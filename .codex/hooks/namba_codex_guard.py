#!/usr/bin/env python3
"""NambaAI Codex lifecycle hook guard.

This script is intentionally conservative. It adds lightweight NambaAI context,
blocks only clearly dangerous shell patterns, and reminds Codex to refresh
generated Namba surfaces when those files are touched.
"""

import json
import os
import re
import subprocess
import sys


DANGEROUS_PATTERNS = [
    (re.compile(r"\bgit\s+reset\s+--hard\b", re.IGNORECASE), "git reset --hard discards repository changes."),
    (re.compile(r"\bgit\s+clean\s+-[^\n;&|]*[fd][^\n;&|]*", re.IGNORECASE), "git clean can delete untracked files."),
    (re.compile(r"\bgit\s+push\b[^\n;&|]*(?:\s-f\b|\s--force(?:-with-lease)?\b)", re.IGNORECASE), "force-pushing requires an explicit human decision."),
    (re.compile(r"\brm\s+-[^\n;&|]*r[^\n;&|]*f[^\n;&|]*\s+(?:/|\*|\.|~|\$HOME)(?:\s|$|/)", re.IGNORECASE), "broad rm -rf targets are blocked by the NambaAI hook guard."),
    (re.compile(r"\brm\s+-[^\n;&|]*f[^\n;&|]*r[^\n;&|]*\s+(?:/|\*|\.|~|\$HOME)(?:\s|$|/)", re.IGNORECASE), "broad rm -rf targets are blocked by the NambaAI hook guard."),
    (re.compile(r"\bchmod\s+-R\s+777\b", re.IGNORECASE), "recursive chmod 777 is too broad for an automated turn."),
    (re.compile(r"\b(?:curl|wget)\b[^\n|;&]*(?:\||>)\s*(?:sudo\s+)?(?:sh|bash)\b", re.IGNORECASE), "piping downloaded content directly into a shell is blocked."),
]

MANAGED_EXACT = {
    "AGENTS.md",
    ".codex/config.toml",
    ".codex/hooks.json",
    ".namba/codex/README.md",
    ".namba/codex/output-contract.md",
    ".namba/codex/validate-output-contract.py",
}

MANAGED_PREFIXES = (
    ".agents/skills/",
    ".codex/agents/",
    ".codex/hooks/",
    ".namba/config/",
)

PROMPT_REFINEMENT_PATTERN = re.compile(
    r"(\bnamba\s+(?:plan|harness|fix|run)\b|\$namba-(?:plan|harness|fix|run|create|queue|review-resolve)|"
    r"\b(?:build|make|fix|improve|update|create|implement)\b|"
    r"(?:만들어|고쳐|수정|개선|구현|추가|해줘|해봐|처리|진행))",
    re.IGNORECASE,
)

VAGUE_MARKERS = (
    "뭔가",
    "대충",
    "적당히",
    "알아서",
    "좋게",
    "이거",
    "그거",
    "애매",
    "문제",
    "버그",
    "개선",
    "만들어줘",
    "해줘",
    "fix it",
    "make it better",
    "improve",
    "something",
    "stuff",
    "thing",
)

REPORT_HEADER = "NAMBA-AI 작업 결과 보고"
REPORT_SECTIONS = ("작업 정의", "판단", "수행한 작업", "현재 이슈", "잠재 문제", "다음 스텝")
CURRENT_PAYLOAD = {}


def read_payload():
    raw = sys.stdin.read()
    if not raw.strip():
        return {}
    try:
        return json.loads(raw)
    except json.JSONDecodeError:
        return {}


def emit(value):
    trace(CURRENT_PAYLOAD, value)
    print(json.dumps(value, ensure_ascii=False, separators=(",", ":")))


def trace(payload, output=None):
    path = os.environ.get("NAMBA_HOOK_TRACE_PATH")
    if not path:
        return
    try:
        event = payload.get("hook_event_name")
        record = {"event": event}
        if event == "UserPromptSubmit":
            record["prompt_refinement"] = bool(prompt_refinement_context(prompt_from(payload)))
        if event in ("PreToolUse", "PermissionRequest"):
            command = command_from(payload)
            record["dangerous"] = bool(dangerous_reason(command))
            record["approval_risk_note"] = bool(approval_risk_note(command))
        if event == "Stop":
            message = payload.get("last_assistant_message")
            record["would_check_report"] = isinstance(message, str) and len(message.strip()) >= 500
        if output is not None:
            record["emitted"] = True
            record["output_keys"] = sorted(output.keys()) if isinstance(output, dict) else []
            if isinstance(output, dict):
                record["decision"] = output.get("decision")
                hook_output = output.get("hookSpecificOutput")
                if isinstance(hook_output, dict):
                    record["hook_event"] = hook_output.get("hookEventName")
        directory = os.path.dirname(path)
        if directory:
            os.makedirs(directory, exist_ok=True)
        with open(path, "a", encoding="utf-8") as handle:
            handle.write(json.dumps(record, ensure_ascii=False, separators=(",", ":")) + "\n")
    except Exception:
        return


def command_from(payload):
    tool_input = payload.get("tool_input")
    if isinstance(tool_input, dict):
        for key in ("command", "cmd"):
            value = tool_input.get(key)
            if isinstance(value, str):
                return value
    return ""


def dangerous_reason(command):
    for pattern, reason in DANGEROUS_PATTERNS:
        if pattern.search(command):
            return reason
    return ""


def approval_risk_note(command):
    command = command.strip()
    if not command:
        return ""
    lowered = command.lower()
    risks = []
    if "sudo" in lowered:
        risks.append("uses sudo and may modify system-level state")
    if re.search(r"\b(?:curl|wget|npm|pnpm|yarn|pip|uv|go)\b", lowered):
        risks.append("may access the network or install executable dependencies")
    if re.search(r"\b(?:git\s+push|gh\s+pr\s+merge|gh\s+release|namba\s+land|namba\s+release)\b", lowered):
        risks.append("may publish, merge, or release remote state")
    if re.search(r"(^|\s)(?:/usr|/opt|/etc|/var|~|\$HOME|/tmp)(?:/|\s|$)", command):
        risks.append("mentions paths outside the repository workspace")
    if re.search(r"\b(?:chmod|chown|launchctl|systemctl|killall)\b", lowered):
        risks.append("may alter permissions or local runtime processes")
    if not risks:
        return ""
    return "NambaAI approval note: " + "; ".join(risks) + ". Approve only if this matches the user's latest intent and the command is scoped."


def prompt_from(payload):
    prompt = payload.get("prompt")
    return prompt if isinstance(prompt, str) else ""


def prompt_refinement_context(prompt):
    normalized = " ".join(prompt.split())
    if not normalized or not PROMPT_REFINEMENT_PATTERN.search(normalized):
        return ""
    lower = normalized.lower()
    namba_workflow = bool(re.search(r"\bnamba\s+(?:plan|harness|fix|run)\b|\$namba-", lower))
    too_short = len(normalized) < 80
    vague = any(marker in lower for marker in VAGUE_MARKERS)
    has_acceptance = bool(re.search(r"acceptance|criteria|success|완료|성공|검증|테스트|제약|constraint|scope|범위", lower))
    if not namba_workflow and not too_short and not vague:
        return ""
    if has_acceptance and len(normalized) >= 120 and not vague:
        return ""
    return (
        "NambaAI prompt-refinement gate: treat this as a spec-first request before implementation. "
        "If the goal, target surface, constraints, acceptance criteria, or existing context are unclear, "
        "ask 1-3 concise clarifying questions before running a Namba workflow or editing files. "
        "Then restate the improved prompt as Goal, Scope, Constraints, Acceptance. "
        "This is inspired by a Socratic/Ontology workflow: reduce ambiguity before code."
    )


def prompt_refinement_block_reason(prompt):
    context = prompt_refinement_context(prompt)
    if not context:
        return ""
    normalized = " ".join(prompt.split())
    korean = bool(re.search(r"[가-힣]", normalized))
    if korean:
        questions = [
            "1. 이 작업의 대상 surface는 무엇인가요? 예: CLI, 웹 앱, API, 특정 모듈.",
            "2. 원하는 사용자 흐름과 제외할 범위는 무엇인가요?",
            "3. 완료 기준과 검증 방법은 무엇인가요?",
        ]
        return (
            "NambaAI 프롬프트 교정 게이트: 실행 전에 모호함을 먼저 줄여야 합니다.\n"
            + "\n".join(questions)
            + "\n\n답변은 가능하면 다음 형식으로 주세요:\n"
            "Goal: ...\nScope: ...\nConstraints: ...\nAcceptance: ..."
        )
    questions = [
        "1. What target surface should this change affect, such as CLI, web app, API, or a specific module?",
        "2. What user flow is desired, and what should stay out of scope?",
        "3. What acceptance criteria and validation should define done?",
    ]
    return (
        "NambaAI prompt-refinement gate: reduce ambiguity before execution.\n"
        + "\n".join(questions)
        + "\n\nPlease answer in this shape when possible:\n"
        "Goal: ...\nScope: ...\nConstraints: ...\nAcceptance: ..."
    )


def handle_session_start():
    emit({
        "hookSpecificOutput": {
            "hookEventName": "SessionStart",
            "additionalContext": (
                "NambaAI lifecycle hook is active. Treat .namba/ as the source "
                "of truth, use AGENTS.md and .agents/skills/ for workflow routing, "
                "and run configured validation after changes. Codex hooks are "
                "guardrails, not a complete security boundary."
            ),
        }
    })


def handle_pre_tool_use(payload):
    reason = dangerous_reason(command_from(payload))
    if not reason:
        return
    emit({
        "hookSpecificOutput": {
            "hookEventName": "PreToolUse",
            "permissionDecision": "deny",
            "permissionDecisionReason": reason,
        }
    })


def handle_permission_request(payload):
    reason = dangerous_reason(command_from(payload))
    if not reason:
        note = approval_risk_note(command_from(payload))
        if note:
            emit({"systemMessage": note})
        return
    emit({
        "hookSpecificOutput": {
            "hookEventName": "PermissionRequest",
            "decision": {
                "behavior": "deny",
                "message": reason,
            },
        }
    })


def handle_user_prompt_submit(payload):
    reason = prompt_refinement_block_reason(prompt_from(payload))
    if not reason:
        return
    emit({
        "decision": "block",
        "reason": reason,
    })


def changed_paths(cwd):
    try:
        result = subprocess.run(
            ["git", "status", "--short"],
            cwd=cwd,
            text=True,
            capture_output=True,
            timeout=5,
            check=False,
        )
    except Exception:
        return []
    if result.returncode != 0:
        return []
    paths = []
    for line in result.stdout.splitlines():
        if len(line) < 4:
            continue
        path = line[3:].strip()
        if " -> " in path:
            path = path.split(" -> ", 1)[1].strip()
        path = path.strip('"')
        if path:
            paths.append(path)
    return paths


def is_namba_surface(path):
    if path in MANAGED_EXACT:
        return True
    return any(path.startswith(prefix) for prefix in MANAGED_PREFIXES)


def handle_post_tool_use(payload):
    cwd = payload.get("cwd") if isinstance(payload.get("cwd"), str) else os.getcwd()
    touched = [path for path in changed_paths(cwd) if is_namba_surface(path)]
    if not touched:
        return
    shown = ", ".join(touched[:6])
    if len(touched) > 6:
        shown += ", ..."
    emit({
        "hookSpecificOutput": {
            "hookEventName": "PostToolUse",
            "additionalContext": (
                "Namba-managed instruction or config surfaces are changed: "
                + shown
                + ". Update source templates first, run namba regen when scaffold "
                "outputs changed, then run the configured validation commands."
            ),
        }
    })


def namba_repo(cwd):
    return os.path.exists(os.path.join(cwd, ".namba")) or os.path.exists(os.path.join(cwd, "AGENTS.md"))


def report_missing_sections(message):
    if REPORT_HEADER not in message:
        return list(REPORT_SECTIONS)
    positions = []
    for section in REPORT_SECTIONS:
        index = message.find(section)
        if index < 0:
            return [section]
        positions.append(index)
    if positions != sorted(positions):
        return ["section order"]
    return []


def handle_stop(payload):
    if payload.get("stop_hook_active") is True:
        return
    message = payload.get("last_assistant_message")
    if not isinstance(message, str) or len(message.strip()) < 500:
        return
    cwd = payload.get("cwd") if isinstance(payload.get("cwd"), str) else os.getcwd()
    if not namba_repo(cwd):
        return
    nambaish = any(token in message.lower() for token in ("namba", ".namba", "codex", "spec-", "검증", "브랜치"))
    if not nambaish:
        return
    missing = report_missing_sections(message)
    if not missing:
        return
    emit({
        "decision": "block",
        "reason": (
            "Before ending, rewrite the final response using the Namba report frame: "
            "# NAMBA-AI 작업 결과 보고, then 🧭 작업 정의, 🧠 판단, 🛠 수행한 작업, "
            "🚧 현재 이슈, ⚠ 잠재 문제, ➡ 다음 스텝. Keep it concise and high-signal."
        ),
    })


def main():
    global CURRENT_PAYLOAD
    payload = read_payload()
    CURRENT_PAYLOAD = payload
    trace(payload)
    event = payload.get("hook_event_name")
    if event == "SessionStart":
        handle_session_start()
    elif event == "PreToolUse":
        handle_pre_tool_use(payload)
    elif event == "PermissionRequest":
        handle_permission_request(payload)
    elif event == "UserPromptSubmit":
        handle_user_prompt_submit(payload)
    elif event == "PostToolUse":
        handle_post_tool_use(payload)
    elif event == "Stop":
        handle_stop(payload)


if __name__ == "__main__":
    main()

#!/bin/sh
# NambaAI Codex lifecycle hook launcher for POSIX shells.

script_dir=${0%/*}
if [ "$script_dir" = "$0" ]; then
    script_dir="."
fi
script_path="$script_dir/namba_codex_guard.py"
raw_input=""

while IFS= read -r line || [ -n "$line" ]; do
    raw_input="${raw_input}${line}
"
done

failure_message="NambaAI hook guard failed: Python 3 was not found or could not run .codex/hooks/namba_codex_guard.py. Install Python 3, then run namba regen and review /hooks again. Hooks are guardrails, so continue only with explicit validation."
failed_candidates=""

emit_failure() {
    message=$1
    printf '%s\n' "$message" >&2
    printf '{"hookSpecificOutput":{"hookEventName":"Unknown","additionalContext":"%s"}}\n' "$message"
}

run_candidate() {
    executable=$1
    shift
    if ! command -v "$executable" >/dev/null 2>&1; then
        return 127
    fi

    stderr_path="${TMPDIR:-/tmp}/namba_codex_hook_$$.stderr"
    : > "$stderr_path" 2>/dev/null || stderr_path=""

    if [ -n "$stderr_path" ]; then
        output=$(printf '%s' "$raw_input" | "$executable" "$@" 2>"$stderr_path")
        status=$?
        stderr_text=""
        while IFS= read -r err_line || [ -n "$err_line" ]; do
            stderr_text="${stderr_text}${err_line} "
        done < "$stderr_path"
        : > "$stderr_path" 2>/dev/null || true
    else
        output=$(printf '%s' "$raw_input" | "$executable" "$@" 2>&1)
        status=$?
        stderr_text=""
    fi

    if [ "$status" -eq 0 ]; then
        if [ -n "$stderr_text" ]; then
            printf '%s\n' "$stderr_text" >&2
        fi
        if [ -n "$output" ]; then
            printf '%s\n' "$output"
        fi
        exit 0
    fi

    failed_candidates="${failed_candidates}${executable} exited ${status}; "
    if [ -n "$stderr_text" ]; then
        printf '%s failed: %s\n' "$executable" "$stderr_text" >&2
    fi
    return 1
}

run_candidate python3 "$script_path"
run_candidate python "$script_path"
run_candidate py -3 "$script_path"

if [ -n "$failed_candidates" ]; then
    printf 'NambaAI hook launcher candidates failed: %s\n' "$failed_candidates" >&2
fi
emit_failure "$failure_message"
exit 1

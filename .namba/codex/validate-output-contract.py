#!/usr/bin/env python3
"""Validate the NambaAI output contract from stdin or a file."""

from __future__ import annotations

import argparse
import pathlib
import re
import sys

SPEC = {
  "header": "NAMBA-AI 작업 결과 보고",
  "header_aliases": [
    "NAMBA-AI 작업 결과 보고",
    "NAMBA-AI 작업 보고",
    "NAMBA-AI 엔지니어링 보고"
  ],
  "sections": [
    {
      "emoji": "🧭",
      "primary": "작업 정의",
      "aliases": [
        "작업 정의",
        "정의",
        "정의한 범위",
        "문제 정의"
      ]
    },
    {
      "emoji": "🧠",
      "primary": "판단",
      "aliases": [
        "판단",
        "내린 판단",
        "핵심 판단",
        "결정"
      ]
    },
    {
      "emoji": "🛠",
      "primary": "수행한 작업",
      "aliases": [
        "수행한 작업",
        "진행한 작업",
        "작업 내용",
        "적용한 작업"
      ]
    },
    {
      "emoji": "🚧",
      "primary": "현재 이슈",
      "aliases": [
        "현재 이슈",
        "이슈",
        "남은 이슈",
        "현재 문제"
      ]
    },
    {
      "emoji": "⚠",
      "primary": "잠재 문제",
      "aliases": [
        "잠재 문제",
        "잠재 리스크",
        "위험 요소",
        "잠재 이슈"
      ]
    },
    {
      "emoji": "➡",
      "primary": "다음 스텝",
      "aliases": [
        "다음 스텝",
        "다음 단계",
        "추천",
        "권장 흐름"
      ]
    }
  ]
}


def build_pattern(aliases: list[str]) -> re.Pattern[str]:
    escaped = "|".join(re.escape(alias) for alias in aliases)
    return re.compile(r"^\s*(?:#{1,6}\s*|[-*]\s+)?(?:\*\*)?[\W_]*(?P<label>(" + escaped + r"))(?:\*\*)?\s*(?:[:：-].*)?$")


def read_text(args: argparse.Namespace) -> str:
    if args.file:
        return pathlib.Path(args.file).read_text(encoding="utf-8")
    return sys.stdin.read().lstrip('\ufeff')


def find_first_match(lines: list[str], aliases: list[str], start: int = 0) -> int:
    pattern = build_pattern(aliases)
    for index, line in enumerate(lines[start:], start=start):
        if pattern.match(line.strip()):
            return index
    return -1


def main() -> int:
    parser = argparse.ArgumentParser(description="Validate the NambaAI output contract.")
    parser.add_argument("--file", help="Path to a saved response file.")
    args = parser.parse_args()

    text = read_text(args)
    if not text.strip():
        print("output-contract: empty input", file=sys.stderr)
        return 1

    lines = [line.lstrip('\ufeff') for line in text.lstrip('\ufeff').splitlines()]
    header_index = find_first_match(lines, SPEC['header_aliases'])
    if header_index < 0:
        print(f"output-contract: missing header '{SPEC['header']}'", file=sys.stderr)
        return 1

    previous = header_index
    for section in SPEC['sections']:
        found = find_first_match(lines, section['aliases'], start=previous + 1)
        if found < 0:
            print(f"output-contract: missing section '{section['primary']}'", file=sys.stderr)
            return 1
        if found <= previous:
            print(f"output-contract: section '{section['primary']}' is out of order", file=sys.stderr)
            return 1
        previous = found

    print("output-contract: ok")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

#!/usr/bin/env python3
"""Validate the NambaAI output contract from stdin or a file."""

from __future__ import annotations

import argparse
import pathlib
import re
import sys

SECTIONS = [
    ("오늘의 결정", ["오늘의 결정", "핵심 판단", "이번 수", "결론"]),
    ("판단 근거", ["판단 근거", "왜 이렇게 봤나", "근거", "이유"]),
    ("검증 경로", ["검증 경로", "검증 방법", "확인 루트", "검증"]),
    ("무너지는 조건", ["무너지는 조건", "실패 조건", "경계 조건", "리스크"]),
    ("다음 수", ["다음 수", "추천", "권장 흐름", "다음 단계"]),
]


def build_pattern(aliases: list[str]) -> re.Pattern[str]:
    escaped = "|".join(re.escape(alias) for alias in aliases)
    return re.compile(r"^\s*(?:#{1,6}\s*|[-*]\s+)?(?:\*\*)?(?P<label>(" + escaped + r"))(?:\*\*)?\s*(?:[:：-].*)?$")


def read_text(args: argparse.Namespace) -> str:
    if args.file:
        return pathlib.Path(args.file).read_text(encoding="utf-8")
    return sys.stdin.read().lstrip('\ufeff')


def main() -> int:
    parser = argparse.ArgumentParser(description="Validate the NambaAI output contract.")
    parser.add_argument("--file", help="Path to a saved response file.")
    args = parser.parse_args()

    text = read_text(args)
    if not text.strip():
        print("output-contract: empty input", file=sys.stderr)
        return 1

    lines = [line.lstrip('\ufeff') for line in text.lstrip('\ufeff').splitlines()]
    positions: list[tuple[str, int]] = []
    for expected, aliases in SECTIONS:
        pattern = build_pattern(aliases)
        found = -1
        for index, line in enumerate(lines):
            if pattern.match(line.strip()):
                found = index
                break
        if found < 0:
            print(f"output-contract: missing section '{expected}'", file=sys.stderr)
            return 1
        positions.append((expected, found))

    previous = -1
    for expected, found in positions:
        if found <= previous:
            print(f"output-contract: section '{expected}' is out of order", file=sys.stderr)
            return 1
        previous = found

    print("output-contract: ok")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

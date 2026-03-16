#!/usr/bin/env python3
"""
bench/check_regression.py — Regression guard for agent mode tokens.
Compares bench/results/tokens.csv against bench/results/baseline.csv.
Exits non-zero if any command's agent token count increased.

Usage:
    python3 bench/check_regression.py [--baseline bench/results/baseline.csv] [--current bench/results/tokens.csv]

To update the baseline after an intentional improvement:
    cp bench/results/tokens.csv bench/results/baseline.csv
"""
import argparse
import csv
import os
import sys


def load_csv(path: str) -> dict:
    """Returns {command: tokens} for agent mode rows only."""
    result = {}
    with open(path, newline="", encoding="utf-8") as f:
        for row in csv.DictReader(f):
            if row["mode"] == "agent":
                result[row["command"]] = int(row["tokens"])
    return result


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--baseline", default="bench/results/baseline.csv")
    parser.add_argument("--current", default="bench/results/tokens.csv")
    args = parser.parse_args()

    if not os.path.exists(args.baseline):
        print(f"No baseline found at {args.baseline} — saving current as baseline.")
        import shutil
        shutil.copy(args.current, args.baseline)
        print("Baseline created. Re-run to compare.")
        sys.exit(0)

    baseline = load_csv(args.baseline)
    current = load_csv(args.current)

    # Allow ±5 token variance (request_id UUIDs change on each call)
    TOLERANCE = 5

    regressions = []
    improvements = []

    for cmd, cur_tokens in current.items():
        base_tokens = baseline.get(cmd)
        if base_tokens is None:
            print(f"  NEW   {cmd}: {cur_tokens} tokens (not in baseline)")
            continue
        diff = cur_tokens - base_tokens
        if diff > TOLERANCE:
            regressions.append((cmd, base_tokens, cur_tokens, diff))
        elif diff < -TOLERANCE:
            improvements.append((cmd, base_tokens, cur_tokens, diff))

    if improvements:
        print("Improvements detected (agent tokens decreased):")
        for cmd, b, c, d in improvements:
            print(f"  ✓  {cmd}: {b} → {c} ({d:+d} tokens)")

    if regressions:
        print("\nREGRESSIONS detected (agent tokens increased):")
        for cmd, b, c, d in regressions:
            print(f"  ✗  {cmd}: {b} → {c} ({d:+d} tokens)")
        print("\nFix the render output or run:")
        print("  cp bench/results/tokens.csv bench/results/baseline.csv")
        print("to accept the new baseline intentionally.")
        sys.exit(1)

    if not regressions:
        print("No regressions. Agent token counts are stable or improved.")
        sys.exit(0)


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""
bench/compare.py — Human vs Agent comparison report.
Reads bench/results/tokens.csv and generates bench/results/report.md.

Usage:
    python3 bench/compare.py [--csv bench/results/tokens.csv] [--out bench/results/report.md]
"""
import argparse
import csv
import os
import sys
from datetime import date


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--csv", default="bench/results/tokens.csv")
    parser.add_argument("--out", default="bench/results/report.md")
    args = parser.parse_args()

    if not os.path.exists(args.csv):
        print(f"CSV not found: {args.csv}", file=sys.stderr)
        print("Run bench/corpus.sh then bench/count_tokens.py first.", file=sys.stderr)
        sys.exit(1)

    # Load rows
    human = {}
    agent = {}
    with open(args.csv, newline="", encoding="utf-8") as f:
        for row in csv.DictReader(f):
            cmd = row["command"]
            if row["mode"] == "human":
                human[cmd] = row
            elif row["mode"] == "agent":
                agent[cmd] = row

    # Build comparison table
    commands = sorted(set(human) | set(agent))
    rows = []
    total_human_tokens = 0
    total_agent_tokens = 0
    total_human_bytes = 0
    total_agent_bytes = 0

    for cmd in commands:
        h = human.get(cmd)
        a = agent.get(cmd)
        if not h or not a:
            continue
        ht = int(h["tokens"])
        at = int(a["tokens"])
        hb = int(h["bytes"])
        ab = int(a["bytes"])
        tok_saved = ht - at
        tok_pct = round((tok_saved / ht) * 100, 1) if ht > 0 else 0
        byte_saved = hb - ab
        byte_pct = round((byte_saved / hb) * 100, 1) if hb > 0 else 0
        rows.append({
            "command": cmd,
            "human_bytes": hb,
            "agent_bytes": ab,
            "byte_saved": byte_saved,
            "byte_pct": byte_pct,
            "human_tokens": ht,
            "agent_tokens": at,
            "tok_saved": tok_saved,
            "tok_pct": tok_pct,
        })
        total_human_tokens += ht
        total_agent_tokens += at
        total_human_bytes += hb
        total_agent_bytes += ab

    total_tok_saved = total_human_tokens - total_agent_tokens
    total_tok_pct = round((total_tok_saved / total_human_tokens) * 100, 1) if total_human_tokens > 0 else 0
    total_byte_saved = total_human_bytes - total_agent_bytes
    total_byte_pct = round((total_byte_saved / total_human_bytes) * 100, 1) if total_human_bytes > 0 else 0

    # Print to stdout
    print(f"{'COMMAND':<45} {'H.TOK':>7} {'A.TOK':>7} {'SAVED':>7} {'%':>6}")
    print("-" * 76)
    for r in rows:
        print(f"{r['command']:<45} {r['human_tokens']:>7} {r['agent_tokens']:>7} {r['tok_saved']:>7} {r['tok_pct']:>5}%")
    print("-" * 76)
    print(f"{'TOTAL':<45} {total_human_tokens:>7} {total_agent_tokens:>7} {total_tok_saved:>7} {total_tok_pct:>5}%")

    # Write markdown report
    os.makedirs(os.path.dirname(args.out), exist_ok=True)
    with open(args.out, "w", encoding="utf-8") as f:
        f.write(f"# Benchmark Report — Human vs Agent\n\n")
        f.write(f"Generated: {date.today()}  \n")
        f.write(f"Encoding: cl100k_base (tiktoken)  \n\n")

        f.write("## Token comparison\n\n")
        f.write("| Command | Human tokens | Agent tokens | Saved | % |\n")
        f.write("|---|---:|---:|---:|---:|\n")
        for r in rows:
            sign = "−" if r["tok_saved"] >= 0 else "+"
            f.write(f"| `{r['command']}` | {r['human_tokens']} | {r['agent_tokens']} | {sign}{abs(r['tok_saved'])} | {r['tok_pct']}% |\n")
        f.write(f"| **TOTAL** | **{total_human_tokens}** | **{total_agent_tokens}** | **−{total_tok_saved}** | **{total_tok_pct}%** |\n\n")

        f.write("## Byte comparison\n\n")
        f.write("| Command | Human bytes | Agent bytes | Saved | % |\n")
        f.write("|---|---:|---:|---:|---:|\n")
        for r in rows:
            sign = "−" if r["byte_saved"] >= 0 else "+"
            f.write(f"| `{r['command']}` | {r['human_bytes']} | {r['agent_bytes']} | {sign}{abs(r['byte_saved'])} | {r['byte_pct']}% |\n")
        f.write(f"| **TOTAL** | **{total_human_bytes}** | **{total_agent_bytes}** | **−{total_byte_saved}** | **{total_byte_pct}%** |\n\n")

        f.write("## Summary\n\n")
        f.write(f"- Token reduction: **{total_tok_pct}%** ({total_tok_saved} tokens saved over {len(rows)} commands)\n")
        f.write(f"- Byte reduction: **{total_byte_pct}%** ({total_byte_saved} bytes saved)\n")
        f.write(f"- Agent mode returns pure minified JSON — zero decorative output on stdout\n")

    print(f"\nReport saved to {args.out}")


if __name__ == "__main__":
    main()

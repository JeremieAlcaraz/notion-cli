#!/usr/bin/env python3
"""
bench/count_tokens.py — Count tokens for each raw output file.
Uses tiktoken with cl100k_base encoding (GPT-4 / Claude-compatible approximation).

Usage:
    python3 bench/count_tokens.py [--raw-dir bench/results/raw] [--out bench/results/tokens.csv]

Output CSV columns: mode, command, bytes, tokens, tokens_per_byte
"""
import argparse
import csv
import os
import sys

try:
    import tiktoken
except ImportError:
    print("tiktoken not installed. Run: uv pip install tiktoken", file=sys.stderr)
    sys.exit(1)

enc = tiktoken.get_encoding("cl100k_base")


def count_tokens(text: str) -> int:
    return len(enc.encode(text))


def main():
    parser = argparse.ArgumentParser(description="Count tokens in benchmark outputs")
    parser.add_argument("--raw-dir", default="bench/results/raw", help="Directory with raw output files")
    parser.add_argument("--out", default="bench/results/tokens.csv", help="Output CSV path")
    args = parser.parse_args()

    raw_dir = args.raw_dir
    if not os.path.isdir(raw_dir):
        print(f"Raw dir not found: {raw_dir}", file=sys.stderr)
        print("Run bench/corpus.sh first.", file=sys.stderr)
        sys.exit(1)

    rows = []
    for fname in sorted(os.listdir(raw_dir)):
        if not fname.endswith(".json"):
            continue
        # filename format: {mode}_{command...}.json
        parts = fname.replace(".json", "").split("_", 1)
        if len(parts) < 2:
            continue
        mode = parts[0]
        command = parts[1].replace("_", " ")

        fpath = os.path.join(raw_dir, fname)
        with open(fpath, "r", encoding="utf-8") as f:
            content = f.read()

        byte_count = len(content.encode("utf-8"))
        token_count = count_tokens(content)
        tpb = round(token_count / byte_count, 4) if byte_count > 0 else 0

        rows.append({
            "mode": mode,
            "command": command,
            "bytes": byte_count,
            "tokens": token_count,
            "tokens_per_byte": tpb,
        })

    os.makedirs(os.path.dirname(args.out), exist_ok=True)
    with open(args.out, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=["mode", "command", "bytes", "tokens", "tokens_per_byte"])
        writer.writeheader()
        writer.writerows(rows)

    # Pretty-print to stdout
    print(f"{'MODE':<8} {'COMMAND':<45} {'BYTES':>8} {'TOKENS':>8} {'TOK/B':>7}")
    print("-" * 82)
    for r in rows:
        print(f"{r['mode']:<8} {r['command']:<45} {r['bytes']:>8} {r['tokens']:>8} {r['tokens_per_byte']:>7}")

    print(f"\nSaved to {args.out}")


if __name__ == "__main__":
    main()

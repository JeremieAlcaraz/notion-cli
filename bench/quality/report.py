#!/usr/bin/env python3
"""
bench/quality/report.py — Cross-report: tokens saved × quality score.

Runs each scenario in baseline and optimized mode, computes:
  - tokens baseline vs optimized
  - quality score (PASS/FAIL) for each mode
  - verdict: ✅ optimisé / ❌ trop agressif / ➡ neutre

Usage:
    python3 bench/quality/report.py [--scenarios bench/quality/scenarios.yaml]
                                     [--out bench/quality/results/report.md]
"""
import argparse
import json
import os
import sys
from datetime import date

try:
    import yaml
except ImportError:
    print("pyyaml not installed.")
    sys.exit(1)

# Reuse eval logic
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from eval import evaluate_scenario


def verdict(baseline_pass: bool, optimized_pass: bool, tok_saved: int) -> str:
    if not baseline_pass:
        return "⚠️  baseline fail"
    if optimized_pass and tok_saved > 0:
        return "✅ optimisé"
    if optimized_pass and tok_saved == 0:
        return "➡️  neutre"
    if not optimized_pass and tok_saved > 0:
        return "❌ trop agressif"
    return "❌ dégradé"


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--scenarios", default="bench/quality/scenarios.yaml")
    parser.add_argument("--out", default="bench/quality/results/report.md")
    args = parser.parse_args()

    with open(args.scenarios) as f:
        data = yaml.safe_load(f)
    scenarios = data["scenarios"]

    rows = []
    for scenario in scenarios:
        print(f"  {scenario['id']} — {scenario['description']} ...", flush=True)
        result = evaluate_scenario(scenario, mode="both")

        b = result["modes"].get("baseline", {})
        o = result["modes"].get("optimized", {})

        b_tokens = b.get("tokens", 0)
        o_tokens = o.get("tokens", 0)
        tok_saved = b_tokens - o_tokens
        tok_pct = round((tok_saved / b_tokens) * 100, 1) if b_tokens > 0 else 0

        b_pass = b.get("verdict") == "PASS"
        o_pass = o.get("verdict") == "PASS"

        rows.append({
            "id": scenario["id"],
            "description": scenario["description"],
            "baseline_cmd": scenario["baseline_command"].replace(scenario["baseline_command"].split()[0], "notion"),
            "optimized_cmd": scenario.get("optimized_command", ""),
            "b_tokens": b_tokens,
            "o_tokens": o_tokens,
            "tok_saved": tok_saved,
            "tok_pct": tok_pct,
            "b_verdict": b.get("verdict", "?"),
            "o_verdict": o.get("verdict", "?"),
            "verdict": verdict(b_pass, o_pass, tok_saved),
        })

    # Console summary
    print()
    print(f"{'ID':<5} {'B.TOK':>7} {'O.TOK':>7} {'SAVED':>7} {'%':>6}  {'B':^6} {'O':^6}  VERDICT")
    print("─" * 80)
    total_b = total_o = 0
    for r in rows:
        b_sym = "✓" if r["b_verdict"] == "PASS" else "✗"
        o_sym = "✓" if r["o_verdict"] == "PASS" else "✗"
        print(f"{r['id']:<5} {r['b_tokens']:>7} {r['o_tokens']:>7} {r['tok_saved']:>7} {r['tok_pct']:>5}%  {b_sym:^6} {o_sym:^6}  {r['verdict']}")
        total_b += r["b_tokens"]
        total_o += r["o_tokens"]

    total_saved = total_b - total_o
    total_pct = round((total_saved / total_b) * 100, 1) if total_b > 0 else 0
    optimized_quality = sum(1 for r in rows if r["o_verdict"] == "PASS")
    print("─" * 80)
    print(f"{'TOT':<5} {total_b:>7} {total_o:>7} {total_saved:>7} {total_pct:>5}%  {'':^6} {optimized_quality}/10   overall")

    # Markdown report
    os.makedirs(os.path.dirname(args.out), exist_ok=True)
    with open(args.out, "w", encoding="utf-8") as f:
        f.write(f"# Quality × Efficiency Report\n\n")
        f.write(f"Generated: {date.today()}  \n")
        f.write(f"Baseline: raw agent output — Optimized: with `--fields` / `--format` flags\n\n")

        f.write("## Results\n\n")
        f.write("| Scenario | Baseline tokens | Optimized tokens | Saved | % | Baseline ✓ | Optimized ✓ | Verdict |\n")
        f.write("|---|---:|---:|---:|---:|:---:|:---:|---|\n")
        for r in rows:
            b_sym = "✓" if r["b_verdict"] == "PASS" else "✗"
            o_sym = "✓" if r["o_verdict"] == "PASS" else "✗"
            sign = "−" if r["tok_saved"] >= 0 else "+"
            f.write(f"| **{r['id']}** {r['description']} | {r['b_tokens']} | {r['o_tokens']} | {sign}{abs(r['tok_saved'])} | {r['tok_pct']}% | {b_sym} | {o_sym} | {r['verdict']} |\n")
        f.write(f"| **TOTAL** | **{total_b}** | **{total_o}** | **−{total_saved}** | **{total_pct}%** | | {optimized_quality}/10 | |\n\n")

        f.write("## Optimization flags used\n\n")
        seen = set()
        for r in rows:
            opt = r["optimized_cmd"]
            # Extract flags
            flags = []
            for flag in ["--fields", "--format", "--strip-meta", "--summary", "--all"]:
                if flag in opt and flag not in seen:
                    flags.append(flag)
                    seen.add(flag)
            if flags:
                f.write(f"- `{r['id']}`: {', '.join(f'`{fl}`' for fl in flags)}\n")

        f.write("\n## Key findings\n\n")

        safe = [r for r in rows if r["verdict"].startswith("✅")]
        aggressive = [r for r in rows if r["verdict"].startswith("❌")]
        neutral = [r for r in rows if r["verdict"].startswith("➡")]

        f.write(f"- **{len(safe)}/10 scenarios** can be optimized without quality loss\n")
        if aggressive:
            f.write(f"- **{len(aggressive)}/10 scenarios** lose quality with current flags (too aggressive)\n")
            for r in aggressive:
                f.write(f"  - `{r['id']}`: {r['description']}\n")
        if neutral:
            f.write(f"- **{len(neutral)}/10 scenarios** show no token reduction with current flags\n")
        f.write(f"- Total token reduction across all scenarios: **{total_pct}%** ({total_saved} tokens saved)\n")
        f.write(f"- Optimized quality score: **{optimized_quality}/10**\n")

        f.write("\n## Per-scenario detail\n\n")
        for r in rows:
            f.write(f"### {r['id']} — {r['description']}\n\n")
            f.write(f"- Baseline: `{r['baseline_cmd'].split('notion ')[1] if 'notion ' in r['baseline_cmd'] else r['baseline_cmd']}`\n")
            opt_short = r['optimized_cmd'].replace(r['optimized_cmd'].split()[0], 'notion') if r['optimized_cmd'] else ''
            f.write(f"- Optimized: `{opt_short.split('notion ')[1] if 'notion ' in opt_short else opt_short}`\n")
            f.write(f"- Tokens: {r['b_tokens']} → {r['o_tokens']} ({r['tok_pct']}% reduction)\n")
            f.write(f"- Verdict: {r['verdict']}\n\n")

    print(f"\nReport saved to {args.out}")


if __name__ == "__main__":
    main()

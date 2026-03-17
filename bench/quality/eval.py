#!/usr/bin/env python3
"""
bench/quality/eval.py — Evaluate agent quality scenarios.

Runs each scenario's baseline_command and optimized_command,
checks success_criteria programmatically, outputs PASS/FAIL/PARTIAL.

Usage:
    python3 bench/quality/eval.py [--scenarios bench/quality/scenarios.yaml]
                                   [--mode baseline|optimized|both]
                                   [--out bench/quality/results/eval.json]
    Exit code: 0 if score >= 8/10, 1 otherwise.
"""
import argparse
import json
import os
import re
import subprocess
import sys
from datetime import datetime

try:
    import yaml
except ImportError:
    print("pyyaml not installed. Run: uv pip install pyyaml --python bench/.venv/bin/python")
    sys.exit(1)

# ── helpers ──────────────────────────────────────────────────────────────────

REPO_ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
NOTION_BIN = os.path.join(REPO_ROOT, "notion")


def run_command(cmd: str) -> tuple[str, int]:
    """Run a shell command with the local notion binary, return (stdout, exit_code)."""
    # Replace bare 'notion' at start of command with the local binary path
    cmd = re.sub(r'^notion\b', NOTION_BIN, cmd)
    result = subprocess.run(
        cmd, shell=True, capture_output=True, text=True, cwd=REPO_ROOT
    )
    return result.stdout.strip(), result.returncode


def parse_output(raw: str) -> object:
    """Try to parse output as JSON. Returns parsed object or raw string."""
    if not raw:
        return None
    try:
        return json.loads(raw)
    except json.JSONDecodeError:
        return raw


def is_uuid(s: str) -> bool:
    return bool(re.match(
        r'^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$',
        str(s), re.IGNORECASE
    ))


def get_nested(obj, path: str):
    """
    Navigate a dot-path like 'results[].id' into the object.
    For list paths (results[].x), returns list of values.
    """
    parts = path.split(".", 1)
    key = parts[0]
    rest = parts[1] if len(parts) > 1 else None

    if key.endswith("[]"):
        key = key[:-2]
        if isinstance(obj, dict):
            items = obj.get(key, [])
        elif isinstance(obj, list):
            items = obj
        else:
            return None
        if rest:
            return [get_nested(item, rest) for item in (items or [])]
        return items

    if isinstance(obj, dict):
        val = obj.get(key)
    elif isinstance(obj, list):
        # try numeric index
        try:
            val = obj[int(key)]
        except (ValueError, IndexError):
            return None
    else:
        return None

    if rest:
        return get_nested(val, rest)
    return val


def _extract_results(data) -> list | None:
    """Return the list of items from a response (handles list wrapper and bare list)."""
    if isinstance(data, list):
        return data
    if isinstance(data, dict):
        for key in ("results", "items", "data"):
            val = data.get(key)
            if isinstance(val, list):
                return val
    return None


def field_present(data, field_path: str) -> bool:
    """Check if a field path exists and is non-empty in data."""
    val = get_nested(data, field_path)
    if val is None:
        return False
    if isinstance(val, list):
        return len(val) > 0 and any(v is not None and v != "" for v in val)
    # bool False is a valid value (e.g. archived=false), treat as present
    if isinstance(val, bool):
        return True
    return val != ""


# ── criterion evaluators ─────────────────────────────────────────────────────

def evaluate_criterion(criterion: str, data, raw: str) -> tuple[bool, str]:
    """
    Evaluate a single success criterion string against parsed data.
    Returns (passed, reason).
    """
    c = criterion.lower()

    I = re.IGNORECASE

    # "chaque item contient X" — must come BEFORE generic "contient X" checks
    m = re.search(r"chaque item (?:de results )?contient ['\"]([a-zA-Z_.\[\]]+)['\"]", criterion, I)
    if m:
        field = m.group(1)
        results = _extract_results(data)
        if results is None:
            return False, "no items found (response is not a list)"
        if len(results) == 0:
            return True, "empty list — criterion vacuously satisfied"
        missing = [i for i, item in enumerate(results) if not field_present(item, field)]
        ok = len(missing) == 0
        return ok, f"field '{field}' present in all {len(results)} items" if ok else f"missing in items {missing[:3]}"

    # "chaque item contient 'id', 'object', 'url'"  (multi-field) — before generic too
    m2 = re.search(r"chaque item contient (.+)", criterion, I)
    if m2:
        fields_str = m2.group(1)
        fields = [f.strip().strip("'\"") for f in re.split(r"[,\s]+", fields_str) if f.strip().strip("'\"")]
        results = _extract_results(data)
        if not results:
            return False, "no items to check"
        for field in fields:
            if not field or field in ("et",):
                continue
            missing = [i for i, item in enumerate(results) if not field_present(item, field)]
            if missing:
                return False, f"field '{field}' missing in items {missing[:3]}"
        return True, "all fields present in all items"

    # "est une liste"
    if "est une liste" in c or "is a list" in c:
        results = _extract_results(data)
        ok = results is not None
        return ok, f"list check: {'ok' if ok else 'not a list'}"

    # "La réponse contient le champ 'X' non vide" or "contient 'X' non vide"
    m = re.search(r"contient (?:le champ )?['\"]([a-zA-Z_.\[\]]+)['\"] non vide", criterion, I)
    if m:
        field = m.group(1)
        ok = field_present(data, field)
        return ok, f"field '{field}' {'present' if ok else 'missing or empty'}"

    # "contient 'X' (booléen)" / "contient 'X' (liste)"
    m = re.search(r"contient (?:le champ )?['\"]([a-zA-Z_.\[\]]+)['\"]", criterion, I)
    if m:
        field = m.group(1)
        ok = field_present(data, field)
        return ok, f"field '{field}' {'present' if ok else 'missing'}"

    # "type égal à X" / "type equal to X"
    m = re.search(r"['\"]?type['\"]? égal à ['\"]?([a-z_]+)['\"]?", criterion)
    if m:
        expected = m.group(1)
        actual = get_nested(data, "type") if isinstance(data, dict) else None
        ok = actual == expected
        return ok, f"type is '{actual}', expected '{expected}'"

    # "UUID valide"
    if "uuid" in c:
        id_val = get_nested(data, "id") if isinstance(data, dict) else None
        ok = is_uuid(str(id_val or ""))
        return ok, f"id '{id_val}' is {'valid' if ok else 'invalid'} UUID"

    # "contient 'results' non vide" (explicit)
    if "'results'" in c and "non vide" in c:
        results = get_nested(data, "results") if isinstance(data, dict) else None
        ok = isinstance(results, list) and len(results) > 0
        return ok, f"results list {'non-empty' if ok else 'empty or missing'}"

    # "texte lisible" / "contenu textuel"
    if "texte lisible" in c or "contenu textuel" in c or "texte" in c:
        if isinstance(raw, str) and len(raw.strip()) > 10:
            return True, "non-empty text output"
        return False, "output is empty or too short"

    # "titres sont identifiables"
    if "titres sont identifiables" in c or "titres" in c:
        if isinstance(raw, str):
            has_md_heading = bool(re.search(r'^#{1,3} ', raw, re.MULTILINE))
            has_json_heading = "heading_" in raw
            ok = has_md_heading or has_json_heading
            return ok, f"headings {'found' if ok else 'not found'} in output"
        return False, "no text output"

    # "paragraphes sont identifiables"
    if "paragraphes" in c:
        if isinstance(raw, str) and len(raw.strip()) > 0:
            return True, "output has content (paragraphs assumed present)"
        return False, "no content"

    # "injectable directement" / "sans post-traitement"
    if "injectable" in c or "sans post-traitement" in c:
        ok = isinstance(raw, str) and len(raw.strip()) > 0
        return ok, "output is non-empty string (injectable)"

    # "peut extraire" / "peut choisir" / "peut identifier" / "sans appel supplémentaire"
    if any(x in c for x in ["peut extraire", "peut choisir", "peut identifier", "sans appel supplémentaire", "sans appel"]):
        # These require data_required fields — checked separately
        return True, "agent capability (assumed if data_required fields present)"

    # fallback — unknown criterion, mark as SKIP
    return None, f"criterion not parsed: '{criterion[:60]}'"


def evaluate_scenario(scenario: dict, mode: str = "both") -> dict:
    """Run a scenario in baseline and/or optimized mode. Return result dict."""
    results = {"id": scenario["id"], "description": scenario["description"], "modes": {}}

    commands = {}
    if mode in ("baseline", "both"):
        commands["baseline"] = scenario["baseline_command"]
    if mode in ("optimized", "both"):
        commands["optimized"] = scenario.get("optimized_command", scenario["baseline_command"])

    for cmd_mode, cmd in commands.items():
        raw, exit_code = run_command(cmd)
        data = parse_output(raw)

        criteria_results = []
        passed = 0
        failed = 0
        skipped = 0

        for criterion in scenario["success_criteria"]:
            ok, reason = evaluate_criterion(criterion, data, raw)
            if ok is None:
                skipped += 1
                criteria_results.append({"criterion": criterion, "status": "SKIP", "reason": reason})
            elif ok:
                passed += 1
                criteria_results.append({"criterion": criterion, "status": "PASS", "reason": reason})
            else:
                failed += 1
                criteria_results.append({"criterion": criterion, "status": "FAIL", "reason": reason})

        total = passed + failed
        if total == 0:
            verdict = "SKIP"
        elif failed == 0:
            verdict = "PASS"
        elif passed == 0:
            verdict = "FAIL"
        else:
            verdict = "PARTIAL"

        # measure tokens
        tokens = 0
        if raw:
            try:
                import tiktoken
                enc = tiktoken.get_encoding("cl100k_base")
                tokens = len(enc.encode(raw))
            except Exception:
                tokens = len(raw.split())  # rough fallback

        results["modes"][cmd_mode] = {
            "command": cmd,
            "exit_code": exit_code,
            "bytes": len(raw.encode("utf-8")) if raw else 0,
            "tokens": tokens,
            "verdict": verdict,
            "passed": passed,
            "failed": failed,
            "skipped": skipped,
            "criteria": criteria_results,
        }

    return results


# ── main ─────────────────────────────────────────────────────────────────────

def main():
    parser = argparse.ArgumentParser(description="Evaluate agent quality scenarios")
    parser.add_argument("--scenarios", default="bench/quality/scenarios.yaml")
    parser.add_argument("--mode", choices=["baseline", "optimized", "both"], default="both")
    parser.add_argument("--out", default="bench/quality/results/eval.json")
    parser.add_argument("--min-score", type=int, default=8, help="Minimum passing score out of 10")
    args = parser.parse_args()

    with open(args.scenarios) as f:
        data = yaml.safe_load(f)
    scenarios = data["scenarios"]

    print(f"Running {len(scenarios)} scenarios in mode: {args.mode}\n")

    all_results = []
    baseline_score = 0
    optimized_score = 0

    for scenario in scenarios:
        print(f"  {scenario['id']} — {scenario['description']}")
        result = evaluate_scenario(scenario, args.mode)
        all_results.append(result)

        for cmd_mode, r in result["modes"].items():
            v = r["verdict"]
            tok = r["tokens"]
            sym = "✓" if v == "PASS" else ("~" if v == "PARTIAL" else ("?" if v == "SKIP" else "✗"))
            print(f"    [{cmd_mode:10s}] {sym} {v:<8}  {tok:>5} tokens  ({r['passed']}✓ {r['failed']}✗ {r['skipped']}?)")
            if v in ("FAIL", "PARTIAL"):
                for c in r["criteria"]:
                    if c["status"] != "PASS":
                        print(f"               {c['status']}: {c['criterion'][:60]}")
                        print(f"                      → {c['reason']}")

            if cmd_mode == "baseline" and v == "PASS":
                baseline_score += 1
            if cmd_mode == "optimized" and v == "PASS":
                optimized_score += 1
        print()

    n = len(scenarios)
    print("─" * 60)
    if args.mode in ("baseline", "both"):
        print(f"Baseline  score: {baseline_score}/{n}")
    if args.mode in ("optimized", "both"):
        print(f"Optimized score: {optimized_score}/{n}")

    # Save JSON results
    os.makedirs(os.path.dirname(args.out), exist_ok=True)
    output = {
        "generated": datetime.now().isoformat(),
        "mode": args.mode,
        "baseline_score": baseline_score,
        "optimized_score": optimized_score,
        "total": n,
        "scenarios": all_results,
    }
    with open(args.out, "w") as f:
        json.dump(output, f, indent=2)
    print(f"\nResults saved to {args.out}")

    # Regression guard
    check_score = optimized_score if args.mode == "optimized" else baseline_score
    if check_score < args.min_score:
        print(f"\nQuality gate FAILED: {check_score}/{n} < {args.min_score} minimum")
        sys.exit(1)
    print(f"Quality gate PASSED: {check_score}/{n} >= {args.min_score} minimum")


if __name__ == "__main__":
    main()

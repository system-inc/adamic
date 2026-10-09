"""Replay the exact audit mutations against only the requested test leaves."""
import json
from pathlib import Path
import re
import subprocess

evidence = Path("review/compiler/enum-guards")
names = [
    "TestFlagEnumsDomain", "TestEnumNeverDefault", "TestFlagEnumLiteralSpellings",
    "TestFlagEnumMemberAliases", "TestEnumNameEnumeration",
    "TestFlagEnumInlineIteration", "TestNumericEnumsAreOpen",
]
results = []


def run_test(label, selected):
    command = ["go", "test", "./internal/lower", "-run",
               "^(" + "|".join(selected) + ")$", "-count=1", "-v", "-timeout", "90s"]
    with (evidence / (label + ".log")).open("w") as output:
        result = subprocess.run(command, stdout=output, stderr=subprocess.STDOUT, timeout=120)
    log = (evidence / (label + ".log")).read_text()
    failed = re.findall(r"--- FAIL: (\w+) ", log)
    results.append({"label": label, "command": command, "exit": result.returncode, "failed": failed})
    (evidence / "replay-results.json").write_text(json.dumps(results, indent=2) + "\n")
    print(label, "exit", result.returncode, "failed", ", ".join(failed), flush=True)
    if result.returncode != 1 or not set(failed).intersection(names):
        raise RuntimeError(label + " was not caught by a strengthened lowering assertion")
    if "panic:" in log or "[build failed]" in log or "test timed out" in log:
        raise RuntimeError(label + " failed for an unintended reason")


for mutant in ["M01", "M02", "M04", "M14", "M15", "M16", "P_LOWER"]:
    patch = str(evidence / (mutant + ".diff"))
    subprocess.run(["git", "apply", patch], check=True, timeout=10)
    try:
        if mutant == "P_LOWER":
            for name in names:
                run_test(mutant + "-" + name, [name])
        else:
            selected = names.copy()
            if mutant in ["M01", "M02"]:
                selected.append("TestEnumMemberValuesAndReverseNameMatchNode")
            run_test(mutant, selected)
    finally:
        subprocess.run(["git", "apply", "-R", patch], check=True, timeout=10)

"""Separate observable disagreements from mutants with no JavaScript effect."""
import json
from pathlib import Path
import re
import subprocess

root = Path("review/compiler/enum-guards")
evidence = root / "behavior"
names = ["TestFlagEnumsDomain", "TestEnumNeverDefault", "TestFlagEnumLiteralSpellings",
         "TestFlagEnumMemberAliases", "TestEnumNameEnumeration",
         "TestFlagEnumInlineIteration", "TestNumericEnumsAreOpen"]
results = []


def run(label, selected, expected):
    command = ["go", "test", "./internal/lower", "-run", "^(" + "|".join(selected) + ")$",
               "-count=1", "-v", "-timeout", "90s"]
    with (evidence / (label + ".log")).open("w") as output:
        result = subprocess.run(command, stdout=output, stderr=subprocess.STDOUT, timeout=120)
    log = (evidence / (label + ".log")).read_text()
    failed = re.findall(r"--- FAIL: (\w+) ", log)
    results.append({"label": label, "command": command, "exit": result.returncode, "failed": failed})
    (evidence / "results.json").write_text(json.dumps(results, indent=2) + "\n")
    print(label, "exit", result.returncode, "failed", ", ".join(failed), flush=True)
    if result.returncode != expected or (expected == 1 and not failed):
        raise RuntimeError(label + " had an unexpected result")
    if "panic:" in log or "[build failed]" in log or "test timed out" in log:
        raise RuntimeError(label + " failed for an unintended reason")


for mutant in ["M01", "M02", "M04", "M14", "M15", "M16", "P_LOWER"]:
    patch = str(root / (mutant + ".diff"))
    subprocess.run(["git", "apply", patch], check=True, timeout=10)
    try:
        if mutant == "P_LOWER":
            for name in names:
                run(mutant + "-" + name, [name], 1)
        else:
            run(mutant + "-behavior", names, 1 if mutant in ["M01", "M02"] else 0)
            if mutant in ["M01", "M02"]:
                run(mutant + "-both-backends", ["TestEnumMemberValuesAndReverseNameMatchNode"], 1)
            elif mutant == "M04":
                run(mutant + "-targeted", ["TestEnumReverseMappingUsesSingleSlot"], 1)
            else:
                run(mutant + "-targeted", ["TestEnumFlagProofsWithoutObservableLoweringEffect"], 1)
    finally:
        subprocess.run(["git", "apply", "-R", patch], check=True, timeout=10)

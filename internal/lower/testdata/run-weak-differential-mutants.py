#!/usr/bin/env python3
"""Revert each Weak representation fix independently and require its Node proof to fail."""
import json
import os
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[3]
EXPRESSION = ROOT / "internal/lower/expression.go"
CALLBACK = ROOT / "internal/lower/weak_representations.go"
EVIDENCE = ROOT / "review/compiler/lowering-chain/source-member-12"
METHOD = """				if inside := l.checker.GetPropertyOfType(from, viewed.Name); inside != nil && !l.sameWeakMethodSlots(l.checker.GetTypeOfSymbol(inside), l.checker.GetTypeOfSymbol(viewed)) {
					return false
				}
"""
FILTER = 'return l.notYet(node, callee.Name().Text()+" callback over narrowed Weak elements whose parameter needs handle-to-target conversion; use a loop with an explicit narrowed copy")'
MUTANTS = [
    ("method-views", EXPRESSION, METHOD, "", "^TestWeakDifferential(ViewMethodParameter|ViewMethod)$",
     ["both backends unexpectedly admit the pinned view NotYet", 'release: exit=0 stdout="t1 none', "release: stdout differs", "AddressSanitizer"]),
    ("union-liveness", EXPRESSION, "\t\tpresent = !l.includesUndefined(read)\n", "", "^TestWeakDifferentialUnionNarrowed$",
     ["release: exit codes differ", "sanitized: exit codes differ", "runtime error: member access within null pointer"]),
    ("weak-callback", CALLBACK, FILTER, "return nil", "^TestWeakDifferential(Each|Find|Map|Reduce|Some|Sort)$",
     ["both backends unexpectedly admit the pinned callback NotYet", "AddressSanitizer: heap-buffer-overflow", "sanitized: exit codes differ"]),
]


def main():
    environment = dict(os.environ, ADAMIC_GATE_UNCACHED="1")
    results = []
    EVIDENCE.mkdir(parents=True, exist_ok=True)
    for name, source, before, after, selected, markers in MUTANTS:
        original = source.read_text()
        if original.count(before) != 1:
            raise SystemExit(f"{name}: expected one mutation site, found {original.count(before)}")
        log = EVIDENCE / ("mutant-" + name + ".log")
        try:
            source.write_text(original.replace(before, after, 1))
            with log.open("w") as output:
                run = subprocess.run(["go", "test", "./internal/oracle", "-run", selected,
                                      "-count=1", "-v", "-timeout", "5m"], cwd=ROOT,
                                     env=environment, stdout=output, stderr=subprocess.STDOUT,
                                     timeout=360)
        finally:
            source.write_text(original)
        text = log.read_text()
        caught = run.returncode != 0 and all(marker in text for marker in markers)
        results.append({"mutant": name, "exit": run.returncode, "caught": caught,
                        "markers": markers, "log": log.name, "tests": selected})
        print(f"{name}: exit={run.returncode}, caught={caught}", flush=True)
        if not caught:
            raise SystemExit(f"{name}: survived or failed outside its intended runtime/refusal proof; see {log}")
    (EVIDENCE / "mutants.json").write_text(json.dumps(results, indent=2) + "\n")


if __name__ == "__main__":
    main()

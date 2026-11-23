#!/usr/bin/env python3
"""Put back the fail-open Adamic named for the advisory checklist (#v3q9ayd, Nov 23 14:33Z) and require its sealed test to
go red: a gate with no evidence dropped from the report.

The mutant is put into a copy of cloud/ in a scratch directory, so the tree is never written. It's caught only when its
test fails on an assertion (`FAIL: <test>`), never on an error: a mutant that breaks the module proves nothing about the
check. Exit 0 when it's caught, 1 when it escapes or fails for the wrong reason. The authority rounds' 36 mutants are in
Loom's feed (#v3q9ayd, through J-ke9nhtw); the advisory script carries the one its ruling asked for.
"""

from pathlib import Path
import shutil
import subprocess
import sys
import tempfile

source = Path(__file__).resolve().parent
script, compare = "submit-gates.py", "gates/units-compare.py"
tests = "submit_gates_test.SubmitGatesTests."

# (name, file, [(old, new), ...], [tests that must go red])
mutants = [
    ("a gate with no evidence dropped from the report", script, [
        ('    report["missing"] = [entry["name"] for entry in report["gates"] + report["rows"]',
         '    report["missing"] = [entry["name"] for entry in report["rows"]'),
    ], [tests + "test_a_missing_gate_is_reported_missing"]),
]


def copy(scratch):
    shutil.copy(source / script, scratch / script)
    shutil.copy(source / "submit_gates_test.py", scratch / "submit_gates_test.py")
    shutil.copytree(source / "gates", scratch / "gates")


def main():
    scratch = Path(tempfile.mkdtemp(prefix="submit-gates-mutants-"))
    escaped = 0
    for number, (name, path, edits, named) in enumerate(mutants):
        tree = scratch / f"mutant-{number}"
        tree.mkdir()
        copy(tree)
        text = (tree / path).read_text()
        for old, new in edits:
            if text.count(old) != 1:
                print(f"FAIL {name}: its original code isn't once in {path}: {old[:70]!r}")
                return 1
            text = text.replace(old, new)
        (tree / path).write_text(text)
        result = subprocess.run([sys.executable, "-B", "-m", "unittest", *named], cwd=tree,
                                capture_output=True, text=True)
        output = result.stdout + result.stderr
        missed = [test for test in named if f"FAIL: {test.rsplit('.', 1)[-1]} " not in output]
        if result.returncode == 0 or missed or "ERROR: " in output:
            escaped += 1
            print(f"FAIL {name}: escaped or failed for the wrong reason (exit {result.returncode}); "
                  f"not red: {', '.join(missed) or 'none'}\n{output[-1500:]}")
        else:
            print(f"PASS {name}: caught by {', '.join(test.rsplit('.', 1)[-1] for test in named)}")
    shutil.rmtree(scratch)
    print(f"mutants: {len(mutants) - escaped} of {len(mutants)} caught")
    return 1 if escaped else 0


if __name__ == "__main__":
    sys.exit(main())

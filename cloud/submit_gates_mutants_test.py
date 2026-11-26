#!/usr/bin/env python3
"""Require the advisory checklist's fail-open mutant and the two steps.sh summary mutants to go red on their named tests.

The advisory ruling (#v3q9ayd, Nov 23 14:33Z) asked for the missing-gate mutant, caught by
test_a_missing_gate_is_reported_missing. The package-skip mutant goes red on
test_latent_whole_rejects_a_package_with_a_skipped_test. The dropped-failure mutant goes red on
test_latent_whole_rejects_a_failing_test.

Each mutant is put into a copy of cloud/ in a scratch directory, so the tree is never written. A mutant is caught only when its
test fails on an assertion (`FAIL: <test>`), never on an error: a mutant that breaks the module proves nothing about the
check. It's a test itself, run as `python3 -B -m unittest -v submit_gates_mutants_test` from cloud/, red when the mutant
escapes or fails for the wrong reason. A test by name, as the Go mutants beside their packages are, since a script that
isn't one and runs submit_gates_test.py is gate logic at the queue's door. The authority rounds' 36 mutants are in
Loom's feed (#v3q9ayd, through J-ke9nhtw).
"""

from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest

source = Path(__file__).resolve().parent
script, compare = "submit-gates.py", "gates/units-compare.py"
tests = "submit_gates_test.SubmitGatesTests."

# (name, file, [(old, new), ...], [tests that must go red])
mutants = [
    ("a gate with no evidence dropped from the report", script, [
        ('    report["missing"] = [entry["name"] for entry in report["gates"] + report["rows"]',
         '    report["missing"] = [entry["name"] for entry in report["rows"]'),
    ], [tests + "test_a_missing_gate_is_reported_missing"]),
    ("a package skip accepted with test results", "gates/steps.sh", [
        ("(action == 'skip' and not counted)", "action == 'skip'"),
    ], ["submit_gates_test.GoTestSummaryTests.test_latent_whole_rejects_a_package_with_a_skipped_test"]),
    ("a failing test dropped from the summary verdict", "gates/steps.sh", [
        ("passed and not failed", "passed"),
    ], ["submit_gates_test.GoTestSummaryTests.test_latent_whole_rejects_a_failing_test"]),
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


class SubmitGatesMutants(unittest.TestCase):
    def test_each_mutant_turns_its_named_test_red(self):
        self.assertEqual(main(), 0)


if __name__ == "__main__":
    unittest.main()

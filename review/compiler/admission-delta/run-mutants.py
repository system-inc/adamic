#!/usr/bin/env python3
"""Mutants of the admission gate itself. Always restore the original source."""
import pathlib
import subprocess
import difflib

source = pathlib.Path('cmd/adamic-admission-delta/main.go')
original = source.read_text()
evidence = pathlib.Path('review/compiler/admission-delta')
mutants = [
 ('crash-as-refusal', 'return "error"', 'return "refused"', 'TestClassification'),
 ('ignore-output', 'return a.Error == "" &&', 'return true || a.Error == "" &&', 'TestOutputsAgree|TestNewAdmissionMismatch'),
 ('omit-witness', 'witnesses = append(witnesses, i)', 'others = append(others, i)', 'TestSampling|TestNewAdmissionMismatch'),
 ('ignore-blob', 'if blob != p.Blob {', 'if false && blob != p.Blob {', 'TestManifestMismatch'),
 ('dirty-input', 'execute(headTree, *limit,', 'execute(root, *limit,', 'TestHeadInputIsolation'),
]
for name, before, after, tests in mutants:
    assert before in original, name
    mutated = original.replace(before, after)
    (evidence / (name + '.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True), mutated.splitlines(True), fromfile=str(source), tofile=str(source))))
    try:
        source.write_text(mutated)
        with (evidence / (name + '.log')).open('w') as log:
            result = subprocess.run(['go', 'test', './cmd/adamic-admission-delta', '-run', tests, '-count=1', '-timeout', '60s'], stdout=log, stderr=subprocess.STDOUT, timeout=90)
        if result.returncode == 0:
            raise RuntimeError(name + ' survived')
        print(name + ': caught, exit ' + str(result.returncode), flush=True)
    finally:
        source.write_text(original)

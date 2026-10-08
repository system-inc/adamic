#!/usr/bin/env python3
"""Independently mutate lane-owned helper guards and pin semantic test failures."""
import os
import argparse
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
logs = Path('/tmp/views-callables-mutants')
logs.mkdir(exist_ok=True)
native = root / 'internal/native/runtime/view_callables_contract.h'
js = root / 'internal/javascript/view_callables_contract.go'
lower = root / 'internal/lower/view_callables_contract.go'
mutants = [
 ('lower-skip-descriptor', lower, 'contract.Parameters, contract.Result = parameters, result', '_ = parameters; _ = result', 'lower', ''),
 ('lower-accept-generic', lower, 'len(signature.TypeParameters()) != 0 || ', '', 'lower', ''),
 ('lower-accept-rest', lower, 'signature.HasRestParameter() || signature.MinArgumentCount() != len(signature.Parameters())', 'false', 'lower', '.*values.*'),
 ('lower-accept-overload', lower, 'len(signatures) != 1', 'len(signatures) == 0', 'lower', ''),
 ('lower-accept-optional', lower, ' || signature.MinArgumentCount() != len(signature.Parameters())', '', 'lower', ''),
 ('lower-accept-unknown-parameter', lower, 'child == 0', 'false', 'lower', ''),
 ('lower-accept-unknown-result', lower, 'result == 0', 'false', 'lower', ''),
 ('native-skip-shape', native, '    if (optional && value == NULL)', '    if (value != NULL) { return value; }\n    if (optional && value == NULL)', 'native', 'wrong-arity'),
 ('native-accept-arity', native, 'recorded->arity != expected->arity', 'false', 'native', 'wrong-arity'),
 ('native-accept-number', native, 'value->kind == adamic_kind_closure && recorded', 'recorded', 'native', 'number'),
 ('native-accept-parameter', native, 'recorded->parameters[index] == expected->parameters[index]', 'true', 'native', 'wrong-parameter'),
 ('native-accept-result', native, 'recorded->result != expected->result', 'false', 'native', 'wrong-result'),
 ('native-accept-unknown-result', native, 'recorded->result == 0 || expected->result == 0', 'false', 'native', 'unknown-result'),
 ('native-accept-unknown-parameter', native, 'recorded->parameters[index] != 0 && expected->parameters[index] != 0 && ', '', 'native', 'unknown-parameter'),
 ('js-skip-shape', js, '    if (optional && value === undefined)', '    if (value !== undefined) return value;\n    if (optional && value === undefined)', 'javascript', 'wrong-arity'),
 ('js-accept-arity', js, 'recorded.parameters.length !== expected.parameters.length', 'false', 'javascript', 'wrong-arity'),
 ('js-call-number', js, 'if (found === "function")', 'if (found === "function" || found === "number")', 'javascript', 'number'),
 ('js-accept-parameter', js, 'representation === expected.parameters[index]', 'true', 'javascript', 'wrong-parameter'),
 ('js-accept-result', js, 'recorded.result !== expected.result', 'false', 'javascript', 'wrong-result'),
 ('js-accept-unknown-result', js, 'recorded.result === 0 || expected.result === 0', 'false', 'javascript', 'unknown-result'),
 ('js-accept-unknown-parameter', js, 'representation !== 0 && expected.parameters[index] !== 0 && ', '', 'javascript', 'unknown-parameter'),
]
arguments = argparse.ArgumentParser()
arguments.add_argument("--lower-only", action="store_true")
arguments.add_argument("--only")
options = arguments.parse_args()
if options.lower_only:
    mutants = [mutant for mutant in mutants if mutant[4] == "lower"]
if options.only:
    mutants = [mutant for mutant in mutants if mutant[0] == options.only]
    assert mutants, "unknown mutant"
for name, path, before, after, package, probe in mutants:
    original = path.read_text()
    assert original.count(before) == 1, name
    try:
        path.write_text(original.replace(before, after))
        test = {'native': 'TestViewCallableShapeNative', 'javascript': 'TestViewCallableShapeNode', 'lower': 'TestViewCallableShapeContract'}[package]
        command = ['go', 'test', './internal/' + package, '-run', '^' + test + ('/' + probe if probe else '') + '$', '-count=1', '-timeout', '10m']
        with (logs / (name + '.log')).open('w') as log:
            result = subprocess.run(command, cwd=root, env=os.environ, stdout=log, stderr=subprocess.STDOUT)
        output = (logs / (name + '.log')).read_text()
        assert result.returncode != 0 and '--- FAIL: ' + test in output, name + ' survived or failed to build'
        print(name + ': caught by ' + test + '/' + probe, flush=True)
    finally:
        path.write_text(original)
print(str(len(mutants)) + ' helper mutants caught; compiler dispatch mutants remain pending')

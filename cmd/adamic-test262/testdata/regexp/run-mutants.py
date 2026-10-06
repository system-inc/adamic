#!/usr/bin/env python3
"""Run real runner/lowering mutations and restore each source unconditionally."""
import pathlib
import subprocess
import sys

root = pathlib.Path(__file__).resolve().parents[4]
mutants = [
    ('range-end-exclusive', 'cmd/adamic-test262/regexp_prelude.go', 'point <= end', 'point < end', 'TestRegExpRunnerNode', 'serial/parallel mismatch'),
    ('negative-string-condition-inverted', 'cmd/adamic-test262/regexp_prelude.go', 'if (regExp.test(negatives.join(""))) {', 'if (!regExp.test(negatives.join(""))) {', 'TestRegExpRunnerNode', 'serial/parallel mismatch'),
    ('capture-check-omitted', 'cmd/adamic-test262/regexp_prelude.go', 'assertCompareArray<string | undefined>(match, expectedEntries, "Match entries");', '', 'TestRegExpRunnerNode', 'serial/parallel mismatch'),
    ('index-check-omitted', 'cmd/adamic-test262/regexp_prelude.go', 'assertSameValue(match.index, expectedIndex, "Match index");', '', 'TestRegExpRunnerNode', 'serial/parallel mismatch'),
    ('input-check-omitted', 'cmd/adamic-test262/regexp_prelude.go', 'assertSameValue(match.input, expectedInput, "Match input");', '', 'TestRegExpRunnerNode', 'serial/parallel mismatch'),
    ('original-node-guard-bypassed', 'cmd/adamic-test262/run.go', 'base.Kind == outcomePass && test.Original != ""', 'false && base.Kind == outcomePass && test.Original != ""', 'TestRegExpRunnerNode', 'false pass accepted'),
    ('written-let-frozen', 'internal/lower/regexp_constants.go', 'return !written', 'return true', 'TestRegExpNativeRefusals', 'expected loud NotYet'),
    ('clone-flags-lost', 'internal/lower/regexp.go', 'pattern, flags, ok = l.constantRegExp(args[0], 0)', 'pattern, flags, ok = l.constantRegExp(args[0], 0); flags = ""', 'TestRegExpRunnerNode', 'serial/parallel mismatch'),
    ('constructor-identity-guard-bypassed', 'cmd/adamic-test262/run.go', 'base.Kind == outcomePass && test.ConstructorAssertion', 'false && base.Kind == outcomePass && test.ConstructorAssertion', 'TestRegExpRunnerNode', 'wrong constructor accepted'),
    ('omitted-pattern-overloads-removed', 'internal/load/regexp_library.go', r'\n    new (): RegExp;\n    (): RegExp;', '', 'TestRegExpOmittedPattern', 'error TS2554'),
    ('literal-flags-lost', 'cmd/adamic-test262/regexp_adapt.go', 'flags := t.text[last+1:]', 'flags := t.text[last+1:]; if flags == "gy" { flags = "" }', 'TestRegExpRunnerNode', 'serial/parallel mismatch'),
    ('code-point-value-changed', 'cmd/adamic-test262/prelude.go', '\treturn point;', '\treturn point + 1;', 'TestRegExpRunnerNode', 'serial/parallel mismatch'),
    ('non-string-collector-annotated', 'cmd/adamic-test262/regexp_adapt.go', 't[args[0][0]].kind != "str"', 'false', 'TestRegExpAdaptGuards', 'uncertain program adapted'),
    ('flow-undefined-frozen', 'internal/lower/regexp_constants.go', 'l.intrinsicUndefined(node) && ', '', 'TestRegExpNativeRefusals', 'expected loud NotYet'),
]
for name, file, old, new, test, witness in mutants:
    if len(sys.argv) > 1 and name not in sys.argv[1:]:
        continue
    target = root / file
    original = target.read_text()
    if original.count(old) != 1:
        raise SystemExit(f'{name}: mutation site count {original.count(old)}')
    log = pathlib.Path('/tmp') / f'regex-runner-mutant-{name}.log'
    try:
        target.write_text(original.replace(old, new))
        package = './internal/lower' if test == 'TestRegExpNativeRefusals' else './cmd/adamic-test262'
        if test == 'TestRegExpOmittedPattern':
            package = './internal/load'
        with log.open('w') as output:
            result = subprocess.run(['go', 'test', package, '-run', '^'+test+'$', '-count=1', '-v'], cwd=root, stdout=output, stderr=subprocess.STDOUT)
        text = log.read_text()
        failures = [line.strip() for line in text.splitlines() if witness in line]
        if result.returncode == 0 or not failures:
            raise SystemExit(f'{name}: NOT CAUGHT, exit={result.returncode}; {log}')
        if witness == 'serial/parallel mismatch' and 'Fail:1 Refused:0 Crashed:0' not in failures[0]:
            raise SystemExit(f'{name}: not caught by the execution check; {log}')
        print(f'{name}: caught, exit={result.returncode}; {log}', flush=True)
        print(failures[0], flush=True)
    finally:
        target.write_text(original)

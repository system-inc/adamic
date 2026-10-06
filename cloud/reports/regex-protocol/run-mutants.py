#!/usr/bin/env python3
"""Run feature mutants in an isolated checkout and restore every source."""
import os
import pathlib
import subprocess
import sys

root = pathlib.Path(sys.argv[1]).resolve()
if root == pathlib.Path.cwd().resolve():
    raise SystemExit('Use an isolated checkout; the live survey must not see mutants.')
mutants = [
    ('unhandled-diagnostic-admitted', 'internal/lower/regexp_diagnostics.go', 'if l.regexpDiagnosticLeaves(l.result.Main, functions, closures) {', 'if false && l.regexpDiagnosticLeaves(l.result.Main, functions, closures) {', './internal/lower', '^TestRegExpUncaughtDiagnosticsRefuse$', 'unhandled parser diagnostic must refuse'),
    ('syntax-error-is-type-error', 'internal/native/emit_expressions.go', 'adamic_thrown = adamic_builtin_error_new(2,', 'adamic_thrown = adamic_builtin_error_new(1,', './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/regexp_errors', 'stdout differs'),
    ('constructor-is-ancestry', 'internal/native/runtime/exceptions.c', '(!exact && kind == 0)', '(kind == 0)', './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/regexp_errors', 'stdout differs'),
    ('non-global-matchall-repeats', 'internal/native/runtime/regexp.c', 'else if (!(regex_program(regex)->flags & 8))', 'else if (false)', './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/regexp_protocol', 'stdout differs'),
    ('search-loses-negative-zero', 'internal/native/runtime/regexp.c', 'regex->slots[1].number = previous;', 'regex->slots[1].number = 0;', './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/regexp_protocol', 'stdout differs'),
    ('named-indices-omitted', 'internal/native/runtime/regexp.c', 'indices->properties->slots[2].reference = regex_groups(p, indices);', 'indices->properties->slots[2].reference = NULL;', './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/regexp_indices', 'stdout differs'),
    ('dynamic-let-frozen', 'internal/lower/regexp_constants.go', 'return !written', 'return true', './internal/lower', '^TestRegExpNativeRefusals$', 'expected loud NotYet'),
    ('harness-constructor-check-omitted', 'cmd/adamic-test262/prelude.go', 'if (!correct)', 'if (false)', './cmd/adamic-test262', '^TestRegExpRunnerNode$', 'wrong constructor accepted'),
    ('symbol-match-returns-null', 'internal/native/runtime/regexp.c', 'if (!(p->flags & 8))\n\t\treturn adamic_regex_exec(regex, input);', 'if (!(p->flags & 8))\n\t\treturn NULL;', './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/regexp_protocol', 'stdout differs'),
    ('symbol-split-drops-captures', 'internal/native/runtime/regexp.c', 'for (size_t k = 1; k <= p->captures; k++)', 'for (size_t k = 1; k < 1; k++)', './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/regexp_protocol', 'stdout differs'),
    ('symbol-replace-literal-named-group', 'internal/native/runtime/regexp.c', "case '<': {\n\t\t\tif (p->group_count == 0)", "case '<': {\n\t\t\tif (true)", './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/regexp_protocol', 'stdout differs'),
    ('replacement-callback-resets-state', 'internal/native/runtime/regexp.c', 'adamic_string *whole = match->elements[0].reference;', 'adamic_string *whole = match->elements[0].reference; regex->slots[1].number = 0;', './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/regexp_replacement_callback', 'stdout differs'),
    ('syntax-error-lacks-exception-edge', 'internal/flow/build.go', 'throws = throws || value.Interface().(ir.RegExpNew).Invalid', 'throws = throws || false', './internal/flow', '^TestRegExpProtocolExceptionEdges$', 'missing or spurious exception edge'),
    ('regexp-string-drops-flags', 'internal/native/runtime/regexp.c', '&slash, regex->slots[2].reference, &slash, regex->slots[3].reference', '&slash, regex->slots[2].reference, &slash, &adamic_string_empty', './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/regexp_string', 'stdout differs'),
]
for name, file, old, new, package, test, witness in mutants:
    if len(sys.argv) > 2 and name not in sys.argv[2:]:
        continue
    target = root / file
    original = target.read_text()
    if original.count(old) != 1:
        raise SystemExit(f'{name}: mutation site count {original.count(old)}')
    log = pathlib.Path('/tmp') / f'regex-protocol-mutant-{name}.log'
    try:
        target.write_text(original.replace(old, new))
        with log.open('w') as output:
            result = subprocess.run(['go', 'test', package, '-run', test, '-count=1', '-v', '-timeout', '10m'], cwd=root, env=dict(os.environ, GOFLAGS='-buildvcs=false'), stdout=output, stderr=subprocess.STDOUT)
        failures = [line.strip() for line in log.read_text().splitlines() if witness in line]
        if result.returncode == 0 or not failures:
            raise SystemExit(f'{name}: NOT CAUGHT, exit={result.returncode}; {log}')
        print(f'{name}: caught, exit={result.returncode}; {log}', flush=True)
        print(failures[0], flush=True)
    finally:
        target.write_text(original)

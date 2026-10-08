#!/usr/bin/env python3
"""Each mutant runs alone, logs its failure, and restores the exact source bytes."""
from pathlib import Path
import subprocess
import os

repository = Path(__file__).resolve().parents[3]
logs = Path('/tmp/method-values-mutants')
logs.mkdir(exist_ok=True)
mutants = [
    ('bound-extraction', 'internal/native/emit_expressions.go',
     '\t\t\treturn value\n\t\t}\n\t\tif taken, ok := e.take(expression);',
     '\t\t\treturn e.own(ir.Closure, fmt.Sprintf("adamic_method_bind(%s, %s)", value, object))\n\t\t}\n\t\tif taken, ok := e.take(expression);',
     './internal/oracle', 'TestMethodValuesTypeScript/(reading|uncaught)', ['stdout differs', 'exit code differs']),
    ('missing-this-check', 'internal/lower/method_values.go',
     'if !extracted {', 'if extracted || !extracted {',
     './internal/oracle', 'TestMethodValuesTypeScript/(reading|uncaught)', ['runtime error:', 'AddressSanitizer']),
    ('receiver-without-retain', 'internal/native/runtime/closure.c',
     'bound->bound = adamic_retain(receiver);', 'bound->bound = receiver;',
     './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/method_values/bind', ['heap-use-after-free']),
    ('stale-unbound-cache', 'internal/native/runtime/heap.c',
     'closure->original->unbound = NULL;', '(void)closure;',
     './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/method_values/free', ['heap-use-after-free']),
    ('bound-cycle-hidden', 'internal/lower/cycles.go',
     'for _, bound := range f.boundMethods {', 'for _, bound := range f.boundMethods[:0] {',
     './internal/lower', 'TestMethodBindCycleIsRefused', ['want bound receiver cycle refusal']),
    ('static-throw-hidden', 'internal/lower/exceptions.go',
     'for _, lowered := range l.statics {', 'for _, lowered := range map[*ast.Symbol]*instance{} {',
     './internal/oracle', 'TestMethodValuesTypeScript/static_reading', ['stdout differs']),
    ('static-own-hidden', 'internal/native/runtime/object.c',
     'methods->own_static[index] && key[0]', 'false && methods->own_static[index] && key[0]',
     './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/method_values/static', ['stdout differs']),
    ('optional-receiver-guard', 'internal/native/emit_expressions.go',
     'e.line("if (%s != NULL) {", object)', 'e.line("if (%s == NULL) {", object)',
     './internal/oracle', 'TestMethodValuesTypeScript/optional', ['runtime error:', 'AddressSanitizer']),
    ('optional-member-hidden', 'internal/native/runtime/closure.c',
     'if (!found) { return NULL; }', 'if (found) { return NULL; }',
     './internal/oracle', 'TestMethodValuesTypeScript/optional', ['exit codes differ']),
    ('rest-reference-without-retain', 'internal/native/runtime/closure.c',
     'if (references) { adamic_retain(value.reference); }', 'if (references) { (void)value.reference; }',
     './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/method_values/rest', ['heap-use-after-free']),
    ('rest-spread-without-owner', 'internal/native/method_values.go',
     'adamic_retain(%s->elements[%s].reference)', '(%s->elements[%s].reference)',
     './internal/oracle', 'TestNativeAgreesWithNode/internal/oracle/testdata/method_values/rest', ['heap-use-after-free']),
]
for name, relative, original, replacement, package, test, catchers in mutants:
    selected = os.environ.get("METHOD_VALUES_MUTANTS")
    if selected and name not in selected.split(","):
        continue
    source = repository / relative
    saved = source.read_bytes()
    text = saved.decode()
    if text.count(original) != 1:
        raise RuntimeError(f'{name}: mutation point must be unique')
    try:
        source.write_text(text.replace(original, replacement))
        log = logs / (name + '.log')
        with log.open('w') as output:
            result = subprocess.run(
                ['bash', '-lc', 'source /workspace/adamic-tools/env.sh && '
                 f"ADAMIC_GATE_UNCACHED=1 go test {package} -run '{test}' -count=1 -timeout 10m"],
                cwd=repository, stdout=output, stderr=subprocess.STDOUT)
        report = log.read_text()
        if result.returncode == 0 or not any(catcher in report for catcher in catchers):
            raise RuntimeError(f'{name}: not caught by the intended check; see {log}')
        print(f'{name}: caught, exit {result.returncode}; {log}', flush=True)
    finally:
        source.write_bytes(saved)

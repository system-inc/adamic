#!/usr/bin/env python3
"""Run independent compiler mutations, restoring each before the next."""
import json
import os
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parents[2]
LOGS = Path('/tmp/optional-indexing-mutants')
LOGS.mkdir(exist_ok=True)
results = []


def run(name, relative, change, fixture, catcher, package="oracle", selector=None):
    path = ROOT / relative
    original = path.read_text()
    mutated = change(original)
    assert mutated != original, name
    command = ['go', 'test', './internal/' + package, '-run',
               selector or '^TestNativeAgreesWithNode/internal/oracle/testdata/' + fixture + '$',
               '-count=1']
    try:
        path.write_text(mutated)
        with (LOGS / (name + '.log')).open('w') as log:
            result = subprocess.run(command, cwd=ROOT, env={**os.environ, 'ADAMIC_GATE_UNCACHED': '1'}, stdout=log, stderr=subprocess.STDOUT)
    finally:
        path.write_text(original)
    output = (LOGS / (name + '.log')).read_text()
    assert result.returncode == 1 and '--- FAIL:' in output and '[build failed]' not in output, output
    assert catcher in output, output
    results.append({'name': name, 'command': command, 'exit': result.returncode,
                    'catcher': catcher, 'log': str(LOGS / (name + '.log'))})
    print(name + ': caught by ' + catcher, flush=True)


def absent_index(text):
    needle = 'WhenNot: fit(ir.Undefined{}, of)'
    assert text.count(needle) == 1
    return text.replace(needle, 'WhenNot: ir.Effects{Body: []ir.Statement{ir.Evaluate{Value: value.(ir.ArrayIndex).Index}}, Result: fit(ir.Undefined{}, of)}')


run('evaluate-absent-index', 'internal/lower/optional_indexing.go', absent_index,
    'optional_indexing_array.a', 'stdout differs')


def unchecked_slot(text):
    needle = 'e.line("adamic_value *%s = %s(%s, %s);", slot, lookup, array, index)'
    assert text.count(needle) == 1
    return text.replace(needle, '_ = lookup\n\te.line("adamic_value *%s = &%s->elements[(size_t)%s];", slot, array, index)')


run('omit-array-range-check', 'internal/native/emit_slots.go', unchecked_slot,
    'optional_indexing_array.a', 'AddressSanitizer: heap-buffer-overflow')


def absent_chain_index(text):
    needle = 'WhenNot: fit(ir.Undefined{}, of)'
    assert text.count(needle) == 1
    return text.replace(needle, 'WhenNot: ir.Effects{Body: []ir.Statement{ir.Evaluate{Value: indexed.(ir.ArrayIndex).Index}}, Result: fit(ir.Undefined{}, of)}')


run('evaluate-absent-chain-index', 'internal/lower/optional_indexing_chain.go', absent_chain_index,
    'optional_indexing_chain.a', 'exit codes differ')


def omit_null(text):
    needle = 'ir.Binary{Operator: ir.Or, Left: ir.IsUndefined{Value: read}, Right: ir.IsNull{Value: read}}'
    assert text.count(needle) == 1
    return text.replace(needle, 'ir.IsUndefined{Value: read}')


run('omit-null-guard', 'internal/lower/optional_indexing_map.go', omit_null,
    'optional_indexing_map.a', 'JavaScript backend: exit codes differ')


def reselect(text):
    needle = 'l.optionalIndexValue(node, read)'
    assert text.count(needle) == 1
    return text.replace(needle, 'l.optionalIndexValue(node, base)')


run('reselect-base', 'internal/lower/optional_indexing.go', reselect,
    'optional_indexing_array.a', 'stdout differs')
run('omit-index-null-guard', 'internal/lower/optional_indexing.go', omit_null,
    'optional_indexing_array.a', 'JavaScript backend:')


def absent_map_key(text):
    needle = 'WhenNot: fit(ir.Undefined{}, of)'
    assert text.count(needle) == 1
    return text.replace(needle, 'WhenNot: ir.Effects{Body: []ir.Statement{ir.Evaluate{Value: key}}, Result: fit(ir.Undefined{}, of)}')


run('evaluate-absent-map-key', 'internal/lower/optional_indexing_map.go', absent_map_key,
    'optional_indexing_map.a', 'stdout differs')


def typed_index_continuation(text):
    needle = "\t\tif access.QuestionDotToken == nil && node.Flags&ast.NodeFlagsOptionalChain != 0 {\n\t\t\tvalue, err := l.optionalIndexContinuation(node)\n\t\t\treturn value, true, err\n\t\t}\n"
    assert text.count(needle) == 1
    return text.replace(needle, '')


run('omit-typed-index-continuation', 'internal/lower/typed_arrays.go', typed_index_continuation,
    'optional_indexing_typed_array.a', 'UndefinedBehaviorSanitizer')


def typed_property_boundary(text):
    needle = "\t\tif access.QuestionDotToken == nil && node.Flags&ast.NodeFlagsOptionalChain != 0 {\n\t\t\treturn nil, true, l.notYet(node, \"an optional chain longer than one step\")\n\t\t}\n"
    assert text.count(needle) == 1
    return text.replace(needle, '')


run('omit-typed-property-boundary', 'internal/lower/typed_arrays.go', typed_property_boundary,
    '', 'want NotYet containing', package='lower', selector='^TestOptionalIndexingKeepsUnsupportedStorageNotYet/typed_ordinary_property_continuation$')

def structural_map_boundary(text):
    needle = '\tif l.optionalMapStructuralReceiver(access.Expression) {\n\t\treturn nil, true, l.notYet(node, "an optional Map get through a structural receiver")\n\t}\n'
    assert text.count(needle) == 1
    return text.replace(needle, '')

run('omit-structural-map-boundary', 'internal/lower/optional_indexing_map.go', structural_map_boundary,
    '', 'want structural Map receiver boundary', package='lower', selector='^TestOptionalIndexingMapShapeRefused$')

def numeric_index_boundary(text):
    needle = '\tif index.Type() != ir.Number {\n\t\treturn nil, l.notYet(node, "an optional index that isn\'t a number")\n\t}\n'
    assert text.count(needle) == 1
    return text.replace(needle, '')

run('omit-numeric-index-boundary', 'internal/lower/optional_indexing.go', numeric_index_boundary,
    '', 'want NotYet containing', package='lower', selector='^TestOptionalIndexingKeepsUnsupportedStorageNotYet/string_numeric_key$')

(ROOT / 'review/optional-indexing/mutants.json').write_text(json.dumps(results, indent=2) + '\n')

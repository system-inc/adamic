#!/usr/bin/env python3
"""Compare owned numeric listener declarations with pinned production Go inputs."""
from pathlib import Path
import json
import hashlib
import re

ROOT = Path(__file__).resolve().parents[4]
OWNED = ROOT / 'stage1/cohere/typeaware'
RULES = {
    'no_import_assign': 'core/no_import_assign',
    'prefer_exponentiation_operator': 'core/prefer_exponentiation_operator',
    'use_isnan': 'core/use_isnan',
    'no_global_listener_target_assertion': 'nexus/correctness_no_global_listener_target_assertion',
    'no_leaked_number_render': 'nexus/correctness_no_leaked_number_render',
    'no_mock_on_module_namespace': 'nexus/correctness_no_mock_on_module_namespace',
    'no_interpolated_shell_command': 'nexus/security_no_interpolated_shell_command',
    'no_interpolated_sql_string': 'nexus/security_no_interpolated_sql_string',
    'no_alert': 'core/no_alert',
    'no_new_func': 'core/no_new_func',
    'no_new_native_nonconstructor': 'core/no_new_native_nonconstructor',
    'no_new_wrappers': 'core/no_new_wrappers',
    'no_throw_literal': 'core/no_throw_literal',
    'no_useless_backreference': 'core/no_useless_backreference',
    'prefer_arrow_callback': 'core/prefer_arrow_callback',
}


def declaration(text):
    matches = re.findall(
        r'export function listenerKinds\(\): readonly number\[\] \{\s*'
        r'return \[([0-9, ]+)\];\s*\}', text,
    )
    if len(matches) != 1:
        raise AssertionError('one numeric-only listener declaration required')
    numbers = [int(n.strip()) for n in matches[0].split(',')]
    if numbers != sorted(set(numbers)):
        raise AssertionError('listener kinds must be unique and ordered')
    return numbers


def compare(text, expected):
    actual = declaration(text)
    if actual != expected:
        raise AssertionError(f'Go kinds {expected}, Adamic kinds {actual}')


enum = (ROOT / 'cohere/TypeScript/tsc/internal/ast/kind_generated.go').read_text()
body = enum.split('const (', 1)[1].split('\n)', 1)[0]
values = {}
for line in body.splitlines():
    match = re.fullmatch(r'\s*(Kind\w+)(?: Kind = iota)?\s*(?://.*)?', line)
    if match:
        values[match[1]] = len(values)
    elif re.match(r'\s*Kind\w+\s*=', line):
        break  # The enum ends before its range aliases.
if values.get('KindUnknown') != 0 or 'KindCount' not in values:
    raise AssertionError('unrecognized pinned parser enum')

results = []
for name, production in RULES.items():
    go = (ROOT / f'cohere/internal/lint/rules/{production}.go').read_text()
    keys = re.findall(r'^\t\t\tast\.(Kind\w+):\s*(?:func\(|reportIfFunctionConstructor)', go, re.M)
    if not keys:
        raise AssertionError(f'{name}: production listener keys not found')
    expected = sorted({values[key] for key in keys})
    source = (OWNED / f'{name}.a').read_text()
    compare(source, expected)
    # Change only a numeric declaration, keeping its syntax valid. This static
    # witness is additional to the existing compiled semantic rule mutants.
    replacement = list(expected)
    replacement[-1] += 1
    original = ', '.join(map(str, expected))
    mutated = source.replace(f'return [{original}];',
                             'return [' + ', '.join(map(str, replacement)) + '];', 1)
    if mutated == source:
        raise AssertionError(f'{name}: declaration mutation did not apply')
    try:
        compare(mutated, expected)
    except AssertionError as caught:
        print(f'PASS {name}: {expected}; numeric-kind mutant caught: {caught}')
    else:
        raise AssertionError(f'{name}: numeric-kind mutant survived')
    results.append({'module': name, 'go_listener_kinds': keys,
                    'numeric_kinds': expected, 'mutant_kinds': replacement,
                    'mutant_caught': True,
                    'go_source_sha256': hashlib.sha256(go.encode()).hexdigest(),
                    'adamic_source_sha256': hashlib.sha256(source.encode()).hexdigest(),
                    'enum_sha256': hashlib.sha256(enum.encode()).hexdigest()})
Path(__file__).with_name('results.json').write_text(json.dumps(results, indent=2) + '\n')
print(f'PASS {len(results)} declarations and {len(results)} static numeric-kind mutants')
Path(__file__).with_name('expected.stdout').write_text(''.join(
    r['module'] + '\t' + ','.join(map(str, r['numeric_kinds'])) + '\n'
    for r in results))

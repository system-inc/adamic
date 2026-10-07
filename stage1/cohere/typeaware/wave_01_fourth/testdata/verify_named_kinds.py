"""Compare owned manifests with unchanged Go rule listener registrations."""
import json
from pathlib import Path
import re

repo = Path(__file__).resolve().parents[5]
root = repo / 'stage1/cohere/typeaware'
pairs = [
 ('wave_01_third/rules/no_implicit_return', 'nexus/correctness_no_implicit_return.go'),
 ('wave_01_third/rules/no_deprecated', 'typescript/no_deprecated.go'),
 ('wave_01_third/rules/no_else_return', 'core/no_else_return.go'),
 ('wave_01_third/rules/child_process_error_listener', 'nexus/correctness_require_child_process_error_listener.go'),
 ('wave_01_third/rules/response_status_check', 'nexus/correctness_require_response_status_check.go'),
 ('wave_01_third/rules/independent_await_in_loop', 'nexus/performance_no_independent_await_in_loop.go'),
 ('wave_01_fourth/symbol_description', 'core/symbol_description.go'),
 ('wave_01_fourth/require_atomic_updates', 'core/require_atomic_updates.go'),
 ('wave_01_fourth/require_await', 'core/require_await.go'),
]
for directory, production in pairs:
    text = (repo / 'cohere/internal/lint/rules' / production).read_text()
    expected = re.findall(r'^\t{3}ast.Kind(\w+):', text, re.M)
    assert expected, production
    actual = json.loads((root / directory / 'rule.json').read_text())['kinds']
    def compare(kinds):
        assert kinds == expected, (directory, kinds, expected)
    compare(actual)
    mutant = actual.copy()
    mutant[0] = 'CallExpression' if actual[0] != 'CallExpression' else 'Identifier'
    try:
        compare(mutant)
    except AssertionError:
        pass
    else:
        raise AssertionError('wrong listener kind survived: ' + directory)
    print(directory, 'PASS named registration; wrong-kind mutant caught')

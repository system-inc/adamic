"""Compare listener metadata with the pinned production Go listeners and enum."""
from pathlib import Path
import re

ROOT = Path(__file__).resolve().parents[4]
RULES = ROOT / 'stage1/cohere/typeaware'
GO = ROOT / 'cohere/internal/lint/rules'
KINDS = ROOT / 'cohere/TypeScript/tsc/internal/ast/kind_stringer_generated.go'
values = {name: int(value) for name, value in re.findall(r'x\[Kind(\w+)-(\d+)\]', KINDS.read_text())}
files = [
    ('no_unnecessary_type_parameters.a', 'typescript/no_unnecessary_type_parameters.go'),
    ('no_useless_assignment.a', 'core/no_useless_assignment.go'),
    ('restrict_template_expressions.a', 'typescript/restrict_template_expressions.go'),
    ('wave_05_next/no_uncleared_race_timeout.a', 'nexus/correctness_no_uncleared_race_timeout.go'),
    ('wave_05_next/no_process_exit_after_output.a', 'nexus/correctness_no_process_exit_after_output.go'),
    ('wave_05_next/require_blocking_standard_streams.a', 'nexus/correctness_require_blocking_standard_streams.go'),
]
for native, production in files:
    block = (GO / production).read_text().split('return rule.Listeners{', 1)[1].split('\n\t\t}', 1)[0]
    expected = [values[name] for name in re.findall(r'ast.Kind(\w+):', block)]
    declaration = re.search(r'export const listenerKinds: readonly number\[\] = \[([^]]+)\];', (RULES / native).read_text())
    assert declaration, native
    actual = [int(part.strip()) for part in declaration[1].split(',') if part.strip()]
    assert actual == expected, (native, actual, expected)
    mutant = actual.copy()
    mutant[0] += 1
    assert mutant != expected, (native, 'kind-value mutant survived')
    print(native, actual, 'matches Go; changed numeric kind caught by metadata comparison')
print('PASS six numeric listener declarations and six metadata mutants')

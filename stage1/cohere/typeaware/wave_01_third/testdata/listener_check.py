"""Compare listener metadata with production Go registrations and enum values."""
import json
from pathlib import Path
import re
import sys
repo = Path(sys.argv[1])
values = dict((name.removeprefix('Kind'), int(number)) for name, number in (line.split() for line in Path(sys.argv[2]).read_text().splitlines()))
pairs = [
 ('no_implicit_return.a', 'nexus/correctness_no_implicit_return.go'),
 ('no_deprecated.a', 'typescript/no_deprecated.go'),
 ('no_else_return.a', 'core/no_else_return.go'),
 ('wave_01_next/child_process_error_listener.a', 'nexus/correctness_require_child_process_error_listener.go'),
 ('wave_01_next/response_status_check.a', 'nexus/correctness_require_response_status_check.go'),
 ('wave_01_next/independent_await_in_loop.a', 'nexus/performance_no_independent_await_in_loop.go'),
]
for native, oracle in pairs:
 source = (repo/'cohere/internal/lint/rules'/oracle).read_text()
 names = re.findall(r'^\t{3}ast.Kind(\w+):', source, re.M)
 expected = [values[name] for name in names]
 text = (repo/'stage1/cohere/typeaware'/native).read_text()
 actual = json.loads(re.search(r'listenerKinds: readonly number\[\] = (\[[^;]+\]);', text).group(1))
 assert actual == expected, (native, actual, expected)
 mutant = actual.copy(); mutant[0] += 1
 assert mutant != expected, native
 print(native, 'PASS numeric registration; +1 mutant caught')

#!/usr/bin/env python3
# coverage.py <first seed> <count>: generate each program the run used (adamic-fuzz -print, the same
# patched generator) and count how many programs contain each construct.
import re, subprocess, sys
first, count = int(sys.argv[1]), int(sys.argv[2])
probes = {
	# fuzz.Features, by what each one writes
	'field-updates': r'counter\.\w+ (\+|-|\*)=|counter\.\w+(\+\+|--)',
	'number-tostring': r'\)\.toString\(\)',
	'number-functions': r'Number\.(parseInt|parseFloat|isInteger|isNaN|isFinite)',
	'string-index or array-index ([] ?? or .at)': r'\w\[[^\]]+\] \?\?|\.at\(',
	'string-search': r'\.(lastIndexOf|replaceAll|trimStart|trimEnd)\(',
	'array-write': r'^\s*[\w.]+\[[^\]]+\] = ',
	'array-spread': r'\[\.\.\.',
	'array-search': r'\.(indexOf|includes)\(',
	'array-methods': r'\.(reverse|concat|reduce|filter|find|findIndex|some|every)\(',
	'sort-callback': r'\.sort\(',
	'map-iteration': r'for \(const \[\w+, \w+\] of table',
	'closures-deep': r'pending\w*\.push\(',
	'map-mutation': r'of table\.(keys|values)\(\)|table\.delete\(',
	'splice': r'\.splice\(',
	'surrogates': r'\\u[dD][89abAB]',
	'case-mapping': r'\.to(Upper|Lower)Case\(',
	'defaults': r'parameter\d+: [\w\[\]]+ = ',
	'optional-chains': r'\?\.next',
	'number-formats': r'\.to(Exponential|Precision)\(',
	'array-from': r'Array\.from\(',
	# the language beyond the list
	'class': r'^class ',
	'switch': r'\bswitch \(',
	'do...while': r'\bdo \{',
	'while': r'\bwhile \(',
	# Operands, not types: ' | ' alone matches every Link | undefined.
	'bitwise operators': r'\) (&|\^|<<|>>>?|\|) \(|~\(',
	'Set': r'\bSet<|new Set\(',
	'tuples': r'\]: \[',
	'typeof': r'\btypeof\b',
	'instanceof': r'\binstanceof\b',
	'string | number unions': r'string \| number',
	'generics': r'<T>|<T,',
	'String.fromCharCode': r'fromCharCode|fromCodePoint',
	'padStart/padEnd': r'\.pad(Start|End)\(',
	'repeat': r'\.repeat\(',
	'split': r'\.split\(',
	'normalize': r'\.normalize\(',
	'slice': r'\.slice\(',
	'trim': r'\.trim\(',
	'toFixed': r'\.toFixed\(',
	'Math.round/trunc/sign': r'Math\.(round|trunc|sign)\(',
	'Math.pow / **': r'Math\.pow\(|\*\*',
	'panic': r'\bpanic\(',
	'checked cast (as)': r' as [A-Z]',
	'Map.get': r'table\.get\(',
	'Map spread': r'\[\.\.\.table',
	'nested functions': r'^\t+function ',
	'break/continue': r'\b(break|continue);',
	'readTextFile/writeTextFile': r'(read|write)TextFile',
	'programArguments': r'programArguments',
	'console.error': r'console\.error',
}
counts = {name: 0 for name in probes}
for seed in range(first, first + count):
	source = subprocess.run(['/tmp/adamic-gate/adamic-fuzz-r', '-seed', str(seed), '-print'], capture_output=True, text=True).stdout
	for name, pattern in probes.items():
		if re.search(pattern, source, re.MULTILINE):
			counts[name] += 1
for name, seen in counts.items():
	print(f'{seen:6d} {name}')

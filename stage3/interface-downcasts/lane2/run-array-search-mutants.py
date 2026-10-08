#!/usr/bin/env python3
"""Selected-read contracts, lazy search bounds, identity and NaN controls."""
import pathlib, subprocess, sys
root = pathlib.Path(__file__).resolve().parents[3]
logs = pathlib.Path('/tmp/adamic-lane2-array-search-mutants')
logs.mkdir(exist_ok=True)
cases = [
 ('reference-source-contract', 'internal/native/runtime/view_arrays.c', 'if (storage == 4 || storage == 5 || storage == 6 || storage == 8 || storage == 9 || storage == 10)', 'if (false && (storage == 4 || storage == 5 || storage == 6 || storage == 8 || storage == 9 || storage == 10))', 'native-array-reference-write'),
 ('javascript-reference-source-contract', 'internal/javascript/view_arrays.go', 'if (storage === 4 || storage === 5 || storage === 6 || storage === 8 || storage === 9 || storage === 10)', 'if (false && (storage === 4 || storage === 5 || storage === 6 || storage === 8 || storage === 9 || storage === 10))', 'native-array-reference-write'),
 ('search-contract', 'internal/lower/view_arrays.go', 'Last: last, ViewRead: l.viewArrayUse(node, receiver, element, false)', 'Last: last, ViewRead: ir.ArrayViewRead{}', 'native-array-search-bad'),
 ('search-literal', 'internal/native/view_arrays.go', 'if len(read.ViewAllowed) != 0 {', 'if false && len(read.ViewAllowed) != 0 {', 'native-array-search-literal'),
 ('search-lazy-stop', 'internal/native/view_array_search.go', 'e.line("    break;")', 'e.line("    continue;")', 'native-array-search-lazy'),
 ('search-same-value-zero', 'internal/native/view_array_search.go', 'equal += " || (isnan("', 'equal += " || (false && isnan("', 'native-array-search'),
 ('search-negative-bound', 'internal/native/view_array_search.go', 'e.line("if (%s < 0) %s += %s;", start, start, count)', 'e.line("if (false && %s < 0) %s += %s;", start, start, count)', 'native-array-search'),
 ('search-identity', 'internal/native/view_array_search.go', 'equal = slot + "->reference == " + value', 'equal = "true"', 'native-array-search-object'),
 ('javascript-negative-bound', 'internal/javascript/view_array_search.go', 'relative := "if (start < 0) start += count;"', 'relative := "if (false && start < 0) start += count;"', 'native-array-search'),
]
for name, relative, before, after, probe in cases:
 if len(sys.argv) > 1 and name not in sys.argv[1:]: continue
 path = root / relative
 original = path.read_bytes()
 assert original.decode().count(before) == 1, name
 try:
  path.write_text(original.decode().replace(before, after))
  log = logs / (name + '.log')
  with log.open('wb') as output:
   result = subprocess.run(['go','test','./internal/oracle','-run','^TestCheckedViewArraySearch/'+probe+'$','-count=1','-timeout','10m'],cwd=root,stdout=output,stderr=subprocess.STDOUT)
  observed = log.read_text()
  assert result.returncode != 0 and '--- FAIL:' in observed and '[build failed]' not in observed, observed
  assert 'clang failed' not in observed and 'SyntaxError' not in observed, observed
  print(name+': caught by '+probe+'; '+str(log), flush=True)
 finally:
  path.write_bytes(original)

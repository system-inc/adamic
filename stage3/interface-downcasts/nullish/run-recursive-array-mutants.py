#!/usr/bin/env python3
"""Prove each new recursive array boundary independently, restoring each edit."""
import json
import os
from pathlib import Path
import subprocess

repository = Path(__file__).resolve().parents[3]
output = Path(os.environ.get("ADAMIC_RECURSIVE_ARRAY_MUTANT_LOGS", "/tmp/views-lazy-array-mutants"))
output.mkdir(parents=True, exist_ok=True)
mutants = [
 ("native-write", "internal/native/view_array_writes.go",
  'e.mapEntryNominalCertificate(id, value, "<array write>")',
  'e.line("(void)%s;", value)',
  '^TestCheckedViewRecursiveMutableArrayMutants$/(push|index)$'),
 ("javascript-write", "internal/javascript/view_array_writes.go",
  'adamicArrayNominalProducers[target](value);', 'void value;',
  '^TestCheckedViewRecursiveMutableArrayMutants$/(push|index)$'),
 ("source-certificate", "internal/lower/view_array_writes.go",
  'return l.mapEntrySlot(node, element)', 'return 0',
  '^TestCheckedViewRecursiveMutableArrays$/write$'),
 ("mutable-admission", "internal/lower/view_maps_nested_nominal.go",
  'return l.recursiveNominalObjectPath(l.viewArrayElementType(target), seen)',
  'return l.isLibraryType(base, "ReadonlyArray") && l.recursiveNominalObjectPath(l.viewArrayElementType(target), seen)',
  '^TestCheckedViewRecursiveMutableArrays$/control$'),
]
results = []
for name, relative, old, new, selection in mutants:
 source = repository / relative
 original = source.read_text()
 if original.count(old) != 1:
  raise RuntimeError(f"{name}: replacement does not identify one guard")
 try:
  source.write_text(original.replace(old, new, 1))
  command = ['go','test','./internal/oracle','-run',selection,'-v','-count=1','-timeout','5m']
  with (output / (name+'.log')).open('w') as log:
   result = subprocess.run(command,cwd=repository,stdout=log,stderr=subprocess.STDOUT)
  text = (output / (name+'.log')).read_text()
  caught = result.returncode == 1 and '--- FAIL:' in text and '[build failed]' not in text and '[no test files]' not in text
  results.append({'name':name,'command':command,'exit':result.returncode,'caught':caught,'log':str(output/(name+'.log'))})
  print(f"{name}: exit={result.returncode}, caught={caught}",flush=True)
  if not caught:
   raise RuntimeError(f"{name}: mutant survived or failed only at build time")
 finally:
  source.write_text(original)
(output/'results.json').write_text(json.dumps(results,indent=2)+'\n')

#!/usr/bin/env python3
"""Prove contextual callback checks reject unserved contracts."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
root = Path(__file__).resolve().parents[4]
source = root / 'internal/lower/overload_callback.go'
original = source.read_text()
mutants = [
 ('input', original.replace('if !l.censusRelated(given, takes) || !l.overloadCallbackStorage(given, takes) {', 'if false && !l.censusRelated(given, takes) || !l.overloadCallbackStorage(given, takes) {')),
 ('result', original.replace('if !l.censusRelated(produced, result) || !l.overloadCallbackStorage(produced, result) {', 'if false && !l.censusRelated(produced, result) || !l.overloadCallbackStorage(produced, result) {')),
 ('representation', original.replace('actual == output && !censusCallableSlotless(actual)', '(actual == output || actual != output) && !censusCallableSlotless(actual)')),
]
mutants.append(('field-storage', original.replace(' || !l.overloadCallbackStorage(given, takes)', '')))
rows=[]
for name, changed in mutants:
 assert changed != original
 with tempfile.TemporaryDirectory(prefix='callback-mutant-') as scratch:
  scratch=Path(scratch)
  replacement=scratch/'overload_callback.go'
  replacement.write_text(changed)
  overlay=scratch/'overlay.json'
  overlay.write_text(json.dumps({'Replace':{str(source):str(replacement)}}))
  selector='^TestOverloadCallbackFieldStorage$' if name == 'field-storage' else '^TestOverloadCallbackUnserved$/'+name+'$'
  command=['go','test','-overlay',str(overlay),'./internal/lower','-run',selector,'-count=1','-v']
  path=Path(__file__).parent/(name+'-mutant.log.txt')
  with path.open('w') as log:
   result=subprocess.run(command,cwd=root,env=os.environ,stdout=log,stderr=subprocess.STDOUT)
  output=path.read_text()
  caught=result.returncode != 0 and '--- FAIL:' in output and 'got <nil>' in output and '[build failed]' not in output
  rows.append(dict(mutant=name,exit=result.returncode,caught=caught))
  print(name, 'caught' if caught else 'NOT CAUGHT',flush=True)
(Path(__file__).parent/'mutants.json').write_text(json.dumps(rows,indent=2)+'\n')
assert all(row['caught'] for row in rows)

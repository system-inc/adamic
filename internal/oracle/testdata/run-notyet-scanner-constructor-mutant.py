#!/usr/bin/env python3
"""Run each computed-name mutant through an isolated source overlay."""
from pathlib import Path
import os
import json
import subprocess

root = Path(__file__).resolve().parents[3]
logs = Path('/tmp/destructuring-scanner-mutant')
logs.mkdir(exist_ok=True)
helper = 'internal/lower/computed_field_name.go'
mutants = [
 ('scanner-constructor-key', helper, 'return left + right, true', 'return left + "wrong" + right, true', 'oracle', 'stdout differs'),
]

for name, file, before, after, package, catcher in mutants:
 path = root / file
 original = path.read_text()
 assert original.count(before) == 1, (name, original.count(before))
 source = logs / (name + '.go')
 source.write_text(original.replace(before, after, 1))
 overlay = logs / (name + '.json')
 overlay.write_text(json.dumps({'Replace': {str(path): str(source)}}))
 pattern = '^TestNativeAgreesWithNode/internal/oracle/testdata/notyet_computed_scanner_constructor.a$' if package == 'oracle' else '^TestComputedFieldNameGapsStayExplicit$'
 if name == 'enum-key': pattern = '^TestNativeAgreesWithNode/internal/oracle/testdata/notyet_computed_numeric_names.a$'
 with (logs / (name + '.log')).open('w') as log:
  result = subprocess.run(['go','test','-overlay='+str(overlay),'./internal/'+package,'-run',pattern,'-count=1','-timeout','10m'],cwd=root,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1'),stdout=log,stderr=subprocess.STDOUT)
 output = (logs / (name + '.log')).read_text()
 assert result.returncode != 0 and catcher in output and '[build failed]' not in output, (name,result.returncode,output)
 print(name + ': caught by ' + catcher,flush=True)

#!/usr/bin/env python3
"""Source Node and compile-admission observations, not a runtime oracle pass."""
import json
import pathlib
import subprocess

root = pathlib.Path(__file__).resolve().parents[3]
lane = pathlib.Path(__file__).resolve().parent
observations = []
for source in sorted(lane.glob('*.a')):
 row = {'fixture':str(source.relative_to(root))}
 node = subprocess.run(['node','--input-type=module-typescript'],input=source.read_bytes(),capture_output=True)
 row['node'] = {'exit':node.returncode,'stdout':node.stdout.decode(),'stderr':node.stderr.decode()}
 assert node.returncode == 0, row
 for backend in ['c','js']:
  result = subprocess.run(['go','run','./cmd/adamic',backend,str(source)],cwd=root,capture_output=True)
  row[backend] = {'exit':result.returncode,'stdout':result.stdout.decode(),'stderr':result.stderr.decode()}
  expected = 'refuses an index signature' if 'versionpaths' in source.name else 'stage 0 can\'t lower'
  assert result.returncode != 0 and expected in result.stderr.decode(), row
 observations.append(row)
 print(source.name + ': Node exit 0; native and JavaScript compilation remain closed')
(lane / 'probe-observations.json').write_text(json.dumps(observations,indent=2)+'\n')

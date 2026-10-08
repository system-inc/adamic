#!/usr/bin/env python3
"""Run three actual-source fixtures and require each source mutant to change Node bytes."""
import argparse
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile

H = Path(__file__).resolve().parent
REPO = H.parents[3]
p = argparse.ArgumentParser()
p.add_argument('api', type=Path)
a = p.parse_args()
out = H / 'evidence/fixtures'
out.mkdir(parents=True, exist_ok=True)
expected = {
 '01_primitive_config.a': 'number 7 0\nobject null 0\nobject [1,true] 0\nobject {"x":2} 0\n',
 '02_unvalidated_list.a': 'true\ntrue\nfalse\n',
 '03_timer_roundtrip.a': 'callback function args string object boolean\nhandle number\ncancel-same true\ncallback function args string object boolean\ncancel-same true\ncallback function args string object boolean\nhandle object\ncancel-same true\ncallback function args string object boolean\ncancel-same true\n',
}
mutations = {
 '01_primitive_config.a': ('/*returnValue*/ true', '/*returnValue*/ false'),
 '02_unvalidated_list.a': ('return isArray(value);', 'return false;'),
 '03_timer_roundtrip.a': ('hostWithWatch.clearTimeout(state.timerToBuildInvalidatedProject);', 'hostWithWatch.clearTimeout(undefined);'),
}
status = []
env = dict(os.environ, PUBLIC_ANY_API=str(a.api.resolve()))
for file, wanted in expected.items():
 source = H / 'fixtures' / file
 command = ['node', '--disable-warning=ExperimentalWarning', '--import', str(H/'api-loader.mjs'), str(REPO/'oracle/node.mjs'), str(source)]
 result = subprocess.run(command, capture_output=True, env=env, timeout=30)
 (out/(file+'.node.stdout')).write_bytes(result.stdout)
 (out/(file+'.node.stderr')).write_bytes(result.stderr)
 assert result.returncode == 0 and result.stdout == wanted.encode() and not result.stderr, (file, result)
 before, after = mutations[file]
 text = source.read_text()
 assert text.count(before) == 1, file
 with tempfile.TemporaryDirectory(prefix='public-any-mutant-') as temp:
  mutant = Path(temp)/file
  mutant.write_text(text.replace(before, after))
  changed = subprocess.run(command[:-1]+[str(mutant)], capture_output=True, env=env, timeout=30)
 (out/(file+'.mutant.stdout')).write_bytes(changed.stdout)
 (out/(file+'.mutant.stderr')).write_bytes(changed.stderr)
 assert changed.returncode == 0 and not changed.stderr, 'mutant must execute successfully on Node'
 assert changed.stdout != result.stdout, 'source mutant survived byte comparison'
 with (out/(file+'.build.stdout')).open('wb') as stdout, (out/(file+'.build.stderr')).open('wb') as stderr:
  compiled = subprocess.run(['go', 'run', './cmd/adamic', 'build', str(source), '-o', str(Path('/tmp')/('public-any-'+file+'.native'))], cwd=REPO, stdout=stdout, stderr=stderr, timeout=180)
 diagnostic = (out/(file+'.build.stderr')).read_text()
 if compiled.returncode == 0:
  native = subprocess.run([str(Path('/tmp')/('public-any-'+file+'.native'))], capture_output=True, timeout=30)
  assert (native.stdout,native.stderr,native.returncode)==(result.stdout,result.stderr,result.returncode), 'silent miscompile'
  outcome='Compiles'
 else:
  outcome='Refused' if 'refuses' in diagnostic else 'NotYet' if "can't lower" in diagnostic else 'Checker'
 row=dict(file=file,tsc=[line.removeprefix('// From TypeScript 6.0.3, ') for line in text.splitlines() if line.startswith('// From')],reason=text.splitlines()[0].removeprefix('// Census reason: '),node=dict(stdout=result.stdout.decode(),stderr='',exit=0),stage0=dict(outcome=outcome,what=diagnostic),mutant=dict(change=before+' -> '+after,node_exit=changed.returncode,caught_by='exact stdout against unchanged Node fixture',stdout=changed.stdout.decode()))
 status.append(row)
 print(file,'Node PASS; mutant caught; stage0',outcome)
(H/'fixtures/status.json').write_text(json.dumps(status,indent=2)+'\n')
print('PASS: three Node fixtures, three executable source mutants')

#!/usr/bin/env python3
"""Prove the mode evidence checks reject deliberately corrupted observations."""
import json,shutil,subprocess,tempfile
from pathlib import Path
here=Path(__file__).resolve().parent; source=here/'evidence/mode-c09c22b8'
with tempfile.TemporaryDirectory(prefix='tsc-entry-mode-baseline-') as t:
 with (Path(t)/'baseline.log').open('wb') as log:
  subprocess.run(['python3',str(here/'verify-mode.py'),str(source)],stdout=log,stderr=subprocess.STDOUT,check=True)
def edit_json(e,name,fn):
 p=e/name;data=json.loads(p.read_text());fn(data);p.write_text(json.dumps(data))
mutants=[('classification',lambda e:edit_json(e,'admission.json',lambda d:d[0].update(mode_status='scheduled'))),('missing-stop',lambda e:edit_json(e,'admission.json',lambda d:d.pop())),('split-result',lambda e:(e/'entry-1.exit').write_text('0\n')),('first-diagnostic',lambda e:(e/'entry-0.stderr').write_text('wrong\n')),('source-identity',lambda e:edit_json(e,'identity.json',lambda d:d['after'].update(fake='changed'))),('control-identity',lambda e:edit_json(e,'identity.json',lambda d:d.update(main_control_diff='changed'))),('ledger-scope',lambda e:edit_json(e,'ledger-roots.json',lambda d:d.remove('src/compiler/builder.ts'))),('node-oracle',lambda e:(e/'witnesses/app/node.stdout').write_text('8\n')),('outside-scheduled-check',lambda e:edit_json(e,'witnesses/outside/load.json',lambda d:d.update(dispositions=[]))),('ordinary-error',lambda e:edit_json(e,'09-load.json',lambda d:d.update(ordinary_diagnostics=[],loaded=True)))]
for name,mutate in mutants:
 with tempfile.TemporaryDirectory(prefix='tsc-entry-mode-mutant-') as t:
  e=Path(t)/'evidence';shutil.copytree(source,e);mutate(e)
  with (Path(t)/'check.log').open('wb') as log:r=subprocess.run(['python3',str(here/'verify-mode.py'),str(e)],stdout=log,stderr=subprocess.STDOUT)
  assert r.returncode!=0,name
  print(name+': caught by verify-mode.py',flush=True)
print('PASS: all ten mutants rejected')

#!/usr/bin/env python3
"""Prove checked-stop stdout and exit checks with three independent runs."""
import argparse
import json
import os
import subprocess
import time
from pathlib import Path
root=Path(__file__).resolve().parents[3]
parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('--scratch',type=Path,required=True)
arguments=parser.parse_args()
work=arguments.scratch.resolve()
work.mkdir(parents=True,exist_ok=False)
source=root/'internal/native/emit_statements.go'
original=source.read_text()
needle='e.line("adamic_write_line(%s, %s);", stream, value)'
assert original.count(needle)==1
replacements={
 'different-line':r'''_ = value
        e.line("{ static adamic_string mutant = ADAMIC_STRING(\"different line\"); adamic_write_line(%s, &mutant); }", stream)''',
 'missing-line':r'''_ = stream
        e.line("(void)%s;", value)''',
}
results=[]
for name in ['different-line','missing-line','never-stops']:
 case=work/name
 case.mkdir(exist_ok=True)
 env=os.environ.copy()
 command=['go','test','./stage3/fixtures','-run','^TestFixturesEntries$/^entries$/^03_hidden_other_type.a$','-count=1','-timeout','90s','-v']
 # Change real C output through a scratch Go overlay. The repository's emitter
 # and the JavaScript backend remain unchanged, and clang must still succeed.
 if name in replacements:
  overlay_source=case/'emit_statements.go'
  overlay_source.write_text(original.replace(needle,replacements[name]))
  overlay=case/'overlay.json'
  overlay.write_text(json.dumps({'Replace':{str(source):str(overlay_source)}}))
  env['GOFLAGS']=(env.get('GOFLAGS','')+' -overlay='+str(overlay)).strip()
 else:
  bucket=case/'fixtures'/'entries'
  bucket.mkdir(parents=True,exist_ok=True)
  fixture=root/'stage3/fixtures/entries/03_hidden_other_type.a'
  (bucket/fixture.name).write_text(fixture.read_text().replace("hidden: 'wrong'",'hidden: 2'))
  rows=json.loads((root/'stage3/fixtures/entries/status.json').read_text())
  row=next(r for r in rows if r['file']==fixture.name)
  # Keep source Node truthful so only the promised stop, not its golden, fails.
  row['node']['stdout']=row['node']['stdout'].replace('hidden:wrong','hidden:2')
  (bucket/'status.json').write_text(json.dumps([row]))
  command+=['-args','-fixtures',str(case/'fixtures')]
 started=time.monotonic()
 with (case/'test.log').open('w') as log:
  log.write('COMMAND: '+str(command)+'\n'); log.flush()
  result=subprocess.run(command,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT,timeout=120)
 elapsed=time.monotonic()-started
 output=(case/'test.log').read_text()
 assert result.returncode==1,(name,result.returncode,output)
 assert 'pinned native check byte comparison failed' in output,(name,output)
 if name!='never-stops':
  assert 'native versus checked Node byte comparison failed' in output,(name,output)
  assert 'got exit=70' in output,(name,output)
 else:
  assert 'got exit=0' in output,(name,output)
 results.append(dict(name=name,exit=result.returncode,wall=elapsed,log=str(case/'test.log')))
 (work/'results.json').write_text(json.dumps(results,indent=2)+'\n')
 print(name,'CAUGHT',round(elapsed,3),flush=True)

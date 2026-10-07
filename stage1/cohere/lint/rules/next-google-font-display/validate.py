#!/usr/bin/env python3
"""Validate the blocked rule's URL decisions against the actual Go JSX rule."""
from pathlib import Path
import json,subprocess,shutil
owned=Path(__file__).resolve().parent
repo=owned.parents[4]
scratch=Path('/tmp/w05-google-decision');scratch.mkdir(exist_ok=True)
def command(args,name,cwd=repo):
 with (scratch/(name+'.log')).open('wb') as out, (scratch/(name+'.stderr.log')).open('wb') as err:
  subprocess.run(args,cwd=cwd,stdout=out,stderr=err,check=True)
 assert (scratch/(name+'.stderr.log')).stat().st_size == 0, name+' wrote stderr'
virtual=owned/'build_probe.go'
(scratch/'build-overlay.json').write_text(json.dumps({'Replace':{str(virtual):str(owned/'testdata/build_probe.go.txt')}}))
go_virtual=repo/'cohere/adamic_google_decision.go'
(scratch/'oracle-overlay.json').write_text(json.dumps({'Replace':{str(go_virtual):str(owned/'testdata/decision_oracle.go.txt')}}))
command(['go','build','-overlay='+str(scratch/'oracle-overlay.json'),'-o',str(scratch/'oracle'),str(go_virtual)],'oracle-build',repo/'cohere')
corpus=str(owned/'testdata/hrefs.txt')
command([str(scratch/'oracle'),corpus],'Go')
for mutant in [False,True]:
 module=owned
 if mutant:
  module=scratch/'mutant';shutil.copytree(owned,module,dirs_exist_ok=True)
  spec=json.loads((module/'mutant.json').read_text());path=module/spec['file'];text=path.read_text()
  assert text.count(spec['from'])==1
  path.write_text(text.replace(spec['from'],spec['to']))
 prefix=str(scratch/('mutant-bin' if mutant else 'baseline-bin'))
 probe=str(module/'testdata/decision_probe.a')
 command(['go','run','-overlay='+str(scratch/'build-overlay.json'),str(virtual),probe,prefix],('mutant' if mutant else 'baseline')+'-build')
 for side,args in [('Node',['node','--disable-warning=ExperimentalWarning',str(repo/'oracle/node.mjs'),probe,corpus]),('JavaScript',['node','--disable-warning=ExperimentalWarning',str(repo/'oracle/node.mjs'),prefix+'.mjs',corpus]),('native',[prefix,corpus])]:
  name=('mutant-' if mutant else '')+side
  command(args,name)
  same=(scratch/(name+'.log')).read_bytes()==(scratch/'Go.log').read_bytes()
  assert same!=mutant,(side,mutant)
  print(side,'mutant caught' if mutant else 'identical',len((scratch/'Go.log').read_bytes()),'bytes')

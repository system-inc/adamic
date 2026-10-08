#!/usr/bin/env python3
"""Source Node is the oracle; each intentional hatch must be refused by main."""
import argparse,json,subprocess,tempfile
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('--compiler',required=True);p.add_argument('--output',required=True,type=Path);p.add_argument('--check',action='store_true');a=p.parse_args()
unit=Path(__file__).resolve().parent;repo=unit.parents[3];out=a.output.resolve();out.mkdir(parents=True,exist_ok=True)
specs=json.loads((unit/'fixtures.json').read_text());results=[]
def run(command,label):
 with (out/(label+'.stdout')).open('wb') as stdout,(out/(label+'.stderr')).open('wb') as stderr:
  r=subprocess.run(command,cwd=repo,stdout=stdout,stderr=stderr,timeout=120)
 return dict(exit=r.returncode,stdout=(out/(label+'.stdout')).read_text(),stderr=(out/(label+'.stderr')).read_text())
with tempfile.TemporaryDirectory(prefix='step09-mutants-') as tmp:
 for s in specs:
  f=unit/s['file'];label=f.stem
  node=run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(f)],label+'.node');assert node['exit']==0 and node['stderr']=='',node
  build=run([a.compiler,'build',str(f),'-o',str(out/(label+'.bin'))],label+'.build')
  assert build['exit']==1 and 'Adamic 0.1 refuses' in build['stderr'],build
  assert 'cast the runtime' in build['stderr'] if label.startswith('01') else 'Object.defineProperty' in build['stderr'] if label.startswith('02') else 'type predicate whose return is not proven' in build['stderr']
  text=f.read_text();assert text.count(s['before'])==1
  mutant=Path(tmp)/(label+'.a');mutant.write_text(text.replace(s['before'],s['after']))
  observed=run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(mutant)],label+'.mutant.node')
  assert observed!=node,f'{label}: source mutant survived the Node oracle'
  results.append(dict(file=s['file'],node=node,build=build,stage0='Refused',mutant=dict(change=s['mutation'],node=observed,caughtBy='original Node stdout/stderr/exit comparison; native mutation not claimed')))
  print(label,'Refused; source mutant caught',flush=True)
(out/'results.json').write_text(json.dumps(results,indent=2)+'\n')
if a.check:
 want=json.loads((unit/'fixture-results.json').read_text())
 for x,y in zip(results,want,strict=True):
  for key in ['file','node','stage0']:assert x[key]==y[key],(x['file'],key)
  # Keep raw diagnostics while allowing the checkout path to differ.
  assert x['build']['exit']==y['build']['exit'] and x['build']['stdout']==y['build']['stdout']
  assert x['build']['stderr'].replace(str(repo),'REPO')==y['build']['stderr'].replace('/workspace/adamic','REPO')
  assert x['mutant']['node']==y['mutant']['node'],(x['file'],'mutant Node result')
print('PASS: 3 Node-held baselines, 3 intended refusals, 3 source mutants caught')

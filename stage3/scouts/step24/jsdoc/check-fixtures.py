"""Node behavior and expected main result, with one real source mutant per witness."""
import argparse,json,subprocess,tempfile
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('--compiler',required=True);p.add_argument('--output',type=Path,required=True);p.add_argument('--record',action='store_true');a=p.parse_args();u=Path(__file__).resolve().parent;repo=u.parents[3];a.output.mkdir(parents=True,exist_ok=True)
def run(cmd,label):
 with (a.output/(label+'.stdout')).open('wb') as o,(a.output/(label+'.stderr')).open('wb') as e:r=subprocess.run(cmd,cwd=repo,stdout=o,stderr=e,timeout=120)
 return dict(exit=r.returncode,stdout=(a.output/(label+'.stdout')).read_text(),stderr=(a.output/(label+'.stderr')).read_text())
rows=[]
with tempfile.TemporaryDirectory(prefix='step24-mutants-') as scratch:
 for spec in json.loads((u/'fixtures.json').read_text()):
  f=u/spec['file'];label=f.stem
  node=run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(f)],label+'.node');assert node['exit']==0 and node['stderr']=='',node
  native=run([a.compiler,'build',str(f),'-o',str(a.output/(label+'.bin'))],label+'.build')
  if native['exit']==0:
   outcome='Compiles';output=run([str(a.output/(label+'.bin'))],label+'.native');assert output==node,dict(silent_miscompile=label,node=node,native=output)
  else:
   assert native['exit']==1,native
   outcome='Refused' if 'Adamic 0.1 refuses ' in native['stderr'] else 'Checker' if ' error TS' in native['stderr'] else 'NotYet' if "can't lower" in native['stderr'] and ' yet' in native['stderr'] else 'Unexpected'
   assert outcome!='Unexpected',native;output=None
  text=f.read_text();assert text.count(spec['before'])==1,(label,spec['before']);mut=Path(scratch)/(label+'.a');mut.write_text(text.replace(spec['before'],spec['after']))
  mutant=run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(mut)],label+'.mutant');assert mutant['exit']==0 and mutant['stderr']=='' and mutant['stdout']!=node['stdout'],mutant
  mutant_build=run([a.compiler,'build',str(mut),'-o',str(a.output/(label+'.mutant.bin'))],label+'.mutant.build');assert mutant_build['exit']==0,mutant_build
  mutant_native=run([str(a.output/(label+'.mutant.bin'))],label+'.mutant.native');assert mutant_native==mutant and mutant_native!=node,mutant_native
  checked={}
  for mode in ['sanitize','count']:
   binary=a.output/(label+'.'+mode+'.bin');build=run([a.compiler,'build',str(f),'-o',str(binary),'--'+mode],label+'.'+mode+'.build');assert build['exit']==0,build
   observation=run([str(binary)],label+'.'+mode);assert observation['exit']==0 and observation['stdout']==node['stdout'],observation
   if mode=='sanitize':assert observation['stderr']=='',observation
   checked[mode]=observation
  rows.append(dict(file=spec['file'],node=node,stage0=dict(outcome=outcome,build=native,output=output),checked=checked,mutant=dict(change=[spec['before'],spec['after']],node=mutant,native=mutant_native,caught='original Node stdout comparison, source and native mutants both differ')))
  print('PASS:',label,outcome,'source mutant caught',flush=True)
if a.record:(u/'fixture-results.json').write_text(json.dumps(rows,indent=2)+'\n')
else:
 expected=json.loads((u/'fixture-results.json').read_text());assert rows==expected,(rows,expected)
print('PASS: 3 witnesses, 3 source mutants')

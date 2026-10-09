#!/usr/bin/env python3
"""Measure witnesses and source mutants against stock Node and a scratch compiler."""
import argparse, json, os, re, subprocess, tempfile
from pathlib import Path
p=argparse.ArgumentParser(description=__doc__); p.add_argument('--compiler',required=True); p.add_argument('--output',required=True,type=Path); p.add_argument('--check',action='store_true'); a=p.parse_args()
root=Path(__file__).resolve().parent; repo=root.parents[3]; out=a.output.resolve(); out.mkdir(parents=True,exist_ok=True)
specs=json.loads((root/'witnesses.json').read_text()); results=[]
def run(command, label, env=None):
 with (out/(label+'.stdout')).open('wb') as stdout, (out/(label+'.stderr')).open('wb') as stderr:
  proc=subprocess.run(command,cwd=repo,stdout=stdout,stderr=stderr,timeout=120,env=env)
 return dict(exit=proc.returncode,stdout=(out/(label+'.stdout')).read_text(),stderr=(out/(label+'.stderr')).read_text())
def node(source,label): return run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(source)],label)
def native(source,label):
 binary=out/(label+'.bin'); build=run([a.compiler,'build',str(source),'-o',str(binary)],label+'.build')
 r=dict(build=build)
 if build['exit']==0:r['run']=run([str(binary)],label+'.run')
 return r
def agrees(n,r):return 'run' in r and r['run']==n
with tempfile.TemporaryDirectory(prefix='step12-mutants-') as tmp:
 (Path(tmp)/'fail.a').write_bytes((root/'witnesses/fail.a').read_bytes())
 for s in specs:
  source=root/s['file']; label=source.stem; n=node(source,label+'.node'); assert n['exit']==0 and n['stderr']=='',(label,n)
  r=native(source,label); outcome='Matches' if agrees(n,r) else 'Mismatch' if 'run' in r else 'Checker' if 'TS' in r['build']['stderr'] else 'Refused' if 'refuses' in r['build']['stderr'] else 'NotYet'
  assert outcome!='Mismatch',(label,n,r)
  text=source.read_text(); m=s['mutant']; assert text.count(m['before'])==1
  mutant=Path(tmp)/(label+'.a');mutant.write_text(text.replace(m['before'],m['after']))
  mn=node(mutant,label+'.mutant.node'); assert mn['exit']==0 and mn['stderr']=='', (label,mn)
  assert mn!=n, f'{label}: source mutant survived Node expectation'
  mr=native(mutant,label+'.mutant')
  if 'run' in mr:
   assert agrees(mn,mr),f'{label}: mutant silent miscompile'
   assert mr['run']!=n,f'{label}: native mutant survived original expectation'
  r.update(file=s['file'],node=n,outcome=outcome,mutant=dict(change=m,node=mn,native=mr,caughtBy='Node stdout comparison'+(' and native stdout comparison' if 'run' in mr else '; native blocked, no native mutation proof')))
  if outcome=='Matches':
   counted=out/(label+'.count.bin'); cb=run([a.compiler,'build',str(source),'-o',str(counted),'--count'],label+'.count.build');assert cb['exit']==0
   r['count']=run([str(counted)],label+'.count.run');assert r['count']['stdout']==n['stdout'] and r['count']['exit']==0
   counts=[int(x) for x in re.findall(r'\d+',r['count']['stderr'])];assert len(counts)==6 and counts[0]==counts[1]+counts[5],counts
   sanitized=out/(label+'.sanitize.bin'); sb=run([a.compiler,'build',str(source),'-o',str(sanitized),'--sanitize'],label+'.sanitize.build');assert sb['exit']==0
   r['sanitize']=run([str(sanitized)],label+'.sanitize.run',dict(os.environ,ASAN_OPTIONS='detect_leaks=1',UBSAN_OPTIONS='halt_on_error=1'));assert r['sanitize']==n,(label,r['sanitize'])
  results.append(r); print(label,outcome,r['mutant']['caughtBy'],flush=True)
(out/'results.json').write_text(json.dumps(results,indent=2)+'\n')
if a.check:
 expected=json.loads((root/'results.json').read_text())
 # Scratch temp filenames occur in mutant build diagnostics; compare stable baseline evidence.
 for observed,want in zip(results,expected,strict=True):
  for key in ['file','node','outcome','build','count','sanitize']:
   assert observed.get(key)==want.get(key),(observed['file'],key)
  assert observed['mutant']['node']==want['mutant']['node']
print('PASS: all Node baselines and mutants; every compiled binary matches its own Node source')

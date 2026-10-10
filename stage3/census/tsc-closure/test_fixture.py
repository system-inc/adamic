#!/usr/bin/env python3
"""Two-file known-answer measurement; dropping the outside root must fail."""
import argparse,json,os,subprocess
from pathlib import Path
p=argparse.ArgumentParser();p.add_argument('binary',type=Path);p.add_argument('hidden_tools',type=Path);p.add_argument('output',type=Path);a=p.parse_args();here=Path(__file__).resolve().parent;out=a.output.resolve();out.mkdir();source=out/'source';(source/'src/compiler').mkdir(parents=True)
for name in ['src/compiler/entry.a','src/outside.a']:
 text=(here/'fixtures'/name).read_text().replace('../outside.a','../outside.js');(source/name.replace('.a','.ts')).write_text(text)
def run(name,cmd,env=None,okay=True):
 with (out/(name+'.log')).open('wb') as log:r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,env=env)
 if okay:assert r.returncode==0,(name,(out/(name+'.log')).read_text())
 return r.returncode
run('closure',['node',str(here/'closure.cjs'),str(source),'src/compiler/entry.ts',str(out/'closure.json')]);run('stock',['node',str(a.hidden_tools/'units.cjs'),str(source),str(out/'stock.json')])
m=json.loads((out/'closure.json').read_text());drop={**m,'files':[f for f in m['files'] if f['file'].startswith('src/compiler/')]};(out/'drop.json').write_text(json.dumps(drop))
for kind,manifest in [('baseline','closure.json'),('dropped','drop.json')]:
 for mode in ['latent','full']:
  run(kind+'-'+mode,[str(a.binary.resolve()),str(source/'src/compiler'),str(out/(kind+'-'+mode+'.jsonl'))],{**os.environ,'LATENT_ROOT_MANIFEST':str(out/manifest),'LATENT_ASSERT_NO_OUTPUT':'1','LATENT_FULL':'1' if mode=='full' else '0'})
def summarize(latent,full,target):return ['python3',str(here/'measure.py'),str(out/'closure.json'),str(out/latent),str(out/full),str(out/'stock.json'),str(a.hidden_tools/'hidden.py'),str(out/target)]
run('answer',summarize('baseline-latent.jsonl','baseline-full.jsonl','answer.json'))
r=json.loads((out/'answer.json').read_text());assert r['latent']['all']['checker_by_code']=={'TS2322':1};assert r['latent']['all']['counts']=={}
assert r['hidden']['groups']['all']['bytes']==183 and r['hidden']['groups']['all']['hidden_bytes']==54
assert r['hidden']['groups']['outside_compiler']['bytes']==88 and r['hidden']['groups']['outside_compiler']['hidden_bytes']==54
assert r['hidden']['files']['src/outside.ts']['hidden_ranges']==[[33,87]]
for mode in ['latent','full']:
 first='dropped-latent.jsonl' if mode=='latent' else 'baseline-latent.jsonl';second='dropped-full.jsonl' if mode=='full' else 'baseline-full.jsonl'
 assert run('drop-'+mode,summarize(first,second,'invalid-'+mode+'.json'),okay=False)!=0
 assert 'closure coverage mismatch' in (out/('drop-'+mode+'.log')).read_text()
 print('outside-root-drop '+mode+': caught by closure coverage')
print('PASS: two sources, one outside TS2322, no lowering refusals, 54/183 hidden bytes; both real root-drop mutants caught')

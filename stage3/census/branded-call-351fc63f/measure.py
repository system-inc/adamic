"""Run one same-input compiler pairing sequentially and preserve all streams."""
import hashlib,json,os,subprocess,sys
from pathlib import Path
out=Path(__file__).resolve().parent
repo=out.parents[2]
label,tree=sys.argv[1],Path(sys.argv[2]).resolve()
assert label in ['A','B']

def run(name,cmd):
 rc=subprocess.run([sys.executable,str(out/'run.py'),name,str(repo),*map(str,cmd)]).returncode
 if rc:raise SystemExit(rc)

def hashes():
 return {str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for d in ['src/compiler','src/tsc'] for p in sorted((tree/d).rglob('*')) if p.is_file() and p.suffix in ['.ts','.a']}
before=hashes();(out/(label+'-source-hashes.json')).write_text(json.dumps(before,indent=2)+'\n')
with (out/(label+'-sites.json')).open('w') as f:
 rc=subprocess.run(['node',str(out/'sites.cjs'),str(tree)],stdout=f,env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))).returncode
if rc:raise SystemExit(rc)
for mode in ['entry','core']:
 for side in ['parent','head']:
  name=f'{label}-{side}-{mode}'
  env=['env','-u','LATENT_FULL','-u','LATENT_SPECULATIVE','-u','LATENT_ONLY_FILE','LATENT_ASSERT_NO_OUTPUT=1',f'LATENT_TIMES={out}/{name}-units.jsonl']
  if mode=='core':env+=['LATENT_FULL=1',f'LATENT_ONLY_FILE={tree}/src/compiler/core.ts']
  binary=f'/tmp/branded-{side}-'+('census' if mode=='entry' else 'core')
  project=tree/('src/tsc/tsc.ts' if mode=='entry' else 'src/compiler')
  run(name,env+[binary,project,out/(name+'.jsonl')])
 run(f'{label}-{mode}-pair',['python3',repo/'stage3/census/pair/pair.py',out/f'{label}-parent-{mode}.jsonl',out/f'{label}-head-{mode}.jsonl',tree,'--typescript',repo/'stage3/api/node_modules/typescript','--json',out/f'{label}-{mode}-pair.json','--report',out/f'{label}-{mode}-pair.md'])
sites=json.loads((out/(label+'-sites.json')).read_text())
unique={row['unit']:row for row in reversed(sites['checker'])}
for i,(unit,row) in enumerate(sorted(unique.items()),1):
 for side in ['parent','head']:
  name=f'{label}-{side}-focus-{i}'
  run(name,['env','LATENT_FULL=1','LATENT_ASSERT_NO_OUTPUT=1',f'/tmp/branded-{side}-focus','-project',tree/'src/tsc/tsc.ts','-where',row['site'],'-output',out/(name+'.jsonl')])
 run(f'{label}-focus-{i}-pair',['python3',repo/'stage3/census/pair/pair.py',out/f'{label}-parent-focus-{i}.jsonl',out/f'{label}-head-focus-{i}.jsonl',tree,'--typescript',repo/'stage3/api/node_modules/typescript','--json',out/f'{label}-focus-{i}-pair.json','--report',out/f'{label}-focus-{i}-pair.md'])
assert hashes()==before,'adapted input changed during measurement'
(out/(label+'-complete.json')).write_text(json.dumps(dict(pairing=label,tree=str(tree),focused_units=[dict(unit=u,site=r['site'],owners=r['owners']) for u,r in sorted(unique.items())]),indent=2)+'\n')
print(label,'complete',flush=True)

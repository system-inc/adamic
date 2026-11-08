"""Pairing A's focused full-unit inventory, separate from entry counts."""
import json,subprocess,sys,os,hashlib
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2]
marker=out/'focus-complete.json'
if marker.exists():
 saved=json.loads(marker.read_text())
 assert saved['head']==subprocess.check_output(['git','rev-parse','HEAD'],cwd='/tmp/fa-head',text=True).strip()
 assert saved['source_hashes']==json.loads((out/'source-hashes.json').read_text())
 assert saved['driver_sha256']==hashlib.sha256((out/'focus-driver.go.txt').read_bytes()).hexdigest()
 for file,sha in saved['result_hashes'].items():assert hashlib.sha256((out/file).read_bytes()).hexdigest()==sha
 for side,sha in saved['binary_hashes'].items():assert hashlib.sha256(Path('/tmp/fa-'+side+'-focus').read_bytes()).hexdigest()==sha
 print('reused complete same-pin same-source focused observations child_exit=0',flush=True)
 raise SystemExit(0)

def run(name,cwd,cmd,env=None):
 rc=subprocess.run([sys.executable,str(out/'run.py'),name,str(cwd),*cmd],env=env).returncode
 assert rc==0,(name,rc)
with (out/'sites.json').open('w') as log:
 rc=subprocess.run(['node',str(out/'sites.cjs'),'/tmp/fa-adapted'],env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules')),stdout=log).returncode
assert rc==0
sites=json.loads((out/'sites.json').read_text());units={r['unit']:r['site'] for r in reversed(sites['checker'])}
units.update({r['site']:r['site'] for r in sites['findAncestor'] if r['has_body']})
index={u:i for i,u in enumerate(sorted(units),1)};(out/'focus-index.json').write_text(json.dumps(index,indent=2)+'\n')
for side,root in [('parent',repo),('head',Path('/tmp/fa-head'))]:
 meta=json.loads(Path('/tmp/fa-'+side+'-entry/overlay.json').read_text())
 key=next(k for k in meta['Replace'] if k.endswith('/stage3/census/latent/replay/worker/main.go'))
 meta['Replace'][key]=str(out/'focus-driver.go.txt')
 overlay=out/(side+'-focus-overlay.json');overlay.write_text(json.dumps(meta,indent=2)+'\n')
 env=dict(os.environ,GOWORK='/tmp/fa-'+side+'.work')
 run(side+'-focus-build',root,['go','build','-p=2','-buildvcs=false','-overlay='+str(overlay),'-o','/tmp/fa-'+side+'-focus','./stage3/census/latent/replay/worker'],env)
 for u,selector in sorted(units.items()):
  name=side+'-focus-'+str(index[u])
  run(name,root,['env','-u','LATENT_SPECULATIVE','-u','LATENT_ONLY_FILE','LATENT_FULL=1','LATENT_ASSERT_NO_OUTPUT=1','/tmp/fa-'+side+'-focus','-project','/tmp/fa-adapted/src/tsc/tsc.ts','-where',selector,'-output',str(out/(name+'.jsonl'))])
 # Archive the complete focused overlay, Go evidence always has a text suffix.
 dest=out/'overlays'/side/'focus';dest.mkdir(parents=True,exist_ok=True);manifest=[]
 for i,(logical,actual) in enumerate(sorted(meta['Replace'].items())):
  name=f'{i:02d}-'+Path(actual).name
  if name.endswith('.go'):name+='.txt'
  data=Path(actual).read_bytes();(dest/name).write_bytes(data);manifest.append(dict(logical=logical,evidence=str((dest/name).relative_to(out)),sha256=hashlib.sha256(data).hexdigest()))
 (dest/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
(out/'focus-complete.json').write_text(json.dumps({'units':len(units),'exit':0,'head':subprocess.check_output(['git','rev-parse','HEAD'],cwd='/tmp/fa-head',text=True).strip(),'source_hashes':json.loads((out/'source-hashes.json').read_text()),'driver_sha256':hashlib.sha256((out/'focus-driver.go.txt').read_bytes()).hexdigest(),'result_hashes':{p.name:hashlib.sha256(p.read_bytes()).hexdigest() for side in ['parent','head'] for p in out.glob(side+'-focus-*.jsonl')},'binary_hashes':{side:hashlib.sha256(Path('/tmp/fa-'+side+'-focus').read_bytes()).hexdigest() for side in ['parent','head']}},indent=2)+'\n')
print('focus child_exit=0',flush=True)

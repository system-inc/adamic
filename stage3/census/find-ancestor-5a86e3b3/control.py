"""One frozen-input entry first-error pairing, then requested focused unit inventories."""
import hashlib,json,os,subprocess,sys,time
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2];adapted=Path('/tmp/fa-adapted')
parent='72f23939c773ba0045aeed5d6afb9b025b6473a0';head='5a86e3b3588914eb367f800eb7339db8d1bd70f9'

def hashes():
 return {str(p.relative_to(adapted)):hashlib.sha256(p.read_bytes()).hexdigest() for d in ['src/compiler','src/tsc'] for p in sorted((adapted/d).rglob('*')) if p.is_file() and p.suffix in ['.ts','.a']}
def git(*args):return subprocess.check_output(['git',*args],cwd=repo,text=True).strip()
def checkpoint(event):
 subprocess.run(['git','add','-f','--',str(out)],cwd=repo,check=True)
 tree=git('write-tree')
 sha=subprocess.check_output(['git','commit-tree',tree,'-p',parent],cwd=repo,input='census: find ancestor '+event+' checkpoint\n',text=True).strip()
 ref='refs/heads/typescript/find-ancestor-census-checkpoint-'+event
 with (out/('checkpoint-'+event+'.log')).open('w') as f:rc=subprocess.run(['git','push','origin',sha+':'+ref],cwd=repo,stdout=f,stderr=subprocess.STDOUT).returncode
 print('checkpoint',event,'child_exit='+str(rc),sha,flush=True)
 assert rc==0
 (out/('checkpoint-'+event+'.json')).write_text(json.dumps(dict(sha=sha,ref=ref,exit=rc),indent=2)+'\n')

assert json.loads((out/'apply.status.json').read_text())['exit']==0
assert json.loads((out/'parent-build.status.json').read_text())['exit']==0
assert json.loads((out/'head-build.status.json').read_text())['exit']==0
before=hashes();(out/'source-hashes.json').write_text(json.dumps(before,indent=2)+'\n')
(out/'patch-set.md').write_bytes((adapted/'patch-set.md').read_bytes())
provenance={}
for side,tree,sha in [('parent',repo,parent),('head',Path('/tmp/fa-head'),head)]:
 assert subprocess.check_output(['git','rev-parse','HEAD'],cwd=tree,text=True).strip()==sha
 assert subprocess.check_output(['git','rev-parse',sha+':cohere'],cwd=tree,text=True).strip()==subprocess.check_output(['git','rev-parse','HEAD'],cwd='/tmp/fa-report/cohere',text=True).strip()
 scopes=['internal/lower','internal/load','stage3/census/latent','stage3/meter/entry.go.txt','stage3/meter/entry_overlay.py']
 assert subprocess.run(['git','diff','--quiet',sha,'--',*scopes],cwd=tree).returncode==0
 files=subprocess.check_output(['git','ls-tree','-r','--name-only',sha,'--',*scopes],cwd=tree,text=True).splitlines()
 code={f:hashlib.sha256((tree/f).read_bytes()).hexdigest() for f in files if (f.endswith(('.go','.go.txt','.py','.cjs')) and '/data/' not in f and '/evidence/' not in f and '/testdata/' not in f)}
 provenance[side]=dict(sha=sha,source_hashes=code,binary_sha256=hashlib.sha256(Path('/tmp/fa-'+side+'-census').read_bytes()).hexdigest())
 meta=json.loads(Path('/tmp/fa-'+side+'-entry/overlay.json').read_text());dest=out/'overlays'/side;dest.mkdir(parents=True,exist_ok=True);rows=[]
 for i,(logical,actual) in enumerate(sorted(meta['Replace'].items())):
  name=f'{i:02d}-'+Path(actual).name
  if name.endswith('.go'):name+='.txt'
  data=Path(actual).read_bytes();(dest/name).write_bytes(data)
  rows.append(dict(logical=logical,evidence=str((dest/name).relative_to(out)),sha256=hashlib.sha256(data).hexdigest()))
 (dest/'manifest.json').write_text(json.dumps(rows,indent=2)+'\n')
 (out/(side+'-workspace.txt')).write_bytes(Path('/tmp/fa-'+side+'.work').read_bytes())
(out/'provenance.json').write_text(json.dumps(provenance,indent=2)+'\n')
(out/'REPORT.md').write_text('# Find ancestor census checkpoint\n\nTask 448a51k. Parent '+parent+'; head '+head+'.\n\nThe parent apply completed successfully. Both guarded entry drivers are built from their own pins. Source hashes and overlay evidence are retained. The first-error comparison is pending; no regression verdict is claimed. No full-core or second pairing is run. Focused unit inventories follow the entry pass, as in pairing A.\n\nCheckpoint refs are separate so the final branch can contain one commit on the parent without force-pushing.\n')
checkpoint('apply')
# No measured source is removed. Only unused upstream test fixtures are omitted.
import shutil
shutil.rmtree(adapted/'tests',ignore_errors=True)

def launch(side):
 name=side+'-run';env=['env','-u','LATENT_FULL','-u','LATENT_SPECULATIVE','-u','LATENT_ONLY_FILE','LATENT_ASSERT_NO_OUTPUT=1','LATENT_TIMES='+str(out/(name+'-units.jsonl'))]
 cmd=[sys.executable,str(out/'run.py'),name,str(repo),*env,'/tmp/fa-'+side+'-census',str(adapted/'src/tsc/tsc.ts'),str(out/(name+'.jsonl'))]
 print('launch',side,flush=True)
 return subprocess.Popen(cmd,cwd=repo)
p=launch('parent');h=None;started=time.monotonic();last_print=0
while p.poll() is None or h is None or h.poll() is None:
 if h is None:
  f=out/'parent-run-units.jsonl';rows=[]
  if f.exists():
   for s in f.read_text().splitlines():
    try:rows.append(json.loads(s))
    except json.JSONDecodeError:pass
  files=len({r['file'] for r in rows})
  stat={k:int(v) for k,v,*_ in (line.split() for line in Path('/sys/fs/cgroup/memory.stat').read_text().splitlines())}
  if p.poll() is not None or ((files>=20 or time.monotonic()-started>=900) and stat['anon']+stat['shmem']<10*1024**3):
   if p.poll() not in (None,0):raise SystemExit(p.returncode)
   (out/'head-start.json').write_text(json.dumps(dict(parent_timed_files=files,parent_running=p.poll() is None,anon_plus_shmem=stat['anon']+stat['shmem']),indent=2)+'\n')
   h=launch('head')
 if time.monotonic()-last_print>60:
  print('entry runs active; parent_exit='+str(p.poll())+' head_exit='+str(h.poll() if h else 'not started'),flush=True);last_print=time.monotonic()
 time.sleep(2)
assert p.wait()==0 and h.wait()==0
assert hashes()==before,'adapted program sources changed'
for side in ['parent','head']:
 rows=[json.loads(s) for s in (out/(side+'-run.jsonl')).read_text().splitlines()]
 assert rows[0]['status']=='measurement' and rows[0]['latent_mode']=='first-error' and len(rows)==1+len(rows[0]['sources'])
 print(side,'files='+str(len(rows)-1),'checker_diagnostics='+str(len(rows[0]['diagnostics'])),flush=True)
cmd=['python3',str(repo/'stage3/census/pair/pair.py'),str(out/'parent-run.jsonl'),str(out/'head-run.jsonl'),str(adapted),'--typescript',str(repo/'stage3/api/node_modules/typescript'),'--json',str(out/'paired.json'),'--report',str(out/'paired.md')]
rc=subprocess.run([sys.executable,str(out/'run.py'),'pair',str(repo),*cmd]).returncode
assert rc==0
print('pair complete',flush=True)

rc=subprocess.run([sys.executable,str(out/"focus.py")]).returncode
assert rc==0

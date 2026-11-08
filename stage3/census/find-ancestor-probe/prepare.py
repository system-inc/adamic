import os,sys,subprocess,json,time,signal,hashlib,re
from pathlib import Path
out=Path('/workspace/scratch/probe-report/stage3/census/find-ancestor-probe')
out.mkdir(parents=True,exist_ok=True)
repo=out.parents[2]
def run(name,cmd,cwd,env=None):
 start=time.monotonic()
 with (out/(name+'.log')).open('w') as log:
  p=subprocess.Popen(cmd,cwd=cwd,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  last=time.monotonic();stamp=None;stop=None
  while p.poll() is None:
   stat=(out/(name+'.log')).stat();cur=(stat.st_size,stat.st_mtime_ns)
   if cur!=stamp:last=time.monotonic();stamp=cur
   if time.monotonic()-last>1800:
    stop='silent for 1800 seconds';os.killpg(p.pid,signal.SIGTERM)
    try:p.wait(timeout=10)
    except subprocess.TimeoutExpired:os.killpg(p.pid,signal.SIGKILL)
    break
   time.sleep(1)
  rc=p.wait()
  try:os.killpg(p.pid,signal.SIGKILL)
  except ProcessLookupError:pass
 (out/(name+'.status.json')).write_text(json.dumps(dict(command=cmd,cwd=str(cwd),exit=rc,wall_seconds=time.monotonic()-start,stopped=stop),indent=2)+'\n')
 print(name,'child_exit='+str(rc),flush=True)
 if rc:raise SystemExit(rc)
# npm on the apply tree, exactly as requested.
a=Path('/workspace/scratch/probe-apply')
run('npm',['npm','ci','--prefix','stage3/api'],a)
assert subprocess.check_output(['git','rev-parse','HEAD'],cwd='/workspace/scratch/typescript-6.0.3',text=True).strip()=='050880ce59e30b356b686bd3144efe24f875ebc8'
os.environ['ADAMIC_TYPESCRIPT_SOURCE']='/workspace/scratch/typescript-6.0.3'
run('apply',['bash','stage3/apply.sh','/workspace/scratch/fa-adapted'],a)
adapted=Path('/workspace/scratch/fa-adapted')
actual={str(p.relative_to(adapted)):hashlib.sha256(p.read_bytes()).hexdigest() for d in ['src/compiler','src/tsc'] for p in sorted((adapted/d).rglob('*')) if p.is_file() and p.suffix in ['.ts','.a']}
expected=json.loads(Path('/workspace/scratch/probe-evidence/source-hashes.json').read_text())
(out/'source-hashes.json').write_text(json.dumps(actual,indent=2)+'\n')
(out/'source-validation.json').write_text(json.dumps(dict(matches=actual==expected,files=len(actual)),indent=2)+'\n')
if actual!=expected:raise SystemExit('STOP: source hash mismatch')
print('source hashes match',len(actual),flush=True)
for side in ['parent','head']:
 tree=Path('/workspace/scratch/probe-'+side)
 (tree/'cohere').rmdir();(tree/'cohere').symlink_to('/workspace/adamic/cohere',target_is_directory=True)
 work=Path('/workspace/scratch/probe-'+side+'.work')
 pairs=re.findall(r'^\s*(\S+) => (\./cohere\S*)',(tree/'go.mod').read_text(),re.M)
 replacements={m:str(tree/p) for m,p in pairs if m!='github.com/microsoft/TypeScript/tsc'}
 work.write_text('go 1.27\n\ntoolchain go1.27.1\n\nuse (\n'+str(tree)+'\n'+str(tree/'cohere/TypeScript/tsc')+'\n)\n\nreplace (\n'+''.join(m+' => '+p+'\n' for m,p in replacements.items())+')\n')
 env=dict(os.environ,GOWORK=str(work))
 overlay=Path('/workspace/scratch/probe-'+side+'-overlay');entry=Path('/workspace/scratch/probe-'+side+'-entry')
 run(side+'-overlay',['python3','stage3/census/latent/make_overlay.py',str(tree),str(overlay)],tree,env)
 run(side+'-entry-overlay',['python3','stage3/meter/entry_overlay.py',str(overlay),str(entry)],tree,env)
 run(side+'-entry-build',['go','build','-p=2','-buildvcs=false','-overlay='+str(entry/'overlay.json'),'-o','/workspace/scratch/probe-'+side+'-entry-bin','./stage3/census/latent/tool'],tree,env)
 manifest=[]
 for i,(logical,actual) in enumerate(sorted(json.loads((entry/'overlay.json').read_text())['Replace'].items())):
  dest=out/'overlays'/side;dest.mkdir(parents=True,exist_ok=True)
  name=f'{i:02d}-'+Path(actual).name
  if name.endswith('.go'):name+='.txt'
  data=Path(actual).read_bytes();(dest/name).write_bytes(data)
  manifest.append(dict(logical=logical,evidence=name,sha256=hashlib.sha256(data).hexdigest()))
 (dest/'manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
 (out/(side+'-workspace.txt')).write_bytes(work.read_bytes())
print('prepare complete child_exit=0',flush=True)

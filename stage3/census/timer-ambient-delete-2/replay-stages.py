import hashlib,json,os,subprocess,sys,tarfile,tempfile
from pathlib import Path
repo=Path(sys.argv[1]);tree=Path(sys.argv[2]);results=[]
def snapshot(root):
 return {str(p.relative_to(root)):(hashlib.sha256(p.read_bytes()).hexdigest(),p.stat().st_mtime_ns) for p in (root/'src').rglob('*') if p.is_file()}
for stage in ['41-explicit-any-remaining','93-builtin-imports','99-wasm-crypto-import','99-correctness-timer-closures']:
 with tempfile.TemporaryDirectory(prefix='road5-replay-') as scratch:
  root=Path(scratch)
  with (root/'source.tar').open('wb') as stream:
   subprocess.run(['git','-C',str(tree),'archive','refs/stage3/'+stage+'-after','src','scripts','package.json'],stdout=stream,check=True)
  with tarfile.open(root/'source.tar') as archive:archive.extractall(root,filter='data')
  before=snapshot(root)
  env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))
  r=subprocess.run(['node',str(repo/'stage3/adapt'/stage/'adapt.cjs'),str(root)],text=True,capture_output=True,env=env)
  assert r.returncode==0,(stage,r.stdout,r.stderr)
  assert snapshot(root)==before,stage
  results.append(dict(stage=stage,child_exit=r.returncode,source_files=len(before),bytesAndMtimesUnchanged=True,stdout=r.stdout,stderr=r.stderr))
print(json.dumps(results,indent=2))

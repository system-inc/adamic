"""Replay every amended stage on its exact recorded after tree without writes."""
import hashlib,io,json,os,pathlib,subprocess,tarfile,tempfile
repo=pathlib.Path('/workspace/adamic');tree=pathlib.Path('/workspace/scratch/timer-final')
output=pathlib.Path(__file__).resolve().parent
stages=['41-explicit-any-remaining','93-builtin-imports','99-correctness-axolotl-x','99-wasm-crypto-import']
env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))
results=[]
for stage in stages:
 with tempfile.TemporaryDirectory(prefix='timer-replay-',dir='/workspace/scratch') as temporary:
  root=pathlib.Path(temporary)
  raw=subprocess.check_output(['git','-C',str(tree),'archive','refs/stage3/'+stage+'-after','src'])
  with tarfile.open(fileobj=io.BytesIO(raw)) as archive:archive.extractall(root,filter='data')
  def snapshot():return {str(p.relative_to(root)):(hashlib.sha256(p.read_bytes()).hexdigest(),p.stat().st_mtime_ns) for p in (root/'src').rglob('*') if p.is_file()}
  before=snapshot()
  child=subprocess.run(['node',str(repo/'stage3/adapt'/stage/'adapt.cjs'),str(root)],env=env,text=True,capture_output=True)
  (output/('replay-'+stage+'.stdout')).write_text(child.stdout)
  (output/('replay-'+stage+'.stderr')).write_text(child.stderr)
  assert child.returncode==0,(stage,child.returncode,child.stderr)
  assert snapshot()==before,stage
  result=json.loads(child.stdout)
  assert all(result.get(key,0)==0 for key in ['files','edits','rewrites']),(stage,result)
  if isinstance(result.get('sites'),int):assert result['sites']==0,(stage,result)
  results.append(dict(stage=stage,child_exit=child.returncode,sourceFiles=len(before),bytesAndMtimesPreserved=True,result=result))
(output/'replay-stages.json').write_text(json.dumps(results,indent=2)+'\n')
print(json.dumps(results,indent=2))

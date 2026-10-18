from pathlib import Path
import importlib.util,subprocess,json,sys,time,gzip,difflib,hashlib
root=Path.cwd();ev=root/'review/compiler/loom-r-views-fx-a-ac'
spec=importlib.util.spec_from_file_location('views_a_gate','/tmp/views-ac-gate.py');gate=importlib.util.module_from_spec(spec);sys.modules[spec.name]=gate;spec.loader.exec_module(gate)
Path(str(root)+'-binaries').mkdir(exist_ok=True)
class Process:
 def __init__(self,p,path,h):self.p=p;self.path=path;self.h=h
 def communicate(self):
  out,err=self.p.communicate(timeout=45);self.returncode=self.p.returncode
  self.h.diagnostics[self.path]=dict(exit=self.returncode,stderr=err,stdout_bytes=len(out.encode()))
  return out,err
class Harness:
 def __init__(self,label):
  self.label=label;self.arguments=type('Args',(),{'tree':str(root)})();self.result={};self.steps={};self.exits={};self.failure=None;self.diagnostics={}
 def step(self,name,cmd):
  start=time.monotonic()
  with (ev/(self.label+'-build.log')).open('w') as log:
   result=subprocess.run(cmd,cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=90)
  self.steps[name]=round(time.monotonic()-start,4)
  return result.returncode==0
 def spawn(self,cmd,stdout,stderr):return Process(subprocess.Popen(cmd,cwd=root,stdout=stdout,stderr=stderr,text=True),cmd[-1],self)
 def fail(self,name,detail):self.failure=detail

def run(label,paths):
 h=Harness(label);gate.Gate.aCheck(h,paths)
 row=dict(label=label,result=h.result,steps=h.steps,exits=h.exits,failure=h.failure,diagnostics=h.diagnostics)
 (ev/(label+'.json')).write_text(json.dumps(row,indent=2)+'\n')
 print(label,h.exits,h.failure or 'PASS',flush=True);return row
paths=json.loads((ev/'paths.json').read_text());rows=json.loads((ev/'statuses.json').read_text())
known=[x['path'] for x in rows if x['status']=='failed']
full=run('predicate-full',paths)
assert full['exits']['a-check']==(1 if known else 0)
if known:
 assert f'for {len(known)} files:' in full['failure']
 assert all(path+': failed, expected checked' in full['failure'] for path in known)
healthy=[p for p in paths if p not in known]
assert run('predicate-accepted',healthy)['exits']['a-check']==0
path='internal/lower/testdata/loom_views_a/key.a';file=root/path;original=file.read_bytes();assert original.startswith(b'// a-check: refused ')
mutants=[]
try:
 for name,changed in [('removed-header',original.split(b'\n',1)[1]),('wrong-reason',b'// a-check: refused intentionally wrong reason for header predicate mutant\n'+original.split(b'\n',1)[1])]:
  file.write_bytes(changed)
  patch=''.join(difflib.unified_diff(original.decode().splitlines(True),changed.decode().splitlines(True),fromfile=path,tofile=path))
  (ev/(name+'.patch.gz')).write_bytes(gzip.compress(patch.encode(),mtime=0))
  result=run('mutant-'+name,[path]);assert result['exits']['a-check']==1 and 'expected' in result['failure']
  assert result['result']['a_check'][path]['outcome']=='refused'
  mutants.append(dict(mutant=name,path=path,caught=True,exit=result['exits']['a-check'],failure=result['failure']))
  file.write_bytes(original)
finally:file.write_bytes(original)
(ev/'header-mutants.json').write_text(json.dumps(mutants,indent=2)+'\n')
final=run('predicate-final',paths);assert final['exits']==full['exits']
for row in rows:
 now=(root/row['path']).read_bytes();body=now.split(b'\n',1)[1] if now.startswith(b'// a-check:') else now
 assert hashlib.sha256(body).hexdigest()==row['body_sha256_before']
(ev/'pinned-gate.py.gz').write_bytes(gzip.compress(Path('/tmp/views-ac-gate.py').read_bytes(),mtime=0))
(ev/'gate-provenance.json').write_text(json.dumps(dict(commit='fbac28c62493f27a788edc02a18bc8edb68de5da',path='cloud/fast-gate/run.py',sha256=hashlib.sha256(Path('/tmp/views-ac-gate.py').read_bytes()).hexdigest(),method='Gate.aCheck',modified=False),indent=2)+'\n')

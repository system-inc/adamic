import pathlib,subprocess,importlib.util,json,sys,hashlib,re,io,builtins
sys.dont_write_bytecode=True
out=pathlib.Path(__file__).resolve().parent;root=out.parents[2];binary=out/'adamic-acheck'
spec=importlib.util.spec_from_file_location('pinned_gate',out/'pinned-gate.py');module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
paths=subprocess.check_output(['git','diff','--name-only','origin/main','--','*.a'],cwd=root,text=True).splitlines()
class Process:
 def __init__(self,p,path,records):self.p=p;self.path=path;self.records=records
 def communicate(self):
  stdout,stderr=self.p.communicate(timeout=95);self.returncode=self.p.returncode;self.records[self.path]=dict(exit=self.returncode,stdout=stdout,stderr=stderr);return stdout,stderr
class Harness:
 def __init__(self,replay=None):self.arguments=type('Args',(),dict(tree=str(root)))();self.result={};self.steps={};self.exits={};self.failure=None;self.records={};self.replay=replay
 def step(self,name,command):
  if self.replay is not None:return True
  command=list(command);command[3]=str(binary)
  with(out/'acheck-build.log').open('w')as log:r=subprocess.run(['timeout','180',*command],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  assert r.returncode==0;return True
 def spawn(self,command,stdout,stderr):
  if self.replay is not None:
   row=self.replay[command[-1]]
   class Recorded:
    returncode=row['exit']
    def communicate(self):return row['stdout'],row['stderr']
   return Recorded()
  command=list(command);command[0]=str(binary);return Process(subprocess.Popen(['timeout','90',*command],cwd=root,stdout=stdout,stderr=stderr,text=True),command[-1],self.records)
 def fail(self,name,detail):self.failure=detail
h=Harness();module.Gate.aCheck(h,paths);assert h.exits['a-check']==0,h.failure
(out/'a-check.json').write_text(json.dumps(dict(paths=paths,exits=h.exits,results=h.result,diagnostics=h.records),indent=2)+'\n');statuses=[]
for path in paths:
 data=(root/path).read_bytes();assert data==subprocess.check_output(['git','show','2dad6ae36c3b:'+path],cwd=root);diag=h.records[path];errors=diag['stderr'];status='checked'if diag['exit']==0 else'NotYet'if"can't lower"in errors and' yet'in errors else'refused'if'Adamic 0.1 refuses'in errors else'type error'if' error TS'in errors else'failed'
 statuses.append(dict(path=path,status=status,header=data.splitlines()[0].decode(),diagnostic=errors,body_sha256=hashlib.sha256(data.split(b'\n',1)[1] if data.startswith(b'// a-check:') else data).hexdigest()))
(out/'a-check-status.json').write_text(json.dumps(statuses,indent=2)+'\n');mutants=[];selected=next(r['path'] for r in statuses if r['status']=='refused');path=root/selected;original=path.read_bytes()
for name,changed in [('removed-header',original.split(b'\n',1)[1]),('wrong-reason',b'// a-check: refused deliberately incorrect checked-write reason\n'+original.split(b'\n',1)[1])]:
 try:
  module.open=lambda name,*args,**kwargs: io.StringIO(changed.decode()) if pathlib.Path(name)==path else builtins.open(name,*args,**kwargs);probe=Harness(h.records);module.Gate.aCheck(probe,[selected]);assert probe.exits['a-check']==1;mutants.append(dict(mutant=name,catcher='unchanged Gate.aCheck using captured actual branch diagnostics',exit=1,failure=probe.failure))
 finally:module.open=builtins.open
restored=Harness(h.records);module.Gate.aCheck(restored,paths);assert restored.exits['a-check']==0
(out/'header-mutants.json').write_text(json.dumps(mutants,indent=2)+'\n');(out/'predicate-provenance.json').write_text(json.dumps(dict(source='fbac28c62493f27a788edc02a18bc8edb68de5da:cloud/fast-gate/run.py',sha256=hashlib.sha256((out/'pinned-gate.py').read_bytes()).hexdigest(),predicate_unchanged=True,binary_output_redirected_to_evidence=True),indent=2)+'\n')
print('PASS actual Gate.aCheck',len(paths),'files:',[r['status']for r in statuses],'; bodies byte-identical to source; both header mutants caught; restored predicate passes')

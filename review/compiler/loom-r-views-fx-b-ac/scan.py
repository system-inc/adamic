import collections, hashlib, importlib.util, json, os, pathlib, re, subprocess, sys, time
root=pathlib.Path.cwd(); out=root/'review/compiler/loom-r-views-fx-b-ac'
spec=importlib.util.spec_from_file_location('pinned_gate',out/'pinned-gate.py'); gate=importlib.util.module_from_spec(spec);spec.loader.exec_module(gate)
paths=subprocess.check_output(['git','diff','--name-only','--diff-filter=AM','dbdf977f5b','--','*.a'],text=True).splitlines()
assert len(paths)==17,paths
originals={p:(root/p).read_bytes() for p in paths}
class Process:
 def __init__(self,h,cmd):self.h=h;self.cmd=cmd
 def communicate(self):
  start=time.monotonic();path=self.cmd[-1]
  if self.h.replay is None:
   p=subprocess.run(self.cmd,cwd=root,capture_output=True,text=True,timeout=30)
   self.returncode=p.returncode;stdout,stderr=p.stdout,p.stderr
  else:
   row=self.h.replay[path];self.returncode=row['exit'];stdout,stderr='',row['stderr']
  self.h.diagnostics[path]={'command':self.cmd,'exit':self.returncode,'stderr':stderr,'stdout_bytes':len(stdout.encode()),'seconds':round(time.monotonic()-start,4)}
  return stdout,stderr
class Harness:
 def __init__(self,label,replay=None):
  self.arguments=type('Args',(),{'tree':str(root)})();self.result={};self.steps={};self.exits={};self.failure=None;self.label=label;self.diagnostics={};self.replay=replay
 def step(self,name,cmd):
  if self.replay is not None:return True
  with (out/(self.label+'-build.log')).open('w') as f: code=subprocess.run(cmd,cwd=root,stdout=f,stderr=subprocess.STDOUT,timeout=120).returncode
  assert code==0,code
  return True
 def spawn(self,cmd,stdout,stderr):return Process(self,cmd)
 def fail(self,name,detail):self.failure=detail
 def run(self,selected):
  gate.Gate.aCheck(self,selected)
  rows={p:dict(**v,**self.diagnostics[p]) for p,v in self.result['a_check'].items()}
  result={'paths':selected,'rows':rows,'failure':self.failure,'exits':self.exits,'steps':self.steps}
  (out/(self.label+'.json')).write_text(json.dumps(result,indent=2)+'\n');return result
before=Harness('before').run(paths)
statuses=[]
for path in paths:
 row=before['rows'][path];errors=row['stderr'];status='checked';reason='';header=None
 if row['exit']==0:pass
 elif "can't lower" in errors and ' yet' in errors:status='NotYet';reason=errors.strip()
 elif 'Adamic 0.1 refuses' in errors:
  status='refused';reason=re.search(r'Adamic 0\.1 refuses ([^\n;]+)',errors).group(1);header='// a-check: refused '+reason
 elif ' error TS' in errors:
  status='type error';reason=re.search(r' error (TS\d+)',errors).group(1);header='// a-check: type error '+reason
 else:raise RuntimeError((path,row))
 body=originals[path]
 if header:
  assert not body.startswith(b'// a-check:'),path
  (root/path).write_bytes(header.encode()+b'\n'+body)
 actual=(root/path).read_bytes();below=actual.split(b'\n',1)[1] if header else actual
 assert below==body,path
 statuses.append({'path':path,'status':status,'reason':reason,'header_added':header,'body_sha256_before':hashlib.sha256(body).hexdigest(),'body_sha256_after':hashlib.sha256(below).hexdigest()})
(out/'statuses.json').write_text(json.dumps(statuses,indent=2)+'\n')
after=Harness('after').run(paths);assert after['exits']['a-check']==0,after['failure']
negative=next((x for x in statuses if x['header_added']),None)
if negative is None:
 path='stage3/fixtures/entries/06_symbol_key.a'
 control=Harness('mutant-control').run([path]);assert control['exits']['a-check']==0,control['failure']
 replay=control['rows'];restored=(root/path).read_bytes();body=restored.split(b'\n',1)[1]
else:
 path=negative['path'];replay=after['rows'];restored=(root/path).read_bytes();body=originals[path]
mutants=[]
for name,data in [('removed-header',body),('wrong-reason',b'// a-check: refused deliberate wrong reason mutant\n'+body)]:
 try:
  (root/path).write_bytes(data)
  result=Harness(name,replay).run([path]);caught=result['exits']['a-check']==1
  mutants.append({'mutant':name,'path':path,'caught':caught,'failure':result['failure'],'predicate':'unchanged Gate.aCheck at fbac28c62493f27a788edc02a18bc8edb68de5da','diagnostics':'captured real compiler output after headers'})
  assert caught,result
 finally:(root/path).write_bytes(restored)
assert (root/path).read_bytes()==restored
(out/'header-mutants.json').write_text(json.dumps(mutants,indent=2)+'\n')
final=Harness('final').run(paths);assert final['exits']['a-check']==0,final['failure']
for x in statuses:
 actual=(root/x['path']).read_bytes();body=actual.split(b'\n',1)[1] if x['header_added'] else actual
 assert hashlib.sha256(body).hexdigest()==x['body_sha256_before'],x['path']
print(dict(collections.Counter(x['status'] for x in statuses)), 'headers',sum(bool(x['header_added']) for x in statuses),'predicate PASS; both mutants caught; final real-compiler scan PASS',flush=True)
for x in statuses:print(x['path'],x['status'],x['header_added'] or '',flush=True)

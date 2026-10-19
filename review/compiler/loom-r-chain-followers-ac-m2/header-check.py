import collections,hashlib,importlib.util,json,subprocess,time
from pathlib import Path
from types import SimpleNamespace
root=Path.cwd();out=root/'review/compiler/loom-r-chain-followers-ac-m2';source='09aa4ee0aad721c1a027b15eba5b23344a588a83'
paths=subprocess.check_output(['git','diff','--name-only',source+'~1',source,'--','*.a'],text=True).splitlines();assert len(paths)==21
spec=importlib.util.spec_from_file_location('pinned_gate',root/'review/compiler/loom-r-chain-followers-ac/gate.py');mod=importlib.util.module_from_spec(spec);spec.loader.exec_module(mod)
observations={};commands=[]
class Process:
 def __init__(self,command):
  self.command=command;self.start=time.monotonic();self.process=subprocess.Popen(command,cwd=root,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
 def communicate(self):
  stdout,stderr=self.process.communicate(timeout=90);self.returncode=self.process.returncode
  path=self.command[-1];observations[path]=dict(exit=self.returncode,stdout=stdout,stderr=stderr,seconds=round(time.monotonic()-self.start,3),command=self.command)
  (out/'compiler-logs').mkdir(exist_ok=True);(out/'compiler-logs'/(path.replace('/','__')+'.log')).write_text(stderr)
  return stdout,stderr
class Harness:
 def __init__(self):self.arguments=SimpleNamespace(tree=str(root));self.result={};self.steps={};self.exits={};self.failure=None
 def step(self,name,command):
  assert name=='a-check-build';start=time.monotonic()
  with (out/'compiler-build.log').open('w') as stream:done=subprocess.run(command,cwd=root,stdout=stream,stderr=subprocess.STDOUT,timeout=180)
  commands.append(dict(label=name,command=command,exit=done.returncode,seconds=round(time.monotonic()-start,3)));assert done.returncode==0;return True
 def spawn(self,command,stdout,stderr):return Process(command)
 def fail(self,name,detail):self.failure=detail
before={p:(root/p).read_bytes() for p in paths}
h=Harness();mod.Gate.aCheck(h,paths)
rows=[]
for path in paths:
 result=observations[path];text=result['stderr'];data=(root/path).read_bytes();assert data==before[path]
 original=subprocess.check_output(['git','show',source+':'+path]);assert original==data
 header=data.split(b'\n',1)[0].decode();body=data.split(b'\n',1)[1] if header.startswith('// a-check:') else data
 if result['exit']==0:status='checked';reason='';code=''
 elif "can't lower" in text and ' yet' in text:status='NotYet';reason=text.strip();code=''
 elif 'Adamic 0.1 refuses ' in text:status='refused';reason=text.split('Adamic 0.1 refuses ',1)[1].splitlines()[0];code=''
 elif ' error TS' in text:
  import re
  status='type error';code=re.search(r'error (TS\d+)',text).group(1);reason=text.strip()
 else:raise AssertionError((path,result))
 rows.append(dict(path=path,status=status,reason=reason,code=code,header=header if header.startswith('// a-check:') else '',header_added=False,header_changed=False,body_sha256=hashlib.sha256(body).hexdigest(),file_sha256=hashlib.sha256(data).hexdigest(),stdout_sha256=hashlib.sha256(result['stdout'].encode()).hexdigest(),stdout_bytes=len(result['stdout'].encode()),stderr=text,exit=result['exit'],seconds=result['seconds'],command=result['command'],predicate=h.result['a_check'][path]))
(out/'statuses.json').write_text(json.dumps(rows,indent=2));(out/'statuses.tsv').write_text('path\tstatus\tcode\treason\theader\n'+''.join('\t'.join([r['path'],r['status'],r['code'],r['reason'].replace('\n',' | '),r['header']])+'\n' for r in rows))
(out/'predicate.json').write_text(json.dumps(dict(gate_commit='fbac28c62493f27a788edc02a18bc8edb68de5da',exit=h.exits['a-check'],result=h.result,failure=h.failure),indent=2));(out/'commands.json').write_text(json.dumps(commands,indent=2))
print('statuses',dict(collections.Counter(r['status'] for r in rows)),'predicate',h.exits['a-check'],flush=True)
assert h.exits['a-check']==0,h.failure
# Proof uses the exact unchanged predicate and captured real diagnostics, as in the README.
selected=next(r for r in rows if r['status']=='refused' and r['path'].startswith('review/'));path=root/selected['path'];original=path.read_bytes();first,body=original.split(b'\n',1)
class Captured:
 def __init__(self,row):self.row=row;self.returncode=row['exit']
 def communicate(self):return self.row['stdout'],self.row['stderr']
class Replay(Harness):
 def step(self,name,command):return True
 def spawn(self,command,stdout,stderr):return Captured(observations[selected['path']])
mutants=[]
try:
 for mode,data in [('removed-header',body),('wrong-reason',b'// a-check: refused deliberately-unmatched-chain-followers-reason\n'+body)]:
  path.write_bytes(data);test=Replay();mod.Gate.aCheck(test,[selected['path']]);assert test.exits['a-check']==1 and test.failure,(mode,test.result)
  mutants.append(dict(path=selected['path'],mutation=mode,exit=test.exits['a-check'],caught_by='unchanged pinned Gate.aCheck replaying the actual branch compiler diagnostic',failure=test.failure,result=test.result))
finally:path.write_bytes(original)
assert path.read_bytes()==original
(out/'header-mutants.json').write_text(json.dumps(mutants,indent=2))
(out/'headers.json').write_text(json.dumps(dict(newly_added=[],changed=[],retained=[dict(path=r['path'],header=r['header']) for r in rows if r['header']],body_bytes_unchanged=True,source=source),indent=2))
print('Both removed-header and wrong-reason mutants caught; all original bytes restored; no header edits required.',flush=True)

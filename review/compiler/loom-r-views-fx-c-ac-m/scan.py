import sys
sys.dont_write_bytecode=True
import importlib.util,json,pathlib,subprocess,sys,time,collections,hashlib,re
ROOT=pathlib.Path.cwd(); E=ROOT/'review/compiler/loom-r-views-fx-c-ac-m'
spec=importlib.util.spec_from_file_location('pinned_gate',ROOT/'review/compiler/loom-r-views-fx-c-ac/gate.py');gate=importlib.util.module_from_spec(spec);spec.loader.exec_module(gate)
phase=sys.argv[1];diagnostics={};paths=json.loads((E/'paths.json').read_text())
class Process:
 def __init__(self,command):
  command=['/tmp/views-fx-c-adamic',*command[1:]];self.command=command;self.p=subprocess.Popen(command,cwd=ROOT,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
 def communicate(self):
  try:out,err=self.p.communicate(timeout=45)
  except subprocess.TimeoutExpired:self.p.kill();self.p.communicate();raise
  self.returncode=self.p.returncode;diagnostics[self.command[-1]]=dict(exit=self.returncode,stderr=err,stdout_bytes=len(out.encode()))
  return out,err
class Harness:
 def __init__(self):
  self.arguments=type('Args',(),dict(tree=str(ROOT)))();self.result={};self.steps={};self.exits={};self.failure=None
 def step(self,name,command):
  assert name=='a-check-build' and command[-1]=='./cmd/adamic'
  # Already built by validate.py; Process invokes that exact branch compiler.
  return True
 def spawn(self,command,stdout,stderr):return Process(command)
 def fail(self,name,detail):self.failure=detail
h=Harness();gate.Gate.aCheck(h,paths)
rows=[]
for path in paths:
 data=(ROOT/path).read_bytes();d=diagnostics[path];err=d['stderr'];reason='';code=''
 if d['exit']==0:status='checked'
 elif "can't lower" in err and ' yet' in err:status='NotYet';reason=err.strip()
 elif 'Adamic 0.1 refuses ' in err:status='refused';reason=err.split('Adamic 0.1 refuses ',1)[1].split(';',1)[0]
 elif re.search(r' error (TS\d+)',err):status='type error';code=re.search(r' error (TS\d+)',err)[1]
 else:raise AssertionError((path,d))
 rows.append(dict(path=path,status=status,reason=reason,code=code,sha256=hashlib.sha256(data).hexdigest(),diagnostic=d,predicate=h.result['a_check'][path]))
(E/(phase+'.json')).write_text(json.dumps(rows,indent=2)+'\n');print(collections.Counter(r['status'] for r in rows));print('Gate.aCheck exit',h.exits['a-check']);print(h.failure or 'PASS')
if phase=='final':assert h.exits['a-check']==0

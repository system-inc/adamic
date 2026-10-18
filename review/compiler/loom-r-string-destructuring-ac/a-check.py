import collections, hashlib, importlib.machinery, json, pathlib, subprocess, sys, time
sys.dont_write_bytecode=True
root=pathlib.Path.cwd(); out=pathlib.Path(__file__).resolve().parent
reference=out/'gate-reference.py.txt'
gate=importlib.machinery.SourceFileLoader('pinned_fast_gate',str(reference)).load_module()
paths=subprocess.check_output(['git','diff','--name-only','dbdf977f5b','c7302ad4','--','*.a'],text=True).splitlines()
class Process:
 def __init__(self,command,**kw):
  self.path=command[-1];self.process=subprocess.Popen(command,cwd=root,**kw,text=True)
 def communicate(self):
  stdout,stderr=self.process.communicate(timeout=60);self.returncode=self.process.returncode
  diagnostics[self.path]={'exit':self.returncode,'stderr':stderr,'stdoutBytes':len(stdout.encode())}
  return stdout,stderr
class Harness:
 def __init__(self,label):
  self.label=label;self.arguments=type('Args',(),{'tree':str(root)})();self.result={};self.steps={};self.exits={};self.failure=None
 def step(self,name,command):
  with (out/(self.label+'-build.log.txt')).open('w') as log:
   code=subprocess.run(command,cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=90).returncode
  assert code==0,code
  return True
 def spawn(self,command,stdout,stderr):return Process(command,stdout=stdout,stderr=stderr)
 def fail(self,name,detail):self.failure=detail

def run(label,files,want):
 global diagnostics
 diagnostics={};h=Harness(label);start=time.monotonic();gate.Gate.aCheck(h,files)
 for path,row in h.result['a_check'].items():row.update(diagnostics[path])
 result={'paths':h.result['a_check'],'exit':h.exits['a-check'],'failure':h.failure,'seconds':round(time.monotonic()-start,3)}
 (out/(label+'.json')).write_text(json.dumps(result,indent=2)+'\n')
 assert result['exit']==want,result
 print(label,result['exit'],result['failure'] or 'PASS',flush=True)
 return result
run('a-check-branch',paths,0)
control='internal/oracle/testdata/non_null_refused/initialized.a'; file=root/control
original=file.read_bytes();header,body=original.split(b'\n',1)
assert header.startswith(b'// a-check: refused ')
run('control-header',[control],0)
try:
 file.write_bytes(body)
 run('removed-header-mutant',[control],1)
 file.write_bytes(b'// a-check: refused deliberately-wrong-reason\n'+body)
 run('wrong-reason-mutant',[control],1)
finally:file.write_bytes(original)
assert file.read_bytes()==original
(out/'body-preservation.json').write_text(json.dumps({'changedPrograms':{path:{'sourceBlob':subprocess.check_output(['git','rev-parse','c7302ad4:'+path],text=True).strip(),'sha256':hashlib.sha256((root/path).read_bytes()).hexdigest(),'unchanged':subprocess.check_output(['git','show','c7302ad4:'+path])==(root/path).read_bytes()} for path in paths},'control':control,'controlRestored':True,'controlBodySha256':hashlib.sha256(body).hexdigest(),'headersAdded':[]},indent=2)+'\n')

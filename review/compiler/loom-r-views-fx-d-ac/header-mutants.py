import difflib, hashlib, importlib.util, json, os, pathlib, subprocess, tempfile, time
root=pathlib.Path.cwd();ev=root/'review/compiler/loom-r-views-fx-d-ac'
spec=importlib.util.spec_from_file_location('pinned_gate',ev/'pinned-gate.py');gate=importlib.util.module_from_spec(spec);spec.loader.exec_module(gate)
source=root/'stage3/census/repro/r37/main.a';original=source.read_bytes();rows=[]
assert original.startswith(b'// a-check: refused delete\n')
class Process:
 def __init__(self,command):self.command=command
 def communicate(self):
  result=subprocess.run(['timeout','55',*self.command],stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
  self.returncode=result.returncode;self.errors=result.stderr
  assert self.returncode!=124
  return result.stdout,result.stderr
class Harness:
 def __init__(self,label):self.arguments=type('Args',(),{'tree':str(root)})();self.result={};self.steps={};self.exits={};self.failure=None;self.label=label;self.diagnostics=[]
 def step(self,name,command):
  with (ev/(self.label+'-build.log')).open('w') as log:result=subprocess.run(['timeout','180',*command],stdout=log,stderr=subprocess.STDOUT)
  assert result.returncode==0;return True
 def spawn(self,command,stdout,stderr):
  p=Process(command);self.diagnostics.append(p);return p
 def fail(self,name,detail):self.failure=detail
with tempfile.TemporaryDirectory(prefix='views-d-a-check-',dir=os.environ['TMPDIR']) as directory:
 path=pathlib.Path(directory)/'control.a'
 for label,data,want in [('control',original,0),('removed-header',original.split(b'\n',1)[1],1),('wrong-reason',original.replace(b'delete',b'deliberately wrong refusal reason',1),1)]:
  path.write_bytes(data)
  if label!='control':(ev/(label+'.patch')).write_text(''.join(difflib.unified_diff(original.decode().splitlines(True),data.decode().splitlines(True),fromfile='a/control.a',tofile='b/control.a')))
  started=time.monotonic();h=Harness(label);gate.Gate.aCheck(h,[str(path)])
  row=dict(name=label,predicate_exit=h.exits['a-check'],seconds=round(time.monotonic()-started,4),result=list(h.result['a_check'].values())[0],compiler_exit=h.diagnostics[0].returncode,stderr=h.diagnostics[0].errors,failure=h.failure)
  assert row['predicate_exit']==want,row
  assert row['compiler_exit']!=0 and 'Adamic 0.1 refuses' in row['stderr'],row
  rows.append(row);print(json.dumps(row),flush=True)
assert source.read_bytes()==original
(ev/'header-mutants.json').write_text(json.dumps(dict(source='stage3/census/repro/r37/main.a',unchanged_source_sha256=hashlib.sha256(original).hexdigest(),temporary_copy=True,rows=rows),indent=2)+'\n')

"""Invoke unchanged Gate.aCheck on this scout and prove each header can fail."""
import importlib.util,json,pathlib,subprocess,sys
root=pathlib.Path(__file__).resolve().parent;repo=root.parents[3]
spec=importlib.util.spec_from_file_location('fastgate',sys.argv[1]);gate=importlib.util.module_from_spec(spec);spec.loader.exec_module(gate)
output=pathlib.Path(sys.argv[2]);output.mkdir(parents=True,exist_ok=True)
class Harness:
 def __init__(self,label):
  self.arguments=type('Args',(),{'tree':str(repo)})();self.result={};self.steps={};self.exits={};self.failure=None;self.label=label
 def step(self,name,cmd):
  with (output/(self.label+'-build.log')).open('w') as log:r=subprocess.run(cmd,cwd=repo,stdout=log,stderr=subprocess.STDOUT)
  assert r.returncode==0;return True
 def spawn(self,cmd,stdout,stderr):return subprocess.Popen(cmd,cwd=repo,stdout=stdout,stderr=stderr,text=True)
 def fail(self,name,detail):self.failure=detail
paths=sorted(str(p.relative_to(repo)) for p in (root/'fixtures').glob('*.a'))
def check(label,paths):
 h=Harness(label);gate.Gate.aCheck(h,paths)
 record=dict(exit=h.exits['a-check'],results=h.result['a_check'],failure=h.failure)
 (output/(label+'.json')).write_text(json.dumps(record,indent=2)+'\n');return record
baseline=check('baseline',paths);assert baseline['exit']==0,baseline
mutants=[]
for i,name in enumerate(paths):
 p=repo/name;original=p.read_bytes();first,body=original.split(b'\n',1);assert first.startswith(b'// a-check: refused ')
 for mode in ['missing','wrong']:
  try:
   p.write_bytes(body if mode=='missing' else b'// a-check: refused unrelated rule\n'+body)
   record=check(str(i)+'-'+mode,[name]);assert record['exit']==1 and 'expected' in record['failure']
   mutants.append(dict(file=name,change=mode+' header',exit=record['exit']))
  finally:p.write_bytes(original)
final=check('final',paths);assert final['exit']==0,final
(root/'evidence/a-check.json').write_text(json.dumps(dict(gate='fbac28c62493f27a788edc02a18bc8edb68de5da',baseline=baseline,mutants=mutants,final=final),indent=2)+'\n')
print('PASS: a-check 3/3; six real missing/wrong-header mutants caught; restored final check 3/3')

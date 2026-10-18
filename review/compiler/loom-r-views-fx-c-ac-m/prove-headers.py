import sys
sys.dont_write_bytecode=True
import pathlib,json,importlib.util,hashlib
ROOT=pathlib.Path.cwd();E=ROOT/'review/compiler/loom-r-views-fx-c-ac-m'
spec=importlib.util.spec_from_file_location('gate',ROOT/'review/compiler/loom-r-views-fx-c-ac/gate.py');gate=importlib.util.module_from_spec(spec);spec.loader.exec_module(gate)
row=next(r for r in json.loads((E/'final.json').read_text()) if r['path'].endswith('0587b8e6f528/fxspptb_oct9_views_p22_callable_union_fewer_parameters.a'))
p=ROOT/row['path'];original=p.read_bytes();body=original.split(b'\n',1)[1]
class Captured:
 returncode=row['diagnostic']['exit']
 def communicate(self):return '',row['diagnostic']['stderr']
class Harness:
 def __init__(self):self.arguments=type('Args',(),dict(tree=str(ROOT)))();self.result={};self.steps={};self.exits={};self.failure=None
 def step(self,*args):return True
 def spawn(self,*args,**kwargs):return Captured()
 def fail(self,name,detail):self.failure=detail
results=[]
for name,source in [('removed-header',body),('wrong-reason',b'// a-check: refused deliberately wrong reason\n'+body)]:
 try:
  p.write_bytes(source);h=Harness();gate.Gate.aCheck(h,[row['path']]);assert h.exits['a-check']==1 and h.failure
  results.append(dict(mutant=name,path=row['path'],exit=h.exits['a-check'],catcher=h.failure,body_sha256=hashlib.sha256(body).hexdigest()))
  print(name,'CAUGHT',h.failure)
 finally:p.write_bytes(original)
assert p.read_bytes()==original
(E/'header-mutants.json').write_text(json.dumps(results,indent=2)+'\n')

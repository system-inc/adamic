import importlib.util,json,pathlib,subprocess,time,types
root=pathlib.Path.cwd();out=root/'review/compiler/loom-r-views-fx-b-ac-m'
spec=importlib.util.spec_from_file_location('gate',root/'review/compiler/loom-r-views-fx-b-ac/pinned-gate.py');gate=importlib.util.module_from_spec(spec);spec.loader.exec_module(gate)
paths=subprocess.check_output(['git','diff','HEAD','--name-only','--','*.a'],text=True).splitlines()
paths=sorted(set(p for p in paths if (root/p).is_file())|set(str(p) for p in (root/'internal/oracle/testdata/review/refused').glob('fxspptb_oct9_views_p[046]*.a') if p.name in {'fxspptb_oct9_views_p07_element_access_union_wrong.a','fxspptb_oct9_views_p49_union_destructure_wrong.a'}))
paths=[str(pathlib.Path(p).relative_to(root)) if p.startswith(str(root)) else p for p in paths]
# Include moved/untracked refused sources before staging.
paths=sorted(set(paths)|{'internal/lower/testdata/fx6_refused/p49.a','internal/oracle/testdata/review/refused/fxspptb_oct9_views_p07_element_access_union_wrong.a','internal/oracle/testdata/review/refused/fxspptb_oct9_views_p49_union_destructure_wrong.a'})
pathlib.Path('/tmp/views-b-gate-tree-binaries').mkdir(exist_ok=True)
class Process:
 def __init__(self,h,cmd):self.h=h;self.cmd=cmd
 def communicate(self):
  start=time.monotonic();p=subprocess.run(self.cmd,cwd=root,capture_output=True,text=True,timeout=30);self.returncode=p.returncode
  self.h.diagnostics[self.cmd[-1]]={'exit':p.returncode,'stderr':p.stderr,'seconds':time.monotonic()-start}
  return p.stdout,p.stderr
class Harness:
 def __init__(self,label):self.arguments=types.SimpleNamespace(tree='/tmp/views-b-gate-tree');self.result={};self.steps={};self.exits={};self.failure=None;self.diagnostics={};self.label=label
 def step(self,name,cmd):
  if self.label!='final':return True
  with (out/'a-check-build.log').open('w') as f:p=subprocess.run(cmd,cwd=root,stdout=f,stderr=subprocess.STDOUT,timeout=120)
  assert p.returncode==0;return True
 def spawn(self,cmd,stdout,stderr):return Process(self,cmd)
 def fail(self,name,detail):self.failure=detail
 def run(self,paths):
  gate.Gate.aCheck(self,paths);data={'result':self.result,'exits':self.exits,'failure':self.failure,'diagnostics':self.diagnostics};(out/('a-check-'+self.label+'.json')).write_text(json.dumps(data,indent=2)+'\n');return self.exits['a-check']
h=Harness('final');assert h.run(paths)==0,h.failure
p=root/'internal/lower/testdata/fx6_refused/p49.a';original=p.read_bytes();body=original.split(b'\n',1)[1]
for name,source in [('removed-header',body),('wrong-reason',b'// a-check: refused wrong mutant reason\n'+body)]:
 try:
  p.write_bytes(source);m=Harness(name);assert m.run([str(p.relative_to(root))])==1
 finally:p.write_bytes(original)
print('a-check PASS:',len(paths),'files; two real-compiler header mutants caught')

from pathlib import Path
import types,sys,gzip,json,subprocess,hashlib,time
root=Path.cwd();out=root/'review/compiler/loom-r-views-fx-a-ac-m'
source=gzip.decompress((root/'review/compiler/loom-r-views-fx-a-ac/pinned-gate.py.gz').read_bytes())
gate=types.ModuleType('views_a_unchanged_gate');gate.__file__=str(out/'pinned-gate.py');sys.modules[gate.__name__]=gate;exec(compile(source,gate.__file__,'exec'),gate.__dict__)
class Process:
 def __init__(self,h,cmd):self.h=h;self.cmd=cmd
 def communicate(self):
  p=subprocess.run(self.cmd,cwd=root,capture_output=True,text=True,timeout=30);self.returncode=p.returncode;self.h.diagnostics[self.cmd[-1]]={'exit':p.returncode,'stderr':p.stderr};return p.stdout,p.stderr
class Harness:
 def __init__(self,label):self.arguments=types.SimpleNamespace(tree=str(root));self.result={};self.steps={};self.exits={};self.failure=None;self.diagnostics={};self.label=label
 def step(self,name,cmd):
  if self.label!='full':return True
  with (out/'a-check-build.log').open('w') as f:p=subprocess.run(cmd,cwd=root,stdout=f,stderr=subprocess.STDOUT,timeout=120)
  return p.returncode==0
 def spawn(self,cmd,stdout,stderr):return Process(self,cmd)
 def fail(self,name,detail):self.failure=detail
 def run(self,paths):
  gate.Gate.aCheck(self,paths);d={'exits':self.exits,'result':self.result,'failure':self.failure,'diagnostics':self.diagnostics};(out/('a-check-'+self.label+'.json')).write_text(json.dumps(d,indent=2)+'\n');return self.exits['a-check']
paths=subprocess.check_output(['git','diff','HEAD','--name-only','--','*.a'],text=True).splitlines();paths=[p for p in paths if (root/p).is_file()];moved='internal/oracle/testdata/review/refused/fxspptb_oct9_views_p07_element_access_union_wrong.a';paths=sorted(set(paths)|{row['to'] for row in json.loads((out/'moved-programs.json').read_text())})
h=Harness('full');assert h.run(paths)==0,h.failure
rows=[]
for path,diagnostic in h.diagnostics.items():
 s=diagnostic['stderr'];status='checked' if diagnostic['exit']==0 else 'NotYet' if "can't lower" in s and ' yet' in s else 'refused' if 'Adamic 0.1 refuses' in s else 'type error' if ' error TS' in s else 'failed';rows.append({'path':path,'status':status,**diagnostic})
(out/'program-statuses.json').write_text(json.dumps(rows,indent=2)+'\n')
p=root/moved;original=p.read_bytes();body=original.split(b'\n',1)[1]
for name,data in [('removed-header',body),('wrong-reason',b'// a-check: refused wrong mutant reason\n'+body)]:
 try:p.write_bytes(data);m=Harness(name);assert m.run([moved])==1,m.failure
 finally:p.write_bytes(original)
assert hashlib.sha256(body).hexdigest()==json.loads((out/'moved-programs.json').read_text())[0]['body_sha256']
(out/'a-check-provenance.json').write_text(json.dumps({'gate_source_commit':'fbac28c62493f27a788edc02a18bc8edb68de5da','sha256':hashlib.sha256(source).hexdigest(),'modified':False,'method':'Gate.aCheck','files':len(paths),'counts':{s:sum(r['status']==s for r in rows) for s in ['checked','NotYet','refused','type error','failed']},'header_mutants_caught':2},indent=2)+'\n')
print('PASS',len(paths),'files; removed-header and wrong-reason caught')
# Inspect even pending agree witnesses; a pending sidecar must not hide a new refusal.
binary=str(root)+'-binaries/adamic-acheck';remaining=[]
for p in sorted((root/'internal/oracle/testdata/review/agree').glob('*.a')):
 q=subprocess.run([binary,'c',str(p)],capture_output=True,text=True,timeout=30)
 if 'Adamic 0.1 refuses' in q.stderr:remaining.append({'path':str(p.relative_to(root)),'diagnostic':q.stderr,'pending':p.with_suffix('.pending').exists()})
(out/'agree-refusal-scan.json').write_text(json.dumps(remaining,indent=2)+'\n');print('Remaining agree refusals:',len(remaining))

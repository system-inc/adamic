import collections,hashlib,json,runpy,subprocess,time
from pathlib import Path
from types import SimpleNamespace
root=Path.cwd();e=root/'review/compiler/loom-r-v4-2-ac';Gate=runpy.run_path('/tmp/v4-ac-gate.py')['Gate'];diagnostics={}
class Process:
 def __init__(self,process,path):self.process=process;self.path=path
 def communicate(self):
  output,error=self.process.communicate(timeout=60);self.returncode=self.process.returncode
  diagnostics[self.path]={'exit':self.returncode,'stderr':error,'stdout_bytes':len(output.encode())}
  return output,error
class Harness:
 def __init__(self,captured=None):
  self.arguments=SimpleNamespace(tree=str(root));self.result={};self.steps={};self.exits={};self.failure=None;self.captured=captured
 def step(self,name,command):
  assert name=='a-check-build'
  # Reuse the branch compiler already built by the recorded go build command.
  assert Path('/tmp/adamic-v4-ac').is_file()
  return True
 def spawn(self,command,stdout,stderr):
  path=command[-1]
  if self.captured is not None:
   capture=self.captured[path]
   class Captured:
    returncode=capture['exit']
    def communicate(self):return '',capture['stderr']
   return Captured()
  return Process(subprocess.Popen(['/tmp/adamic-v4-ac',*command[1:]],cwd=root,stdout=stdout,stderr=stderr,text=True),path)
 def fail(self,name,detail):self.failure=detail
paths=(e/'paths.txt').read_text().splitlines();baseline=Harness();Gate.aCheck(baseline,paths);assert baseline.exits['a-check']==0,baseline.failure
(e/'gate-baseline.json').write_text(json.dumps(dict(results=baseline.result,steps=baseline.steps,exits=baseline.exits,diagnostics=diagnostics),indent=2)+'\n')
selected='internal/oracle/testdata/views_v4_refused/view-result-wrong.a';p=root/selected;original=p.read_bytes();header,body=original.split(b'\n',1);mutants=[]
try:
 for name,contents in [('removed-header',body),('wrong-reason',b'// a-check: refused deliberately unmatched V4 reason\n'+body)]:
  p.write_bytes(contents);h=Harness(diagnostics);Gate.aCheck(h,[selected]);assert h.exits['a-check']==1 and h.failure,(name,h.result)
  mutants.append(dict(path=selected,mutant=name,exit=h.exits['a-check'],caught_by='unchanged Gate.aCheck',failure=h.failure))
finally:p.write_bytes(original)
(e/'header-mutants.json').write_text(json.dumps(mutants,indent=2)+'\n')
rows=[]
for r in json.loads((e/'statuses.json').read_text()):
 p=root/r['path'];current=p.read_bytes();assert hashlib.sha256(current).hexdigest()==r['sha256'];r['body_unchanged']=True;r['header']=r['first_line'] if r['first_line'].startswith('// a-check:') else None;rows.append(r)
(e/'statuses.json').write_text(json.dumps(rows,indent=2)+'\n');summary=dict(files=len(paths),statuses=dict(collections.Counter(r['status'] for r in rows)),headers_added=0,existing_headers=sum(r['header'] is not None for r in rows),predicate_exit=baseline.exits['a-check'],mutants_caught=len(mutants),gate_commit='fbac28c62493f27a788edc02a18bc8edb68de5da',gate_source_sha256=hashlib.sha256(Path('/tmp/v4-ac-gate.py').read_bytes()).hexdigest());(e/'summary.json').write_text(json.dumps(summary,indent=2)+'\n');print(json.dumps(summary,indent=2))

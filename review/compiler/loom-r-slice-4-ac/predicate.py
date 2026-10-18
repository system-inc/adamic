import importlib.util,json,subprocess,types,hashlib
from pathlib import Path
root=Path.cwd();out=root/'review/compiler/loom-r-slice-4-ac'
spec=importlib.util.spec_from_file_location('pinned_gate',out/'gate.py');gate=importlib.util.module_from_spec(spec);spec.loader.exec_module(gate)
rows=json.loads((out/'statuses.json').read_text());records={r['path']:r for r in rows}
class Process:
 def __init__(self,path):self.row=records[path];self.returncode=self.row['exit']
 def communicate(self):return '',(root/self.row['diagnostics']).read_text()
class Harness:
 def __init__(self):self.arguments=types.SimpleNamespace(tree=str(root));self.result={};self.steps={};self.exits={};self.failure=None
 def step(self,name,command):assert name=='a-check-build';return True
 def spawn(self,command,*args):return Process(command[-1])
 def fail(self,name,detail):self.failure=detail
h=Harness();gate.Gate.aCheck(h,list(records));assert h.exits['a-check']==0,h.failure
results=[dict(kind='all real compiler diagnostics',files=len(records),exit=0,result=h.result)]
p=Path(next(r['path'] for r in rows if r['status']=='refused'));original=p.read_bytes();body=original.split(b'\n',1)[1]
try:
 for name,contents in [('removed-header',body),('wrong-reason',b'// a-check: refused mutant reason that is absent from compiler diagnostics\n'+body)]:
  p.write_bytes(contents);h=Harness();gate.Gate.aCheck(h,[str(p)]);assert h.exits['a-check']==1 and 'expected' in h.failure
  results.append(dict(mutant=name,path=str(p),exit=h.exits['a-check'],caught=True,catcher=h.failure))
finally:p.write_bytes(original)
assert p.read_bytes()==original
(out/'predicate-results.json').write_text(json.dumps(results,indent=2)+'\n');print('82 accepted; removed-header and wrong-reason each caught; source restored')

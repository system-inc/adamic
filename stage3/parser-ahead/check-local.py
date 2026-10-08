#!/usr/bin/env python3
"""Run pinned Gate.aCheck; write local pre-push evidence without claiming the full gate."""
import hashlib, importlib.util, json, os, pathlib, subprocess, sys
root=pathlib.Path(__file__).resolve().parents[2]
spec=importlib.util.spec_from_file_location('fastgate','/tmp/step24-ahead-gate.py')
mod=importlib.util.module_from_spec(spec);spec.loader.exec_module(mod)
class Harness:
 def __init__(self):
  self.arguments=type('Args',(),{'tree':str(root)})();self.result={};self.steps={};self.exits={};self.failure=None
 def step(self,name,cmd):
  with open('/tmp/step24-ahead-'+name+'.log','w') as log:
   code=subprocess.run(cmd,cwd=root,stdout=log,stderr=subprocess.STDOUT).returncode
  return code==0
 def spawn(self,cmd,stdout,stderr):return subprocess.Popen(cmd,cwd=root,stdout=stdout,stderr=stderr,text=True)
 def fail(self,name,detail):self.failure=detail
paths=sorted(str(p.relative_to(root)) for p in (root/'stage3/parser-ahead').rglob('*.a'))
h=Harness();mod.Gate.aCheck(h,paths)
record={'head':subprocess.check_output(['git','rev-parse','HEAD'],cwd=root,text=True).strip(),
 'a_check':h.result.get('a_check'), 'exits':h.exits,'failure':h.failure,
 'scope':'changed .a only; pre-push record, not the full fast gate',
 'input_sha256':{p:hashlib.sha256((root/p).read_bytes()).hexdigest() for p in paths},
 'worktree_status':subprocess.check_output(['git','status','--short'],cwd=root,text=True)}
pathlib.Path(os.environ.get('STEP24_RECORD',str(root/'stage3/parser-ahead/evidence/pre-push.json'))).write_text(json.dumps(record,indent=2)+'\n')
print(json.dumps(record,indent=2));sys.exit(h.exits['a-check'])

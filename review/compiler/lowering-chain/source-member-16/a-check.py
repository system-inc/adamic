import importlib.util
import json
from pathlib import Path
import subprocess
import sys
import time

root=Path.cwd()
evidence=root/'review/compiler-generators-main'
spec=importlib.util.spec_from_file_location('fastgate','/tmp/generators-fast-gate.py')
module=importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)
class Harness:
 def __init__(self):
  self.arguments=type('Args',(),{'tree':str(root)})()
  self.result={};self.steps={};self.exits={};self.failure=None
 def step(self,name,command):
  with (evidence/(name+'.log')).open('w') as log:
   code=subprocess.run(command,cwd=root,stdout=log,stderr=subprocess.STDOUT).returncode
  return code==0
 def spawn(self,command,stdout,stderr):
  return subprocess.Popen(command,cwd=root,stdout=stdout,stderr=stderr,text=True)
 def fail(self,name,detail):self.failure=detail
paths=sorted(str(p.relative_to(root)) for p in (root/'internal/oracle/testdata/generators').glob('*.a'))+['internal/oracle/testdata/notyet/generator_cycle.a']
harness=Harness()
module.Gate.aCheck(harness,paths)
(evidence/'a-check.json').write_text(json.dumps(harness.result,indent=2)+'\n')
print(harness.failure or 'PASS',len(paths),harness.steps,harness.exits)
sys.exit(harness.exits['a-check'])

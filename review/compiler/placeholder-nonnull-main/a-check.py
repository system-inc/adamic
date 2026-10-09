"""Use integration's a-check implementation on changed Adamic fixtures only."""
import importlib.util,json,pathlib,subprocess,types,sys
sys.dont_write_bytecode=True
root=pathlib.Path(__file__).resolve().parents[3]
evidence=pathlib.Path(__file__).resolve().parent
source=subprocess.check_output(['git','show','origin/devtools/fast-gate:cloud/fast-gate/run.py'],cwd=root)
runner=evidence/'gate-run.py.txt';runner.write_bytes(source)
from importlib.machinery import SourceFileLoader
module=types.ModuleType('gate');SourceFileLoader('gate',str(runner)).exec_module(module)
gate=module.Gate.__new__(module.Gate)
gate.arguments=types.SimpleNamespace(tree=str(root));gate.exits={};gate.steps={};gate.result={}
def step(name,command):
 with (evidence/'logs'/(name+'.log')).open('w') as log:
  return subprocess.run(command,cwd=root,stdout=log,stderr=subprocess.STDOUT).returncode==0
gate.step=step
gate.spawn=lambda command,out,err:subprocess.Popen(command,cwd=root,stdout=out,stderr=err,text=True)
gate.fail=lambda name,message:print(message)
paths=subprocess.check_output(['git','diff','--name-only','origin/main','--','*.a'],cwd=root,text=True).splitlines()
gate.aCheck(paths)
(evidence/'a-check-results.json').write_text(json.dumps(gate.result,indent=2)+'\n')
print(len(paths),'changed .a files;',gate.exits,gate.steps)
assert gate.exits['a-check']==0

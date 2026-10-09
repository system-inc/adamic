import importlib.util, pathlib, subprocess, json, sys
root=pathlib.Path.cwd()
# Pass a saved cloud/fast-gate/run.py from the documented tooling pin.
spec=importlib.util.spec_from_file_location('fastgate',sys.argv[1]); mod=importlib.util.module_from_spec(spec);spec.loader.exec_module(mod)
class Harness:
 def __init__(self):self.arguments=type('Args',(),{'tree':str(root)})();self.result={};self.steps={};self.exits={};self.failure=None
 def step(self,name,cmd):
  with open('/tmp/overload-hatch-acheck-build.log','w') as log:r=subprocess.run(cmd,cwd=root,stdout=log,stderr=subprocess.STDOUT)
  return r.returncode==0
 def spawn(self,cmd,stdout,stderr):return subprocess.Popen(cmd,cwd=root,stdout=stdout,stderr=stderr,text=True)
 def fail(self,name,detail):self.failure=detail
h=Harness();paths=sorted(str(p.relative_to(root)) for p in (root/'internal/oracle/testdata/overload_field_hatch').glob('*.a'));mod.Gate.aCheck(h,paths)
pathlib.Path('docs/overload-results/groups/field-hatch/a-check.json').write_text(json.dumps(h.result,indent=2)+'\n')
print(h.result);print(h.failure or 'PASS');sys.exit(h.exits['a-check'])

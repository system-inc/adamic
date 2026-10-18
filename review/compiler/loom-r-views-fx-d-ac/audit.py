import collections, hashlib, importlib.util, json, pathlib, subprocess, sys, time
root=pathlib.Path.cwd();ev=root/'review/compiler/loom-r-views-fx-d-ac'
spec=importlib.util.spec_from_file_location('pinned_gate',ev/'pinned-gate.py');gate=importlib.util.module_from_spec(spec);spec.loader.exec_module(gate)
paths=subprocess.check_output(['git','diff','--name-only','dbdf977f5b','45d2ef9ac165fd78b791df03d61d9c058de0b16b','--','*.a'],text=True).splitlines()
label=sys.argv[1] if len(sys.argv)>1 else 'initial';observations={}
class Process:
 def __init__(self,command):self.command=command;self.path=command[-1]
 def communicate(self):
  started=time.monotonic();p=subprocess.run(['timeout','55',*self.command],cwd=root,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
  self.returncode=p.returncode
  assert p.returncode!=124,('compiler timeout',self.path)
  body=(root/self.path).read_bytes()
  outcome='checked' if p.returncode==0 else 'NotYet' if "can't lower" in p.stderr and ' yet' in p.stderr else 'refused' if 'Adamic 0.1 refuses' in p.stderr else 'type error' if ' error TS' in p.stderr else 'failed'
  observations[self.path]=dict(status=outcome,exit=p.returncode,stderr=p.stderr,stdout_bytes=len(p.stdout.encode()),stdout_sha256=hashlib.sha256(p.stdout.encode()).hexdigest(),seconds=round(time.monotonic()-started,4),body_sha256=hashlib.sha256(body).hexdigest())
  return p.stdout,p.stderr
class Harness:
 def __init__(self):self.arguments=type('Args',(),{'tree':str(root)})();self.result={};self.steps={};self.exits={};self.failure=None
 def step(self,name,command):
  with (ev/(label+'-build.log')).open('w') as log:p=subprocess.run(['timeout','180',*command],cwd=root,stdout=log,stderr=subprocess.STDOUT)
  assert p.returncode==0,('build failed',p.returncode)
  return True
 def spawn(self,command,stdout,stderr):return Process(command)
 def fail(self,name,detail):self.failure=detail
h=Harness();gate.Gate.aCheck(h,paths)
for path,row in h.result['a_check'].items():row.update(observations[path])
(ev/(label+'-statuses.json')).write_text(json.dumps(h.result['a_check'],indent=2)+'\n')
summary=dict(files=len(paths),statuses=dict(collections.Counter(row['status'] for row in observations.values())),predicate_exit=h.exits['a-check'],seconds=h.steps['a-check'],failure=h.failure)
(ev/(label+'-summary.json')).write_text(json.dumps(summary,indent=2)+'\n');print(json.dumps(summary,indent=2));sys.exit(h.exits['a-check'])

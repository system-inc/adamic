"""Run the README-pinned, unchanged Gate.aCheck using this worktree's compiler."""
import collections, hashlib, importlib.util, json, pathlib, re, subprocess, sys, time
out=pathlib.Path(__file__).resolve().parent;root=out.parents[2]
spec=importlib.util.spec_from_file_location('slice5_gate',out/'gate-source.py');gate=importlib.util.module_from_spec(spec);spec.loader.exec_module(gate)
paths=[sys.argv[2]] if len(sys.argv)>2 else (out/'paths.txt').read_text().splitlines();observations={};label=sys.argv[1]
class Process:
 def __init__(self,cmd):self.cmd=cmd;self.started=time.monotonic();self.process=subprocess.Popen(cmd,cwd=root,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
 def communicate(self):
  try:stdout,stderr=self.process.communicate(timeout=45)
  except subprocess.TimeoutExpired:self.process.kill();self.process.communicate();raise
  self.returncode=self.process.returncode
  status='checked' if self.returncode==0 else 'NotYet' if "can't lower" in stderr and ' yet' in stderr else 'refused' if 'Adamic 0.1 refuses' in stderr else 'type error' if ' error TS' in stderr else 'failed'
  reason=stderr.split('Adamic 0.1 refuses ',1)[1].split('; ',1)[0].splitlines()[0] if status=='refused' else None
  code=re.search(r'error (TS\d+)',stderr).group(1) if status=='type error' else None
  observations[self.cmd[-1]]=dict(status=status,reason=reason,code=code,exit=self.returncode,stderr=stderr,stdout_bytes=len(stdout.encode()),stdout_sha256=hashlib.sha256(stdout.encode()).hexdigest(),seconds=time.monotonic()-self.started)
  return stdout,stderr
class Harness:
 def __init__(self):self.arguments=type('Args',(),{'tree':str(root)})();self.result={};self.steps={};self.exits={};self.failure=None
 def step(self,name,cmd):
  pathlib.Path(cmd[3]).parent.mkdir(exist_ok=True)
  with (out/'logs'/(label+'-build.log')).open('w') as f:r=subprocess.run(cmd,cwd=root,stdout=f,stderr=subprocess.STDOUT,timeout=120)
  return r.returncode==0
 def spawn(self,cmd,stdout,stderr):return Process(cmd)
 def fail(self,name,detail):self.failure=detail
h=Harness();gate.Gate.aCheck(h,paths)
record=dict(gate_sha='fbac28c62493f27a788edc02a18bc8edb68de5da',paths=observations,predicate=h.result.get('a_check'),exits=h.exits,failure=h.failure,seconds=h.steps,counts=dict(collections.Counter(r['status'] for r in observations.values())))
(out/(label+'.json')).write_text(json.dumps(record,indent=2)+'\n');print(record['counts'],h.exits);print(h.failure or 'PASS')
if label=='final':assert h.exits['a-check']==0

sys.exit(h.exits['a-check'])

from pathlib import Path
import subprocess, tempfile, types, json, time, os, hashlib
p=Path('review/compiler/loom-r-slice-7-ac'); root=Path.cwd()
pin='fbac28c62493f27a788edc02a18bc8edb68de5da'
source=subprocess.check_output(['git','show',pin+':cloud/fast-gate/run.py'])
module=types.ModuleType('pinned_fast_gate'); exec(compile(source,'pinned-fast-gate','exec'),module.__dict__)
class Harness:
 def __init__(self,label):
  self.arguments=types.SimpleNamespace(tree=str(root)); self.result={}; self.steps={}; self.exits={}; self.failure=None; self.label=label; self.processes=[]
 def step(self,name,cmd):
  with (p/(self.label+'-build.log')).open('wb') as log: r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,timeout=120)
  return r.returncode==0
 def spawn(self,cmd,stdout,stderr):
  process=subprocess.Popen(cmd,stdout=stdout,stderr=stderr,text=True); self.processes.append(process)
  class Process:
   def communicate(inner):
    try: output=process.communicate(timeout=60)
    except subprocess.TimeoutExpired:
     process.kill(); process.communicate(); raise
    inner.returncode=process.returncode; return output
  return Process()
 def fail(self,name,detail): self.failure=detail

def check(label,paths):
 h=Harness(label); module.Gate.aCheck(h,paths)
 r=dict(predicate_pin=pin,source_sha256=hashlib.sha256(source).hexdigest(),result=h.result,steps=h.steps,exits=h.exits,failure=h.failure)
 (p/(label+'.json')).write_text(json.dumps(r,indent=2)+'\n'); print(label,h.exits,h.failure or 'PASS',flush=True); return h
paths=[r['path'] for r in json.loads((p/'statuses.json').read_text())]
assert check('predicate-scan',paths).exits['a-check']==0
# All changed programs are NotYet and require no header. Use an independent
# temporary refusal control to prove both required header mutations can fail.
body=b"interface Dictionary { [key: string]: string; }\nconsole.log('control');\n"
(p/'refused-control.a.txt').write_bytes(body)
with tempfile.TemporaryDirectory(prefix='eep-header-control-') as temporary:
 f=Path(temporary)/'control.a'; f.write_bytes(body)
 r=subprocess.run([str(root)+'-binaries/adamic-acheck','c',str(f)],capture_output=True,timeout=60)
 diagnostic=r.stderr.decode(); assert r.returncode==1 and 'Adamic 0.1 refuses ' in diagnostic,diagnostic
 (p/'refused-control.stderr.log').write_text(diagnostic)
 reason=diagnostic.split('Adamic 0.1 refuses ',1)[1].splitlines()[0]
 header=('// a-check: refused '+reason+'\n').encode(); f.write_bytes(header+body)
 assert check('control-header',[str(f)]).exits['a-check']==0
 f.write_bytes(body)
 assert check('mutant-removed-header',[str(f)]).exits['a-check']==1
 f.write_bytes(b'// a-check: refused deliberately wrong reason\n'+body)
 assert check('mutant-wrong-reason',[str(f)]).exits['a-check']==1
 (p/'mutants.json').write_text(json.dumps(dict(control='refused-control.a.txt (temporary .a only)',reason=reason,control_header=header.decode().strip(),removed_header='caught',wrong_reason='caught',scope='Independent refusal control; no changed program needs a header'),indent=2)+'\n')

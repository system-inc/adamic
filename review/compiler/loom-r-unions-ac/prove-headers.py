import difflib,hashlib,json,os,subprocess,time
from pathlib import Path
root=Path.cwd();e=root/'review/compiler/loom-r-unions-ac';ns={'__name__':'unions_header_gate'};exec(compile((e/'gate.py.txt').read_bytes(),str(e/'gate.py.txt'),'exec'),ns);Gate=ns['Gate'];observations={};builds=[]
class Process:
 def __init__(self,cmd,label):self.cmd=cmd;self.label=label;self.started=time.monotonic();self.p=subprocess.Popen(cmd,cwd=root,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
 def communicate(self):
  out,err=self.p.communicate(timeout=60);self.returncode=self.p.returncode;name=self.label+'-'+self.cmd[-1].replace('/','-');(e/(name+'.stderr.log')).write_text(err);(e/(name+'.stdout.txt')).write_text(out)
  observations.setdefault(self.label,[]).append(dict(command=self.cmd,exit=self.returncode,seconds=round(time.monotonic()-self.started,3),stderr=err,stdout_bytes=len(out.encode())))
  return out,err
class Harness:
 def __init__(self,label):self.label=label;self.arguments=type('Args',(),{'tree':str(root)})();self.result={};self.steps={};self.exits={};self.failure=None
 def step(self,name,cmd):
  if builds:return True
  start=time.monotonic()
  with (e/'predicate-build.log').open('w') as out:r=subprocess.run(cmd,cwd=root,stdout=out,stderr=subprocess.STDOUT,timeout=180)
  builds.append(dict(command=cmd,seconds=round(time.monotonic()-start,3),exit=r.returncode));assert r.returncode==0;return True
 def spawn(self,cmd,stdout,stderr):return Process(cmd,self.label)
 def fail(self,name,detail):self.failure=detail
rows=[]
def check(label,paths,want):
 h=Harness(label);Gate.aCheck(h,paths);row=dict(label=label,result=h.result,exits=h.exits,seconds=h.steps['a-check'],failure=h.failure);rows.append(row);assert h.exits['a-check']==want,(label,h.failure);return row
paths=json.loads((e/'paths.json').read_text());check('changed-programs',paths,0)
path='internal/oracle/testdata/non_null_refused/initialized.a';p=root/path;original=p.read_bytes();header,body=original.split(b'\n',1);assert header==b'// a-check: refused the non-null assertion !'
check('control-header',[path],0)
try:
 for label,mutated in [('removed-header',body),('wrong-reason',b'// a-check: refused impossible union header mutant reason\n'+body)]:
  (e/(label+'.patch')).write_text(''.join(difflib.unified_diff(original.decode().splitlines(True),mutated.decode().splitlines(True),fromfile='a/'+path,tofile='b/'+path)))
  p.write_bytes(mutated);check(label,[path],1)
finally:p.write_bytes(original)
check('restored-header',[path],0);assert p.read_bytes()==original
(e/'predicate-results.json').write_text(json.dumps(dict(gate_commit='fbac28c62493f27a788edc02a18bc8edb68de5da',builds=builds,checks=rows,compiler_observations=observations,control_path=path,control_sha256=hashlib.sha256(original).hexdigest(),control_restored=True),indent=2)+'\n');print([(r['label'],r['exits']) for r in rows])

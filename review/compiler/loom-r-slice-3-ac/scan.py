import collections,hashlib,importlib.util,json,os,re,subprocess,sys,time
from pathlib import Path
root=Path.cwd(); ev=Path(__file__).resolve().parent
os.sched_setaffinity(0,sorted(os.sched_getaffinity(0))[:4])
os.environ['GOMAXPROCS']='4'
spec=importlib.util.spec_from_file_location('pinned_gate',ev/'gate.py'); gate=importlib.util.module_from_spec(spec);spec.loader.exec_module(gate)
paths=(ev/'paths.txt').read_text().splitlines()
class Process:
 def __init__(self,cmd,rows):
  self.cmd=cmd;self.rows=rows;self.returncode=None
 def communicate(self):
  start=time.monotonic();r=subprocess.run(['timeout','60',*self.cmd],cwd=root,capture_output=True,text=True,timeout=65);self.returncode=r.returncode
  text=r.stderr
  status='checked' if r.returncode==0 else 'NotYet' if "can't lower" in text and ' yet' in text else 'refused' if 'Adamic 0.1 refuses' in text else 'type error' if ' error TS' in text else 'failed'
  reason=text.split('Adamic 0.1 refuses ',1)[1].strip() if status=='refused' else None
  code=re.search(r' error (TS\d+)',text)
  self.rows[self.cmd[-1]]=dict(status=status,reason=reason,code=code[1] if code else None,exit=r.returncode,stderr=text,stdout_bytes=len(r.stdout.encode()),stdout_sha256=hashlib.sha256(r.stdout.encode()).hexdigest(),seconds=round(time.monotonic()-start,6))
  return r.stdout,r.stderr
class Harness:
 def __init__(self,label):
  self.arguments=type('Args',(),{'tree':str(root)})();self.result={};self.steps={};self.exits={};self.failure=None;self.rows={};self.label=label
 def step(self,name,cmd):
  # Only the build output location is redirected into the evidence tool directory.
  # Gate.aCheck's classification and predicate remain unchanged.
  (ev/'tools').mkdir(exist_ok=True)
  cmd=list(cmd);cmd[cmd.index('-o')+1]=str(ev/'tools/adamic')
  with (ev/(self.label+'-build.log')).open('w') as f:r=subprocess.run(['timeout','90',*cmd],cwd=root,stdout=f,stderr=subprocess.STDOUT,timeout=95)
  assert r.returncode==0,r.returncode
  return True
 def spawn(self,cmd,stdout,stderr):
  return Process([str(ev/'tools/adamic'),*cmd[1:]],self.rows)
 def fail(self,name,detail):self.failure=detail
 def run(self,chosen):
  gate.Gate.aCheck(self,chosen)
  for name,row in self.rows.items():row['predicate']=self.result['a_check'][name]
  result=dict(rows=self.rows,counts=dict(collections.Counter(r['status'] for r in self.rows.values())),exit=self.exits['a-check'],failure=self.failure,seconds=self.steps['a-check'])
  (ev/(self.label+'.json')).write_text(json.dumps(result,indent=2)+'\n')
  print(self.label,result['counts'],'predicate exit',result['exit'],flush=True)
  return result
mode=sys.argv[1]
if mode!='mutants':
 result=Harness(mode).run(paths)
 sys.exit(result['exit'])
original_rows=json.loads((ev/'after.json').read_text())['rows']
path=next(p for p,r in original_rows.items() if r['status']=='refused');p=root/path;original=p.read_bytes();header,body=original.split(b'\n',1);results=[]
try:
 for name,changed in [('removed-header',body),('wrong-reason',b'// a-check: refused intentionally wrong reason 999999\n'+body)]:
  p.write_bytes(changed)
  result=Harness(name).run([path]);assert result['exit']==1 and result['rows'][path]['status']=='refused'
  (ev/(name+'.a.txt')).write_bytes(changed)
  results.append(dict(name=name,path=path,caught=True,compiler_status='refused',predicate_exit=result['exit']))
finally:p.write_bytes(original)
(ev/'mutants.json').write_text(json.dumps(results,indent=2)+'\n')

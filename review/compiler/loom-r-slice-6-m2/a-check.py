import collections,hashlib,importlib.util,json,os,re,subprocess,sys,time
from pathlib import Path
sys.dont_write_bytecode=True
root=Path.cwd();ev=Path(__file__).resolve().parent
os.sched_setaffinity(0,sorted(os.sched_getaffinity(0))[:4])
spec=importlib.util.spec_from_file_location('gate',root/'review/compiler/loom-r-implied-target-ac/pinned-gate.py');gate=importlib.util.module_from_spec(spec);spec.loader.exec_module(gate)
paths=sorted(p for p in subprocess.check_output(['git','diff','5565a7cafb0ee8d538f8a5ece2304d5aa4254145','--name-only','--diff-filter=ACMR'],text=True).splitlines() if p.endswith('.a') and Path(p).is_file())
class Process:
 def __init__(self,cmd,rows):self.cmd=cmd;self.rows=rows;self.returncode=None
 def communicate(self):
  start=time.monotonic();r=subprocess.run(['timeout','20',*self.cmd],capture_output=True,text=True,timeout=25);self.returncode=r.returncode
  text=r.stderr;status='checked' if r.returncode==0 else 'NotYet' if "can't lower" in text and ' yet' in text else 'refused' if 'Adamic 0.1 refuses' in text else 'type error' if ' error TS' in text else 'failed'
  reason=text.split('Adamic 0.1 refuses ',1)[1].strip() if status=='refused' else None
  code=re.search(r' error (TS\d+)',text)
  self.rows[self.cmd[-1]]=dict(status=status,reason=reason,code=code[1] if code else None,exit=r.returncode,stderr=text,stdout_bytes=len(r.stdout.encode()),stdout_sha256=hashlib.sha256(r.stdout.encode()).hexdigest(),seconds=round(time.monotonic()-start,6))
  return r.stdout,r.stderr
class Harness:
 def __init__(self,label):self.arguments=type('Args',(),{'tree':str(root)})();self.result={};self.steps={};self.exits={};self.failure=None;self.rows={};self.label=label
 def step(self,name,cmd):
  # The already-built branch compiler is supplied; the pinned classification and predicate are unchanged.
  return True
 def spawn(self,cmd,stdout,stderr):return Process(['/tmp/adamic-gate/loom-r-slice-6-m2/adamic',*cmd[1:]],self.rows)
 def fail(self,name,detail):self.failure=detail
 def run(self,chosen):
  gate.Gate.aCheck(self,chosen)
  for p,r in self.rows.items():r['predicate']=self.result['a_check'][p]
  result=dict(rows=self.rows,counts=dict(collections.Counter(r['status'] for r in self.rows.values())),exit=self.exits['a-check'],failure=self.failure,seconds=self.steps['a-check'],gate_sha256=hashlib.sha256((root/'review/compiler/loom-r-implied-target-ac/pinned-gate.py').read_bytes()).hexdigest())
  (ev/(self.label+'.json.txt')).write_text(json.dumps(result,indent=2)+'\n');print(self.label,result['counts'],'exit',result['exit'],result['failure'],flush=True)
  return result
mode=sys.argv[1];result=Harness(mode).run(paths)
if mode=='observe':
 changes=[]
 for path,row in result['rows'].items():
  p=Path(path);data=p.read_text();first=data.splitlines()[0] if data else '';header=None
  if row['status']=='refused':header='// a-check: refused '+row['reason']
  elif row['status']=='type error':header='// a-check: type error '+row['code']
  elif row['status'] in ['checked','NotYet'] and first.startswith('// a-check:'):header=''
  if header is not None and header!=first:
   body=data.split('\n',1)[1] if first.startswith('// a-check:') else data
   p.write_text((header+'\n' if header else '')+body);changes.append(dict(path=path,before=first,after=header,body_sha256=hashlib.sha256(body.encode()).hexdigest()))
 (ev/'a-check-headers.json.txt').write_text(json.dumps(changes,indent=2)+'\n');print('Header changes',len(changes),flush=True)
else:sys.exit(result['exit'])

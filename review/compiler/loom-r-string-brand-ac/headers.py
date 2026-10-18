import collections,difflib,hashlib,importlib.util,json,pathlib,re,subprocess,time
root=pathlib.Path.cwd();out=root/'review/compiler/loom-r-string-brand-ac'
paths=subprocess.check_output(['git','diff','--name-only','dbdf977f5b','2194037a','--','*.a'],text=True).splitlines()
compiler='/tmp/string-brand-ac-adamic'
spec=importlib.util.spec_from_file_location('pinned_gate',out/'pinned-gate.py');module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
class Process:
 def __init__(self,cmd,label):self.cmd=cmd;self.label=label
 def communicate(self):
  r=subprocess.run([compiler,*self.cmd[1:]],capture_output=True,text=True,timeout=20)
  self.returncode=r.returncode
  key=self.cmd[-1];self.label.observations[key]={'exit':r.returncode,'stderr':r.stderr,'stdout_bytes':len(r.stdout.encode())}
  return r.stdout,r.stderr
class Harness:
 def __init__(self):self.arguments=type('Args',(),{'tree':str(root)})();self.result={};self.steps={};self.exits={};self.failure=None;self.observations={}
 def step(self,name,cmd):
  # The branch compiler was already built by the recorded bounded go build.
  assert name=='a-check-build' and cmd[-1]=='./cmd/adamic';return True
 def spawn(self,cmd,stdout,stderr):return Process(cmd,self)
 def fail(self,name,detail):self.failure=detail

def scan(label,files):
 h=Harness();module.Gate.aCheck(h,files)
 record={'exit':h.exits['a-check'],'results':h.result['a_check'],'observations':h.observations,'failure':h.failure}
 (out/(label+'.json')).write_text(json.dumps(record,indent=2)+'\n')
 print(label,record['exit'],flush=True);return record

before=scan('initial-scan',paths);statuses=[]
for name in paths:
 p=root/name;body=p.read_bytes();row=before['observations'][name];err=row['stderr'];header='';code=None;reason=None
 if row['exit']==0:status='checked'
 elif "can't lower" in err and ' yet' in err:status='NotYet';reason=err.strip()
 elif 'Adamic 0.1 refuses ' in err:
  status='refused';reason=err.split('Adamic 0.1 refuses ',1)[1].splitlines()[0];header='// a-check: refused '+reason
 elif ' error TS' in err:
  status='type error';code=re.search(r' error (TS\d+)',err).group(1);header='// a-check: type error '+code
 else:raise AssertionError((name,row))
 assert not body.startswith(b'// a-check:')
 if header:p.write_bytes(header.encode()+b'\n'+body)
 assert p.read_bytes()==(header.encode()+b'\n' if header else b'')+body
 statuses.append(dict(path=name,status=status,reason=reason,code=code,header=header or None,body_sha256=hashlib.sha256(body).hexdigest(),body_unchanged=True,diagnostic_before=err.strip()))
(out/'statuses.json').write_text(json.dumps({'counts':dict(collections.Counter(r['status'] for r in statuses)),'files':statuses},indent=2)+'\n')
after=scan('headers-scan',paths);assert after['exit']==0,after['failure']
selected=next(r['path'] for r in statuses if r['status']=='refused');p=root/selected;original=p.read_bytes();first,body=original.split(b'\n',1);mutants=[]
for mode in ['removed-header','wrong-reason']:
 try:
  mutant=body if mode=='removed-header' else b'// a-check: refused deliberately unmatched phantom-brand reason\n'+body
  p.write_bytes(mutant)
  (out/(mode+'.patch')).write_text(''.join(difflib.unified_diff(original.decode().splitlines(True),mutant.decode().splitlines(True),fromfile='a/'+selected,tofile='b/'+selected)))
  result=scan(mode,[selected]);assert result['exit']==1 and 'expected' in result['failure'],result
  assert result['results'][selected]['outcome']=='refused'
  mutants.append(dict(file=selected,mutation=mode,exit=1,caught_by='unchanged pinned Gate.aCheck with real branch compiler',failure=result['failure']))
 finally:p.write_bytes(original)
final=scan('final-scan',paths);assert final['exit']==0,final['failure']
assert all((root/r['path']).read_bytes()==((r['header']+'\n').encode() if r['header'] else b'')+subprocess.check_output(['git','show','2194037a:'+r['path']]) for r in statuses)
(out/'header-mutants.json').write_text(json.dumps(mutants,indent=2)+'\n')
(out/'gate-provenance.json').write_text(json.dumps({'commit':'fbac28c62493f27a788edc02a18bc8edb68de5da','source':'cloud/fast-gate/run.py','sha256':hashlib.sha256((out/'pinned-gate.py').read_bytes()).hexdigest(),'method':'unchanged Gate.aCheck; step reuses recorded branch compiler; spawn substitutes only compiler binary path, with a 20-second subprocess timeout'},indent=2)+'\n')
print('PASS all',len(paths),'programs; both mutants caught; bodies unchanged',flush=True)

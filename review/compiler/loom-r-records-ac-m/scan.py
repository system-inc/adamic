import collections,hashlib,importlib.util,json,pathlib,subprocess,sys,time
out=pathlib.Path(__file__).resolve().parent;root=out.parents[2]
spec=importlib.util.spec_from_file_location('records_ac_gate',root/'review/compiler/loom-r-records-ac/gate.py');gate=importlib.util.module_from_spec(spec);spec.loader.exec_module(gate)
paths=subprocess.check_output(['git','diff','--name-only','--diff-filter=AM','5a94465370ea0172e25aae8c13478e59e40c2bb4','--','*.a'],cwd=root,text=True).splitlines()
label=sys.argv[1];diagnostics={};commands=[]
if len(sys.argv)>2:paths=json.loads(pathlib.Path(sys.argv[2]).read_text())
class Process:
 def __init__(self,args,stdout,stderr):self.args=args;self.process=subprocess.Popen(args,cwd=root,stdout=stdout,stderr=stderr,text=True)
 def communicate(self):
  start=time.monotonic()
  try:stdout,stderr=self.process.communicate(timeout=45)
  except subprocess.TimeoutExpired:self.process.kill();self.process.communicate();raise
  self.returncode=self.process.returncode;diagnostics[self.args[-1]]={'exit':self.returncode,'stderr':stderr,'stdout_bytes':len(stdout.encode()),'seconds':round(time.monotonic()-start,3)}
  return stdout,stderr
class Harness:
 def __init__(self):self.arguments=type('Args',(),{'tree':str(root)})();self.result={};self.steps={};self.exits={};self.failure=None
 def step(self,name,args):
  with (out/(label+'-build.log')).open('w') as log:code=subprocess.run(args,cwd=root,stdout=log,stderr=subprocess.STDOUT,timeout=120).returncode
  commands.append({'command':args,'exit':code});return code==0
 def spawn(self,args,stdout,stderr):return Process(args,stdout,stderr)
 def fail(self,name,detail):self.failure=detail
h=Harness();gate.Gate.aCheck(h,paths)
rows=[]
for path in paths:
 row={'path':path,**h.result['a_check'][path],**diagnostics[path]};row['status']='NotYet' if row['outcome']=='checked' and row['exit']!=0 else row['outcome'];rows.append(row)
(out/(label+'.json')).write_text(json.dumps({'commands':commands,'counts':dict(collections.Counter(r['status'] for r in rows)),'exits':h.exits,'failure':h.failure,'rows':rows},indent=2)+'\n')
print(label,len(rows),dict(collections.Counter(r['status'] for r in rows)),h.exits,flush=True)
print(h.failure or 'PASS',flush=True)
sys.exit(h.exits['a-check'])

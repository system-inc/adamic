import subprocess,time,json,pathlib,sys,os,signal
p=pathlib.Path(__file__).resolve().parent;label=sys.argv[1];pattern=sys.argv[2] if len(sys.argv)>2 else '.'
cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./cmd/adamic-gate/','-run',pattern];t=time.monotonic()
with (p/(label+'.log')).open('w') as f:
 proc=subprocess.Popen(cmd,stdout=f,stderr=subprocess.STDOUT,start_new_session=True)
 try:code=proc.wait(timeout=121)
 except subprocess.TimeoutExpired:os.killpg(proc.pid,signal.SIGKILL);code=124
r={'command':' '.join(cmd),'exit':code,'wall':time.monotonic()-t,'seconds':None,'failed':[],'passed':[],'skipped':[],'outputs':{}}
for l in (p/(label+'.log')).read_text().splitlines():
 try:e=json.loads(l)
 except:continue
 n=e.get('Test');a=e.get('Action')
 if n and a=='output':r['outputs'].setdefault(n,[]).append(e['Output'].rstrip())
 if n and a in ['pass','fail','skip']:r[{'pass':'passed','fail':'failed','skip':'skipped'}[a]].append(n)
 if not n and a in ['pass','fail']:r['seconds']=e.get('Elapsed')
p.joinpath(label+'.json').write_text(json.dumps(r,indent=2));print(label,code,round(r['wall'],3),r['seconds'],r['failed'],flush=True)

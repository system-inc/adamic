import subprocess,time,json,pathlib,sys,os,signal
p=pathlib.Path(__file__).resolve().parent
label=sys.argv[1]; command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./bridge/tsgo/','-run',sys.argv[2]]
t=time.monotonic()
with (p/(label+'.log')).open('w') as f:
 process=subprocess.Popen(command,stdout=f,stderr=subprocess.STDOUT,start_new_session=True)
 try: code=process.wait(timeout=121)
 except subprocess.TimeoutExpired:
  os.killpg(process.pid,signal.SIGKILL);code=124
r={'command':' '.join(command),'wall':time.monotonic()-t,'exit':code,'seconds':None,'failed':[],'passed':[],'skipped':[],'outputs':{}}
for line in (p/(label+'.log')).read_text().splitlines():
 try: event=json.loads(line)
 except: continue
 name=event.get('Test');action=event.get('Action')
 if name and action=='output':r['outputs'].setdefault(name,[]).append(event['Output'].rstrip())
 if name and action in ['pass','fail','skip']:r[{'pass':'passed','fail':'failed','skip':'skipped'}[action]].append(name)
 if not name and action in ['pass','fail']:r['seconds']=event.get('Elapsed')
p.joinpath(label+'.json').write_text(json.dumps(r,indent=2));print(label,code,round(r['wall'],3),r['seconds'],r['failed'])

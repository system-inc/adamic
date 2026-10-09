import pathlib,json,subprocess,time,os,concurrent.futures
r=pathlib.Path('/workspace/adamic');o=pathlib.Path('/tmp/u009');names=json.loads((o/'names.json').read_text());ms=json.loads((o/'mutants.json').read_text())
def run(m,n):
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./cmd/adamic-meter/','-run','^'+n+'$'];env=os.environ.copy();env['ADAMIC_MUTANT']=m;t=time.monotonic();label=m+'-'+n
 with open(o/(label+'.log'),'w') as f:x=subprocess.run(cmd,cwd=r,env=env,stdout=f,stderr=subprocess.STDOUT)
 z=dict(name=label,command=cmd,seconds=time.monotonic()-t,exit=x.returncode)
 with open(o/'commands.jsonl','a') as f:f.write(json.dumps(z)+'\n')
 return z
jobs=[]
for m in ms:
 events=[]
 for l in (o/(m['id']+'-matrix.log')).read_text().splitlines():
  try:events.append(json.loads(l))
  except:pass
 observed={x['Test'] for x in events if x.get('Test') in names and x['Action'] in ['pass','fail','skip']}
 if observed!=set(names):
  print('recover',m['id'],flush=True)
  jobs += [(m['id'],n) for n in names]
with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:
 for z in pool.map(lambda args:run(*args),jobs):print(z,flush=True)

import pathlib,subprocess,time,json,concurrent.futures
P=pathlib.Path('/workspace/adamic/review/test-defend/internal-lower-module_namespace/session-6eef5ae5');names=['TestNamespaceClassEarlyConstructionStaysLoud','TestCallableNamespaceLimitsStayLoud','TestNamespaceAmbientHostInitialization','TestTscNamespaceDeclarationShapes'];results=[]
def run(n):
 cmd=['timeout','120','go','test','-count=1','-timeout','90s','./internal/lower/','-run','^'+n+'$','-coverpkg=github.com/system-inc/adamic/internal/lower','-coverprofile='+str(P/(n+'.cover'))];t=time.monotonic()
 with (P/(n+'.coverage.log')).open('w') as f:r=subprocess.run(cmd,cwd='/workspace/adamic',stdout=f,stderr=subprocess.STDOUT)
 return dict(test=n,seconds=time.monotonic()-t,exit=r.returncode,command=cmd)
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:results=list(pool.map(run,names))
(P/'coverage-runs.json').write_text(json.dumps(results,indent=2))
def covered(n):
 d=set()
 for l in (P/(n+'.cover')).read_text().splitlines()[1:]:
  loc,stm,cnt=l.split();f,pos=loc.rsplit(':',1);a,b=pos.split(',')
  if int(cnt):d.update((f,x) for x in range(int(a.split('.')[0]),int(b.split('.')[0])+1))
 return d
unique=sorted(covered(names[0])-covered(names[1]));(P/'class-exclusive.json').write_text(json.dumps(unique,indent=2));print(results);print('class exclusive',len(unique))

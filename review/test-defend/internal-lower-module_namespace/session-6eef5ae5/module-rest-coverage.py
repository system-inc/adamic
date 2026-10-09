import pathlib,subprocess,json,time
P=pathlib.Path('/workspace/adamic/review/test-defend/internal-lower-module_namespace/session-6eef5ae5');A='TestNamespaceAmbientHostInitialization';S='TestTscNamespaceDeclarationShapes';names=[l for l in (P/'list.log').read_text().splitlines() if l.startswith('Test') and l not in [A,S]];cmd=['timeout','120','go','test','-count=1','-timeout','90s','./internal/lower/','-run','^('+ '|'.join(names)+')$','-coverpkg=github.com/system-inc/adamic/internal/lower','-coverprofile='+str(P/'rest.cover')];t=time.monotonic()
with (P/'rest.coverage.log').open('w') as f:r=subprocess.run(cmd,cwd='/workspace/adamic',stdout=f,stderr=subprocess.STDOUT)
(P/'rest-coverage-run.json').write_text(json.dumps(dict(command=cmd,rows=names,seconds=time.monotonic()-t,exit=r.returncode),indent=2))
def covered(n):
 out=set()
 for l in (P/(n+'.cover')).read_text().splitlines()[1:]:
  loc,_,cnt=l.split();f,pos=loc.rsplit(':',1);a,b=pos.split(',')
  if int(cnt):out.update((f,x) for x in range(int(a.split('.')[0]),int(b.split('.')[0])+1))
 return out
if r.returncode==0:
 rest=covered('rest');d={A:sorted(covered(A)-(rest|covered(S))),S:sorted(covered(S)-(rest|covered(A)))};(P/'untrue-exclusive.json').write_text(json.dumps(d,indent=2));print({k:len(v) for k,v in d.items()})
print('exit',r.returncode,'seconds',time.monotonic()-t)

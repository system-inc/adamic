import subprocess,pathlib,json,time,os
p=pathlib.Path('/workspace/adamic/review/test-defend/internal-oracle-interface_cast/session'); rows=['TestInterfaceCastOracle','TestInterfaceCastImportedConstruction','TestInterfaceCastScalarTags','TestInterfaceCastChecksMalformedRead']; out=[]
for n in rows:
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverpkg=github.com/system-inc/adamic/internal/lower,github.com/system-inc/adamic/internal/native,github.com/system-inc/adamic/internal/javascript','-coverprofile='+str(p/(n+'.cover')),'./internal/oracle/','-run','^'+n+'$']; start=time.monotonic()
 with (p/(n+'-coverage.log')).open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1'})
 out.append({'test':n,'command':cmd,'exit':r.returncode,'wall':time.monotonic()-start}); (p/'coverage-runs.json').write_text(json.dumps(out,indent=2))
 if r.returncode:break

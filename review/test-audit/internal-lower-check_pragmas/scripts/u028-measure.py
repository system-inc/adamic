import subprocess,time,json
from pathlib import Path
r=Path('review/test-audit/internal-lower-check_pragmas')
rows=['TestCheckPragmasAreRefused','TestCheckPragmaNeighborsCompile','TestTsgoHonorsNoCheckInAdamicFiles','TestClassFeaturesReadonlyChecker','TestClassFeaturesPrivateChecker','TestClassFeaturesPrivateStorage','TestClassFeaturesAccessorRefusals','TestClassFeaturesStaticDeclarationsExecute','TestClassFeaturesStaticSoundness','TestClassFeaturesAccessorCaptureCycle','TestClassFeaturesNarrowedAccessor','TestClassFeaturesStaticParentCycle','TestClassFeaturesStaticInterfaceCycle']
(r/'scope.json').write_text(json.dumps(rows,indent=2)+'\n')
regex='^('+'|'.join(rows)+')$'
start=time.monotonic()
with (r/'coverage.log').open('w') as f:
 c=subprocess.run(['timeout','90','go','test','-count=1','-timeout','90s','-coverpkg=./internal/lower,./internal/load','-coverprofile='+str(r/'coverage.out'),'./internal/lower/','-run',regex],stdout=f,stderr=subprocess.STDOUT)
(r/'coverage-wall.json').write_text(json.dumps({'exit':c.returncode,'seconds':time.monotonic()-start}))
assert c.returncode==0
for row in rows:
 for i in range(1,4):
  with (r/(row+'-time-'+str(i)+'.log')).open('w') as f:
   c=subprocess.run(['timeout','90','go','test','-count=1','-timeout','90s','./internal/lower/','-run','^'+row+'$'],stdout=f,stderr=subprocess.STDOUT)
  assert c.returncode==0

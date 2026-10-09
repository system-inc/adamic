import json,pathlib,subprocess,time,os
p=pathlib.Path('review/test-defend/internal-lower-class_inheritance')
names=['TestInheritanceKeepsCheckerConstructorRules','TestClassFeaturesPrivateChecker','TestInheritanceCycleFinderIncludesInheritedFields','TestInheritanceGenericSoundness','TestInheritanceRefusesThisBeforeSuperReturns','TestInheritanceConditionalThisRules']
runs=[]
for n in names:
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverpkg=./internal/lower,./internal/load','-coverprofile='+str(p/(n+'.cover')),'./internal/lower/','-run','^'+n+'$']; begin=time.monotonic()
 with (p/(n+'-coverage.log')).open('w') as f: done=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env=dict(os.environ,ADAMIC_GATE_UNCACHED='1',ADAMIC_BUILD_CACHE_DIR='/tmp/defend-class-inheritance/cache/coverage'))
 runs.append(dict(test=n,command=cmd,exit=done.returncode,wall=time.monotonic()-begin));print(n,done.returncode,flush=True)
 if done.returncode: break
(p/'coverage-runs.json').write_text(json.dumps(runs,indent=2)+'\n')
def blocks(n):
 out={}
 for line in (p/(n+'.cover')).read_text().splitlines()[1:]:
  loc,statements,count=line.split();out[loc]=int(count)
 return out
if len(runs)==6:
 out=[]
 for a,b in zip(names[::2],names[1::2]):
  aa,bb=blocks(a),blocks(b);out.append(dict(test=a,subsumer=b,exclusive_blocks=[x for x,v in aa.items() if v>0 and bb.get(x,0)==0]))
 (p/'exclusive-coverage.json').write_text(json.dumps(out,indent=2)+'\n')

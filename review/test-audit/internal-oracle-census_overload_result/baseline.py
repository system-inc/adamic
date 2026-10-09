import pathlib,subprocess,json,time,os
p=pathlib.Path('review/test-audit/internal-oracle-census_overload_result');names=['TestCensusAppendResultProof','TestCensusOverloadResultStop','TestCensusSmallStoppedSourceOnNode'];records=[];env=os.environ.copy();env['ADAMIC_GATE_UNCACHED']='1'
def run(label,args):
 t=time.monotonic()
 with (p/(label+'.log')).open('w') as f:r=subprocess.run(args,env=env,stdout=f,stderr=subprocess.STDOUT)
 records.append(dict(label=label,command='ADAMIC_GATE_UNCACHED=1 '+' '.join(args),exit=r.returncode,wall=time.monotonic()-t));(p/'baseline-commands.json').write_text(json.dumps(records,indent=2));print(label,r.returncode,round(records[-1]['wall'],2),flush=True)
 return r.returncode
if run('bounded-baseline',['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^('+'|'.join(names)+')$'])!=0:raise SystemExit('RED bounded baseline')
for n in names:
 for i in range(1,4):
  if run(n+'.'+str(i),['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^'+n+'$'])!=0:raise SystemExit('RED row baseline')
run('coverage',['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverpkg=./internal/lower','-coverprofile='+str(p/'coverage.out'),'./internal/oracle/','-run','^('+'|'.join(names)+')$'])

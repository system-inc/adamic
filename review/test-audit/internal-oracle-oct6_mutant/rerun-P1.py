import pathlib,json,subprocess,os,time
scope=json.load(open('review/test-audit/internal-oracle-oct6_mutant/scope.json'));runs=[];env=os.environ.copy();env.update(ADAMIC_MUTANT='P1',ADAMIC_GATE_UNCACHED='1',ADAMIC_BUILD_CACHE_DIR='/tmp/u066/cache/P1')
for row in scope['rows']:
 reg=scope['native_regex'] if row=='TestNativeAgreesWithNode' else '^'+row+'$';cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',reg];start=time.monotonic()
 with open('/tmp/u066/P1-alone-'+row+'.log','w') as out:p=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
 runs.append(dict(id='P1',test=row,command='ADAMIC_MUTANT=P1 ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR=/tmp/u066/cache/P1 '+' '.join(cmd),seconds=time.monotonic()-start,returncode=p.returncode));print(row,p.returncode,flush=True)
pathlib.Path('/tmp/u066/P1-alone-runs.json').write_text(json.dumps(runs,indent=2))

import json,pathlib,subprocess,os,time
r=pathlib.Path('review/test-audit/internal-oracle-oct6_mutant');scope=json.loads((r/'scope.json').read_text());runs=[]
for id in ['clean','M1','M2','M3','M4','W1','W2','W3','P1','P2','P3','P4']:
 env=os.environ.copy();env.update(ADAMIC_MUTANT='' if id=='clean' else id,ADAMIC_GATE_UNCACHED='1',ADAMIC_BUILD_CACHE_DIR='/tmp/u066/cache/'+id)
 for name,reg in [('rows',scope['regex']),('native',scope['native_regex'])]:
  if id.startswith('W') and name=='native':continue
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',reg];start=time.monotonic()
  with open('/tmp/u066/'+id+'-'+name+'.log','w') as out:p=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
  runs.append(dict(id=id,part=name,command='ADAMIC_MUTANT='+env['ADAMIC_MUTANT']+' ADAMIC_GATE_UNCACHED=1 ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd),seconds=time.monotonic()-start,returncode=p.returncode));pathlib.Path('/tmp/u066/runs.json').write_text(json.dumps(runs,indent=2));print(id,name,p.returncode,round(runs[-1]['seconds'],3),flush=True)
  if id=='clean' and p.returncode:raise SystemExit('red switch control')

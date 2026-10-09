import pathlib,json,subprocess,time,os,shlex
p=pathlib.Path('review/test-audit/internal-oracle-census_overload_result');names=['TestCensusAppendResultProof','TestCensusOverloadResultStop','TestCensusSmallStoppedSourceOnNode'];primary='^('+'|'.join(names)+')$';family='^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^census_(append_overload|overload_lie)\\.a$';records=[];base=os.environ.copy();base['ADAMIC_GATE_UNCACHED']='1';items=json.loads((p/'manifest.json').read_text())
def run(id,label,regex):
 env=base.copy();env['ADAMIC_MUTANT']=id;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u056/cache/'+(id or 'baseline');cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',regex];t=time.monotonic()
 with (p/(label+'.log')).open('w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
 records.append(dict(id=id,label=label,command='ADAMIC_GATE_UNCACHED=1 ADAMIC_MUTANT='+id+' ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+shlex.join(cmd),exit=r.returncode,wall=time.monotonic()-t));(p/'matrix-commands.json').write_text(json.dumps(records,indent=2));print(label,r.returncode,round(records[-1]['wall'],2),flush=True);return (p/(label+'.log')).read_text()
if run('','switched-control',primary) and records[-1]['exit']!=0:raise SystemExit('red control')
for i in range(1,4):
 run('','family-baseline.'+str(i),family)
 if records[-1]['exit']!=0:raise SystemExit('red family baseline')
for item in items:
 id=item['id']
 if id.startswith('P'):
  for n in names[:2]:run(id,id+'.'+n,'^'+n+'$')
  continue
 s=run(id,id+'.primary',primary)
 if 'panic:' in s or records[-1]['exit']==124:
  for n in names:run(id,id+'.'+n,'^'+n+'$')
 s=run(id,id+'.family',family)
 # A panic on the selected shared-family inputs is narrowed to each observed fixture.
 if 'panic:' in s or records[-1]['exit']==124:
  for fixture in ['append_overload','overload_lie']:run(id,id+'.family.'+fixture,'^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^census_'+fixture+'\\.a$')

import pathlib,subprocess,json,time,os,shlex
p=pathlib.Path('review/test-audit/internal-lower-non_null_impossible');scope=json.loads((p/'scope.json').read_text());names=scope['names'];rows=scope['rows'];records=[];items=json.loads((p/'manifest.json').read_text());base=os.environ.copy();base['OPTIONAL_WIDENING_CONFIG']=str(p.resolve()/'census-project/tsconfig.json');base['OPTIONAL_WIDENING_OUTPUT']='/tmp/u040-census.json'
def run(id,regex,label):
 env=base.copy();env['ADAMIC_MUTANT']=id;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u040/cache/'+id;cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run',regex];t=time.monotonic()
 with (p/(label+'.log')).open('w') as out:r=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
 rec=dict(id=id,label=label,command='OPTIONAL_WIDENING_CONFIG='+env['OPTIONAL_WIDENING_CONFIG']+' OPTIONAL_WIDENING_OUTPUT='+env['OPTIONAL_WIDENING_OUTPUT']+' ADAMIC_MUTANT='+id+' ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+shlex.join(cmd),exit=r.returncode,wall=time.monotonic()-t);records.append(rec);(p/'matrix-commands.json').write_text(json.dumps(records,indent=2))
 census=pathlib.Path(base['OPTIONAL_WIDENING_OUTPUT'])
 if census.exists():(p/(label+'.census.json')).write_bytes(census.read_bytes())
 print(label,r.returncode,round(rec['wall'],2),flush=True);return (p/(label+'.log')).read_text()
def regex(row):return '^('+'|'.join(scope['family'])+')$' if row==rows[-1] else '^'+row+'$'
for x in items:
 id=x['id']
 if id.startswith('P'):
  owners={'P01':[r for r in rows if r not in ['TestOptionalWideningCensus','TestOptionalWideningSpreadOverwrite','TestOverloadInferenceWitnesses']],'P02':['TestOptionalWideningSpreadOverwrite'],'P03':['TestOptionalWideningCensus'],'P04':['TestOverloadInferenceWitnesses']}[id]
  for r in owners:run(id,regex(r),id+'.'+r.replace(' ','_'))
  continue
 s=run(id,'.',id)
 if 'panic:' in s or records[-1]['exit']==124:
  for r in rows:run(id,regex(r),id+'.'+r.replace(' ','_'))

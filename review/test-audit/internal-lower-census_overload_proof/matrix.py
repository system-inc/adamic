import pathlib,subprocess,json,time,os,shlex
p=pathlib.Path('review/test-audit/internal-lower-census_overload_proof');names=json.loads((p/'scope.json').read_text());items=json.loads((p/'manifest.json').read_text()); records=[]
def run(id,regex,label):
 env=os.environ.copy();env['ADAMIC_MUTANT']=id;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u027/cache/'+id
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run',regex];t=time.monotonic()
 with (p/(label+'.log')).open('w') as out:r=subprocess.run(cmd,env=env,stdout=out,stderr=subprocess.STDOUT)
 record=dict(id=id,label=label,command='ADAMIC_MUTANT='+id+' ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+shlex.join(cmd)+' > '+str(p/(label+'.log'))+' 2>&1',exit=r.returncode,wall=time.monotonic()-t)
 records.append(record);(p/'matrix-commands.json').write_text(json.dumps(records,indent=2)); print(label,r.returncode,round(record['wall'],3),flush=True)
 return (p/(label+'.log')).read_text()
for x in items:
 id=x['id'];s=run(id,'.',id)
 if 'panic:' in s or 'test timed out' in s or records[-1]['exit']==124:
  # Panic leaves later rows unknown. Rerun each requested row, never infer an unseen catch.
  for n in names:run(id,'^'+n+'$',id+'.'+n)

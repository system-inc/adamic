import pathlib,subprocess,os,json,time,shlex
p=pathlib.Path('review/test-audit/internal-oracle-stage3_front');scope=json.loads((p/'scope.json').read_text());names=scope['rows'];pattern='^('+'|'.join(names)+')$';family=scope['extra_pattern'];plan=json.loads((p/'plan.json').read_text());rs=[]
def run(mid,label,pattern):
 env=os.environ.copy();env['ADAMIC_GATE_UNCACHED']='1';env['ADAMIC_MUTANT']='' if mid=='control' else mid;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u070/cache/'+mid;env['ADAMIC_STAGE3_FIXTURE']=str(pathlib.Path('internal/oracle/testdata/switch_case_declarations/fallthrough.a').resolve());env['ADAMIC_STAGE3_RESULT']='/tmp/u070/hook-'+label+'.json';pathlib.Path('/tmp/u070-selector').write_text('' if mid=='control' else mid)
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',pattern];t=time.monotonic()
 with open(p/(label+'.log'),'w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env=env)
 events=[]
 for line in (p/(label+'.log')).read_text().splitlines():
  try:events.append(json.loads(line))
  except:pass
 failed=sorted(set(e['Test'].split('/')[0] for e in events if e.get('Action')=='fail' and e.get('Test')));panic=any(e.get('Output','').startswith('panic:') for e in events)
 record=dict(id=mid,label=label,command='ADAMIC_GATE_UNCACHED=1 ADAMIC_MUTANT='+env['ADAMIC_MUTANT']+' ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' ADAMIC_STAGE3_FIXTURE='+env['ADAMIC_STAGE3_FIXTURE']+' ADAMIC_STAGE3_RESULT='+env['ADAMIC_STAGE3_RESULT']+' '+shlex.join(cmd)+' > '+label+'.log 2>&1',exit=r.returncode,wall_seconds=time.monotonic()-t,failed_rows=failed,panic=panic,pattern=pattern,ran_rows=sorted(set(e['Test'].split('/')[0] for e in events if e.get('Action')=='run' and e.get('Test'))),completed_rows=sorted(set(e['Test'].split('/')[0] for e in events if e.get('Action') in ['pass','fail','skip'] and e.get('Test'))));rs.append(record);(p/'matrix.json').write_text(json.dumps(rs,indent=2));print(label,r.returncode,failed,'panic',panic,round(record['wall_seconds'],2),flush=True);return record
for mid in ['control']+[m['id'] for m in plan['mutants']]+[q['id'] for q in plan['probes']]:
 rec=run(mid,mid,pattern)
 if mid=='control' and rec['exit']:break
 if rec['panic']:
  for name in names:
   if name not in rec['completed_rows']:run(mid,mid+'-'+name,'^'+name+'$')
 if mid!='P01':run(mid,mid+'-agreement',family)

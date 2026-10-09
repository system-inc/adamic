import os,subprocess,time,json,pathlib,re
p=pathlib.Path('review/test-audit/bridge-tsgo-bridge'); rows=['TestTSGoRequiresLink','TestBridgeUnitsCoverEveryPiece','TestBridgeProductCacheIsVerified']; pattern='^('+'|'.join(rows)+')$'; records=[]
for id in ['switch-control']+[x['id'] for x in json.loads((p/'plan.json').read_text())]:
 env=os.environ.copy(); env['ADAMIC_MUTANT']='' if id=='switch-control' else id; env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u001/cache/'+id
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./bridge/tsgo/','-run',pattern]; start=time.monotonic()
 with (p/(id+'.log')).open('w') as f: result=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
 events=[]
 for line in (p/(id+'.log')).read_text().splitlines():
  try: events.append(json.loads(line))
  except ValueError: pass
 failures=[e['Test'] for e in events if e['Action']=='fail' and e.get('Test') in rows]; passed=[e['Test'] for e in events if e['Action']=='pass' and e.get('Test') in rows]
 record=dict(id=id,command='ADAMIC_MUTANT='+env['ADAMIC_MUTANT']+' ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd),exit=result.returncode,wall_seconds=time.monotonic()-start,failed=failures,passed=passed,unknown=[r for r in rows if r not in failures+passed],bounded=True,matrix_rows=rows)
 record['failure_lines']=[e['Output'].strip() for e in events if e.get('Test') in failures and e['Action']=='output' and re.search(r'\w+_test.go:\d+:',e.get('Output',''))]
 records.append(record); (p/'matrix.json').write_text(json.dumps(records,indent=2)+'\n'); print(id,result.returncode,failures,record['failure_lines'],round(record['wall_seconds'],3),flush=True)
 if id=='switch-control' and result.returncode: break

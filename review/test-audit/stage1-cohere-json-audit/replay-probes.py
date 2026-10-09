exec(open('/tmp/u102/matrix.py').read().split("for id in ['control']")[0])
records=json.loads(pathlib.Path('/tmp/u102/matrix.json').read_text())
for id,runrows in [('P01',[n]) for n in rows]+[(x['id'],rows) for x in plan if x['kind']=='probe' and x['id']!='P01']:
 env=dict(os.environ,ADAMIC_JSON_PRETTIER='/tmp/u102/prettier',ADAMIC_BUILD_CACHE_DIR='/tmp/u102/cache/'+id,ADAMIC_MUTANT=id)
 pat='^('+'|'.join(runrows)+')$';cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/json/','-run',pat];suffix='-'+runrows[0] if len(runrows)==1 else '';logpath=out/(id+suffix+'.log');t=time.monotonic()
 with logpath.open('w') as log:p=subprocess.run(cmd,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
 events=[]
 for line in logpath.read_text().splitlines():
  try:events.append(json.loads(line))
  except:pass
 failures=list(dict.fromkeys(v['Test'].split('/')[0] for v in events if v.get('Action')=='fail' and v.get('Test')))
 records.append(dict(id=id,command=cmd,env={k:env[k] for k in ['ADAMIC_JSON_PRETTIER','ADAMIC_BUILD_CACHE_DIR','ADAMIC_MUTANT']},wall_seconds=time.monotonic()-t,exit=p.returncode,failed=failures,passed=[v['Test'] for v in events if v.get('Action')=='pass' and v.get('Test') in runrows],skipped=[v['Test'] for v in events if v.get('Action')=='skip' and v.get('Test') in runrows],matrix_rows=runrows,log=logpath.name))
 pathlib.Path('/tmp/u102/matrix.json').write_text(json.dumps(records,indent=2));print(id,suffix,p.returncode,round(records[-1]['wall_seconds'],2),','.join(failures),flush=True)

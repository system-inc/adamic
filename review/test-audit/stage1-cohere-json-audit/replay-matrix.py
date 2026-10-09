import pathlib,json,subprocess,time,os
root=pathlib.Path('/workspace/adamic');out=root/'review/test-audit/stage1-cohere-json-audit'
rows=list(dict.fromkeys(x['test'] for x in json.loads(pathlib.Path('/tmp/u102/timing.json').read_text())))+['TestRepositoryCorpusMutants','TestRepositoryLandingWithoutPinEdit','TestRepositoryRequiresGit']
pattern='^('+'|'.join(rows)+')$'; plan=json.loads(pathlib.Path('/tmp/u102/plan.json').read_text()); records=[]
for id in ['control']+[x['id'] for x in plan]:
 env=dict(os.environ,ADAMIC_JSON_PRETTIER='/tmp/u102/prettier',ADAMIC_BUILD_CACHE_DIR='/tmp/u102/cache/'+id,ADAMIC_MUTANT='' if id=='control' else id)
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/json/','-run',pattern]
 started=time.monotonic()
 with (out/(id+'.log')).open('w') as log:p=subprocess.run(cmd,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
 seconds=time.monotonic()-started; fails=[]; evidence={}; passed=[];skips=[];panic=False
 for line in (out/(id+'.log')).read_text().splitlines():
  try:v=json.loads(line)
  except:continue
  name=v.get('Test','');top=name.split('/')[0]
  if v.get('Action')=='fail' and name and top not in fails:fails.append(top)
  if v.get('Action')=='pass' and name==top:passed.append(top)
  if v.get('Action')=='skip' and name==top:skips.append(top)
  if v.get('Action')=='output':
   output=v.get('Output','').strip()
   if 'panic:' in output:panic=True
   if name and not output.startswith(('===','---','PASS','FAIL')) and any(k in output for k in ['gap changed','wrong shard','lost/','lost final','survived','failed to catch','sample changed','toolchain','ordered control','fixture baseline','export without','landing changed','native:','panic:']):evidence.setdefault(top,output)
 records.append(dict(id=id,command=cmd,env={k:env[k] for k in ['ADAMIC_JSON_PRETTIER','ADAMIC_BUILD_CACHE_DIR','ADAMIC_MUTANT']},wall_seconds=seconds,exit=p.returncode,failed=fails,passed=passed,skipped=skips,evidence=evidence,panic=panic,matrix_rows=rows))
 pathlib.Path('/tmp/u102/matrix.json').write_text(json.dumps(records,indent=2))
 print(id,p.returncode,round(seconds,2),','.join(fails),','.join(skips),flush=True)
 if id=='control' and p.returncode:raise SystemExit('Switched clean baseline red; stop')
 if seconds>=90 or panic:raise SystemExit('Cooked/panic: follow up individually')

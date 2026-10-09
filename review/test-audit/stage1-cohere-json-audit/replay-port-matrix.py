import pathlib,json,subprocess,time,os
root=pathlib.Path('/workspace/adamic');out=root/'review/test-audit/stage1-cohere-json-audit';rows=list(dict.fromkeys(x['test'] for x in json.loads(pathlib.Path('/tmp/u102/timing.json').read_text())));rows.remove('TestProfileSnapshotsAgree');rows+=['TestRepositoryCorpusMutants','TestRepositoryLandingWithoutPinEdit','TestRepositoryRequiresGit'];records=[]
for id,f in [('M04','stage1/cohere/json/doc.ts'),('P10','stage1/cohere/json/main.ts')]:
 base=(root/f).read_bytes();p=subprocess.run(['git','apply','--check',str(out/(id+'.diff'))],cwd=root,stdout=subprocess.PIPE,stderr=subprocess.STDOUT)
 if p.returncode:raise SystemExit('Standalone diff does not apply')
 subprocess.run(['git','apply',str(out/(id+'.diff'))],cwd=root,check=True)
 env=dict(os.environ,ADAMIC_JSON_PRETTIER='/tmp/u102/prettier',ADAMIC_BUILD_CACHE_DIR='/tmp/u102/cache/'+id,ADAMIC_MUTANT='')
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/json/','-run','^('+'|'.join(rows)+')$'];t=time.monotonic()
 with (out/(id+'-other-rows.log')).open('w') as log:p=subprocess.run(cmd,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
 (root/f).write_bytes(base)
 records.append(dict(id=id,command=cmd,env={k:env[k] for k in ['ADAMIC_JSON_PRETTIER','ADAMIC_BUILD_CACHE_DIR']},exit=p.returncode,wall_seconds=time.monotonic()-t,log=id+'-other-rows.log',matrix_rows=rows,apply_check_exit=0))
 pathlib.Path('/tmp/u102/port-matrix.json').write_text(json.dumps(records,indent=2));print(id,p.returncode,round(records[-1]['wall_seconds'],2),flush=True)

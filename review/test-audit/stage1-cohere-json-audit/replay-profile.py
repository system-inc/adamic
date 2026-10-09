import pathlib,subprocess,time,os,json,difflib
root=pathlib.Path('/workspace/adamic');out=root/'review/test-audit/stage1-cohere-json-audit';records=[]
plan=[dict(id='M04',kind='production-port',file='stage1/cohere/json/doc.ts',old="const indent = node.kind === 'indent' ? command.indent + 2 : command.indent;",new="const indent = node.kind === 'indent' ? command.indent + 3 : command.indent;"),dict(id='P10',kind='probe-port',file='stage1/cohere/json/main.ts',old='module body',new='export {};\n')]
for x in plan:
 s=subprocess.check_output(['git','show','origin/main:'+x['file']],cwd=root,text=True);x['line']=s[:s.index(x['old'])].count('\n')+1 if x['id']=='M04' else 1;x['source']=s.replace(x['old'],x['new']) if x['id']=='M04' else x['new']
 (out/(x['id']+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),x['source'].splitlines(True),fromfile='a/'+x['file'],tofile='b/'+x['file'])))
(out/'profile-plan.txt').write_text('Written before catches. Code under test: Documents.print indentation and main.ts module entry. Menu: change constant +2 to +3. P10 drops the entire module body to produce an empty native main. Historical final snapshot is rebuilt from 921b2e809266aa53cc30a068e369c2449bce072b; mutant and probe products are built from origin/main.\n'+json.dumps([{k:v for k,v in x.items() if k!='source'} for x in plan],indent=2))
# One final snapshot cuts the over-budget two-snapshot row without changing its corpus or checker.
for i in range(3):
 env=dict(os.environ,ADAMIC_MUTANT='',ADAMIC_JSON_PRETTIER='/tmp/u102/prettier',ADAMIC_JSON_PROFILE_BINARIES='/tmp/u102/snapshots/final/snapshot',ADAMIC_BUILD_CACHE_DIR='/tmp/u102/cache/profile-control')
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/json/','-run','^TestProfileSnapshotsAgree$'];t=time.monotonic()
 with (out/('profile-timing-'+str(i)+'.log')).open('w') as log:p=subprocess.run(cmd,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
 records.append(dict(id='profile-control',trial=i,command=cmd,env={k:env[k] for k in ['ADAMIC_JSON_PRETTIER','ADAMIC_JSON_PROFILE_BINARIES','ADAMIC_BUILD_CACHE_DIR']},exit=p.returncode,wall_seconds=time.monotonic()-t,log='profile-timing-'+str(i)+'.log'))
 pathlib.Path('/tmp/u102/profile.json').write_text(json.dumps(records,indent=2));print('profile-control',i,p.returncode,round(records[-1]['wall_seconds'],2),flush=True)
 if p.returncode:raise SystemExit('Final snapshot baseline failed/cooked: stop')
for x in plan:
 id=x['id'];d=pathlib.Path('/tmp/u102/port')/id;d.mkdir(parents=True,exist_ok=True)
 for f in ['parser.ts','doc.ts','formatter.ts','main.ts','width.ts','widthTables.ts','identifierTables.ts']:(d/f).write_bytes(subprocess.check_output(['git','show','origin/main:stage1/cohere/json/'+f],cwd=root))
 (d/pathlib.Path(x['file']).name).write_text(x['source'])
 for stage,cmd in [('lower',['timeout','90','go','run','./cmd/adamic','c',str(d/'main.ts')]),('clang',['timeout','90','clang','-std=c11','-O2','-g','-ffp-contract=off','-fno-optimize-sibling-calls','-Iinternal/native/runtime',str(d/'generated.c')]+[str(y) for y in (root/'internal/native/runtime').glob('*.c')]+['-lm','-o',str(d/'product')])]:
  t=time.monotonic()
  with (out/(id+'-'+stage+'.log')).open('w') as log:
   if stage=='lower':
    with (d/'generated.c').open('w') as c:p=subprocess.run(cmd,cwd=root,env=dict(os.environ,ADAMIC_MUTANT='',ADAMIC_BUILD_CACHE_DIR='/tmp/u102/cache/'+id),stdout=c,stderr=log)
   else:p=subprocess.run(cmd,cwd=root,stdout=log,stderr=subprocess.STDOUT)
  records.append(dict(id=id,stage=stage,command=cmd,exit=p.returncode,wall_seconds=time.monotonic()-t,log=id+'-'+stage+'.log'))
  pathlib.Path('/tmp/u102/profile.json').write_text(json.dumps(records,indent=2));print(id,stage,p.returncode,round(records[-1]['wall_seconds'],2),flush=True)
  if p.returncode:raise SystemExit('Port standalone diff build failed')
 env=dict(os.environ,ADAMIC_MUTANT='',ADAMIC_JSON_PRETTIER='/tmp/u102/prettier',ADAMIC_JSON_PROFILE_BINARIES=str(d/'product'),ADAMIC_BUILD_CACHE_DIR='/tmp/u102/cache/'+id)
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/json/','-run','^TestProfileSnapshotsAgree$'];t=time.monotonic()
 with (out/(id+'-profile.log')).open('w') as log:p=subprocess.run(cmd,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
 records.append(dict(id=id,command=cmd,env={k:env[k] for k in ['ADAMIC_JSON_PRETTIER','ADAMIC_JSON_PROFILE_BINARIES','ADAMIC_BUILD_CACHE_DIR']},exit=p.returncode,wall_seconds=time.monotonic()-t,log=id+'-profile.log',matrix_rows=['TestProfileSnapshotsAgree']))
 pathlib.Path('/tmp/u102/profile.json').write_text(json.dumps(records,indent=2));print(id,'profile',p.returncode,round(records[-1]['wall_seconds'],2),flush=True)

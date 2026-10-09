import pathlib,subprocess,time,os,json
root=pathlib.Path('/workspace/adamic');out=root/'review/test-audit/stage1-cohere-json-audit';records=[]
commits={'baseline':'adb4e0aa911b010c038c4925ec0a5c949f313acb','final':'921b2e809266aa53cc30a068e369c2449bce072b'}
for name,commit in commits.items():
 d=pathlib.Path('/tmp/u102/snapshots')/name;d.mkdir(parents=True,exist_ok=True)
 for f in ['parser.ts','doc.ts','formatter.ts','main.ts','width.ts','widthTables.ts','identifierTables.ts']:(d/f).write_bytes(subprocess.check_output(['git','show',commit+':stage1/cohere/json/'+f],cwd=root))
 for stage,cmd in [('lower',['timeout','90','go','run','./cmd/adamic','c',str(d/'main.ts')]),('clang',['timeout','90','clang','-std=c11','-O2','-g','-ffp-contract=off','-fno-optimize-sibling-calls','-Iinternal/native/runtime',str(d/'generated.c')]+[str(x) for x in (root/'internal/native/runtime').glob('*.c')]+['-lm','-o',str(d/'snapshot')])]:
  t=time.monotonic()
  with (out/('snapshot-'+name+'-'+stage+'.log')).open('w') as log:
   if stage=='lower':
    with (d/'generated.c').open('w') as c:p=subprocess.run(cmd,cwd=root,env=dict(os.environ,ADAMIC_MUTANT=''),stdout=c,stderr=log)
   else:p=subprocess.run(cmd,cwd=root,stdout=log,stderr=subprocess.STDOUT)
  records.append(dict(name=name,commit=commit,stage=stage,command=cmd,exit=p.returncode,wall_seconds=time.monotonic()-t))
  pathlib.Path('/tmp/u102/snapshot-builds.json').write_text(json.dumps(records,indent=2));print(name,stage,p.returncode,round(records[-1]['wall_seconds'],2),flush=True)
  if p.returncode:raise SystemExit('Snapshot reconstruction blocked')

import pathlib,json,subprocess,time,os,statistics
D=pathlib.Path('/tmp/u103/evidence');env=os.environ.copy();env['ADAMIC_GATE_SAMPLE']=(D/'origin.txt').read_text().strip();env['ADAMIC_JSON_PRETTIER']='/tmp/u086/library/node_modules/prettier'
rows=['TestPortMatchesGoCohereSplit_Setup','TestPortMatchesGoCohereSplitUnion','TestPortMatchesGoCohere family','TestProduct_JSON family','TestThreePortMutantsAreCaught','TestAdditionalJSONBoundaries','TestSingleFileStdoutDriver','TestProgressGuard','TestRepositoryCorpusMutants','TestRepositoryLandingWithoutPinEdit','TestRepositoryRequiresGit'];commands=[];timings={}
for row in rows:
 regex='^TestPortMatchesGoCohere_[0-9]{3}$' if row=='TestPortMatchesGoCohere family' else '^TestProduct_JSON(GoOracle|LoweredPort|NativeRelease|NativeSanitized)$' if row=='TestProduct_JSON family' else '^'+row+'$'
 vals=[]
 for i in range(1,4):
  label='time-'+row.replace(' ','-')+'-'+str(i);args=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/json/','-run',regex];t=time.monotonic()
  with (D/(label+'.log')).open('w') as out:rc=subprocess.call(args,stdout=out,stderr=subprocess.STDOUT,env=env)
  commands.append(dict(label=label,args=args,wall_seconds=round(time.monotonic()-t,3),exit=rc));(D/'timing-commands.json').write_text(json.dumps(commands,indent=2))
  vals.append(None)
  for l in (D/(label+'.log')).read_text().splitlines():
   try:e=json.loads(l)
   except:continue
   if e.get('Action')=='pass' and not e.get('Test'):vals[-1]=e['Elapsed']
  print(label,rc,vals[-1],flush=True)
  if rc:raise RuntimeError('clean timing failed '+label)
 timings[row]=dict(runs=vals,median=statistics.median(vals));(D/'timings.json').write_text(json.dumps(timings,indent=2))

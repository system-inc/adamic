import pathlib,subprocess,os,json,time
p=pathlib.Path('/workspace/adamic/review/test-audit/stage1-cohere-typeaware-typeaware_top_shards');scope=json.loads((p/'scope.json').read_text())['names']
groups=[('TestTypeAwareTopPlanted family',[n for n in scope if n.startswith('TestTypeAwareTopPlanted_')]),('TestTypeAwareTopShardPlantedFailure',['TestTypeAwareTopShardPlantedFailure']),('TestTypeAwareAgreementAndMutants family',[n for n in scope if n.startswith('TestTypeAwareAgreementAndMutants') and not n.endswith('_Setup')]),('TestTypeAwareAgreementAndMutants_Setup',['TestTypeAwareAgreementAndMutants_Setup']),('TestProduct_volume_guard family',[n for n in scope if n.startswith('TestProduct_volume_guard_')]),('TestVolumeConfigGuardAndMutant_Setup',['TestVolumeConfigGuardAndMutant_Setup']),('TestVolumeConfigGuardAndMutant family',[n for n in scope if n.startswith('TestVolumeConfigGuardAndMutant_') and n[-3:].isdigit() and n[-3:]!='000']+['TestVolumeConfigGuardAndMutantUnion']),('TestVolumeConfigGuardAndMutant_000',['TestVolumeConfigGuardAndMutant_000']),('TestVolumeConfigGuardAndMutantPlantedFailure',['TestVolumeConfigGuardAndMutantPlantedFailure'])]
(p/'groups.json').write_text(json.dumps(dict(groups),indent=2)+'\n');env=os.environ.copy();env.update(ADAMIC_TYPESCRIPT_SOURCE='/tmp/u146/pinned-typescript',ADAMIC_TYPEAWARE_BENCH='1');records=[]
for group,names in groups:
 for trial in range(1,4):
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/typeaware/','-run','^('+'|'.join(names)+')$'];log=p/(group.replace(' ','-')+'-'+str(trial)+'.log');start=time.monotonic()
  with log.open('w') as f:r=subprocess.run(cmd,cwd='/workspace/adamic',env=env,stdout=f,stderr=subprocess.STDOUT)
  es=[]
  for l in log.read_text().splitlines():
   try:es.append(json.loads(l))
   except ValueError:pass
  seconds=next((e['Elapsed'] for e in reversed(es) if e.get('Action')=='pass' and 'Test' not in e),None)
  records.append(dict(group=group,trial=trial,seconds=seconds,wall=time.monotonic()-start,exit=r.returncode,command=cmd,log=log.name));(p/'timings.json').write_text(json.dumps(records,indent=2)+'\n');print(group,trial,r.returncode,seconds,flush=True)
  if r.returncode!=0:print('COOKED OR RED: inspect before further dependent audit',flush=True);raise SystemExit(1)

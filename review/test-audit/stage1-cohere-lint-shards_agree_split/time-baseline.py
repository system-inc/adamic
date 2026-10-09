import pathlib,subprocess,time,json,os
p=pathlib.Path('review/test-audit/stage1-cohere-lint-shards_agree_split');env=os.environ.copy();env['ADAMIC_TYPESCRIPT_SOURCE']='/tmp/u112/TypeScript-050880ce59e30b356b686bd3144efe24f875ebc8'
groups=[('suggestion','^TestSuggestionAlongsideAutomaticFix_[0-9]+$'),('witness-kind','^TestWitnessScriptKind_[0-9]+$'),('shards','^TestShardsAgree_[0-9]+$'),('shards-union','^TestShardsAgree_Union$'),('shards-setup','^TestShardsAgree_Setup$'),('shards-required','^TestShardsAgree_SetupRequired$'),('suggestion-setup','^TestSuggestionAlongsideAutomaticFix_Setup$'),('suggestion-planted','^TestSuggestionAlongsideAutomaticFixPlantedFailure$'),('suggestion-required','^TestSuggestionAlongsideAutomaticFixSetupIsRequired$'),('product-lower','^TestProduct_WitnessScriptKindLowered$'),('product-native','^TestProduct_WitnessScriptKindNative$'),('product-oracle','^TestProduct_WitnessScriptKindGoOracle$'),('kind-planted','^TestWitnessScriptKindPlantedFailure$'),('kind-setup','^TestWitnessScriptKind_Setup$'),('kind-standalone','^TestWitnessScriptKindStandalone$'),('kind-union','^TestWitnessScriptKind$')]
results=[]
for name,regex in groups:
 for i in range(3):
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/','-run',regex];t=time.monotonic();log=p/f'clean-{name}-{i+1}.log'
  with log.open('w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
  events=[]
  for l in log.read_text().splitlines():
   try:events.append(json.loads(l))
   except:pass
  timed=any('test timed out' in e.get('Output','') for e in events) or r.returncode==124
  x={'group':name,'command':cmd,'log':str(log),'exit':r.returncode,'wall':time.monotonic()-t,'timeout':timed,'events':[e for e in events if e.get('Action') in ['pass','fail','skip']]};results.append(x);(p/'clean-runs.json').write_text(json.dumps(results,indent=2));print(name,i+1,r.returncode,flush=True)
  if r.returncode and not timed:print('RED BASELINE: stop',flush=True);raise SystemExit(1)
  if timed:break

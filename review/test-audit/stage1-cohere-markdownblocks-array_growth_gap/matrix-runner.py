import pathlib,json,subprocess,time,os
root=pathlib.Path('/workspace/adamic');out=root/'review/test-audit/stage1-cohere-markdownblocks-array_growth_gap';menu={x['id']:x for x in json.loads((out/'menu.json').read_text())}
scope='^(TestArrayGrowthWitness|TestMarkdownASTPreprocessing|TestWholeDocumentOraclePreflight_Setup|TestMicromarkInputChunks|TestMarkdownSourceDecoding|TestTokenizerEventShardUnion|TestTokenizerEventShardPlantedDisagreement|TestTokenizerEventShardGrowth|TestTokenizerEventsUnion|TestTokenizerEvents_(000|001|002|255|511)|TestFrontMatterStage|TestParserRepresentationProbes|TestMdastIdentifierScalars|TestMarkdownLeafComposition_Setup|TestMarkdownLeafCompositionUnion|TestMarkdownLeafCompositionPlantedDisagreement|TestMarkdownLeafComposition_00[0-6])$'
# All production mutants use the same bounded representative matrix; unseen members remain unknown.
results=[]
for id in ['M1','M2','M3','M4','P_C','S1','S2','S3','S4','W1','W2']:
 m=menu[id];p=root/m['file'];original=p.read_text();assert original.count(m['from_'])==1
 p.write_text(original.replace(m['from_'],m['to'],1))
 try:
  if id=='M1':
   with (out/'M1-vet.log').open('w') as f:v=subprocess.run(['go','vet','./internal/native/'],cwd=root,stdout=f,stderr=subprocess.STDOUT)
   if v.returncode:raise RuntimeError('M1 vet failed')
  if id.startswith('M') or id=='P_C':pattern=scope
  else:pattern={'S1':'^TestWholeDocumentOraclePreflight_Setup$','S2':'^TestTokenizerEventShardUnion$','S3':'^TestTokenizerEventShardGrowth$','S4':'^TestMarkdownLeafComposition_Setup$','W1':'^TestTokenizerEventShardPlantedDisagreement$','W2':'^TestMarkdownLeafComposition(Union|PlantedDisagreement)$'}[id]
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u126/cache/'+id
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/markdownblocks/','-run',pattern]
  start=time.time()
  with (out/(id+'.log')).open('w') as f:r=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
  events=[]
  for l in (out/(id+'.log')).read_text().splitlines():
   try:events.append(json.loads(l))
   except:pass
  result=dict(id=id,command=cmd,cache=env['ADAMIC_BUILD_CACHE_DIR'],wall=time.time()-start,exit=r.returncode,rows=[e['Test'] for e in events if e.get('Action')=='run' and '/' not in e.get('Test','/')],failures=[e['Test'] for e in events if e.get('Action')=='fail' and '/' not in e.get('Test','/')],passed=[e['Test'] for e in events if e.get('Action')=='pass' and '/' not in e.get('Test','/')],cooked=any('panic: test timed out' in e.get('Output','') for e in events),builds=[e.get('Output','').strip() for e in events if 'build ' in e.get('Output','')])
  results.append(result);(out/'matrix.json').write_text(json.dumps(results,indent=2));print(id,result['wall'],result['failures'],flush=True)
 finally:p.write_text(original)

import pathlib,json,subprocess,time,os
root=pathlib.Path('/workspace/adamic');out=root/'review/test-audit/stage1-cohere-markdownblocks-array_growth_gap'
while True:
 try:
  m=json.loads((out/'matrix.json').read_text())
  if len(m)>=11:break
 except:pass
 time.sleep(2)
menu={x['id']:x for x in json.loads((out/'menu.json').read_text())};results=[]
patterns={'P_chunks':'^(TestMicromarkInputChunks|TestTokenizerEvents_003)$','P_decode':'^TestMarkdownSourceDecoding$','P_identifier':'^TestMdastIdentifierScalars$','P_front':'^TestFrontMatterStage$','P_ast':'^TestMarkdownASTPreprocessing$','P_leaf':'^TestMarkdownLeafComposition_00[0-3]$'}
for id,pattern in patterns.items():
 m=menu[id];p=root/m['file'];old=p.read_text();assert old.count(m['from_'])==1
 p.write_text(old.replace(m['from_'],m['to'],1))
 try:
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u126/cache/'+id
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/markdownblocks/','-run',pattern];start=time.time()
  with (out/(id+'.log')).open('w') as f:r=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
  events=[]
  for l in (out/(id+'.log')).read_text().splitlines():
   try:events.append(json.loads(l))
   except:pass
  result=dict(id=id,command=cmd,cache=env['ADAMIC_BUILD_CACHE_DIR'],wall=time.time()-start,exit=r.returncode,rows=[e['Test'] for e in events if e.get('Action')=='run' and '/' not in e.get('Test','/')],failures=[e['Test'] for e in events if e.get('Action')=='fail' and '/' not in e.get('Test','/')],passed=[e['Test'] for e in events if e.get('Action')=='pass' and '/' not in e.get('Test','/')],cooked=any('panic: test timed out' in e.get('Output','') for e in events))
  results.append(result);(out/'probes.json').write_text(json.dumps(results,indent=2));print(result,flush=True)
 finally:p.write_text(old)

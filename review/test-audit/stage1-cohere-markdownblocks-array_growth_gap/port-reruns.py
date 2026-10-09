import pathlib,shutil,json,subprocess,time,os
base=pathlib.Path('/tmp/u126/clean-source');out=pathlib.Path('/workspace/adamic/review/test-audit/stage1-cohere-markdownblocks-array_growth_gap');menu={m['id']:m for m in json.loads((out/'menu.json').read_text())};results=[]
for id,patterns in [('M3',['^TestMicromarkInputChunks$','^TestTokenizerEvents_000$','^TestTokenizerEvents_001$','^TestTokenizerEvents_002$','^TestTokenizerEvents_255$','^TestTokenizerEvents_511$']),('M4',['^TestMarkdownLeafComposition_00[0-3]$'])]:
 dst=pathlib.Path('/tmp/u126/'+id+'-source');dst.mkdir(exist_ok=True)
 for p in base.iterdir():
  if p.name=='stage1':continue
  (dst/p.name).symlink_to(p,target_is_directory=p.is_dir())
 (dst/'stage1/cohere').mkdir(parents=True,exist_ok=True)
 for p in (base/'stage1').iterdir():
  if p.name!='cohere':(dst/'stage1'/p.name).symlink_to(p,target_is_directory=p.is_dir())
 for p in (base/'stage1/cohere').iterdir():
  q=dst/'stage1/cohere'/p.name
  if p.name=='markdownblocks':shutil.copytree(p,q)
  else:q.symlink_to(p,target_is_directory=p.is_dir())
 m=menu[id];p=dst/m['file'];old=p.read_text();assert old.count(m['from_'])==1;p.write_text(old.replace(m['from_'],m['to'],1))
 for k,pattern in enumerate(patterns):
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u126/cache/'+id+'-rerun'
  cmd=['timeout','100','go','tool','test2json','-t','-p','github.com/system-inc/adamic/stage1/cohere/markdownblocks','/tmp/u126/clean.test','-test.v=test2json','-test.count=1','-test.timeout=90s','-test.run='+pattern];start=time.time()
  path=out/f'{id}-bounded-{k}.log'
  with path.open('w') as f:r=subprocess.run(cmd,cwd=dst/'stage1/cohere/markdownblocks',env=env,stdout=f,stderr=subprocess.STDOUT)
  events=[]
  for l in path.read_text().splitlines():
   try:events.append(json.loads(l))
   except:pass
  result=dict(id=id,command=cmd,wall=time.time()-start,exit=r.returncode,rows=[e['Test'] for e in events if e.get('Action')=='run' and '/' not in e.get('Test','/')],failures=[e.get('Test') for e in events if e.get('Action')=='fail' and e.get('Test')],cooked=any('panic: test timed out' in e.get('Output','') for e in events));results.append(result);(out/'port-reruns.json').write_text(json.dumps(results,indent=2));print(result,flush=True)

import pathlib,subprocess,time,os,json
out=pathlib.Path('/workspace/adamic/review/test-audit/stage1-cohere-markdownblocks-array_growth_gap');results=[]
for row in ['TestMarkdownASTPreprocessing','TestMicromarkInputChunks','TestMarkdownSourceDecoding','TestMdastIdentifierScalars','TestParserRepresentationProbes','TestTokenizerEvents_000']:
 env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u126/cache/M1-rerun'
 cmd=['timeout','100','go','tool','test2json','-t','-p','github.com/system-inc/adamic/stage1/cohere/markdownblocks','/tmp/u126/M1.test','-test.v=test2json','-test.count=1','-test.timeout=90s','-test.run=^'+row+'$'];start=time.time()
 with (out/('M1-alone-'+row+'.log')).open('w') as f:r=subprocess.run(cmd,cwd='/tmp/u126/clean-source/stage1/cohere/markdownblocks',env=env,stdout=f,stderr=subprocess.STDOUT)
 events=[]
 for l in (out/('M1-alone-'+row+'.log')).read_text().splitlines():
  try:events.append(json.loads(l))
  except:pass
 result=dict(id='M1',row=row,command=cmd,wall=time.time()-start,exit=r.returncode,failures=[e.get('Test') for e in events if e.get('Action')=='fail' and e.get('Test')],cooked=any('panic: test timed out' in e.get('Output','') for e in events));results.append(result);(out/'reruns.json').write_text(json.dumps(results,indent=2));print(result,flush=True)

import pathlib,re,json,difflib,subprocess,time,os
root=pathlib.Path('/workspace/adamic'); out=root/'review/test-audit/stage1-cohere-lint-harness'; pkg='./stage1/cohere/lint/'; jsx='stage1/cohere/lint/jsx_shards_test.go'
spec=[]
def body(id,file,name,new,regex):
 s=(root/file).read_text(); start=s.index('func '+name+'('); a=s.index('{',start); nextfunc=s.find('\nfunc ',a+1); end=(nextfunc if nextfunc>=0 else len(s)); b=s.rfind('\n}',a,end)+2
 old=s[a:b]
 spec.append(dict(id=id,file=file,line=s[:a].count('\n')+1,old=old,new='{\n'+new+'\n}',kind='probe',regex=regex,change='empty-answer entry probe for '+name))
body('P2',jsx,'jsxGoOracle','\treturn ""','^TestProduct_jsx_(membership|parser)$')
body('P3',jsx,'jsxTreeLowered','\treturn ""','^TestProduct_jsx_tree_lowered$')
body('P4',jsx,'jsxTreeNative','\treturn ""','^TestProduct_jsx_tree_native$')
body('P5',jsx,'jsxPrepareTrees','\treturn ""','^(TestProduct_jsx_tree_setup|TestJsxLintTrees_Setup|TestJsxLintTreesSetupIsolation)$')
body('P6',jsx,'jsxFetchTrees','\treturn jsxTreeBundle{}, ""','^TestJsxLintTreesUnion$')
body('P7','stage1/cohere/lint/emitted_javascript_shards_test.go','emittedMismatchSetup','','^TestEmittedJavaScriptMismatch_Setup$')
body('P8','stage1/cohere/lint/emitted_javascript_shards_test.go','emittedMismatchUnion','','^TestEmittedJavaScriptMismatch$')
body('P9','stage1/cohere/lint/suggestion_alongside_shards_test.go','suggestionAlongsideUnion','','^TestSuggestionAlongsideAutomaticFix$')
body('P10',jsx,'jsxTreeShards','\treturn nil, nil','^TestJsxLintTreesShardCoverage$')
body('P11',jsx,'jsxValidateUnion','\treturn nil','^TestJsxLintTreesShardCoverage$')
body('P12',jsx,'jsxShardSelection','\treturn nil, nil','^TestJsxLintTreesShardCoverage$')
(out/'probe-plan.json').write_text(json.dumps(spec,indent=2)+'\n')
results=json.loads((out/"probe-results.json").read_text())
for m in spec:
 if m["id"]=="P2":continue
 path=root/m['file']; s=path.read_text(); modified=s.replace(m['old'],m['new'],1)
 if m['id']=='P3':
  for imp in ['github.com/system-inc/adamic/internal/load','github.com/system-inc/adamic/internal/lower']: modified=modified.replace('\t"'+imp+'"\n','')
 if m['id'] in ['P8','P9']:
  for imp in ['go/ast','go/parser','go/token'] + (['strconv'] if m['id']=='P9' else []): modified=modified.replace('\t"'+imp+'"\n','')
 diff=''.join(difflib.unified_diff(s.splitlines(True),modified.splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file']))
 (out/'diffs'/f"{m['id']}.diff").write_text(diff)
 path.write_text(modified)
 try:
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u109/cache/'+m['id'];env['ADAMIC_NATIVE_SPLIT']='1'
  start=time.monotonic()
  with (out/f"{m['id']}-vet.log").open('w') as f:v=subprocess.run(['timeout','90','go','vet',pkg],cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
  if v.returncode:raise RuntimeError('vet failed '+m['id'])
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s',pkg,'-run',m['regex']]
  with (out/f"{m['id']}.log").open('w') as f:p=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
  if m['id']=='P12' and 'panic:' in (out/'P12.log').read_text():
   with (out/'P12-recheck.log').open('w') as f:subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
  results.append(dict(id=m['id'],exit=p.returncode,wall=time.monotonic()-start,command=cmd)); (out/'probe-results.json').write_text(json.dumps(results,indent=2)+'\n')
  print(m['id'],p.returncode,results[-1]['wall'],flush=True)
 finally:path.write_text(s)

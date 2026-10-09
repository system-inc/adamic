import subprocess,time,json,pathlib,re
root=pathlib.Path('/workspace/adamic'); out=root/'review/test-audit/stage1-cohere-lint-harness'
rows=['TestEmittedJavaScriptMismatch','TestEmittedJavaScriptMismatch_Setup','TestCompleteSuggestionSerialization','TestSuggestionAlongsideAutomaticFix','TestJsxLintTrees family','TestProduct_jsx_oracle family','TestProduct_jsx_tree_lowered','TestProduct_jsx_tree_native','TestProduct_jsx_tree_setup','TestJsxLintTreesSetupIsolation','TestJsxLintTreesShardCoverage','TestJsxLintTrees_Setup','TestJsxLintTreesUnion','TestJsxLintTreesShardDisagreement']
patterns={'TestJsxLintTrees family':'TestJsxLintTrees_[0-9]{3}','TestProduct_jsx_oracle family':'TestProduct_jsx_(membership|parser)'}
records=[]
for i,row in enumerate(rows):
 for trial in range(3):
  regex='^('+patterns.get(row,row)+')$'
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/','-run',regex]
  start=time.monotonic(); log=out/f'timing-{i:02}-{trial+1}.log'
  with log.open('w') as f: p=subprocess.run(cmd,cwd=root,stdout=f,stderr=subprocess.STDOUT)
  elapsed=time.monotonic()-start
  seconds=None
  for line in log.read_text().splitlines():
   try:r=json.loads(line)
   except:continue
   if r.get('Action')=='pass' and 'Test' not in r: seconds=r.get('Elapsed')
  records.append(dict(test=row,trial=trial+1,seconds=seconds,wall=elapsed,exit=p.returncode,command=cmd,log=log.name))
  (out/'timings.json').write_text(json.dumps(records,indent=2)+'\n')
  print(row,trial+1,seconds,p.returncode,flush=True)
  if p.returncode: break

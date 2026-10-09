from pathlib import Path
import subprocess,json,difflib,time,os
root=Path('/workspace/adamic');out=root/'review/test-audit/stage1-cohere-graphql';env=os.environ.copy();env['ADAMIC_GRAPHQL_LIBRARY']='/tmp/u096/library'
metrics=[]
for id,file in [('P01','stage1/cohere/graphql/main.ts'),('P02','internal/lower/lower.go')]:
 p=root/file;s=p.read_text()
 t=s[:s.index('const casesPath =')] if id=='P01' else s[:s.index('\n\tfiles := program.Files()')]+ '\n\treturn nil, nil' + s[s.index('\n}\n\ntype lowering'):]
 if id=='P02':t=t.replace('\t"fmt"\n','').replace('\t"path/filepath"\n','')
 (out/'diffs'/f'{id}.diff').write_text(''.join(difflib.unified_diff(s.splitlines(True),t.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 p.write_text(t);env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u096/cache/'+id
 try:
  if id=='P02':
   with (out/'logs'/'P02-vet.log').open('w') as log:subprocess.run(['go','vet','./internal/lower/'],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT,check=True)
  if id=='P01': runs=[('P01','.')]
  else: runs=[('P02-gap1','^TestEachGapStandsWhereGapsMdSaysItDoes$/^gaps$/^1_throwing_function_value.ts$'),('P02-gap2','^TestEachGapStandsWhereGapsMdSaysItDoes$/^gaps$/^2_error_made_elsewhere.ts$')]
  for label,regex in runs:
   start=time.monotonic()
   with (out/'logs'/f'{label}.log').open('w') as log:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/graphql/','-run',regex],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
   metrics.append(dict(id=label,wall=time.monotonic()-start,exit=r.returncode));(out/'probe-times.json').write_text(json.dumps(metrics,indent=2))
 finally:p.write_text(s)

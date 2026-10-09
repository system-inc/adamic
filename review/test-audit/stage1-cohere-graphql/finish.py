from pathlib import Path
import os,subprocess,json,time,difflib
root=Path('/workspace/adamic');out=root/'review/test-audit/stage1-cohere-graphql';env=os.environ.copy();env['ADAMIC_GRAPHQL_LIBRARY']='/tmp/u096/library';metrics=[]
for m in json.loads((out/'plan.json').read_text()):
 p=root/m['file'];s=p.read_text();p.write_text(s.replace(m['old'],m['new']));env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u096/cache/'+m['id']
 try:
  subprocess.run(['git','apply','--reverse','--check',str(out/'diffs'/f"{m['id']}.diff")],cwd=root,check=True)
  with (out/'logs'/f"{m['id']}-build.log").open('w') as log:
   start=time.monotonic();r=subprocess.run(['timeout','90','go','run',str(out/'build.go'),str(root/'stage1/cohere/graphql/main.ts'),'/tmp/u096/'+m['id']],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
  metrics.append(dict(id=m['id'],wall=time.monotonic()-start,exit=r.returncode));(out/'build-times.json').write_text(json.dumps(metrics,indent=2))
  assert r.returncode==0
 finally:p.write_text(s)
p=root/'stage1/cohere/graphql/graphql_test.go';s=p.read_text();a=s.index('func firstDifference(');b=s.index('\nfunc lowered(',a);t=s[:a]+'func firstDifference(got string, want string) string { return "" }\n'+s[b:];p.write_text(t)
(out/'diffs'/'W01-witness-only.diff').write_text(''.join(difflib.unified_diff(s.splitlines(True),t.splitlines(True),fromfile='a/stage1/cohere/graphql/graphql_test.go',tofile='b/stage1/cohere/graphql/graphql_test.go')))
try:
 env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u096/cache/W01'
 with (out/'logs'/'W01.log').open('w') as log:r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/graphql/','-run','^TestThePortParsesAsGoCohereDoes$/^catches_'],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
 assert r.returncode==1
finally:p.write_text(s)
env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u096/cache/final'
with (out/'logs'/'final-baseline.log').open('w') as log:
 r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverpkg=./internal/lower,./internal/native','-coverprofile=/tmp/u096/final.cover','./stage1/cohere/graphql/','-run','.'],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
assert r.returncode==0
with (out/'all-Go-functions.txt').open('w') as log:subprocess.run(['go','tool','cover','-func=/tmp/u096/final.cover'],cwd=root,stdout=log,stderr=subprocess.STDOUT,check=True)

from pathlib import Path
import json,os,subprocess,difflib,time
root=Path('/workspace/adamic');out=root/'review/test-defend/stage1-typescript-parser-compiler_expressions_shards'
while not Path('/tmp/defend-parser/adaptive-done').exists():time.sleep(1)
p=root/'stage1/typescript/parser/grammar.ts';s=p.read_text();old='return 15;';new='return 13;';assert s.count(old)==1;(out/'D4.diff').write_text(''.join(difflib.unified_diff(s.splitlines(True),s.replace(old,new).splitlines(True),fromfile='a/stage1/typescript/parser/grammar.ts',tofile='b/stage1/typescript/parser/grammar.ts')))
env=os.environ.copy();env['ADAMIC_TYPESCRIPT_SOURCE']='/tmp/u030/typescript';env['ADAMIC_BUILD_CACHE_DIR']='/tmp/defend-parser/cache/D4';results=[]
try:
 p.write_text(s.replace(old,new));commands=[['timeout','120','go','run','./cmd/adamic','build','stage1/typescript/parser/main.ts','-o','/tmp/defend-parser/D4-native','--sanitize'],['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/typescript/parser/','-run','^(TestGeneratedExpressionsAgree|TestCompilerExpressionsAgree(Union|_[0-9]{3})|TestExpressionsAgree|TestWholeGeneratedAgrees|TestWholeCompilerAgrees|TestEveryTypeNodeKindAgrees|TestYieldLookaheadAgrees)$']]
 for i,cmd in enumerate(commands):
  start=time.monotonic()
  with open('/tmp/defend-parser/D4-'+str(i)+'.log','w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
  results.append(dict(command=' '.join(cmd),status=r.returncode,wall=time.monotonic()-start));(out/'D4-runs.json').write_text(json.dumps(results,indent=2));print(i,r.returncode,flush=True)
finally:p.write_text(s)
Path('/tmp/defend-parser/D4-done').write_text('done')

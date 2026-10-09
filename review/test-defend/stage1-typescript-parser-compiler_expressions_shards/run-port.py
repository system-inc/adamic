from pathlib import Path
import os,subprocess,json,time,difflib
root=Path('/workspace/adamic');out=root/'review/test-defend/stage1-typescript-parser-compiler_expressions_shards';p=root/'stage1/typescript/parser/parser.ts';s=p.read_text();old="new ParseNode('PrivateIdentifier', old.pos, old.end, [])";new="new ParseNode('Identifier', old.pos, old.end, [])";assert s.count(old)==1
(out/'D3.diff').write_text(''.join(difflib.unified_diff(s.splitlines(True),s.replace(old,new).splitlines(True),fromfile='a/stage1/typescript/parser/parser.ts',tofile='b/stage1/typescript/parser/parser.ts')))
base=os.environ.copy();base['ADAMIC_TYPESCRIPT_SOURCE']='/tmp/u030/typescript';base['ADAMIC_BUILD_CACHE_DIR']='/tmp/defend-parser/cache/D3';runs=[]
names=[l.strip() for l in Path('/tmp/defend-parser/list.log').read_text().splitlines() if l.startswith('Test')]
groups=[[],[],[],[]]
for n in names:
 i=0 if (n.startswith('TestCompilerExpressionsAgree') or n in ['TestStrongAstParentGap','TestPushSpreadGap','TestClassMethodInterfaceGap','TestClosedOptionalFunctionValueGap','TestClosedConditionalEmptyArrayGap','TestGeneratedExpressionsAgree','TestTypeOnlyImportCycleCompiles']) else (1 if ('Expressions' in n or n.startswith('TestProduct_ParserOracle')) else (2 if 'Jsx' in n else 3))
 groups[i].append(n)
try:
 p.write_text(s.replace(old,new));start=time.monotonic()
 cmd=['timeout','120','go','run','./cmd/adamic','build','stage1/typescript/parser/main.ts','-o','/tmp/defend-parser/D3-native','--sanitize']
 with open('/tmp/defend-parser/D3-native-build.log','w') as f:r=subprocess.run(cmd,env=base,stdout=f,stderr=subprocess.STDOUT)
 (out/'D3-build.json').write_text(json.dumps(dict(command=' '.join(cmd),status=r.returncode,wall=time.monotonic()-start),indent=2))
 if r.returncode:raise SystemExit('Native validation failed')
 for i,rows in enumerate(groups):
  regex='^('+'|'.join(rows)+')$';cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/typescript/parser/','-run',regex];start=time.monotonic()
  with open('/tmp/defend-parser/D3-group'+str(i)+'.log','w') as f:r=subprocess.run(cmd,env=base,stdout=f,stderr=subprocess.STDOUT)
  runs.append(dict(group=i,rows=rows,status=r.returncode,wall=time.monotonic()-start,command=' '.join(cmd)));(out/'D3-runs.json').write_text(json.dumps(runs,indent=2));print(i,r.returncode,runs[-1]['wall'],flush=True)
finally:p.write_text(s)
Path('/tmp/defend-parser/port-done').write_text('done')

import pathlib,subprocess,os,time,json
p=pathlib.Path('/workspace/adamic/review/test-defend/stage1-cohere-yaml-gaps')
groups={'core':['TestLexerGaps','TestStructuralPositionRefusal','TestClosedStringPresenceGap','TestClosedLexerGaps','TestClosedValuePresenceGap','TestLexerMatchesGo','TestPropsMatchGo','TestSharedSliceAppendMatchesNode','TestLexerMutants','TestPropsMutants','TestScalarMutants'],'other':['TestComposeMatchGo','TestComposeMutants','TestCSTMatchesGo','TestCSTMutants','TestSchemaMatchesGo','TestSchemaMutants','TestUnistMatchesGo','TestUnistMutants','TestWidthsMatchGo'],'formatter':['TestFormatterMatchesGo','TestBundledParserDifference'],'files':['TestFileDriverUnion'],'scalars':['TestScalarsMatchGoUnion']}
(p/'groups.json').write_text(json.dumps(groups,indent=2));results=[]
for name,rows in groups.items():
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/yaml/','-run','^('+'|'.join(rows)+')$'];start=time.monotonic()
 with (p/('baseline-'+name+'.log')).open('w') as f:r=subprocess.run(cmd,cwd='/workspace/adamic',env=dict(os.environ,ADAMIC_YAML_LIBRARY='/tmp/u152/library'),stdout=f,stderr=subprocess.STDOUT)
 results.append(dict(group=name,command=cmd,exit=r.returncode,wall=time.monotonic()-start));(p/'baseline-groups.json').write_text(json.dumps(results,indent=2));print(results[-1],flush=True)

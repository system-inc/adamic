import os,subprocess,json,time
from pathlib import Path
root=Path.cwd();ev=root/'review/compiler/loom-slice-3-2'
cases=[('cssstrings','TestMultiPushGap'),('json','TestClosedMultiplePushGap'),('markdownblocks','TestParserRepresentationProbes'),('selector','TestEachGapStandsWhereGapsMdSaysItDoes'),('yaml','TestLexerGaps|TestProduct_YAMLClosedGaps(Lowered|Native)'),('typescript/parser','TestPushSpreadGap'),('typescript/scanner','TestGapStandsWhereGapsMdSays'),('fixtures','TestFixturesAssignmentProofs|TestFixtureDirectoriesHaveTopLevelTests')]
results=[]
for package,tests in cases:
 path='./stage3/fixtures' if package=='fixtures' else './stage1/'+(package if package.startswith('typescript/') else 'cohere/'+package)
 command=['timeout','180','go','test',path,'-run','^('+tests+')$','-count=1','-timeout','90s','-parallel','4','-json']
 start=time.monotonic()
 with (ev/('gaps-'+package.replace('/','-')+'-green.jsonl')).open('w') as log: result=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,env={**os.environ,'GOMAXPROCS':'4','ADAMIC_GATE_UNCACHED':'1'})
 results.append(dict(package=path,command=command,exit=result.returncode,seconds=time.monotonic()-start));(ev/'gaps-green-results.json').write_text(json.dumps(results,indent=2)+'\n');print(path,result.returncode,flush=True)
 if result.returncode:break

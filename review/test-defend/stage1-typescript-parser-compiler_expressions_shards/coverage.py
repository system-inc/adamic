import os,subprocess,time,json
from pathlib import Path
base=os.environ.copy();base['ADAMIC_TYPESCRIPT_SOURCE']='/tmp/u030/typescript';results=[]
rows=['TestClosedConditionalEmptyArrayGap','TestPushSpreadGap','TestTypeOnlyImportCycleCompiles','TestClosedOptionalFunctionValueGap','TestGeneratedExpressionsAgree','TestCompilerExpressionsAgree']
for n in rows:
 regex='^TestCompilerExpressionsAgree(Union|_[0-9]{3})$' if n=='TestCompilerExpressionsAgree' else '^'+n+'$'
 env=base.copy();env['NODE_V8_COVERAGE']='/tmp/defend-parser/v8/'+n
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverpkg=./internal/load,./internal/lower,./internal/native,./internal/javascript','-coverprofile=/tmp/defend-parser/'+n+'.cover','./stage1/typescript/parser/','-run',regex];start=time.monotonic()
 with open('/tmp/defend-parser/'+n+'-coverage.log','w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
 results.append(dict(test=n,status=r.returncode,wall=time.monotonic()-start,command=' '.join(cmd)));Path('/tmp/defend-parser/coverage-runs.json').write_text(json.dumps(results,indent=2));print(n,r.returncode,results[-1]['wall'],flush=True)
 if r.returncode and 'panic: test timed out' not in Path('/tmp/defend-parser/'+n+'-coverage.log').read_text():break

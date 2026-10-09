import pathlib,os,subprocess,json,time
p=pathlib.Path('/workspace/adamic/review/test-defend/stage1-cohere-estree-deep/session');env=os.environ.copy();env.update(ADAMIC_ESTREE_LIBRARY='/tmp/estree-deep-defense/library',ADAMIC_NATIVE_SPLIT='1')
results=[]
for name in ['TestDeepGrammar','TestGeneratedAgreement','TestDecoratedExports','TestRecoveredExpressions']:
 env['NODE_V8_COVERAGE']=str(p/('v8-'+name))
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverpkg=github.com/system-inc/adamic/internal/lower,github.com/system-inc/adamic/internal/native','-coverprofile='+str(p/(name+'.cover')),'./stage1/cohere/estree/','-run','^'+name+'$']
 start=time.monotonic()
 with (p/(name+'-coverage.log')).open('w') as f:r=subprocess.run(cmd,cwd='/workspace/adamic',env=env,stdout=f,stderr=subprocess.STDOUT)
 results.append(dict(test=name,command=cmd,wall=time.monotonic()-start,exit=r.returncode));(p/'coverage-runs.json').write_text(json.dumps(results,indent=2))
 if r.returncode:break

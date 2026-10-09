import os,pathlib,subprocess,json,time
p=pathlib.Path('/workspace/adamic/review/test-defend/stage1-typescript-parser-jsx_rejection');env=os.environ.copy();env['TMPDIR']='/workspace/scratch/defend-jsx/tmp';pathlib.Path(env['TMPDIR']).mkdir(exist_ok=True);env['ADAMIC_BUILD_CACHE_DIR']='/workspace/scratch/defend-jsx/cache/clean'
runs=[]
for name in ['TestJsxMemberNameRejection','TestJsxNode','TestJsxNative']:
 env['NODE_V8_COVERAGE']='/workspace/scratch/defend-jsx/v8/'+name;pathlib.Path(env['NODE_V8_COVERAGE']).mkdir(parents=True,exist_ok=True)
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverpkg=github.com/system-inc/adamic/internal/load,github.com/system-inc/adamic/internal/lower,github.com/system-inc/adamic/internal/native','-coverprofile='+str(p/(name+'.cover')),'./stage1/typescript/parser/','-run','^'+name+'$'];t=time.monotonic()
 with (p/(name+'-coverage.log')).open('w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
 runs.append(dict(test=name,command=cmd,exit=r.returncode,wall=time.monotonic()-t));(p/'coverage-runs.json').write_text(json.dumps(runs,indent=2));print(name,r.returncode,flush=True)

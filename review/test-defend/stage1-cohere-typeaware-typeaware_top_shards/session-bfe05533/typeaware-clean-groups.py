import pathlib,subprocess,json,time,os
R=pathlib.Path('/workspace/adamic');P=R/'review/test-defend/stage1-cohere-typeaware-typeaware_top_shards/session-bfe05533';scope=json.loads((P/'matrix-scope.json').read_text());groups=[dict(id='baseline-agreement',rows=[n for n in scope['reached_tests'] if not n.endswith('_010')]),dict(id='baseline-kind-witness',rows=['TestTypeAwareAgreementAndMutants_010'])];runs=[]
for g in groups:
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/typeaware/','-run','^('+'|'.join(g['rows'])+')$','-coverpkg=github.com/system-inc/adamic/stage1/cohere/typeaware','-coverprofile='+str(P/(g['id']+'.cover'))];t=time.monotonic()
 with (P/(g['id']+'.log')).open('w') as f:r=subprocess.run(cmd,cwd=R,stdout=f,stderr=subprocess.STDOUT)
 runs.append(dict(id=g['id'],rows=g['rows'],seconds=time.monotonic()-t,exit=r.returncode,command=cmd));(P/'baseline-groups.json').write_text(json.dumps(runs,indent=2));print(g['id'],r.returncode,time.monotonic()-t,flush=True)
 if r.returncode:break

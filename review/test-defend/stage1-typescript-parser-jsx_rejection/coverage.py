import pathlib,os,time,subprocess,json
E=pathlib.Path('/tmp/d159/evidence');meta=[]
for name in ['TestCompilerExpressionsAgree_Setup','TestPerformance','TestWholePerformance']:
 env=os.environ.copy();env.update(ADAMIC_TYPESCRIPT_SOURCE='/tmp/u159/typescript',ADAMIC_PARSER_BENCH='1',ADAMIC_BUILD_CACHE_DIR='/tmp/d159/cache/coverage/'+name,NODE_V8_COVERAGE='/tmp/d159/v8/'+name)
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverpkg=./internal/native,./internal/lower,./internal/load','-coverprofile='+str(E/(name+'.cover')),'./stage1/typescript/parser/','-run','^'+name+'$'];start=time.monotonic()
 with open(E/('coverage-'+name+'.log'),'w') as out: result=subprocess.run(cmd,cwd='/workspace/adamic',env=env,stdout=out,stderr=subprocess.STDOUT)
 meta.append(dict(test=name,command=cmd,exit=result.returncode,wall=time.monotonic()-start));(E/'coverage-runs.json').write_text(json.dumps(meta,indent=2));print(name,result.returncode,meta[-1]['wall'],flush=True)

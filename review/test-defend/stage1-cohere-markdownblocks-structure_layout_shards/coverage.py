import os,pathlib,time,json,subprocess
p=pathlib.Path('/tmp/defend-markdown');env=os.environ.copy();env['ADAMIC_MARKDOWNWIDTH_DEPS']=str(p/'width-deps');env['ADAMIC_NATIVE_SPLIT']='1';runs=[]
for name,regex in [('structure','^TestMarkdownStructureLayout(Union|_[0-9]{3})$'),('table','^TestMarkdownTableLayout(Union|_[0-9]{3})$'),('width','^TestMarkdownUnicodeWidths$')]:
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','-coverpkg=./internal/load,./internal/lower,./internal/native,./internal/javascript','-coverprofile='+str(p/(name+'.cover')),'./stage1/cohere/markdownblocks/','-run',regex]
 env['NODE_V8_COVERAGE']=str(p/'v8'/name);start=time.monotonic()
 with (p/(name+'-clean.log')).open('w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
 runs.append(dict(name=name,command=' '.join(cmd),status=r.returncode,wall=time.monotonic()-start));(p/'coverage-runs.json').write_text(json.dumps(runs,indent=2));print(name,r.returncode,flush=True)
 if r.returncode:break

import pathlib,json,subprocess,time,os
p=pathlib.Path('review/test-audit/stage1-cohere-css-top_level_shards');base=json.loads((p/'original-source.json').read_text());env=os.environ.copy();env.update(ADAMIC_CSS_FIXTURES='/tmp/u080-css-fixtures',ADAMIC_CSS_LIBRARY='/tmp/u080-css-library');results=[]
for id in ['M01','M02','M03','M04','P01']:
 for f,s in base.items():pathlib.Path(f).write_text(s)
 subprocess.run(['git','apply','--check',str(p/(id+'.diff'))],check=True);subprocess.run(['git','apply',str(p/(id+'.diff'))],check=True)
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/css/','-run','^TestProduct_CSSParserParserNative$'];start=time.monotonic()
 with (p/(id+'-build.log')).open('w') as log:r=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
 results.append({'id':id,'command':' '.join(cmd),'exit':r.returncode,'wall':time.monotonic()-start});(p/'standalone-validation.json').write_text(json.dumps(results,indent=2));print(id,results[-1],flush=True)
for f,s in base.items():pathlib.Path(f).write_text(s)

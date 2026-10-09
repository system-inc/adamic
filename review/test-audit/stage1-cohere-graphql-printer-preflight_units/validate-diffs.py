import pathlib,json,subprocess,os,time,shlex
p=pathlib.Path('/tmp/u099/evidence');base=json.loads((p/'original-source.json').read_text());env=os.environ.copy();env['ADAMIC_GRAPHQL_PRETTIER']='/tmp/u099-prettier';records=[]
try:
 for id in ['M01','M02','M03','M04','P01']:
  for f,text in base.items():pathlib.Path(f).write_text(text)
  subprocess.run(['git','apply','--check',str(p/(id+'.diff'))],check=True);subprocess.run(['git','apply',str(p/(id+'.diff'))],check=True)
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/graphql/printer/','-run','^TestProduct_GraphQLPrinterSanitized$'];start=time.monotonic()
  with (p/(id+'-build.log')).open('w') as log:r=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
  records.append({'id':id,'command':'ADAMIC_GRAPHQL_PRETTIER=/tmp/u099-prettier '+shlex.join(cmd),'exit':r.returncode,'wall':time.monotonic()-start});(p/'standalone-validation.json').write_text(json.dumps(records,indent=2));print(id,records[-1],flush=True)
finally:
 for f,text in base.items():pathlib.Path(f).write_text(text)

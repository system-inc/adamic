from pathlib import Path
import json,difflib,subprocess,os,time
root=Path('/workspace/adamic');out=root/'review/test-defend/stage1-typescript-parser-compiler_expressions_shards';base=os.environ.copy();base['ADAMIC_TYPESCRIPT_SOURCE']='/tmp/u030/typescript'
menu=[('D1','internal/javascript/javascript.go','e.value(expression.WhenTrue) + " : " + e.value(expression.WhenNot)','e.value(expression.WhenNot) + " : " + e.value(expression.WhenTrue)','^(TestClosedConditionalEmptyArrayGap|TestClosedOptionalFunctionValueGap)$'),('D2','internal/lower/class_inheritance.go','} else if instance.base != nil {','} else if instance.base == nil {','^(TestTypeOnlyImportCycleCompiles|TestClassMethodInterfaceGap)$')];records=[]
for id,file,old,new,regex in menu:
 s=(root/file).read_text();assert s.count(old)==1
 edited=s.replace(old,new,1);scratch=Path('/tmp/defend-parser')/(id+'.go');scratch.write_text(edited);overlay=Path('/tmp/defend-parser')/(id+'-overlay.json');overlay.write_text(json.dumps({'Replace':{str(root/file):str(scratch)}}));(out/(id+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),edited.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 with open('/tmp/defend-parser/'+id+'-vet.log','w') as f:vet=subprocess.run(['go','vet','-overlay='+str(overlay),'./'+file.rsplit('/',1)[0]+'/'],stdout=f,stderr=subprocess.STDOUT)
 env=base.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/defend-parser/cache/'+id;cmd=['timeout','120','go','test','-overlay='+str(overlay),'-json','-count=1','-timeout','90s','./stage1/typescript/parser/','-run',regex]
 start=time.monotonic()
 with open('/tmp/defend-parser/'+id+'.log','w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
 record=dict(id=id,file=file,line=s[:s.index(old)].count('\n')+1,old=old,new=new,regex=regex,command=' '.join(cmd),status=r.returncode,vet=vet.returncode,wall=time.monotonic()-start);records.append(record);(out/'go-mutants.json').write_text(json.dumps(records,indent=2));print(id,r.returncode,flush=True)
 if 'panic:' in Path('/tmp/defend-parser/'+id+'.log').read_text():
  for n in ['TestTypeOnlyImportCycleCompiles','TestClassMethodInterfaceGap']:
   cmd[-1]='^'+n+'$'
   with open('/tmp/defend-parser/'+id+'-'+n+'.log','w') as f:subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)

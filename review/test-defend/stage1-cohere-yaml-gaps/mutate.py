import pathlib,subprocess,time,json,os
root=pathlib.Path('/workspace/adamic');p=root/'review/test-defend/stage1-cohere-yaml-gaps'
ident='D01';file='internal/native/runtime/string_share.c';old='shared->capacity = 0;';new='shared->capacity = size + 1;'
f=root/file;source=f.read_text();assert source.count(old)==1
(p/'plan.json').write_text(json.dumps(dict(mutant=ident,file_line=file+':'+str(source[:source.index(old)].count('\n')+1),change=new,menu='change constant',difference='A shared slice is given one byte of append room in its owner, violating immutable alias ownership.'),indent=2))
f.write_text(source.replace(old,new,1))
try:
 (p/(ident+'.diff')).write_bytes(subprocess.check_output(['git','diff','--',file],cwd=root))
 flags=['-std=c11','-Wall','-Wextra','-Werror','-Wcast-function-type-strict','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O2']
 cmd=['clang']+flags+['-I','internal/native/runtime','-c',file,'-o','/tmp/defend-yaml/D01-string-share.o']
 with (p/(ident+'.clang.log')).open('w') as log:r=subprocess.run(cmd,cwd=root,stdout=log,stderr=subprocess.STDOUT)
 assert r.returncode==0
 groups=json.loads((p/'groups.json').read_text());results=[]
 for name,rows in groups.items():
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/yaml/','-run',('^'+rows[0].replace('/','$/^')+'$' if len(rows)==1 and '/' in rows[0] else '^('+'|'.join(rows)+')$')];start=time.monotonic()
  env=dict(os.environ,ADAMIC_YAML_LIBRARY='/tmp/u152/library',ADAMIC_BUILD_CACHE_DIR='/tmp/defend-yaml/cache/'+ident)
  with (p/(ident+'-'+name+'.log')).open('w') as log:r=subprocess.run(cmd,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
  events=[]
  for line in (p/(ident+'-'+name+'.log')).read_text().splitlines():
   try:events.append(json.loads(line))
   except:pass
  failed=sorted({x['Test'].split('/')[0] for x in events if x.get('Test') and x['Action']=='fail'})
  results.append(dict(group=name,command=cmd,exit=r.returncode,wall=time.monotonic()-start,rows_failed=failed));(p/(ident+'-runs.json')).write_text(json.dumps(results,indent=2));print(results[-1],flush=True)
finally:f.write_text(source)

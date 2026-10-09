import pathlib,json,subprocess,time,os,difflib
p=pathlib.Path('review/test-defend/internal-native-radix');(p/'diffs').mkdir(exist_ok=True);scope=json.loads((p/'scope.json').read_text())
plan=[('D01','internal/regexp/matcher.go','{0x2028, 0x2029}','{0x2029, 0x2029}','TestRegExpBytecodeTest262','Off by one: stop excluding U+2028 from dot without dotAll.'),('D02','internal/native/runtime/regexp.c','i->unbounded || reg.count < i->maximum','i->unbounded || i->maximum == 0 || reg.count < i->maximum','TestRegExpBytecodeRandomNode family','Flip repeat-body condition: interpret closed maximum zero as unbounded.')]
items=[]
for ident,file,old,new,target,change in plan:
 original=pathlib.Path(file).read_text();assert original.count(old)==1
 line=original[:original.index(old)].count('\n')+1
 (p/'diffs'/(ident+'.diff')).write_text(''.join(difflib.unified_diff(original.splitlines(True),original.replace(old,new).splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
 items.append(dict(mutant=ident,file=file,file_line=f'{file}:{line}',old=old,new=new,target=target,change=change))
(p/'mutant-plan.json').write_text(json.dumps(items,indent=2))
results=[]
for item in items:
 ident=item['mutant'];file=pathlib.Path(item['file']);original=file.read_text()
 try:
  file.write_text(original.replace(item['old'],item['new']))
  if file.suffix=='.go':cmd=['go','vet','./internal/regexp/']
  else:cmd=['clang','-std=c11','-Wall','-Wextra','-Werror','-Wcast-function-type-strict','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-DADAMIC_COUNT','-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all','-I','internal/native/runtime','-c',str(file),'-o','/tmp/native-defend-'+ident+'.o']
  t=time.monotonic()
  with (p/'logs'/('compile-'+ident+'.log')).open('w') as log:code=subprocess.call(cmd,stdout=log,stderr=log)
  assert code==0,'compile failed'
  item['compile_command']=cmd;item['compile_exit']=code;item['compile_seconds']=time.monotonic()-t
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run','^('+'|'.join(scope['matrix_members'])+')$'];env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/defend-native-radix/cache/'+ident)
  t=time.monotonic()
  with (p/'logs'/(ident+'.log')).open('w') as log:code=subprocess.call(cmd,env=env,stdout=log,stderr=log)
  events=[]
  for s in (p/'logs'/(ident+'.log')).read_text().splitlines():
   try:events.append(json.loads(s))
   except ValueError:pass
  failed=sorted(set(e['Test'].split('/')[0] for e in events if e.get('Action')=='fail' and e.get('Test')))
  grouped=sorted(set('TestRegExpBytecodeRandomNode family' if n.startswith('TestRegExpBytecodeRandomNodeUnit') else n for n in failed))
  results.append(dict(item,command=cmd,env={'ADAMIC_BUILD_CACHE_DIR':env['ADAMIC_BUILD_CACHE_DIR']},exit=code,wall_seconds=time.monotonic()-t,failed_members=failed,rows_failed=grouped,passed_members=sorted(e['Test'] for e in events if e.get('Action')=='pass' and e.get('Test') and '/' not in e['Test']),errors=[e for e in events if e.get('OutputType')=='error'],package_result=[e for e in events if e.get('Action') in ['pass','fail'] and not e.get('Test')]))
  (p/'matrix.json').write_text(json.dumps(results,indent=2));print(ident,code,results[-1]['wall_seconds'],grouped,flush=True)
 finally:file.write_text(original)

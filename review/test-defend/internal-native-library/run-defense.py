import pathlib,subprocess,json,time,os
p=pathlib.Path('review/test-defend/internal-native-library')
mutations=[
 ('D1','internal/native/library.go','if info, statErr := os.Stat(library); statErr != nil || !info.Mode().IsRegular() {','if info, statErr := os.Stat(library); statErr == nil || !info.Mode().IsRegular() {','cache-rows.json'),
 ('D2','internal/native/emit_statements.go','''\t\tif e.elementBorrows[e.at] {
\t\t\te.line("%s %s = %s;", cType(statement.Element), e.localName(statement.Local), element)
\t\t} else {
\t\t\te.declareLocal(statement.Local, element, false)
\t\t}''','''\t\te.declareLocal(statement.Local, element, false)''','emit-rows.json'),
 ('D3','internal/native/runtime/map_set.c','return hash & 0x3fffffffu;','return hash & 0xffu;','matrix-rows.json')]
import sys
if len(sys.argv)>1 and sys.argv[1]=='D4':
 mutations=[('D4','internal/native/runtime/map_set.c','return hash & 0x3fffffffu;','return hash & 0x3ffu;','matrix-rows.json')]
 results=json.loads((p/'matrix.json').read_text())
else:
 results=[]
for mid,file,before,after,rowsfile in mutations:
 f=pathlib.Path(file);original=f.read_text();assert before in original
 line=original[:original.index(before)].count('\n')+1
 f.write_text(original.replace(before,after,1))
 try:
  (p/(mid+'.diff')).write_bytes(subprocess.check_output(['git','diff','--',file]))
  env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/defend-native-library/cache/'+mid)
  env['GOCACHE']=subprocess.check_output(['go','env','GOCACHE']).decode().strip()
  validation=['go','vet','./internal/native/'] if file.endswith('.go') else ['clang','-std=c11','-Wall','-Wextra','-Werror','-Wcast-function-type-strict','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all','-I','internal/native/runtime','-c',file,'-o','/tmp/native-library-D3.o']
  start=time.monotonic()
  with (p/(mid+'-compile.log')).open('w') as log: rc=subprocess.run(validation,stdout=log,stderr=subprocess.STDOUT,env=env).returncode
  checktime=time.monotonic()-start
  assert rc==0,(mid,rc)
  rows=json.loads((p/rowsfile).read_text());pattern='^('+'|'.join(rows)+')$'
  command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run',pattern]
  start=time.monotonic()
  with (p/(mid+'.log')).open('w') as log: rc=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,env=env).returncode
  wall=time.monotonic()-start
  events=[]
  for l in (p/(mid+'.log')).read_text().splitlines():
   try: events.append(json.loads(l))
   except ValueError: pass
  failed=sorted({e['Test'] for e in events if e['Action']=='fail' and 'Test' in e and '/' not in e['Test']})
  passed=sorted({e['Test'] for e in events if e['Action']=='pass' and 'Test' in e and '/' not in e['Test']})
  skipped=sorted({e['Test'] for e in events if e['Action']=='skip' and 'Test' in e and '/' not in e['Test']})
  record=dict(mutant=mid,file_line=file+':'+str(line),change=before+' => '+after,rows_failed=failed,rows_passed=passed,rows_skipped=skipped,rows_unknown=sorted(set(rows)-set(failed+passed+skipped)),command=command,validation=validation,validation_seconds=checktime,wall_seconds=wall,exit=rc,matrix_rows=rows)
  results.append(record);(p/'matrix.json').write_text(json.dumps(results,indent=2)+'\n')
  print(mid,failed,'pass',len(passed),'skip',len(skipped),'unknown',len(record['rows_unknown']),'wall',round(wall,3),flush=True)
 finally: f.write_text(original)

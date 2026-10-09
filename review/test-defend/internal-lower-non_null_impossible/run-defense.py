import pathlib,subprocess,json,time,os,sys
p=pathlib.Path('review/test-defend/internal-lower-non_null_impossible')
mutations=[('D1','internal/lower/class_inheritance.go','strconv.Quote(next.Parameters()[index].Name)','strconv.Quote(next.Parameters()[0].Name)','Change diagnostic parameter-index option to the first parameter'),('D2','internal/lower/optional_widening.go','property.Flags&ast.SymbolFlagsOptional != 0 && !isClassInstance(source)','property.Flags&ast.SymbolFlagsOptional != 0','Drop class-instance exemption condition'),('D3','internal/lower/optional_widening.go','case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:','case ast.KindNewExpression, ast.KindArrayLiteralExpression:','Change fresh-object syntax constant to constructor syntax')]
if len(sys.argv)>1:
 mutations=[('D4','internal/lower/optional_widening.go','case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:','case ast.KindIdentifier, ast.KindArrayLiteralExpression:','Change fresh-object syntax constant to identifier syntax')]
 results=json.loads((p/'matrix.json').read_text())
else: results=[]
for mid,file,before,after,description in mutations:
 f=pathlib.Path(file);original=f.read_text();assert original.count(before)==1
 line=original[:original.index(before)].count('\n')+1;f.write_text(original.replace(before,after,1))
 try:
  (p/(mid+'.diff')).write_bytes(subprocess.check_output(['git','diff','--',file]))
  env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/defend-lower-non-null/cache/'+mid)
  start=time.monotonic()
  with (p/(mid+'-vet.log')).open('w') as log: rc=subprocess.run(['go','vet','./internal/lower/'],env=env,stdout=log,stderr=subprocess.STDOUT).returncode
  check=time.monotonic()-start;assert rc==0,(mid,rc)
  command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'];start=time.monotonic()
  with (p/(mid+'.log')).open('w') as log:rc=subprocess.run(command,env=env,stdout=log,stderr=subprocess.STDOUT).returncode
  wall=time.monotonic()-start;events=[]
  for l in (p/(mid+'.log')).read_text().splitlines():
   try: events.append(json.loads(l))
   except:pass
  failed=sorted({e['Test'] for e in events if e['Action']=='fail' and 'Test' in e and '/' not in e['Test']})
  passed=sorted({e['Test'] for e in events if e['Action']=='pass' and 'Test' in e and '/' not in e['Test']})
  skipped=sorted({e['Test'] for e in events if e['Action']=='skip' and 'Test' in e and '/' not in e['Test']})
  allrows=[l for l in (p/'test-list.log').read_text().splitlines() if l.startswith('Test')]
  record=dict(mutant=mid,file_line=file+':'+str(line),change=description,before=before,after=after,rows_failed=failed,rows_passed=passed,rows_skipped=skipped,rows_unknown=sorted(set(allrows)-set(failed+passed+skipped)),command=command,cache=env['ADAMIC_BUILD_CACHE_DIR'],vet_seconds=check,wall_seconds=wall,exit=rc)
  results.append(record);(p/'matrix.json').write_text(json.dumps(results,indent=2)+'\n');print(mid,failed,'passed',len(passed),'skipped',len(skipped),'unknown',len(record['rows_unknown']),'wall',round(wall,3),flush=True)
 finally:f.write_text(original)

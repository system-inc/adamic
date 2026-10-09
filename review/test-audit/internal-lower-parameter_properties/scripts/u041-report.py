from pathlib import Path
import subprocess,json,re,statistics
r=Path('review/test-audit/internal-lower-parameter_properties');scope=json.loads((r/'scope.json').read_text());runs=json.loads((r/'run-results.json').read_text());menu=json.loads((r/'menu.json').read_text());matrix={};timing={};subsumer='TestClassFeaturesPrivateChecker'
def events(log):
 es=[]
 for l in (r/log).read_text().splitlines():
  try:es.append(json.loads(l))
  except:pass
 return es
def failing_line(log,row):
 es=events(log);outputs=[e.get('Output','').strip() for e in es if e.get('Test','').split('/')[0]==row]
 for line in outputs:
  if re.search(r'_test.go:\d+:',line):return line
 for e in es:
  if 'panic:' in e.get('Output',''):return e['Output'].strip()
 return 'failed process; see '+log
for row in scope+[subsumer]:
 vals=[float(re.search(r'\t([0-9.]+)s',(r/(row+'-time-'+str(i)+'.log')).read_text())[1]) for i in range(1,4)]
 timing[row]={'runs':vals,'median':statistics.median(vals)}
for m in menu:
 mid=m['id'];run=runs[mid];bounded='isolated'in run or 'bounded'in run
 if 'isolated'in run: failed=[row for row,v in run['isolated'].items() if v['exit']!=0];observed=scope
 else:
  es=events(run.get('bounded',run)['log']);failed=sorted({e['Test'].split('/')[0] for e in es if e['Action']=='fail' and 'Test'in e});observed=scope if bounded else [l for l in (r/'list.log').read_text().splitlines() if l.startswith('Test')]
 matrix[mid]={'failed_rows':failed,'bounded':bounded,'matrix_rows':observed,'logs':{},'failing_lines':{},'whole_command_wall_seconds':run['wall_seconds']}
 for row in scope+[subsumer]:
  if 'isolated'in run and row not in scope:continue
  log=run['isolated'][row]['log'] if 'isolated'in run else run.get('bounded',run)['log'];matrix[mid]['logs'][row]=log
  if row in failed:matrix[mid]['failing_lines'][row]=failing_line(log,row)
sets={row:{m['id'] for m in menu if m['id'].startswith('M') and row in matrix[m['id']]['failed_rows']} for row in scope+[subsumer]}
rows=[]
oracles={scope[0]:'Self-written Refused type and reason fragments across six unsafe parameter-property sources. Does not execute accepted programs or inspect initialized field values.',scope[1]:'Self-written load.CheckError type for readonly/private/protected writes and nil error for positive lowering. Negative cases accept any CheckError, not an exact diagnostic. The positive Lower subcase passes P1.',scope[2]:'Self-written NotYet type and this parameter used as a value fragment; no exact location or runtime output.',scope[3]:'Self-written helper argument/index-fix fragments plus fixture basename:5:10 location. No external expected-answer authority.'}
for row in scope:
 kills=sorted(sets[row]);unique=[mid for mid in kills if not matrix[mid]['bounded'] and matrix[mid]['failed_rows']==[row]];by=[];subs_seconds=None
 if unique:verdict='sacred'
 elif not kills:verdict='untrue'
 else:
  candidates=[x for x in scope+[subsumer] if x!=row and sets[row]<=sets[x]]
  if candidates:verdict='subsumed';by=[min(candidates,key=lambda x:(timing[x]['median'],x))];subs_seconds=timing[by[0]]['median']
  else:verdict='overlapping';by=sorted({x for x in scope+[subsumer] if x!=row and sets[x]&sets[row]})
 own=['P3'] if row==scope[3] else ['P1','P2'] if row==scope[1] else ['P1'];p_kills=[p for p in own if row in matrix[p]['failed_rows']]
 last=kills[-1] if kills else None;line=matrix[last]['failing_lines'][row] if last else None;log=matrix[last]['logs'][row] if last else None
 pattern='^'+row+'$' if last and matrix[last]['bounded'] else '.'
 command='ADAMIC_MUTANT='+last+' ADAMIC_BUILD_CACHE_DIR=/tmp/u041/cache/'+last+' timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run '+pattern+' > review/test-audit/internal-lower-parameter_properties/'+log+' 2>&1' if last else 'No production kill'
 file='internal/lower/predicate_refusals_test.go' if row==scope[3] else 'internal/lower/parameter_properties_test.go';b=subprocess.check_output(['git','show','origin/main:'+file],text=True);lineno=b[:b.index('func '+row+'(')].count('\n')+1
 rows.append(dict(test=row,package='internal/lower',file=file+':'+str(lineno),seconds=timing[row]['median'],oracle=oracles[row],oracle_kind='self',kills=kills,unique_kills=unique,last_proven_fail=(last+': '+line) if last else None,verdict=verdict,subsumed_by=by,mutants_in_matrix=12,probe_kills=p_kills,subsumer_seconds=subs_seconds,vacuous=not p_kills,vacuous_subcases=['positive mutable parameter-property Lower subcase passes P1'] if row==scope[1] else [],bounded=True,matrix_rows=scope,evidence=command+('; '+line if line else ''),subsumption_mutant_count=len(kills) if verdict=='subsumed' else None))
(r/'rows.json').write_text(json.dumps(rows,indent=2)+'\n');(r/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n');(r/'timings.json').write_text(json.dumps(timing,indent=2)+'\n')
print([(row['test'],row['verdict'],row['kills'],row['unique_kills']) for row in rows])

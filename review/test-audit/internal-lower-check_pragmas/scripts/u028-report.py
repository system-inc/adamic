from pathlib import Path
import subprocess,json,re,statistics
r=Path('review/test-audit/internal-lower-check_pragmas'); scope=json.loads((r/'scope.json').read_text()); runs=json.loads((r/'run-results.json').read_text()); menu=json.loads((r/'menu.json').read_text()); matrix={}; timing={}
def events(log):
 out=[]
 for l in (r/log).read_text().splitlines():
  try: out.append(json.loads(l))
  except: pass
 return out
def failureline(log,row):
 es=events(log)
 candidates=[e.get('Output','').strip() for e in es if e.get('Test','').split('/')[0]==row]
 for line in candidates:
  if re.search(r'_test.go:\d+:',line):return line
 for line in candidates:
  if 'panic:' in line:return line
 for e in es:
  if 'panic:' in e.get('Output',''):return e['Output'].strip()
 return next((c for c in candidates if c.startswith('--- FAIL')), 'process failed; inspect '+log)
for row in scope:
 vals=[float(re.search(r'\t([0-9.]+)s',(r/(row+'-time-'+str(i)+'.log')).read_text())[1]) for i in range(1,4)]
 timing[row]={'runs':vals,'median':statistics.median(vals)}
for m in menu:
 mid=m['id']; run=runs[mid]; bounded='isolated' in run or 'bounded' in run
 if 'isolated' in run:
  killed=[row for row,v in run['isolated'].items() if v['exit']!=0]
  observed=scope
 else:
  es=events(run.get('bounded',run)['log']); killed=sorted({e['Test'].split('/')[0] for e in es if e['Action']=='fail' and 'Test'in e})
  observed=scope if bounded else [l for l in (r/'list-retry.log').read_text().splitlines() if l.startswith('Test')]
 matrix[mid]={'failed_rows':killed,'bounded':bounded,'matrix_rows':observed,'failing_lines':{},'logs':{},'whole_command_wall_seconds':run['wall_seconds']}
 for row in scope:
  log=run['isolated'][row]['log'] if 'isolated'in run else run.get('bounded',run)['log']
  matrix[mid]['logs'][row]=log
  if row in killed:matrix[mid]['failing_lines'][row]=failureline(log,row)
sets={row:{m['id'] for m in menu if m['id'].startswith('M') and row in matrix[m['id']]['failed_rows']} for row in scope}
rows=[]
for row in scope:
 kills=sorted(sets[row]);unique=[mid for mid in kills if not matrix[mid]['bounded'] and matrix[mid]['failed_rows']==[row]]
 others=[other for other in scope if other!=row and sets[row] and sets[row]<=sets[other]]
 by=[];seconds=None
 if unique: verdict='sacred'
 elif not kills:verdict='untrue'
 elif others:
  verdict='subsumed';by=[min(others,key=lambda x:(timing[x]['median'],x))];seconds=timing[by[0]]['median']
 else:
  verdict='overlapping';by=sorted({min([other for other in scope if other!=row and mid in sets[other]],key=lambda x:(timing[x]['median'],x)) for mid in kills})
 file='internal/lower/check_pragmas_test.go' if row in scope[:3] else 'internal/lower/class_features_test.go'
 base=subprocess.check_output(['git','show','origin/main:'+file],text=True);lineno=base[:base.index('func '+row+'(')].count('\n')+1
 oracle='Self-written expected Adamic refusal text and type.';kind='self'
 if row=='TestCheckPragmasAreRefused':oracle='Self-written Refused type, source location, pragma name, removal instruction and fix fragments.'
 if row=='TestCheckPragmaNeighborsCompile':oracle='Self-written expectation of nil error only; never inspects the returned program. Empty Lower passes.'
 if row=='TestTsgoHonorsNoCheckInAdamicFiles':oracle='Self-written acceptance/rejection expectations for in-process tsgo through Load, plus Adamic Refused type. First loader rejection accepts any error.'
 if row=='TestClassFeaturesReadonlyChecker':oracle='tsgo TS2540 read-only diagnostic fragment, checked against pinned upstream definition this session; nil-error expectation for positive lowering subcase. Does not assert exact diagnostic code or positive IR.';kind=['external-authority','self']
 if row=='TestClassFeaturesPrivateChecker':oracle='tsgo TS18013 private identifier diagnostic fragment, checked against pinned upstream definition this session. Does not assert exact code.';kind='external-authority'
 if row=='TestClassFeaturesPrivateStorage':oracle='Self-written private-field count >= 2. Counts visibility metadata, not field identity or native privacy.'
 if row=='TestClassFeaturesAccessorRefusals':oracle='Self-written accessor override and getter may throw diagnostic fragments; no exact code/location.'
 if row=='TestClassFeaturesStaticDeclarationsExecute':oracle='Self-written Classes nonempty, first class Static flag, Main length >= 2. Does not execute; survives M15 removal of initializer calls.'
 if row=='TestClassFeaturesStaticSoundness':oracle='Self-written non-nil-error expectation only for ten unsafe sources; any unrelated lowering error also satisfies it.'
 if row=='TestClassFeaturesAccessorCaptureCycle':oracle='Self-written cycle diagnostic fragment; no exact location/code.'
 if row=='TestClassFeaturesNarrowedAccessor':oracle='Self-written narrowed accessor reread fragment. M09 still passes when refusal moves from column 207 to the earlier read at 167; witness logs prove wrong-read rejection.'
 if row in ['TestClassFeaturesStaticParentCycle','TestClassFeaturesStaticInterfaceCycle']:oracle='Self-written cycle diagnostic fragment; no exact location/code.'
 own=['P2'] if row=='TestClassFeaturesPrivateChecker' else ['P1']
 if row in ['TestTsgoHonorsNoCheckInAdamicFiles','TestClassFeaturesReadonlyChecker']:own=['P1','P2']
 probe_kills=[p for p in own if row in matrix[p]['failed_rows']]
 vacuous=not probe_kills
 subcases=['positive readonly/mutable-content lowering subcase passes P1; Lower entry is vacuous'] if row=='TestClassFeaturesReadonlyChecker' else []
 last=kills[-1] if kills else None; fail=matrix[last]['failing_lines'][row] if last else None
 log=matrix[last]['logs'][row] if last else matrix['P1']['logs'][row]
 pattern='^'+row+'$' if last and matrix[last]['bounded'] else '.'
 command=('ADAMIC_MUTANT='+last+' ADAMIC_BUILD_CACHE_DIR=/tmp/u028/cache/'+last+' timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run '+pattern+' > review/test-audit/internal-lower-check_pragmas/'+log+' 2>&1') if last else 'No production mutant killed this row. P1 isolated run passed; see '+log
 rows.append(dict(test=row,package='internal/lower',file=file+':'+str(lineno),seconds=timing[row]['median'],oracle=oracle,oracle_kind=kind,kills=kills,unique_kills=unique,last_proven_fail=(last+': '+fail) if last else None,verdict=verdict,subsumed_by=by,mutants_in_matrix=18,probe_kills=probe_kills,subsumer_seconds=seconds,vacuous=vacuous,vacuous_subcases=subcases,bounded=True,matrix_rows=scope,evidence=command+('; '+fail if fail else ''),subsumption_mutant_count=len(kills) if verdict=='subsumed' else None))
(r/'rows.json').write_text(json.dumps(rows,indent=2)+'\n');(r/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n');(r/'timings.json').write_text(json.dumps(timing,indent=2)+'\n')
print([(x['test'],x['verdict'],x['kills'],x['unique_kills'],x['vacuous']) for x in rows])

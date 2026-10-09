import pathlib,json,statistics,re
p=pathlib.Path('review/test-audit/internal-oracle-interface_cast'); scope=json.loads((p/'scope.json').read_text()); matrix=json.loads((p/'matrix.json').read_text()); timings=json.loads((p/'timings.json').read_text()); isolated=json.loads((p/'witness-isolated.json').read_text()); witnesses={x['row'] for x in isolated}; production=[x for x in scope if x not in witnesses]; medians={x:statistics.median(v) for x,v in timings.items()}; files={}
for file in pathlib.Path('internal/oracle').glob('*_test.go'):
 for row in re.findall(r'^func (Test\w+)\(',file.read_text(),re.M): files[row]=str(file)
def diagnostic(log,row):
 candidates=[]
 for line in (p/log).read_text().splitlines():
  try: event=json.loads(line)
  except: continue
  if event.get('Action')!='output' or event.get('Test','').split('/')[0]!=row: continue
  text=event.get('Output','').strip()
  if re.search(r'(mutant survived|caught by|want |got .*want|not caught|Node did not catch|malformed view read|stdout differs|stderr differs|exit codes differ|panic: runtime)',text) and not text.startswith('==='): candidates.append(text)
 return candidates[0] if candidates else 'See raw log'
for result in matrix:
 result['failure_evidence']={row:diagnostic(result['log'],row) for row in result['failed']}
(p/'matrix-with-evidence.json').write_text(json.dumps(matrix,indent=2))
prod=[r for r in matrix if r['id'].startswith('M')]; killsets={row:{r['id'] for r in prod if row in r['failed']} for row in production}; entries={x:['P01','P03','P04'] for x in production[:4]}; entries['TestJSONStringifyRefusals']=['P01','P02']; entries['TestJSONStringifyResultMayBeUndefined']=['P02']; entries['TestLibraryMapSetIteratorCopiesRefused']=['P01']; objects=[]
for row in scope:
 iswitness=row in witnesses; kills=sorted(killsets.get(row,set())); unique=[r['id'] for r in prod if r['failed']==[row]]; subsumer=None
 if iswitness: verdict='witness'
 elif unique: verdict='sacred'
 elif kills:
  candidates=[other for other in production if other!=row and set(kills)<=killsets[other]]; subsumer=min(candidates,key=lambda x:medians[x]) if candidates else None; verdict='subsumed' if subsumer else 'overlapping'
 else: verdict='untrue'
 if iswitness:
  observation=next(x for x in isolated if x['row']==row); failure='W01: '+diagnostic(observation['log'],row); evidence=observation['command']+' => '+diagnostic(observation['log'],row); probes=['P05']; vacuous=False
 else:
  chosen=next(r for r in reversed(prod) if row in r['failed']); failure=chosen['id']+': '+chosen['failure_evidence'][row]; evidence=chosen['command']+' => '+chosen['failure_evidence'][row]; probes=[r['id'] for r in matrix if r['id'] in entries[row] and row in r['failed']]
  if row=='TestJSONStringifyRefusals': probes=sorted(set(probes)|{'P02'})
  vacuous=not bool(probes)
 if iswitness: oracle='Node source observations and self-written expectation that disagreement detects the built-in mutant; W01 disabled the witnessed check'
 elif row=='TestJSONStringifyRefusals': oracle='Self-written Adamic refusal text; duplicate TS1117 copied from TypeScript diagnostic authority, checked against cohere/TypeScript/tsc/internal/diagnostics/diagnosticMessages.json this session'
 elif row=='TestJSONStringifyResultMayBeUndefined': oracle='Self-written expected undefined rejection from Adamic public declarations; diagnostic substring must contain undefined'
 elif row=='TestLibraryMapSetIteratorCopiesRefused': oracle='Node copy.next failure plus self-written NotYet inherited-next refusal text'
 elif row=='TestInterfaceCastImportedConstruction': oracle='Node source stdout, stderr and exit status compared to native and JavaScript backends'
 else: oracle='Node source observations plus self-written Adamic inserted-panic contracts, stdout and stderr; scalar row also witnesses omission detection' if row.endswith('ScalarTags') else 'Node source observations plus self-written Adamic inserted-panic contracts, stdout and stderr'
 kind=['self','external-authority'] if row=='TestJSONStringifyRefusals' else 'self' if row=='TestJSONStringifyResultMayBeUndefined' else 'external-run' if row=='TestInterfaceCastImportedConstruction' else ['external-run','self']
 obj={'test':row,'package':'internal/oracle','file':files[row],'seconds':medians[row],'oracle':oracle,'oracle_kind':kind,'kills':kills,'unique_kills':unique,'last_proven_fail':failure,'verdict':verdict,'subsumed_by':[subsumer] if subsumer else [],'mutants_in_matrix':0 if iswitness else 12,'probe_kills':probes,'subsumer_seconds':medians[subsumer] if subsumer else None,'vacuous':vacuous,'bounded':not iswitness,'evidence':evidence}
 if not iswitness: obj['matrix_rows']=production
 else: obj['witness_checks']=['W01','W02']; obj['probe_note']='P05 is the empty-answer W01 weakening, sharing the same diff and observations'
 if row=='TestInterfaceCastScalarTags': obj['witness_component']='W01 also fails the built-in scalar omission checks; production verdict uses only M-series kills'
 if row=='TestJSONStringifyRefusals': obj['entry_note']='Duplicate subcase calls Load; P02-duplicate.log probes that entry separately. Other cases call Lower.'
 objects.append(obj)
(p/'rows.json').write_text(json.dumps(objects,indent=2))
table='| ID | Starting-main file:line | Change | Failed rows |\n|---|---|---|---|\n'
for r in prod:
 where,change=(p/(r['id']+'.location')).read_text().strip().split(': ',1); table+='| '+r['id']+' | '+where+' | `'+change+'` | '+(', '.join(r['failed']) or 'survivor')+' |\n'
(p/'mutants.md').write_text(table)
summary='Starting main: '+(p/'start.txt').read_text().splitlines()[0]+'; all 15 names present, none moved or vanished.\n15 rows: seven production/mixed rows and eight pure witnesses; no extra family grouping.\nClean scoped baseline passed in 8.625s; whole package exceeded 90s.\nBounded verdicts: four sacred, three subsumed; eight witness verdicts.\nEvidence branch: test-audit/internal-oracle-interface_cast; nproc 5.\n'
text=summary+'\n```json\n'+json.dumps(objects,indent=2)+'\n```\n\n'+table+'\nSurvivors:\n\nM03: equivalent candidate. native.C changes ADAMIC_STRING("a b") to ADAMIC_STRING("a\\040b"); rebuilt programs both print "a b\\n".\n\nM11: unguarded by this bounded set. Load changes from TS2375 rejection to acceptance of `{p: undefined}` when ExactOptionalPropertyTypes is false. Outside-set protection remains unknown.\n\n'+(p/'limitations.md').read_text()
compiles=json.loads((p/'compiles.json').read_text()); replay=json.loads((p/'replay-checks.json').read_text()); text+='\nMeasured timings: core setup 0s; npm installation and initial Go/coverage compilation were not separately timed. Row timing runs total '+str(round(sum(sum(v) for v in timings.values()),3))+' binary seconds. Production/witness/probe matrix commands total '+str(round(sum(r['seconds'] for r in matrix),3))+' wall seconds plus '+str(round(sum(part['seconds'] for r in matrix for part in r.get('reruns',[])),3))+'s isolation reruns. Isolated standalone witness checks add '+str(round(sum(r['seconds'] for r in isolated),3))+'s. Standalone production vet checks took '+str(round(sum(x['seconds'] for x in compiles),3))+'s; other replay compilation checks took '+str(round(sum(x.get('seconds',0) for x in replay),3))+'s. Runtime rebuilding is included in these command wall times and was not separately instrumented.\n'
(p/'report.md').write_text(text); print([(o['test'],o['verdict'],o['subsumed_by'],o['last_proven_fail']) for o in objects])

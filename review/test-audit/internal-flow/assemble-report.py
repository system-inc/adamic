import json,re,statistics,gzip,shutil,subprocess
from pathlib import Path
r=Path('/workspace/adamic/review/test-audit/internal-flow')
rows=json.loads((r/'rows.json').read_text()); names=list(rows); mapping={m:n for n,ms in rows.items() for m in ms}
markers=['the graph reads','reads a variable nothing defines',' names ','defined in bb','not dominate','no edge','entry doesn','in the middle','call ended','no point at all','dead','outside its range','invalid ranges for','where no value','nothing was checked','debugger changed','setup was rebuilt','unit for ','no mutation was seen','ran a point','panic:']
def scan(p):
 result={'rows':{},'members':{},'errors':{},'panic':False,'timeout':False,'seconds':None,'command':'timeout 120 go test -json -count=1 -timeout 90s ./internal/flow/ -run .'}
 current=None
 for line in p.open():
  try:x=json.loads(line)
  except:continue
  test=x.get('Test');act=x.get('Action');out=x.get('Output','').strip();row=mapping.get(test.split('/')[0]) if test else None
  if test:current=row
  if act in ['pass','fail','skip'] and test:
   result['members'][test]=act
   if row:
    if act=='fail' or result['rows'].get(row)!='fail': result['rows'][row]=act
  if out and 'test timed out' not in out and (re.search(r'\.go:\d+:',out) or out.startswith('panic:')) and any(m in out for m in markers):
   target=row or current
   if target and target not in result['errors']:result['errors'][target]=out
  if 'panic:' in out:result['panic']=True
  if 'test timed out' in out:result['timeout']=True
  if act in ['pass','fail'] and not test and result['seconds'] is None:result['seconds']=x.get('Elapsed')
  match=re.search(r'^ok\s+github.com/system-inc/adamic/internal/flow\s+([0-9.]+)s$',out)
  if match:result['seconds']=float(match.group(1))
 for row in result['errors']:result['rows'][row]='fail'
 return result
matrix={}
for mid in ['M'+str(i) for i in range(1,10)]:
 full=scan(r/(mid+'.log'));effective=scan(r/(mid+'-bounded.log')) if (r/(mid+'-bounded.log')).exists() else full
 effective['bounded']=(r/(mid+'-bounded.log')).exists();effective['full_run']=full
 if effective is not full:
  for row,state in full['rows'].items():
   if state=='fail':
    effective['rows'][row]='fail'
    effective['errors'].setdefault(row,full['errors'].get(row,'observed FAIL action in full run'))
 matrix[mid]=effective
(r/'matrix.json').write_text(json.dumps(matrix,indent=2))
probes={}
for mid in ['PBuild','PConstruct','PLiveOut','PRanges']:
 full=scan(r/(mid+'.log'));effective=scan(r/(mid+'-bounded.log')) if (r/(mid+'-bounded.log')).exists() else full
 if full['panic'] and not full['timeout']:
  effective={'rows':{},'errors':{},'members':{},'bounded':True}
  for n,row in enumerate(names):
   p=r/(mid+'-row'+str(n)+'.log')
   if p.exists():
    x=scan(p);effective['rows'].update(x['rows']);effective['errors'].update(x['errors']);effective['members'].update(x['members'])
 probes[mid]=effective
(r/'probes.json').write_text(json.dumps(probes,indent=2))
seconds={};timing={}
for n,row in enumerate(names):
 values=[scan(r/('timing-row'+str(n)+'-'+str(i)+'.log'))['seconds'] for i in range(1,4)]
 timing[row]=values;seconds[row]=statistics.median(values) if all(x is not None for x in values) else None
(r/'timings.json').write_text(json.dumps(timing,indent=2))
objects=[]
for row,members in rows.items():
 kills=[mid for mid,x in matrix.items() if x['rows'].get(row)=='fail']
 unique=[mid for mid in kills if all(n in matrix[mid]['rows'] for n in names) and [k for k,v in matrix[mid]['rows'].items() if v=='fail']==[row]]
 oracle_kind='self';oracle='Self: hand-written count of one Evaluate instruction; only the count is checked.';own=['PBuild']
 if row=='TestFlowProgram family':
  oracle_kind=['external-run','self'];oracle='Node runs marked Adamic JavaScript and checks graph paths, observed mutations and future reads. SSA uses self-written reaching definitions, VerifySSA and fmt IR-read counts. Mutation ranges left unset are skipped, so absence is accepted. Coverage and remainder included per corpus-family rule.';own=['PBuild','PConstruct','PLiveOut','PRanges']
 elif 'SingleAssignment' in row:
  oracle='Self: VerifySSA, independent textbook reaching definitions over Build, and fmt IR-read count. No outside expected value checked.';own=['PConstruct']
 elif 'MutationRanges' in row:
  oracle_kind='external-run';oracle='Node observes object changes in Adamic JavaScript; set ranges must contain observed mutations. Unset ranges explicitly skip checks, so an empty table is accepted.';own=['PRanges']
 elif 'GraphPaths' in row:
  oracle_kind='external-run';oracle='Node runs marked Adamic JavaScript; trace must follow Build edges and finish at a return. It requires at least one numeric trace point.';own=['PBuild']
 elif 'Liveness' in row:
  oracle_kind='external-run';oracle='Node trace determines later reads; every such variable must be live. Only under-approximation is rejected; extra live variables are accepted.';own=['PLiveOut']
 elif 'SetupIsShared' in row:
  oracle='Self: repeated preparation must return pointer-identical lowered IR and trace setup. SSetupLower and SSetupTrace independently break these caches.';own=[]
 pk=[mid for mid in own if probes[mid]['rows'].get(row)=='fail'];probe_pass=[mid for mid in own if probes[mid]['rows'].get(row)=='pass']
 vacuous=(bool(probe_pass) if own else None)
 # Combined corpus row has several production entries; report separately rather than conceal the empty-ranges pass.
 if row=='TestFlowProgram family':vacuous=bool(probe_pass)
 subs=[other for other in names if other!=row and kills and all(mid in [k for k,x in matrix.items() if x['rows'].get(other)=='fail'] for mid in kills)]
 subs=sorted(subs,key=lambda x:seconds[x] if seconds[x] is not None else float('inf'))
 if 'SetupIsShared' in row:verdict='setup-check';subs=[];unique=[]
 elif unique:verdict='slow-worthy' if seconds[row] and seconds[row]>60 else 'sacred'
 elif subs:verdict='subsumed';subs=subs[:1]
 elif kills:verdict='overlapping';subs=[other for other in names if other!=row and any(matrix[mid]['rows'].get(other)=='fail' for mid in kills)]
 else:verdict='untrue';subs=[]
 last=next((mid+': '+matrix[mid]['errors'].get(row,'observed FAIL action') for mid in reversed(kills)),None)
 evidence='; '.join(mid+': timeout 120 go test -json -count=1 -timeout 90s ./internal/flow/ -run '+('bounded member regex (bounded-members.json)' if matrix[mid]['bounded'] else '.')+'; '+matrix[mid]['errors'].get(row,'FAIL action') for mid in kills)
 if 'SetupIsShared' in row:
  x=scan(r/'SSetupTrace.log');last='SSetupTrace: '+x['errors'].get(row,'FAIL action');evidence='go test -json -count=1 -timeout 90s ./internal/flow/ -run ^TestFlowCorpusSetupIsShared$; '+last
 obj={'test':row,'package':'internal/flow','file':('internal/flow/corpus_units_test.go; internal/flow/corpus_coverage_test.go' if row=='TestFlowProgram family' else 'internal/flow/corpus_units_test.go') if 'timsort' in row or row=='TestFlowProgram family' else ('internal/flow/debugger_test.go' if 'Debugger' in row else 'internal/flow/corpus_coverage_test.go'),'members':members,'seconds':seconds[row],'oracle':oracle,'oracle_kind':oracle_kind,'kills':kills,'unique_kills':unique,'last_proven_fail':last,'verdict':verdict,'subsumed_by':subs,'mutants_in_matrix':9,'probe_kills':pk,'subsumer_seconds':seconds[subs[0]] if verdict=='subsumed' else None,'vacuous':vacuous,'bounded':any(matrix[m]['bounded'] for m in matrix),'matrix_rows':names,'evidence':evidence,'probe_results':{mid:probes[mid]['rows'].get(row,'unknown') for mid in own},'subsumption_mutants':len(kills) if verdict=='subsumed' else None}
 if row=='TestFlowProgram family':obj['vacuous_subcases']=['PRanges: mutation-range checker accepts an empty table; other checkers still reject failures. PBuild: coverage/remainder construction members are separate from the Build entry.']
 if 'SetupIsShared' in row:obj['setup_kills']=['SSetupLower','SSetupTrace']
 objects.append(obj)
(r/'rows-results.json').write_text(json.dumps(objects,indent=2))
# Compact, human-readable failure evidence retains commands and origin mapping in sibling artifacts.
with (r/'failing-lines.txt').open('w') as f:
 for mid,x in matrix.items():
  for row,line in x['errors'].items():f.write(mid+' | '+row+' | '+line+'\n')
 for mid,x in probes.items():
  for row,line in x.get('errors',{}).items():f.write(mid+' | '+row+' | '+line+'\n')
# Raw observations kept compressed for replay, avoiding very large Git blobs.
for p in r.glob('*.log'):
 with p.open('rb') as src,gzip.open(str(p)+'.gz','wb',compresslevel=6) as dst:shutil.copyfileobj(src,dst)
 p.unlink()
print(json.dumps([{k:v for k,v in x.items() if k in ['test','seconds','kills','unique_kills','verdict','vacuous','subsumed_by']} for x in objects],indent=2))

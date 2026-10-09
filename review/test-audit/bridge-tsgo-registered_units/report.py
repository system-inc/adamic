import json,re,statistics,subprocess
from pathlib import Path
P=Path('review/test-audit/bridge-tsgo-registered_units');scope=json.load(open(P/'scope.json'));rows=scope['rows'];plan=json.load(open(P/'plan.json'));wplan=json.load(open(P/'witness-plan.json'));witnesses={r:w['id'] for w in wplan for r in w['rows']};runs=[]
for f in P.glob('*.json'):
 try:d=json.load(open(f))
 except:continue
 if isinstance(d,dict) and 'events' in d and 'label' in d:runs.append(d)
runmap={r['label']:r for r in runs};matrix=[]
def proof(d,row):
 out=''.join(e.get('Output','') for e in d['events'] if e.get('Test','').split('/')[0]==row)
 line=next((l.strip() for l in out.splitlines() if re.search(r'\.go:\d+:',l) and not ('build bridge-test' in l or 'queries:' in l or 'tsgo:' in l)),None)
 if not line:line=next((l.strip() for l in out.splitlines() if 'panic:' in l),'no failing line observed')
 return {'command':d['command'],'environment':{'ADAMIC_MUTANT':d['selector'],'ADAMIC_TSGO_CORPUS':'/tmp/u004/typescript' if d['corpus'] else '', 'ADAMIC_BUILD_CACHE_DIR':d.get('cache')},'line':line,'output':out,'run':d['label']}
for m in plan:
 ds=[d for d in runs if d['selector']==m['id'] and not '-timing-' in d['label']]; status={};evidence={}
 for d in ds:
  for row,s in d['status'].items():
   status[row]=s
   if s=='fail':evidence[row]=proof(d,row)
 # Package-level panic can omit a row fail event. Do not promote it to a kill.
 matrix.append(dict(m,status=status,failed_rows=[r for r in rows if r not in witnesses and status.get(r)=='fail'],other_failures=[r for r,s in status.items() if s=='fail' and r not in rows],matrix_rows=sorted(status),evidence=evidence))
(P/'matrix.json').write_text(json.dumps(matrix,indent=2))
medians={r:statistics.median(runmap[r+'-timing-'+str(n)]['binary_seconds'] for n in [1,2,3]) for r in rows};audit=[]
for row in rows:
 caught=[m for m in matrix if not m['probe'] and row in m['failed_rows']]; kills=[m['id'] for m in caught];probes=[m['id'] for m in matrix if m['probe'] and m['status'].get(row)=='fail'];unique=[m['id'] for m in caught if m['failed_rows']==[row]];matrixrows=sorted({r for m in caught for r in m['matrix_rows']})
 oracle='direct Go checker, all framed answer bytes';kind='external-run';subs=[];vacuous=None;vacuous_entries=[]
 if row=='TestBridgeABI':oracle='self-written C ABI assertions: nonzero kind/type length, status codes, cleared buffers and handle lifetime; no exact type text';kind='self';entries=['P_CREATE','P_ABI','P_ABI_RELEASE','P_FREE','P_RESULT_FREE']
 elif 'Unlinked' in row:oracle='nonzero exit and self-written unlinked-call diagnostic substring';kind='self';entries=['P_LOWER']
 elif row=='TestBridgeRegion':oracle='Go-derived combined UTF-16 symbol/type length plus self-written regions=1; ignores kind and exact text';kind=['external-run','self'];entries=['P_PROGRAM','P_QUERY','P_RELEASE']
 elif 'Timing' in row:oracle='direct Go checker, all framed answer bytes; timings logged without a performance threshold';entries=['P_PROGRAM','P_QUERY','P_RELEASE']
 elif row in witnesses:
  entries=[]
  if 'WrongPosition' in row:oracle='direct Go checker; planted wrong-position answer must disagree'
  elif row=='TestBridgeLinkage':oracle='self-written inner unlinked-call refusal assertion';kind='self'
  elif row=='TestBridgeStaleHandle':oracle='self-written C stale-handle status assertion';kind='self'
  else:oracle='actual AddressSanitizer/LeakSanitizer execution and expected diagnostic substring';kind=['external-run','self']
 else:entries=['P_PROGRAM','P_QUERY','P_RELEASE']
 # Overall answer vacuity uses the queried entry; additional entries remain explicit.
 for entry in entries:
  m=next(m for m in matrix if m['id']==entry)
  if m['status'].get(row)=='pass':vacuous_entries.append(entry)
 primary='P_ABI' if row=='TestBridgeABI' else 'P_LOWER' if 'Unlinked' in row else 'P_QUERY'
 if entries:
  state=next(m for m in matrix if m['id']==primary)['status'].get(row);vacuous=state=='pass' if state in ['pass','fail'] else None
 if row in witnesses:
  id=witnesses[row];d=runmap.get(id+'-'+row); verdict='witness' if d and d['status'].get(row)=='fail' else 'untrue' if d and d['status'].get(row)=='pass' else 'cannot-judge';p=proof(d,row) if d and d['status'].get(row)=='fail' else None;kills=[];unique=[];matrixrows=[row];last=id+': '+p['line'] if p else 'none';evidence=p if p else {'reason':'weakened-check run unavailable or did not fail'}
 else:
  if unique:verdict='slow-worthy' if medians[row]>60 else 'sacred'
  elif kills:
   candidates=[r for r in rows if r!=row and r not in witnesses and all(next(m for m in matrix if m['id']==id)['status'].get(r)=='fail' for id in kills)]
   if candidates:verdict='subsumed';subs=[min(candidates,key=medians.get)]
   else:verdict='overlapping';subs=sorted({r for m in caught for r in m['failed_rows'] if r!=row})
  else:verdict='untrue'
  p=caught[-1]['evidence'][row] if caught else None;last=caught[-1]['id']+': '+p['line'] if p else 'none';evidence=p if p else {'reason':'no production mutant killed'}
 audit.append({'test':row,'package':'bridge/tsgo','file':'bridge/tsgo/registered_units_test.go','seconds':medians[row],'oracle':oracle,'oracle_kind':kind,'kills':kills,'unique_kills':unique,'last_proven_fail':last,'verdict':verdict,'subsumed_by':subs,'mutants_in_matrix':len([m for m in matrix if not m['probe'] and row in m['status']]) if row not in witnesses else 0,'probe_kills':probes,'subsumer_seconds':medians[subs[0]] if verdict=='subsumed' else None,'vacuous':vacuous,'vacuous_entries':vacuous_entries,'bounded':True,'matrix_rows':matrixrows,'evidence':str(evidence.get('environment',{}))+' '+evidence.get('command','')+'; '+evidence.get('line',evidence.get('reason','')),'evidence_details':evidence,'subsumption_mutants':len(kills) if verdict=='subsumed' else None})
(P/'audit.json').write_text(json.dumps(audit,indent=2))
# Census times include dependency builds and are not additive.
builds=[]
for d in runs:
 for e in d['events']:
  match=re.search(r'build (\S+) (\S+) (miss|off) ([\d.]+)',e.get('Output',''))
  if match:builds.append({'run':d['label'],'product':match[1],'key':match[2],'mode':match[3],'inclusive_seconds':float(match[4])})
(P/'builds.json').write_text(json.dumps(builds,indent=2));env=json.load(open(P/'environment.json'));env['timing_runs_wall_seconds']=sum(d['wall_seconds'] for d in runs if '-timing-' in d['label']);env['matrix_runs_wall_seconds_sum']=sum(d['wall_seconds'] for d in runs if d['selector']);env['witness_runs_wall_seconds_sum']=sum(d['wall_seconds'] for d in runs if d['label'].startswith('W_'));env['active_baseline_runs']=108;env['remaining_skips']=[d['label'] for d in runs if 'skip' in d['status'].values()];(P/'environment.json').write_text(json.dumps(env,indent=2))
print('VERDICTS',[(r['test'],r['verdict'],r['kills'],r['probe_kills'],r['vacuous_entries']) for r in audit]);print('ENV',env)

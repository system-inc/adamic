import json,re,statistics,subprocess,time,os
from pathlib import Path
P=Path('review/test-audit/internal-lower-enum_flags');S=Path('/tmp/u031');names=json.loads((S/'rows.json').read_text());plan=json.loads((P/'plan.json').read_text())
def events(log):
 es=[]
 for l in log.read_text().splitlines():
  try:es.append(json.loads(l))
  except ValueError:pass
 return es
def evidence(d,row):
 output=''.join(e.get('Output','') for e in d['events'] if e.get('Test','').split('/')[0]==row)
 line=next((l.strip() for l in output.splitlines() if re.search(r'_test.go:\d+:',l) and 'function body expansions' not in l),None)
 if not line:line=next((l.strip() for l in output.splitlines() if 'panic:' in l),'No assertion line; see complete events')
 return {'command':d['command'],'line':line,'output':output}
def median(row):
 times=[]
 for i in [1,2,3]:
  log=S/(row+'-'+str(i)+'.log');es=events(log);sec=next(e['Elapsed'] for e in reversed(es) if not e.get('Test') and e['Action']=='pass');times.append(sec)
  (P/(row+'-timing-'+str(i)+'.json')).write_text(json.dumps(dict(json.loads((S/(row+'-'+str(i)+'.time.json')).read_text()),binary_seconds=sec,events=es),indent=2))
 return statistics.median(times)
mat=[]
for m in plan:
 d=json.loads((P/(m['id']+'.json')).read_text());statuses={k:v for k,v in d['status'].items() if '/' not in k};proof={r:evidence(d,r) for r,v in statuses.items() if v=='fail'};subcases=dict(d['status']);bounded=d['aborted']
 if bounded:
  for row in names:
   rerun=json.loads((P/(m['id']+'-'+row+'.json')).read_text());statuses[row]=rerun['status'].get(row,'fail' if rerun['aborted'] else 'unknown');subcases.update(rerun['status']);
   if statuses[row]=='fail':proof[row]=evidence(rerun,row)
 mat.append(dict(m,status=statuses,failed=[k for k,v in statuses.items() if v=='fail'],bounded=bounded,evidence=proof,subcases=subcases))
(P/'matrix.json').write_text(json.dumps(mat,indent=2));timings={r:median(r) for r in names};regular=[m for m in mat if not m['probe']];oracles={
'TestFlagEnumsOpen':'self-written acceptance assertions and Refused type for two literal-promise subcases; reason table is not asserted',
'TestFlagEnumsDomain':'self-written nil-error assertion; no IR or runtime answer checked',
'TestEnumNeverDefault':'self-written nil-error assertion; no IR or runtime answer checked',
'TestFlagEnumLiteralSpellings':'self-written nil-error assertion; literal values not checked',
'TestFlagEnumMemberAliases':'self-written nil-error assertion; alias values not checked',
'TestEnumNameEnumeration':'self-written nil-error assertion; enumerated names not checked',
'TestFlagEnumInlineIteration':'self-written nil-error assertion; iterable values not checked',
'TestFlagEnumAliasBoundaries':'self-written nil-error assertions and Refused type for Mutable<T> and member-slot cases',
'TestEnumInitializationReach':'self-written acceptance, NotYet text and source-line assertions',
'TestEnumInitializationGraphMemo':'self-written reach count and exact body expansion counts',
'TestEnumNamespaceSharedCycle':'self-written reach identities, count and exact body expansion count',
'TestNumericEnumsAreOpen':'self-written nil-error assertions; no IR or runtime answer checked',
'TestStringEnumsStayClosed':'self-written Refused type and diagnostic substring alternatives',
'TestNumericEnumNeverProof':'self-written IR function name enum_never and Panic presence',
'TestNumericEnumLiteralPromises':'self-written Refused type only; unrelated Refused can satisfy it'
};audit=[]
# An aborted package run cannot prove uniqueness. Only completed matrix columns contribute unique kills.
for row in names:
 kills=[m for m in regular if m['status'].get(row)=='fail'];unique=[m['id'] for m in kills if not m['bounded'] and m['failed']==[row]];bounded=any(m['bounded'] for m in regular);subs=[]
 if unique:verdict='slow-worthy' if timings[row]>60 else 'sacred'
 elif not kills:verdict='untrue'
 else:
  candidates=set.intersection(*(set(m['failed'])-{row} for m in kills));candidates-=set(r for r in candidates if any(m['status'].get(r)=='unknown' for m in kills))
  if candidates:
   ordered=sorted(candidates,key=lambda r:(r not in timings,timings.get(r,999),r));subs=[ordered[0]];verdict='subsumed'
   if subs[0] not in timings:
    name=subs[0]
    for i in [1,2,3]:
     cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^'+name+'$'];start=time.monotonic()
     with open(S/(name+'-'+str(i)+'.log'),'w') as f:r=subprocess.run(cmd,env=dict(os.environ,ADAMIC_MUTANT=''),stdout=f,stderr=subprocess.STDOUT)
     (S/(name+'-'+str(i)+'.time.json')).write_text(json.dumps({'command':' '.join(cmd),'exit':r.returncode,'wall_seconds':time.monotonic()-start}));assert r.returncode==0
    timings[name]=median(name)
  else:verdict='overlapping';subs=sorted(set(r for m in kills for r in m['failed'] if r!=row))
 entry='P_REACH' if row in ['TestEnumInitializationGraphMemo','TestEnumNamespaceSharedCycle'] else 'P_LOWER';probe=next(m for m in mat if m['id']==entry);probe_status=probe['status'].get(row,'unknown');passed=[k for k,v in probe['subcases'].items() if k.startswith(row+'/') and v=='pass'];last=kills[-1] if kills else None;
 if row=='TestFlagEnumAliasBoundaries' and probe_status=='fail': passed=['source[0]: cross-enum alias','source[1]: inline iterable with 99','source[2]: number-array iterable','source[3]: mutable for-of binding']
 proof=last['evidence'][row] if last else None
 ep=next(m for m in kills if m['id']==unique[-1])['evidence'][row] if unique else proof
 audit.append(dict(test=row,package='internal/lower',file=('internal/lower/enum_initialization_reach_test.go' if row.startswith('TestEnumInitialization') or row=='TestEnumNamespaceSharedCycle' else 'internal/lower/enums_open_test.go' if row.startswith('TestNumeric') or row=='TestStringEnumsStayClosed' else 'internal/lower/enum_flags_test.go'),seconds=timings[row],oracle=oracles[row],oracle_kind='self',kills=[m['id'] for m in kills],unique_kills=unique,last_proven_fail=last['id']+': '+proof['line'] if last else None,verdict=verdict,subsumed_by=subs,mutants_in_matrix=len(regular),probe_kills=[m['id'] for m in mat if m['probe'] and m['status'].get(row)=='fail'],subsumer_seconds=timings[subs[0]] if verdict=='subsumed' else None,vacuous=probe_status=='pass' if probe_status in ['pass','fail'] else None,vacuous_subcases=passed if probe_status=='fail' else [],bounded=bounded,matrix_rows=names if bounded else [],evidence=ep['command']+'; '+ep['line'] if ep else 'No production mutant failure observed; see matrix.json'))
(P/'audit.json').write_text(json.dumps(audit,indent=2));(P/'baseline.json').write_text(json.dumps(events(S/'baseline.log'),indent=2));(P/'slice.json').write_text(json.dumps(events(S/'slice.log'),indent=2));print(json.dumps([{k:r[k] for k in ['test','verdict','kills','unique_kills','subsumed_by','vacuous']} for r in audit],indent=2))

import json,pathlib,statistics,re
p=pathlib.Path('review/test-audit/internal-native-decode_ascii'); data=json.loads((p/'matrix.json').read_text()); timings=json.loads((p/'timings.json').read_text()); members=json.loads((p/'members.json').read_text()); rows=list(timings); family=rows[0]
def group(test): return family if test.startswith('TestDecodeASCII') else test
def diagnostic(result,row):
 logs=[p/(result['id']+'.log')]
 if result.get('reruns'): logs=[p/(x['id']+'.log') for x in result['reruns']]
 for log in logs:
  for line in log.read_text().splitlines():
   try: event=json.loads(line)
   except: continue
   if group(event.get('Test',''))!=row or event.get('Action')!='output': continue
   out=event.get('Output','').strip()
   if re.search(r'(mismatch|missing|not direct|dispatch|assigned receiver|lost its exact|native \"|answers differ|panic:)',out) and not out.startswith('==='): return out
 return 'See raw log for failure diagnostic'
for result in data: result['kill_rows']=sorted(set(group(test) for test in result['failed']))
for result in data: result['failure_evidence']={row:diagnostic(result,row) for row in result['kill_rows']}
(p/'grouped-matrix.json').write_text(json.dumps(data,indent=2))
objects=[]
oracles=['Node Buffer UTF-8 output, independent decoder snapshot, self-written metadata, indexing expectations and completed-partition counts','Self-written C fragments and call-case counts; expected adapter also uses production exactReceiverMethod, weakening independence','Self-written zero class identity for an assigned synthetic receiver','Node toExponential/toPrecision, all answers compared byte for byte','Node stdout and RangeError text; self-written exit 70 and panic expectations from docs/0.1.md; stderr checked by prefix']
for index,row in enumerate(rows):
 token='DecodeASCII' if index==0 else 'DevirtualizedCalls' if index in [1,2] else 'ToExponential'
 relevant=[r for r in data if r['id'].startswith('M') and token in r['pattern']]; killed=[r for r in relevant if row in r['kill_rows']]; unique=[r['id'] for r in killed if len(r['kill_rows'])==1]
 own={'P01'} if index==0 else {'P02','P03','P04'} if index==1 else {'P02'} if index==2 else {'P05','P06'}; probes=[r for r in data if r['id'] in own]
 verdict='slow-worthy' if index==0 and unique else 'sacred' if unique else 'subsumed' if index==2 and killed else 'untrue'
 chosen=killed[-1] if killed else None; line=diagnostic(chosen,row) if chosen else None
 obj={'test':row,'package':'internal/native','file':['internal/native/decode_ascii_test.go','internal/native/devirtualize_test.go','internal/native/devirtualize_test.go','internal/native/dtoa_test.go','internal/native/dtoa_test.go'][index],'seconds':statistics.median(timings[row]),'oracle':oracles[index],'oracle_kind':['external-run','self'] if index in [0,4] else 'external-run' if index==3 else 'self','kills':[r['id'] for r in killed],'unique_kills':unique,'last_proven_fail':chosen['id']+': '+line if chosen else None,'verdict':verdict,'subsumed_by':[rows[1]] if verdict=='subsumed' else [],'mutants_in_matrix':len(relevant),'probe_kills':[r['id'] for r in probes if row in r['kill_rows']],'subsumer_seconds':statistics.median(timings[rows[1]]) if verdict=='subsumed' else None,'vacuous':all(row not in r['kill_rows'] for r in probes) if probes else None,'bounded':True,'matrix_rows':[family] if index==0 else rows[1:3] if index in [1,2] else rows[3:],'evidence':chosen['command']+' => '+line if chosen else None}
 if index==0: obj['members']=members[family]
 objects.append(obj)
(p/'rows.json').write_text(json.dumps(objects,indent=2))
table='| ID | origin/main file:line | Change | Failed grouped rows |\n|---|---|---|---|\n'
for r in data:
 if not r['id'].startswith('M'): continue
 where,change=(p/(r['id']+'.location')).read_text().strip().split(': ',1); table+='| '+r['id']+' | '+where+' | `'+change+'` | '+', '.join(r['kill_rows'])+' |\n'
(p/'mutants.md').write_text(table)
summary='Starting commit: '+(p/'start.txt').read_text().splitlines()[0]+'; all 94 names present.\n94 functions group into five rows; none moved or vanished.\nClean scoped baseline: 76.942s; whole package exceeded 90s.\n12 mutants and six probes; bounded uniqueness only.\nEvidence branch: test-audit/internal-native-decode_ascii; nproc 5.\n'
text=summary+'\n```json\n'+json.dumps(objects,indent=2)+'\n```\n\n'+table+'\nSurvivors: '+(', '.join(r['id'] for r in data if r['id'].startswith('M') and not r['kill_rows']) or 'none in the bounded matrix')+'.\n\n'+(p/'limitations.md').read_text()
text+='\nMatrix and probe wall total: '+str(round(sum(r['seconds'] for r in data),3))+'s. Timing runs: '+str(round(sum(sum(v) for v in timings.values()),3))+' binary seconds. Individual compile times and inactive build phases are in driver.log and inactive logs.\n'
(p/'report.md').write_text(text)
print(json.dumps([{k:o[k] for k in ['test','seconds','kills','unique_kills','verdict','vacuous']} for o in objects],indent=2))

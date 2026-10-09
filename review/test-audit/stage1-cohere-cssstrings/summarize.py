import json,re,statistics,subprocess
from pathlib import Path
E=Path(__file__).resolve().parent
runs=[json.loads(x) for x in (E/'runs.jsonl').read_text().splitlines()]
records={r['label']:r for r in runs}
def events(label):
 result=[]
 for x in (E/(label+'.log')).read_text().splitlines():
  try:result.append(json.loads(x))
  except:pass
 return result
def row(t):
 if t=='TestMultiPushGap' or t=='TestCSSStringsPlantedDisagreement':return t
 return 'TestCSSStrings'
matrix=[]
mutants=json.loads((E/'mutants.json').read_text())
for m in mutants:
 if m['id'] not in records:continue
 r=records[m['id']];m['failed_tests']=r['failed_tests'];m['failed_rows']=sorted(set(map(row,r['failed_tests'])));m['wall_seconds']=r['wall_seconds'];m['cooked']=r['cooked'];matrix.append(m)
(E/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n')
rows=[]
for name,timing,file,oracle,kind in [('TestCSSStrings','family','port_test.go; shards_test.go; top_level_shards_test.go','Go cohere CSS adjustStrings through bridge; Node and JS/native port outputs compared to Go. Census and built-in witnesses use self checks.','external-run'),('TestMultiPushGap','gap','gaps_test.go','Self-authored NotYet.What string; Node executes fixture and must print self-authored ab newline. Compiler refusal expectation has no outside authority.','self'),('TestCSSStringsPlantedDisagreement','witness','shards_test.go','Self-authored synthetic case IDs and exactly-one-failure assertion on checker child.','self')]:
 kills=[m['id'] for m in matrix if name in m['failed_rows'] and m['id'].startswith('M')]
 unique=[m['id'] for m in matrix if m['failed_rows']==[name] and m['id'].startswith('M')]
 last='W1' if timing=='witness' else kills[-1] if kills else None
 proof=''
 if last:
  for e in events(last):
   if e.get('Test') and row(e['Test'])==name and e.get('Output') and re.search(r'\w+_test.go:\d+:',e['Output']) and any(x in e['Output'] for x in ['disagreement','gap changed','probe exit','mutant survived','build cssstrings port']):proof=e['Output'].strip();break
  if not proof:proof=next((e['Output'].strip() for e in events(last) if e.get('Test') and row(e['Test'])==name and e.get('Output','').startswith('--- FAIL:')),'no diagnostic')
 proof=re.sub(r'strings.ts:(\d+):',lambda m:'strings.ts:'+str(int(m[1])-3 if int(m[1])<=53 else int(m[1])-4)+':',proof)
 proof=re.sub(r'shards_test.go:(\d+):',lambda m:'shards_test.go:'+str(int(m[1])-3 if int(m[1])>374 else int(m[1]))+':',proof)
 probe='E1' if timing=='family' else 'E2-gap' if timing=='gap' else 'W1'
 verdict='witness' if timing=='witness' and name in records.get('W1',{}).get('failed_tests',[]) else 'sacred' if unique else 'cannot-judge'
 rows.append(dict(test=name,package='stage1/cohere/cssstrings',file=file,seconds=statistics.median(records['timing-'+timing+'-'+str(i)]['binary_seconds'] for i in [1,2,3]),oracle=oracle,oracle_kind=kind,kills=kills if timing!='witness' else ['W1'],unique_kills=unique,last_proven_fail=(str(last)+': '+proof) if last else None,verdict=verdict,subsumed_by=[],mutants_in_matrix=len([m for m in matrix if m['id'].startswith('M')]),vacuous=records.get(probe,{}).get('exit')==0,evidence=('selector file /tmp/adamic-u083-mutant='+str(last if last in ['M1','M2','M3','M4','M5','M6','E1'] else '')+'; env '+str(records[last]['environment'])+'; '+' '.join(records[last]['command'])+' > '+str(last)+'.log 2>&1; '+proof) if last else 'No completed kill'))
(E/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
build=[]
for r in runs:
 for e in events(r['label']):
  out=e.get('Output','')
  if 'build cssstrings_' in out and ('_lowering_and_backends' in out or 'sanitized' in out or 'native-fast' in out):build.append({'run':r['label'],'output':out.strip()})
(E/'build-times.json').write_text(json.dumps(build,indent=2)+'\n')
print(json.dumps({'rows':rows,'matrix':[{k:m[k] for k in ['id','file','line','failed_rows','wall_seconds','cooked']} for m in matrix],'recorded_run_wall':round(sum(r['wall_seconds'] for r in runs),3)},indent=2))

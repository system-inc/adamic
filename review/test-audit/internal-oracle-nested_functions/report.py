from pathlib import Path
import json,statistics,hashlib
p=Path(__file__).resolve().parent
rows=json.loads((p/'rows.json').read_text());plan=json.loads((p/'replay-plan.json').read_text());scope=json.loads((p/'matrix-scope.json').read_text())
results={m['id']:json.loads((p/(m['id']+'.json')).read_text()) for m in plan}
normalized={}
for id,r in results.items():
 fails=set(t.split('/')[0] for t in r['failed'])
 if id in ['M08','E00']:
  for row in rows:
   q=json.loads((p/(id+'-'+row+'.json')).read_text());fails.update(t.split('/')[0] for t in q['failed'])
 normalized[id]=sorted(fails)
(p/'matrix.json').write_text(json.dumps({'scope':scope,'columns':normalized,'incomplete_columns':['M08','E00'],'note':'Isolated unit-row reruns complete those six observations only. No package-unique kills proven.'},indent=2)+'\n')
witness={rows[0]:'W01',rows[3]:'W02'}
oracles={rows[0]:('Node for unchanged output; LeakSanitizer for the guarded leak check','external-run'),rows[1]:('Self-authored Adamic cycle-capable refusal requirement; Node runs only if accepted','self'),rows[2]:('TypeScript TS2630; independently checked with TypeScript 6.0.3 in tsc-authority.log','external-authority'),rows[3]:('Node source execution compared by disagreement','external-run'),rows[4]:('Node conversion result 4464, plus self-authored exact NotYet label','external-run'),rows[5]:('Node initializer result 9, plus self-authored exact NotYet label','external-run')}
out=[]
for row in rows:
 times=[json.loads((p/(row+'-'+str(n)+'.json')).read_text())['binary_seconds'] for n in [1,2,3]]
 kills=[m['id'] for m in plan if m['kind']=='production' and row in normalized[m['id']]]
 if row in witness:kills=[witness[row]]
 last='M06' if row==rows[1] else kills[-1]
 r=results[last]; source=row
 if last=='M08':r=json.loads((p/(last+'-'+row+'.json')).read_text())
 diagnostic=next((x.strip() for x in r['outputs'].get(row,[]) if '.go:' in x and '=== ' not in x),next((x.strip() for x in r['outputs'].get(row,[]) if 'panic:' in x),'no diagnostic'))
 verdict='witness' if row in witness else ('overlapping' if row==rows[5] else 'cannot-judge')
 subsumed=[rows[4],'TestNativeAgreesWithNode (selected fixture subtests)'] if verdict=='overlapping' else []
 empty=json.loads((p/('E00-'+row+'.json')).read_text())
 out.append(dict(test=row,package='internal/oracle',file='internal/oracle/'+('nested_functions_test.go' if row in rows[:4] else 'new_expression_test.go'),seconds=statistics.median(times),oracle=oracles[row][0],oracle_kind=oracles[row][1],kills=kills,unique_kills=[],last_proven_fail=last+': '+diagnostic,verdict=verdict,subsumed_by=subsumed,mutants_in_matrix=2 if row in witness else 16,vacuous=row in empty['passed'],evidence='ADAMIC_GATE_UNCACHED=1 python3 review/test-audit/internal-oracle-nested_functions/matrix.py; '+last+'.log: '+diagnostic))
(p/'results.json').write_text(json.dumps(out,indent=2)+'\n')
header='| ID | origin/main file:line | Change | Failed rows in bounded matrix |\n|---|---|---|---|\n'
lines=[]
for m in plan:
 desc=m['new'].replace('\n',' ').strip();old=m['old'].replace('\n',' ').strip()
 lines.append('| '+m['id']+' | '+m['file']+':'+str(m['line'])+' | `'+old.replace('|','\\|')+'` to `'+desc.replace('|','\\|')+'` | '+(', '.join(normalized[m['id']]) or '[]')+(' (incomplete outside six isolated unit rows)' if m['id'] in ['M08','E00'] else '')+' |')
(p/'mutants.md').write_text(header+'\n'.join(lines)+'\n')
print(json.dumps(out,indent=2));print(header+'\n'.join(lines))

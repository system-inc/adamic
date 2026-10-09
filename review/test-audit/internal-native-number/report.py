from pathlib import Path
import json,statistics
p=Path(__file__).resolve().parent;rows=json.loads((p/'rows.json').read_text());scope=json.loads((p/'scope.json').read_text());plan=json.loads((p/'replay-plan.json').read_text());results={m['id']:json.loads((p/(m['id']+'.json')).read_text()) for m in plan};normalized={id:sorted(set(t.split('/')[0] for t in r['failed'])) for id,r in results.items()}
(p/'matrix.json').write_text(json.dumps(dict(bounded=True,rows=scope['rows'],columns=normalized,unknown='Every omitted package row; no package or repo uniqueness proven'),indent=2))
medians={r:statistics.median(json.loads((p/(r+'-'+str(n)+'.json')).read_text())['seconds'] for n in [1,2,3]) for r in scope['rows']};(p/'medians.json').write_text(json.dumps(medians,indent=2))
verdict=['subsumed','sacred','overlapping','sacred'];entries=[['E01'],['E02','E03'],['E01'],['E04']];out=[]
for i,row in enumerate(rows):
 kills=[m['id'] for m in plan if m['kind']=='production' and row in normalized[m['id']]];unique=[id for id in kills if normalized[id]==[row]];last=kills[-1];r=results[last];diagnostic=next(x.strip() for x in r['outputs'][row] if '.go:' in x)
 obj=dict(test=row,package='internal/native',file=['internal/native/number_test.go','internal/native/parse_test.go','internal/native/power_of_two_test.go','internal/native/products_reuse_test.go'][i],seconds=medians[row],oracle='Node execution' if i<3 else 'Self-authored directory identity and build-count expectations',oracle_kind='external-run' if i<3 else 'self',kills=kills,unique_kills=unique,last_proven_fail=last+': '+diagnostic,verdict=verdict[i],subsumed_by=['TestMathAndToFixedMatchJavaScript'] if i in [0,2] else [],mutants_in_matrix=12,vacuous=any(row in results[id]['passed'] for id in entries[i]),bounded=True,matrix_rows=scope['rows'],evidence='python3 review/test-audit/internal-native-number/matrix.py; '+last+'.log: '+diagnostic)
 if i==2:obj['verdict_note']='Single common catcher is 86.9% slower, so it does not qualify for subsumption. The brief lacks this case; overlapping is a provisional retain label, not a claim that no common catcher exists.'
 out.append(obj)
(p/'results.json').write_text(json.dumps(out,indent=2)+'\n')
(p/'mutants.md').write_text('| ID | origin/main file:line | Change | Failed rows |\n|---|---|---|---|\n'+'\n'.join('| '+m['id']+' | '+m['file']+':'+str(m['line'])+' | '+m['menu']+': `'+m['old'].replace('\n',' ').replace('|','\\|')+'` becomes `'+m['new'].replace('\n',' ').replace('|','\\|')+'` | '+(', '.join(normalized[m['id']]) or '[]')+' |' for m in plan)+'\n')
runs=[]
for f in p.glob('*.json'):
 try:r=json.loads(f.read_text())
 except:continue
 if isinstance(r,dict) and 'wall' in r:runs.append(r)
validations=json.loads((p/'validations.json').read_text());timing=dict(nproc=5,setup_seconds=0,setup='Skipped; warm env.sh worked',npm_ci=json.loads((p/'npm-ci.json').read_text()),test_commands=len(runs),test_wall_seconds=sum(r['wall'] for r in runs),test_binary_seconds=sum(r['seconds'] or 0 for r in runs),standalone_validation_seconds=sum(r['seconds'] for r in validations),standalone_validated=16,control_wall=results.get('control',json.loads((p/'control.json').read_text()))['wall'],control_binary=json.loads((p/'control.json').read_text())['seconds'],note='Runtime archive building is inside binary duration. Control wall minus binary is startup/Go-build overhead, not separately measured compiler time. Probe builds not separately timed.')
(p/'timing.json').write_text(json.dumps(timing,indent=2));print(json.dumps(out,indent=2));print(json.dumps(timing,indent=2))

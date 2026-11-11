from pathlib import Path
import json,collections,hashlib
S=Path(__file__).resolve().parent;results={}
for side in ['base','after','mutant-object','mutant-intrinsic']:
 rows=[json.loads(l) for l in (S/(side+'-census.jsonl')).read_text().splitlines()];assert not rows[0]['diagnostics'] and not rows[0]['checker_rejected'];assert rows[-1]['status']=='relation_report';files=[r for r in rows if 'file' in r];assert len(files)==1 and len(files[0]['units'])==23
 findings=files[0]['findings'];casts=[f for f in findings if f['reason']=="a cast the runtime can't check"]
 results[side]=dict(units=23,diagnostics=0,kinds=dict(collections.Counter(f['kind'] for f in findings)),casts=[dict(phase=f['phase'],where=f['where']) for f in casts])
assert len(results['base']['casts'])==3 and not results['after']['casts']
assert [(f['phase'],f['where'].split('tracing.ts:')[1]) for f in results['mutant-object']['casts']]==[('refusal_scan','234:34')]
assert [(f['phase'],f['where'].split('tracing.ts:')[1]) for f in results['mutant-intrinsic']['casts']]==[('refusal_scan','320:33'),('lowering','320:33')]
(S/'census-mutant-audit-results.json').write_text(json.dumps(results,indent=2)+'\n');print(json.dumps(results))
D=Path('/workspace/adamic/stage3/adapt/99-correctness-tracing-dumptypes');frozen={name:hashlib.sha256((D/name).read_bytes()).hexdigest() for name in ['adapt.cjs','pins.json']};(S/'frozen-adapter.json').write_text(json.dumps(frozen,indent=2)+'\n')

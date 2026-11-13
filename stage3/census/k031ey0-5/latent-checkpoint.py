import json,collections
from pathlib import Path
out=Path(__file__).resolve().parent
summary={}
for side in ['parent','head']:
 st=out/(side+'-latent-status.json');raw=out/(side+'-latent.jsonl')
 if not raw.exists():summary[side]={'state':'unrun'};continue
 rows=[json.loads(l) for l in raw.read_text().splitlines()];status=json.loads(st.read_text()) if st.exists() else None
 pre=[l.rsplit('\t',2) for l in (out/(side+'-prepass-times.tsv')).read_text().splitlines()];times=[int(x[-1])/1e9 for x in pre if x[0]=='collectViewedElements']
 units=[u for r in rows for u in r.get('units',[])];findings=[f for r in rows for f in r.get('findings',[])];nf=[f for f in findings if 'which instantiates itself deeper at each step, so no walk over it ends' in f.get('reason','')];errors=collections.Counter(l for l in (out/(side+'-latent.log')).read_text().splitlines() if l.startswith('prepass collectViewedElements error:'))
 timer=out/(side+'-latent-times.jsonl');unit_times=[json.loads(l) for l in timer.read_text().splitlines()]
 summary[side]=dict(state=('completed' if status['exit_code']==0 else 'bounded stop') if status else 'running snapshot',status=status,files_emitted=sum('file' in r for r in rows),units_emitted=len(units),findings_emitted=len(findings),new_reason_findings=nf,new_reason_units=sorted(set(f['unit'] for f in nf)),unit_timers=len(unit_times),latest_unit_timer=unit_times[-1] if unit_times else None,prepass_calls=len(times),prepass_seconds=sum(times),stderr_prepass_errors=dict(errors))
(out/'latent-checkpoint.json').write_text(json.dumps(summary,indent=2)+'\n')
r=['# Entry latent checkpoint','', 'First-error mode, LATENT_ASSERT_NO_OUTPUT=1; LATENT_TIMES and LATENT_PREPASS_TIMES set. Head has only the saved scratch seam. Findings are buffered until the entire pass finishes; a timed-out or running pass cannot supply a census of unit stops. Prepass totals for such a pass are partial and are not whole-pass comparison numbers.','']
for side,s in summary.items():
 r+=['## '+side,'','```json',json.dumps(s,indent=2),'```','']
(out/'LATENT.md').write_text('\n'.join(r))
print(json.dumps(summary,indent=2))

#!/usr/bin/env python3
"""Validate paired ledgers and write reason, unit-transition and timing evidence."""
import collections, gzip, importlib.util, json, statistics, sys
from pathlib import Path
out = Path(__file__).resolve().parent
repo = Path('/workspace/scratch/change-a')
spec = importlib.util.spec_from_file_location('meter_report', repo/'stage3/meter/report.py')
meter = importlib.util.module_from_spec(spec); spec.loader.exec_module(meter)
def dump(name, data): (out/name).write_text(json.dumps(data, indent=2)+'\n')
def read(name):
    path=out/name
    data=path.read_text() if path.exists() else gzip.decompress(Path(str(path)+'.gz').read_bytes()).decode()
    return [json.loads(s) for s in data.splitlines()]
def norm(s): return s.replace('/workspace/scratch/tree/', '')
expected={str(p.resolve()) for p in Path('/workspace/scratch/tree/src/compiler').rglob('*') if p.is_file() and p.suffix in ('.ts','.a')}
a,b=read('A.jsonl'),read('B.jsonl')
validated={label:meter.lowering_summary(rows, expected, meter.MEASUREMENT) for label,rows in [('A',a),('B',b)]}
dump('validation.json',validated)
def sites(rows): return {tuple(norm(f.get(k,'')) for k in ('kind','where','reason','text')) for r in rows[1:] for f in r.get('findings') or []}
sa,sb=sites(a),sites(b)
reasons=[]
for kind,reason in sorted({(s[0],s[2]) for s in sa|sb if s[0] in ('NotYet','Refused')}):
 x=[s for s in sa if s[0]==kind and s[2]==reason]; y=[s for s in sb if s[0]==kind and s[2]==reason]
 if len(x)!=len(y):
  reasons.append(dict(kind=kind,reason=reason,A=len(x),B=len(y),delta=len(y)-len(x),A_examples=[s[1] for s in sorted(set(x)-sb)[:5]],B_examples=[s[1] for s in sorted(set(y)-sa)[:5]],callout=any(w in reason.lower() for w in ('view','branded supertype','predicate'))))
dump('reason-changes.json',reasons)
lines=['| Kind | Reason | A | B | Delta | Removed A examples | Added B examples |','|---|---|---:|---:|---:|---|---|']
for r in reasons:
 esc=lambda s:s.replace('|','&#124;').replace('\n',' ')
 lines.append('| '+ ' | '.join(map(esc,[r['kind'],('**CALLOUT** ' if r['callout'] else '')+r['reason'],str(r['A']),str(r['B']),str(r['delta']),'<br>'.join(r['A_examples']),'<br>'.join(r['B_examples'])]))+' |')
(out/'reason-changes.md').write_text('\n'.join(lines)+'\n')
def units(rows):
 result={}
 for row in rows[1:]:
  for u in row.get('units',[]):
   key=norm(u['where']); result[key]=dict(u)
   result[key]['lowering_stops']=[f for f in row.get('findings') or [] if f.get('unit')==u['where'] and f.get('phase')=='lowering' and f['kind'] in ('NotYet','Refused','error','panic','SkippedDependency')]
 return result
ua,ub=units(a),units(b)
transitions=[]
for key,u in ua.items():
 if u['status']!='attempted' or u['lowering_stops']: continue
 v=ub.get(key)
 if v is None or v['status']!='attempted' or v['lowering_stops']: transitions.append(dict(where=key,A=u,B=v))
dump('lowering-regressions.json',transitions)
newstops=[dict(zip(('kind','where','reason','text'),s)) for s in sorted(sb-sa) if s[0] in ('NotYet','Refused','error','panic')]
dump('new-stops.json',newstops)
(out/'lowering-regressions.md').write_text('A attempted units with no lowering-phase stop, absent or stopped in B: '+str(len(transitions))+'\n\n'+('\n'.join('- `'+x['where']+'`: '+json.dumps(x['B']) for x in transitions) if transitions else 'None.')+'\n\nThis compares ledger units and their lowering-phase findings; the ledger does not enumerate every successful expression. All new distinct stops are in new-stops.json.\n')
def times(name):
 rows=read(name); d=collections.defaultdict(lambda:[0.,0.,0])
 for r in rows:
  x=d[norm(r['where'])]; x[0]+=r['wall_seconds']; x[1]+=r['cpu_seconds']; x[2]+=1
 return d
at,bt=times('A.times.jsonl'),times('B.times.jsonl')
assert at.keys()==bt.keys(), 'timing unit coverage differs'
rows=[]
for k in at:
 x,y=at[k],bt[k]; rows.append(dict(where=k,name=ua.get(k,{}).get('name',''),A_seconds=x[0],B_seconds=y[0],delta_seconds=y[0]-x[0],ratio=y[0]/x[0] if x[0] else None,A_cpu=x[1],B_cpu=y[1],A_samples=x[2],B_samples=y[2]))
rows.sort(key=lambda r:-r['delta_seconds'])
dump('timing-table.json',rows)
# Non-function declarations that reach the same first stop are controls for registration/timer/scheduling costs.
controls=[r for r in rows if ua.get(r['where'],{}).get('kind') in ('KindImportDeclaration','KindImportEqualsDeclaration','KindInterfaceDeclaration','KindTypeAliasDeclaration') and ua.get(r['where'],{}).get('lowering_stops')==ub.get(r['where'],{}).get('lowering_stops')]
def noise(rs):
 ds=sorted(abs(r['delta_seconds']) for r in rs)
 return dict(units=len(ds),median_absolute_delta=statistics.median(ds) if ds else None,p95_absolute_delta=ds[min(len(ds)-1,int(.95*len(ds)))] if ds else None,median_ratio=statistics.median(r['ratio'] for r in rs if r['ratio'] is not None) if rs else None)
summary=dict(A_total_seconds=sum(r['A_seconds'] for r in rows),B_total_seconds=sum(r['B_seconds'] for r in rows),A_cpu_seconds=sum(r['A_cpu'] for r in rows),B_cpu_seconds=sum(r['B_cpu'] for r in rows),units=len(rows),controls=noise(controls),control_definition='Imports, interfaces and type aliases with identical lowering findings: no checked-predicate body to exercise B. These include registration costs; one paired run cannot separate scheduling, GC and common registration overhead.')
dump('timing-summary.json',summary)
lines=['Unit timing includes registration, lowering and panic recovery. Ratio = B/A.','',json.dumps(summary,indent=2),'','| Unit | Name | A seconds | B seconds | B-A seconds | B/A |','|---|---|---:|---:|---:|---:|']
for r in rows: lines.append(f"| {r['where']} | {r['name']} | {r['A_seconds']:.6f} | {r['B_seconds']:.6f} | {r['delta_seconds']:+.6f} | {r['ratio']:.4f} |")
(out/'timing-table.md').write_text('\n'.join(lines)+'\n')
print(json.dumps(dict(counts={k:v['totals'] for k,v in validated.items()},reason_changes=len(reasons),lowering_regressions=len(transitions),timing=summary)))

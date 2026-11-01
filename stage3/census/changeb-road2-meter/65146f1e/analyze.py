#!/usr/bin/env python3
"""Validate paired ledgers and report distinct sites, unit outcomes and timing."""
import collections,hashlib,importlib.util,json,pathlib,re,statistics
repo=pathlib.Path('/workspace/adamic')
out=repo/'stage3/census/changeb-road2-meter/65146f1e'
raw=pathlib.Path('/workspace/scratch/paired-65146f1e')
tree=pathlib.Path('/workspace/scratch/tree')
spec=importlib.util.spec_from_file_location('meter_report',repo/'stage3/meter/report.py')
meter=importlib.util.module_from_spec(spec);spec.loader.exec_module(meter)
def read(p): return [json.loads(s) for s in p.read_text().splitlines()]
def norm(s): return s.replace(str(tree)+'/', '')
def write(name,obj): (out/name).write_text(json.dumps(obj,indent=2)+'\n')
def esc(s): return str(s).replace('|','&#124;').replace('\n',' ')
def table(name,columns,rows):
    (out/name).write_text('| '+' | '.join(columns)+' |\n| '+' | '.join(['---']*len(columns))+' |\n'+''.join('| '+' | '.join(esc(v) for v in row)+' |\n' for row in rows))
expected={str(p.resolve()) for p in (tree/'src/compiler').rglob('*') if p.is_file() and p.suffix in ('.ts','.a')}
assert len(expected)==79
records={s:read(raw/f'{s}.jsonl') for s in 'AB'}
summaries={s:meter.lowering_summary(records[s],expected,meter.MEASUREMENT) for s in 'AB'}
write('validation.json',summaries)
sites={s:{tuple(norm(f[k]) for k in ('kind','where','reason','text')) for r in records[s][1:] for f in r.get('findings') or [] if f['kind'] in ('NotYet','Refused')} for s in 'AB'}
counts={s:collections.Counter((x[0],x[2]) for x in sites[s]) for s in 'AB'}
changed=[]
for key in sorted(counts['A'].keys()|counts['B'].keys()):
    a,b=(counts[s][key] for s in 'AB')
    if a==b: continue
    examples={s:[dict(zip(('kind','where','reason','text'),x)) for x in sorted(sites[s]) if (x[0],x[2])==key][:5] for s in 'AB'}
    changed.append(dict(kind=key[0],reason=key[1],A=a,B=b,delta=b-a,examples=examples))
write('changed-reasons.json',changed)
table('changed-reasons.md',['Kind','Reason','A','B','B−A','A examples (up to 5)','B examples (up to 5)'],[[r['kind'],r['reason'],r['A'],r['B'],r['delta'],'; '.join(x['where'] for x in r['examples']['A']),'; '.join(x['where'] for x in r['examples']['B'])] for r in changed])
focus=[]
for key in sorted(counts['A'].keys()|counts['B'].keys()):
    if re.search(r'view|branded supertype|predicate',key[1],re.I): focus.append(dict(kind=key[0],reason=key[1],A=counts['A'][key],B=counts['B'][key]))
write('view-predicate-reasons.json',focus)
table('view-predicate-reasons.md',['Kind','Reason','A','B'],[[r['kind'],r['reason'],r['A'],r['B']] for r in focus])
units={};stops={}
for side in 'AB':
    units[side]={norm(u['where']):dict(u,file=norm(r['file'])) for r in records[side][1:] for u in r['units']}
    stops[side]=collections.defaultdict(set)
    for r in records[side][1:]:
        for f in r.get('findings') or []:
            if f['phase']=='lowering': stops[side][norm(f['unit'])].add(tuple(norm(f[k]) for k in ('kind','where','reason','text')))
regressions=[];outcome_changes=[]
for where,u in sorted(units['A'].items()):
    v=units['B'].get(where)
    a=sorted(stops['A'][where]);b=sorted(stops['B'][where])
    if not v or u['status']!=v['status'] or a!=b:
        row=dict(where=where,name=u.get('name',''),A_status=u['status'],B_status=v['status'] if v else 'absent',A_stops=a,B_stops=b)
        outcome_changes.append(row)
        if u['status']=='attempted' and not a and (not v or v['status']!='attempted' or b): regressions.append(row)
write('unit-outcome-changes.json',outcome_changes)
write('lowering-regressions.json',regressions)
table('lowering-regressions.md',['Unit','Name','B status','B stops'],[[r['where'],r['name'],r['B_status'],json.dumps(r['B_stops'])] for r in regressions])
times={}
for s in 'AB':
    rows=read(raw/f'{s}-times.jsonl')
    times[s]={norm(r['where']):r for r in rows}
    assert len(times[s])==len(rows),'duplicate timing key'
    assert set(times[s])=={k for k,u in units[s].items() if u['status'] in ('attempted','panic')},'attempted unit timing coverage'
paired=[]
for where in sorted(times['A'].keys()&times['B'].keys()):
    a,b=times['A'][where],times['B'][where]
    paired.append(dict(where=where,name=units['A'][where].get('name',''),kind=units['A'][where]['kind'],A_seconds=a['wall_seconds'],B_seconds=b['wall_seconds'],delta_seconds=b['wall_seconds']-a['wall_seconds'],ratio=b['wall_seconds']/a['wall_seconds'] if a['wall_seconds'] else None,A_cpu_seconds=a['cpu_seconds'],B_cpu_seconds=b['cpu_seconds']))
paired.sort(key=lambda r:(-r['delta_seconds'],r['where']))
write('unit-times.json',paired)
columns=['Unit','Name','A seconds','B seconds','B−A seconds','B/A']
def timingrow(r):return [r['where'],r['name'],f"{r['A_seconds']:.6f}",f"{r['B_seconds']:.6f}",f"{r['delta_seconds']:+.6f}",f"{r['ratio']:.4f}"]
table('timing-top20.md',columns,[timingrow(r) for r in paired[:20]])
table('timing-all.md',columns,[timingrow(r) for r in paired])
# Declaration/import units do not lower a runtime predicate call; their identical
# paths give a local scheduling/GC baseline, not a statistical confidence interval.
noise=[r for r in paired if r['kind'] in ('KindImportDeclaration','KindInterfaceDeclaration','KindTypeAliasDeclaration') and stops['A'][r['where']]==stops['B'][r['where']]]
def pct(xs,p):
    xs=sorted(xs);return xs[min(len(xs)-1,round((len(xs)-1)*p))] if xs else None
noise_summary=dict(definition='Same-outcome import/interface/type-alias units: no runtime branded-predicate lowering path. Shared registration, GC and scheduling remain in these measurements.',n=len(noise),median_signed_delta_seconds=statistics.median(r['delta_seconds'] for r in noise) if noise else None,median_absolute_delta_seconds=statistics.median(abs(r['delta_seconds']) for r in noise) if noise else None,p90_absolute_delta_seconds=pct([abs(r['delta_seconds']) for r in noise],.9),median_ratio=statistics.median(r['ratio'] for r in noise) if noise else None,p10_ratio=pct([r['ratio'] for r in noise],.1),p90_ratio=pct([r['ratio'] for r in noise],.9))
write('noise-baseline.json',dict(summary=noise_summary,units=noise))
totals={s:{k:sum(r[k] for r in times[s].values()) for k in ('wall_seconds','cpu_seconds')}|{'units':len(times[s])} for s in 'AB'}
write('timing-summary.json',dict(totals=totals,noise=noise_summary,unpaired={s:sorted(times[s].keys()-times['B' if s=='A' else 'A'].keys()) for s in 'AB'}))
write('headline.json',dict(totals={s:summaries[s]['totals'] for s in 'AB'},files={s:summaries[s]['source_files'] for s in 'AB'},changed_reason_rows=len(changed),unit_outcome_changes=len(outcome_changes),lowering_regressions=len(regressions),timing_totals=totals,noise=noise_summary))
print((out/'headline.json').read_text(), end='')
table('unit-outcome-changes.md',['Unit','Name','A first stop','B first stop'],[[r['where'],r['name'],'; '.join(f'{x[0]} at {x[1]}: {x[2]}' for x in r['A_stops']) or '(none)','; '.join(f'{x[0]} at {x[1]}: {x[2]}' for x in r['B_stops']) or '(none)'] for r in outcome_changes])
write('unit-coverage.json',{s:dict(total_units=len(units[s]),statuses=dict(collections.Counter(u['status'] for u in units[s].values())),attempted_without_lowering_finding=sum(u['status']=='attempted' and not stops[s][w] for w,u in units[s].items())) for s in 'AB'})
write('prior-regressions-check.json',[dict(where=w,name=units['A'][w].get('name',''),A_status=units['A'][w]['status'],B_status=units['B'][w]['status'],A_stops=sorted(stops['A'][w]),B_stops=sorted(stops['B'][w])) for w in ('src/compiler/scanner.ts:520:1','src/compiler/utilities.ts:1351:1')])
assert all(units[s][w]['status']=='attempted' and not stops[s][w] for s in 'AB' for w in ('src/compiler/scanner.ts:520:1','src/compiler/utilities.ts:1351:1'))

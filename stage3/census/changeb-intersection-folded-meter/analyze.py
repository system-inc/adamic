"""Validate both original ledgers and derive paired sites, units, and timings."""
import collections, gzip, importlib.util, json, math, pathlib, statistics
P=pathlib.Path(__file__).resolve().parent
TREE=pathlib.Path('/workspace/scratch/tree')
REPO=pathlib.Path('/workspace/scratch/compiler-a')
spec=importlib.util.spec_from_file_location('meter',REPO/'stage3/meter/report.py')
meter=importlib.util.module_from_spec(spec); spec.loader.exec_module(meter)
def read(p):
    text=p.read_text() if p.exists() else gzip.open(str(p)+'.gz','rt').read()
    return [json.loads(x) for x in text.splitlines()]
def norm(s): return s.replace(str(TREE)+'/', '')
def save(name,value): (P/name).write_text(json.dumps(value,indent=2)+'\n')
def cell(s): return str(s).replace('|','&#124;').replace('\n',' ')
def table(headers,rows): return '\n'.join(['| '+' | '.join(headers)+' |','|'+'|'.join(['---']*len(headers))+'|']+['| '+' | '.join(cell(v) for v in r)+' |' for r in rows])+'\n'
ledgers={s:read(P/f'{s}.jsonl') for s in 'ab'}
names={x['where']:x['name'] for x in json.loads((P/'direct-callers.json').read_text())['catalog']}
expected={str(p.resolve()) for p in (TREE/'src/compiler').rglob('*') if p.is_file() and p.suffix in ('.ts','.a')}
summary={s:meter.lowering_summary(r,expected,meter.MEASUREMENT) for s,r in ledgers.items()}
save('validation.json',summary)
units={}; findings={}; sites={}
for s,records in ledgers.items():
    units[s]={norm(u['where']):dict(u,file=norm(r['file']),name=names.get(norm(u['where']),u.get('name',''))) for r in records[1:] for u in r['units']}
    findings[s]=collections.defaultdict(list)
    sites[s]=set()
    for r in records[1:]:
        for f in r['findings'] or []:
            n={k:norm(v) if isinstance(v,str) else v for k,v in f.items()}
            if f['phase']=='lowering': findings[s][norm(f['unit'])].append(n)
            sites[s].add(tuple(n[k] for k in ('kind','where','reason','text')))

def outcome(s,w):
    u=units[s].get(w)
    if u is None: return {'outcome':'absent'}
    stops=findings[s].get(w,[])
    if stops: return {'outcome':'stopped','kind':stops[0]['kind'],'reason':stops[0]['reason'],'where':stops[0]['where'],'all_stops':stops}
    if u['status']=='attempted': return {'outcome':'lowers'}
    return {'outcome':u['status'],'checker_diagnostics':u.get('checker_diagnostics',[]),'instantiated':u.get('instantiated',False)}
def label(o):
    if o['outcome']=='stopped': return o['kind']+': '+o['reason']+' @ '+o['where']
    if o['outcome']=='skipped_checker_body': return 'checker body skipped: '+('; '.join(o.get('checker_diagnostics',[])[:1]))[:350]+'… (full diagnostic in caller-outcomes.json)'
    return o['outcome']
changes=[]; regressions=[]; missing=[]
for w in sorted(units['a'].keys()|units['b'].keys()):
    a,b=outcome('a',w),outcome('b',w)
    if {k:v for k,v in a.items() if k!='all_stops'}!={k:v for k,v in b.items() if k!='all_stops'}:
        row={'unit':w,'name':(units['a'].get(w) or units['b'][w]).get('name',''),'A':a,'B':b}
        changes.append(row)
        if a['outcome']=='lowers' and b['outcome']=='stopped': regressions.append(row)
        if a['outcome']=='lowers' and b['outcome']!='lowers': missing.append(row)
save('unit-changes.json',changes); save('a-lowers-b-stops.json',regressions); save('a-lowers-b-absent-or-stopped.json',missing)
(P/'unit-changes.md').write_text(table(['Unit','Name','A','B'],[(r['unit'],r['name'],label(r['A']),label(r['B'])) for r in changes]))
# Preserve every distinct A lowering-phase finding absent from B's ledger, and
# every distinct A finding whose lowering phase occurrence disappears in B.
a_lower={tuple(f[k] for k in ('kind','where','reason','text')) for rows in findings['a'].values() for f in rows}
b_lower={tuple(f[k] for k in ('kind','where','reason','text')) for rows in findings['b'].values() for f in rows}
removed=[dict(zip(('kind','where','reason','text'),x)) for x in sorted(a_lower-sites['b'])]
save('a-lowering-sites-absent-b.json',removed)
save('a-lowering-phase-disappeared.json',[dict(zip(('kind','where','reason','text'),x)) for x in sorted(a_lower-b_lower)])
(P/'a-lowering-sites-absent-b.md').write_text(table(['Kind','Location','Reason'],[(r['kind'],r['where'],r['reason']) for r in removed]))
counts={s:collections.Counter(x[2] for x in v if x[0] in ('NotYet','Refused')) for s,v in sites.items()}
reason_rows=[]
for reason in sorted(counts['a'].keys()|counts['b'].keys()):
    if counts['a'][reason]==counts['b'][reason]: continue
    row={'reason':reason,'A':counts['a'][reason],'B':counts['b'][reason],
         'flagged':any(x in reason.lower() for x in ['view','branded supertype','intersection','predicate']),
         'A_examples':sorted({x[1] for x in sites['a'] if x[2]==reason})[:5],
         'B_examples':sorted({x[1] for x in sites['b'] if x[2]==reason})[:5]}
    reason_rows.append(row)
save('changed-reasons.json',reason_rows)
(P/'changed-reasons.md').write_text(table(['Reason (* = view/brand/intersection/predicate)','A','B','A examples (up to 5)','B examples (up to 5)'],[(('* ' if r['flagged'] else '')+r['reason'],r['A'],r['B'],'; '.join(r['A_examples']),'; '.join(r['B_examples'])) for r in reason_rows]))
callers=json.loads((P/'direct-callers.json').read_text())
caller_rows=[]
for c in callers['callers']:
    # AST and meter use the first token's location. Fall back to containing span
    # for declaration attributes if their first-token conventions ever differ.
    w=c['where']
    if w not in units['a']:
        matches=[k for k,u in units['a'].items() if u['file']==c['file'] and u.get('name')==c['name']]
        assert len(matches)==1,(c,matches)
        w=matches[0]
    caller_rows.append(dict(c,unit=w,A=outcome('a',w),B=outcome('b',w)))
assert any(c['name']=='nullParenthesizerRules' for c in caller_rows)
save('caller-outcomes.json',caller_rows)
(P/'caller-outcomes.md').write_text(table(['Unit','Name','Direct calls','A','B'],[(r['unit'],r['name'],len(r['calls']),label(r['A']),label(r['B'])) for r in caller_rows]))
times={s:{norm(r['where']):r for r in read(P/f'{s}-times.jsonl')} for s in 'ab'}
assert times['a'].keys()==times['b'].keys(),'timing unit populations differ'
for s in 'ab':
    assert times[s].keys()=={w for w,u in units[s].items() if u['status'] in ('attempted','panic')},'timing coverage differs from attempted units'
tr=[]
for w,a in times['a'].items():
    b=times['b'][w]; delta=b['wall_seconds']-a['wall_seconds']
    tr.append({'unit':w,'name':units['a'][w].get('name',''),'kind':units['a'][w]['kind'],
               'A_seconds':a['wall_seconds'],'B_seconds':b['wall_seconds'],'B_minus_A_seconds':delta,
               'B_over_A':b['wall_seconds']/a['wall_seconds'] if a['wall_seconds'] else None,
               'A_cpu_seconds':a['cpu_seconds'],'B_cpu_seconds':b['cpu_seconds']})
tr.sort(key=lambda x:(-x['B_minus_A_seconds'],x['unit']))
save('unit-times.json',tr)
(P/'unit-times.tsv').write_text('unit\tname\tA_seconds\tB_seconds\tB_minus_A_seconds\tB_over_A\tA_cpu_seconds\tB_cpu_seconds\n'+''.join('\t'.join(str(r[k]) for k in ['unit','name','A_seconds','B_seconds','B_minus_A_seconds','B_over_A','A_cpu_seconds','B_cpu_seconds'])+'\n' for r in tr))
def time_table(rows): return table(['Unit','Name','A seconds','B seconds','B−A seconds','B/A'],[(r['unit'],r['name'],f"{r['A_seconds']:.6f}",f"{r['B_seconds']:.6f}",f"{r['B_minus_A_seconds']:+.6f}",f"{r['B_over_A']:.3f}") for r in rows])
(P/'timing-top20.md').write_text(time_table(tr[:20]))
# Import declarations execute no property read or predicate test and cannot
# exercise B's intersection rule; they still pay the shared registration cost.
control=[r for r in tr if r['kind']=='KindImportDeclaration']
assert control
absd=sorted(abs(r['B_minus_A_seconds']) for r in control)
def quantile(xs,p): return sorted(xs)[round((len(xs)-1)*p)]
noise={'selection':'ImportDeclaration: no lowered receiver read, destructure or predicate; registration still included','units':len(control),
       'median_abs_seconds':statistics.median(absd),'p90_abs_seconds':quantile(absd,.9),'p95_abs_seconds':quantile(absd,.95),'max_abs_seconds':max(absd),
       'median_B_over_A':statistics.median(r['B_over_A'] for r in control),
       'p05_B_over_A':quantile([r['B_over_A'] for r in control],.05),'p95_B_over_A':quantile([r['B_over_A'] for r in control],.95),
       'A_total_seconds':sum(r['A_seconds'] for r in control),'B_total_seconds':sum(r['B_seconds'] for r in control)}
save('timing-noise.json',noise)
(P/'timing-controls.md').write_text(time_table(control))
totals={s:{'wall_seconds':sum(r['wall_seconds'] for r in times[s].values()),'cpu_seconds':sum(r['cpu_seconds'] for r in times[s].values()),'units':len(times[s])} for s in 'ab'}
save('timing-totals.json',totals)
headline={'files':len(expected),'summary':{s:summary[s]['totals'] for s in 'ab'},'changed_reasons':len(reason_rows),'unit_changes':len(changes),'A_lowers_B_stops':len(regressions),'A_lowers_B_other':len(missing),'removed_A_lowering_sites':len(removed),'direct_calls':callers['calls'],'caller_units':callers['units'],'timing_totals':totals,'noise':noise}
save('headline.json',headline)
print(json.dumps(headline,indent=2))

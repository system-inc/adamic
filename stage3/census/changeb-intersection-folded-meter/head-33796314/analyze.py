"""Compare the alone-run C head against frozen A and B ledgers."""
import collections,gzip,hashlib,importlib.util,json,pathlib,statistics
P=pathlib.Path(__file__).resolve().parent
BASE=P.parent
TREE=pathlib.Path('/workspace/scratch/tree')
GUARD='a union or optional field read in a program with record storage'
REFS={'A':'3aec8230abb53ac93b92b24f5e3ae916f9cd36a8','B':'5e5f57d3b530264cb09eba5712a8a362375aede2','C':'33796314e5bd17ec159c0c4a92dd633552bf27ab'}
def read(p):
    return [json.loads(x) for x in (p.read_text() if p.exists() else gzip.open(str(p)+'.gz','rt').read()).splitlines()]
def norm(s): return s.replace(str(TREE)+'/', '')
def save(path,value): path.write_text(json.dumps(value,indent=2)+'\n')
def cell(v): return str(v).replace('&','&amp;').replace('<','&lt;').replace('>','&gt;').replace('|','&#124;').replace('\n',' ')
def table(headers,rows): return '\n'.join(['| '+' | '.join(headers)+' |','|'+'|'.join(['---']*len(headers))+'|']+['| '+' | '.join(cell(x) for x in row)+' |' for row in rows])+'\n'
spec=importlib.util.spec_from_file_location('meter','/workspace/scratch/compiler-a/stage3/meter/report.py')
meter=importlib.util.module_from_spec(spec);spec.loader.exec_module(meter)
records={'A':read(BASE/'a.jsonl'),'B':read(BASE/'b.jsonl'),'C':read(P/'c.jsonl')}
expected={str(x.resolve()) for x in (TREE/'src/compiler').rglob('*') if x.is_file() and x.suffix in ('.ts','.a')}
assert len(expected)==79
summary={s:meter.lowering_summary(r,expected,meter.MEASUREMENT) for s,r in records.items()}
save(P/'validation.json',summary)
assert records['A'][0]==records['B'][0]==records['C'][0],'checker headers differ'
callers=json.loads((BASE/'direct-callers.json').read_text())
names={r['where']:r['name'] for r in callers['catalog']}
units={};first={};sites={};owners={};lower_sites={}
for s,rs in records.items():
    units[s]={norm(u['where']):dict(u,name=names.get(norm(u['where']),u.get('name','')),file=norm(r['file'])) for r in rs[1:] for u in r['units']}
    findings=collections.defaultdict(list);sites[s]=set();owners[s]=collections.defaultdict(set);lower_sites[s]=set()
    for r in rs[1:]:
        for f in r['findings'] or []:
            f={k:norm(v) if isinstance(v,str) else v for k,v in f.items()}
            key=tuple(f[k] for k in ('kind','where','reason','text'))
            sites[s].add(key);owners[s][key].add(f['unit'])
            if f['phase']=='lowering': findings[f['unit']].append(f);lower_sites[s].add(key)
    first[s]={}
    for w,u in units[s].items():
        fs=findings.get(w,[])
        if fs: o={'outcome':'stopped',**{k:fs[0][k] for k in ('kind','where','reason')},'all_stops':fs}
        elif u['status']=='attempted': o={'outcome':'lowers'}
        else: o={'outcome':u['status'],'checker_diagnostics':u.get('checker_diagnostics',[]),'instantiated':u.get('instantiated',False)}
        first[s][w]=o

def outcome(s,w): return first[s].get(w,{'outcome':'absent'})
def fingerprint(o): return {k:o[k] for k in ('outcome','kind','where','reason') if k in o}
def label(o):
    if o['outcome']=='stopped': return o['kind']+': '+o['reason']+' @ '+o['where']
    if o['outcome']=='skipped_checker_body': return 'checker body skipped: '+norm('; '.join(o.get('checker_diagnostics',[])[:1]))[:350]+'… (full diagnostic in JSON)'
    return o['outcome']
def site_rows(s,keys):
    return [dict(zip(('kind','where','reason','text'),k),units=[{'where':u,'name':names.get(u,'')} for u in sorted(owners[s][k])]) for k in sorted(keys)]

pair_summaries={}
for left,sub in [('A','vs-a'),('B','vs-5e5f57d3')]:
    out=P/sub;out.mkdir(exist_ok=True)
    changes=[];reg=[];missing=[]
    for w in sorted(units[left].keys()|units['C'].keys()):
        a,c=outcome(left,w),outcome('C',w)
        if fingerprint(a)!=fingerprint(c):
            row={'unit':w,'name':names.get(w,''),left:a,'C':c,'record_guard':c.get('reason')==GUARD}
            changes.append(row)
            if a['outcome']=='lowers' and c['outcome']=='stopped': reg.append(row)
            if a['outcome']=='lowers' and c['outcome']!='lowers': missing.append(row)
    metadata=[{'unit':w,'name':names.get(w,''),left:units[left].get(w),'C':units['C'].get(w)} for w in sorted(units[left].keys()|units['C'].keys()) if units[left].get(w)!=units['C'].get(w)]
    save(out/'unit-metadata-changes.json',metadata)
    save(out/'unit-changes.json',changes)
    (out/'unit-changes.md').write_text(table(['Unit','Name',left,'C','New stop is record guard'],[(r['unit'],r['name'],label(r[left]),label(r['C']),r['record_guard']) for r in changes]))
    guarded=[r for r in reg if r['record_guard']];other=[r for r in reg if not r['record_guard']]
    save(out/'lowers-to-stops-record-guard.json',guarded);save(out/'lowers-to-stops-other.json',other)
    save(out/'lowers-to-absent-or-nonlowering.json',missing)
    for name,rows in [('record-guard-regressions.md',guarded),('other-regressions.md',other)]:
        (out/name).write_text(table(['Unit','Name','C reason','C location'],[(r['unit'],r['name'],r['C']['reason'],r['C']['where']) for r in rows]))
    removed=site_rows(left,sites[left]-sites['C']);added=site_rows('C',sites['C']-sites[left])
    save(out/'removed-sites.json',removed);save(out/'added-sites.json',added)
    (out/'site-movements.md').write_text(table(['Change','Kind','Location','Reason','Units'],[(side,r['kind'],r['where'],r['reason'],'; '.join(u['name']+' @ '+u['where'] for u in r['units'])) for side,rs in [('removed',removed),('added',added)] for r in rs]))
    absent=site_rows(left,lower_sites[left]-sites['C'])
    save(out/'lowering-sites-absent-c.json',absent)
    (out/'lowering-sites-absent-c.md').write_text(table(['Kind','Location','Reason'],[(r['kind'],r['where'],r['reason']) for r in absent]))
    a_count=collections.Counter(k[2] for k in sites[left] if k[0] in ('NotYet','Refused'))
    c_count=collections.Counter(k[2] for k in sites['C'] if k[0] in ('NotYet','Refused'))
    reasons=[]
    for reason in sorted(a_count.keys()|c_count.keys()):
        if a_count[reason]==c_count[reason]: continue
        reasons.append({'reason':reason,left:a_count[reason],'C':c_count[reason],'flagged':any(t in reason.lower() for t in ['view','branded supertype','intersection','predicate']),left+'_examples':sorted({k[1] for k in sites[left] if k[2]==reason})[:5],'C_examples':sorted({k[1] for k in sites['C'] if k[2]==reason})[:5]})
    save(out/'changed-reasons.json',reasons)
    (out/'changed-reasons.md').write_text(table(['Reason (* = view/brand/intersection/predicate)',left,'C',left+' examples (up to 5)','C examples (up to 5)'],[(('* ' if r['flagged'] else '')+r['reason'],r[left],r['C'],'; '.join(r[left+'_examples']),'; '.join(r['C_examples'])) for r in reasons]))
    pair_summaries[left+'-C']={'changed_reasons':len(reasons),'unit_changes':len(changes),'lowers_to_stops_record_guard':len(guarded),'lowers_to_stops_other':len(other),'lowers_to_absent_or_nonlowering':len(missing),'removed_distinct_sites_all_kinds':len(removed),'added_distinct_sites_all_kinds':len(added),'removed_lowering_sites':len(absent),'all_changed_unit_new_stops_are_record_guard':all(r['record_guard'] for r in changes),'all_added_sites_are_record_guard':all(r['reason']==GUARD for r in added),'non_guard_added_sites':[r for r in added if r['reason']!=GUARD]}

guard_stops=[{'unit':w,'name':names.get(w,''),'A':outcome('A',w),'B':outcome('B',w),'C':o} for w,o in sorted(first['C'].items()) if o.get('reason')==GUARD]
save(P/'record-guard-first-stops.json',guard_stops)
(P/'record-guard-first-stops.md').write_text(table(['Unit','Name','A','B','C'],[(r['unit'],r['name'],label(r['A']),label(r['B']),label(r['C'])) for r in guard_stops]))
guard_sites=site_rows('C',{k for k in sites['C'] if k[2]==GUARD})
save(P/'record-guard-sites.json',guard_sites)
(P/'record-guard-sites.md').write_text(table(['Kind','Site','Reason','Units'],[(r['kind'],r['where'],r['reason'],'; '.join(u['name']+' @ '+u['where'] for u in r['units'])) for r in guard_sites]))
cr=[]
for caller in callers['callers']:
    w=caller['where'];assert w in units['A'] and w in units['C'],caller
    cr.append(dict(caller,**{s:outcome(s,w) for s in 'ABC'}))
save(P/'caller-outcomes.json',cr)
(P/'caller-outcomes.md').write_text(table(['Unit','Name','Calls','A','B (5e5f57d3)','C (33796314)'],[(r['where'],r['name'],len(r['calls']),label(r['A']),label(r['B']),label(r['C'])) for r in cr]))

times={'A':read(BASE/'a-times.jsonl'),'B':read(BASE/'b-times.jsonl'),'C':read(P/'c-times.jsonl')}
times={s:{norm(r['where']):r for r in rows} for s,rows in times.items()}
for s in 'ABC': assert times[s].keys()=={w for w,u in units[s].items() if u['status'] in ('attempted','panic')},s+' timing coverage'
assert times['A'].keys()==times['B'].keys()==times['C'].keys()
total={s:{'wall_seconds':sum(r['wall_seconds'] for r in rs.values()),'cpu_seconds':sum(r['cpu_seconds'] for r in rs.values()),'units':len(rs)} for s,rs in times.items()}
save(P/'timing-totals.json',total)
noise={}
def quantile(xs,p): return sorted(xs)[round((len(xs)-1)*p)]
for left,sub in [('A','vs-a'),('B','vs-5e5f57d3')]:
    out=P/sub;rows=[]
    for w,a in times[left].items():
        c=times['C'][w]
        rows.append({'unit':w,'name':names.get(w,''),'kind':units[left][w]['kind'],left+'_seconds':a['wall_seconds'],'C_seconds':c['wall_seconds'],'C_minus_'+left+'_seconds':c['wall_seconds']-a['wall_seconds'],'C_over_'+left:c['wall_seconds']/a['wall_seconds'] if a['wall_seconds'] else None,left+'_cpu_seconds':a['cpu_seconds'],'C_cpu_seconds':c['cpu_seconds']})
    rows.sort(key=lambda r:(-r['C_minus_'+left+'_seconds'],r['unit']))
    save(out/'unit-times.json',rows)
    keys=['unit','name',left+'_seconds','C_seconds','C_minus_'+left+'_seconds','C_over_'+left,left+'_cpu_seconds','C_cpu_seconds']
    (out/'unit-times.tsv').write_text('\t'.join(keys)+'\n'+''.join('\t'.join(str(r[k]) for k in keys)+'\n' for r in rows))
    def tt(rs): return table(['Unit','Name',left+' seconds','C seconds','C−'+left+' seconds','C/'+left],[(r['unit'],r['name'],f"{r[left+'_seconds']:.6f}",f"{r['C_seconds']:.6f}",f"{r['C_minus_'+left+'_seconds']:+.6f}",f"{r['C_over_'+left]:.3f}") for r in rs])
    (out/'timing-top20.md').write_text(tt(rows[:20]))
    controls=[r for r in rows if r['kind']=='KindImportDeclaration'];assert len(controls)==91
    (out/'timing-controls.md').write_text(tt(controls))
    ds=[abs(r['C_minus_'+left+'_seconds']) for r in controls];ratios=[r['C_over_'+left] for r in controls]
    noise[left+'-C']={'selection':'91 import declarations: no lowered receiver read, destructure or predicate; includes registration; C alone versus concurrent original run','median_absolute_seconds':statistics.median(ds),'p95_absolute_seconds':quantile(ds,.95),'max_absolute_seconds':max(ds),'median_ratio':statistics.median(ratios),'p05_ratio':quantile(ratios,.05),'p95_ratio':quantile(ratios,.95),'control_totals':{left:sum(r[left+'_seconds'] for r in controls),'C':sum(r['C_seconds'] for r in controls)}}
save(P/'timing-noise.json',noise)
headline={'refs':REFS,'files':len(expected),'totals':{s:summary[s]['totals'] for s in 'ABC'},'pairs':pair_summaries,'direct_calls':callers['calls'],'caller_units':callers['units'],'timing_totals':total,'timing_noise':noise,'C_record_guard_first_stops':len(guard_stops),'C_record_guard_distinct_sites':len(guard_sites),'timing_design':'A and B simultaneous; C alone; observational differences, not a controlled cost estimate'}
save(P/'headline.json',headline)
print(json.dumps(headline,indent=2))

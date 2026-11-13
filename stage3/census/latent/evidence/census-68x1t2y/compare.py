"""Compare the two complete one-file streams; preserve every finding unchanged."""
import collections, gzip, json, pathlib, re
base=pathlib.Path(__file__).resolve().parent
data={}
for name in ('parent','tip'):
    source=base/name/'census.jsonl'
    raw=source.read_bytes() if source.exists() else gzip.decompress(source.with_suffix('.jsonl.gz').read_bytes())
    rows=[json.loads(line) for line in raw.decode().splitlines()]
    assert rows[0]['latent_mode']=='full'
    assert rows[-1]['status']=='relation_report', 'incomplete census'
    assert rows[-1]['findings']==[], 'relation findings'
    assert rows[-1]['relation_overflow']['after_load']=='' and rows[-1]['relation_overflow']['end']=='', 'relation overflow'
    files=[r for r in rows if 'file' in r]
    assert len(files)==1 and files[0]['file'].endswith('/src/compiler/transformers/destructuring.ts')
    file=files[0]; root=file['file'].removesuffix('/src/compiler/transformers/destructuring.ts')
    def normalize(value):
        if isinstance(value,str): return value.replace(root,'<adapted>')
        if isinstance(value,list): return [normalize(x) for x in value]
        if isinstance(value,dict): return {k:normalize(v) for k,v in value.items()}
        return value
    units={normalize(u['where']):u for u in file['units']}
    findings=collections.defaultdict(list)
    for f in file['findings']: findings[normalize(f['unit'])].append(f)
    keys={(f['kind'],f['where'],f['reason'],f['text']) for f in file['findings']}
    ownkeys={k for k in keys if k[1].startswith(file['file']+':')}
    lowerkeys={(f['kind'],f['where'],f['reason'],f['text']) for f in file['findings'] if f['phase']=='lowering'}
    data[name]=dict(rows=rows,file=file,root=root,units=units,findings=findings,normalized_findings=normalize(dict(findings)),totals=dict(declarations=len(file['declarations']),units=len(units),unit_statuses=dict(collections.Counter(u['status'] for u in units.values())),raw_findings=dict(collections.Counter(f['kind'] for f in file['findings'])),unique_sites=dict(collections.Counter(k[0] for k in keys)),unique_sites_attributed_to_destructuring=dict(collections.Counter(k[0] for k in ownkeys)),unique_lowering_sites=dict(collections.Counter(k[0] for k in lowerkeys))))
def match(f):
    if 'overload' in f.get('reason','').lower(): return True
    for key in ('where','site_where'):
        m=re.search(r'/(debug|destructuring)\.ts:(\d+):',f.get(key,''))
        if m and ((m[1]=='debug' and 244<=int(m[2])<=320) or (m[1]=='destructuring' and 600<=int(m[2])<=616)):return True
    return False
selected=sorted({u for d in data.values() for u,fs in d['findings'].items() if any(match(f) for f in fs)})
common=sorted(set(data['parent']['units']) & set(data['tip']['units']))
def first(d,u):
    fs=d['normalized_findings'].get(u,[])
    return fs[0] if fs else None
changed=[u for u in common if first(data['parent'],u)!=first(data['tip'],u)]
tracked=sorted(set(selected)|{u for u in common if u.endswith(':244:1') or u.endswith(':318:5')})
result=dict(totals={n:d['totals'] for n,d in data.items()},selected_units=[dict(unit=u,**{n:dict(declaration=d['units'].get(u),findings=d['findings'].get(u,[])) for n,d in data.items()}) for u in tracked],selection_rule_units=selected,first_finding_changed=[dict(unit=u,**{n:next(iter(d['findings'].get(u,[])),None) for n,d in data.items()}) for u in changed],units_only_in_parent=sorted(set(data['parent']['units'])-set(data['tip']['units'])),units_only_in_tip=sorted(set(data['tip']['units'])-set(data['parent']['units'])))
result['first_lowering_finding_changed']=[u for u in common if next((f for f in data['parent']['normalized_findings'].get(u,[]) if f['phase']=='lowering'),None)!=next((f for f in data['tip']['normalized_findings'].get(u,[]) if f['phase']=='lowering'),None)]
result['all_findings_changed_units']=[u for u in sorted(set(data['parent']['findings'])|set(data['tip']['findings'])) if data['parent']['normalized_findings'].get(u,[])!=data['tip']['normalized_findings'].get(u,[])]
(base/'comparison.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(dict(totals=result['totals'],selected_units=[(u,{n:d['units'].get(u,{}).get('name') for n,d in data.items()}) for u in tracked],selection_rule_units=selected,first_finding_changed=changed),indent=2))

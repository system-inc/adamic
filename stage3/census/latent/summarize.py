"""Unique finding sites, unit eligibility, and feature deltas from raw measurement events."""
import collections, gzip, hashlib, json, pathlib, shutil, subprocess, sys
scratch=pathlib.Path(sys.argv[1]).resolve(); adapted=pathlib.Path(sys.argv[2]).resolve()
root=pathlib.Path(__file__).resolve().parent; data=root/'data'; data.mkdir(exist_ok=True)
LABEL='measured on a checker-rejected program'
runs=json.loads((scratch/'runs.json').read_text())
manifest=[dict(file=str(p.relative_to(adapted)),sha256=hashlib.sha256(p.read_bytes()).hexdigest(),generated='.generated.' in p.name) for p in sorted((adapted/'src/compiler').rglob('*.ts'))]
def normalize(v):
    if isinstance(v,str): return v.replace(str(adapted)+'/', '')
    if isinstance(v,list): return [normalize(x) for x in v]
    if isinstance(v,dict): return {k:normalize(x) for k,x in v.items()}
    return v

def key(f): return (f['kind'],f['where'],f['reason'],f['text'])
def location_file(where): return where.rsplit(':',2)[0]
def counts(sites):
    c=collections.Counter(k[0] for k in sites)
    return {kind:c[kind] for kind in ['NotYet','Refused','SkippedDependency','error','panic']}
def reasons(sites): return dict(sorted(collections.Counter(k[0]+': '+k[2] for k in sites).items()))
def units_count(units):
    return dict(total=len(units),functions=sum(u['kind']=='KindFunctionDeclaration' for u in units),skipped_checker_body=sum(u['status']=='skipped_checker_body' for u in units),attempted_functions=sum(u['kind']=='KindFunctionDeclaration' and u['status']!='skipped_checker_body' for u in units),attempted_statements=sum(u['kind']!='KindFunctionDeclaration' for u in units),panicked_units=sum(u['status']=='panic' for u in units))
site_sets={}; eligible={}; unit_sites={}
for run in runs:
    name=run['name']; run['measurement']=LABEL
    assert run['status']=='complete',run
    for phase in ['overlay','build','run']:
        shutil.copyfile(scratch/(name+'-'+phase+'.log'),data/(name+'-'+phase+'.log'))
    raw=[normalize(json.loads(line)) for line in (scratch/(name+'.jsonl')).read_text().splitlines()]
    with gzip.open(data/(name+'.jsonl.gz'),'wt') as f:
        for row in raw: f.write(json.dumps(row)+'\n')
    header=raw[0]; assert header['checker_rejected'] and header['status']=='measurement'
    run['checker_rejected']=True; run['checker_total']=len(header['diagnostics'])
    run['checker_per_reason']=dict(sorted(collections.Counter(d.split('error ')[1].split(':')[0] for d in header['diagnostics']).items()))
    assert {row['file'] for row in raw[1:]}=={f['file'] for f in manifest}
    sites={key(f) for row in raw[1:] for f in row['findings']}; site_sets[name]=sites
    all_units=[u for row in raw[1:] for u in row['units']]
    eligible[name]={u['where'] for u in all_units if u['status']!='skipped_checker_body'}
    # Map events to their attempted unit, retaining caller context even if failure lives in a dependency.
    unit_sites[name]=collections.defaultdict(set)
    for row in raw[1:]:
        for f in row['findings']: unit_sites[name][f['unit']].add(key(f))
    run['lowering_counts']=counts(sites);run['unique_sites']=len(sites);run['recorded_events']=sum(len(r['findings']) for r in raw[1:]);run['units']=units_count(all_units)
    run['per_reason']=reasons(sites);run['per_file']={}
    for row in raw[1:]:
        file=row['file']; own={s for s in sites if location_file(s[1])==file}
        run['per_file'][file]=dict(measurement=LABEL,**counts(own),unique_sites=len(own),recorded_events=len(row['findings']),checker=sum(d['file']==file for d in header['diagnostic_sites']),units=units_count(row['units']),per_reason=reasons(own))
    outside={s for s in sites if location_file(s[1]) not in run['per_file']}
    run['outside_roots']=dict(measurement=LABEL,**counts(outside),unique_sites=len(outside),per_reason=reasons(outside))
    run['findings']=[dict(measurement=LABEL,kind=s[0],where=s[1],reason=s[2],text=s[3]) for s in sorted(sites)]
    run['overlay_sha256']={p.name:hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((scratch/(name+'-overlay')).glob('*.go'))}
    run['binary_sha256']=hashlib.sha256((scratch/(name+'-census')).read_bytes()).hexdigest()
base=runs[0]
for run in runs[1:]:
    name=run['name'];removed=site_sets['main']-site_sets[name];added=site_sets[name]-site_sets['main'];common=eligible['main']&eligible[name]
    common_base=set().union(*(unit_sites['main'][u] for u in common)); common_other=set().union(*(unit_sites[name][u] for u in common))
    delta=dict(measurement=LABEL,counts={k:run['lowering_counts'][k]-base['lowering_counts'][k] for k in base['lowering_counts']},removed_sites=len(removed),added_sites=len(added),removed_per_reason=reasons(removed),added_per_reason=reasons(added),checker=run['checker_total']-base['checker_total'],skipped_bodies=run['units']['skipped_checker_body']-base['units']['skipped_checker_body'],newly_eligible_units=len(eligible[name]-eligible['main']),no_longer_eligible_units=len(eligible['main']-eligible[name]),common_eligible_units=len(common),common_eligible_counts_delta={k:counts(common_other)[k]-counts(common_base)[k] for k in base['lowering_counts']},common_eligible_removed_sites=len(common_base-common_other),common_eligible_added_sites=len(common_other-common_base),per_reason={r:run['per_reason'].get(r,0)-base['per_reason'].get(r,0) for r in sorted(run['per_reason'].keys()|base['per_reason'].keys())},per_file={f:dict(measurement=LABEL,**{k:run['per_file'][f][k]-base['per_file'][f][k] for k in base['lowering_counts']}) for f in base['per_file']})
    run['feature_lowering_delta']=delta
base['feature_lowering_delta']=None
preparation_refs=['origin/codex/stage3-base','origin/codex/tsc-census','origin/codex/stage3-type-imports','origin/codex/stage3-optional-declarations']
preparation_shas={ref:subprocess.check_output(['git','rev-parse',ref],cwd=root,text=True).strip() for ref in preparation_refs}
result=dict(preparation_shas=preparation_shas, dependencies=dict(cohere='715ba94f3608a6500086b1076ce5cb7e51b836db',typescript_go='8d550c837c90bd1805b047b7eeccc2baac2d5e7a'), complete=True,measurement=LABEL,count_definition='unique (kind, where, reason, text) sites across all attempts; raw JSONL retains each owning unit and phase',source=dict(typescript='6.0.3',commit='050880ce59e30b356b686bd3144efe24f875ebc8',files=manifest),runs=runs)
(root/'REPORT.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps([dict(name=r['name'],measurement=LABEL,checker=r['checker_total'],counts=r['lowering_counts'],units=r['units'],delta=r['feature_lowering_delta']['counts'] if r['feature_lowering_delta'] else None) for r in runs],indent=2))

"""Independently recount raw evidence, verify body eligibility, and kill report mutants."""
import collections, copy, gzip, hashlib, json, pathlib, sys
root=pathlib.Path(__file__).resolve().parent.parent; adapted=pathlib.Path(sys.argv[1]).resolve()
report=json.loads((root/'REPORT.json').read_text()); LABEL='measured on a checker-rejected program'
raw_by_name={}
for run in report['runs']:
    with gzip.open(root/'data'/(run['name']+'.jsonl.gz'),'rt') as f: raw_by_name[run['name']]=[json.loads(line) for line in f]
def keys(raw): return {(f['kind'],f['where'],f['reason'],f['text']) for row in raw[1:] for f in row['findings']}
def kinds(s):
    c=collections.Counter(k[0] for k in s)
    return {k:c[k] for k in ['NotYet','Refused','SkippedDependency','error','panic']}
def unit_counts(us):
    return dict(total=len(us),functions=sum(u['kind']=='KindFunctionDeclaration' for u in us),skipped_checker_body=sum(u['status']=='skipped_checker_body' for u in us),attempted_functions=sum(u['kind']=='KindFunctionDeclaration' and u['status']!='skipped_checker_body' for u in us),attempted_statements=sum(u['kind']!='KindFunctionDeclaration' for u in us),panicked_units=sum(u['status']=='panic' for u in us))
def check(value):
    expected={str(p.relative_to(adapted)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (adapted/'src/compiler').rglob('*.ts')}
    assert {s['file']:s['sha256'] for s in value['source']['files']}==expected,'source manifest'
    assert value['measurement']==LABEL,'measurement label'
    baseline=value['runs'][0]
    for run in value['runs']:
        assert run['measurement']==LABEL and run['checker_rejected'],'measurement label'
        assert set(run['per_file'])==set(expected),'file coverage'
        raw=raw_by_name[run['name']]; header=raw[0]
        assert header['checker_rejected'] and header['measurement']==LABEL,'measurement label'
        assert len(raw[1:])==len(expected),'file coverage'
        assert run['checker_total']==len(header['diagnostics']),'checker recount'
        assert run['checker_per_reason']==dict(sorted(collections.Counter(d.split('error ')[1].split(':')[0] for d in header['diagnostics']).items())),'checker recount'
        assert len(header['diagnostic_sites'])==len(header['diagnostics']),'checker recount'
        assert collections.Counter(d['text'] for d in header['diagnostic_sites'])==collections.Counter(header['diagnostics']),'checker recount'
        sites=keys(raw)
        assert run['lowering_counts']==kinds(sites) and run['unique_sites']==len(sites),'finding recount'
        assert run['per_reason']==dict(sorted(collections.Counter(k[0]+': '+k[2] for k in sites).items())),'reason recount'
        assert {(f['kind'],f['where'],f['reason'],f['text']) for f in run['findings']}==sites,'finding ledger'
        assert run['recorded_events']==sum(len(row['findings']) for row in raw[1:]),'event recount'
        assert run['units']==unit_counts([u for row in raw[1:] for u in row['units']]),'unit recount'
        for row in raw[1:]:
            assert row['measurement']==LABEL,'measurement label'
            file=row['file']; own={k for k in sites if k[1].rsplit(':',2)[0]==file}; entry=run['per_file'][file]
            assert entry['measurement']==LABEL,'measurement label'
            assert {k:entry[k] for k in ['NotYet','Refused','SkippedDependency','error','panic']}==kinds(own) and entry['unique_sites']==len(own),'per-file recount'
            assert entry['units']==unit_counts(row['units']),'unit recount'
            assert entry['checker']==sum(d['file']==file for d in header['diagnostic_sites']),'per-file recount'
            assert entry['per_reason']==dict(sorted(collections.Counter(k[0]+': '+k[2] for k in own).items())),'per-file reason recount'
            skipped=set()
            for u in row['units']:
                assert u['measurement']==LABEL,'measurement label'
                if 'body_start' not in u: continue
                # Raw positions are independent of the overlay's reported skipped status.
                ds=[d['text'] for d in header['diagnostic_sites'] if d['file']==file and d['start']<u['body_end'] and (d['end']>u['body_start'] or d['start']>=u['body_start'])]
                assert u['checker_diagnostics']==ds,'body eligibility'
                assert (u['status']=='skipped_checker_body')==bool(ds),'body eligibility'
                if ds: skipped.add(u['where'])
            assert not any(f['unit'] in skipped for f in row['findings']),'body eligibility'
            assert all(f['measurement']==LABEL for f in row['findings']),'measurement label'
        outside={k for k in sites if k[1].rsplit(':',2)[0] not in expected}
        assert {k:run['outside_roots'][k] for k in kinds(outside)}==kinds(outside) and run['outside_roots']['unique_sites']==len(outside),'outside-root recount'
        if run is baseline: continue
        delta=run['feature_lowering_delta']; bs=keys(raw_by_name['main'])
        assert delta['measurement']==LABEL,'measurement label'
        assert delta['counts']=={k:run['lowering_counts'][k]-baseline['lowering_counts'][k] for k in baseline['lowering_counts']},'delta recount'
        assert delta['removed_sites']==len(bs-sites) and delta['added_sites']==len(sites-bs),'delta recount'
        assert delta['checker']==run['checker_total']-baseline['checker_total'],'delta recount'
        assert delta['skipped_bodies']==run['units']['skipped_checker_body']-baseline['units']['skipped_checker_body'],'delta recount'
        reasons=run['per_reason'].keys()|baseline['per_reason'].keys()
        assert delta['per_reason']=={k:run['per_reason'].get(k,0)-baseline['per_reason'].get(k,0) for k in sorted(reasons)},'delta recount'
        br=raw_by_name['main']
        be={u['where'] for row in br[1:] for u in row['units'] if u['status']!='skipped_checker_body'}
        oe={u['where'] for row in raw[1:] for u in row['units'] if u['status']!='skipped_checker_body'}
        common=be&oe
        bk={(f['kind'],f['where'],f['reason'],f['text']) for row in br[1:] for f in row['findings'] if f['unit'] in common}
        ok={(f['kind'],f['where'],f['reason'],f['text']) for row in raw[1:] for f in row['findings'] if f['unit'] in common}
        assert delta['common_eligible_units']==len(common) and delta['newly_eligible_units']==len(oe-be) and delta['no_longer_eligible_units']==len(be-oe),'common-unit recount'
        assert delta['common_eligible_counts_delta']=={k:kinds(ok)[k]-kinds(bk)[k] for k in kinds(bk)},'common-unit recount'
        assert delta['common_eligible_removed_sites']==len(bk-ok) and delta['common_eligible_added_sites']==len(ok-bk),'common-unit recount'
        for file in expected:
            assert delta['per_file'][file]['measurement']==LABEL,'measurement label'
            assert {k:delta['per_file'][file][k] for k in baseline['lowering_counts']}=={k:run['per_file'][file][k]-baseline['per_file'][file][k] for k in baseline['lowering_counts']},'delta recount'
check(report)
with gzip.open(root/'data/corpus-mutant.jsonl.gz','rt') as f:
    planted=[json.loads(line) for line in f]
before=raw_by_name['main']
assert planted[0]==before[0]
assert len(planted)==len(before)
for a,b in zip(before[1:],planted[1:]):
    assert a['units']==b['units']
    if a['file']!='src/compiler/binder.ts': assert a==b
    else:
        old={(f['kind'],f['where'],f['reason'],f['text']) for f in a['findings']}
        new={(f['kind'],f['where'],f['reason'],f['text']) for f in b['findings']}
        assert len(new-old)==1 and not old-new
        extra=next(iter(new-old))
        assert extra[0]=='NotYet' and extra[1]=='src/compiler/binder.ts:330:1' and extra[2]=='latent planted extra NotYet'
assert len(keys(planted)-keys(before))==1
print('real-corpus mutant evidence recounted: binder.ts +1, every other file unchanged')
mutants=[
 ('source manifest',lambda r:r['source']['files'][0].__setitem__('sha256','0'*64)),
 ('file coverage',lambda r:r['runs'][0]['per_file'].pop(next(iter(r['runs'][0]['per_file'])))),
 ('checker recount',lambda r:r['runs'][0].__setitem__('checker_total',r['runs'][0]['checker_total']+1)),
 ('finding recount',lambda r:r['runs'][0]['lowering_counts'].__setitem__('NotYet',r['runs'][0]['lowering_counts']['NotYet']+1)),
 ('reason recount',lambda r:r['runs'][0]['per_reason'].__setitem__(next(iter(r['runs'][0]['per_reason'])),-1)),
 ('per-file recount',lambda r:r['runs'][0]['per_file'][next(iter(r['runs'][0]['per_file']))].__setitem__('NotYet',-1)),
 ('unit recount',lambda r:r['runs'][0]['units'].__setitem__('skipped_checker_body',-1)),
 ('measurement label',lambda r:r.__setitem__('measurement','unqualified')),
 ('outside-root recount',lambda r:r['runs'][0]['outside_roots'].__setitem__('NotYet',-1)),
 ('common-unit recount',lambda r:r['runs'][1]['feature_lowering_delta'].__setitem__('common_eligible_units',-1)),
 ('delta recount',lambda r:r['runs'][1]['feature_lowering_delta']['counts'].__setitem__('NotYet',123456)),
]
for name,mutate in mutants:
    value=copy.deepcopy(report); mutate(value)
    try: check(value)
    except AssertionError as failure:
        assert str(failure)==name,(name,failure);print(name+' mutant caught')
    else:raise AssertionError(name+' mutant survived')
# Mutate raw body status itself, without altering diagnostic bytes.
row=next(r for r in raw_by_name['main'][1:] if any(u['status']=='skipped_checker_body' for u in r['units']))
u=next(u for u in row['units'] if u['status']=='skipped_checker_body'); old=u['status'];u['status']='attempted'
changed=copy.deepcopy(report)
for c in [changed['runs'][0]['units'], changed['runs'][0]['per_file'][row['file']]['units']]:
    c['skipped_checker_body']-=1;c['attempted_functions']+=1
try:check(changed)
except AssertionError as failure:
    assert str(failure)=='body eligibility'; print('raw body status mutant caught: '+str(failure))
else:raise AssertionError('raw body status mutant survived')
finally:u['status']=old
print('report evidence audit passed')

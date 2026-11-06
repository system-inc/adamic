"""Independently check the saved piece arithmetic, admission gate and feature results."""
from collections import Counter,defaultdict
import gzip
import hashlib
import json
from pathlib import Path
import subprocess

O=Path(__file__).resolve().parent;E=O/'evidence';P=E/'pieces'
def read(p):return json.loads(p.read_text())
def load(p):
    with gzip.open(p,'rt') as f:return [json.loads(l) for l in f]
def norm(v):
    if isinstance(v,str):return v.replace(str(O/'scratch/adapted')+'/', '')
    if isinstance(v,list):return [norm(x) for x in v]
    if isinstance(v,dict):return {k:norm(x) for k,x in v.items()}
    return v
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def us(rows):return {u['where']:u for r in rows[1:] for u in r['units']}
def lowers(rows):
    out=defaultdict(list)
    for r in rows[1:]:
        for f in r['findings']:
            if f['phase']=='lowering':out[f['unit']].append(f)
    return out
def fk(f):return tuple(f[k] for k in ('kind','where','reason','text'))
def occurrences(rows):return [f for r in rows[1:] for f in r['findings']]
def counts(rows):
    c=Counter(k[0] for k in {fk(f) for f in occurrences(rows)})
    return {k:c[k] for k in ('NotYet','Refused','panic','error')}

d=read(O/'delta.json');proof=read(P/'union-proof.json');plan=read(P/'plan.json');builds=read(P/'builds.json')
assert d['status']=='complete' and d['latent_mode']=='first-error'
assert sha(O.parent/'pair/pair.py')==d['pair_py_sha256']==read(E/'provenance.json')['pair_py_sha256']
whole=norm(load(E/'base-latent.jsonl.gz'));wu=us(whole);wl=lowers(whole)
stock={u['where'] for r in read(E/'stock-units.json') for u in r['units']}
assert stock==set(wu) and len(wu)==4555
assign=Counter(w for p in plan['pieces'] for w in p['expected_units'])
assert assign==Counter({w:1 for w in wu})
streams={};timers={}
for label,side in [('base','before'),('head','after')]:
    allu=[];allt=[];rawfind=[]
    for piece in plan['pieces']:
        folder=P/label/piece['id'];m=read(folder/'metrics.json');c=read(folder/'checkpoint.json');i=m['identity']
        assert m==read(P/'INDEX.json')[label][piece['id']] and m['returncode']==0 and m['failure'] is None
        assert 0<m['process_wall_seconds']<1200
        assert i['label']==label and i['piece']==piece['id']
        assert i['plan_sha256']==sha(P/'plan.json') and i['selection_sha256']==sha(P/(piece['id']+'.selection.json'))
        assert i['binary_sha256']==builds[label]['binary_sha256'] and i['source_manifest_sha256']==sha(E/'source-manifest.json')
        assert (folder/'memory-before.txt').read_bytes()==(folder/'memory-after.txt').read_bytes()
        for kind in ('latent','times'):
            assert hashlib.sha256(gzip.decompress((folder/(kind+'.jsonl.gz')).read_bytes())).hexdigest()==m[kind+'_sha256']==c[kind+'_sha256']
        rows=norm(load(folder/'latent.jsonl.gz'));ts=norm(load(folder/'times.jsonl.gz'))
        assert rows[0]==whole[0] and rows[0]['latent_mode']=='first-error'
        assert Counter(r['file'] for r in rows[1:])==Counter(f['file'] for f in piece['files'])
        owners=[u['where'] for r in rows[1:] for u in r['units']]
        assert Counter(owners)==Counter({w:1 for w in piece['expected_units']})
        assert Counter(t['where'] for t in ts)==Counter({w:1 for w in piece['expected_timed_units']})
        assert all(f['unit'] in owners for f in occurrences(rows) if f['phase']=='lowering')
        allu+=owners;allt+=ts;rawfind+=occurrences(rows)
    rows=load(P/(label+'-union.jsonl.gz'));streams[label]=rows;timers[label]=allt
    assert rows[0]==whole[0] and set(us(rows))==stock and Counter(allu)==assign
    assert len(allt)==len({t['where'] for t in allt})==4056
    assert {fk(f) for f in occurrences(rows)}=={fk(f) for f in rawfind}
    assert counts(rows)==d['counts'][label]
    for kind in ('NotYet','Refused','panic','error'):
        assert d['raw_findings']['totals'][side].get(kind,0)==d['counts'][label][kind]
    for k in ('wall_seconds','cpu_seconds'):assert sum(t[k] for t in allt)==d['unit_time_sums'][label][k]
    pins=read(E/(label+'-actual-pins.json'))
    assert pins=={'.':d[label+'_commit'],'cohere':'06af2e399c6c1fb93fd570b774027942747400b4','cohere/TypeScript':'d92d9bfee114c80be2c375d72edae966176e3a4f'}
    assert not any(read(E/(label+'-post-run-status.json')).values())

bu=us(streams['base']);bl=lowers(streams['base']);hu=us(streams['head']);hl=lowers(streams['head'])
assert all(wl[w]==bl[w] for w in wu)
differences={w for w in wu if wu[w]!=bu[w] or wl[w]!=bl[w]}
assert differences==set(proof['excluded_units'])=={r['where'] for r in proof['unit_differences']}
assert len(differences)==65 and proof['agreeing_units']==4490 and proof['agreeing_timed_units']==4056
for w in differences:
    assert wu[w]['instantiated'] is True and bu[w]['instantiated'] is False
    assert wu[w]['status']=='deferred_generic_instantiation' and bu[w]['status']=='deferred_generic_instantiation'
    assert {k:v for k,v in wu[w].items() if k!='instantiated'}=={k:v for k,v in bu[w].items() if k!='instantiated'}
for w,b in zip(whole[1:],streams['base'][1:]):
    assert {k:v for k,v in w.items() if k not in ('units','findings')}=={k:v for k,v in b.items() if k not in ('units','findings')}
    assert [f for f in w['findings'] if f['phase']!='lowering']==[f for f in b['findings'] if f['phase']!='lowering']
assert proof['first_stop_differences']==0 and not proof['scan_differences']
for label in streams:
    filtered=load(P/(label+'-paired.jsonl.gz'))
    assert set(us(filtered))==stock-differences
    assert all(f['unit'] not in differences for f in occurrences(filtered) if f['phase']=='lowering')
    assert occurrences(filtered)==occurrences(streams[label])
assert counts(whole)==d['whole_base_counts']==d['counts']['base']
# Independently recompute raw first-stop changes, newly failing units and sites.
compact=lambda f:fk(f)[:3] if f else None
first=lambda ll,w:ll[w][0] if ll[w] else None
changed={w for w in stock-differences if compact(first(bl,w))!=compact(first(hl,w))}
assert changed=={r['unit']['where'] for r in d['raw_findings']['unit_first_stop_delta']}
basekeys={fk(f) for f in occurrences(streams['base'])};basewhere={f['where'] for f in occurrences(streams['base'])}
new={fk(f) for f in occurrences(streams['head'])}-basekeys
assert new=={fk(r['finding']) for r in d['raw_findings']['new_findings']}
newscan={fk(f) for f in occurrences(streams['head']) if f['phase']=='refusal_scan' and f['where'] not in basewhere}
reg=d['raw_findings']['regressions']
assert newscan=={fk(r['finding']) for r in reg if r['type']=='previously_finding_free_scan_site'}
assert {w for w in stock-differences if not bl[w] and hl[w]}=={r['unit']['where'] for r in reg if r['type']=='previously_finding_free_unit'}
assert not changed and len(newscan)==1 and not any(f['kind']=='panic' for f in occurrences(streams['head']))
# Check all any-member occurrences, including old NotYet reasons, and lexical owner flags.
afields=('kind','where','reason','text','unit','phase')
expected={tuple(f[k] for k in afields) for f in occurrences(streams['head']) if f['reason'].startswith('a member typed any')}
assert expected=={tuple(r['finding'][k] for k in afields) for r in d['any_member_findings']}
assert len(expected)==12 and d['any_member_unit_first_stops']==0
for r in d['any_member_findings']:
    f=r['finding'];w=r['lexical_owner']['where'];assert r['lexical_owner']==hu[w]
    assert r['lexical_owner_first_lowering_stop']==first(hl,w)
    assert r['is_unit_first_stop']==(f['phase']=='lowering' and first(hl,f['unit'])==f)
    assert r['is_lexical_owner_first_stop']==bool(hl[w] and fk(hl[w][0])==fk(f))
    assert not r['is_lexical_owner_first_stop'] and r['present_on_base']==(fk(f) in basekeys)
assert sum(not r['present_on_base'] for r in d['any_member_findings'])==1
manifest=read(E/'source-manifest.json');tree=O/'scratch/adapted'
assert manifest=={str(p.relative_to(tree)):sha(p) for p in sorted((tree/'src').rglob('*')) if p.is_file()}
for name in ('base-latent.jsonl.gz','base-times.jsonl.gz'):
    original=subprocess.check_output(['git','show','711cca91ed2928b618cb207b1b8152c22eb65ebe:stage3/census/any-members/evidence/'+name],cwd=O)
    assert original==(E/name).read_bytes()
for name,digest in read(O/'artifacts.sha256.json').items():assert sha(O/name)==digest,name
print('PASS: 46 successful checkpoints and hashes; exact disjoint 4555-unit / 4056-timer inventories; source and dependency pins; whole-base streams byte-preserved; base union has only 65 deferred witness changes, zero first-stop or scan differences; exclusions applied symmetrically; 12 any-member scan findings, zero unit first stops; one new scan regression, zero changed first stops or panics; kind counts and timing sums; artifact hashes.')

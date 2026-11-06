"""Audit first-error piece union against the retained whole base before pairing."""
from collections import Counter, defaultdict
import gzip
import hashlib
import json
from pieces_common import OUT, E, TREE, digest, dump, load, norm, validate_piece, verify_partition


def key(value): return json.dumps(value, sort_keys=True)
def units(rows): return {u['where']:u for r in rows[1:] for u in r['units']}
def lower(rows):
    result = defaultdict(list)
    for r in rows[1:]:
        for f in r['findings']:
            if f['phase'] == 'lowering': result[f['unit']].append(f)
    return result

def write_stream(path, rows):
    with gzip.GzipFile(filename=str(path), mode='wb', mtime=0) as f:
        for row in rows: f.write((json.dumps(row)+'\n').encode())

def audit():
    plan = json.loads((E/'plan.json').read_text())
    manifest = json.loads((OUT/'evidence/source-manifest.json').read_text())
    assert manifest == {str(p.relative_to(TREE)):digest(p) for p in sorted((TREE/'src').rglob('*')) if p.is_file()}
    whole = norm(load(OUT/'evidence/base-latent.jsonl.gz'))
    wu = units(whole); wl = lower(whole)
    verify_partition(plan['pieces'], wu)
    builds = json.loads((E/'builds.json').read_text())
    index = json.loads((E/'INDEX.json').read_text())
    unions, times, metrics, repeats = {}, {}, {}, []
    for label in ('base','head'):
        byfile = defaultdict(list); timers = []; metrics[label] = []
        assert set(index[label]) == {p['id'] for p in plan['pieces']}
        for p in plan['pieces']:
            d = E/label/p['id']; m = json.loads((d/'metrics.json').read_text())
            assert m == index[label][p['id']] and m['returncode'] == 0 and m['failure'] is None
            i = m['identity']
            assert i == dict(schema=1,label=label,piece=p['id'],plan_sha256=digest(E/'plan.json'), selection_sha256=digest(E/(p['id']+'.selection.json')),binary_sha256=builds[label]['binary_sha256'],source_manifest_sha256=digest(OUT/'evidence/source-manifest.json'))
            c = json.loads((d/'checkpoint.json').read_text())
            for kind in ('latent','times'):
                raw = gzip.decompress((d/(kind+'.jsonl.gz')).read_bytes())
                assert hashlib.sha256(raw).hexdigest() == m[kind+'_sha256'] == c[kind+'_sha256']
            rows, ts = validate_piece(p,load(d/'latent.jsonl.gz'),load(d/'times.jsonl.gz'),whole[0])
            assert (d/'memory-before.txt').read_bytes() == (d/'memory-after.txt').read_bytes()
            for r in rows[1:]: byfile[r['file']].append(r)
            timers += ts; metrics[label].append(m)
        assert set(byfile) == {r['file'] for r in whole[1:]}
        merged = [whole[0]]
        for original in whole[1:]:
            file = original['file']; chunks = byfile[file]
            metadata = [{k:v for k,v in r.items() if k not in ('units','findings')} for r in chunks]
            assert all(v == metadata[0] for v in metadata), 'repeated file metadata changed'
            scans = [[f for f in r['findings'] if f['phase'] != 'lowering'] for r in chunks]
            assert all(v == scans[0] for v in scans), 'repeated scans differ'
            if len(chunks)>1: repeats.append(dict(label=label,file=file,copies=len(chunks),scan_findings=len(scans[0])))
            uu = {u['where']:u for r in chunks for u in r['units']}
            assert sum(len(r['units']) for r in chunks) == len(uu)
            ll = lower([whole[0]]+chunks)
            us = [uu[u['where']] for u in original['units']]
            merged.append(dict(metadata[0], units=us,findings=scans[0]+[f for u in us for f in ll[u['where']]]))
        assert len(units(merged)) == 4555 and len(timers) == len({t['where'] for t in timers}) == 4056
        assert {t['where'] for t in timers} == {w for w,u in units(merged).items() if u['status'] in ('attempted','panic')}
        unions[label], times[label] = merged, timers
        write_stream(E/(label+'-union.jsonl.gz'),merged)
        write_stream(E/(label+'-union-times.jsonl.gz'),timers)
    bu = units(unions['base']); bl = lower(unions['base']); differences = []
    for where,u in wu.items():
        changes = {k:dict(whole=u.get(k),pieces=bu[where].get(k)) for k in u.keys()|bu[where].keys() if u.get(k)!=bu[where].get(k)}
        if changes or wl[where] != bl[where]:
            witness = set(changes)=={'instantiated'} and wl[where] == bl[where]
            differences.append(dict(where=where,name=u.get('name'),metadata_changes=changes,whole_lowering=wl[where],pieces_lowering=bl[where],first_stop_equal=wl[where][:1]==bl[where][:1],explanation='Only the deferred-generic instantiated witness changed: this owner piece does not run callers in other pieces. The census-wide witness map annotates units after all selected calls finish.' if witness else 'Selected calls changed this unit observation; see exact metadata and ordered lowering findings. Excluded conservatively.',classification='witness_only' if witness else 'lowering_or_metadata'))
    scan_diff = []
    for w,b in zip(whole[1:],unions['base'][1:]):
        assert w['file']==b['file']
        ws=[f for f in w['findings'] if f['phase']!='lowering'];bs=[f for f in b['findings'] if f['phase']!='lowering']
        if ws != bs: scan_diff.append(dict(file=w['file'],whole=ws,pieces=bs))
        assert {k:v for k,v in w.items() if k not in ('units','findings')} == {k:v for k,v in b.items() if k not in ('units','findings')}
    excluded={r['where'] for r in differences}; scan_excluded={r['file'] for r in scan_diff}
    for label,rows in unions.items():
        filtered = [rows[0]]+[dict(r,units=[u for u in r['units'] if u['where'] not in excluded],findings=[f for f in r['findings'] if (f['unit'] not in excluded if f['phase']=='lowering' else r['file'] not in scan_excluded)]) for r in rows[1:]]
        write_stream(E/(label+'-paired.jsonl.gz'),filtered)
    proof=dict(schema=1,mode='first-error',source_files=len(manifest),compiler_files=79,units=4555,timed_units=4056,pieces_per_label=23,successful_children=46,unit_differences=differences,scan_differences=scan_diff,excluded_units=sorted(excluded),agreeing_units=4555-len(excluded),agreeing_timed_units=sum(u['status'] in ('attempted','panic') and w not in excluded for w,u in wu.items()),first_stop_differences=sum(not d['first_stop_equal'] for d in differences),repeated_scans=repeats)
    dump(E/'union-proof.json',proof)
    lines=['# Base piece union against complete whole base','',f"Units: 4,555; exact agreements: {proof['agreeing_units']}; excluded: {len(excluded)}. First-stop differences: {proof['first_stop_differences']}. Scan-context differences: {len(scan_diff)}.",'','Every unit metadata field and ordered lowering finding is compared. Repeated scans are checked byte-for-byte after shared-root normalization, then retained once. No instantiated flags are ORed or masked. All differences below are excluded from both feature streams.','']
    for d in differences:
        lines += [f"- `{d['where']}` ({d['name']}): {json.dumps(d['metadata_changes'])}. First stop equal: {d['first_stop_equal']}. {d['explanation']}"]
        if not d['first_stop_equal'] or d['whole_lowering']!=d['pieces_lowering']:
            lines += [f"  Whole lowering: `{json.dumps(d['whole_lowering'])}`",f"  Pieces lowering: `{json.dumps(d['pieces_lowering'])}`"]
    for d in scan_diff: lines += [f"- Scan `{d['file']}`: `{json.dumps(d)}`"]
    if not differences and not scan_diff: lines += ['No unit or scan differences.']
    (OUT/'UNION.md').write_text('\n'.join(lines)+'\n')
    return proof,unions,times,metrics

if __name__=='__main__':
    proof,*_ = audit()
    print(json.dumps({k:v for k,v in proof.items() if k not in ('unit_differences','repeated_scans','excluded_units','scan_differences')},indent=2))
    print('Difference classifications:',dict(Counter(d['classification'] for d in proof['unit_differences'])))

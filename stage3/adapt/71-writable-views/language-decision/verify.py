#!/usr/bin/env python3
"""Check corpus coverage and immutable source provenance; run destructive mutants only on copies."""
import copy, hashlib, json, pathlib, sys, tempfile, shutil
HERE = pathlib.Path(__file__).resolve().parent
UNIT = HERE.parent

def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def verify(tree, library, result):
    selected = json.loads((UNIT / 'evidence/followup/remaining-selected.json').read_text())
    assert len(selected) == 615, 'input cohort'
    records = result['records']
    assert len(records) == 615, 'coverage: missing or extra retained site'
    assert [(r['id'], r['where'], r['reason']) for r in records] == [(i+1, r['where'], r['reason']) for i,r in enumerate(selected)], 'coverage: site identity or order'
    assert len({(r['where'],r['reason']) for r in records}) == 615, 'coverage: duplicate site'
    counts = {k:0 for k in 'abcd'}
    families, sources = {}, {}
    for r in records:
        k=r['classification']; assert k in counts, 'classification'
        counts[k]+=1
        for out,key in [(families,r['family']), (sources,r['source_family'])]:
            out.setdefault(key, dict(a=0,b=0,c=0,d=0,total=0)); out[key][k]+=1; out[key]['total']+=1
        if k in 'ab': assert not r['unresolved'], 'closed graph must have no unresolved effects'
        if k == 'a': assert not r['writes'], 'reader has a write'
        if k == 'b': assert r['writes'] and all(w['compatible'] is True for w in r['writes']), 'compatible writer'
        if k == 'd': assert r['unresolved'], 'unknown needs a reason'
    assert counts == result['counts'] and families == result['families'] and sources == result['source_families'], 'count totals'
    assert {k:v['total'] for k,v in families.items()} == {'shared-never':82,'other':370,'diagnostics':163}, 'cohort families'
    assert result['c_sites'] == [r['id'] for r in records if r['classification']=='c'], 'all c sites'
    pin = json.loads((UNIT/'evidence/internal-views/source-identity.json').read_text())['sha256']
    assert len(result['source_sha256']) == result['source_files'] == 603, 'source coverage'
    for name, expected in result['source_sha256'].items():
        assert name.startswith('src/'), 'source path'
        assert expected == pin[name[4:]] == sha(tree/name), 'source identity: '+name
    emitted = json.loads((UNIT/'evidence/internal-views/emitted-identity.json').read_text())
    expected = next(r['sha256'] for r in emitted['javascript'] if r['file']=='built/local/typescript.js')
    assert sha(library) == expected, 'built library identity'
    assert sha(library.with_suffix('.d.ts')) == emitted['api_sha256'], 'public API identity'
    return dict(sites=615,counts=counts,source_files=603,library_sha256=expected,public_api_sha256=emitted['api_sha256'],input_sha256=sha(UNIT/'evidence/followup/remaining-selected.json'))

if __name__ == '__main__':
    tree,library=map(pathlib.Path,sys.argv[1:3]); result=json.loads((HERE/'results.json').read_text())
    report=verify(tree,library,result); mutants=[]
    broken=copy.deepcopy(result); broken['records'].pop()
    try: verify(tree,library,broken)
    except AssertionError as e:
        assert 'coverage:' in str(e); mutants.append(dict(name='missing-site',caught_by=str(e)))
    else: raise AssertionError('coverage mutant survived')
    # Copy only the indexed files, mutate one comment, and prove the physical source hash is checked.
    with tempfile.TemporaryDirectory(prefix='unit71-source-mutant-') as tmp:
        target=pathlib.Path(tmp)
        for name in result['source_sha256']:
            dest=target/name; dest.parent.mkdir(parents=True,exist_ok=True); shutil.copyfile(tree/name,dest)
        name=next(iter(result['source_sha256'])); dest=target/name
        with dest.open('a') as f: f.write('\n// source identity mutant\n')
        try: verify(target,library,result)
        except AssertionError as e:
            assert 'source identity:' in str(e); mutants.append(dict(name='source-byte-drift',caught_by=str(e)))
        else: raise AssertionError('source identity mutant survived')
    report['mutants']=mutants; print(json.dumps(report,indent=2))

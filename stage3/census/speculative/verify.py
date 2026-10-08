"""Independent result arithmetic with targeted artifact mutants."""
from collections import Counter
from copy import deepcopy
import json
from pathlib import Path
import sys

raw, result_path, root = map(Path, sys.argv[1:4])
rows = [json.loads(line) for line in raw.read_text().splitlines()]
result = json.loads(result_path.read_text())


def verify(artifact):
    unique = {}
    for row in rows[1:]:
        for finding in row['findings']:
            if finding['kind'] in ('NotYet','Refused') and finding['where'].startswith(str(root)+'/'):
                identity = tuple(finding[k] for k in ('kind','where','reason','text'))
                previous = unique.setdefault(identity, finding)
                assert previous['depth'] == finding['depth'], identity
    for kind in ('NotYet','Refused'):
        expected = [sum(f['kind']==kind and min(4,f['depth'])==depth for f in unique.values()) for depth in range(5)]
        assert artifact['counts'][kind] == expected, (kind,expected)
    reasons = Counter((f['kind'],f['reason']) for f in unique.values())
    ranking = sorted(reasons,key=lambda key:(-reasons[key],key))[:20]
    actual = [(r['kind'],r['root_kind']) for r in artifact['top20']]
    assert actual == ranking, (actual,ranking)
    for item in artifact['top20']:
        expected = [sum(f['kind']==item['kind'] and f['reason']==item['root_kind'] and min(4,f['depth'])==depth for f in unique.values()) for depth in range(5)]
        assert item['counts']==expected,item
    observed = {str(Path(row['file']).relative_to(root)):row['speculative_coverage'] for row in rows[1:]}
    actual = {f['file']:f for f in artifact['files']}
    expected_sources = {str(p.relative_to(root)) for p in root.rglob('*') if p.is_file() and p.suffix in ('.ts','.a')}
    assert set(actual)==set(observed)==expected_sources
    assert all(f['unvisited_nodes']==0 and f['visited_nodes']>=f['total_nodes'] for f in actual.values())
    examined = sum((root/name).stat().st_size for name in expected_sources)
    total = sum(p.stat().st_size for p in root.rglob('*') if p.is_file())
    assert artifact['examined_bytes']==examined and artifact['source_bytes']==total
    assert artifact['examined_share']==examined/total
    assert artifact['typescript_source_bytes']==examined and artifact['typescript_examined_share']==1.0
    return len(unique)


count = verify(result)
for name, mutate in [
    ('headline-depth-total',lambda x:x['counts']['NotYet'].__setitem__(0,x['counts']['NotYet'][0]+1)),
    ('top20-depth-count',lambda x:x['top20'][0]['counts'].__setitem__(0,x['top20'][0]['counts'][0]+1)),
    ('omit-source-file',lambda x:x['files'].pop()),
    ('byte-share',lambda x:x.__setitem__('examined_bytes',x['examined_bytes']+1)),
]:
    mutant=deepcopy(result)
    mutate(mutant)
    try:
        verify(mutant)
    except AssertionError:
        print(name+' artifact mutant caught')
    else:
        raise AssertionError(name+' mutant survived')
print(f'PASS: {count} unique sites; exact depth totals and top20; complete source inventory; independent byte denominator; all four artifact mutants caught')

#!/usr/bin/env python3
"""Join disjoint call observations only when all body proofs agree."""
import copy, json, sys
from pathlib import Path

def merge(reports):
    count = len(reports)
    assert count > 0, 'no shards'
    assert {int(r['shard']) for r in reports} == set(range(count)), 'shard coverage drift'
    assert all(int(r['shards']) == count for r in reports), 'shard count drift'
    first = reports[0]
    result = copy.deepcopy(first)
    result['shard'] = 'merged'
    result['shards'] = count
    for row in result['predicates']:
        row['calls'] = []
    for report in reports:
        assert report['roots'] == first['roots'], 'root drift between shards'
        assert report['matchedCalls'] == first['matchedCalls'], 'call inventory drift between shards'
        assert report['checkerDiagnostics'] == first['checkerDiagnostics'], 'checker diagnostic drift between shards'
        assert len(report['predicates']) == len(result['predicates']), 'body coverage drift between shards'
        for destination, observed in zip(result['predicates'], report['predicates']):
            assert {k:v for k,v in destination.items() if k != 'calls'} == {k:v for k,v in observed.items() if k != 'calls'}, 'body proof drift between shards'
            for call in observed['calls']:
                assert call['ordinal'] % count == int(report['shard']), 'call belongs to another shard'
                destination['calls'].append(copy.deepcopy(call))
    ordinals = [c['ordinal'] for r in result['predicates'] for c in r['calls']]
    assert len(ordinals) == len(set(ordinals)) == result['matchedCalls'] and set(ordinals) == set(range(result['matchedCalls'])), 'call coverage drift'
    for row in result['predicates']:
        row['calls'].sort(key=lambda c:c['ordinal'])
    return result

if __name__ == '__main__':
    reports = [json.loads(Path(p).read_bytes()) for p in sys.argv[2:]]
    result = merge(reports)
    Path(sys.argv[1]).write_text(json.dumps(result, indent=2)+'\n')
    print(json.dumps({'shards': len(reports), 'bodies': len(result['predicates']), 'matchedCalls': result['matchedCalls']}, indent=2))

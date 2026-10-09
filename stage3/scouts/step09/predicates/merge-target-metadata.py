#!/usr/bin/env python3
"""Join independently resolved view targets without changing any proof or call result."""
import copy, json, sys
from pathlib import Path

def join(measured, metadata):
    result = copy.deepcopy(measured)
    targets = {r['where']: r['viewTarget'] for r in metadata}
    calls = [c for p in result['predicates'] for c in p['calls']]
    assert len(targets) == len(metadata) == result['matchedCalls'], 'target metadata coverage drift'
    assert set(targets) == {c['where'] for c in calls}, 'target metadata coordinate drift'
    for call in calls:
        assert type(targets[call['where']]) is bool, 'invalid target metadata'
        call['initialViewTarget'] = call['viewTarget']
        call['viewTarget'] = targets[call['where']]
    result['targetMetadata'] = 'resolved union and intersection members'
    return result

if __name__ == '__main__':
    measured = json.loads(Path(sys.argv[1]).read_bytes())
    metadata = json.loads(Path(sys.argv[2]).read_bytes())
    result = join(measured, metadata)
    Path(sys.argv[3]).write_text(json.dumps(result, indent=2)+'\n')
    print(json.dumps({'calls': result['matchedCalls'], 'newObjectIntersectionTargets': sum(not c['initialViewTarget'] and c['viewTarget'] for p in result['predicates'] for c in p['calls'])}, indent=2))

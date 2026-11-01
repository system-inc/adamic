#!/usr/bin/env python3
"""Compare the meter's entry-root diagnostic observations, including the one-edit mutant."""
import json
from pathlib import Path
import re
import sys

before_path, after_path, mutant_path = map(Path, sys.argv[1:])

def observations(path):
    records = [json.loads(line) for line in path.read_text().splitlines()]
    assert len(records) == 2
    assert all(row['kind'] == 'checker' for row in records)
    def normalize(row):
        return sorted(d.replace(str(Path(row['roots'][0]).parents[2]) + '/', '') for d in row['diagnostics'])
    # The census emits the same single root and the aggregate root. The meter
    # deduplicates observations, rather than adding these two counts.
    assert normalize(records[0]) == normalize(records[1])
    return normalize(records[0])

before, after, mutant = map(observations, (before_path, after_path, mutant_path))
def own(diagnostics):
    return [d for d in diagnostics if re.match(r'src/compiler/checker.ts:\d+:\d+: error TS', d)]
removed = sorted(set(before) - set(after))
assert not set(after) - set(before), 'new whole-program diagnostic'
assert len(before) == 144 and len(after) == 131 and len(mutant) == 132
assert len(own(before)) == 30 and len(own(after)) == 17 and len(own(mutant)) == 18
inventory = json.loads(Path(__file__).with_name('sites.json').read_text())
expected_locations = sorted((site['line'], site['column']) for site in inventory['sites'])
def location(d):
    return tuple(map(int, re.search(r'checker.ts:(\d+):(\d+): error TS', d).groups()))
assert sorted(map(location, removed)) == expected_locations
assert not set(after) - set(mutant)
restored = sorted(set(mutant) - set(after))
assert len(restored) == 1 and location(restored[0]) == (52426, 13)
assert restored[0] in before
print(json.dumps(dict(before=dict(wholeProgram=len(before), checkerOwnFile=len(own(before))),
    after=dict(wholeProgram=len(after), checkerOwnFile=len(own(after))),
    mutant=dict(wholeProgram=len(mutant), checkerOwnFile=len(own(mutant))),
    removed=removed, added=[], mutantRestored=restored), indent=2))

#!/usr/bin/env python3
"""Exact entry-root comparison, using 95a's meter observation deduplication."""
import json
from pathlib import Path
import re
import sys

def observations(path):
    rows = [json.loads(line) for line in Path(path).read_text().splitlines()]
    assert len(rows) == 2 and all(row['kind'] == 'checker' for row in rows)
    def normalized(row):
        return sorted(d.replace(str(Path(row['roots'][0]).parents[2]) + '/', '') for d in row['diagnostics'])
    assert normalized(rows[0]) == normalized(rows[1])
    return normalized(rows[0])

before, after, mutant = map(observations, sys.argv[1:])
def own(rows):
    return [d for d in rows if re.match(r'src/compiler/checker.ts:\d+:\d+: error TS', d)]
def location(d):
    return tuple(map(int, re.search(r'checker.ts:(\d+):(\d+): error TS',d).groups()))
removed = sorted(set(before) - set(after))
assert not set(after) - set(before), 'new diagnostic anywhere'
assert [len(before),len(after),len(mutant)] == [131,128,129]
assert list(map(lambda r: len(own(r)), [before,after,mutant])) == [17,14,15]
assert sorted(map(location,removed)) == [(5311,45),(17890,21),(38430,64)]
assert not set(after) - set(mutant)
restored = sorted(set(mutant) - set(after))
assert len(restored) == 1 and location(restored[0]) == (5311,45) and restored[0] in before
print(json.dumps(dict(before=dict(wholeProgram=len(before),checkerOwnFile=len(own(before))),
    after=dict(wholeProgram=len(after),checkerOwnFile=len(own(after))),
    mutant=dict(wholeProgram=len(mutant),checkerOwnFile=len(own(mutant))),removed=removed,added=[],mutantRestored=restored),indent=2))

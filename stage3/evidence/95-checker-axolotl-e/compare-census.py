#!/usr/bin/env python3
"""Exact 95a entry-root census difference for the ruled public API repair."""
import json
from pathlib import Path
import re
import sys

def observations(path):
    rows=[json.loads(line) for line in Path(path).read_text().splitlines()]
    assert len(rows)==2 and all(row['kind']=='checker' for row in rows)
    def normalized(row):
        return sorted(d.replace(str(Path(row['roots'][0]).parents[2])+'/','') for d in row['diagnostics'])
    assert normalized(rows[0])==normalized(rows[1])
    return normalized(rows[0])
before,after,mutant=map(observations,sys.argv[1:])
def own(rows):
    return [d for d in rows if re.match(r'src/compiler/checker.ts:\d+:\d+: error TS',d)]
removed=sorted(set(before)-set(after))
assert [len(before),len(after),len(mutant)]==[131,130,131]
assert [len(own(r)) for r in [before,after,mutant]]==[17,16,17]
assert not set(after)-set(before),'new whole-program diagnostic'
assert len(removed)==1 and removed[0].startswith('src/compiler/checker.ts:1683:9: error TS2322:')
assert before==mutant,'undo must restore the exact control census'
print(json.dumps(dict(before=dict(wholeProgram=131,checkerOwnFile=17),after=dict(wholeProgram=130,checkerOwnFile=16),mutant=dict(wholeProgram=131,checkerOwnFile=17),removed=removed,added=[],mutantRestored=removed),indent=2))

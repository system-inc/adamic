#!/usr/bin/env python3
"""Compare the real meter entry-root observations, counting duplicate observations once."""
import json
from pathlib import Path
import re
import sys

def read(path):
    rows=[json.loads(line) for line in Path(path).read_text().splitlines()]
    assert len(rows)==2 and all(row['kind']=='checker' for row in rows)
    def normalize(row):
        prefix=str(Path(row['roots'][0]).parents[2])+'/'
        return sorted(d.replace(prefix,'') for d in row['diagnostics'])
    assert normalize(rows[0])==normalize(rows[1])
    return normalize(rows[0])

before,after,mutant=map(read,sys.argv[1:])
expected={(17946,29,2769),(19044,25,2769),(20595,32,2722),(22498,29,2556),(22498,32,2488),(43517,20,2412),(46114,128,2345),(46114,208,2345),(46128,128,2345)}
def key(d):
    m=re.match(r'src/compiler/checker.ts:(\d+):(\d+): error TS(\d+)',d)
    assert m,d
    return tuple(map(int,m.groups()))
def own(ds):return sum(d.startswith('src/compiler/checker.ts:') for d in ds)
removed=sorted(set(before)-set(after));added=sorted(set(after)-set(before))
assert not added and set(map(key,removed))==expected
assert (len(before),own(before),len(after),own(after))==(131,17,122,8)
restored=sorted(set(mutant)-set(after))
assert not set(after)-set(mutant)
assert set(map(key,restored))=={(22498,29,2556),(22498,32,2488)}
assert (len(mutant),own(mutant))==(124,10)
print(json.dumps(dict(before=dict(wholeProgram=131,checkerOwnFile=17),after=dict(wholeProgram=122,checkerOwnFile=8),mutant=dict(wholeProgram=124,checkerOwnFile=10),removed=removed,added=added,mutantRestored=restored),indent=2))

import json
from pathlib import Path
import sys

before_path, after_path, original_proof = map(Path, sys.argv[1:])
def read(path):
    rows=[json.loads(line) for line in path.read_text().splitlines()]
    assert len(rows)==2 and all(r['kind']=='checker' for r in rows)
    def normalize(row):
        prefix=str(Path(row['roots'][0]).parents[2])+'/'
        return sorted(d.replace(prefix,'') for d in row['diagnostics'])
    assert normalize(rows[0])==normalize(rows[1])
    return normalize(rows[0])
before,after=read(before_path),read(after_path)
removed=sorted(set(before)-set(after));added=sorted(set(after)-set(before))
assert removed==json.loads(original_proof.read_text())['removed'] and not added
own=lambda ds:sum(d.startswith('src/compiler/checker.ts:') for d in ds)
assert (len(before),own(before),len(after),own(after))==(67,21,54,8)
print(json.dumps(dict(before=dict(wholeProgram=67,checkerOwnFile=21),after=dict(wholeProgram=54,checkerOwnFile=8),removed=removed,added=added,exactHistorical13Rows=True),indent=2))

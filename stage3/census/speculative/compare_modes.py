"""Compare only completed whole-project full and no-stubs records.
Usage: compare_modes.py FULL_RUN NO_STUBS_RUN OUTPUT_JSON
"""
import json
from pathlib import Path
import sys
full, mutant, output = map(Path,sys.argv[1:4])
a=json.loads((full/'INPUT.json').read_text());b=json.loads((mutant/'INPUT.json').read_text())
assert a['mode']=='full' and b['mode']=='no-stubs'
a.pop('mode');b.pop('mode')
assert a==b,'baseline inputs or limits differ'
def records(root):return {p.name:p for p in (root/'records').glob('*.jsonl')}
x,y=records(full),records(mutant)
common=sorted(x.keys()&y.keys())
assert common,'no completed baseline intersection'
for name in common:assert x[name].read_bytes()==y[name].read_bytes(),name
sample=x[common[0]].read_bytes()
try:assert sample==sample+b'\n'
except AssertionError:pass
else:raise AssertionError('baseline byte mutant survived')
result=dict(full_completed=len(x),no_stubs_completed=len(y),input_files=len(a['files']),
            common_completed=len(common),byte_identical=common,full_only=sorted(x.keys()-y.keys()),
            no_stubs_only=sorted(y.keys()-x.keys()),byte_mutant='caught')
output.write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result))

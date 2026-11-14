import json,pathlib
s=pathlib.Path('/workspace/scratch/path-mint-618/evidence')
a=json.loads((s/'base-stock/javascript.json').read_text());b=json.loads((s/'after-stock/javascript.json').read_text())
r={'baseFiles':len(a),'afterFiles':len(b),'new':sorted(b.keys()-a.keys()),'gone':sorted(a.keys()-b.keys()),'changed':[k for k in sorted(a.keys()&b.keys()) if a[k]!=b[k]]}
(s/'javascript-comparison.json').write_text(json.dumps(r,indent=2)+'\n');print(r);assert a==b

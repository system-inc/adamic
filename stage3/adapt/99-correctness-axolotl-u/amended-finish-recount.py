from pathlib import Path
from collections import Counter
import json,gzip,shutil
s=Path('/workspace/scratch');p=Path('/workspace/adamic/stage3/adapt/99-correctness-axolotl-u')
report=json.loads((s/'99u2-recount-oracle/report.json').read_text());assert report['counts']=={'passing':106366,'failing':1,'pending':0},report['counts'];assert report['baseline_diffs']==['api/typescript.d.ts'],report['baseline_diffs']
rows=[json.loads(line) for line in (s/'99u2-recount-counts.jsonl').read_text().splitlines()];assert all(r['id']=='G11' for r in rows)
assert sum(r['runs'] for r in rows)==26;assert sum(r['wouldFail'] for r in rows)==12
failures=[json.loads(line) for line in (s/'99u2-recount-failures.jsonl').read_text().splitlines()];assert len(failures)==12
old=json.loads((p/'default-expression-failures.json').read_text());print('old failure artifact keys',list(old) if isinstance(old,dict) else 'list')
# Compare against the retained raw earlier observations, ignoring process identifiers.
with gzip.open(p/'default-expression-failures.jsonl.gz','rt') as f: prior=[json.loads(line) for line in f]
key=lambda r:(r['title'],r['kind'],r['numericKind'])
assert Counter(map(key,failures))==Counter(map(key,prior))
assert Counter(r['kind'] for r in failures)=={'MethodSignature':6,'TypeReference':6}
summary={'sourceBase':'31d7a244808fa8092643926d21ab9d310a54cce3','runs':26,'nonExpressions':12,'kinds':dict(Counter(r['kind'] for r in failures)),'titleAndKindMultisetMatchesOriginal99u':True,'failures':failures}
(p/'amended-default-recount.json').write_text(json.dumps(summary,indent=2)+'\n')
for source,dest in [('99u2-recount-counts.jsonl','amended-default-counters.jsonl.gz'),('99u2-recount-failures.jsonl','amended-default-failures.jsonl.gz'),('99u2-recount-identities.jsonl','amended-recount-identities.jsonl.gz')]:
 with gzip.open(p/dest,'wb') as f:f.write((s/source).read_bytes())
shutil.copyfile(s/'99u2-recount-oracle/report.json',p/'amended-recount-oracle.json')
with gzip.open(p/'amended-recount-tests.log.gz','wb') as f:f.write((s/'99u2-recount-oracle/tests.log').read_bytes())
print('Verified amended recount: 26 / 12, six MethodSignature and six TypeReference, same title/kind multiset')

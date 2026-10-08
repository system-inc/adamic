#!/usr/bin/env python3
"""Compare raw recorded stdout, stderr and exit codes without normalization."""
import base64,collections,gzip,json,pathlib,sys

def read(path):
 path=pathlib.Path(path)
 if path.is_dir(): rows=[json.loads(p.read_text()) for p in path.glob('*.json')]
 else:
  opener=gzip.open if path.suffix=='.gz' else open
  with opener(path,'rt') as handle:rows=[json.loads(line) for line in handle]
 result={(r['Test'],r['Number']):{field:(base64.b64decode(value or '') if field in ('Stdout','Stderr') else value) for field,value in r['Run'].items()} for r in rows}
 assert len(result)==len(rows), 'duplicate execution key'
 return result

before,after=map(read,sys.argv[1:3])
missing=sorted(set(before)-set(after));added=sorted(set(after)-set(before))
mismatch=[{'test':k[0],'number':k[1],'fields':[f for f in before[k] if before[k][f]!=after[k][f]]} for k in sorted(set(before)&set(after)) if before[k]!=after[k]]
main=lambda k:k[0].startswith('TestNativeAgreesWithNode/')
summary={'before_executions':len(before),'after_executions':len(after),'identical_executions':sum(before[k]==after[k] for k in before.keys()&after.keys()),'missing':missing,'added':added,'mismatches':mismatch,'main_executions':sum(map(main,before)),'main_tests':len({k[0] for k in before if main(k)}),'main_missing':[k for k in missing if main(k)],'main_added':[k for k in added if main(k)],'main_mismatches':[m for m in mismatch if main((m['test'],m['number']))]}
print(json.dumps(summary,indent=2))
if missing or added or mismatch:sys.exit(1)

import gzip,json
from pathlib import Path
root=Path(__file__).resolve().parent
memo=root/'memo'
def read(path):
 if not path.exists():path=path.with_suffix(path.suffix+'.gz')
 data=gzip.decompress(path.read_bytes()) if path.suffix=='.gz' else path.read_bytes()
 return [json.loads(line) for line in data.decode().splitlines()]
queries=json.loads((root/'queries.json').read_text());results=[]
for q in queries:
 for side in ['main','stack']:
  name=side+'-'+q['label'];status=json.loads((memo/(name+'.status.json')).read_text());assert status['exit']==0,name
  old=root/(name+'.jsonl')
  if side=='stack' and q['label']=='hidden-08':continue
  assert read(old)==read(memo/(name+'.jsonl')),name
  results.append(name)
assert len(results)==19
(root/'memo-equivalence.json').write_text(json.dumps({'complete_reference_ledgers':len(results),'identical':results,'newly_completed':'stack-hidden-08','basis':'Exact equality of every parsed JSON record, including diagnostic signatures, unit order, findings and coverage; same instrumentation on main and stack.'},indent=2)+'\n')
print('PASS: all 19 complete original ledgers identical; 20 memoized ledgers complete')

from pathlib import Path
import hashlib,json,subprocess
out=Path(__file__).resolve().parent
rows=[];raw_all=b''
for group,size in enumerate([20,20,20,19],1):
 r=json.loads((out/f'group-{group}-rows.json').read_text());data=(out/f'group-{group}.jsonl').read_bytes();raw=[json.loads(x) for x in data.splitlines()]
 assert len(r)==size and len(raw)==2*size
 for i,row in enumerate(r):
  assert row==raw[2*i]
  assert {k:v for k,v in row.items() if k!='seconds'}=={k:v for k,v in raw[2*i+1].items() if k!='seconds'}
 rows+=r;raw_all+=data
roots=json.loads((out/'roots.json').read_text())
assert [r['roots'][0] for r in rows]==roots and len(set(roots))==79
(out/'census.jsonl').write_bytes(raw_all)
cp=json.loads((out/'CHECKPOINT.json').read_text())
assert hashlib.sha256((out/'adapted-product.manifest.json').read_bytes()).hexdigest()==cp['manifest_sha256']
paths=subprocess.check_output(['git','diff','--name-only','origin/main'],text=True).splitlines()
assert all(p.startswith('stage3/census/next-stops-37f8314b/') for p in paths)
assert not list(out.rglob('*.go'))
for p in out.rglob('*'):
 if p.is_file() and not p.name.endswith('.gz'):
  try:text=p.read_text()
  except UnicodeDecodeError:continue
  assert '\u2013' not in text and '\u2014' not in text,str(p)
artifacts=json.loads((out/'root-order.json').read_text())
result=dict(ordinary_roots=79,unique_roots=79,raw_rows=158,group_sizes=[20,20,20,19],raw_pairs_agree=True,root_order_matches=True,manifest_hash_matches=True,scope_report_directory_only=True,artifact_status=artifacts['status'])
(out/'validation.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result,indent=2))

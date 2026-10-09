"""Retain parser driver selection and verify the pinned input bytes."""
import hashlib,json,sys
from pathlib import Path
repo,tree,checkout,out=map(Path,sys.argv[1:]);reference=json.loads((repo/'stage3/drivers/parser/cases-reference.json').read_text());rows=[]
for p in sorted((tree/'src/compiler').rglob('*')):
 if p.is_file():rows.append(dict(name=p.relative_to(tree/'src/compiler').as_posix(),group='compiler',text=p.read_text(),raw_sha256=hashlib.sha256(p.read_bytes()).hexdigest()))
for r in reference['cases']:
 raw=(checkout/r['path']).read_bytes();assert hashlib.sha256(raw).hexdigest()==r['source_sha256']
 text=raw[2:].decode('utf-16-be','replace') if raw.startswith(b'\xfe\xff') else raw[2:].decode('utf-16-le','replace') if raw.startswith(b'\xff\xfe') else raw.decode('utf-8-sig','replace')
 assert hashlib.sha256(text.encode()).hexdigest()==r['input_sha256']
 rows.append(dict(name=r['input_path'],group=r['bucket'],text=text,raw_sha256=r['source_sha256'],reference_nodes=r['node']['node_count']))
for name,text in json.loads((repo/'stage3/drivers/parser/jsdoc-inputs.json').read_text()).items():rows.append(dict(name=name,group='directed',text=text))
out.write_text(json.dumps(rows));print('PASS inputs:',len(rows),'cases:',len(reference['cases']))

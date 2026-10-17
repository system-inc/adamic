import json,pathlib
root=pathlib.Path(__file__).resolve().parent
owners={}
for path in root.glob('recut-TestProduct_NativeRuntimes*.jsonl'):
 for line in path.read_text().splitlines():
  row=json.loads(line)
  if row.get('name')=='native-runtime' and row['outcome']=='settled':
   assert row['nameKey'] not in owners,('two owner settlements',row['nameKey'])
   owners[row['nameKey']]=(path.name,row['key'])
for test in ['Sanitized','WASI']:
 path=root/f'recut-TestProduct_CommandRuntime{test}.jsonl'
 rows=[json.loads(line) for line in path.read_text().splitlines()]
 runtime=[row for row in rows if row.get('name')=='native-runtime']
 assert len(runtime)==1,(path,runtime)
 row=runtime[0]
 assert row['outcome']=='found',(path,row)
 assert row['nameKey'] in owners and owners[row['nameKey']][1]==row['key'],(path,row)
 print(f"{test}: found owner {owners[row['nameKey']][0]} nameKey={row['nameKey']} key={row['key']}; zero builds")
path=root/'recut-final-counted.jsonl'
rows=[json.loads(line) for line in path.read_text().splitlines()]
rows=[row for row in rows if row.get('name')=='native-runtime']
assert len(rows)==1 and rows[0]['outcome']=='settled',rows
row=rows[0]
assert row['nameKey'] not in owners,row
print(f"WASICounted: sole builder cmd/adamic nameKey={row['nameKey']} key={row['key']}")
print(f"PASS: {len(owners)+1} observed runtime keys each have one builder; both shared command products find their owner's exact key")

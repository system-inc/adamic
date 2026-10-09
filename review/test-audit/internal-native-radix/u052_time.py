import re,json,subprocess,time,pathlib
p=pathlib.Path('review/test-audit/internal-native-radix')
files=['radix','record','regexp_search','regexp','runtime_profile']
names=[]
for f in files:
 for n in re.findall(r'^func (Test\w+)\(',pathlib.Path(f'internal/native/{f}_test.go').read_text(),re.M):names.append(n)
rows=[]
for n in names:
 if n.startswith('TestRecordMutantsUnit') or n.startswith('TestRegExpBytecodeRandomNodeUnit'):continue
 rows.append((n,[n]))
for prefix in ['TestRecordMutants','TestRegExpBytecodeRandomNode']:
 rows.append((prefix+' family',[n for n in names if n.startswith(prefix+'Unit')]))
(p/'scope.json').write_text(json.dumps(rows,indent=2))
listed=pathlib.Path('/tmp/u052-list.log').read_text().splitlines()
assert all(n in listed for n in names)
for row,members in rows:
 results=[]
 for i in range(3):
  log=p/f'time-{row.replace(" ","-")}-{i+1}.log'
  pattern='^('+'|'.join(members)+')$'
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run',pattern]
  with log.open('w') as out:r=subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT)
  events=[]
  for line in log.read_text().splitlines():
   try:events.append(json.loads(line))
   except:pass
  terminal=[e for e in events if e.get('Action') in ['pass','fail'] and 'Test' not in e]
  results.append({'exit':r.returncode,'seconds':terminal[-1].get('Elapsed') if terminal else None,'command':'ADAMIC_RECORD_BENCH=1 '+' '.join(cmd)})
 (p/f'timing-{row.replace(" ","-")}.json').write_text(json.dumps(results,indent=2))

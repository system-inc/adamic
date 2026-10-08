import json,pathlib,statistics,sys
p=pathlib.Path(sys.argv[1]);builds=json.loads((p/'builds.json').read_text());frontend=json.loads((p/'end-to-end.json').read_text());result=[]
for version in ('before','after'):
 times={}
 for round in range(3):
  for line in (p/f'{round}-{version}-sanitized.jsonl').read_text().splitlines():
   row=json.loads(line);assert row['exit']==0;a=row['args']
   if '-c' not in a:continue
   n=a[a.index('-c')+1]
   if pathlib.Path(n).name != n:continue
   times.setdefault(n,[]).append(row['seconds'])
 rows=[b['Seconds'] for b in builds if b['Source']==version]
 full=[b['Seconds'] for b in frontend if b['Source']==version]
 result.append({'source':version,'c_to_native_median':statistics.median(rows),'c_to_native_runs':rows,'full_build_median':statistics.median(full),'full_build_runs':full,'width_table_units':{n:statistics.median(v) for n,v in sorted(times.items()) if 'widthTables' in n},'all_units':len(times),'unit_runs':times})
(p/'summary.json').write_text(json.dumps(result,indent=2)+'\n')
for row in result:print(row['source'],row['c_to_native_median'],row['full_build_median'],row['all_units'])

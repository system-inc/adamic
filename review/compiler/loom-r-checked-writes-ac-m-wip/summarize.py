import pathlib,json
out=pathlib.Path(__file__).parent
selected=[('internal/ir','package-02.log'),('internal/fresh','package-03.log'),('internal/javascript','package-04.log'),('internal/flow','package-05.log'),('internal/regexp','package-06.log'),('internal/lower','lower-final.log'),('internal/native','native-green.log'),('internal/native/runtimeproducts','runtimeproducts-local-cache.log'),('cmd/adamic','remaining-final-00.log'),('stage1/cohere/values','remaining-green-00.log'),('stage3/fixtures','remaining-green-01.log')]
rows=[]
for package,log in selected:
 events=[];results=[]
 for line in(out/log).read_text().splitlines():
  try:e=json.loads(line)
  except ValueError:continue
  if package=='internal/native' and e.get('Package','').endswith('/runtimeproducts'):continue
  if e.get('Action')in('pass','fail','skip'):
   if e.get('Test'):events.append(e)
   else:results.append(e)
 names={(e['Package'],e['Test'])for e in events}
 leaves=[e for e in events if not any(p==e['Package']and n.startswith(e['Test']+'/')for p,n in names)]
 assert results and all(e['Action']=='pass'for e in results),(package,results)
 assert all(e['Action']!='fail'for e in events),package
 row=dict(package=package,log=log,packages=[dict(name=e['Package'],seconds=e.get('Elapsed',0))for e in results],test_events=len(events),leaf_count=len(leaves),max_leaf_seconds=max((e.get('Elapsed',0)for e in leaves),default=0),skips=[e['Test']for e in events if e['Action']=='skip'],over_budget=[dict(test=e['Test'],seconds=e.get('Elapsed'))for e in leaves if e.get('Elapsed',0)>=60])
 rows.append(row)
coverage=json.loads((out/'oracle-coverage.json').read_text());assert coverage['all_other_roots_pass']
batches=json.loads((out/'oracle-remainder-results.json').read_text());parts=json.loads((out/'oracle-remainder.json').read_text());kept=set(parts['completed']);oracle_events={};oracle_packages=[]
for log,restrict in [('oracle-final.log',True)]+[(r['log'],False)for r in batches]:
 for line in(out/log).read_text().splitlines():
  try:e=json.loads(line)
  except:continue
  if e.get('Action')not in('pass','fail','skip'):continue
  if e.get('Test'):
   if restrict and e['Test'].split('/')[0]not in kept:continue
   oracle_events[e['Test']]=e
  elif not restrict:
   assert e['Action']=='pass';oracle_packages.append(dict(name=e['Package'],seconds=e.get('Elapsed',0)))
assert all(e['Action']!='fail'for e in oracle_events.values())
names=set(oracle_events);leaves=[e for n,e in oracle_events.items()if not any(child.startswith(n+'/')for child in names)]
rows.append(dict(package='internal/oracle',log='oracle-final.log + oracle-remainder-*.log',packages=oracle_packages,test_events=len(oracle_events),leaf_count=len(leaves),max_leaf_seconds=max(e.get('Elapsed',0)for e in leaves),skips=[n for n,e in oracle_events.items()if e['Action']=='skip'],over_budget=[dict(test=e['Test'],seconds=e.get('Elapsed'))for e in leaves if e.get('Elapsed',0)>=60],coverage=coverage,initial_aggregate_timeout_seconds=570.103,status='927 of 928 roots completed; excluded over-budget probe was not rerun'))
(out/'final-package-results.json').write_text(json.dumps(rows,indent=2)+'\n')
text='| Package (recursive) | Package seconds | Test events | Slowest leaf seconds | Log |\n|---|---:|---:|---:|---|\n'
for r in rows:text+='| '+r['package']+' | '+', '.join(str(p['seconds'])for p in r['packages'])+' | '+str(r['test_events'])+' | '+str(r['max_leaf_seconds'])+' | '+r['log']+' |\n'
(out/'final-package-results.md').write_text(text)
print(text)

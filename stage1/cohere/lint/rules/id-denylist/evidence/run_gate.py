import glob,json,os,statistics,subprocess,time
from pathlib import Path
root=Path('stage1/cohere/lint/rules/id-denylist/evidence')
before=set(glob.glob('/tmp/adamic-gate/lint-shared-*'))
seen=set(); records={}; samples=[]; start=time.monotonic()
with (root/'lint-package.jsonl').open('w') as log:
 p=subprocess.Popen(['go','test','-json','-count=1','-timeout=3h','./stage1/cohere/lint'],stdout=log,stderr=log)
 while p.poll() is None:
  samples.append({'elapsed':time.monotonic()-start,'load':os.getloadavg()})
  for directory in set(glob.glob('/tmp/adamic-gate/lint-shared-*'))-before:
   for filename in glob.glob(directory+'/upstream-*/capture/*.jsonl'):
    path=Path(filename)
    try:
     stamp=(filename,path.stat().st_mtime_ns,path.stat().st_size)
     if stamp in seen: continue
     seen.add(stamp)
     for line in path.read_text().splitlines():
      try: record=json.loads(line)
      except json.JSONDecodeError: continue
      if record.get('rule') not in ('id-denylist','no-shadow-restricted-names'): continue
      records[json.dumps(record,sort_keys=True)]=record
    except FileNotFoundError: pass
  if records:
   (root/'upstream-capture.jsonl').write_text(''.join(json.dumps(v,sort_keys=True)+'\n' for v in records.values()))
  time.sleep(5)
wall=time.monotonic()-start
counts={'pass':0,'fail':0,'skip':0}; top=counts.copy(); failures=[]; skipped=[]
for line in (root/'lint-package.jsonl').read_text().splitlines():
 try: record=json.loads(line)
 except json.JSONDecodeError: continue
 action=record.get('Action'); name=record.get('Test')
 if name and action in counts:
  counts[action]+=1
  if '/' not in name: top[action]+=1
  if action=='fail': failures.append(name)
  if action=='skip': skipped.append(name)
unique={}
for record in records.values():
 key=json.dumps([record.get('rule'),record.get('file'),record.get('source'),record.get('options')],sort_keys=True)
 unique[key]=record
by_rule={name:sum(r.get('rule')==name for r in unique.values()) for name in ['id-denylist','no-shadow-restricted-names']}
result={'exit':p.returncode,'wall':wall,'nproc':len(os.sched_getaffinity(0)),'counts':counts,'top_level':top,'failures':failures,'skipped':skipped,'captured_unique':by_rule,'load_min':min(s['load'][0] for s in samples),'load_median':statistics.median(s['load'][0] for s in samples),'load_max':max(s['load'][0] for s in samples)}
(root/'gate-result.json').write_text(json.dumps(result,indent=2)+'\n')
(root/'load.json').write_text(json.dumps(samples)+'\n')
print(json.dumps(result))

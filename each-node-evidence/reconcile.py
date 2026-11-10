from pathlib import Path
import json,hashlib,collections,tarfile
s=Path('/workspace/scratch/each-node');repo=Path('/workspace/adamic');plan=json.loads((repo/'stage3/lane/shard-plan.json').read_text());summary=json.loads((s/'lane-summary.json').read_text());assert summary['complete'];oracle=collections.Counter();totals=collections.Counter();sources={};processes=0;units=[]
for item in plan['units']:
 name=item['name'];record=json.loads((s/(name+'-result.json')).read_text());report=json.loads((s/'lane'/name/'report.json').read_text());assert report==record['report'];assert record['exit']==0 and report['status']=='pass' and not report['errors'];rows=[json.loads(p.read_text()) for p in sorted((s/'counts'/name).glob('*.json'))];assert rows==record['processes'] and rows;units.append(name);oracle.update(report['counts'])
 for row in rows:
  processes+=1
  for k in ['defined','exposed','throws']:totals[k]+=row[k]
  assert sum(f['calls'] for f in row['functions'])==row['defined'];assert sum(f['throws'] for f in row['functions'])==row['throws'];assert sum(f['calls'] for f in row['functions'] if f['canSeeSecond'])==row['exposed']
  for f in row['functions']:
   x=sources.setdefault(f['source'],dict(f,calls=0,throws=0));assert x['parameters']==f['parameters'] and x['canSeeSecond']==f['canSeeSecond'];x['calls']+=f['calls'];x['throws']+=f['throws']
assert len(units)==len(set(units))==len(plan['units']);assert dict(oracle)==summary['oracle']==plan['counts'];assert dict(totals)==summary['totals'];assert sorted(sources.values(),key=lambda f:f['source'])==sorted(summary['functions'],key=lambda f:f['source']);assert totals['exposed']==0
result=dict(passed=True,acceptedUnitsExactlyOnce=len(units),planSha256=hashlib.sha256((repo/'stage3/lane/shard-plan.json').read_bytes()).hexdigest(),nodeProcesses=processes,oracle=dict(oracle),totals=dict(totals),distinctCallbackSources=len(sources),allSourceClassificationsNonSensitive=True)
(s/'reconciled.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))
# Retain the raw identity, API, baseline and count guards, not just summaries.
for start in range(0,len(units),20):
 archive=s/f'lane-raw-{start//20+1:02}.tar.gz'
 with tarfile.open(archive,'w:gz') as tf:
  for name in units[start:start+20]:tf.add(s/'lane'/name,arcname=name,recursive=True)
 assert archive.stat().st_size<20_000_000,(archive,archive.stat().st_size)

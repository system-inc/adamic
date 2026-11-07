import json,sys,hashlib,tarfile
from pathlib import Path
from collections import Counter
s=Path('/workspace/scratch/levenshtein');repo=Path('/workspace/adamic');sys.path.insert(0,str(repo/'stage3/lane'))
from shard_plan import load_plan,names_digest
plan=load_plan();summary=json.loads((s/'COUNT.json').read_text());assert summary['complete'] and summary['differences']==0
assert summary['completedUnits']==[u['name'] for u in plan['units']]
classes=Counter();calls=0;diffs=0;records=[];totals=Counter();rawdiff=b''
for piece in summary['pieces']:
 directory=s/f'count-{piece["piece"]:02d}'
 # Reopen the retained archive and verify every member against its internal manifest.
 with tarfile.open(s/(directory.name+'.tar.gz')) as tar:
  prefix=directory.name+'/'
  manifest=tar.extractfile(prefix+'SHA256SUMS').read().decode()
  for line in manifest.splitlines():
   digest,name=line.split('  ',1);assert hashlib.sha256(tar.extractfile(prefix+name).read()).hexdigest()==digest
  for member in tar.getmembers():
   if '/calls/' not in member.name or not member.isfile():continue
   for line in tar.extractfile(member):
    r=json.loads(line);assert r['equal'] and r['original']==r['adapted'];calls+=1
    expected={'equalStrings':r['s1']==r['s2'],'emptyS1':not r['s1'],'emptyS2':not r['s2'],'caseOnly':r['s1']!=r['s2'] and r['s1'].lower()==r['s2'].lower(),'giveUp':r['original']=='undefined','maxBelowOne':r['max']<1}
    for key,value in expected.items():assert r[key]==value;classes[key]+=value
for unit in plan['units']:
 directory=s/'count-results'/unit['name'];report=json.loads((directory/'report.json').read_text());assert report['status']=='pass' and report['counts']==unit['expect'];observed=json.loads((directory/'observations.json').read_text());assert names_digest(observed)==unit['names_digest'];records+=observed
 totals.update(report['counts']);rawdiff+=(directory/'oracle/baseline.diff').read_bytes()
whole=[dict(record,task=row['task'],status=status) for row in map(json.loads,(s/'whole-tasks.jsonl').read_text().splitlines()) for status,key in [('pass','passes'),('fail','errors')] for record in row[key]]
assert names_digest(records)==names_digest(whole)
assert rawdiff==(s/'whole/oracle/baseline.diff').read_bytes()
assert dict(totals)==summary['counts']==plan['counts'];assert dict(classes)==summary['classes'];assert calls==summary['calls']
result={'complete':True,'calls':calls,'differences':0,'classes':dict(classes),'counts':dict(totals),'units':len(plan['units']),'wholeLaneIdentitiesMatch':True,'wholeLaneBaselineByteIdentical':True,'allArchivedHashesVerified':True,'namesDigest':names_digest(records)}
(s/'reconciled.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)

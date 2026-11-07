import hashlib,json,tarfile
from collections import Counter
from pathlib import Path
def digest(value):return hashlib.sha256(json.dumps(value,sort_keys=True,separators=(',',':')).encode()).hexdigest()
out=Path(__file__).resolve().parent;repo=out.parents[2];plan=json.loads((repo/'stage3/lane/shard-plan.json').read_text());summary=json.loads((out/'lane-summary.json').read_text());byName={u['name']:u for u in plan['units']};seen=[];counts=Counter();calls=reads=diff=tdz=uninitialized=exportReads=exportUninitialized=exportTdz=0;kinds=Counter();exportSites=Counter();offsetClasses=Counter();returnKinds=Counter()
for piece in summary['pieces']:
 num=piece['piece'];archive=out/f'lane-{num:02d}.tar.gz';assert hashlib.sha256(archive.read_bytes()).hexdigest()==(out/f'lane-{num:02d}.sha256').read_text().split()[0]
 with tarfile.open(archive) as tar:
  prefix=f'piece-{num:02d}/'
  manifest=tar.extractfile(prefix+'SHA256SUMS').read().decode().splitlines()
  for line in manifest:
   expectedHash,name=line.split('  ',1);assert hashlib.sha256(tar.extractfile(prefix+name).read()).hexdigest()==expectedHash,name
  record=json.load(tar.extractfile(prefix+'piece.json'));assert record['sourceFrozen']
  for child in record['children']:
   assert child['exit']==0 and child['report']['status']=='pass',child
   name=child['name'];unit=byName[name];report=child['report']
   assert report['name']==name and report['counts']==unit['expect'] and report['baseline_diffs']==unit['baseline_diffs'] and report['errors']==[]
   assert report['artifact']=='stage3-adapted-tree' and report['inputs_hash']==digest(dict(artifact='stage3-adapted-tree',api=plan['api_digest'],runtime=plan['runtime'],unit=unit))
   observations=json.load(tar.extractfile(prefix+name+'/observations.json'))
   titles=Counter((r['task']['runner'],r['task']['file'],' '.join(part for part in r['name'] if part)) for r in observations)
   assert digest(sorted((*key,count) for key,count in titles.items()))==unit['names_digest']
   actual={'pass':sum(r['status']=='pass' for r in observations),'fail':sum(r['status']=='fail' for r in observations),'skip':sum(r['status']=='pending' for r in observations)}
   assert actual==report['counts'];seen.append(name);counts.update(report['counts'])
  for member in tar.getmembers():
   if '/calls/' not in member.name or not member.isfile():continue
   for line in tar.extractfile(member):
    r=json.loads(line)
    if member.name.endswith('.exports.jsonl'):
     exportReads+=1;exportUninitialized+=not r['initialized'];exportTdz+=r['tdz'];exportSites[r['site']]+=1
    elif member.name.endswith('.reads.jsonl'):reads+=1;tdz+=r['tdz'];uninitialized+=not r['initialized'];kinds[r['kind']]+=1
    else:
     calls+=1;diff+=not r['equal'];returnKinds[r['kind']]+=1
     offset=r['offset'];length=r['length'];offset=0 if offset=='-0' else offset
     integer=isinstance(offset,(int,float)) and not isinstance(offset,bool) and (isinstance(offset,int) or offset.is_integer())
     classification='array undefined' if length is None else 'non-integer' if not integer else 'out-of-range integer' if offset>=length or offset < -length else 'in-range negative integer' if offset<0 else 'in-range non-negative integer'
     offsetClasses[classification]+=1
assert len(seen)==len(set(seen));assert sorted(seen)==sorted(summary['completedUnits']);assert counts==summary['counts'];assert calls==summary['calls'] and reads==summary['reads'] and diff==summary['differences'] and tdz==summary['tdz'];assert dict(kinds)==summary['readKinds']
if summary['complete']:assert sorted(seen)==sorted(u['name'] for u in plan['units']);assert counts==plan['counts'];assert calls<=reads+exportReads
assert exportReads==summary['exportReads'] and exportUninitialized==summary['exportUninitialized'] and exportTdz==summary['exportTdz'] and dict(exportSites)==summary['exportSites']
result={'units':len(seen),'complete':summary['complete'],'counts':dict(counts),'calls':calls,'reads':reads,'differences':diff,'tdz':tdz,'uninitializedReads':uninitialized+exportUninitialized,'exportReads':exportReads,'exportUninitialized':exportUninitialized,'exportTdz':exportTdz,'exportSites':dict(exportSites),'readKinds':dict(kinds),'archiveAndMemberHashesValid':True,'duplicateUnits':0,'offsetClasses':dict(offsetClasses),'returnKinds':dict(returnKinds),'rawTestIdentitiesAndCountsVerified':True,'inputHashesAndBaselineGuardsVerified':True}
(out/'reconciliation.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result))

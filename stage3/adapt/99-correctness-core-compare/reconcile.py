import json,hashlib,tarfile,sys
from collections import Counter
from pathlib import Path
s=Path('/workspace/scratch/core-compare');r=Path('/workspace/adamic');summary=json.loads((s/'COUNT.json').read_text());assert summary['complete'];sys.path.insert(0,str(r/'stage3/lane'));from shard_plan import load_plan,names_digest
plan=load_plan();units={u['name']:u for u in plan['units']};seen=[];counts=Counter();oracle=Counter();histogram=Counter();archiveHashes={};types=['undefined','number','string']
for piece in summary['pieces']:
 archive=s/f"piece-{piece['piece']:02d}.tar.gz";archiveHashes[archive.name]=hashlib.sha256(archive.read_bytes()).hexdigest();prefix=f"piece-{piece['piece']:02d}/"
 with tarfile.open(archive) as tar:
  for line in tar.extractfile(prefix+'SHA256SUMS').read().decode().splitlines():
   digest,name=line.split('  ',1);assert hashlib.sha256(tar.extractfile(prefix+name).read()).hexdigest()==digest,(archive,name)
  local=Counter()
  for name in piece['units']:
   unit=units[name];report=json.load(tar.extractfile(prefix+name+'/report.json'));observed=json.load(tar.extractfile(prefix+name+'/observations.json'));assert report['status']=='pass' and report['counts']==unit['expect'];assert names_digest(observed)==unit['names_digest'];oracle.update(report['counts']);seen.append(name)
  for member in tar.getmembers():
   if not member.isfile() or not member.name.startswith(prefix+'calls/'):continue
   assert member.name.endswith('.bin'),('counterexample sidecar or unexpected record',member.name)
   assert member.name.split('/')[-1].rsplit('.',2)[0] in piece['units']
   local.update(tar.extractfile(member).read())
  assert {str(k):v for k,v in local.items()}==json.load(tar.extractfile(prefix+'binary-histogram.json'))
  histogram.update(local)
mixed=diff=calls=0;matrix=[]
for code,n in sorted(histogram.items()):
 assert code<81
 pair,value=divmod(code,9);a,b=types[pair//3],types[pair%3];original,adapted=value//3-1,value%3-1
 counts[a+'/'+b]+=n;mixed+=({a,b}=={'string','number'})*n;diff+=(original!=adapted)*n;calls+=n;matrix.append({'typeofA':a,'typeofB':b,'original':original,'adapted':adapted,'calls':n})
assert seen==[u['name'] for u in plan['units']]==summary['units'];assert len(seen)==len(set(seen));assert dict(oracle)==plan['counts'];assert dict(counts)==summary['counts'];assert mixed==summary['mixed']==0 and diff==summary['differences']==0 and calls==summary['calls']
assert hashlib.sha256((s/'count-preload.cjs').read_bytes()).hexdigest()==json.loads((s/'recorder-provenance.json').read_text())['sha256']
result={'complete':True,'units':len(seen),'calls':calls,'counts':dict(counts),'mixed':mixed,'differences':diff,'resultMatrix':matrix,'oracleCounts':dict(oracle),'independentClassificationAgrees':True,'allRawMemberHashesVerified':True,'exactlyOneCompleteUnitUnion':True,'recorderVersion':4,'archives':archiveHashes};(s/'RECONCILED.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)

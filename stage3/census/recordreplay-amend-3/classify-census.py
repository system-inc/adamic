from pathlib import Path
from collections import defaultdict,Counter
import json,re,sys
D=Path('/workspace/scratch/recordreplay-amend-3-proof')
mainPath=Path(sys.argv[1]);finalPath=Path(sys.argv[2]);mainTree=Path(sys.argv[3]);finalTree=Path(sys.argv[4]);output=Path(sys.argv[5])
def relative(s):return re.sub(r'^.*?/src/','src/',s)
def message(f):
 text=re.sub(r'^.*?sys\.ts:\d+:\d+:\s*','',f['text'])
 text=re.sub(r'/workspace/scratch/[^ ]*?/src/','src/',text)
 return re.sub(r':\d+:\d+','',text)
def key(f):return f['kind'],f['phase'],f['reason'],message(f)
def read(p):
 rows=[json.loads(l) for l in p.read_text().splitlines()];h=rows[0]
 assert h['module_format']=='CommonJSBundle' and h['latent_mode']=='full' and not h['checker_rejected'] and not h['diagnostics'];assert rows[-1]['status']=='relation_report'
 files=[r for r in rows if r.get('file')];assert len(files)==1 and files[0]['file'].endswith('/src/compiler/sys.ts');return files[0]
a=read(mainPath);b=read(finalPath);buckets=defaultdict(list)
for f in a['findings']:buckets[key(f)].append(f)
added=[];paired=[];matched=set()
# Complete exact unit/site matches across the whole stream before allowing
# a same-message record in an earlier unit to consume a later unit's match.
for mode in ['unit-and-site','unit','tuple']:
 for i,f in enumerate(b['findings']):
  if i in matched:continue
  old=buckets[key(f)]
  def eligible(x):
   if mode=='tuple':return True
   if relative(x['unit'])!=relative(f['unit']):return False
   return mode=='unit' or relative(x['where'])==relative(f['where'])
  at=next((j for j,x in enumerate(old) if eligible(x)),None)
  if at is not None:paired.append((old.pop(at),f));matched.add(i)
added=[f for i,f in enumerate(b['findings']) if i not in matched]
removed=[f for values in buckets.values() for f in values];mainByUnit=defaultdict(list)
for f in a['findings']:mainByUnit[relative(f['unit'])].append(f)
baselineUnits={relative(u['where']):u for u in a['units']};finalUnits={relative(u['where']):u for u in b['units']}
def quote(site,tree):
 m=re.fullmatch(r'(src/.*):(\d+):(\d+)',site)
 return (tree/m[1]).read_text().splitlines()[int(m[2])-1].strip() if m else None
items=[]
for i,f in enumerate(added,1):
 unit=relative(f['unit']);baseline=mainByUnit[unit];primary=[x for x in baseline if x['kind']!='Boundary'];exists=unit in baselineUnits
 classification='revealed_later_stop' if baseline else 'newly_stopped_unit'
 oldStops=[dict(kind=x['kind'],phase=x['phase'],site=relative(x['where']),reason=x['reason'],message=message(x)) for x in primary]
 actual=re.match(r'^(.*?sys\.ts:\d+:\d+): ',f['text']);stopSite=relative(actual[1]) if actual else relative(f['where'])
 items.append(dict(id=i,unit=unit,unitName=quote(unit,finalTree),unitKind=finalUnits.get(unit,{}).get('kind'),site=relative(f['where']),stopSite=stopSite,kind=f['kind'],phase=f['phase'],reason=f['reason'],message=message(f),classification=classification,unitExistedOnBaseline=exists,baselineUnitPrimaryFindingCount=len(primary),baselineUnitStops=oldStops,sourceLine=quote(stopSite,finalTree)))
oldStopped={u for u,fs in mainByUnit.items() if fs}
newByUnit=defaultdict(list)
for f in b['findings']:
 newByUnit[relative(f['unit'])].append(f)
newlyStopped=sorted(u for u in newByUnit if u not in oldStopped)
regressions=[dict(unit=u,existedOnMain=u in baselineUnits,findings=[dict(site=relative(f['where']),kind=f['kind'],phase=f['phase'],reason=f['reason'],message=message(f),sourceLine=quote(relative(f['where']),finalTree)) for f in fs]) for u,fs in newByUnit.items() if u not in oldStopped]
report=dict(baselineFindings=len(a['findings']),finalFindings=len(b['findings']),baselineUnits=len(a['units']),finalUnits=len(b['units']),paired=len(paired),addedRecords=len(added),removedRecords=len(removed),newNotYet=sum(f['kind']=='NotYet' for f in added),newBoundary=sum(f['kind']=='Boundary' for f in added),classes=dict(Counter(i['classification'] for i in items)),newlyStoppedUnits=newlyStopped,newlyStoppedUnitDetails=regressions,newFindingFreeUnits=sorted(u for u in finalUnits if u not in baselineUnits and u not in newByUnit),pairing=['kind','phase','reason','message with position stripped'],records=items,removed=[dict(unit=relative(f['unit']),site=relative(f['where']),kind=f['kind'],phase=f['phase'],reason=f['reason'],message=message(f)) for f in removed])
output.write_text(json.dumps(report,indent=2)+'\n')
print(json.dumps({k:v for k,v in report.items() if k not in ['records','removed','newlyStoppedUnitDetails']},indent=2))

if newlyStopped: raise SystemExit(1)

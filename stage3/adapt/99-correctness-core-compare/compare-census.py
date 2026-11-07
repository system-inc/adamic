import json,re,sys
from pathlib import Path
s=Path('/workspace/scratch/core-compare');pin=json.loads((Path('/workspace/adamic/stage3/adapt/99-correctness-core-compare/owners.json')).read_text())[0]
before=(s/'before/src/compiler/core.ts').read_text();start=before.index(pin['before']);end=start+len(pin['before']);first=before[:start].count('\n')+1;last=before[:end].count('\n')+1;delta=pin['after'].count('\n')-pin['before'].count('\n')
def load(side):
 actualSide=sys.argv[1] if side=='after' and len(sys.argv)>1 else side
 raw=[json.loads(l) for l in (s/('census-'+actualSide)/'raw.jsonl').read_text().splitlines()];assert not raw[0]['checker_rejected'] and raw[0]['diagnostics']==[];return next(r for r in raw if r.get('file','').endswith('/core.ts'))
def sites(side):
 result=set()
 for f in load(side)['findings']:
  if f['kind']=='Boundary':continue
  m=re.search(r'core.ts:(\d+):(\d+)$',f['where']);assert m,f
  line,column=map(int,m.groups())
  if side=='after' and line>last+delta:line-=delta
  result.add((line,column,f['kind'],f['reason'],f['phase']))
 return result
old=sites('before');now=sites('after');mutant=sites('mutant');removed=old-now;added=now-old
ownerUnits={}
for side in ['before','after','mutant']:
 ownerUnits[side]=[u for u in load(side)['units'] if u['kind']=='KindFunctionDeclaration' and first<=int(re.search(r':(\d+):(\d+)$',u['where']).group(1))<=last]
 assert len(ownerUnits[side])==3 and all(u['status']=='attempted' for u in ownerUnits[side]),ownerUnits[side]
expected='a BinaryExpression with a union of differently held members and a union of differently held members'
assert removed and all(first<=row[0]<=last and row[2]=='NotYet' and row[3]==expected for row in removed),removed
assert not added,added
assert mutant==old,(mutant-old,old-mutant)
report={'beforeDistinctSites':len(old),'afterDistinctSites':len(now),'removed':[list(r) for r in sorted(removed)],'newSites':[],'fullCoreCompared':True,'mutantRestoresExactlyOriginalSites':True,'ownerStartLine':first,'ownerEndLine':last,'lineDelta':delta,'checkerDiagnostics':0,'ownerProbes':ownerUnits};(s/'CENSUS.json').write_text(json.dumps(report,indent=2)+'\n');print(json.dumps(report),flush=True)

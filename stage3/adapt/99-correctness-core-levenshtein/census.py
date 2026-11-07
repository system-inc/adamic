import json,os,time,signal,subprocess,shutil,re
from pathlib import Path
s=Path('/workspace/scratch/levenshtein');repo=Path('/workspace/adamic')
for label in ['before','after','mutant']:
 tree=s/label
 if label=='mutant':
  shutil.copytree(s/'after/src',tree/'src')
  pin=json.loads((repo/'stage3/adapt/99-correctness-core-levenshtein/owners.json').read_text())[0];p=tree/'src/compiler/core.ts';raw=p.read_bytes();assert raw.decode().replace('\r\n','\n').count(pin['after'])==1;assert raw.count(b'new Array<number>(s2.length + 1).fill(0)')==2;restored=raw.replace(b'new Array<number>(s2.length + 1).fill(0)',b'new Array(s2.length + 1)');p.write_bytes(restored)
  names=sorted(p.relative_to(s/'after/src') for p in (s/'after/src').rglob('*') if p.is_file());changed=[str(name) for name in names if (s/'after/src'/name).read_bytes()!=(tree/'src'/name).read_bytes()];assert changed==['compiler/core.ts'];(s/'mutant-scope.json').write_text(json.dumps({'sourceFiles':len(names),'changed':changed,'onlyOriginalInitializersRestored':True,'sixAssertionsRetained':True},indent=2)+'\n')
 if not (s/(label+'-census-result.json')).exists():
  subprocess.run(['python3',str(s/'run-census.py'),label],check=True)

def read(label):
 rows=[json.loads(l) for l in (s/(label+'-census.jsonl')).read_text().splitlines()];assert not rows[0]['checker_rejected'];record=next(r for r in rows if r.get('file','').endswith('/core.ts'));sites={}
 for f in record['findings']:
  if f['kind']=='Boundary':continue
  key=tuple([f['where'].split('/core.ts:')[-1]]+[f[k] for k in ['kind','reason','phase']]);sites[key]=f
 return record,sites
before,bs=read('before');after,asites=read('after');mutant,ms=read('mutant');removed=sorted(set(bs)-set(asites));added=sorted(set(asites)-set(bs))
assert not added,added
# The mutant keeps all six checked reads and restores only the initializers.
# Scope the owner by declaration offsets and source line range.
text=(s/'before/src/compiler/core.ts').read_text();pin=json.loads((repo/'stage3/adapt/99-correctness-core-levenshtein/owners.json').read_text())[0];start=text.index(pin['before']);lo=text[:start].count('\n')+1;hi=lo+pin['before'].count('\n')
owned=lambda sites:[k for k in sites if lo<=int(k[0].split(':')[0])<=hi]
assert owned(bs) and not owned(asites)
assert len(removed)==len(owned(bs))==15,removed
initSites=[k for k in owned(bs) if k[2]=='new an Identifier'];assert len(initSites)==2 and all(k in ms for k in initSites),initSites
rangeLine=text[:text.index('export function getOwnValues')].count('\n')+1
rangeSites=sorted((k for k in bs if int(k[0].split(':')[0])>=rangeLine),key=lambda k:tuple(map(int,k[0].split(':'))))
removedIds=['F'+str(rangeSites.index(k)+1) for k in removed];assert set(removedIds)=={'F'+str(i) for i in range(22,37)},removedIds
result={'beforeSites':len(bs),'afterSites':len(asites),'ownerLineRange':[lo,hi],'ownerBeforeSites':owned(bs),'ownerAfterSites':owned(asites),'removed':removed,'removedHistoricalIds':removedIds,'added':added,'mutantOwnerSites':owned(ms),'mutantRestoresInitializers':True,'beforeUnits':len(before['units']),'afterUnits':len(after['units'])}
(s/'census-comparison.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result),flush=True)

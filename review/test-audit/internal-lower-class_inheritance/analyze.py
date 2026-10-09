import json,pathlib,statistics,re
out=pathlib.Path('/workspace/adamic/review/test-audit/internal-lower-class_inheritance');names=json.loads((out/'names.json').read_text());timings=json.loads((out/'timings.json').read_text());plans=json.loads((out/'plan.json').read_text());runs=json.loads((out/'runs.json').read_text());matrix={};evidence={};subpasses={}
for run in runs:
 id=run['id'];files=[out/(id+'.log')]
 if run['bounded']:files += [out/(id+'-'+n+'.log') for n in names]
 fail=set();seen=set();elapsed={};lines={};passed=set()
 for file in files:
  raw=file.read_text()
  if file.name != id+'.log' and 'panic:' in raw:
   solo=file.name[len(id)+1:-4];fail.add(solo);seen.add(solo)
  for l in raw.splitlines():
   try:d=json.loads(l)
   except:continue
   test=d.get('Test','');top=test.split('/')[0]
   if d.get('Action')=='fail' and test:fail.add(top)
   if d.get('Action') in ['pass','fail','skip'] and test:seen.add(top)
   if d.get('Action')=='pass' and '/' in test:passed.add(test)
   if d.get('Action')=='output' and test:
    text=d.get('Output','').strip()
    if re.search(r'\w+_test\.go:\d+:',text):lines.setdefault(top,(str(file.name),text))
 matrix[id]={'failed':sorted(fail),'seen':sorted(seen),'bounded':run['bounded'],'command':'timeout 120 go test -json -count=1 -timeout 90s ./internal/lower/ -run .','code':run['code'],'wall':run['wall']};evidence[id]=lines;subpasses[id]=passed
(out/'matrix.json').write_text(json.dumps(matrix,indent=2))
(out/'matrix-raw.json').write_text(json.dumps(matrix,indent=2))
family=json.loads((out/'family.json').read_text());members=family['members'];fn=family['name'];timings[fn]=family['times']
override_members=['TestAbfe962OverrideDefaultAddedRefusesNativeSignature','TestOverrideOptionalNumberRefusesNativeSignature','TestOverrideOptionalRefusesNativeSignature']
for id,m in matrix.items():
 for key in ['failed','seen']:m[key]=sorted({fn if x in members else 'TestOverride family' if x in override_members else x for x in m[key]})
 for member in members:
  if member in evidence[id]:evidence[id].setdefault(fn,evidence[id][member])
(out/'matrix.json').write_text(json.dumps(matrix,indent=2))
names=[fn if n==members[0] else n for n in names if n not in members[1:]]
rows=[]
for n in names:
 kills=[m['id'] for m in plans if m['id'].startswith('M') and m['id']!='M11' and n in matrix.get(m['id'],{}).get('failed',[])]
 unique=[id for id in kills if len(matrix[id]['failed'])==1]
 others=[x for x in names if x!=n and kills and all(x in matrix[id]['failed'] for id in kills)]
 if unique:verdict='sacred';subs=[]
 elif not kills:verdict='untrue';subs=[]
 elif others:verdict='subsumed';subs=[min(others,key=lambda x:statistics.median(timings[x]))]
 else:
  allothers=set().union(*(set(matrix[id]['failed'])-{n} for id in kills));common=set.intersection(*(set(matrix[id]['failed'])-{n} for id in kills))
  verdict='subsumed' if common else 'overlapping';subs=sorted(common)[:1] if common else sorted(allothers)
  if not common:
   import itertools
   candidates=sorted(allothers,key=lambda x:(x not in names,x))
   for width in range(2,5):
    options=[list(c) for c in itertools.combinations(candidates,width) if all(any(x in matrix[id]['failed'] for x in c) for id in kills)]
    if options:
     subs=min(options,key=lambda c:(sum(x not in names for x in c),sum(statistics.median(timings.get(x,[1])) for x in c)));break
 owned='P_LOAD' if n=='TestInheritanceKeepsCheckerConstructorRules' else 'P_FIELDS' if n=='TestInheritanceCycleFinderIncludesInheritedFields' else 'P_LOWER'
 observed=n in matrix.get(owned,{}).get('seen',[]);pk=[owned] if n in matrix.get(owned,{}).get('failed',[]) else []
 id=kills[-1] if kills else None;entry=evidence.get(id,{}).get(n) if id else None
 proof=unique[-1] if unique else next((x for x in reversed(kills) if x!='M17'),id)
 proofentry=evidence.get(proof,{}).get(n) if proof else None
 oracle='Handwritten acceptance/refusal and diagnostic substring assertions'
 if n=='TestInheritanceKeepsCheckerConstructorRules':oracle='Handwritten *load.CheckError and abstract/super substring checks; upstream checker is not independently run or compared'
 if n in ['TestInheritanceHasClassIdentity','TestInheritanceGenericMonomorphizations','TestInheritanceGenericFactoryLayouts']:oracle='Handwritten IR class identity or field-layout assertions'
 if n=='TestInheritanceCycleFinderIncludesInheritedFields':oracle='Handwritten inherited property-symbol membership assertion'
 row={'test':n,'package':'internal/lower','file':'internal/lower/class_inheritance_test.go','seconds':statistics.median(timings[n]),'oracle':oracle,'oracle_kind':'self','kills':kills,'unique_kills':unique,'last_proven_fail':(id+': '+entry[1]) if entry else None,'verdict':verdict,'subsumed_by':subs,'mutants_in_matrix':sum(m['id'].startswith('M') and m['id']!='M11' and m['id'] in matrix for m in plans),'probe_kills':pk,'subsumer_seconds':statistics.median(timings[subs[0]]) if verdict=='subsumed' and subs[0] in timings else None,'vacuous':(not pk) if observed else None,'bounded':any(matrix[id]['bounded'] for id in kills),'matrix_rows':names if any(matrix[id]['bounded'] for id in kills) else [],'evidence':('ADAMIC_MUTANT='+proof+' ADAMIC_BUILD_CACHE_DIR=/tmp/u029/cache/'+proof+' '+matrix[proof]['command']+' > '+proofentry[0]+' 2>&1; '+proofentry[1]) if proofentry else 'No production mutant caught this row in the completed matrix'}
 if owned in subpasses:
  positives=sorted(x for x in subpasses[owned] if x.startswith(n+'/'))
  if pk and positives:row['vacuous_subcases']=positives
 if n==fn:row['members']=members;row['oracle']='Handwritten success-only err == nil checks; no IR or execution result checked'
 if n=='TestInheritanceRefusesGrowingGenericClasses':row['oracle']+='; only caught M17, an earlier computed-base rejection, not a mutation of the recursion bound'
 rows.append(row)
(out/'rows.json').write_text(json.dumps(rows,indent=2));print([(r['test'],r['verdict'],r['kills'],r['unique_kills'],r['subsumed_by'],r['vacuous']) for r in rows])

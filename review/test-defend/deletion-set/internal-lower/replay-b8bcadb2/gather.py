import subprocess,json,pathlib,re,os
P=pathlib.Path('review/test-defend/deletion-set/internal-lower/replay-b8bcadb2')
C=['TestAdamicNullishAssertionsAreRefused','TestArgumentsLengthRefusalFixtures','TestArgumentsLengthRefusals','TestArrayPredicateCannotInventAnElementContract','TestClassFeaturesReadonlyChecker','TestClassFeaturesStaticParentCycle','TestClassWrongOutputPrivateRepair','TestDefaultTaggedInterfaceNeedsNoFlag','TestDefiniteAssignmentSoundNeighbors','TestEveryNeedsCallbackEffects','TestInheritanceGenericFactoryLayouts','TestInputSpreadArgumentsAreNotYet','TestOptionalWideningAllowed','TestOptionalWideningCensus','TestPredicateOverloadCallback']
json.dump(C,open(P/'candidates.json','w'),indent=2)
allmutants=[];inventory=[]
def show(b,path):return subprocess.check_output(['git','show','origin/'+b+':'+path])
def records(d,path,output,test=None):
 if isinstance(d,list):
  for x in d:records(x,path,output,test)
 elif isinstance(d,dict):
  test=d.get('test',test)
  mid=d.get('mutant',d.get('id'))
  for key in ['rows_failed','failed_rows','failed','failures','kills']:
   val=d.get(key)
   if mid and isinstance(val,list) and all(isinstance(x,str) for x in val) and any(x.startswith('Test') for x in val):output.append((path,str(mid),val))
  if test and isinstance(d.get('kills'),list):
   for m in d['kills']:output.append((path,str(m),[test]))
  for k,v in d.items():
   if re.fullmatch('[MD][0-9]+',k):
    if isinstance(v,list) and all(isinstance(x,str) for x in v):output.append((path,k,v))
    elif isinstance(v,dict) and isinstance(v.get('rows_failed'),list):output.append((path,k,v['rows_failed']))
   records(v,path,output,test)
for unit in ['arguments_length','check_pragmas','class_inheritance','class_instance_key','enums','interface_cast','non_null_impossible','predicates_overload']:
 for wave in ['audit','defend']:
  b='test-'+wave+'/internal-lower-'+unit;root='review/'+b+'/'
  paths=subprocess.check_output(['git','ls-tree','-r','--name-only','origin/'+b,root]).decode().splitlines(); rec=[]
  for path in paths:
   if not path.endswith('.json'):continue
   try:records(json.loads(show(b,path)),path,rec)
   except (ValueError,TypeError):continue
  prefix='M' if wave=='audit' else 'D'
  for path in paths:
   name=pathlib.Path(path).name
   if not re.fullmatch(prefix+r'[0-9]+(?:-replay)?\.diff',name):continue
   mid=name.split('.')[0].split('-')[0];matching=[r for r in rec if r[1]==mid]
   if not matching:inventory.append(dict(branch=b,file=path,reason='no parsed failing-row metadata'));continue
   def proximity(r):return (len(os.path.commonpath([str(pathlib.Path(path).parent),str(pathlib.Path(r[0]).parent)]).split('/')), -len(pathlib.Path(r[0]).parent.parts))
   closest=max(map(proximity,matching));used=[r for r in matching if proximity(r)==closest]
   failed=sorted({x.split('/')[0] for _,_,rows in used for x in rows if x.startswith('Test')});candidate=sorted(set(failed)&set(C))
   if not candidate:continue
   identifier=f'{wave}-{unit}-{mid}'+('-replay' if '-replay' in name else '')
   # Preserve separate paths with reused identifiers.
   if any(x['key']==identifier for x in allmutants):identifier+='-'+str(len(allmutants))
   raw=show(b,path);dest=P/'diffs'/ (identifier+'.diff');dest.parent.mkdir(exist_ok=True);dest.write_bytes(raw)
   check=subprocess.run(['git','apply','--check',str(dest)],capture_output=True)
   hunk=re.search(rb'^@@ -(\d+)',raw,re.M); file=re.search(rb'^--- a/(.+)$',raw,re.M)
   allmutants.append(dict(key=identifier,mutant=mid,file=str(dest),source_file=path,file_line=(file[1].decode()+':'+hunk[1].decode()) if file and hunk else '',branch=b,candidates_failed=candidate,other_rows_failed=sorted(set(failed)-set(C)),metadata=[r[0] for r in used],stale=check.returncode!=0,apply_error=check.stderr.decode()))
json.dump(allmutants,open(P/'mutant-list.json','w'),indent=2);json.dump(inventory,open(P/'metadata-unresolved.json','w'),indent=2)
print('gathered',len(allmutants),'stale',sum(x['stale'] for x in allmutants))
for x in allmutants:print(x['key'],x['candidates_failed'],'stale' if x['stale'] else 'applies')

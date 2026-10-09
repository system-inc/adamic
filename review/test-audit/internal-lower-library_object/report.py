import pathlib,json,re,statistics,collections
p=pathlib.Path('/workspace/adamic/review/test-audit/internal-lower-library_object');rows=json.loads((p/'scope.json').read_text());menu=json.loads((p/'menu.json').read_text());meta=json.loads((p/'run-meta.json').read_text())
def events(f):
 es=[]
 for line in f.read_text().splitlines():
  try:es.append(json.loads(line))
  except:pass
 return es
def observed(mid):
 logs=[p/(mid+'.log')]
 if (p/(mid+'-bounded.log')).exists():logs=[p/(mid+'-bounded.log')]
 if list(p.glob(mid+'-Test*.log')):logs=list(p.glob(mid+'-Test*.log'))
 outcomes={};fail={};passedsubs=collections.defaultdict(list)
 for f in logs:
  es=events(f)
  for e in es:
   test=e.get('Test','');r=test.split('/')[0]
   if r not in rows:continue
   if e['Action'] in ('pass','fail','skip'):
    if test==r:outcomes[r]=e['Action']
    elif e['Action']=='pass':passedsubs[r].append(test)
   if e['Action']=='output' and re.search(r'\.go:\d+:',e.get('Output','')):fail[r]=e['Output'].strip()
  if 'panic:' in f.read_text():
   roots={e.get('Test','').split('/')[0] for e in es if e['Action']=='run' and e.get('Test')}
   if len(roots)==1:
    r=next(iter(roots));outcomes[r]='fail';fail[r]=next((e['Output'].strip() for e in es if 'panic:' in e.get('Output','')),'panic')
 return outcomes,fail,passedsubs
matrix={};failure={};subs={}
for m in menu:
 mid=m['id'];matrix[mid],failure[mid],subs[mid]=observed(mid)
 m['failed_rows']=[r for r in rows if matrix[mid].get(r)=='fail'];m['unknown_rows']=[r for r in rows if r not in matrix[mid]]
 es=events(p/(mid+'.log'))
 m['outside_failed_rows']=sorted({e['Test'] for e in es if e['Action']=='fail' and e.get('Test') and '/' not in e['Test'] and e['Test'] not in rows})
 m['full_package_status']=meta[mid]['status']
 m['full_package_wall']=meta[mid]['wall']
 m['cooked']=meta[mid]['status']==124 or 'test timed out after' in (p/(mid+'.log')).read_text()
(p/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n');(p/'mutants.json').write_text(json.dumps(menu,indent=2)+'\n')
secs={r:statistics.median(float(re.search(r'\bok\s+\S+\s+([\d.]+)s',(p/f'timing-{r}-{n}.log').read_text())[1]) for n in range(3)) for r in rows}
files={}
for f in pathlib.Path('/workspace/adamic/internal/lower').glob('*_test.go'):
 for r in re.findall(r'func (Test\w+)\(',f.read_text()):files[r]='internal/lower/'+f.name
oracles={
'TestObjectRefusalsExplainSoundness':'Requires Refused plus handwritten reason substrings. Does not assert the whole diagnostic.',
'TestObjectUnprovenShapesStayNotYet':'Requires only errors.As(NotYet); ignores location and reason. M05 passes despite a wrong Object.collectBy diagnostic; M02 passes several cases with a different NotYet reason.',
'TestLibraryRegexOffsetRequiresNoCaptures':'Requires a producer be visited and the capture-free proof return false for four unproven inputs. No positive capture-free control; its false entry probe passes.',
'TestLibraryStringRefusals':'Requires an error containing a handwritten reason; does not require a particular error class.',
'TestConsoleLowersToWriteLine':'Exact handwritten IR statements, streams, string deduplication and source filename.',
'TestWhatStageZeroCannotLowerIsRefusedWithWhereAndWhat':'Requires NotYet and handwritten diagnostic suffixes containing position and construct.',
'TestWhatZeroOneRefusesIsRefusedWithAFix':'Requires Refused and handwritten diagnostic substrings. Some expectations stop before the fix text, so the name promises more than every subcase checks.',
'TestAMutableLocationSeenWiderIsRefused':'Requires Refused and handwritten diagnostic substrings describing unsafe views. Does not execute the alleged unsound writes on Node.',
'TestAViewThatCantWriteIsNotRefused':'Rejects only errors.As(Refused). Accepts nil and all other errors, including NotYet. Its nil Lower entry probe passes all 26 subcases.',
'TestATupleSeenAsAnArrayIsNotYet':'Requires NotYet and handwritten diagnostic suffixes naming representation mismatch and workaround.',
'TestAMethodReadAsAValueIsRefused':'Negative cases require Refused and diagnostic substrings; four positive neighbors reject only Refused and accept an empty Lower answer.'}
result=[];ms=[m['id'] for m in menu if m['kind']=='mutant']
for r in rows:
 kills=[mid for mid in ms if matrix[mid].get(r)=='fail']
 unique=[mid for mid in kills if sum(matrix[mid].get(x)=='fail' for x in rows)==1 and all(x in matrix[mid] for x in rows)]
 candidates=[x for x in rows if x!=r and kills and all(matrix[mid].get(x)=='fail' for mid in kills)]
 if unique:verdict='sacred';names=[]
 elif candidates:verdict='subsumed';names=[min(candidates,key=lambda x:secs[x])]
 elif kills:verdict='overlapping';names=sorted({x for mid in kills for x in rows if x!=r and matrix[mid].get(x)=='fail'})
 else:verdict='untrue';names=[]
 last=kills[-1] if kills else None;line=failure.get(last,{}).get(r)
 if line and len(line)>500:line=line[:500]+' [full output in log]'
 own='P02' if r=='TestLibraryRegexOffsetRequiresNoCaptures' else 'P01'
 cmd=meta.get(last+'-'+r,meta.get(last+'-bounded',meta.get(last,{}))).get('command') if last else None
 cmd=(('ADAMIC_BUILD_CACHE_DIR=/tmp/u036/cache/M19-corrected ' if last=='M19' else 'ADAMIC_MUTANT='+last+' ADAMIC_BUILD_CACHE_DIR=/tmp/u036/cache/'+last+' ')+cmd) if last and cmd else None
 o=dict(test=r,package='internal/lower',file=files[r],seconds=secs[r],oracle=oracles[r],oracle_kind='self',kills=kills,unique_kills=unique,last_proven_fail=(last+': '+str(line)) if last else None,verdict=verdict,subsumed_by=names,mutants_in_matrix=20,probe_kills=[mid for mid in ['P01','P02'] if matrix[mid].get(r)=='fail'],subsumer_seconds=secs[names[0]] if verdict=='subsumed' else None,vacuous=matrix[own].get(r)=='pass' if r in matrix[own] else None,vacuous_subcases=subs[own].get(r,[]),bounded=True,matrix_rows=rows,evidence=(str(cmd)+'; '+str(line)) if last else None,subsumption_mutants=len(kills) if verdict=='subsumed' else None)
 result.append(o)
(p/'rows.json').write_text(json.dumps(result,indent=2)+'\n')
print([(r['test'],r['verdict'],r['kills'],r['unique_kills'],r['vacuous']) for r in result])
print('SLICE SURVIVORS',[(m['id'],m['full_package_status'],m['outside_failed_rows']) for m in menu if m['kind']=='mutant' and not m['failed_rows']])

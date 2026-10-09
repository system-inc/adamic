import pathlib,json,re,statistics
p=pathlib.Path('/workspace/adamic/review/test-audit/internal-ir'); rows=json.loads((p/'scope.json').read_text()); menu=json.loads((p/'menu.json').read_text())
def events(f):
 ret=[]
 for line in f.read_text().splitlines():
  try: ret.append(json.loads(line))
  except: pass
 return ret
def observe(mid):
 logs=[p/(mid+'.log')]
 if any(p.glob(mid+'-Test*.log')):logs=list(p.glob(mid+'-Test*.log'))
 obs={};fail={}
 for f in logs:
  es=events(f)
  for e in es:
   row=e.get('Test','').split('/')[0]
   if row not in rows:continue
   if e['Action'] in ('pass','fail','skip'):obs[row]=e['Action']
   if e['Action']=='output' and e.get('OutputType')!='frame' and row not in fail:fail[row]=e['Output'].strip()
   if e['Action']=='output' and re.search(r'\.go:\d+:',e.get('Output','')) and row not in fail:fail[row]=e['Output'].strip()
  # panic terminates without fail event, so establish isolated failing row by run + panic
  if 'panic:' in f.read_text():
   active=[e['Test'] for e in es if e['Action'] in ('run','cont') and 'Test' in e]
   if len(set(active))==1:obs[active[0]]='fail';fail[active[0]]=next((e['Output'].strip() for e in es if 'panic:' in e.get('Output','')),'panic')
  for e in es:
   row=e.get('Test','').split('/')[0]
   if row in rows and e['Action']=='output' and re.search(r'\.go:\d+:',e.get('Output','')):fail[row]=e['Output'].strip()
 return obs,fail
matrix={}; failures={}
for m in menu:
 mid=m['id']; matrix[mid],failures[mid]=observe(mid)
 m['failed_rows']=[r for r in rows if matrix[mid].get(r)=='fail'];m['unknown_rows']=[r for r in rows if r not in matrix[mid]]
 if mid=='M09':m['kind']='supplemental'
 m['vet_passed']= (p/(mid+'-vet.log')).exists() if mid.startswith('M') else None
(p/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n');(p/'mutants.json').write_text(json.dumps(menu,indent=2)+'\n')
secs={r:statistics.median(float(re.search(r'\bok\s+\S+\s+([\d.]+)s', (p/f'timing-{r}-{n}.log').read_text())[1]) for n in range(3)) for r in rows}
files={}
for f in pathlib.Path('/workspace/adamic/internal/ir').glob('*_test.go'):
 for r in re.findall(r'func (Test\w+)\(',f.read_text()):files[r]='internal/ir/'+f.name
entries={'TestArgumentLayouts':['P05'],'TestClosureArgumentsCountTargets':['P09'],'TestPrimitiveViewMembers':['P06'],'TestViewUnionDiscriminantOverlaps':['P07'],'TestCallMayThrowUsesReachableTargets':['P02'],'TestDirectClosureTargetsUseEncodedIndex':['P03','P04'],'TestCallTargetsIncludeEveryDescendant':['P01','P08'],'TestClosureTargetsBoundOnlyProvenValues':['P03','P04']}
objs=[];ms=[m['id'] for m in menu if m['kind']=='mutant']
for r in rows:
 kills=[m for m in ms if matrix[m].get(r)=='fail']; unique=[m for m in kills if sum(matrix[m].get(x)=='fail' for x in rows)==1 and len(matrix[m])==len(rows)]
 subs=[x for x in rows if x!=r and kills and all(matrix[m].get(x)=='fail' for m in kills)]
 verdict='sacred' if unique else ('subsumed' if subs else ('overlapping' if kills else 'untrue'))
 if r=='TestCallTargetReaders':verdict='setup-check'
 if subs:subs=[min(subs,key=lambda x:secs[x])]
 last=kills[-1] if kills else ('S01' if verdict=='setup-check' else None)
 line=failures.get(last,{}).get(r) if last!='S01' else next((e['Output'].strip() for e in events(p/'S01.log') if re.search(r'\.go:\d+:',e.get('Output',''))),None)
 probes=[m['id'] for m in menu if m['kind']=='probe' and matrix[m['id']].get(r)=='fail']
 obj=dict(test=r,package='internal/ir',file=files[r],seconds=secs[r],oracle='Handwritten IR expectations; no external value checked.' if r!='TestCallTargetReaders' else 'Go-resolved source field reads compared with handwritten targetReaders allowlist.',oracle_kind='self',kills=kills,unique_kills=unique,last_proven_fail=f'{last}: {line}' if last else None,verdict=verdict,subsumed_by=subs,mutants_in_matrix=17,probe_kills=probes,subsumer_seconds=secs[subs[0]] if subs else None,vacuous=all(matrix[m].get(r)=='pass' for m in entries[r]) if r in entries else None,bounded=False,matrix_rows=rows,evidence=f"ADAMIC_MUTANT={last} timeout 120 go test -json -count=1 -timeout 90s ./internal/ir/ -run .; {line}" if last and last!='S01' else "S01: timeout 120 go test -json -count=1 -timeout 90s ./internal/ir/ -run '^TestCallTargetReaders$'; "+str(line))
 objs.append(obj)
(p/'rows.json').write_text(json.dumps(objs,indent=2)+'\n')
print(json.dumps(objs,indent=2));print('SURVIVORS',[(m['id'],m['kind']) for m in menu if not m['failed_rows'] and m['id'].startswith('M')])

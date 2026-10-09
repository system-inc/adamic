import pathlib,json,re,statistics,subprocess
p=pathlib.Path(__file__).parent;subprocess.run(['python3',str(p/'summarize.py')],check=True,stdout=subprocess.DEVNULL)
raw=json.loads((p/'results.json').read_text());(p/'individual-results.json').write_text(json.dumps(raw,indent=2));families=json.loads((p/'families.json').read_text());family,members=next(iter(families.items()));matrix=json.loads((p/'matrix.json').read_text());unit=[o for o in raw if o['test'] not in members]
f=dict(next(o for o in raw if o['test']==members[0]));f['test']=family;f['members']=members;f['kills']=sorted(set(sum([o['kills'] for o in raw if o['test'] in members],[])));f['seconds']=statistics.median([float(re.search(r'\t([0-9.]+)s',(p/('TestNodeFSFile_family-'+str(i)+'.log')).read_text())[1]) for i in range(3)]);f['oracle']='Handwritten err == nil acceptance only. Four source-input wrappers share lowerSource and the same assertion; IR contents are not checked.';f['vacuous']=True;f['probe_kills']=[];buf=next(o for o in raw if o['test']=='TestNodeFSFileBufferBorrow');f['last_proven_fail']=buf['last_proven_fail'];f['evidence']=buf['evidence'];unit.insert(0,f)
for m,v in matrix.items():v['failed_rows']=sorted(set(family if r in members else r for r in v['failed_tests']))
(p/'matrix.json').write_text(json.dumps(matrix,indent=2));allrows=sorted(set(sum([v['failed_rows'] for m,v in matrix.items() if m.startswith('M')],[])));ks={r:[m for m,v in matrix.items() if m.startswith('M') and r in v['failed_rows']] for r in allrows};known={o['test']:o['seconds'] for o in unit};extra='TestEmptyNeverMapCannotGainWritableInhabitants';known[extra]=statistics.median([float(re.search(r'\t([0-9.]+)s',(p/(extra+f'-{i}.log')).read_text())[1]) for i in range(3)])
for o in unit:
 row=o['test'];kills=o['kills'];o['unique_kills']=[m for m in kills if matrix[m]['failed_rows']==[row]];o['subsumed_by']=[];o['subsumer_seconds']=None
 if o['unique_kills']:o['verdict']='sacred'
 elif not kills:o['verdict']='untrue'
 else:
  cand=[r for r in allrows if r!=row and set(kills)<=set(ks[r])];measured=[r for r in cand if r in known]
  if cand:
   assert measured,(row,cand);sub=min(measured,key=known.get);o['verdict']='subsumed';o['subsumed_by']=[sub];o['subsumer_seconds']=known[sub]
  else:
   o['verdict']='overlapping';remain=set(kills)
   while remain:
    sub=max([r for r in allrows if r!=row],key=lambda r:(len(remain&set(ks[r])),r in known));assert remain&set(ks[sub]);o['subsumed_by'].append(sub);remain-=set(ks[sub])
(p/'results.json').write_text(json.dumps(unit,indent=2));print(json.dumps([(o['test'],o['seconds'],o['kills'],o['unique_kills'],o['verdict'],o['subsumed_by']) for o in unit],indent=2))

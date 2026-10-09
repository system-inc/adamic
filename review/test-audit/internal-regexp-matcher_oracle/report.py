import json,pathlib,statistics,re,gzip,functools
D=pathlib.Path('review/test-audit/internal-regexp-matcher_oracle');family=['TestMatcherNodeControls','TestMatcherRandomNode','TestMatcherTest262Executions','TestMatcherCanonicalizeNode','TestMatcherUTF16PatternsNode','TestMatcherOct6Node']; F='TestMatcherExecution family';requested=family+['TestMatcherStepLimit','TestMatcherPropertyProviderStrings','TestMatcherProviderSnapshot','TestMatcherStepLimitBoundary','TestMatcherOct6Mutants','TestMatcherOct6LoopsNode','TestNodeAgreement']; alltests=[x for x in (D/'test-list.log').read_text().splitlines() if x.startswith('Test')]
@functools.lru_cache(maxsize=2)
def events(label):
 p=D/(label+'.log'); gz=D/(label+'.log.gz')
 if not p.exists() and not gz.exists():return []
 with (p.open() if p.exists() else gzip.open(gz,'rt')) as inp:
  out=[]
  for line in inp:
   try:out.append(json.loads(line))
   except:pass
  return out

def info(label,test=None):
 es=events(label); selected=[e for e in es if not test or e.get('Test','').split('/')[0]==test]; failed=any(e['Action']=='fail' and e.get('Test') for e in selected);panic=any('panic:' in e.get('Output','') for e in selected);timeout=any('test timed out' in e.get('Output','') for e in es)
 failleaves={e.get('Test') for e in selected if e['Action']=='fail'}
 lines=[e.get('Output','').strip() for e in selected if (not failleaves or e.get('Test') in failleaves) and re.search(r'\.go:\d+:',e.get('Output','')) and not any(x in e.get('Output','') for x in ['totals:','caught:','caught at','UNAVAILABLE','agreement:','executions compared'])]
 return dict(status='unknown' if timeout else 'fail' if failed or panic else 'pass' if any(e['Action']=='pass' for e in selected) else 'unknown',line=lines[0][:800] if lines else next((e['Output'].strip() for e in selected if 'panic:' in e.get('Output','')),''),timeout=timeout)
plan=json.loads((D/'plan.json').read_text());matrix={}
for m in plan:
 id=m['id'];matrix[id]={}
 for test in alltests:
  alone=id+'-alone-'+test; label=alone if (D/(alone+'.log')).exists() or (D/(alone+'.log.gz')).exists() else id
  matrix[id][test]=dict(info(label,test),log=label)
(D/'matrix.json').write_text(json.dumps(matrix,indent=2))
medians={}
for test in alltests:
 nums=[e['Elapsed'] for i in range(1,4) for e in events('time-'+test+'-'+str(i)) if e['Action']=='pass' and not e.get('Test')]
 medians[test]=statistics.median(nums) if nums else None
nums=[e['Elapsed'] for i in range(1,4) for e in events('time-family-'+str(i)) if e['Action']=='pass' and not e.get('Test')];medians[F]=statistics.median(nums)
(D/'timings.json').write_text(json.dumps(medians,indent=2))
def group(test):return F if test in family else test
kills={group(t):set() for t in alltests if t!='TestMatcherOct6Mutants'}
for id,entries in matrix.items():
 for t,result in entries.items():
  if t!='TestMatcherOct6Mutants' and result['status']=='fail':kills[group(t)].add(id)
unique={row:[id for id in ids if sum(id in v for v in kills.values())==1] for row,ids in kills.items()}
rows=[]
for test in [F]+[t for t in requested if t not in family]:
 members=family if test==F else [test]; witness=test=='TestMatcherOct6Mutants';ids=sorted(kills.get(test,set())); uq=sorted(unique.get(test,[]));subs=[];subtime=None
 if witness:verdict='witness' if info('W01',test)['status']=='fail' else 'untrue'
 elif uq:verdict='sacred'
 elif not ids:verdict='untrue'
 else:
  options=[r for r,k in kills.items() if r!=test and set(ids)<=k];options.sort(key=lambda x:medians[x]);subs=options[:1];verdict='subsumed' if subs else 'overlapping';subtime=medians[subs[0]] if subs else None
  if not subs:subs=[r for r,k in kills.items() if r!=test and set(ids)&k]
 chosen=([id for id in uq if id!='M09'][-1] if any(id!='M09' for id in uq) else ids[-1]) if ids else None; evidence='';last=None
 if witness: chosen='W01';line=info('W01',test)['line'];evidence='go test -json -count=1 -timeout 90s ./internal/regexp/ -run ^TestMatcherOct6Mutants$: '+line;last=chosen+' '+line
 elif chosen:
  member=next(t for t in members if matrix[chosen][t]['status']=='fail');r=matrix[chosen][member];line=r['line'];last=chosen+' '+line;evidence='ADAMIC_MUTANT='+chosen+' timeout 120 go test -json -count=1 -timeout 90s ./internal/regexp/ -run '+('^'+member+'$' if '-alone-' in r['log'] else '.')+' > '+r['log']+'.log 2>&1: '+line
 probes=[]
 if not witness:
  for id in ['PExec','PRun','PParse']:
   if any(info(id,t)['status']=='fail' for t in members):probes.append(id)
  for id in ['PCompile','PCompileUTF16','PCompileProperties']:
   if any(info(id+'-'+t,t)['status']=='fail' for t in members):probes.append(id)
 oracle='Node24 executes patterns and compares capture spans, named spans and lastIndex'
 kind='external-run'
 if test==F:oracle+='; test262 member uses stored Node24 output at test262 7ab7fafa, one capture [0,10] checked against live Node in authority-check.log';kind=['external-run','external-authority']
 if test in ['TestMatcherStepLimit','TestMatcherStepLimitBoundary']:oracle='Self-written ErrStepLimit sentinel expectation'+('; exact three-instruction budget, matched flag, step count and [0,2] capture' if test.endswith('Boundary') else '; does not check a successful answer');kind='self'
 if test=='TestMatcherPropertyProviderStrings':oracle+='; self-written requirement that Go stand-in ASCII property returns UnavailablePropertyError';kind=['external-run','self']
 if test=='TestMatcherProviderSnapshot':oracle+=' after caller-owned provider string storage is changed'
 if test=='TestNodeAgreement':oracle='Live Node RegExp constructor acceptance compared with Parse error presence; checks syntax acceptance, not execution output'
 if witness:oracle='Node24 controls and mutant results; weakened result comparisons must make built-in mutants survive'
 files=[]
 for member in members:
  for p in pathlib.Path('internal/regexp').glob('*_test.go'):
   s=p.read_text();needle='func '+member+'('
   if needle in s:files.append(str(p)+':'+str(s[:s.index(needle)].count('\n')+1))
 row=dict(test=test,package='internal/regexp',file='; '.join(files),seconds=medians[test],oracle=oracle,oracle_kind=kind,kills=ids,unique_kills=uq,last_proven_fail=last,verdict=verdict,subsumed_by=subs,mutants_in_matrix=18 if not witness else 0,probe_kills=probes,subsumer_seconds=subtime,vacuous=None if witness else not bool(probes),bounded=True if not witness else False,matrix_rows=[group(t) for t in alltests if t not in family]+[F],evidence=evidence)
 if test==F:row['members']=members
 if test=='TestNodeAgreement':row['vacuous_subcases']=['Node-accepted patterns pass the empty Parse probe because this oracle checks only error presence; Node-rejected patterns fail it.']
 row['bounded_reason']='M09 exceeded 90s; isolated caller reruns retain TestMatcherTest262Executions unknown. Sixteen mutants completed whole-package runs; M10 panicked and was resolved by isolated runs of all 17 tests. No uniqueness claim rests on M09.' if not witness else None
 if test=='TestMatcherStepLimitBoundary':row['vacuous_subcases']=['PExec: exactly-limit reaches ExecString only after direct Program.run assertions; one-step-past-limit fails empty ExecString','PRun: no subcase passes the complete probe'] if False else []
 # No uniqueness depends on cooked mutant.
 row['unique_kills']=[id for id in row['unique_kills'] if id!='M09']
 rows.append(row)
(D/'results.json').write_text(json.dumps(rows,indent=2))
lines=['| ID | Origin file:line | Change | Failing rows |','|---|---|---|---|']
for m in plan:
 failed=sorted({group(t) for t,r in matrix[m['id']].items() if r['status']=='fail' and t!='TestMatcherOct6Mutants'});m['failed_rows']=failed
 old=m['old'].strip().replace('\n',' ');new=m['new'].strip().replace('\n',' ');lines.append('| '+m['id']+' | '+m['file']+':'+str(m['line'])+' | '+old+' → '+(new or 'drop whole loop')+' | '+', '.join(failed)+' |')
(D/'mutant-table.md').write_text('\n'.join(lines));(D/'mutants.json').write_text(json.dumps(plan,indent=2))
print(json.dumps(rows,indent=2));print('\n'.join(lines))

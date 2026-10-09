import pathlib,json,statistics,re,shutil
p=pathlib.Path('review/test-audit/internal-oracle-census_overload_result');names=['TestCensusAppendResultProof','TestCensusOverloadResultStop','TestCensusSmallStoppedSourceOnNode'];family='TestNativeAgreesWithNode family (append and lie inputs only)';matrix_rows=names+[family];items=json.loads((p/'manifest.json').read_text());commands=json.loads((p/'matrix-commands.json').read_text());bylabel={c['label']:c for c in commands}
def events(label):
 out=[]
 for l in (p/(label+'.log')).read_text().splitlines():
  try:out.append(json.loads(l))
  except:pass
 return out
def binary(label):return next((e['Elapsed'] for e in reversed(events(label)) if e.get('Action') in ['pass','fail'] and not e.get('Test')),None)
def med(n):return statistics.median(binary(n+'.'+str(i)) for i in range(1,4))
matrix=[];kills={n:[] for n in matrix_rows};unique={n:[] for n in matrix_rows};proof={};diagnostics=[]
for x in items:
 id=x['id']
 if id.startswith('P'):continue
 primary=events(id+'.primary');shared=events(id+'.family');fail={e['Test'] for e in primary if e.get('Action')=='fail' and e.get('Test') in names}
 if any(e.get('Action')=='fail' and e.get('Test')=='TestNativeAgreesWithNode' for e in shared):fail.add(family)
 obs={n:('fail' if n in fail else 'pass' if any(e.get('Action')=='pass' and e.get('Test')==n for e in primary) else 'unknown') for n in names};obs[family]='fail' if family in fail else ('pass' if any(e.get('Action')=='pass' and e.get('Test')=='TestNativeAgreesWithNode' for e in shared) else 'unknown')
 for n in fail:
  kills[n].append(id)
  if n in names:
   line=next(e['Output'].strip() for e in primary if e.get('OutputType')=='error' and e.get('Test','').split('/')[0]==n);proof[n]=(id,line)
 if len(fail)==1:unique[next(iter(fail))].append(id)
 for e in primary+shared:
  if e.get('OutputType')=='error':
   d=dict(mutant=id,test=e.get('Test'),line=e['Output'].strip())
   for field in ['stdout','stderr']:
    m=re.search(field+r':\[\]uint8\{([^}]*)\}',d['line'])
    if m:d[field]=bytes(int(v.strip(),0) for v in m[1].split(',') if v.strip()).decode('utf-8','replace')
   diagnostics.append(d)
 matrix.append(dict(id=id,bounded=True,matrix_rows=matrix_rows,failed_rows=sorted(fail),observations=obs,outside_scope='unknown',survivor=not fail))
(p/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n');(p/'diagnostics.json').write_text(json.dumps(diagnostics,indent=2)+'\n')
rows=[]
oracles=['Handwritten absence of strings containing of append result:. No positive IR-content or execution assertion; an empty IR program passes P02.', 'Untouched source actually runs on Node and must match handwritten called/undefined stdout. Native and emitted JavaScript must match handwritten stdout, exact panic stderr and exit 70. These checked backend results deliberately differ from the original source.', 'Untouched stopped-source fixtures actually run on Node and must match handwritten stdout, zero exit and empty stderr. No Adamic lowering or backend is invoked.']
for i,n in enumerate(names):
 if i<2:
  id,line=proof[n];evidence=bylabel[id+'.primary']['command']+'; '+line;last=id+' '+line;probe_kills=[id for id in ['P01','P02'] if any(e.get('Action')=='fail' and e.get('Test')==n for e in events(id+'.'+n))];vacuous=any(e.get('Action')=='pass' and e.get('Test')==n for e in events('P02.'+n));verdict='sacred' if unique[n] else 'untrue'
 else:last=None;evidence='Bounded uncached baseline and three isolated runs pass. Production mutants do not apply: this row only invokes the external Node oracle on untouched fixtures.';probe_kills=[];vacuous=None;verdict='cannot-judge'
 r=dict(test=n,package='internal/oracle',file='internal/oracle/'+('census_overload_result_test.go:'+str([21,40][i]) if i<2 else 'census_small_test.go:34'),seconds=med(n),oracle=oracles[i],oracle_kind='self' if i==0 else ['external-run','self'] if i==1 else 'external-run',kills=kills[n],unique_kills=unique[n],last_proven_fail=last,verdict=verdict,subsumed_by=[],mutants_in_matrix=9,probe_kills=probe_kills,subsumer_seconds=None,vacuous=vacuous,bounded=True,matrix_rows=matrix_rows,evidence=evidence,timing_samples=[binary(n+'.'+str(j)) for j in range(1,4)])
 if i==0:r.update(empty_answer_probe='P02',nil_probe_failed_by_dereference=True,empty_ir_passes=True)
 if i==1:r['empty_answer_probe']='P02'
 if i==2:r['reason']='No Adamic production entry is reached. A meaningful production mutant cannot affect the asserted observation; mutating Node or its reference input would violate the oracle restriction. Not labeled untrue.'
 rows.append(r)
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
md=['| ID | Origin file:line | Change | Failed rows |','|---|---|---|---|']
for x,m in zip(items,matrix):
 if x['id'].startswith('P'):continue
 md.append('| '+x['id']+' | '+x['file']+':'+str(x['line'])+' | '+(x['old']+' -> '+x['new']).strip().replace('\n',' ').replace('|','\\|')+' | '+', '.join(m['failed_rows'])+' |')
(p/'mutants.md').write_text('\n'.join(md)+'\n')
base=json.loads((p/'baseline-commands.json').read_text());valid=json.loads((p/'standalone-validation.json').read_text());surv=json.loads((p/'survivor-commands.json').read_text());ctl=json.loads((p/'restored-control.json').read_text());timings=dict(setup_seconds=0,nproc=5,whole_package_baseline_binary_seconds=binary('baseline'),whole_package_baseline_timed_out=True,bounded_baseline_binary_seconds=binary('bounded-baseline'),isolated_timing_and_coverage_command_seconds=sum(x['wall'] for x in base),matrix_and_controls_command_seconds=sum(c['wall'] for c in commands),matrix_and_controls_binary_seconds=sum(binary(c['label']) or 0 for c in commands),standalone_vet_seconds=sum(x['seconds'] for x in valid),independent_probe_build_seconds=surv[0]['wall'],restored_control_command_seconds=ctl['wall'],restored_control_exit=ctl['exit']);(p/'timings.json').write_text(json.dumps(timings,indent=2)+'\n')
for src in ['plan','matrix','baseline','vet','report']:shutil.copyfile('/tmp/u056-'+src+'.py',p/(src+'.py'))
print(json.dumps(rows,indent=2));print(json.dumps(timings,indent=2))

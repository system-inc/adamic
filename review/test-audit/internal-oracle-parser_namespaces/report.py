import pathlib,json,statistics,re,shutil
p=pathlib.Path('review/test-audit/internal-oracle-parser_namespaces');scope=json.loads((p/'scope.json').read_text());rows=scope['rows'];family=scope['family'];witnesses=scope['witnesses'];shared='TestNativeAgreesWithNode family (eight selected inputs)';prod=[rows[0]]+rows[-3:];matrix_rows=prod+[shared];items=json.loads((p/'manifest.json').read_text());commands=json.loads((p/'matrix-commands.json').read_text());bylabel={c['label']:c for c in commands}
def events(label):
 out=[]
 for l in (p/(label+'.log')).read_text().splitlines():
  try:out.append(json.loads(l))
  except:pass
 return out
def binary(label):return next((e['Elapsed'] for e in reversed(events(label)) if e.get('Action') in ['pass','fail'] and not e.get('Test')),None)
def med(n):return statistics.median(binary(n.replace(' ','_')+'.'+str(i)) for i in range(1,4))
def group(n):return rows[0] if n in family else n
kills={n:[] for n in matrix_rows};unique={n:[] for n in matrix_rows};proof={};matrix=[];diagnostics=[]
for x in items:
 id=x['id']
 if not id.startswith('M'):continue
 es=events(id+'.primary');ss=events(id+'.family');fail={group(e['Test']) for e in es if e.get('Action')=='fail' and e.get('Test') and '/' not in e['Test']}
 if any(e.get('Test')=='TestNativeAgreesWithNode' and e.get('Action')=='fail' for e in ss):fail.add(shared)
 panic=any(e.get('Output','').lstrip().startswith('panic:') for e in es);obs={}
 for n in prod:
  members=family if n==rows[0] else [n];status=[e.get('Action') for e in es if e.get('Test') in members and e.get('Action') in ['pass','fail']];obs[n]='fail' if n in fail else 'pass' if len(status)==len(members) and all(s=='pass' for s in status) else 'unknown'
  if n in fail:
   kills[n].append(id);line=next(e['Output'].strip() for e in es if e.get('OutputType')=='error' and group(e.get('Test','').split('/')[0])==n);proof[n]=(id,line)
 if shared in fail:kills[shared].append(id)
 obs[shared]='fail' if shared in fail else 'pass' if any(e.get('Action')=='pass' and e.get('Test')=='TestNativeAgreesWithNode' for e in ss) else 'unknown'
 if len(fail)==1:unique[next(iter(fail))].append(id)
 for e in es+ss:
  if e.get('OutputType')=='error':diagnostics.append(dict(id=id,test=e.get('Test'),line=e['Output'].strip()))
 matrix.append(dict(id=id,bounded=True,matrix_rows=matrix_rows,failed_rows=sorted(fail),observations=obs,go_panic=panic,outside_scope='unknown',survivor=not fail))
(p/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n');(p/'diagnostics.json').write_text(json.dumps(diagnostics,indent=2)+'\n')
secs={n:med(n) for n in rows};sharedsec=statistics.median(binary('family-baseline.'+str(i)) for i in range(1,4));oracles={rows[0]:'Untouched source runs on Node; sanitized native and emitted JavaScript must match stdout, stderr and exit, plus sanitizer leak check.',rows[-3]:'Own counts.md snapshot plus internal site/direction validity and aggregate consistency. No independent authority.',rows[-2]:'Untouched source runs on Node against handwritten stdout; handwritten refusal/pending classes and path-bearing diagnostic substrings, or exact backend panic stdout/stderr/exit and aggregate checked counts.',rows[-1]:'Untouched source Node exit 0 and empty stderr only, with no stdout assertion; handwritten Refused class and complete own .refused diagnostic snapshot.'};result=[]
files={rows[0]:'parser_namespaces_test.go:29',witnesses[0]:'parser_namespaces_test.go:44',witnesses[1]:'parser_namespaces_test.go:131',witnesses[2]:'parser_namespaces_test.go:166',rows[-3]:'predicate_counts_test.go:81',rows[-2]:'predicate_refusals_test.go:32',rows[-1]:'presence_refused_test.go:28'}
# Resolve actual starting lines directly; tests have only the allowed disagreement switch elsewhere.
for n in rows:
 target=family[0] if n==rows[0] else n
 for f in pathlib.Path('internal/oracle').glob('*_test.go'):
  for k,l in enumerate(f.read_text().splitlines(),1):
   if l.startswith('func '+target+'('):files[n]=f.name+':'+str(k)
for n in rows:
 if n in witnesses:
  label='W01.'+n;es=events(label);failed=any(e.get('Action')=='fail' and e.get('Test')==n for e in es);line=next((e['Output'].strip() for e in es if e.get('OutputType')=='error'),'No failure');id='W01';verdict='witness' if failed else 'untrue';one=dict(kills=[],unique_kills=[],probe_kills=[],vacuous=None,subsumed_by=[],subsumer_seconds=None,mutants_in_matrix=0,matrix_rows=witnesses,witness_edits=['W01']);oracle='Node observation and the agreement checker applied to a built-in IR mutation. Only W01 comparison weakening determines the verdict.';kind=['external-run','self'] if n==witnesses[1] else 'external-run';last=id+' '+line;evidence=bylabel[label]['command']+'; '+line
 else:
  id,line=proof[n];verdict='sacred' if unique[n] else 'subsumed' if set(kills[n])<=set(kills[shared]) else 'untrue';sub=[shared] if verdict=='subsumed' else [];pk=[]
  for pid in ['P01','P02']:
   es=events(pid+'.'+n.replace(' ','_'));members=family if n==rows[0] else [n]
   if any(e.get('Action')=='fail' and e.get('Test') in members for e in es):pk.append(pid)
  es=events('P02.'+n.replace(' ','_'));members=family if n==rows[0] else [n];vacuous=all(any(e.get('Action')=='pass' and e.get('Test')==m for e in es) for m in members);one=dict(kills=kills[n],unique_kills=unique[n],probe_kills=pk,vacuous=vacuous,subsumed_by=sub,subsumer_seconds=sharedsec if sub else None,mutants_in_matrix=14,matrix_rows=matrix_rows,empty_answer_probe='P02');oracle=oracles[n];kind='external-run' if n==rows[0] else 'self' if n==rows[-3] else ['external-run','self'];last=id+' '+line;evidence=bylabel[id+'.primary']['command']+'; '+line
 r=dict(test=n,package='internal/oracle',file='internal/oracle/'+files[n],seconds=secs[n],oracle=oracle,oracle_kind=kind,last_proven_fail=last,verdict=verdict,bounded=True,evidence=evidence,**one)
 if n==rows[0]:r['members']=family;r['subsumption_mutants']=len(kills[n])
 result.append(r)
(p/'rows.json').write_text(json.dumps(result,indent=2)+'\n');md=['| ID | Origin file:line | Change | Failed rows |','|---|---|---|---|']
for x,m in zip(items,matrix):
 if not x['id'].startswith('M'):continue
 change=x['old']+' -> '+x['new']
 if x['mode']=='early':change='return value at entry'
 md.append('| '+x['id']+' | '+x['file']+':'+str(x['line'])+' | '+change.strip().replace('\n',' ').replace('|','\\|')+' | '+', '.join(m['failed_rows'])+' |')
(p/'mutants.md').write_text('\n'.join(md)+'\n')
base=json.loads((p/'baseline-commands.json').read_text());v=json.loads((p/'standalone-validation.json').read_text());surv=json.loads((p/'survivor-commands.json').read_text());ctl=json.loads((p/'restored-control.json').read_text());extra=[c for c in commands if c['id'].startswith('M') and c['label'] not in [c['id']+'.primary',c['id']+'.family']];t=dict(setup_seconds=0,nproc=5,whole_baseline_binary_seconds=binary('baseline'),bounded_baseline_binary_seconds=binary('bounded-baseline'),timing_and_coverage_command_seconds=sum(c['wall'] for c in base),matrix_controls_probes_witness_command_seconds=sum(c['wall'] for c in commands),matrix_controls_probes_witness_binary_seconds=sum(binary(c['label']) or 0 for c in commands),unnecessary_isolated_repeat_command_seconds=sum(c['wall'] for c in extra),standalone_vet_seconds=sum(c['seconds'] for c in v),independent_compiler_build_seconds=surv[0]['wall'],restored_control_command_seconds=ctl['wall'],restored_control_exit=ctl['exit']);(p/'timings.json').write_text(json.dumps(t,indent=2)+'\n')
for src in ['plan','baseline','matrix','vet','report','survivors','witnesses']:shutil.copyfile('/tmp/u067-'+src+'.py',p/(src+'.py'))
print(json.dumps([(r['test'],r['verdict'],r['kills'],r['unique_kills'],r['probe_kills'],r['vacuous']) for r in result],indent=2));print(json.dumps(t,indent=2))

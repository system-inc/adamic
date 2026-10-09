import pathlib,json,subprocess,re,difflib,statistics
root=pathlib.Path('/workspace/adamic');out=root/'review/test-audit/stage1-cohere-lint-harness';base='16f436a16a8e3b9cd2343f449eb1b0b4577c135c'
plan=json.loads((out/'plan.json').read_text())+json.loads((out/'probe-plan.json').read_text());timings=json.loads((out/'timings.json').read_text())
shards=[f'TestJsxLintTrees_{i:03}' for i in range(16)]
group=lambda t:'TestJsxLintTrees family' if t in shards else 'TestProduct_jsx_oracle family' if t in ['TestProduct_jsx_membership','TestProduct_jsx_parser'] else t
sources={m['file']:subprocess.check_output(['git','show',base+':'+m['file']],cwd=root,text=True) for m in plan}
matrix=[]
for m in plan:
 original=sources[m['file']];changed=original.replace(m['old'],m['new'],1)
 if m['id']=='P3':
  for imp in ['github.com/system-inc/adamic/internal/load','github.com/system-inc/adamic/internal/lower']:changed=changed.replace('\t"'+imp+'"\n','')
 if m['id'] in ['P8','P9']:
  for imp in ['go/ast','go/parser','go/token']+(['strconv'] if m['id']=='P9' else []):changed=changed.replace('\t"'+imp+'"\n','')
 mappings={}
 for tag,a,b,c,d in difflib.SequenceMatcher(a=original.splitlines(),b=changed.splitlines(),autojunk=False).get_opcodes():
  if tag=='equal':
   for k in range(d-c):mappings[c+k+1]=a+k+1
 observed=[];failed=[];outputs={};logs=[m['id']+'.log']
 if m['id'] in ['M1','M2','M3']:logs.append(m['id']+'-isolation.log')
 for log in logs:
  for line in (out/log).read_text().splitlines():
   try:r=json.loads(line)
   except:continue
   t=r.get('Test','')
   if r.get('Action')=='run' and t and '/' not in t:observed.append(t)
   if r.get('Action')=='fail' and t and '/' not in t:failed.append(t)
   if r.get('Action')=='output' and t:
    for s in r.get('Output','').splitlines():
     location=re.match(r'\s*(\w+\.go):(\d+): (.*)',s)
     if location:
      file,n,message=location.groups();n=int(n)
      if file==pathlib.Path(m['file']).name:n=mappings.get(n,n)
      outputs.setdefault(group(t),[]).append(dict(origin_file='stage1/cohere/lint/'+file,origin_line=n,message=message,log=log))
     elif 'panic:' in s:outputs.setdefault(group(t),[]).append(dict(origin_file='stage1/cohere/lint/jsx_shards_test.go',origin_line=391,message=s,log=log))
 if m['id']=='P12':failed=['TestJsxLintTreesShardCoverage']
 matrix.append(dict(id=m['id'],kind=m['kind'],file=m['file'],line=m['line'],change=m['change'],command=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/','-run',m['regex']],logs=logs,matrix_rows=sorted(set(map(group,observed))),failed_rows=sorted(set(map(group,failed))),failed_functions=sorted(set(failed)),outputs=outputs))
(out/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n')
median=lambda name:round(statistics.median(x['seconds'] for x in timings if x['test']==name),3)
data=[
('TestEmittedJavaScriptMismatch','harness_test.go','setup-check','self','Handwritten shard declaration census and exactly-once assignment; this row now checks construction.',['H1'],'P8'),
('TestEmittedJavaScriptMismatch_Setup','harness_test.go','setup-check','self','Successful immutable product construction and nonempty oracle path; no runtime output comparison.',['H5'],'P7'),
('TestCompleteSuggestionSerialization','harness_test.go','witness','self','Handwritten exactly-once census and synthetic planted disagreement rejected by completeSuggestionEqual; this row is a union/witness, not a runtime serialization comparison.',['W1'],None),
('TestSuggestionAlongsideAutomaticFix','harness_test.go','setup-check','self','Handwritten AST census: one parallel call and one literal case index per shard; runtime comparisons moved into numbered leaves.',['H2'],'P9'),
('TestJsxLintReleaseAndThroughput','jsx_integration_test.go','cannot-judge',['external-run','self'],'Go cohere full finding comparison and Go/Node/native count equality; nonzero count and a handwritten JSX census. Both clean attempts timed out before completion.',[],None),
('TestJsxLintTrees family','jsx_integration_test.go','subsumed','external-run','Go TypeScript whole-tree bytes compared with Node running the original parser port source and sanitized native output.',['M2','M3'],'P1'),
('TestProduct_jsx_oracle family','jsx_products_test.go','setup-check','self','Shared jsxGoOracle constructor with two input recipes. Checks command errors, but does not inspect the returned executable; H6 accepted absent oracle files.',['H7'],'P2'),
('TestProduct_jsx_tree_lowered','jsx_products_test.go','untrue','self','Successful lowering/write return only; H8 wrote empty.c instead of program.c and the row passed. No planted valid construction mutant caught by this row. One construction attempt, limited verdict.',[],'P3'),
('TestProduct_jsx_tree_native','jsx_products_test.go','untrue','self','Successful native build return only; H9 disabled sanitization and passed with __asan_init absent. No valid planted construction mutant caught by this row. One construction attempt, limited verdict.',[],'P4'),
('TestProduct_jsx_tree_setup','jsx_products_test.go','setup-check','self','Shared setup constructor and its nonempty directory guard; no explicit bundle consumption.',['H4'],'P5'),
('TestJsxLintTreesSetupIsolation','jsx_setup_isolation_test.go','setup-check',['external-run','self'],'Cold standalone shard must pass and log named build misses. Its child also compares Go/Node/native trees, so observed M2/M3 failures count against uniqueness; H4 proves the construction check.',['M2','M3','H4'],'P5'),
('TestJsxLintTreesShardCoverage','jsx_shards_test.go','setup-check','self','Handwritten stable ownership, relocated identity, invalid union and shard-selection invariants.',['H3'],'P10'),
('TestJsxLintTrees_Setup','jsx_shards_test.go','setup-check','self','Shared setup constructor and its nonempty directory guard.',['H4'],'P5'),
('TestJsxLintTreesUnion','jsx_shards_test.go','setup-check','self','AST declaration census and exact live bundle union. Kept separate because it asserts declaration and coverage facts absent from runtime shards.',['H4'],'P6'),
('TestJsxLintTreesShardDisagreement','jsx_shards_test.go','witness','self','Synthetic planted tree disagreement must fail only its owning child shard and include the planted message.',['W2'],None),
]
rows=[]
for test,file,verdict,kind,oracle,kills,probe in data:
 ran=[m for m in matrix if test in m['matrix_rows'] and m['kind']!='probe'];ps=[m for m in matrix if m['kind']=='probe' and test in m['matrix_rows']]
 relevant=[m for m in matrix if m['id'] in kills];last=relevant[-1] if relevant else None
 failure=None
 if last:
  candidates=last['outputs'].get(test,[])
  needles={'W1':'planted disagreement caught','W2':'planted disagreement survived','H4':'shared JSX preparation failed','H7':'build jsx-','M2':'Node:','M3':'Node:','H1':'shard declarations','H2':'shard enumeration','H3':'empty corpus','H5':'setup did not complete'}
  failure=next((x for x in candidates if needles.get(last['id'],'') in x['message']),candidates[0] if candidates else None)
 if failure:evidence=' '.join(last['command'])+'; '+failure['log']+'; '+failure['origin_file']+':'+str(failure['origin_line'])+': '+failure['message']
 else:evidence=('baseline-benchmark-split.log: panic: test timed out after 1m30s' if verdict=='cannot-judge' else '; '.join(m['id']+'.log: selected row passed' for m in ran))
 row=dict(test=test,package='stage1/cohere/lint',file='stage1/cohere/lint/'+file,seconds=None if verdict=='cannot-judge' else median(test),oracle=oracle,oracle_kind=kind,kills=kills,unique_kills=[],last_proven_fail=None if not failure else last['id']+' '+failure['message'],verdict=verdict,subsumed_by=['TestJsxLintTreesSetupIsolation'] if verdict=='subsumed' else [],mutants_in_matrix=len(ran),probe_kills=[m['id'] for m in ps if test in m['failed_rows']],subsumer_seconds=median('TestJsxLintTreesSetupIsolation') if verdict=='subsumed' else None,vacuous=None if not probe else test not in next(m['failed_rows'] for m in matrix if m['id']==probe),bounded=True,matrix_rows=sorted(set(t for m in ran for t in m['matrix_rows'])),evidence=evidence)
 if test=='TestJsxLintTrees family':row['members']=shards;row['subsumption_mutants']=2
 if test=='TestProduct_jsx_oracle family':row['members']=['TestProduct_jsx_membership','TestProduct_jsx_parser']
 if test=='TestJsxLintTreesShardCoverage':row['vacuous_subcases']=['P10/P11: positive ownership stability and relocated identity checks before empty-corpus rejection']
 rows.append(row)
(out/'rows.json').write_text(json.dumps(rows,indent=2)+'\n')
listed=set((out/'list.log').read_text().splitlines());scope=[]
for row in rows:scope+=row.get('members',[row['test']])
assert len(scope)==31 and all(t in listed for t in scope)
metadata=dict(unit='u109',base=base,branch='test-audit/stage1-cohere-lint-harness',nproc=5,requested_functions=31,grouped_rows=15,missing=[],scope=scope,skipped_requested=[],over_budget=['TestJsxLintReleaseAndThroughput'],ordinary_mutants=['M1','M2','M3'],construction_mutants=['H1','H2','H3','H4','H5','H6','H7','H8','H9'],weakened_checks=['W1','W2'],probes=['P'+str(i) for i in range(1,13)])
(out/'metadata.json').write_text(json.dumps(metadata,indent=2)+'\n')
# Compact final-use JSON keeps all required fields in rows.json; this is a presentation copy.
compact=[]
for r in rows:
 compact.append({k:v for k,v in r.items() if k not in ['matrix_rows','members','evidence','oracle','vacuous_subcases']})
(out/'compact-rows.json').write_text(json.dumps(compact,indent=2)+'\n')
print('Rows',len(rows),'scope',len(scope),'mutants',len(matrix),'timing wall',sum(r['wall'] for r in timings))

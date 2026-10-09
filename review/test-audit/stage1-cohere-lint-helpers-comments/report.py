import pathlib,json,re,statistics,subprocess
root=pathlib.Path('/workspace/adamic');out=root/'review/test-audit/stage1-cohere-lint-helpers-comments';pkg='stage1/cohere/lint/helpers/comments'
def events(file):
 result=[]
 for l in file.read_text().splitlines():
  try:result.append(json.loads(l))
  except:pass
 return result
runs=json.loads((out/'runs.json').read_text());runmap={r['id']:r for r in runs};matrix=['TestCommentsMatchCohere','TestConsumerCommentHelpers','TestJsxParserGapIsExplicit'];kills={r:[] for r in matrix};probes={r:[] for r in matrix};mutants=[]
for id in ['M1','M2','M3','M4','P1']:
 ev=events(out/(id+'.log'));failed=sorted({e['Test'].split('/')[0] for e in ev if e['Action']=='fail' and 'Test'in e});
 for r in failed:
  (kills if id[0]=='M' else probes)[r].append(id)
 mutants.append({'id':id,'failed':failed,'bounded':True,'matrix_rows':matrix})
def failure(id,row):
 ev=events(out/(id+'.log'));items=[e.get('Output','').strip() for e in ev if e.get('Test','').split('/')[0] in row and 'Output'in e];
 useful=[s for s in items if ('got ' in s or 'survived' in s or 'did not catch' in s or 'union count' in s or 'enumerated ' in s or 'refusalcheck' in s or 'refusal check' in s or 'refuse ' in s or 'gap silently accepted' in s)]
 return (useful[0] if useful else next((s for s in items if '--- FAIL' in s),'no failing diagnostic recorded'))
def median(regex):
 times=[]
 for i in range(1,4):
  ev=events(out/('timing-'+regex+'-'+str(i)+'.log'))
  lines=[e['Output'] for e in ev if e['Action']=='output' and 'Output'in e and e['Output'].startswith('ok')];assert lines
  times.append(float(re.search(r'\t([0-9.]+)s',lines[-1])[1]))
 return statistics.median(times),times
rows=[];specs=[('TestCommentMutants family','TestCommentMutants_[0-9]+','comment_mutants_test.go:24','W1','witness','Live Go cohere complete byte output; forced-equality comparison rejects each compiled built-in mutant.','external-run'),('TestCommentMutantsUnion','TestCommentMutantsUnion','comment_mutants_test.go:140','S1','setup-check','Self-written independent expected and actual shard-ID sets; checks identities and count, not only count.','self'),('TestCommentMutantsPlantedFailure','TestCommentMutantsPlantedFailure','comment_mutants_test.go:190','W2','witness','Self-written child failure protocol checks exit, named test failure and survived diagnostic, with a deliberately equal answer only in shard 002.','self'),('TestCommentsMatchCohere','TestCommentsMatchCohere','comments_test.go:91',None,None,'Live Go cohere full byte output compared with Node and sanitized native port on 47 boundary witnesses.','external-run'),('TestConsumerCommentHelpers','TestConsumerCommentHelpers','comments_test.go:104',None,None,'Live Go cohere full byte output; Go parser provides AST spans, Node and native port compute comment answers for consumer corpus.','external-run'),('TestJsxParserGapIsExplicit','TestJsxParserGapIsExplicit','comments_test.go:121',None,None,'Live Go parser validates zero JSX diagnostics; self-written refusal requires exit 70, NotYet label and identical Node/native stderr. Exit alone cannot pass.',['external-run','self']),('TestJsxAdapterGuardMutant','TestJsxAdapterGuardMutant','comments_test.go:166','W3','witness','Self-written refusal predicate must reject a native guard-drop mutant that exits normally, emits comments, and has empty stderr.','self')]
for name,regex,file,w,verdict,oracle,kind in specs:
 seconds,times=median(regex);kk=kills.get(name,[]);unique=[id for id in kk if sum(id in x for x in kills.values())==1];subs=[]
 if verdict is None:
  if unique:verdict='slow-worthy' if seconds>60 else 'sacred'
  elif kk:
   others=[r for r in matrix if r!=name and set(kk)<=set(kills[r])]
   if others:verdict='subsumed';subs=[min(others,key=lambda r:median(r)[0])]
   else:verdict='overlapping';subs=[r for r in matrix if r!=name and set(kk)&set(kills[r])]
  else:verdict='untrue'
 id=w or (kk[-1] if kk else None);members=[f'TestCommentMutants_{i:03}' for i in range(5)] if name.endswith('family') else [name];diag=failure(id,members) if id else None
 row={'test':name,'package':pkg,'file':pkg+'/'+file,'seconds':seconds,'timing_runs':times,'oracle':oracle,'oracle_kind':kind,'kills':kk,'unique_kills':unique,'last_proven_fail':id+': '+diag if id else None,'verdict':verdict,'subsumed_by':subs,'mutants_in_matrix':4,'probe_kills':probes.get(name,[]) if name in matrix else (['P2'] if name=='TestCommentMutantsUnion' else []),'subsumer_seconds':median(subs[0])[0] if verdict=='subsumed' else None,'vacuous':not bool(probes[name]) if name in matrix else (False if name=='TestCommentMutantsUnion' else None),'bounded':name in matrix,'matrix_rows':matrix if name in matrix else [],'evidence':(runmap[id]['command']+'; '+diag) if id else 'no planted mutant caught','members':members,'verdict_basis':len(kk) if verdict=='subsumed' else None}
 rows.append(row)
(out/'rows.json').write_text(json.dumps(rows,indent=2));(out/'matrix.json').write_text(json.dumps(mutants,indent=2))
plan=json.loads((out/'plan.json').read_text());table=[]
for id,file,old,new,kind in plan['menu']:
 source=subprocess.check_output(['git','show',plan['base']+':'+pkg+'/'+file],cwd=root,text=True);pos=source.index(old);line=source[:pos].count('\n')+1;table.append({'id':id,'file':pkg+'/'+file,'line':line,'old':old.strip(),'new':new.strip() or '(drop statement)','kind':kind,'failed':next(m['failed'] for m in mutants if m['id']==id)})
(out/'mutants.json').write_text(json.dumps(table,indent=2));survivors=[m for m in mutants if m['id'].startswith('M') and not m['failed']];(out/'survivors.json').write_text(json.dumps(survivors,indent=2));print(json.dumps(rows,indent=2));print(json.dumps(table,indent=2))

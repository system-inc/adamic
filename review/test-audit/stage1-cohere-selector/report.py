import pathlib,json,statistics,re,subprocess
root=pathlib.Path('/workspace/adamic');pkg='stage1/cohere/selector';out=root/'review/test-audit/stage1-cohere-selector'
def ev(id):
 result=[]
 for l in (out/(id+'.log')).read_text().splitlines():
  try:result.append(json.loads(l))
  except:pass
 return result
rows=['TestEachGapStandsWhereGapsMdSaysItDoes','TestThePortParsesAsGoCohereDoes','TestSelectorThroughput','TestCorpusKeepsEveryParseableFile','TestTheLibraryDoesNotReturnOnUnconsumedNamespaceBars'];runs=json.loads((out/'runs.json').read_text());cmd={r['id']:r['command'] for r in runs};kills={r:[]for r in rows};matrix=[]
for id in ['M1','M2','M3','M4']:
 failed=sorted({x['Test'].split('/')[0] for x in ev(id) if x['Action']=='fail' and 'Test'in x});matrix.append({'id':id,'failed':failed,'bounded':False,'matrix_rows':rows})
 for r in failed:kills[r].append(id)
def median(r):
 vals=[]
 for i in range(1,4):
  lines=[x['Output']for x in ev('timing-'+r+'-'+str(i))if x['Action']=='output' and x.get('Output','').startswith('ok')];vals.append(float(re.search(r'\t([0-9.]+)s',lines[-1])[1]))
 return statistics.median(vals),vals
def diagnostic(id,row):
 outputs=[x.get('Output','').strip() for x in ev(id) if x.get('Test','').split('/')[0]==row];selected=[s for s in outputs if any(k in s for k in ['differ:','refuses with','checksums differ','cannot see it','lost selectors','parser returned within','stopped before entering','no such file'])];return selected[0] if selected else next((s for s in outputs if 'panic:' in s or '--- FAIL' in s),'no diagnostic')
oracles={rows[0]:('Node executes six fixtures; self-written GAPS.md exact NotYet labels and expected stdout; closed gap also runs sanitized native.',['external-run','self']),rows[1]:('Full canonical output compared with live Go cohere, Node, native, JS backend and postcss-selector-parser 2.2.3; upstream known hangs explicitly excluded.','external-run'),rows[2]:('Only aggregate parsed/attempt/root-child counts: native supplies initial expectation, Node source and live upstream library compare. M1/M3 AST defects survive.',['self','external-run']),rows[3]:('Live PostCSS extracts synthetic CSS; self-written exact selectors, counts, exclusion errors and explicit file-list contract.',['external-run','self']),rows[4]:('Live pinned upstream parser invoked on four inputs; self-written marker and no-return observation for one second. This proves bounded observation, not infinite execution.',['external-run','self'])}
result=[]
for i,r in enumerate(rows):
 kk=kills[r];unique=[m for m in kk if sum(m in k for k in kills.values())==1];subs=[];seconds,times=median(r);id=kk[-1] if kk else ('S1'if i==3 else 'S2')
 if i>=3:verdict='setup-check'
 elif unique:verdict='slow-worthy'if seconds > 60 else 'sacred'
 elif kk:
  others=[s for s in rows if s!=r and set(kk)<=set(kills[s])];subs=[min(others,key=lambda s:median(s)[0])];verdict='subsumed'
 else:verdict='untrue'
 probe='P2'if i==0 else 'P1'if i in [1,2] else 'P3'if i==3 else 'P4';probeid='P2-'+r if probe=='P2'else probe;pe=ev(probeid);failed=any(x['Action']=='fail'and x.get('Test','').split('/')[0]==r for x in pe)or(any('panic:'in x.get('Output','')for x in pe)and not any(x['Action']=='pass'and x.get('Test')==r for x in pe));diag=diagnostic(id,r)
 sourcefile='gaps_test.go' if i==0 else 'selector_test.go';source=subprocess.check_output(['git','show',json.loads((out/'plan.json').read_text())['base']+':'+pkg+'/'+sourcefile],cwd=root,text=True);sourceline=source[:source.index('func '+r+'(')].count('\n')+1
 obj={'test':r,'package':pkg,'file':pkg+'/'+sourcefile+':'+str(sourceline),'seconds':seconds,'timing_runs':times,'oracle':oracles[r][0],'oracle_kind':oracles[r][1],'kills':kk,'unique_kills':unique,'last_proven_fail':id+': '+diag,'verdict':verdict,'subsumed_by':subs,'mutants_in_matrix':4,'probe_kills':[probe]if failed else [],'subsumer_seconds':median(subs[0])[0]if subs else None,'vacuous':not failed,'bounded':False,'matrix_rows':[],'evidence':cmd[id]+'; '+diag}
 if i==1:obj['witness_evidence']=cmd['W1']+'; '+diagnostic('W1',r);obj['vacuous_subcases']=[x['Test']for x in pe if x['Action']=='pass'and x.get('Test','').startswith(r+'/')];obj['probe_scope_note']='Passing upstream-only subcase does not reach the port; compiled-mutant subcases pass despite empty driver.'
 if i==2:obj['vacuous_subcases']=['round 1 native checksum','round 1 Node source checksum'];obj['vacuous_without_library']=any(x['Action']=='pass'and x.get('Test')==r for x in ev('P1-no-library'))
 if i==4:obj['classification_note']='Upstream invocation and exception-observation setup; S2 drops the call without editing the external parser.'
 result.append(obj)
(out/'rows.json').write_text(json.dumps(result,indent=2));(out/'matrix.json').write_text(json.dumps(matrix,indent=2));table=[];plan=json.loads((out/'plan.json').read_text())
for m in plan['mutants']:
 s=subprocess.check_output(['git','show',plan['base']+':'+m['file']],cwd=root,text=True);table.append({**m,'line':s[:s.index(m['old'])].count('\n')+1,'failed':next(x['failed']for x in matrix if x['id']==m['id'])})
(out/'mutants.json').write_text(json.dumps(table,indent=2));(out/'survivors.json').write_text(json.dumps([x for x in matrix if not x['failed']],indent=2));print(json.dumps(result,indent=2))

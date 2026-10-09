from pathlib import Path
import json,re,statistics
p=Path(__file__).parent
names=json.loads((p/'rows.json').read_text());timings=json.loads((p/'timings.json').read_text());runs=json.loads((p/'run-status.json').read_text())
def events(id,n):
 return [json.loads(l) for l in (p/(id+'-'+n+'.log')).read_text().splitlines() if l.startswith('{')]
def fail(id,n):
 ev=events(id,n)
 if not any(x.get('Action')=='fail' and x.get('Test')==n for x in ev):return None
 out=[x['Output'].strip() for x in ev if x.get('OutputType')=='error' and x.get('Test','').split('/')[0]==n]
 if out:return out[0]
 return next(x['Output'].strip() for x in ev if x.get('Output','').startswith('--- FAIL'))
mutants={}
for r in runs:
 mutants.setdefault(r['id'],{})[r['test']]={'exit':r['exit'],'failure':fail(r['id'],r['test']),'wall_seconds':r['wall'],'command':r['command']+' > '+r['id']+'-'+r['test']+'.log 2>&1'}
(p/'matrix.json').write_text(json.dumps(mutants,indent=2))
source={}
for f in Path('stage1/typescript/parser').glob('*test.go'):
 for n in names:
  m=re.search(r'^func '+n+r'\(',f.read_text(),re.M)
  if m:source[n]=str(f)+':'+str(f.read_text()[:m.start()].count('\n')+1)
jsx=[names[0],names[1],names[3],names[4]];perf=names[7:9]
notes={names[0]:'Go rejects with diagnostic 1003; Node/native diagnostic class and exact mutual stderr equality. Location is not compared with Go. M1 passes original refusal at wrong token and fails during built-in permissive parsing.',names[1]:'Go rejects private/escaped JSX names; Node/native refusal markers and exact mutual stderr equality.',names[2]:'Go whole AST bytes; built-in scanner mutations must disagree in Node and sanitized native. Judged by W1, not production precondition failures.',names[3]:'Go whole AST bytes and Go JSX kind inventory coverage, compared with Node port.',names[4]:'Go whole AST bytes compared with sanitized and release native port.',names[5]:'Go expression AST bytes; built-in precedence, optional and arrow mutants must disagree. Judged by W1.',names[6]:'Self: successfully construct pinned corpus, Go oracle and native product. Returned product values are not asserted; S1 drops binary construction and passes.',names[7]:'Go recursive node count compared with Node/native, including warmup and five rotated samples. No speed threshold; equal counts can conceal different trees.',names[8]:'Go full AST preflight plus Go recursive node counts in Node/native. No speed threshold.',names[9]:'Go AST and counts; built-in count mutant must preserve AST but change count. Judged by W2.'}
subs={names[0]:names[3],names[3]:names[4],names[4]:names[3],names[7]:names[8],names[8]:names[7]}
witness={names[2]:'W1',names[5]:'W1',names[9]:'W2'}
result=[]
for n in names:
 ids=[id for id,d in mutants.items() if n in d and id.startswith('M')]
 kills=[id for id in ids if mutants[id][n]['failure']]
 unique=[id for id in kills if sum(bool(x['failure']) for x in mutants[id].values())==1]
 verdict='sacred' if unique else 'subsumed' if kills and n in subs else 'untrue'
 proof=kills[-1] if kills else None;bounded=True;matrix_rows=jsx if n in jsx else perf if n in perf else [n]
 if n in witness:
  matrix_rows=[names[2],names[5]] if witness[n]=='W1' else [n]
  id=witness[n];ids=[id];kills=[id] if mutants[id][n]['failure'] else [];unique=[];proof=id if kills else None;verdict='witness' if kills else 'untrue'
 if n==names[6]:ids=['S1'];verdict='untrue'
 probe=['P1'] if n in mutants.get('P1',{}) and mutants['P1'][n]['failure'] else []
 vacuous=(not bool(probe)) if n in mutants.get('P1',{}) else None
 id=proof or ('S1' if n==names[6] else None)
 evidence=mutants[id][n]['command']+'; '+(mutants[id][n]['failure'] or '--- PASS: '+n) if id else ''
 obj={'test':n,'package':'stage1/typescript/parser','file':source[n],'seconds':timings[n]['median_seconds'],'oracle':notes[n],'oracle_kind':['external-run','self'] if n in names[:2] else 'self' if n==names[6] else 'external-run','kills':kills,'unique_kills':unique,'last_proven_fail':(proof+': '+mutants[proof][n]['failure']) if proof else None,'verdict':verdict,'subsumed_by':[subs[n]] if verdict=='subsumed' else [],'mutants_in_matrix':len(ids),'probe_kills':probe,'subsumer_seconds':timings[subs[n]]['median_seconds'] if verdict=='subsumed' else None,'vacuous':vacuous,'bounded':bounded,'matrix_rows':matrix_rows,'evidence':evidence,'matrix_ids':ids}
 result.append(obj)
(p/'audit.json').write_text(json.dumps(result,indent=2))
meta=json.loads((p/'mutants.json').read_text())
for m in meta:
 if m['id']=='M4':m['line']=58
 m['failed_rows']=[n for n,x in mutants.get(m['id'],{}).items() if x['failure']]
(p/'mutant-table.json').write_text(json.dumps(meta,indent=2))
print(json.dumps(result,indent=2))

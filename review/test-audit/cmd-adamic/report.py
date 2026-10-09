import json,statistics,re
from pathlib import Path
p=Path('review/test-audit/cmd-adamic'); runs=json.load(open(p/'matrix-runs.json')); times=json.load(open(p/'timings.json')); mutants=json.load(open(p/'mutants.json')); rows=list(runs[0]['status']); matrix=[]
for m in mutants:
 rs=[r for r in runs if r['mutant']==m['id']]; status={}; evidence={}
 for r in rs:
  status.update({k:v for k,v in r['status'].items() if v!='unknown'})
  events=[]
  for line in open(r['log']):
   try: events.append(json.loads(line))
   except ValueError: pass
  for row in r['failed_rows']:
   output=''.join(e.get('Output','') for e in events if e.get('Test','').split('/')[0]==row)
   lines=[x for k,v in r['failing_lines'].items() if k.split('/')[0]==row for x in v]
   evidence[row]={'command':r['command'],'failing_line':lines[0] if lines else next((x.strip() for x in output.splitlines() if 'panic:' in x),'test failed'),'output':output}
 matrix.append(dict(m,status=status,failed_rows=[x for x in rows if status.get(x)=='fail'],evidence=evidence))
json.dump(matrix,open(p/'matrix.json','w'),indent=2)
files=['checks_test.go','checks_test.go','non_null_checks_test.go']+['request_test.go']*5+['target_test.go']*2
unique={row:[m['id'] for m in matrix if not m['probe'] and m['failed_rows']==[row]] for row in rows}
audit=[]
for row,file in zip(rows,files):
 kills=[m['id'] for m in matrix if row in m['failed_rows']]; last=next((m for m in reversed(matrix) if row in m['failed_rows']),None); proof=next((m for m in matrix if m['id'] in unique[row]),None) or next((m for m in matrix if row in m['failed_rows']),None)
 oracle='hand-written expected '+('explanation text' if 'Explain' in row else 'selection, diagnostics, exit status or IR values')
 kind='self'
 if row=='TestWASIRequest': oracle='Node runs source independently and compares WASM response; hand-written ABI, lifetime and region-count checks'; kind=['external-run','self']
 if row=='TestWASIRequestThrows': oracle='hand-written exit 70 and panic text; Node hosts WASM only'
 if row=='TestExplainChecksDriver': oracle='no assertion in ordinary invocation; subprocess helper'
 audit.append({'test':row,'package':'cmd/adamic','file':'cmd/adamic/'+file,'seconds':statistics.median(t['binary_seconds'] for t in times if t['test']==row),'oracle':oracle,'oracle_kind':kind,'kills':kills,'unique_kills':unique[row],'last_proven_fail':last['id']+': '+last['evidence'][row]['failing_line'] if last else 'none; E_RUN passed','verdict':'sacred' if unique[row] else 'subsumed' if row=='TestBuildTargetParsing' else 'untrue','subsumed_by':['TestWASIRejectsTSGoArchive'] if row=='TestBuildTargetParsing' else [],'mutants_in_matrix':len(matrix),'vacuous':row=='TestExplainChecksDriver','bounded':False,'matrix_rows':[],'evidence':proof['evidence'][row]['command']+'; '+proof['evidence'][row]['failing_line'] if proof else 'ADAMIC_MUTANT=E_RUN ADAMIC_NATIVE_SPLIT=u007-E_RUN timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic/ -run .; TestExplainChecksDriver PASS'})
json.dump(audit,open(p/'audit.json','w'),indent=2)
d=json.load(open(p/'environment.json')); d['timing_runs_wall_seconds']=sum(t['wall_seconds'] for t in times);d['matrix_wall_seconds']=sum(r['wall_seconds'] for r in runs);d['matrix_runs']=len(runs);json.dump(d,open(p/'environment.json','w'),indent=2)
print('totals',d); print('matrix',[(m['id'],m['failed_rows']) for m in matrix]);print('audit',json.dumps(audit))

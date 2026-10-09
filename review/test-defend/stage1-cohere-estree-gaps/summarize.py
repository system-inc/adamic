import json
from pathlib import Path
p=Path(__file__).resolve().parent
pattern='^(TestRawInputGap|TestInterfaceDefaultGap|TestInterfaceTypeMethodGap|TestPostfixValueGap|TestMethodReplacementGap|TestJSXAgreement(_[0-9]{3})?)$'
meta={
'D1':('internal/native/runtime/directory.c:170','stat size + 1'),
'D2':('internal/lower/class_inheritance.go:761','missing optional callback argument: break -> return to'),
'D3':('internal/lower/iteration_origin.go:95','drop prototype-origin diagnostic assignment'),
'D4':('internal/lower/iteration_origin.go:88','class parent == -> !='),
'D5':('internal/lower/expression.go:1024','swap conditional branch operands'),
'D6':('internal/lower/expression.go:508','Boolean literal TrueKeyword == -> FalseKeyword =='),
'D7':('internal/lower/expression.go:783','PlusToken ir.Add -> ir.Subtract')}
def events(file):
 out=[]
 for s in file.read_text().splitlines():
  try: out.append(json.loads(s))
  except ValueError: pass
 return out
matrix={}
for ident,(line,change) in meta.items():
 ev=events(p/(ident+'.log'));states={e['Test']:e['Action'] for e in ev if 'Test'in e and '/'not in e['Test'] and e['Action'] in ('pass','fail','skip')}
 ran=sorted({e['Test'] for e in ev if 'Test'in e and '/'not in e['Test']})
 narrow=[]
 for f in sorted(p.glob(ident+'-Test*.log')):
  narrow.append(f.name)
  for e in events(f):
   if 'Test'in e and '/'not in e['Test'] and e['Action'] in ('pass','fail','skip'):states[e['Test']]=e['Action']
 outputs=[e['Output'].strip() for e in ev if 'Output'in e and ('gap changed:'in e['Output'] or 'reader changed:'in e['Output'] or 'default-free control:'in e['Output'] or 'exit status 70'in e['Output'])]
 matrix[ident]={'file_line':line,'change':change,'command':f"ADAMIC_BUILD_CACHE_DIR=/tmp/defend-estree/cache/{ident} timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/estree/ -run '{pattern}' > {ident}.log 2>&1",'rows_failed':sorted(n for n,s in states.items() if s=='fail'),'rows_passed':sorted(n for n,s in states.items() if s=='pass'),'rows_unknown':sorted(set(ran)-set(states)),'bounded':True,'binary_seconds':next((e['Elapsed'] for e in reversed(ev) if 'Test'not in e and e.get('Action')in ('pass','fail')),None),'cooked':any('test timed out' in e.get('Output','') for e in ev),'narrow_logs':narrow,'failure_lines':outputs}
rows=[]
for test,prior,ids in [('TestRawInputGap','TestJSXAgreement family',['D1']),('TestInterfaceDefaultGap','TestInterfaceTypeMethodGap',['D2','D3','D4']),('TestInterfaceTypeMethodGap','TestInterfaceDefaultGap',['D5','D6','D7'])]:
 defended=test=='TestRawInputGap'
 r={'test':test,'package':'stage1/cohere/estree','prior_verdict':'subsumed','subsumed_by':[prior],'defense':'defended'if defended else'not defended','unique_mutant':'D1 internal/native/runtime/directory.c:170'if defended else None,'attempts':[{'mutant':i,'file_line':meta[i][0],'change':meta[i][1],'rows_failed':sorted(set('TestJSXAgreement family' if n.startswith('TestJSXAgreement') else n for n in matrix[i]['rows_failed']))} for i in ids],'evidence':' ; '.join(matrix[i]['command']+' ; '+next((s for s in matrix[i]['failure_lines'] if ('reader changed:'in s if defended else ('gaps_test.go:135:'in s if test=='TestInterfaceDefaultGap' else 'default-free control:'in s))),matrix[i]['failure_lines'][0]) for i in ids),'bounded':True,'matrix_rows':matrix['D1']['rows_passed']+[test]if defended else sorted(set(matrix[ids[0]]['rows_passed']+matrix[ids[0]]['rows_failed']))}
 rows.append(r)
(p/'rows.json').write_text(json.dumps(rows,indent=2)+'\n');(p/'matrix.json').write_text(json.dumps(matrix,indent=2)+'\n')
print(json.dumps({i:{'failed':m['rows_failed'],'unknown':m['rows_unknown'],'seconds':m['binary_seconds'],'cooked':m['cooked']}for i,m in matrix.items()},indent=2))

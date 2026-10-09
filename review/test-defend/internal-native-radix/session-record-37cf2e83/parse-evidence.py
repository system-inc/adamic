import pathlib,json,re
p=pathlib.Path('/workspace/adamic/review/test-defend/internal-native-radix/session-record-37cf2e83')
def events(name):
 result=[]
 for line in (p/name).read_text().splitlines():
  try:result.append(json.loads(line))
  except:pass
 return result
plans=json.loads((p/'plan.json').read_text()); matrix=[]; attempts=[]
for plan in plans:
 es=events(plan['mutant']+'.log')
 failed=sorted({e['Test'].split('/')[0] for e in es if e.get('Action')=='fail' and e.get('Test')})
 passed=sorted({e['Test'] for e in es if e.get('Action')=='pass' and e.get('Test') and '/' not in e['Test']})
 errors=[e.get('Output','').strip() for e in es if e.get('Test','').split('/')[0]=='TestRecordBenchmark' and re.search(r'record_test.go:\d+:',e.get('Output',''))]
 matrix.append(dict(mutant=plan['mutant'],rows_failed=failed,rows_passed=passed,benchmark_output=errors,package_result=[e for e in es if e.get('Action') in ['pass','fail'] and not e.get('Test')]))
 attempts.append(dict(mutant=plan['mutant'],file_line=f"{plan['file']}:{plan['line']}",change=plan['change'],rows_failed=failed))
(p/'matrix.json').write_text(json.dumps(matrix,indent=2))
row=dict(test='TestRecordBenchmark',package='internal/native',prior_verdict='subsumed',subsumed_by='TestRuntimeStringEquality',defense='not defended',unique_mutant=None,attempts=attempts,evidence='ADAMIC_RECORD_BENCH=1 ADAMIC_BUILD_CACHE_DIR=/tmp/record-defense/cache/<D01-D03> timeout 330 go test -json -count=1 -timeout 300s ./internal/native/ -run "^(TestRegExp.*|TestRecord.*|TestRuntimeStringEquality|TestMapHash.*|TestLibraryMapSetIteratorResources)$"; see matrix.json and D01-D03.log for exact observed failures',bounded=True)
(p/'rows.json').write_text(json.dumps([row],indent=2))
print(json.dumps(matrix,indent=2)[:10000])

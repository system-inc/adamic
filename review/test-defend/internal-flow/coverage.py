import subprocess,pathlib,json,time
root=pathlib.Path('/workspace/adamic');p=root/'review/test-defend/internal-flow';tmp=pathlib.Path('/tmp/defend-flow')
for f in ['REPORT.md','README.md','rows.json','rows-results.json','mutants.json','bounded-members.json']:
 (p/('audit-'+f)).write_bytes(subprocess.check_output(['git','show','origin/test-audit/internal-flow:review/test-audit/internal-flow/'+f],cwd=root))
rows=['TestFlowSingleAssignment____oracle_testdata_timsort_a_63f1305dfb95','TestFlowMutationRanges____oracle_testdata_timsort_a_63f1305dfb95','TestFlowGraphPaths____oracle_testdata_timsort_a_63f1305dfb95']
runs=[]
for label,pattern in [(n,'^'+n+'$') for n in rows]+[('family','^(TestFlowProgram.*|TestFlowCorpusUnitsCoverEveryProgram|TestFlowCorpusRemainder)$')]:
 cmd=f"source /workspace/adamic-tools/env.sh; ADAMIC_BUILD_CACHE_DIR=/tmp/defend-flow/cache/coverage-{label} timeout 120 go test -count=1 -timeout 90s ./internal/flow/ -run '{pattern}' -coverpkg=./internal/flow -coverprofile={p}/{label}.cover"
 start=time.monotonic()
 with (p/(label+'-coverage.log')).open('w') as log:r=subprocess.run(['bash','-c',cmd],cwd=root,stdout=log,stderr=subprocess.STDOUT)
 runs.append(dict(label=label,command=cmd,exit=r.returncode,wall_seconds=time.monotonic()-start));(p/'coverage-runs.json').write_text(json.dumps(runs,indent=2));print(label,r.returncode,flush=True)

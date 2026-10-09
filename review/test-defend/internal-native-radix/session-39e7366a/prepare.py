import pathlib,re,json,subprocess,time,gzip
p=pathlib.Path('review/test-defend/internal-native-radix/session-39e7366a');(p/'coverage').mkdir(exist_ok=True)
root='origin/test-audit/internal-native-radix:review/test-audit/internal-native-radix/'
audit=json.loads(subprocess.check_output(['git','show',root+'rows.json']));selected=[r for r in audit if r['test'] in ['TestRegExpBytecodeTest262','TestRegExpBytecodeRandomNode family']];(p/'audit-rows.json').write_text(json.dumps(selected,indent=2))
alltests=[s for s in (p/'logs/list.log').read_text().splitlines() if s.startswith('Test')]
names=[]
for file in ['internal/native/regexp_test.go','internal/native/regexp_search_test.go']:
 names+=re.findall(r'^func (Test\w+)\(',pathlib.Path(file).read_text(),re.M)
names+=['TestRegexProgramsKeepCheckedFieldReads','TestClosureConventionRuntimeDropCount']
assert all(n in alltests for n in names)
old=json.loads(subprocess.check_output(['git','show',root+'scope.json']))
(p/'scope.json').write_text(json.dumps({'base':subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip(),'all_package_tests':alltests,'matrix_members':names,'matrix_rows':['TestRegExpBytecodeRandomNode family']+[n for n in names if not n.startswith('TestRegExpBytecodeRandomNodeUnit')],'family_members':[n for n in names if n.startswith('TestRegExpBytecodeRandomNodeUnit')],'reach_basis':'All top-level regexp runtime tests in regexp_test.go and regexp_search_test.go, plus regex field-read runtime integration and regexp callback compilation witness. Go regexp uses in other native tests are host harness utilities, not native VM execution. Kills outside this statically reached set are unknown.'},indent=2))
cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run','^('+'|'.join(names)+')$'];t=time.monotonic()
with (p/'logs/slice-baseline.log').open('w') as log:code=subprocess.call(cmd,stdout=log,stderr=log)
(p/'slice-baseline-run.json').write_text(json.dumps({'command':cmd,'exit':code,'wall_seconds':time.monotonic()-t},indent=2))
assert code==0,'red slice baseline'
results=[]
for label,regex in [('TestRegExpBytecodeTest262','^TestRegExpBytecodeTest262$'),('TestRegExpBytecodeRandomNode-family','^TestRegExpBytecodeRandomNodeUnit[0-9]+$'),('TestRegExpSearchNode','^TestRegExpSearchNode$')]:
 cmd=['timeout','120','go','test','-count=1','-timeout','90s','./internal/native/','-run',regex,'-coverpkg=./internal/native,./internal/regexp','-coverprofile='+str(p/'coverage'/(label+'.out'))];t=time.monotonic()
 with (p/'logs'/('coverage-'+label+'.log')).open('w') as log:code=subprocess.call(cmd,stdout=log,stderr=log)
 results.append({'label':label,'command':cmd,'exit':code,'wall_seconds':time.monotonic()-t});(p/'coverage-runs.json').write_text(json.dumps(results,indent=2));print(label,code,round(results[-1]['wall_seconds'],2),flush=True);assert code==0

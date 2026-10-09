import pathlib,json,subprocess,os,time
p=pathlib.Path('/tmp/def-oct6'); old=json.loads((p/'audit-results.json').read_text()); names=[x['test'] for x in old if x['test']!='TestNativeAgreesWithNode'];names+=['TestInputAgreesWithNode','TestCensusAppendResultProof','TestCensusOverloadResultStop','TestNodeFSFileAgreesWithNode','TestNodeFSFileMutants','TestScannerNestedOverloadImplementationMutantIsCaught','TestSwitchCaseOverloadIsNotYet']
current=[x for x in (p/'list.log').read_text().splitlines() if x.startswith('Test')]; names=[x for x in names if x in current]
fixtures=json.loads((p/'caller-fixtures.json').read_text()); paths=sorted(set(fixtures['readTextFile']+fixtures['overload_signatures']+['internal/oracle/testdata/interleaved.a','internal/oracle/testdata/write_stdout_order.a','internal/oracle/testdata/write_stderr_order.a','internal/oracle/testdata/omitted_scanner.a','internal/oracle/testdata/omitted_reader.a']))
import re
patterns=['^('+'|'.join(names)+')$','^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^('+'|'.join(re.escape(x.rsplit('/',1)[1]) for x in paths)+')$'];(p/'scope.json').write_text(json.dumps({'top_level_rows':names+['TestNativeAgreesWithNode'],'native_fixture_candidates':paths,'patterns':patterns,'bounded':True},indent=2))
results=[]
for i,pattern in enumerate(patterns):
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',pattern];start=time.monotonic()
 with (p/f'bounded-baseline-{i}.log').open('w') as f:q=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1','ADAMIC_BUILD_CACHE_DIR':'/tmp/def-oct6/cache/clean'})
 results.append({'command':' '.join(cmd),'exit':q.returncode,'wall_seconds':time.monotonic()-start});(p/'baseline-runs.json').write_text(json.dumps(results,indent=2));print(i,q.returncode,flush=True)

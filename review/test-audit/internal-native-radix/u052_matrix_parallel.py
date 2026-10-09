import pathlib,json,subprocess,os,time
out=pathlib.Path('/workspace/adamic/review/test-audit/internal-native-radix'); rows=json.loads((out/'scope.json').read_text()); names=[n for _,ms in rows for n in ms]
pattern='^('+'|'.join(names)+')$'
def run(ident):
 env=os.environ.copy();env.update(ADAMIC_MUTANT=ident,ADAMIC_RECORD_BENCH='1',ADAMIC_BUILD_CACHE_DIR='/tmp/u052/cache/'+ident)
 # Probes run only direct users, excluding witness rows whose preconditions are not verdict evidence.
 probe_patterns={'P01':'^TestToStringWithARadix','P02':'^TestRegExp','P03':'^TestRegExp','P04':'^TestRegExpIteratorResultShape$','P05':'^TestRuntimeStringEquality$','P06':'^(TestRuntimeReleasePaths|TestRuntimeStringEquality)$','P07':'^(TestRecordsAgainstNode|TestRecordBenchmark)$','P08':'^(TestRecordsAgainstNode|TestRecordBenchmark)$'}
 chosen=probe_patterns.get(ident,pattern)
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run',chosen]
 start=time.monotonic()
 with (out/(ident+'.log')).open('w') as log:r=subprocess.run(cmd,cwd='/workspace/adamic',env=env,stdout=log,stderr=subprocess.STDOUT)
 (out/(ident+'-run.json')).write_text(json.dumps({'command':'ADAMIC_MUTANT='+ident+' ADAMIC_RECORD_BENCH=1 ADAMIC_BUILD_CACHE_DIR='+env['ADAMIC_BUILD_CACHE_DIR']+' '+' '.join(cmd),'exit':r.returncode,'wall':time.monotonic()-start,'pattern':chosen},indent=2))

from concurrent.futures import ThreadPoolExecutor
with ThreadPoolExecutor(max_workers=2) as pool: list(pool.map(run,[f"M{i:02}" for i in range(1,13)]+[f"P{i:02}" for i in range(1,9)]))

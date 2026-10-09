"""Measure the current typeaware child API on the same emitted program and archive."""
from pathlib import Path
import json,os,subprocess,sys,time
root=Path.cwd(); evidence=root/'review/compiler/lowering-chain'; phase=sys.argv[1]
source=(evidence/'member-11-program.c').read_text(); scratch=Path('/workspace/scratch/lowering-chain-tsgo')
truth=subprocess.run(['node','--disable-warning=ExperimentalWarning','oracle/node.mjs','internal/oracle/testdata/arguments_length_extended.a'],capture_output=True,check=True)
results=[]
for name,sanitize in [('release-cold',False),('release-repeat',False),('sanitized-cold',True),('sanitized-repeat',True)]:
 output=scratch/(phase+'-'+name);request=evidence/('member-11-'+phase+'-'+name+'-request.json')
 request.write_text(json.dumps(dict(Source=source,Output=str(output),Archive=str(scratch/'checker.a'),Sanitize=sanitize)))
 log=evidence/('member-11-'+phase+'-'+name+'-tests.jsonl');command=['go','test','-p','1','./stage1/cohere/typeaware','-run','^TestTypeAwareNativeBuildChild$','-count=1','-json']
 started=time.monotonic()
 with log.open('w') as f:r=subprocess.run(command,stdout=f,stderr=subprocess.STDOUT,env={**os.environ,'ADAMIC_TYPEAWARE_NATIVE_REQUEST':str(request)})
 elapsed=time.monotonic()-started;assert r.returncode==0,log
 actual=subprocess.run([str(output)],capture_output=True,env={**os.environ,'ASAN_OPTIONS':'detect_leaks=1'})
 assert (actual.returncode,actual.stdout,actual.stderr)==(truth.returncode,truth.stdout,truth.stderr),(name,actual)
 tests=[json.loads(line) for line in log.read_text().splitlines() if line.startswith('{')];leaf=next(x for x in tests if x['Action']=='pass' and x.get('Test')=='TestTypeAwareNativeBuildChild')
 assert leaf['Elapsed']<60,leaf
 results.append(dict(name=name,wall_seconds=elapsed,test_seconds=leaf['Elapsed'],node_stdout=truth.stdout.decode(),exit=actual.returncode)); print(phase,name,round(elapsed,3),leaf['Elapsed'],flush=True)
(evidence/('member-11-'+phase+'-timings.json')).write_text(json.dumps(results,indent=2)+'\n')

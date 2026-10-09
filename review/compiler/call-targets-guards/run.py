"""Run each guard separately, then the audit overlays, and restored guards."""
import difflib, gzip, json, os, subprocess, time
from pathlib import Path
root=Path(__file__).resolve().parents[3]
out=Path(__file__).resolve().parent
scratch=Path('/workspace/call-targets-overlays'); scratch.mkdir(exist_ok=True)
env=dict(os.environ,GOMAXPROCS='4',ADAMIC_GATE_UNCACHED='1')
rows=[]
def run(label,pkg,test,overlay=None,expected='pass',reason=None):
 command=['go','test',pkg,'-run','^'+test+'$','-count=1','-json','-timeout=90s']
 if overlay: command+=['-overlay',str(overlay)]
 log=out/(label+'.jsonl'); start=time.monotonic()
 with log.open('w') as stream: result=subprocess.run(command,cwd=root,env=env,stdout=stream,stderr=subprocess.STDOUT)
 elapsed=time.monotonic()-start
 events=[]
 for line in log.read_text().splitlines():
  try: events.append(json.loads(line))
  except ValueError: pass
 terminal=[e for e in events if e.get('Test')==test and e.get('Action') in ['pass','fail','skip']]
 row=dict(label=label,command=command,invocation_seconds=round(elapsed,3),test_seconds=terminal[-1].get('Elapsed') if terminal else None,action=terminal[-1]['Action'] if terminal else 'build-failure',exit=result.returncode)
 rows.append(row); (out/'results.json').write_text(json.dumps(rows,indent=2)+'\n'); print(row,flush=True)
 assert row['action']==expected,(label,log.read_text()[-5000:])
 assert (result.returncode==0)==(expected=='pass')
 if reason: assert reason in log.read_text(),label
 if expected=='pass': assert elapsed<60,(label,elapsed)
units=[('./internal/ir','TestCallMayThrowUsesReachableTargets'),('./internal/ir','TestDirectClosureTargetsUseEncodedIndex'),('./internal/oracle','TestCallTargetThrowAgreesWithNode'),('./internal/oracle','TestDirectClosureCallAgreesWithNode')]
for pkg,test in units: run('baseline-'+test,pkg,test)
source=root/'internal/ir/call_targets.go'; text=source.read_text()
mutations=[('M7','func (p *Program) CallMayThrow(call Call) bool {','func (p *Program) CallMayThrow(call Call) bool { return false;',units[0],'direct throwing: CallMayThrow = false, want true'),('M6','return FunctionTargets{Functions: []int{call.Direct - 1}}','return FunctionTargets{Functions: []int{call.Direct}}',units[1],'ClosureTargets = {Functions:[1] Unknown:false}, want known target [0]'),('always-throw','func (p *Program) CallMayThrow(call Call) bool {','func (p *Program) CallMayThrow(call Call) bool { return true;',units[0],'direct nonthrowing: CallMayThrow = true, want false')]
for label,old,new,(pkg,test),reason in mutations:
 assert text.count(old)==1
 replacement=scratch/(label+'.go'); replacement.write_text(text.replace(old,new))
 overlay=scratch/(label+'.json'); overlay.write_text(json.dumps({'Replace':{str(source):str(replacement)}}))
 (out/(label+'.diff.gz')).write_bytes(gzip.compress(''.join(difflib.unified_diff(text.splitlines(True),replacement.read_text().splitlines(True),fromfile='internal/ir/call_targets.go',tofile=label+'/call_targets.go')).encode(),mtime=0))
 run(label,pkg,test,overlay,'fail',reason)
for pkg,test in units: run('restored-'+test,pkg,test)

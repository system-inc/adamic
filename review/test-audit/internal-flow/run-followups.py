import subprocess,time,json,os,re,difflib
from pathlib import Path
r=Path('/workspace/adamic/review/test-audit/internal-flow');pkg='./internal/flow/'
while len((r/'run-times.jsonl').read_text().splitlines())<13 or not (r/'BOUNDED_DONE').exists():time.sleep(3)
listed=[s for s in (r/'u018-list.log').read_text().splitlines() if s.startswith('Test')]
family=[s for s in listed if s.startswith('TestFlowProgram_')]+['TestFlowCorpusUnitsCoverEveryProgram','TestFlowCorpusRemainder']
other=[s for s in listed if s not in family]
rows={'TestFlowProgram family':family,**{s:[s] for s in other}}
(r/'rows.json').write_text(json.dumps(rows,indent=2))
bounded=[s for s in listed if 'testdata_joins_a_' in s or 'testdata_mutations_a_' in s or 'timsort_a_' in s or s in ['TestDebuggerHasNoFlowInstruction','TestFlowCorpusSetupIsShared','TestFlowCorpusUnitsCoverEveryProgram','TestFlowCorpusRemainder']]
(r/'bounded-members.json').write_text(json.dumps(bounded,indent=2))
def run(id,pattern,label):
 env=os.environ.copy();env['ADAMIC_MUTANT']=id;env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u018/cache/'+id;began=time.monotonic()
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s',pkg,'-run',pattern]
 with (r/(label+'.log')).open('w') as f:code=subprocess.call(cmd,stdout=f,stderr=subprocess.STDOUT,env=env)
 with (r/'follow-times.jsonl').open('a') as f:f.write(json.dumps({'id':id,'label':label,'command':cmd,'wall_seconds':time.monotonic()-began,'exit':code})+'\n')
 print(label,code,round(time.monotonic()-began,3),flush=True)
 return code
for mid in ['M'+str(i) for i in range(1,10)]+['PBuild','PConstruct','PLiveOut','PRanges']:
 text=(r/(mid+'.log')).read_text()
 if 'test timed out' in text and not (r/(mid+'-bounded.log')).exists():run(mid,'^('+'|'.join(bounded)+')$',mid+'-bounded')
 elif 'panic:' in text and 'test timed out' not in text:
  for n,(row,members) in enumerate(rows.items()):run(mid,'^'+members[0]+'$',mid+'-row'+str(n))
for n,(row,members) in enumerate(rows.items()):
 for repetition in range(1,4):run('','^('+'|'.join(members)+')$','timing-row'+str(n)+'-'+str(repetition))
# Restore production sources before compiling every standalone diff.
subprocess.run(['git','restore','--','internal/flow/build.go','internal/flow/ssa.go','internal/flow/liveness.go','internal/flow/ranges.go'],check=True)
for diff in sorted(r.glob('*.diff')):
 if diff.name=='switch.diff':continue
 subprocess.run(['git','apply',str(diff)],check=True)
 with (r/(diff.stem+'-vet.log')).open('w') as f:code=subprocess.call(['go','vet',pkg],stdout=f,stderr=subprocess.STDOUT)
 with (r/'diff-vet.jsonl').open('a') as f:f.write(json.dumps({'id':diff.stem,'exit':code})+'\n')
 subprocess.run(['git','apply','-R',str(diff)],check=True)
 print('vet',diff.stem,code,flush=True)
# Suite construction checks, one allowed harness change at a time.
for mid,p,old,new,test in [('SSetupLower','internal/flow/flow_test.go','return entry.program','return &ir.Program{}','TestFlowCorpusSetupIsShared'),('SSetupTrace','internal/flow/trace_test.go','return entry.setup','return &traceSetup{}','TestFlowCorpusSetupIsShared'),('SSetupCorpus','internal/flow/corpus_units_test.go','"../../dedication/dedication.a",','"../load/testdata/0.1/compile/01_hello.ts",','TestFlowCorpusUnitsCoverEveryProgram')]:
 source=Path(p).read_text();assert old in source;changed=source.replace(old,new,1)
 (r/(mid+'.diff')).write_text(''.join(difflib.unified_diff(source.splitlines(True),changed.splitlines(True),fromfile='a/'+p,tofile='b/'+p)))
 Path(p).write_text(changed)
 run('','^'+test+'$',mid)
 with (r/(mid+'-vet.log')).open('w') as f:code=subprocess.call(['go','vet',pkg],stdout=f,stderr=subprocess.STDOUT)
 with (r/'diff-vet.jsonl').open('a') as f:f.write(json.dumps({'id':mid,'exit':code})+'\n')
 Path(p).write_text(source)
# Measured reachability inventory from clean selected package.
with (r/'coverage.log').open('w') as f:subprocess.call(['timeout','120','go','test','-count=1','-timeout','90s','-coverprofile='+str(r/'coverage.out'),pkg],stdout=f,stderr=subprocess.STDOUT)
(r/'DONE').write_text('complete\n')

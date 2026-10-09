import subprocess,time,json,os,re,difflib
from pathlib import Path
r=Path('/workspace/adamic/review/test-audit/internal-flow');pkg='./internal/flow/'
while len((r/'run-times.jsonl').read_text().splitlines())<9:time.sleep(3)
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
for mid in ['M'+str(i) for i in range(1,10)]:
 text=(r/(mid+'.log')).read_text()
 if 'test timed out' in text:run(mid,'^('+'|'.join(bounded)+')$',mid+'-bounded')

(r/'BOUNDED_DONE').write_text('complete\n')

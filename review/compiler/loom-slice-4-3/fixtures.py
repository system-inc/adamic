import subprocess,os,json,time
from pathlib import Path
out=Path('review/compiler/loom-slice-4-3'); rows=[]
paths=['stage3/fixtures/(generics|iteration|iteration-dispatch)','stage3/project-references-source','internal/oracle/testdata/generic_body','internal/oracle/testdata/generators','internal/oracle/testdata/notyet_element_access/(nodearray|sorted|template)','internal/oracle/testdata/class_wrong_output_refused/iterators_(override_this|hidden_return)','internal/oracle/testdata/user_iterators_rest_tdz.a']
for i,path in enumerate(paths):
 cmd=['go','test','./internal/oracle','-run','^TestNativeAgreesWithNode_[0-9]+$/'+path,'-parallel','4','-count=1','-timeout','90s','-json']; log=out/f'fixtures-{i}.jsonl';start=time.monotonic()
 with log.open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,timeout=150)
 events=[json.loads(l) for l in log.read_text().splitlines() if l.startswith('{')]; leaves=[e for e in events if e.get('Action')=='pass' and '/' in e.get('Test','')]
 print(i,r.returncode,len(leaves),time.monotonic()-start,flush=True);rows.append({'command':cmd,'exit':r.returncode,'leaves':len(leaves)})
 assert r.returncode==0 and leaves,(i,'no passing fixtures')
 assert not any(e.get('Elapsed',0)>=60 for e in events if e.get('Test')), 'red leaf exceeds 60 seconds; never rerun'
(out/'fixture-results.json').write_text(json.dumps(rows,indent=2)+'\n')

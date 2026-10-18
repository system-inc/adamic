import json,re,subprocess,time
from pathlib import Path
out=Path('review/compiler/loom-r-slice-4');source=Path('internal/flow/corpus_units_test.go').read_text();names=re.findall(r'^func (Test\w+)\(',source,re.M);names+=['TestFlowCorpusUnitsCoverEveryProgram','TestFlowCorpusSetupIsShared','TestFlowCorpusRemainder'];rows=[]
for i in range(0,len(names),25):
 group=names[i:i+25];log=out/('flow-corpus-%02d.jsonl'%(i//25));cmd=['go','test','./internal/flow','-run','^('+'|'.join(group)+')$','-parallel','4','-count=1','-timeout','90s','-json'];start=time.monotonic()
 with log.open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,timeout=150)
 events=[]
 for line in log.read_text().splitlines():
  try:events.append(json.loads(line))
  except:pass
 terminal=[e for e in events if e.get('Test') and e.get('Action') in ['pass','fail','skip']]
 rows.append({'command':cmd,'exit':r.returncode,'seconds':time.monotonic()-start,'log':str(log)});(out/'flow-corpus-results.json').write_text(json.dumps(rows,indent=2)+'\n');print(i//25,r.returncode,rows[-1]['seconds'],flush=True)
 assert not any(e.get('Elapsed',0)>=60 for e in terminal),'red leaf over 60; split before another run'
 if r.returncode:raise SystemExit(1)
 assert set(group).issubset({e['Test'] for e in terminal})

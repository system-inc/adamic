from pathlib import Path
import json,subprocess,time
out=Path('review/compiler/loom-r-fx3-refusals');rows=[{'package':'./bench/regex','tests':[]},{'package':'./cmd/adamic-fuzz','tests':[]},{'package':'./cmd/adamic-reduce','tests':[]},{'package':'./internal/javascript','tests':['TestViewMixedUnionSelectsNumber']}];results=[]
(out/'extra-reader-selection.json').write_text(json.dumps(rows,indent=2)+'\n')
for row in rows:
    command=['go','test',row['package'],'-run','^('+'|'.join(row['tests'])+')$' if row['tests'] else '^$','-count=1','-v','-timeout','90s','-parallel','4'];began=time.monotonic()
    with (out/('extra-'+row['package'][2:].replace('/','-')+'.log')).open('w') as log:r=subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,timeout=180)
    results.append(dict(**row,command=command,exit=r.returncode,seconds=round(time.monotonic()-began,3)));(out/'extra-reader-results.json').write_text(json.dumps(results,indent=2)+'\n');print(json.dumps(results[-1]),flush=True)
    assert r.returncode==0

from pathlib import Path
import json,re,subprocess,time,sys
out=Path('review/compiler/loom-r-fx3-refusals')
completed={'./internal/lower','./internal/oracle','./internal/ir','./cmd/adamic'}
for f in ['consumer-results.json','consumer-final-results.json','consumer-rest-results.json']:
    if (out/f).exists():completed.update(r['package'] for r in json.loads((out/f).read_text()) if r['exit']==0)
results=json.loads((out/'consumer-rest-results.json').read_text()) if (out/'consumer-rest-results.json').exists() else []
for row in json.loads((out/'consumer-selection.json').read_text()):
    if row['package'] in completed:continue
    pattern='^('+'|'.join(row['tests'])+')$' if row['tests'] else '^$'
    command=['go','test',row['package'],'-run',pattern,'-count=1','-v','-timeout','90s','-parallel','4']
    log=out/('consumer-rest-'+str(len(results))+'-'+row['package'][2:].replace('/','-')+'.log');began=time.monotonic()
    with log.open('w') as output:
        try:code=subprocess.run(command,stdout=output,stderr=subprocess.STDOUT,timeout=180).returncode
        except subprocess.TimeoutExpired:code=124
    result=dict(**row,command=command,exit=code,log=str(log),seconds=round(time.monotonic()-began,3));results.append(result)
    (out/'consumer-rest-results.json').write_text(json.dumps(results,indent=2)+'\n');print(json.dumps(result),flush=True)
    if code:sys.exit(code)
    times=re.findall(r'--- PASS: (\S+) \(([\d.]+)s\)',log.read_text())
    assert all(float(seconds)<60 for _,seconds in times),'over-budget leaf; never rerun'

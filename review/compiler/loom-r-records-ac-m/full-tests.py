from pathlib import Path
import subprocess,json,time,sys
out=Path('review/compiler/loom-r-records-ac-m');results=[]
commands=[
 ('compiler-product',['go','test','./internal/native','-run','^TestProduct_WASICompiler$','-count=1','-json','-timeout','90s','-parallel','4']),
 ('runtime-products',['go','test','./internal/oracle','-run','^TestProduct_NativeRuntimes(Sanitized|Release|Counted|Shipped)[012]$','-count=1','-json','-timeout','300s','-parallel','4']),
]
for label,pkg,seconds in [('lower','./internal/lower/...',840),('native','./internal/native/...',840),('oracle','./internal/oracle/...',1200),('flow','./internal/flow/...',840),('fresh','./internal/fresh/...',840),('ir','./internal/ir/...',240),('javascript','./internal/javascript/...',240),('fixtures','./stage3/fixtures/...',840)]:
 commands.append((label,['go','test',pkg,'-count=1','-json','-timeout',str(seconds)+'s','-parallel','4','-p','1']))
for label,command in commands:
    limit=int(command[command.index('-timeout')+1][:-1])+120;began=time.monotonic();log=out/('full-'+label+'.jsonl')
    print('START '+label+' '+json.dumps(command),flush=True)
    with log.open('w') as output:
        process=subprocess.Popen(command,stdout=output,stderr=subprocess.STDOUT)
        while process.poll() is None:
            if time.monotonic()-began>limit:
                process.terminate()
                try:process.wait(timeout=10)
                except subprocess.TimeoutExpired:process.kill();process.wait()
                break
            time.sleep(2)
    rows=[]
    for line in log.read_text(errors='replace').splitlines():
        try:rows.append(json.loads(line))
        except json.JSONDecodeError:pass
    tests=[{'test':r['Test'],'package':r.get('Package'),'action':r['Action'],'seconds':r.get('Elapsed',0)} for r in rows if r.get('Test') and r['Action'] in ['pass','fail','skip']]
    failures=[r for r in tests if r['action']=='fail'];slow=[r for r in tests if r['action']!='skip' and r['seconds']>=60]
    row={'label':label,'command':command,'exit':process.returncode,'seconds':round(time.monotonic()-began,3),'failures':failures,'over_60s':slow,'passed':sum(r['action']=='pass' for r in tests),'skipped':sum(r['action']=='skip' for r in tests)}
    (out/('full-'+label+'-tests.json')).write_text(json.dumps(tests,indent=2)+'\n');results.append(row);(out/'full-results.json').write_text(json.dumps(results,indent=2)+'\n');print(json.dumps(row),flush=True)

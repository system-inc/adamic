import hashlib,json,statistics,subprocess
from pathlib import Path
base=Path('/workspace/scratch/stack-check-scc')
out=base/'cache'
helper=base/'handler/wait4'
report={'rounds':7,'warmup':'one untimed invocation per variant','instrument':'wait4 direct-child ru_utime; ordinary execution without Valgrind','load_before':Path('/proc/loadavg').read_text().strip(),'samples':[],'binaries':{}}
for side in ['before','natural','after']:
    path=base/('native-'+side)
    report['binaries'][side]={'sha256':hashlib.sha256(path.read_bytes()).hexdigest(),'command':[str(path),'--manifest',str(base/'compiler.txt'),'--count']}
def measure(side,label):
    prefix=out/label
    command=[str(helper),str(prefix)+'.rusage.json',str(prefix)+'.stdout',str(prefix)+'.stderr','300000',*report['binaries'][side]['command']]
    with (out/'timing-helper.log').open('a') as log:
        subprocess.run(command,cwd='/workspace/adamic',check=True,stdout=log,stderr=log)
    result=json.loads(Path(str(prefix)+'.rusage.json').read_text())
    if result['exitCode']!=0 or result['timedOut'] or Path(str(prefix)+'.stdout').read_bytes()!=b'0\n' or Path(str(prefix)+'.stderr').read_bytes():
        raise RuntimeError('benchmark observation differs')
    return {'side':side,'label':label,**result}
for side in ['before','natural','after']:
    measure(side,'warmup-'+side)
for round_number in range(7):
    for side in (['before','natural','after'][round_number%3:]+['before','natural','after'][:round_number%3]):
        report['samples'].append(measure(side,str(round_number)+'-'+side))
report['load_after']=Path('/proc/loadavg').read_text().strip()
report['summary']={}
for side in ['before','natural','after']:
    samples=[x for x in report['samples'] if x['side']==side]
    report['summary'][side]={key:{'best':min(x[key] for x in samples),'median':statistics.median(x[key] for x in samples)} for key in ['userMs','systemMs','wallMs']}
(out/'user-time.json').write_text(json.dumps(report,indent=2)+'\n')
print(json.dumps(report['summary'],indent=2))

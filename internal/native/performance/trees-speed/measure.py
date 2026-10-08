import os,sys,time,resource,subprocess,json,platform,hashlib
from pathlib import Path
out=Path(sys.argv[1]); names=sys.argv[2:]; cmds={}
for item in names:
    name,path=item.split('=',1)
    cmds[name]=[path]+(['--disable-warning=ExperimentalWarning','/workspace/adamic/bench/trees.ts'] if name=='Node' else ['/workspace/adamic/bench/trees.ts'] if name=='Bun' else [])
env=dict(os.environ)
for key in ['NODE_OPTIONS','BUN_OPTIONS','ADAMIC_THREADS']: env.pop(key,None)
meta={'machine':platform.node(),'platform':platform.platform(),'cpu':next(x.split(':',1)[1].strip() for x in Path('/proc/cpuinfo').read_text().splitlines() if x.startswith('model name')),'affinity':list(os.sched_getaffinity(0)),'cpu.max':Path('/sys/fs/cgroup/cpu.max').read_text().strip(),'load_before':os.getloadavg(),'commands':cmds,'sha256':{k:hashlib.sha256(Path(v[0]).read_bytes()).hexdigest() for k,v in cmds.items()},'source_sha256':hashlib.sha256(Path('/workspace/adamic/bench/trees.ts').read_bytes()).hexdigest()}
def run(cmd):
    prev=resource.getrusage(resource.RUSAGE_CHILDREN); start=time.perf_counter()
    p=subprocess.run(cmd,capture_output=True,check=True,env=env,timeout=120)
    wall=time.perf_counter()-start; after=resource.getrusage(resource.RUSAGE_CHILDREN)
    return p,{'wall':wall,'user':after.ru_utime-prev.ru_utime,'system':after.ru_stime-prev.ru_stime}
expected=None
for name,cmd in cmds.items():
    p,_=run(cmd)
    if expected is None: expected=p.stdout
    assert p.stdout==expected and expected,(name,p.stdout)
samples={k:[] for k in cmds}; keys=list(cmds)
for r in range(5):
    for k in keys[r%len(keys):]+keys[:r%len(keys)]:
        p,measure=run(cmds[k]); assert p.stdout==expected,(k,p.stdout)
        samples[k].append(measure)
meta['load_after']=os.getloadavg(); meta['stdout']=expected.decode(); meta['samples']=samples; meta['best']={k:min(v,key=lambda x:x['wall']) for k,v in samples.items()}
out.write_text(json.dumps(meta,indent=2)+'\n'); print(json.dumps(meta['best'],indent=2)); print('load',meta['load_before'],meta['load_after'])

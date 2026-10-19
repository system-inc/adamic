import json,os,subprocess,time
from pathlib import Path
out=Path(__file__).parent
os.sched_setaffinity(0,sorted(os.sched_getaffinity(0))[:4])
env=dict(os.environ,GOMAXPROCS='4',GOPROXY='https://proxy.golang.org|direct',ADAMIC_BUILD_OUTPUT_DIR='/tmp/adamic-gate/loom-r-slice-3-ac-m-work/products',GOCACHE='/tmp/adamic-gate/loom-r-slice-3-ac-m-work/go-cache',ADAMIC_GOCACHE_OFF='1');env.pop('GOCACHEPROG',None)
packages=['internal/ir','internal/javascript','internal/lower','internal/flow','internal/native','internal/oracle','stage1/cohere/cssstrings','stage1/cohere/json','stage1/cohere/markdownblocks','stage1/cohere/selector','stage1/cohere/yaml','stage1/typescript/parser','stage1/typescript/scanner','stage3/fixtures']
commands=[
 ('production-mutants',['timeout','720','python3',str(out/'run-mutants.py')]),
 ('reader-guard',['timeout','120','go','test','./internal/ir','-run','^TestCallTargetReaders$','-count=1','-parallel=4','-timeout=90s']),
 ('vet',['timeout','180','go','vet','-p','1',*['./'+p+'/...' for p in packages]]),
 ('metadata-guards',['timeout','120','go','test','./cmd/adamic-gate','-run','^TestEveryTest(HasOneDescriptiveName|IsParallelOrSaysWhy)$','-count=1','-timeout=90s']),
 ('linker-mutant',['timeout','120','go','test','-overlay',str(out/'linker-mutant/overlay.json.txt'),'./internal/native','-run','^TestTSGoCacheShippedRelease$','-count=1','-timeout=90s','-v']),
 ('markdown-own-focus',['timeout','120','go','test','./stage1/cohere/markdownblocks','-run','^TestParserRepresentationProbes$','-count=1','-timeout=90s','-v'])]
results=[]
for name,command in commands:
 start=time.monotonic()
 with (out/(name+'.log')).open('w') as log:r=subprocess.run(command,env=env,stdout=log,stderr=subprocess.STDOUT,timeout=int(command[1])+10)
 row=dict(name=name,command=command,exit=r.returncode,seconds=round(time.monotonic()-start,3));results.append(row);(out/'finish-results.json.txt').write_text(json.dumps(results,indent=2)+'\n');print(json.dumps(row),flush=True)

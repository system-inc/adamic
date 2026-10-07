import os,subprocess
from pathlib import Path
base=Path('/workspace/scratch/stack-check-scc');out=base/'tail-calls'
processes=[]
for side in ['before','after']:
    command=[str(base/'valgrind/usr/bin/valgrind'),'--tool=callgrind','--cache-sim=yes','--branch-sim=yes','--I1=32768,8,64','--D1=32768,8,64','--LL=268435456,1,64','--callgrind-out-file='+str(out/('parse-'+side+'.callgrind')),str(out/('native-'+side)),'--manifest',str(base/'compiler.txt'),'--count']
    env=os.environ.copy();env['VALGRIND_LIB']=str(base/'valgrind/usr/libexec/valgrind')
    processes.append((side,subprocess.Popen(command,env=env,stdout=(out/('parse-'+side+'.stdout')).open('w'),stderr=(out/('parse-'+side+'.stderr')).open('w'))))
for side,process in processes:
    status=process.wait();print(side,status,flush=True)
    if status: raise SystemExit(status)

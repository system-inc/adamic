import os, subprocess, time, json, sys
binary, label, *tests = sys.argv[1:]
os.environ['GOMAXPROCS']='4'
for test in tests:
    os.environ['ADAMIC_SPLIT_TEST262_MEASUREMENT']=str(time.time_ns())
    start=time.perf_counter()
    p=subprocess.run([binary,'-test.run',test,'-test.v','-test.count=1'],stdout=subprocess.PIPE,stderr=subprocess.STDOUT,text=True)
    elapsed=time.perf_counter()-start
    name=test.replace('^','').replace('$','').replace('/','_')
    open('/workspace/scratch/split-test262/'+label+'-'+name+'.log','w').write(p.stdout)
    print(json.dumps(dict(label=label,test=test,wall=elapsed,exit=p.returncode)),flush=True)
    print(p.stdout,flush=True)

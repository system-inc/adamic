import subprocess,time,json,os,sys,pathlib
label=sys.argv[1]
root=pathlib.Path('/workspace/split-evidence')
env=os.environ.copy();env.update(GOMAXPROCS='4',ADAMIC_GATE_UNCACHED='1',XDG_CACHE_HOME=str(root/(label+'-cache')))
command=['go','test','./internal/oracle','-count=1','-parallel='+sys.argv[3] if len(sys.argv)>3 else '-parallel=4','-json','-run',sys.argv[2] if len(sys.argv)>2 else '^(TestLoopCountersAgreeWithNode|TestCountsAreRecorded)$']
s=time.perf_counter()
with (root/(label+'.jsonl')).open('w') as f:r=subprocess.run(command,stdout=f,stderr=subprocess.STDOUT,env=env)
(root/(label+'.time')).write_text(json.dumps(dict(instrument='Python time.perf_counter',wall=time.perf_counter()-s,exit=r.returncode,command=command,env={k:env[k] for k in ['GOMAXPROCS','ADAMIC_GATE_UNCACHED','XDG_CACHE_HOME','GOCACHE']}),indent=2))
print(label,r.returncode,time.perf_counter()-s)

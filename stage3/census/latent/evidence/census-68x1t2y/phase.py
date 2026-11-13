import json, os, pathlib, subprocess, sys, time
name, phase, *command = sys.argv[1:]
root = pathlib.Path('/tmp/census-68x1t2y')
out = root/'results'/name
started=time.time()
with (out/(phase+'.log')).open('w') as log:
    result=subprocess.run(['timeout','--preserve-status','--kill-after=10s','1700',*command],stdout=log,stderr=subprocess.STDOUT)
row=dict(phase=phase,command=command,cwd=os.getcwd(),exit_code=result.returncode,wall_seconds=round(time.time()-started,3))
(out/(phase+'-status.json')).write_text(json.dumps(row,indent=2)+'\n')
print(json.dumps(row),flush=True)
sys.exit(result.returncode if result.returncode>=0 else 128-result.returncode)

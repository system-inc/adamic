import subprocess, pathlib, json, sys, os, time
out=pathlib.Path(__file__).resolve().parent
name=sys.argv[1]
started=time.monotonic()
with (out/(name+'.log')).open('w') as log:
    child=subprocess.run(sys.argv[2:],stdout=log,stderr=subprocess.STDOUT)
record=dict(command=sys.argv[2:],exit=child.returncode,seconds=time.monotonic()-started)
(out/(name+'.exit.json')).write_text(json.dumps(record,indent=2)+'\n')
print(name+' child exit='+str(child.returncode),flush=True)
sys.exit(child.returncode)

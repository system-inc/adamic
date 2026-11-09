"""Log a child without a pipe, print its actual exit code, and propagate it."""
import json,pathlib,subprocess,sys,time
log=pathlib.Path(sys.argv[1])
started=time.time()
with log.open('w') as output:
 result=subprocess.run(sys.argv[2:],stdout=output,stderr=subprocess.STDOUT)
record={'command':sys.argv[2:],'exit_code':result.returncode,'elapsed_seconds':time.time()-started}
log.with_suffix(log.suffix+'.json').write_text(json.dumps(record,indent=2)+'\n')
print(json.dumps(record),flush=True)
sys.exit(result.returncode)

"""Record the command and the child's own exit code; never mask a failed child."""
import json,os,pathlib,subprocess,sys,time
name,*command=sys.argv[1:]
out=pathlib.Path(__file__).resolve().parent
start=time.monotonic()
with (out/(name+'.stdout')).open('wb') as stdout,(out/(name+'.stderr')).open('wb') as stderr:
 result=subprocess.run(command,stdout=stdout,stderr=stderr)
(out/(name+'.run.json')).write_text(json.dumps({'command':command,'exit':result.returncode,'elapsed_seconds':time.monotonic()-start,'environment':{k:v for k,v in os.environ.items() if k.startswith('LATENT_')}},indent=2)+'\n')
sys.exit(result.returncode)

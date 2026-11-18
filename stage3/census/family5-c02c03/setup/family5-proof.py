import json,os,sys,subprocess,time
from pathlib import Path
out=Path(sys.argv[1]);out.mkdir(parents=True,exist_ok=True)
cmd=sys.argv[2:];start=time.monotonic()
with (out/'stdout.log').open('wb') as stdout,(out/'stderr.log').open('wb') as stderr:
 p=subprocess.Popen(cmd,stdout=stdout,stderr=stderr)
 code=p.wait()
r={'command':cmd,'cwd':os.getcwd(),'environment':{k:v for k,v in os.environ.items() if k.startswith(('LATENT_','CENSUS_','ADAMIC_','FAMILY5_'))},'exit_code':code,'wall_seconds':round(time.monotonic()-start,3)}
(out/'status.json').write_text(json.dumps(r,indent=2)+'\n')
print(('PASS' if code==0 else 'FAIL')+' '+str(out.name)+' child_exit='+str(code),flush=True)
sys.exit(code if code>=0 else 128-code)

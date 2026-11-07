from pathlib import Path
import json,os,signal,subprocess,time
s=Path('/workspace/scratch/arrayof');tree=s/'after';assert (s/'apply-after.exit').read_text().strip()=='0';assert json.loads((s/'CHECKPOINT.json').read_text())['remoteVerified'];phases={}
for label,command in [('install',['npm','ci','--no-audit','--no-fund']),('build',['npm','run','build:compiler'])]:
 start=time.monotonic()
 with (s/(label+'-after.log')).open('w') as log:
  p=subprocess.Popen(command,cwd=tree,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
  try:code=p.wait(timeout=1700)
  finally:
   try:os.killpg(p.pid,signal.SIGKILL)
   except ProcessLookupError:pass
   p.wait()
 phases[label]=dict(exit=code,seconds=time.monotonic()-start);(s/'after-build-phases.json').write_text(json.dumps(phases,indent=2)+'\n');assert code==0
print(json.dumps(phases),flush=True)

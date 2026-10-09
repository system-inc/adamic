import os, sys, subprocess, time, json, signal
from pathlib import Path
root=Path('/workspace/scratch/oracle-split')
label=sys.argv[1]
env=dict(os.environ, GOMAXPROCS='4', ADAMIC_GATE_UNCACHED='1', ADAMIC_ORACLE_WASI='1', WASI_SYSROOT='/workspace/adamic-tools/wasi-sdk-29.0-x86_64-linux/share/wasi-sysroot', XDG_CACHE_HOME=str(root/(label+'-cache')), GOCACHE=str(root/'go-cache'), TMPDIR='/tmp/adamic-gate')
env['PATH']='/workspace/adamic-tools/bin:/workspace/adamic-tools/go/bin:'+env['PATH']
command=['go','test','./internal/oracle','-run',sys.argv[2],'-parallel','4','-count=1','-timeout','90s','-json']
if label == 'before-installed': command.append('-overlay='+str(root/'baseline-overlay.json'))
start=time.monotonic()
with (root/(label+'.json')).open('w') as out, (root/(label+'.err')).open('w') as err:
 p=subprocess.Popen(command,env=env,stdout=out,stderr=err,start_new_session=True)
 cooked=False
 try: p.wait(timeout=90)
 except subprocess.TimeoutExpired:
  cooked=True
  parents={}
  for entry in Path('/proc').iterdir():
   if not entry.name.isdigit(): continue
   try:
    fields=(entry/'stat').read_text().rsplit(')',1)[1].split()
    parents[int(entry.name)]=int(fields[1])
   except (OSError,ValueError,IndexError): pass
  descendants={p.pid}
  while True:
   more={pid for pid,parent in parents.items() if parent in descendants}
   if more <= descendants: break
   descendants |= more
  for pid in sorted(descendants,reverse=True):
   try: os.kill(pid,signal.SIGKILL)
   except ProcessLookupError: pass
  try: os.killpg(p.pid, signal.SIGKILL)
  except ProcessLookupError: pass
  p.wait()
wall=time.monotonic()-start
result={'label':label,'wall_seconds':round(wall,3),'exit_code':p.returncode,'cooked':cooked,'hard_limit_seconds':90,'deadline_failure':cooked,'command':command,'empty_runtime_cache':True}
(root/(label+'.time.json')).write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result))
sys.exit(p.returncode)

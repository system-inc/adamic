import json,os,sys,socket,subprocess,time
from pathlib import Path
from supervise import run,scratch
out=Path(__file__).resolve().parent
while not (scratch/'source-clone-status.json').exists(): time.sleep(1)
assert json.loads((scratch/'source-clone-status.json').read_text())['exit_code']==0
for name,cmd in [
 ('source-checkout',['git','-C','/workspace/typescript-source','checkout','--detach','050880ce59e30b356b686bd3144efe24f875ebc8']),
 ('apply',['bash','stage3/apply.sh','--build-product','/workspace/bcbe5d12-adapted'])]:
 status=run(name,cmd,limit=5400)
 if status['exit_code']:sys.exit(status['exit_code'] if status['exit_code']>0 else 128-status['exit_code'])
pin=subprocess.run(['git','-C','/workspace/typescript-source','rev-parse','HEAD'],capture_output=True,text=True)
print('git rev-parse child_exit='+str(pin.returncode),flush=True);pin.check_returncode()
box={'hostname':socket.gethostname(),'nproc':os.cpu_count(),'cpu_quota':Path('/sys/fs/cgroup/cpu.max').read_text().strip(),'memory':next(l.strip() for l in Path('/proc/meminfo').read_text().splitlines() if l.startswith('MemTotal:'))}
(out/'provenance.json').write_text(json.dumps({'compiler_sha':'bcbe5d1244ee8a218d860415dea19cca114bae46','apply_sha':'bcbe5d1244ee8a218d860415dea19cca114bae46','typescript_pin':pin.stdout.strip(),'tree':'/workspace/bcbe5d12-adapted','box':box},indent=2)+'\n')
status=run('apply-checkpoint',[sys.executable,str(out/'checkpoint.py'),'completed apply'],limit=1800)
sys.exit(status['exit_code'] if status['exit_code']>=0 else 128-status['exit_code'])

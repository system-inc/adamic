import hashlib,json,os,socket,subprocess,sys
from pathlib import Path
from supervise import run,scratch
repo=Path('/workspace/adamic');out=Path(__file__).resolve().parent;base='c3b22184745d809cb281a37ef1989c895596223f';tree=Path('/workspace/c3b22184-adapted');pin='050880ce59e30b356b686bd3144efe24f875ebc8'
for name,cmd in [('setup',['bash','cloud/setup.sh']),('npm-ci',['npm','ci','--prefix','stage3/api'])]:
 status=run(name,cmd,limit=5400)
 if status['exit_code']:sys.exit(status['exit_code'] if status['exit_code']>0 else 128-status['exit_code'])
p=subprocess.run(['git','-C','/workspace/typescript-source','rev-parse','HEAD'],capture_output=True,text=True);print('source git rev-parse child_exit='+str(p.returncode),flush=True);p.check_returncode();assert p.stdout.strip()==pin
status=run('source-clean',['git','-C','/workspace/typescript-source','diff','--exit-code'],limit=1800);assert status['exit_code']==0
status=run('apply',['bash','stage3/apply.sh','--build-product',str(tree)],limit=5400)
if status['exit_code']:sys.exit(status['exit_code'] if status['exit_code']>0 else 128-status['exit_code'])
box={'hostname':socket.gethostname(),'nproc':os.cpu_count(),'cpu_quota':Path('/sys/fs/cgroup/cpu.max').read_text().strip(),'memory':next(l.strip() for l in Path('/proc/meminfo').read_text().splitlines() if l.startswith('MemTotal:'))}
(out/'provenance.json').write_text(json.dumps(dict(compiler_sha=base,apply_sha=base,typescript_pin=pin,tree=str(tree),box=box),indent=2)+'\n')
manifest={str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src').rglob('*')) if p.is_file()};(out/'adapted-source-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
status=run('apply-checkpoint',[sys.executable,str(out/'checkpoint.py'),'completed apply'],limit=1800)
sys.exit(status['exit_code'] if status['exit_code']>=0 else 128-status['exit_code'])

import os,sys,json,subprocess,hashlib,socket
from pathlib import Path
from runner import run,out,raw,repo
pin='050880ce59e30b356b686bd3144efe24f875ebc8'
run('source-checkout',['git','-C','/workspace/typescript-source','checkout','--detach',pin])
assert subprocess.check_output(['git','-C','/workspace/typescript-source','rev-parse','HEAD'],text=True).strip()==pin
os.environ['ADAMIC_TYPESCRIPT_SOURCE']='/workspace/typescript-source'
run('npm-ci',['npm','ci','--prefix','stage3/api'])
run('apply',['bash','stage3/apply.sh','--build-product','/workspace/require-row-adapted'])
tree=Path('/workspace/require-row-adapted')
(out/'provenance.json').write_text(json.dumps(dict(base=subprocess.check_output(['git','rev-parse','71b126b0'],text=True).strip(),head=subprocess.check_output(['git','rev-parse','738b3f31'],text=True).strip(),typescript_pin=pin,box=socket.gethostname(),cpu_quota=Path('/sys/fs/cgroup/cpu.max').read_text(),meminfo=Path('/proc/meminfo').read_text(),tree=str(tree)),indent=2)+'\n')
(out/'adapted-source-manifest.json').write_text(json.dumps({str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src').rglob('*')) if p.is_file()},indent=2)+'\n')
run('checkpoint-apply',[sys.executable,str(out/'checkpoint.py'),'completed apply'])

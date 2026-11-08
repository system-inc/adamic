"""Preserve logs, raw streams and all guarded overlay inputs inside report scope."""
from pathlib import Path
import gzip,json,hashlib,os,subprocess,socket
here=Path(__file__).parent; scratch=Path('/workspace/scratch/jawqwv7'); raw=here/'raw';raw.mkdir(exist_ok=True)
for p in scratch.iterdir():
 if p.is_file() and p.suffix in ('.log','.jsonl','.json'):
  (raw/(p.name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
for p in (scratch/'overlay').rglob('*'):
 if p.is_file():
  name=p.name+'.txt' if p.suffix=='.go' else p.name
  dest=here/'overlay'/p.relative_to(scratch/'overlay').parent;dest.mkdir(parents=True,exist_ok=True);(dest/name).write_bytes(p.read_bytes())
(here/'supervise.py').write_bytes((scratch/'supervise.py').read_bytes())
root=here.parents[2]
def git(*args):return subprocess.check_output(['git',*args],cwd=root,text=True).strip()
source=Path('/workspace/graph-boundary-adapted')
manifest={str(p.relative_to(source)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((source/'src').rglob('*.ts'))} if source.exists() else {}
(here/'source-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
provenance=dict(base_sha='bcbe5d1244ee8a218d860415dea19cca114bae46',reference_report_sha='30057eca',typescript_pin='050880ce59e30b356b686bd3144efe24f875ebc8',typescript_actual=subprocess.check_output(['git','-C','/workspace/typescript-source','rev-parse','HEAD'],text=True).strip(),box=socket.gethostname(),environment='ccarenv_b64_Y2NhcmVudl85M2MyY2I4NzA3ZTA4MTkxODMwYTI0NDZlY2ZhNjAyZg',cpu_quota=Path('/sys/fs/cgroup/cpu.max').read_text().strip(),memory_limit=Path('/sys/fs/cgroup/memory.max').read_text().strip(),visible_cpus=os.cpu_count(),compiler_head='bcbe5d1244ee8a218d860415dea19cca114bae46',report_worktree_head=git('rev-parse','HEAD'),cohere=git('-C',str(root/'cohere'),'rev-parse','HEAD'))
if (scratch/'census').exists():provenance['binary_sha256']=hashlib.sha256((scratch/'census').read_bytes()).hexdigest()
(here/'provenance.json').write_text(json.dumps(provenance,indent=2)+'\n')
files={str(p.relative_to(here)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(here.rglob('*')) if p.is_file() and p.name!='SHA256SUMS' and '.git' not in p.parts and '__pycache__' not in p.parts}
(here/'SHA256SUMS').write_text(''.join(f'{sha}  {name}\n' for name,sha in files.items()))
print('snapshot child_exit=0')

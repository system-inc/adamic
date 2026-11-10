from pathlib import Path
import shutil,subprocess,json,hashlib,sys
s=Path('/workspace/scratch/each-node');w=Path('/workspace/scratch/each-node-evidence');d=w/'each-node-evidence';d.mkdir(exist_ok=True)
for p in s.iterdir():
 if p.name.startswith('stage3-lane-') and not (s/'lane-summary.json').exists():continue
 if p.is_file() and p.suffix in ['.py','.cjs','.json','.log','.patch','.md','.gz'] and p.stat().st_size<15_000_000:shutil.copy2(p,d/p.name)
for name in ['control-counts','built-control-counts','throw-control-counts','counts']:
 if name=='counts' and not (s/'lane-summary.json').exists():continue
 if (s/name).exists():shutil.copytree(s/name,d/name,dirs_exist_ok=True)
for name in ['each-node-setup.log','each-node-api.log','each-node-typescript-deps.log','each-node-source.log']:
 p=Path('/tmp')/name
 if p.exists():shutil.copy2(p,d/p.name)
(d/'BASE.md').write_text('Task 8n7eet3, site 2. Base origin/main: e24f2600a5d8a8521c5128827cad395d47c81c81. TypeScript: 050880ce59e30b356b686bd3144efe24f875ebc8.\nRaw evidence only. No incomplete run is claimed as passing.\n')
(d/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+str(p.relative_to(d))+'\n' for p in sorted(d.rglob('*')) if p.is_file() and p.name!='SHA256SUMS'))
for command in [['git','add','each-node-evidence'],['git','commit','-m',sys.argv[1]],['git','push','origin','HEAD:refs/heads/typescript/each-node-evidence']]:
 p=subprocess.run(command,cwd=w);print('child exit',p.returncode,flush=True)
 if p.returncode:raise SystemExit(p.returncode)

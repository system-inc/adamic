from pathlib import Path
import shutil,os
old=Path('/workspace/scratch/diag-read-view-1');s=Path('/workspace/scratch/diag-read-view-2')
for name in ['TypeScript','gocache']:
 (s/name).symlink_to(old/name,target_is_directory=True)
for name in ['latent','adamic']:shutil.copyfile(old/name,s/name);os.chmod(s/name,0o700)
for name in ['controls.cjs','stock.cjs','javascript.cjs','pair.py']:
 (s/name).write_text((old/name).read_text().replace(str(old),str(s)))
for side in ['main','candidate']:
 shutil.copytree(old/(side+'-tree'),s/(side+'-tree'),symlinks=True,ignore=shutil.ignore_patterns('.git','node_modules','built'))
(s/'run-proof.py').write_text((old/'run-proof.py').read_text().replace(str(old),str(s)).replace("NODE_PATH=str(r/'stage3/api/node_modules')","NODE_PATH=str(r/'stage3/api/node_modules'),PYTHONDONTWRITEBYTECODE='1',npm_config_cache=str(s/'npm-cache')").replace('sys.exit(max(codes))','sys.exit(0 if all(c==0 for c in codes) else 1)'))

from pathlib import Path
import time,subprocess,shutil,hashlib,difflib,json
root=Path('/workspace/scratch/cast-tags');e=root/'evidence'
while not (e/'c-probe-owners.json').exists():time.sleep(5)
for side in ['main','c']:
 tree=root/(side+'-tree')
 if side=='c':
  shutil.copytree(root/'main-tree/scripts',tree/'scripts')
  for name in ['Herebyfile.mjs','package.json','package-lock.json']:shutil.copy2(root/'main-tree'/name,tree/name)
 (tree/'node_modules').symlink_to(root/'typescript/node_modules',target_is_directory=True)
 with (e/(side+'-dts.log')).open('w') as log:
  p=subprocess.run([str(tree/'node_modules/.bin/hereby'),'dts'],cwd=tree,stdout=log,stderr=subprocess.STDOUT)
 (e/(side+'-dts-status.json')).write_text(json.dumps({'exit':p.returncode})+'\n')
 assert p.returncode==0,(side,p.returncode)
 for name in ['typescript.d.ts','tsserverlibrary.d.ts']:shutil.copy2(tree/'built/local'/name,e/(side+'-'+name))
 print(side,'dts complete',flush=True)
for name in ['typescript.d.ts','tsserverlibrary.d.ts']:
 a=(e/('main-'+name)).read_bytes();b=(e/('c-'+name)).read_bytes()
 diff=''.join(difflib.unified_diff(a.decode().splitlines(True),b.decode().splitlines(True),fromfile='main/'+name,tofile='readonly-kind/'+name))
 (e/(name+'.diff')).write_text(diff)
 print(name,hashlib.sha256(a).hexdigest(),hashlib.sha256(b).hexdigest(),flush=True)

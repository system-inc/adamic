from pathlib import Path
import json,os,signal,subprocess,shutil,time
s=Path('/workspace/scratch/arrayof');repo=Path('/workspace/adamic');tree=s/'after';assert all(v['exit']==0 for v in json.loads((s/'after-build-phases.json').read_text()).values())
roots=sorted((tree/'src/compiler').rglob('*.ts'));assert len(roots)==79;(s/'roots-after.json').write_text(json.dumps(list(map(str,roots)),indent=2)+'\n')
with (s/'types-after.log').open('w') as log:
 p=subprocess.Popen([str(s/'adamic'),'types',*map(str,roots)],cwd=tree,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
 try:code=p.wait(timeout=1700)
 finally:
  try:os.killpg(p.pid,signal.SIGKILL)
  except ProcessLookupError:pass
  p.wait()
(s/'types-after-result.json').write_text(json.dumps(dict(exit=code,roots=len(roots),logBytes=(s/'types-after.log').stat().st_size),indent=2)+'\n');assert code==0
for label in ['after']:
 with (s/('census-'+label+'-driver.log')).open('w') as log:subprocess.run(['python3',str(s/'census.py'),label,str(tree)],cwd=repo,stdout=log,stderr=subprocess.STDOUT,check=True)
mutant=s/'mutant';shutil.copytree(tree,mutant,symlinks=True,ignore=shutil.ignore_patterns('.git','node_modules','built'));(mutant/'node_modules').symlink_to(tree/'node_modules',target_is_directory=True);pin=json.loads((repo/'stage3/adapt/99-correctness-core-arrayof/owners.json').read_text())[0];p=mutant/'src/compiler/core.ts';raw=p.read_bytes();before=pin['before'].encode();after=pin['after'].encode()
if b'\r\n' in raw:before=before.replace(b'\n',b'\r\n');after=after.replace(b'\n',b'\r\n')
assert raw.count(after)==1;p.write_bytes(raw.replace(after,before));assert p.read_bytes()==(s/'before/src/compiler/core.ts').read_bytes();(s/'mutant-restore.json').write_text(json.dumps(dict(owner='arrayOf',exactOriginalOwnerRestored=True,coreByteIdenticalToBaseline=True),indent=2)+'\n')
with (s/'census-mutant-driver.log').open('w') as log:subprocess.run(['python3',str(s/'census.py'),'mutant',str(mutant)],cwd=repo,stdout=log,stderr=subprocess.STDOUT,check=True)
with (s/'compare-census.log').open('w') as log:subprocess.run(['python3',str(s/'compare-census.py')],cwd=repo,stdout=log,stderr=subprocess.STDOUT,check=True)
print('after types, full-core census, original-owner mutant: pass',flush=True)

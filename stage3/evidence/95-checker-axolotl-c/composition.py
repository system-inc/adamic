from pathlib import Path
import json,os,shutil,subprocess,hashlib
base=Path('/workspace/scratch/95a-tree')
out=Path('/workspace/scratch/95-composition');out.mkdir()
adapters={
 'b':Path('/workspace/scratch/95-peer-adapters/stage3/adapt/95-checker-axolotl-b/adapt.cjs'),
 'c':Path('/workspace/adamic/stage3/adapt/95-checker-axolotl-c/adapt.cjs'),
 'd':Path('/workspace/scratch/95-peer-adapters/stage3/adapt/95-checker-axolotl-d/adapt.cjs'),
 'e':Path('/workspace/scratch/95e-repo/stage3/adapt/95-checker-axolotl-e/adapt.cjs')}
env=dict(os.environ,NODE_PATH='/workspace/adamic/stage3/api/node_modules')
def sources(tree):
 return {str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src').rglob('*')) if p.is_file()}
def build(order):
 tree=out/''.join(order)
 shutil.copytree(base,tree,ignore=shutil.ignore_patterns('.git','node_modules','built','tests'))
 for index,name in enumerate(order):
  with (out/(''.join(order)+'-'+str(index)+'-'+name+'.log')).open('w') as log:
   subprocess.run(['node',str(adapters[name]),str(tree)],stdout=log,stderr=subprocess.STDOUT,check=True,env=env)
 return tree,sources(tree)
report=[]
for pair in [('b','c'),('d','c'),('c','e'),('b','c','d','e')]:
 left,lmap=build(pair);right,rmap=build(tuple(reversed(pair)))
 assert lmap==rmap,('noncommuting source edits',pair)
 report.append(dict(forward=''.join(pair),reverse=''.join(reversed(pair)),sourceFiles=len(lmap),sourcesByteIdentical=True))
(out/'composition.json').write_text(json.dumps(dict(refs=dict(b='408f8cb16c4e97ca51997d5a340a0b1826b66ba9',d='c6c2c01e6b0ebfe7fd8b767119c489a1843cdbc5'),orders=report),indent=2)+'\n')
print(json.dumps(report,indent=2))

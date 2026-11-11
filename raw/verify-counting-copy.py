import hashlib,json,os
from pathlib import Path
raw=Path(__file__).parent
def inventory(root):
 result={}
 def walk(directory,relative=Path()):
  for entry in os.scandir(directory):
   rel=relative/entry.name
   if len(rel.parts)==1 and entry.name in ('.git','node_modules'):continue
   if entry.is_symlink():result[str(rel)]=['symlink',os.readlink(entry.path)]
   elif entry.is_dir():result[str(rel)]=['directory',''];walk(entry.path,rel)
   else:result[str(rel)]=['file',hashlib.sha256(Path(entry.path).read_bytes()).hexdigest()]
 walk(root);return result
a=inventory('/workspace/scratch/meta-spread-after');b=inventory('/workspace/scratch/meta-spread-counting-final');assert a.keys()==b.keys()
changed=sorted(name for name in a if a[name]!=b[name]);assert changed==['built/local/_tsc.js','built/local/run.js','built/local/typescript.js'],changed
fixtures=[name for name in a if 'node_modules' in Path(name).parts];assert fixtures
result=dict(passed=True,entries=len(a),nestedNodeModulesEntries=len(fixtures),changed=changed,allOtherBytesAndSymlinkTargetsIdentical=True,rootNodeModulesDeliberatelyOmitted=True)
(raw/'counting-copy-verification.json').write_text(json.dumps(result,indent=2)+'\n');print(result)

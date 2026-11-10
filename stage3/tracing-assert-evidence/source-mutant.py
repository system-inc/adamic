from pathlib import Path
import time,shutil,json,hashlib,os
from run import run,raw
before=Path('/workspace/scratch/baobab-before');after=Path('/workspace/scratch/baobab-after');mutant=Path('/workspace/scratch/baobab-mutant')
while not (raw/'after-root-census.json').exists():time.sleep(1)
assert json.loads((raw/'after-types.json').read_text())['exit']==0
names=lambda tree:sorted(str(p.relative_to(tree)) for p in (tree/'src').rglob('*') if p.is_file())
files=names(before);assert names(after)==files
changed=[name for name in files if (before/name).read_bytes()!=(after/name).read_bytes()];assert changed==['src/compiler/tracing.ts'],changed
b=(before/changed[0]).read_bytes();a=(after/changed[0]).read_bytes();old=b'Debug.assert(tracing, "Tracing is not in progress")';new=b'Debug.assert(!!tracing, "Tracing is not in progress")';assert b.count(old)==1;assert a==b.replace(old,new)
mutant.mkdir();shutil.copytree(after/'src',mutant/'src')
for name in ['package.json','tsconfig.json']:
 if (after/name).exists():shutil.copyfile(after/name,mutant/name)
(mutant/changed[0]).write_bytes(a.replace(new,old));assert (mutant/changed[0]).read_bytes()==b
assert all((mutant/name).read_bytes()==(before/name).read_bytes() for name in files)
(raw/'source-validation.json').write_text(json.dumps(dict(files=len(files),changed=changed,onlyArgumentChanged=True,messageAndOtherBytesIdentical=True,mutantRestoresEverySourceByte=True,beforeSha256=hashlib.sha256(b).hexdigest(),afterSha256=hashlib.sha256(a).hexdigest()),indent=2)+'\n')
env=dict(os.environ,BAOBAB_CENSUS_ROOT=str(mutant/'src/compiler/core.ts'))
run('mutant-root-census',[raw/'root-census',mutant/'src/compiler',raw/'mutant-root-census.jsonl'],env=env)
d=json.loads((raw/'mutant-root-census.jsonl').read_text());assert d['kind']=='NotYet';assert d['where'].endswith('/tracing.ts:110:22');assert d['reason']=="a namespace value seen as an argument to a callee whose body can't be named"
print('restoring mutant reproduces tracing.ts:110:22',flush=True)

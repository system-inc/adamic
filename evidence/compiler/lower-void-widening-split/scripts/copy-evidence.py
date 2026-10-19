from pathlib import Path
import shutil,json
source=Path('/tmp/void-split-evidence');target=Path('/tmp/adamic-void-split-evidence-worktree/evidence/compiler/lower-void-widening-split')
for file in source.iterdir():
 if not file.is_file():continue
 name=file.name
 if file.suffix in ['.log','.stdout','.stderr']:
  destination=target/'logs'/name
 elif file.suffix in ['.py','.cjs']:
  destination=target/'scripts'/name
 elif file.suffix=='.go':
  destination=target/'mutants'/(name+'.txt')
 elif name.endswith('-overlay.json'):
  destination=target/'mutants'/name
 elif file.suffix in ['.ts','.a']:
  destination=target/'probes'/(name+'.txt')
 elif file.suffix=='.mjs':
  destination=target/'generated'/name
 elif file.suffix in ['.json','.jsonl','.txt','.md']:
  destination=target/name
 else:continue
 destination.parent.mkdir(parents=True,exist_ok=True)
 shutil.copyfile(file,destination)
print('copied',sum(1 for p in target.rglob('*') if p.is_file()),'files')

from pathlib import Path
import shutil,json
source=Path('/workspace/scratch/baobab-evidence');target=Path('/workspace/scratch/baobab-evidence-branch/stage3/tracing-assert-evidence');target.mkdir(parents=True,exist_ok=True)
suffixes={'.json','.jsonl','.log','.txt','.sha','.py','.cjs','.patch','.attempt1'}
for p in source.iterdir():
 if p.is_file() and p.suffix in suffixes:shutil.copyfile(p,target/p.name)
for name in ['baseline-control-counts','control-counts','lane-counts','lane-counts.attempt1','root-overlay']:
 directory=source/name
 if directory.exists():
  for p in directory.rglob('*'):
   if p.is_file() and p.suffix in suffixes:
    q=target/name/p.relative_to(directory);q.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(p,q)
for p in Path('/tmp').glob('baobab-*.log'):shutil.copyfile(p,target/p.name)
(target/'STATUS.md').write_text('# baobab evidence checkpoint\n\nVerification is in progress. The lane logs in this checkpoint may be partial.\n\nBoth applies, both 79-root type checks, both normal builds, the paired Node witness, source-byte scope, restoring mutant, next wall, all JavaScript hunks and export-key comparison have passed.\n')
print('evidence files copied:',sum(1 for p in target.rglob('*') if p.is_file()))

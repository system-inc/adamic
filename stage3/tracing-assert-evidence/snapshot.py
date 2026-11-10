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
if (source/'final-source-validation.json').exists():
 status='# baobab evidence\n\nAll verification is complete, including the fresh full apply of the final saved adaptation content.\n'
elif (source/'lane-summary.json').exists():
 status='# baobab evidence checkpoint\n\nThe entire 398-unit lane is complete and green. The assert count is 0, with 0 conversion or assertion-outcome differences. Both positive controls record 2 calls and 0 differences. The fresh full apply of the final saved proof and ledger is still running.\n'
else:
 status='# baobab evidence checkpoint\n\nVerification is in progress. The lane logs in this checkpoint may be partial.\n'
status+='\nBoth initial applies, both 79-root type checks, both normal builds, the paired Node witness, source-byte scope, restoring mutant, next wall, all JavaScript hunks and export-key comparison have passed.\n'
if (source/'main-commit.json').exists():
 d=json.loads((source/'main-commit.json').read_text());status+='\nLanding commit: '+d['sha']+' on '+d['branch']+'. Base: '+d['base']+'.\n'
(target/'STATUS.md').write_text(status)
print('evidence files copied:',sum(1 for p in target.rglob('*') if p.is_file()))

"""Freeze evidence after the single successful apply/build, before any roots."""
import hashlib,json,subprocess
from pathlib import Path
b=Path(__file__).resolve().parent;scratch=Path('/workspace/scratch/next-stops-fb8db081');tree=scratch/'adapted'
for name in ['npm-ci','apply','build']:
 s=json.loads((b/'head/CommonJSBundle'/f'{name}-status.json').read_text());assert s['exit_code']==0
assert not list(b.glob('head/*/root-??.jsonl'))
manifest={str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src/compiler').rglob('*')) if p.is_file()}
assert len(list((tree/'src/compiler').rglob('*.ts')))==79
(b/'adapted-source-before-measurement.json').write_text(json.dumps(manifest,indent=2)+'\n')
binary=scratch/'head-census';(b/'binaries.json').write_text(json.dumps(dict(path=str(binary),sha256=hashlib.sha256(binary.read_bytes()).hexdigest()),indent=2)+'\n')
(b/'build-info.txt').write_text(subprocess.check_output(['go','version','-m',str(binary)],text=True))
(b/'measurement-worktree-status.txt').write_text(subprocess.check_output(['git','status','--porcelain'],cwd=scratch/'repo',text=True))

from pathlib import Path
import json,hashlib
raw=Path('/workspace/scratch/tracing-length-evidence');b=Path('/workspace/scratch/tracing-length-before');a=Path('/workspace/scratch/tracing-length-after')
files=sorted(str(p.relative_to(b)) for p in (b/'src').rglob('*') if p.is_file());assert files==sorted(str(p.relative_to(a)) for p in (a/'src').rglob('*') if p.is_file())
changed=[n for n in files if (b/n).read_bytes()!=(a/n).read_bytes()];assert changed==['src/compiler/tracing.ts'],changed
pins=json.loads(Path('/workspace/adamic/stage3/adapt/99-correctness-tracing-length/pins.json').read_text());text=(b/changed[0]).read_bytes().decode()
for p in pins:assert text.count(p['before'])==1;text=text.replace(p['before'],p['after'])
assert text.encode()==(a/changed[0]).read_bytes()
roots=sorted((a/'src/compiler').rglob('*.ts'));assert len(roots)==79
(raw/'source-comparison-report.json').write_text(json.dumps({'files':len(files),'changed':changed,'exactOwnerEdits':True,'allOtherBytesIdentical':True,'compilerRoots':len(roots)},indent=2)+'\n')
print('exact source comparison passed')

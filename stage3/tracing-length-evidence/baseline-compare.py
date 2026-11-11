from pathlib import Path
from run import raw
import json
assert json.loads((raw/'baseline-clean-apply.json').read_text())['exit']==0
b=Path('/workspace/scratch/tracing-length-before');c=Path('/workspace/scratch/tracing-length-before-clean')
files=sorted(str(p.relative_to(b)) for p in (b/'src').rglob('*') if p.is_file())
assert files==sorted(str(p.relative_to(c)) for p in (c/'src').rglob('*') if p.is_file())
assert all((b/n).read_bytes()==(c/n).read_bytes() for n in files)
(raw/'baseline-clean-source.json').write_text(json.dumps({'sourceFiles':len(files),'allBytesIdenticalToBuiltBaseline':True,'cleanApplyExit':0},indent=2)+'\n')
print('clean baseline apply reproduces every built-baseline source byte')

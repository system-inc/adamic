# Scratch-only experiment: never use this deletion in a freezing program.
from pathlib import Path
import re
p=Path('scratch/emitter-speed/no-freeze');p.mkdir(exist_ok=True)
s=Path('scratch/emitter-speed/cold/after.c').read_text()
if 'adamic_object_freeze(' in s:raise RuntimeError('parse program can freeze')
(p/'before.c').write_text(s)
s,n=re.subn(r' \|\| (adamic_\w+)->frozen(?=\))','',s)
if n!=424:raise RuntimeError('baseline guard count differs')
(p/'after.c').write_text(s)
(p/'removed.txt').write_text(str(n)+' public frozen conditions removed\n')

#!/usr/bin/env python3
"""Remove only the unavailable parent guard in a scratch Go overlay."""
import json, subprocess, tempfile
from pathlib import Path
repo=Path(__file__).resolve().parents[3]
source=repo/'internal/lower/interface_cast.go'
text=source.read_text()
check='\tif err := l.optionalObjectViewBoundary(node, target); err != nil {\n\t\treturn nil, err\n\t}\n'
assert text.count(check)==1
scratch=Path(tempfile.mkdtemp(prefix='optional-object-parent-mutant-'))
changed=scratch/'interface_cast.go';changed.write_text(text.replace(check,''))
overlay=scratch/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(source):str(changed)}}))
log=scratch/'result.log'
with log.open('w') as output:
 result=subprocess.run(['go','test','-overlay',str(overlay),'./internal/lower','-run','^TestOptionalObjectIncompleteParentRefused$','-count=1','-parallel=1','-v'],cwd=repo,stdout=output,stderr=subprocess.STDOUT)
observed=log.read_text()
assert result.returncode!=0 and 'placeholder must not erase a child check: <nil>' in observed,observed
print('parent descriptor guard omission: caught by TestOptionalObjectIncompleteParentRefused')
print(log)

#!/usr/bin/env python3
"""Rule-local diagnostic comparison overlay; shared tracked files are untouched."""
import json, subprocess
from pathlib import Path
owned = Path(__file__).resolve().parent
repository = owned.parents[4]
subprocess.run(['python3', str(repository/'stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/validate.py')], check=True)
scratch=Path('/tmp/lint-wave1-10-next')
p=scratch/'oracle.go'
s=p.read_text()
# Deliberately diagnostics-only for the two rules whose disjoint fixes cannot be
# serialized by the shared Finding. Their real Go rules are still run unchanged.
s=s.replace('for _, d := range diagnostics {\n\t\tstart, end', 'for _, d := range diagnostics {\n        if fields[1]=="arrow-body-style" || fields[1]=="no-extra-bind" { d.Fixes=nil; d.Suggestions=nil }\n\t\tstart, end')
s=s.replace('\tfor _, d := range diagnostics { if len(d.Fixes)', '\tif fields[1]=="arrow-body-style" || fields[1]=="no-extra-bind" { fmt.Fprintf(out,"fixed\\t%s\\n",written(source)); return len(diagnostics) }\n\tfor _, d := range diagnostics { if len(d.Fixes)')
p.write_text(s)
overlay=json.loads((scratch/'overlay.json').read_text())
overlay['Replace'][str(repository/'stage1/cohere/lint/owned_next_test.go')]=str(owned/'validation_test.go.txt')
(scratch/'overlay.json').write_text(json.dumps(overlay))
print('max-lines: full findings and unchanged text; arrow-body-style/no-extra-bind: diagnostics only, multi-edit integration blocked')

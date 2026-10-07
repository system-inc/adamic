#!/usr/bin/env python3
"""Rule-owned compatibility overlay; no tracked shared file edits."""
from pathlib import Path
import json,subprocess
owned=Path(__file__).resolve().parent
repository=owned.parents[4]
subprocess.run(['python3',str(repository/'stage1/cohere/lint/rules/max-lines/validate.py')],check=True)
p=Path('/tmp/lint-wave1-10-next/overlay.json')
r=json.loads(p.read_text())
r['Replace'][str(repository/'stage1/cohere/lint/owned_fourth_test.go')]=str(owned/'validation_test.go.txt')
p.write_text(json.dumps(r))
print('fourth overlay ready; only older arrow/bind comparisons are diagnostics-only')

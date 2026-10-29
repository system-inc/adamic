#!/usr/bin/env python3
"""Drop only the discarded-write guard; require the named refusal test red."""
import json
from pathlib import Path
import subprocess
import tempfile

root = Path(__file__).resolve().parents[3]
source = root / 'internal/lower/library_node_host.go'
before = 'if outer.Parent == nil || outer.Parent.Kind != ast.KindExpressionStatement {'
with tempfile.TemporaryDirectory(prefix='stdout-value-mutant-') as tmp:
    changed = Path(tmp) / 'library_node_host.go'
    text = source.read_text()
    if text.count(before) != 1:
        raise SystemExit('discard guard must match exactly once')
    changed.write_text(text.replace(before, 'if false && (outer.Parent == nil || outer.Parent.Kind != ast.KindExpressionStatement) {', 1))
    overlay = Path(tmp) / 'overlay.json'
    overlay.write_text(json.dumps({'Replace': {str(source): str(changed)}}))
    command = ['go', 'test', '-overlay', str(overlay), './internal/lower', '-count=1', '-p', '1', '-parallel', '2', '-timeout', '2m', '-run', '^TestNodeStdoutWriteRefusals$', '-v']
    result = subprocess.run(command, cwd=root, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, timeout=180, text=True)
    print(result.stdout, end='')
    if result.returncode != 1 or '--- FAIL: TestNodeStdoutWriteRefusals' not in result.stdout or 'want its boolean is backpressure' not in result.stdout:
        raise SystemExit('mutant did not red the value refusal')
    print('caught: TestNodeStdoutWriteRefusals rejects admitting the boolean value')

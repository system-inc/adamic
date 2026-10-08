#!/usr/bin/env python3
"""Create a guarded scratch probe that bypasses diagnostics for one exact unit."""
import json
from pathlib import Path
import sys

source, output = map(lambda name: Path(name).resolve(), sys.argv[1:3])
output.mkdir(parents=True, exist_ok=True)
metadata = json.loads((source / 'overlay.json').read_text())
keys = [key for key in metadata['Replace'] if key.endswith('/internal/load/latent_hook.go')]
if len(keys) != 1:
    raise SystemExit('probe: expected exactly one internal/load/latent_hook.go')
key = keys[0]
text = Path(metadata['Replace'][key]).read_text()
needle = 'func (p *Program) LatentDiagnosticsIn(node *ast.Node) []string {'
if text.count(needle) != 1:
    raise SystemExit('probe: expected exactly one Program.LatentDiagnosticsIn')
text = text.replace(needle, needle + '''
    if node != nil && node.Parent != nil && os.Getenv("LATENT_PROBE_UNIT") != "" && p.Where(node.Parent) == os.Getenv("LATENT_PROBE_UNIT") {
        return nil
    }
''', 1)
hook = output / 'probe_load_hook.go'
hook.write_text(text)
metadata['Replace'][key] = str(hook)
(output / 'overlay.json').write_text(json.dumps(metadata, indent=2) + '\n')

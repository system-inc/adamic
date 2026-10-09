"""Recreate the developer probes at the requested revision against the current compiler."""
import json
import subprocess
from pathlib import Path

scratch = Path('/tmp/optional-widening-tools')
paths = {}
for name in ['internal/refusalprobe/catalog.go', 'internal/refusalprobe/probe.go',
             'internal/refusalprobe/probe_test.go', 'cmd/adamic-refusals/main.go']:
    target = scratch / name
    target.parent.mkdir(parents=True, exist_ok=True)
    source = subprocess.check_output(['git', 'show', '5e3c7b7:' + name]).decode()
    if name.endswith('/probe.go'):
        source = source.replace('"checkOverrides": "method-override",',
                                '"refuseOptionalWidening": "optional-widening", "checkOverrides": "method-override",')
    target.write_text(source)
    paths[str(Path(name).resolve())] = str(target)
(scratch / 'overlay.json').write_text(json.dumps({'Replace': paths}))
print(scratch / 'overlay.json')

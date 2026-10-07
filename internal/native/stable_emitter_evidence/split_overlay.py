# Prepare developer-tools' unchanged split implementation without editing native.go
# or adding production splitter files. Run from the repository root.
from pathlib import Path
import json
import subprocess
root = Path.cwd()
scratch = Path('/tmp/stable-names-split')
scratch.mkdir(exist_ok=True)
revision = '14e8372816b1d3bf5bcc37ca57336e46d41b7a41'
replacements = {}
for filename in ['native.go', 'units.go']:
    destination = scratch / filename
    destination.write_bytes(subprocess.check_output(['git', 'show', revision + ':internal/native/' + filename]))
    replacements[str(root / 'internal/native' / filename)] = str(destination)
path = scratch / 'overlay.json'
path.write_text(json.dumps({'Replace': replacements}))
print(path)

"""Scratch-only file selection; retain the complete entry's loader and prepasses."""
import json
from pathlib import Path
import sys

repository, overlay = map(lambda name: Path(name).resolve(), sys.argv[1:3])
path = overlay / 'overlay.json'
metadata = json.loads(path.read_text())
key = str(repository / 'internal/lower/latent_hook.go')
source = Path(metadata['Replace'][key]).read_text()
needle = 'for _, file := range program.Files() {'
assert source.count(needle) == 1
source = source.replace(needle, needle + '\n if !latentFileSelected(program.FileName(file)) { continue }', 1)
target = overlay / 'selected_hook.go.txt'
target.write_text(source)
metadata['Replace'][key] = str(target)
path.write_text(json.dumps(metadata, indent=2) + '\n')

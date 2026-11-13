"""Restore the source even on failure, and return the unchanged guard's exit code."""
import os
from pathlib import Path
import subprocess
import sys

root = Path('/workspace/adamic')
kind = sys.argv[1]
path, old, new, test = {
 'discarded': ('stage3/census/latent/units.go.txt', 'orderErr = base.collectViewedElements(modules)', 'base.collectViewedElements(modules)', 'TestLatentPrepassErrorsAreRecorded'),
 'omission': ('stage3/census/latent/lower.go.txt', 'l.noteOmittedOptionals(modules)', '', 'TestLatentCensusPrepassesTrackProductionSygfnmc'),
 'added': ('internal/lower/lower.go', 'lowering.noteOmittedOptionals(modules)', 'lowering.noteOmittedOptionals(modules)\n lowering.fakeCensusPrepass(modules)', 'TestLatentCensusPrepassesTrackProductionSygfnmc'),
}[kind]
path = root / path
source = path.read_text()
assert old in source
try:
    path.write_text(source.replace(old, new, 1))
    result = subprocess.run(['timeout', '2m', '/workspace/adamic-tools/go/bin/go', 'test', './stage3/census/latent', '-run', '^' + test + '$', '-count=1', '-v', '-timeout=90s'], cwd=root)
finally:
    path.write_text(source)
print(f'{kind}: guard child exit {result.returncode}', flush=True)
sys.exit(result.returncode)

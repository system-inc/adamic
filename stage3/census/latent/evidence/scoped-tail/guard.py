"""Run the source guard against real lowering plus the generated overlay helpers."""
import json
import os
from pathlib import Path
import signal
import subprocess
import sys

repository, overlay = map(lambda name: Path(name).resolve(), sys.argv[1:3])
metadata = json.loads((overlay / 'overlay.json').read_text())['Replace']
source = repository / 'internal/lower'
files = sorted(str(path) for path in source.glob('*.go') if not path.name.endswith('_test.go'))
files += [name for name in metadata
          if Path(name).parent == source and not Path(name).exists()
          and (not name.endswith('_test.go') or name.endswith('latent_scoped_tail_test.go'))]
command = ['go', 'test', '-overlay=' + str(overlay / 'overlay.json'),
           '-run', '^TestLatentScopedTailRecordStopOwnsSnapshotLoom03dg$',
           '-count=1', '-timeout=25s', *files]
code = subprocess.run(command, cwd=repository).returncode
print(json.dumps({'child_exit_code': code}), flush=True)
if code < 0:
    if -code not in (signal.SIGKILL, signal.SIGSTOP):
        signal.signal(-code, signal.SIG_DFL)
    os.kill(os.getpid(), -code)
sys.exit(code)

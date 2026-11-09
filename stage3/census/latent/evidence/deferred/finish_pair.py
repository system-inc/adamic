"""Pair the completed children only after both have reported their own success."""
import hashlib
import json
from pathlib import Path
import subprocess
import sys
import time

folder = Path(__file__).resolve().parent
root = folder.parents[4]
adapted = Path('/workspace/scratch/zb0q0zd-adapted')
while True:
    statuses = []
    for side in ('base', 'head'):
        path = folder / ('first-error-'+side+'.status.json')
        try:
            statuses.append(json.loads(path.read_text()))
        except (FileNotFoundError, json.JSONDecodeError):
            break
    if len(statuses) == 2:
        break
    time.sleep(5)
for status in statuses:
    if status['child_exit'] != 0:
        raise SystemExit('census failed: '+json.dumps(status))
expected = json.loads((folder/'source-hashes.json').read_text())
actual = {str(path.relative_to(adapted)): hashlib.sha256(path.read_bytes()).hexdigest()
          for path in sorted((adapted/'src').rglob('*')) if path.is_file() and path.suffix in ('.ts','.a')}
if actual != expected:
    raise SystemExit('adapted bytes changed between the paired runs')
def stream_path(side):
    path = folder/(side+'.jsonl')
    return path if path.exists() else folder/(side+'.jsonl.gz')

command = [sys.executable, str(root/'stage3/census/pair/pair.py'), str(stream_path('base')), str(stream_path('head')),
           str(adapted), '--typescript', str(root/'stage3/api/node_modules/typescript'),
           '--json', str(folder/'paired.json'), '--report', str(folder/'paired.md')]
with (folder/'pair.log.txt').open('w') as log:
    result = subprocess.run(command, stdout=log, stderr=subprocess.STDOUT)
(folder/'pair.status.json').write_text(json.dumps(dict(command=command,child_exit=result.returncode),indent=2)+'\n')
if result.returncode:
    raise SystemExit(result.returncode)
with (folder/'summary.log.txt').open('w') as log:
    result = subprocess.run([sys.executable,str(folder/'summarize.py'),str(folder)],stdout=log,stderr=subprocess.STDOUT)
print('summary child exit:',result.returncode,flush=True)
raise SystemExit(result.returncode)

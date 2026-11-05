"""Run and await the pinned head census in the foreground, reporting timer progress."""
import datetime
import hashlib
import json
import os
from pathlib import Path
import signal
import socket
import subprocess
import sys
import time

out = Path(__file__).resolve().parent
e = out / 'evidence'
prefix = e / 'rerun-head'
tree = out / 'scratch/adapted'
binary = out / 'scratch/head-bin/latent-census'
expected = json.loads((e / 'source-manifest.json').read_text())
actual = {str(p.relative_to(tree)): hashlib.sha256(p.read_bytes()).hexdigest()
          for p in sorted((tree / 'src').rglob('*')) if p.is_file()}
assert actual == expected
assert hashlib.sha256(binary.read_bytes()).hexdigest() == json.loads((e / 'provenance.json').read_text())['head']
assert not Path(str(prefix) + '-latent.jsonl').exists(), 'refusing to replace a rerun'
now = lambda: datetime.datetime.now(datetime.timezone.utc).isoformat()
meta = dict(start_utc=now(), host=socket.gethostname(), compiler_commit='66fff4ecc05c3bfb2258cf4bb2a7b0a2b762c59b',
            binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(), source_hashes_matched=len(actual),
            mode='first-error', launch='foreground supervisor awaiting its child; no detached/background shell job')
Path(str(prefix) + '-memory-before.txt').write_text(Path('/sys/fs/cgroup/memory.events').read_text())
Path(str(prefix) + '-start.txt').write_text(meta['start_utc'] + '\n')
env = dict(os.environ, LATENT_FULL='0', LATENT_ASSERT_NO_OUTPUT='1',
           LATENT_TIMES=str(prefix) + '-times.jsonl', TMPDIR=str(out / 'scratch/tmp'))
progress = Path(str(prefix) + '-progress.log').open('w', buffering=1)
def announce(message):
    line = now() + ' ' + message
    progress.write(line + '\n')
    print(line, flush=True)
with Path(str(prefix) + '-latent.log').open('w') as log:
    child = subprocess.Popen([str(binary), str(tree / 'src/compiler'), str(prefix) + '-latent.jsonl'],
                             env=env, stdout=log, stderr=subprocess.STDOUT)
    meta['pid'] = child.pid
    Path(str(prefix) + '-run.json').write_text(json.dumps(meta, indent=2) + '\n')
    announce(f'foreground head started, pid={child.pid}, LATENT_TIMES enabled')
    stalled = False
    while True:
        try:
            code = child.wait(timeout=30)
            break
        except subprocess.TimeoutExpired:
            timers = Path(str(prefix) + '-times.jsonl')
            rows = [json.loads(s) for s in timers.read_text().splitlines()] if timers.exists() else []
            age = time.time() - timers.stat().st_mtime if timers.exists() else (datetime.datetime.now(datetime.timezone.utc) - datetime.datetime.fromisoformat(meta['start_utc'])).total_seconds()
            where = rows[-1]['where'].replace(str(tree) + '/', '') if rows else 'no completed unit yet'
            announce(f'completed={len(rows)}/4056 last={where} timer_age_seconds={age:.1f}')
            if age > 300:
                stalled = True
                announce('No completed timer for over 300 seconds; capturing SIGQUIT goroutine dump in rerun-head-latent.log before stopping.')
                child.send_signal(signal.SIGQUIT)
                try:
                    code = child.wait(timeout=60)
                except subprocess.TimeoutExpired:
                    child.kill()
                    code = child.wait()
                break
meta.update(end_utc=now(), returncode=code, stalled=stalled)
Path(str(prefix) + '-run.json').write_text(json.dumps(meta, indent=2) + '\n')
Path(str(prefix) + '-end.txt').write_text(meta['end_utc'] + '\n')
Path(str(prefix) + '-exit.txt').write_text(str(code) + '\n')
Path(str(prefix) + '-memory-after.txt').write_text(Path('/sys/fs/cgroup/memory.events').read_text())
announce(f'foreground head exited returncode={code}')
progress.close()
sys.exit(0 if code == 0 else 1)

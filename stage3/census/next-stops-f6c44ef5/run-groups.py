#!/usr/bin/env python3
"""Run the unmodified pinned census, retaining its first per-file row per root."""
import json, os, signal, subprocess, time
from pathlib import Path
BASE = Path(__file__).resolve().parent
TREE = Path('/workspace/scratch/next-stops-adapted')
BINARY = '/workspace/scratch/next-stops-census'
ROOTS = sorted((TREE / 'src/compiler').rglob('*.ts'))
assert len(ROOTS) == 79, len(ROOTS)
BRANCH = 'typescript/next-stops-f6c44ef5-b'

def git(*args):
    with (BASE / 'checkpoint-push.log').open('a') as log:
        subprocess.run(['git', *args], check=True, stdout=log, stderr=log)

def checkpoint(message):
    git('add', str(BASE))
    git('commit', '-m', message)
    git('push', 'origin', 'HEAD:refs/heads/' + BRANCH)

for start in range(0, len(ROOTS), 20):
    group = BASE / ('group-%02d' % (start // 20 + 1))
    group.mkdir(exist_ok=True)
    (group / 'roots.txt').write_text(''.join(str(p) + '\n' for p in ROOTS[start:start+20]))
    with (group / 'raw.jsonl').open('wb') as raw, (group / 'stderr.log').open('ab') as stderr, (group / 'rows.jsonl').open('w') as rows:
        for root in ROOTS[start:start+20]:
            part = group / 'current.jsonl'
            started = time.monotonic()
            proc = subprocess.Popen([BINARY, str(root), str(part)], stdout=stderr, stderr=stderr, start_new_session=True)
            timed_out = False
            try:
                while proc.poll() is None:
                    if part.exists() and b'\n' in part.read_bytes():
                        break
                    if time.monotonic() - started > 1800:
                        timed_out = True
                        break
                    time.sleep(0.1)
            finally:
                if proc.poll() is None:
                    os.killpg(proc.pid, signal.SIGTERM)
                    try:
                        proc.wait(timeout=5)
                    except subprocess.TimeoutExpired:
                        os.killpg(proc.pid, signal.SIGKILL)
                        proc.wait()
            data = part.read_bytes() if part.exists() else b''
            raw.write(data)
            raw.flush()
            if timed_out or not data:
                (group / 'STOPPED.md').write_text(f'Stopped at {root} after {time.monotonic()-started:.3f} seconds without a row. Exit {proc.returncode}.\n')
                checkpoint('Record stopped next-stop census group')
                raise SystemExit(1)
            first = json.loads(data.splitlines()[0])
            assert first['roots'] == [str(root)]
            first['root'] = str(root.relative_to(TREE))
            rows.write(json.dumps(first) + '\n')
            rows.flush()
            part.unlink()
    checkpoint('Record next-stop census group %d' % (start // 20 + 1))

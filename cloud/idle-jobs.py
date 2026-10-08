#!/usr/bin/env python3
"""Mac coordinator. State is durable before ssh, so uncertain launches replay safely."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shlex
import shutil
import subprocess
import tempfile
import time

HERE = Path(__file__).resolve().parent


def ssh(box, script, *args):
    return subprocess.run(['ssh', box, 'bash -s -- ' + shlex.join(args)],
                          input=script, text=True, capture_output=True, timeout=30, check=True).stdout.strip()


def save(path, data):
    temp = path.with_suffix('.tmp')
    temp.write_text(json.dumps(data, indent=2) + '\n')
    os.replace(temp, path)


def collect(box, destination):
    # ssh tar streams only complete, atomically published directories.
    with tempfile.TemporaryDirectory() as tmp:
        archive = Path(tmp) / 'findings.tar'
        with archive.open('wb') as output:
            subprocess.run(['ssh', box, 'test ! -d ~/idle/findings || tar -C ~/idle/findings -cf - .'],
                           stdout=output, stderr=subprocess.PIPE, timeout=30, check=True)
        if not archive.stat().st_size:
            return
        import tarfile
        with tarfile.open(archive) as tar:
            # Only the known metadata files from hash-named directories are accepted.
            for member in tar.getmembers():
                parts = Path(member.name).parts
                if (member.isfile() and len(parts) == 2 and re.fullmatch('[0-9a-f]{64}', parts[0])
                        and parts[1] in {'program.a', 'reduced.a', 'signature.txt', 'seed', 'main', 'box'}):
                    target = Path(tmp) / parts[0] / parts[1]
                    target.parent.mkdir(exist_ok=True)
                    target.write_bytes(tar.extractfile(member).read())
        for folder in Path(tmp).iterdir():
            required = ['program.a', 'reduced.a', 'signature.txt', 'seed', 'main', 'box']
            if not folder.is_dir() or not all((folder / name).is_file() for name in required):
                continue
            signature = (folder / 'signature.txt').read_bytes()
            reduced = (folder / 'reduced.a').read_bytes()
            key = hashlib.sha256(len(signature).to_bytes(8, 'big') + signature + reduced).hexdigest()
            target = destination / key
            if target.exists():
                continue
            destination.mkdir(parents=True, exist_ok=True)
            staging = Path(tempfile.mkdtemp(prefix='.incoming-', dir=destination))
            for name in required:
                shutil.copyfile(folder / name, staging / name)
            os.rename(staging, target)
            print(f"finding {box} {(folder / 'main').read_text().strip()} {signature.decode().strip()} {target}", flush=True)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--dry-run', action='store_true')
    parser.add_argument('--once', action='store_true')
    parser.add_argument('--count', type=int, default=2000)
    args = parser.parse_args()
    if args.count <= 0:
        parser.error('--count must be positive')
    boxes = os.environ.get('ADAMIC_IDLE_BOXES', 'cloud home server workshop chonchon').split()
    if not boxes or any(not re.fullmatch(r'[A-Za-z0-9_-]+', b) for b in boxes):
        parser.error('ADAMIC_IDLE_BOXES must contain ssh alias names')
    state_dir = Path(os.environ.get('ADAMIC_IDLE_STATE', str(Path.home() / '.adamic-idle-jobs')))
    destination = Path(os.environ.get('ADAMIC_IDLE_FINDINGS', str(Path.home() / 'Projects/system/adamic-gate-logs/idle-findings')))
    probe = (HERE / 'idle-job-box.sh').read_text()
    worker = (HERE / 'idle-worker.sh').read_text()
    launch = probe.replace('cat > ~/idle/job.sh', "cat > ~/idle/job.sh <<'IDLE_WORKER'\n" + worker + '\nIDLE_WORKER')
    lock = state_dir / 'coordinator.lock'
    if not args.dry_run:
        state_dir.mkdir(parents=True, exist_ok=True)
        try:
            lock.mkdir()
        except FileExistsError:
            raise SystemExit(f'coordinator already running; remove {lock} only after checking no coordinator is alive')
        (lock / 'pid').write_text(str(os.getpid()))
    try:
        while True:
            pass_started = time.monotonic()
            remote = subprocess.check_output(['git', '-C', str(HERE.parent), 'ls-remote', 'origin', 'refs/heads/main'], text=True).split()
            sha = remote[0]
            if not re.fullmatch('[0-9a-f]{40}', sha):
                raise RuntimeError('invalid main sha')
            path = state_dir / 'state.json'
            state = json.loads(path.read_text()) if path.exists() else {}
            entry = state.setdefault(sha, {'next': 1, 'completed': 0, 'pending': {}})
            for box in boxes:
                try:
                    status = ssh(box, probe, 'probe')
                    if args.dry_run:
                        enabled = (state_dir / 'enabled').exists()
                        print(f'{box}: {status}; main={sha[:12]} programs={entry.get("by_box", {}).get(box, 0)}; would ' +
                              (f'start seed={entry["next"]} count={args.count}' if status == 'idle' and enabled else 'start no job'), flush=True)
                        if status == 'idle' and enabled:
                            entry['next'] += args.count
                        continue
                    if status == 'disabled':
                        print(f'{box}: disabled; main={sha[:12]} programs={entry.get("by_box", {}).get(box, 0)}', flush=True)
                        # Collect existing findings, but preserve pending ranges and launch nothing.
                        collect(box, destination)
                        continue
                    collect(box, destination)
                    pending = entry['pending'].get(box)
                    # Pending ranges for older mains remain eligible until completed.
                    pending_sha = sha
                    for old_sha, old in state.items():
                        if box in old['pending']:
                            pending_sha, pending = old_sha, old['pending'][box]
                            break
                    old = state[pending_sha]
                    if pending:
                        check = '''set -eu
if [ -f "$HOME/idle/out/${1:0:12}/$2/done" ]; then echo done; exit; fi
if [ -e ~/idle/lock ]; then
  exec 8< ~/idle/lock
  flock -n 8 || { echo running; exit; }
fi
echo unfinished
'''
                        result = ssh(box, check, pending_sha, str(pending['seed']))
                        if result == 'done':
                            old['completed'] += pending['count']
                            totals = old.setdefault('by_box', {})
                            totals[box] = totals.get(box, 0) + pending['count']
                            del old['pending'][box]
                            save(path, state)
                            pending = None
                        elif result == 'running':
                            status = 'running a job'
                    if status == 'idle' and (state_dir / 'enabled').exists():
                        if not pending:
                            pending_sha, old = sha, entry
                            pending = {'seed': entry['next'], 'count': args.count}
                            entry['next'] += args.count
                            entry['pending'][box] = pending
                            save(path, state)
                        result = ssh(box, launch, 'start', pending_sha, str(pending['seed']), str(pending['count']), box)
                        status = 'running a job' if result in ('started', 'running') else result
                    print(f'{box}: {status}; main={sha[:12]} programs={entry.get("by_box", {}).get(box, 0)}', flush=True)
                except (OSError, subprocess.SubprocessError) as error:
                    print(f'{box}: unavailable ({error}); main={sha[:12]} programs={entry.get("by_box", {}).get(box, 0)}', flush=True)
            if args.once or args.dry_run:
                return
            time.sleep(max(0, 60 - (time.monotonic() - pass_started)))
    finally:
        if not args.dry_run:
            shutil.rmtree(lock)


if __name__ == '__main__':
    main()

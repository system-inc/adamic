#!/usr/bin/env python3
"""Dispatch green workers to integration's unchanged merge script (macOS and Linux)."""
import argparse
import ast
import csv
import datetime
import fcntl
import hashlib
import io
import json
import os
from pathlib import Path
import re
import sqlite3
import subprocess
import time

HERE = Path(__file__).resolve().parent.parent
STATE = Path(os.environ.get('ADAMIC_FAST_GATE_WATCH_STATE', str(Path.home() / '.adamic-fast-gate-watch'))).resolve()
IDENTITY = ['-c', 'user.name=kirkouimet', '-c', 'user.email=kirk@kirkouimet.com']
TRAILER = 'Co-Authored-By: Ahra <ahra@ahra.ai>'
HEADER = ['utc', 'branch', 'sha', 'area', 'outcome', 'seconds', 'area_sha_after']
RECORD_REF = 'refs/heads/records/auto-area-merges'
RECORD_FILE = 'documentation/velocity/auto-area-merges.csv'


def log(message):
    print(f'{datetime.datetime.now(datetime.timezone.utc):%H:%M:%S} auto-area-merge: {message}', flush=True)


def run(*args, cwd=None, check=True):
    result = subprocess.run(args, cwd=cwd or HERE, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    if check and result.returncode:
        raise RuntimeError(f'{args[0]} failed ({result.returncode}): {result.stdout}')
    return result


def git(*args, cwd=None, check=True):
    return run('git', *IDENTITY, *args, cwd=cwd, check=check).stdout.strip()


def atomic_json(path, value):
    temporary = path.with_suffix('.tmp')
    temporary.write_text(json.dumps(value) + '\n')
    temporary.replace(path)


def checkout():
    """An independent Git checkout, so gate fetches and integration refreshes don't race."""
    root = STATE / 'area-integration'
    if not (root / '.git').exists():
        git('clone', '-q', '--shared', '--no-checkout', str(HERE), str(root))
        git('remote', 'set-url', 'origin', git('remote', 'get-url', 'origin'), cwd=root)
    ready = root / '.git/auto-area-ready'
    if ready.exists() and git('status', '--porcelain', cwd=root):
        raise RuntimeError(f'refused: integration tools checkout has local changes: {root}')
    git('fetch', '-q', 'origin', '+refs/heads/cloud/merge-tree:refs/remotes/origin/cloud/merge-tree', cwd=root)
    git('switch', '-q', '--detach', 'refs/remotes/origin/cloud/merge-tree', cwd=root)
    ready.touch()
    log(f'integration tools {git("rev-parse", "HEAD", cwd=root)} in {root / "cloud/integration"}')
    return root


def table(path):
    return [line.split('\t') for line in path.read_text().splitlines() if line and not line.startswith('#')]


def areas_for(branch, directory):
    """Overrides beat prefixes. Use integration's census rules, never changed-file heuristics."""
    areas = table(directory / 'areas.tsv')
    overrides = [owner for name, owner in table(directory / 'census-owners.tsv') if name == branch]
    if overrides:
        owners = set(overrides)
        candidates = {area for area, owner, _ in areas if owner in owners}
        # Census uses sub-Circle names for the two distinct cohere areas.
        aliases = {'system_cohere_lint': 'stage1-lint', 'system_cohere_format': 'stage1-format'}
        candidates.update(aliases[o] for o in owners if o in aliases)
    else:
        # Read the literal table without importing (or running) the census program.
        module = ast.parse((directory / 'branch-census.py').read_text())
        prefixes = next(ast.literal_eval(node.value) for node in module.body
                        if isinstance(node, ast.Assign) and any(isinstance(t, ast.Name) and t.id == 'prefixOwners' for t in node.targets))
        owners = {owner for prefix, owner in prefixes if branch.startswith(prefix)}
        candidates = {area for area, owner, _ in areas if owner in owners}
        aliases = {'system_cohere_lint': 'stage1-lint', 'system_cohere_format': 'stage1-format'}
        candidates.update(aliases[o] for o in owners if o in aliases)
        # Area-spelled worker prefixes (codex/runtime-*, codex/compiler-*, etc.) are explicit.
        candidates.update(area for area, _, _ in areas if branch.startswith(f'codex/{area}-'))
    valid = {area for area, _, _ in areas}
    return sorted(candidates & valid)


def csv_line(row):
    output = io.StringIO()
    csv.writer(output, lineterminator='\n').writerow(row)
    return output.getvalue()


def record(row, dry):
    log('csv: ' + csv_line(row).strip())
    if dry:
        path = STATE / 'dry-run' / RECORD_FILE
        path.parent.mkdir(parents=True, exist_ok=True)
        if not path.exists():
            path.write_text(csv_line(HEADER))
        with path.open('a') as output:
            output.write(csv_line(row))
        return
    # Save first: a failed records push must never discard the result or rerun a finished merge.
    outbox = STATE / 'area-records-outbox'
    outbox.mkdir(exist_ok=True)
    key = hashlib.sha256(csv_line(row).encode()).hexdigest()
    atomic_json(outbox / (key + '.json'), row)


def publish_records():
    outbox = STATE / 'area-records-outbox'
    if not outbox.exists():
        return
    records = STATE / 'area-records'
    for pending in sorted(outbox.glob('*.json')):
        if not enabled():
            return
        row = json.loads(pending.read_text())
        remote = git('ls-remote', 'origin', RECORD_REF)
        if remote:
            tracking = 'refs/remotes/origin/records/auto-area-merges'
            git('fetch', '-q', 'origin', f'+{RECORD_REF}:{tracking}')
            base = git('rev-parse', tracking)
        else:
            # hash-object needs empty input rather than the watcher's stdin.
            empty = subprocess.check_output(['git', 'hash-object', '-w', '-t', 'tree', '--stdin'], cwd=HERE, input=b'').decode().strip()
            base = git('commit-tree', empty, '-m', 'Start automatic area merge records', '-m', TRAILER)
        if not records.exists():
            git('worktree', 'add', '-q', '--detach', str(records), base)
        else:
            git('switch', '-q', '--detach', base, cwd=records)
        path = records / RECORD_FILE
        path.parent.mkdir(parents=True, exist_ok=True)
        if not path.exists():
            path.write_text(csv_line(HEADER))
        with path.open(newline='') as source:
            already = row in list(csv.reader(source))
        if not already:
            with path.open('a') as output:
                output.write(csv_line(row))
            git('add', RECORD_FILE, cwd=records)
            git('commit', '-q', '-m', f'Area merge: {row[1]} {row[4]}', '-m', TRAILER, cwd=records)
            result = run('git', 'push', '-q', 'origin', f'HEAD:{RECORD_REF}', cwd=records, check=False)
            if result.returncode:
                log(f'records push failed; row retained in {pending}: {result.stdout.strip()}')
                return
        pending.unlink()


def notify(branch, sha, area, outcome, output, logfile):
    database = Path(os.environ.get('ADAMIC_AI_DATABASE', '/Users/kirkouimet/Projects/ahra/modules/ai/data/ai.db'))
    session = None
    if database.exists():
        try:
            with sqlite3.connect(database.resolve().as_uri() + '?mode=ro', uri=True) as connection:
                session = connection.execute('select session_id from replies where instr(text, ?) > 0 order by fetched_at desc limit 1', (branch,)).fetchone()
        except sqlite3.Error as error:
            log(f'ai.db lookup failed for {branch}: {error}')
    if not session:
        log(f'no session for {branch}; {outcome} details in {logfile}')
        return
    message = logfile.with_suffix('.message.txt')
    text = f'Automatic area merge {outcome}: {branch} {sha} into area/{area}.\nResolve this in your branch and push a new tip for the fast gate.\n\n{output}\n'
    # The script prints conflict names and new test failures; vet/census/lane errors name log files.
    for name in sorted(set(re.findall(r'(/[^\s();,]+\.(?:log|txt))', output))):
        path = Path(name)
        if path.is_file():
            text += f'\n--- {path} ---\n' + path.read_text(errors='replace')[-100000:]
    message.write_text(text)
    result = run('ahra', 'ai', 'send', str(session[0]), '--message-file', str(message),
                 cwd='/Users/kirkouimet/Projects/ahra', check=False)
    log(f'worker {session[0]} notification: {result.returncode} {result.stdout.strip()}')


def enabled():
    return (STATE / 'auto-area-merge').is_file()


def attempt(root, branch, sha, dry):
    start = time.monotonic()
    utc = datetime.datetime.now(datetime.timezone.utc).strftime('%Y-%m-%dT%H:%M:%SZ')
    candidates = areas_for(branch, root / 'cloud/integration')
    area = candidates[0] if len(candidates) == 1 else ''
    output = ''
    if not area:
        log(f'no area for {branch}' + (f' (ambiguous: {", ".join(candidates)})' if candidates else ''))
        outcome = 'no-area'
    else:
        log(f'{branch} {sha} -> area/{area}' + (' (--no-push)' if dry else ''))
        arguments = ['bash', str(root / 'cloud/integration/area-merge.sh'), area, branch, sha]
        if dry:
            arguments.append('--no-push')
        # area-merge owns all local/origin locks, main catch-up, tests, merges and area pushes.
        result = run(*arguments, cwd=root, check=False)
        output = result.stdout
        print(output, end='', flush=True)
        if 'refused: another merge into area/' in output or 'refused: another machine is merging into area/' in output:
            outcome = 'locked'
        elif result.returncode == 3:
            outcome = 'conflict'
        elif result.returncode == 0:
            outcome = 'merged'
        else:
            outcome = 'red'
    area_sha = ''
    if area:
        try:
            area_sha = git('ls-remote', 'origin', f'refs/heads/area/{area}', cwd=root).split()[0]
        except (RuntimeError, IndexError) as error:
            log(f'could not read area sha after: {error}')
    row = [utc, branch, sha, area, outcome, str(round(time.monotonic() - start, 3)), area_sha]
    record(row, dry)
    logs = STATE / ('dry-run' if dry else 'area-logs')
    logs.mkdir(exist_ok=True)
    logfile = logs / (hashlib.sha256((branch + sha + utc).encode()).hexdigest() + '.log')
    logfile.write_text(output)
    if outcome in ('red', 'conflict') and not dry:
        try:
            notify(branch, sha, area, outcome, output, logfile)
        except (OSError, RuntimeError) as error:
            log(f'worker notification failed for {branch}: {error}; details in {logfile}')
    return outcome


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument('--enqueue', nargs=2, metavar=('BRANCH', 'SHA'))
    group.add_argument('--drain', action='store_true')
    group.add_argument('--dry-run', nargs=2, metavar=('BRANCH', 'SHA'))
    args = parser.parse_args()
    STATE.mkdir(parents=True, exist_ok=True)
    if not args.dry_run and not enabled():
        return
    queue = STATE / 'area-queue'
    queue.mkdir(exist_ok=True)
    if args.enqueue:
        branch, sha = args.enqueue
        if not branch.startswith('codex/') or not re.fullmatch('[0-9a-f]{40}', sha):
            parser.error('want codex/* and a full sha')
        key = hashlib.sha256((branch + '\n' + sha).encode()).hexdigest()
        path = queue / (key + '.json')
        if not path.exists() and not (queue / (key + '.done')).exists():
            atomic_json(path, [branch, sha])
            log(f'queued {branch} {sha}')
        return
    # flock is provided by Python on macOS; this dispatcher lock has no stale pid-file lifetime.
    with (STATE / 'area-dispatch.lock').open('a') as lock:
        try:
            fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError:
            return
        if args.dry_run:
            branch, sha = args.dry_run
            if not branch.startswith('codex/') or not re.fullmatch('[0-9a-f]{40}', sha):
                parser.error('want codex/* and a full sha')
            attempt(checkout(), branch, sha, True)
            return
        publish_records()
        if not enabled() or not list(queue.glob('*.json')):
            return
        root = checkout()
        for path in sorted(queue.glob('*.json'), key=lambda p: p.stat().st_mtime):
            if not enabled():
                break
            branch, sha = json.loads(path.read_text())
            outcome = attempt(root, branch, sha, False)
            if outcome != 'locked':
                path.rename(path.with_suffix('.done'))
            # A busy area remains queued. Other queued areas can make progress this poll.
            publish_records()


if __name__ == '__main__':
    try:
        main()
    except (OSError, RuntimeError, ValueError, StopIteration) as error:
        log(str(error))
        raise SystemExit(1)

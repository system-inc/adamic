#!/usr/bin/env python3
"""Dispatch green workers to integration's unchanged merge script (macOS and Linux)."""
import argparse
import csv
import datetime
import fcntl
import importlib.util
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


def route(branch, directory):
    """Use only integration's trusted route API; malformed/unavailable routing is held."""
    try:
        spec = importlib.util.spec_from_file_location('trusted_area_route', directory / 'area-route.py')
        module = importlib.util.module_from_spec(spec)
        # Load trusted source without writing __pycache__ into integration's checkout.
        source = directory / 'area-route.py'
        exec(compile(source.read_text(), str(source), 'exec'), module.__dict__)
        area, how = module.route(branch)
        if not isinstance(area, str) or not isinstance(how, str) or not how.strip():
            return 'hold', 'integration route returned an invalid area/reason; ask system_adamic_integration'
        if area != 'hold' and area not in module.areaNames():
            return 'hold', f'integration route returned unknown area {area!r}; ask system_adamic_integration'
        return area, how
    except Exception as error:
        return 'hold', f'integration routing unavailable: {error}; ask system_adamic_integration'


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


def ai_database():
    return Path(os.environ.get('ADAMIC_AI_DATABASE', '/Users/kirkouimet/Projects/ahra/modules/ai/data/ai.db'))


def worker_session(connection, branch):
    # A reply naming codex/foo-bar must not identify the worker of codex/foo.
    pattern = re.compile(r'(?<![A-Za-z0-9_./-])' + re.escape(branch) + r'(?![A-Za-z0-9_./-])')
    for session, text in connection.execute(
            'select session_id, text from replies where instr(text, ?) > 0 order by fetched_at desc', (branch,)):
        if session and pattern.search(text or ''):
            return str(session)
    return None


def hold_job(branch, sha, reason, dry=False):
    if dry:
        log(f'held {branch} {sha}: {reason} (dry run, not reported)')
        return
    directory = STATE / 'area-held'
    directory.mkdir(exist_ok=True)
    key = hashlib.sha256(branch.encode()).hexdigest()
    path = directory / (key + '.json')
    previous = json.loads(path.read_text()) if path.exists() else {}
    reported = previous.get('reported_to', ['system_adamic_integration'] if previous.get('reported') else [])
    entry = {'branch': branch, 'sha': sha, 'reason': reason, 'reported_to': reported}
    atomic_json(path, entry)
    write_held_list()
    circles = set(re.findall(r'@([A-Za-z0-9_]+)', reason))
    circles.update(re.findall(r'\b(system_[A-Za-z0-9_]+)\b', reason))
    targets = ['system_adamic_integration'] + sorted(circles - {'system_adamic_integration'})
    message = f'Automatic area merge held: {branch} {sha}: {reason}. Job remains .held; integration must decide or clear the refusal.'
    for target in targets:
        if target in entry['reported_to']:
            continue
        try:
            result = run('ahra', 'os', 'send', target, message,
                         cwd='/Users/kirkouimet/Projects/ahra', check=False)
            if result.returncode:
                log(f'hold report to {target} failed for {branch}: {result.stdout.strip()}')
                continue
        except OSError as error:
            log(f'hold report to {target} failed for {branch}: {error}')
            continue
        entry['reported_to'].append(target)
        atomic_json(path, entry)
        log(f'reported held branch {branch} to {target}')


def write_held_list():
    output = io.StringIO()
    writer = csv.writer(output, delimiter='\t', lineterminator='\n')
    writer.writerow(['branch', 'sha', 'reason'])
    for file in sorted((STATE / 'area-held').glob('*.json')):
        item = json.loads(file.read_text())
        if not item.get('resolved'):
            writer.writerow([item['branch'], item['sha'], item['reason']])
    temporary = STATE / 'area-held.tsv.tmp'
    temporary.write_text(output.getvalue())
    temporary.replace(STATE / 'area-held.tsv')


def clear_hold(branch):
    # Keep the per-branch reported marker: future SHAs must not resend the same alert.
    path = STATE / 'area-held' / (hashlib.sha256(branch.encode()).hexdigest() + '.json')
    if path.exists():
        entry = json.loads(path.read_text())
        entry['resolved'] = True
        atomic_json(path, entry)
        write_held_list()


def notify(branch, sha, area, outcome, output, logfile):
    database = ai_database()
    session = None
    if database.is_file():
        try:
            with sqlite3.connect(database.resolve().as_uri() + '?mode=ro', uri=True) as connection:
                session = worker_session(connection, branch)
        except sqlite3.Error as error:
            log(f'ai.db lookup failed for {branch}: {error}')
    if not session:
        log(f'no session for {branch}; {outcome} details in {logfile}')
        return
    message = logfile.with_suffix('.message.txt')
    if outcome == 'merged':
        text = f'Merged {branch} {sha} into area/{area} automatically.\n'
    else:
        text = f'Automatic area merge {outcome}: {branch} {sha} into area/{area}.\nResolve this in your branch and push a new tip for the fast gate.\n\n{output}\n'
    # The script prints conflict names and new test failures; vet/census/lane errors name log files.
    for name in sorted(set(re.findall(r'(/[^\s();,]+\.(?:log|txt))', output))) if outcome != 'merged' else []:
        path = Path(name)
        if path.is_file():
            text += f'\n--- {path} ---\n' + path.read_text(errors='replace')[-100000:]
    message.write_text(text)
    result = run('ahra', 'ai', 'send', session, '--message-file', str(message),
                 cwd='/Users/kirkouimet/Projects/ahra', check=False)
    log(f'worker {session} notification: {result.returncode} {result.stdout.strip()}')


def enabled():
    return (STATE / 'auto-area-merge').is_file()


def attempt(root, branch, sha, dry, routing=None):
    start = time.monotonic()
    utc = datetime.datetime.now(datetime.timezone.utc).strftime('%Y-%m-%dT%H:%M:%SZ')
    area, how = routing or route(branch, root / 'cloud/integration')
    output = ''
    if area == 'hold':
        area = ''
        log(f'no area for {branch}: {how}')
        hold_job(branch, sha, how, dry)
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
        refusals = re.findall(r'^refused:.*$', output, re.MULTILINE)
        if refusals:
            reason = '; '.join(refusals)
            # Live lock contention must retry, but all refusal details go to integration.
            busy = any('another merge into area/' in line or 'another machine is merging into area/' in line for line in refusals)
            outcome = 'locked' if busy else 'held'
            hold_job(branch, sha, reason, dry)
        elif result.returncode == 3:
            outcome = 'conflict'
        elif result.returncode == 0:
            outcome = 'merged'
        elif result.returncode == 1 and re.search(r'^(not merged:|new failures from merging |go vet failed:|gofmt:)', output, re.MULTILINE):
            outcome = 'red'
        else:
            outcome = 'held'
            hold_job(branch, sha, f'area-merge exited {result.returncode} without a recognized verdict: {output.strip()[:2000]}', dry)
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
    if outcome in ('merged', 'red', 'conflict') and not dry:
        clear_hold(branch)
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
        if not path.exists() and not (queue / (key + '.done')).exists() and not (queue / (key + '.held')).exists():
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
        pending = list(queue.glob('*.json')) + list(queue.glob('*.held'))
        if not enabled() or not pending:
            return
        root = checkout()
        for path in sorted(pending, key=lambda p: p.stat().st_mtime):
            if not enabled():
                break
            job = json.loads(path.read_text())
            branch, sha = (job['branch'], job['sha']) if isinstance(job, dict) else job
            routing = route(branch, root / 'cloud/integration')
            if path.suffix == '.held' and isinstance(job, dict) and job.get('routing') == list(routing) and job.get('outcome') != 'locked':
                # A route decision or refusal needs a changed route (or manual requeue), not another attempt each poll.
                hold_job(branch, sha, job['reason'])
                continue
            if path.suffix == '.held':
                queued = path.with_suffix('.json')
                path.rename(queued)
                path = queued
                log(f'requeued held branch {branch}: routing changed or area lock retry')
            outcome = attempt(root, branch, sha, False, routing=routing)
            if outcome in ('merged', 'red', 'conflict'):
                path.rename(path.with_suffix('.done'))
            else:
                held_entry = STATE / 'area-held' / (hashlib.sha256(branch.encode()).hexdigest() + '.json')
                reason = json.loads(held_entry.read_text())['reason'] if held_entry.exists() else f'unrecognized outcome {outcome!r}; ask system_adamic_integration'
                if not held_entry.exists():
                    hold_job(branch, sha, reason)
                atomic_json(path, {'branch': branch, 'sha': sha, 'status': 'held', 'reason': reason,
                                   'routing': list(routing), 'outcome': outcome})
                path.rename(path.with_suffix('.held'))
            publish_records()


if __name__ == '__main__':
    try:
        main()
    except (OSError, RuntimeError, ValueError, StopIteration) as error:
        log(str(error))
        raise SystemExit(1)

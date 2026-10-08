#!/usr/bin/env python3
"""Keep Codex lanes at their floor without a window coordinator.

Settings: ADAMIC_FLEET_FLOOR_STATE, ADAMIC_FLEET_FLOOR_CHECKOUT and
ADAMIC_FLEET_FLOOR_CREDIT_FLOOR (defaults below). Queue files are prompts,
ordered by filename; an optional first line is Label: <label>. A trailing *
fleet selector is passed literally to ahra and matched as a prefix locally.
Failed/uncertain dispatch attempts keep their claimed brief in used: an operator
must inspect the session before supplying another unit. No automatic replay.
Rows spool locally before Git; a failed records push is retried next pass.
"""
import argparse
import csv
import datetime
import fcntl
import json
import math
import os
from pathlib import Path
import re
import subprocess
import time
import uuid

HERE = Path(__file__).resolve().parent.parent
FIELDS = ['time', 'lane', 'running', 'floor', 'queued', 'brief', 'session', 'outcome']
BRANCH = 'records/fleet-floor'
CSV = 'documentation/velocity/fleet-floor.csv'


def run(*args, input=None):
    return subprocess.run(list(map(str, args)), input=input, text=True,
                          capture_output=True, timeout=60, check=True).stdout.strip()


def save(path, value):
    path.parent.mkdir(parents=True, exist_ok=True)
    temp = path.with_name(path.name + '.tmp')
    with temp.open('w') as file:
        json.dump(value, file)
        file.write('\n')
        file.flush()
        os.fsync(file.fileno())
    os.replace(temp, path)
    # Persist renames as well as file contents across a power loss.
    fd = os.open(str(path.parent), os.O_RDONLY)
    try:
        os.fsync(fd)
    finally:
        os.close(fd)


def load(path, default):
    return json.loads(path.read_text()) if path.exists() else default


class Records:
    def __init__(self, state, repo=HERE, dry=False):
        self.state, self.repo, self.dry = state, repo, dry
        self.pending = state / 'rows'
        self.worktree = state / 'records'
        self.transaction = state / 'record-transaction.json'

    def row(self, lane, running, queued, brief='', session='', outcome=''):
        row = [datetime.datetime.now(datetime.timezone.utc).isoformat(timespec='microseconds').replace('+00:00', 'Z'),
               lane['lane'], running, lane['floor'], queued, brief, session, outcome]
        print(','.join(map(str, row)), flush=True)
        if not self.dry:
            save(self.pending / (uuid.uuid4().hex + '.json'), row)
        return row

    def git(self, *args, input=None):
        return run('git', '-C', self.repo, *args, input=input)

    def wg(self, *args):
        return run('git', '-C', self.worktree, *args)

    def acknowledge(self, transaction):
        for name in transaction['files']:
            (self.pending / name).unlink(missing_ok=True)
        self.transaction.unlink(missing_ok=True)

    def flush(self):
        if self.dry or not self.pending.exists():
            return
        try:
            if not list(self.pending.glob('*.json')):
                return
            remote = self.git('ls-remote', 'origin', 'refs/heads/' + BRANCH).split()
            tip = remote[0] if remote else None
            if tip:
                self.git('fetch', '-q', 'origin', 'refs/heads/' + BRANCH)
            if not self.worktree.exists():
                base = tip
                if not base:
                    tree = self.git('hash-object', '-t', 'tree', '--stdin', input='')
                    base = self.git('-c', 'user.name=kirkouimet', '-c',
                                    'user.email=kirk@kirkouimet.com', 'commit-tree', tree,
                                    '-m', 'Start the fleet floor record', '-m',
                                    'Co-Authored-By: Ahra <ahra@ahra.ai>')
                self.git('worktree', 'add', '-q', '--detach', self.worktree, base)
            transaction = load(self.transaction, None)
            if transaction and tip:
                # A successful push followed by a crash is acknowledged, not appended twice.
                check = subprocess.run(['git', '-C', str(self.worktree), 'merge-base',
                                        '--is-ancestor', transaction['commit'], tip],
                                       capture_output=True, timeout=60)
                if check.returncode == 0:
                    self.acknowledge(transaction)
                    transaction = None
                elif check.returncode != 1:
                    raise RuntimeError('cannot verify records ancestry')
            if transaction and transaction['base'] == tip:
                self.wg('push', '-q', 'origin', transaction['commit'] + ':refs/heads/' + BRANCH)
                self.acknowledge(transaction)
                tip = transaction['commit']
                transaction = None
            files = sorted(self.pending.glob('*.json'), key=lambda p: (load(p, [])[0], p.name))
            if not files:
                return
            if tip:
                self.wg('reset', '--hard', tip)
            else:
                # A never-pushed local commit must not become a duplicate base on retry.
                base = self.wg('rev-list', '--max-parents=0', 'HEAD').splitlines()[0]
                self.wg('reset', '--hard', base)
            target = self.worktree / CSV
            target.parent.mkdir(parents=True, exist_ok=True)
            exists = target.exists()
            with target.open('a', newline='') as file:
                writer = csv.writer(file)
                if not exists:
                    writer.writerow(FIELDS)
                writer.writerows(load(path, []) for path in files)
            self.wg('add', CSV)
            self.wg('-c', 'user.name=kirkouimet', '-c', 'user.email=kirk@kirkouimet.com',
                    'commit', '-q', '-m', 'Fleet floor: record pass', '-m',
                    'Co-Authored-By: Ahra <ahra@ahra.ai>')
            transaction = {'base': tip, 'commit': self.wg('rev-parse', 'HEAD'),
                           'files': [path.name for path in files]}
            save(self.transaction, transaction)
            self.wg('push', '-q', 'origin', transaction['commit'] + ':refs/heads/' + BRANCH)
            self.acknowledge(transaction)
        except (OSError, ValueError, RuntimeError, subprocess.SubprocessError) as error:
            print('records retained locally: ' + str(error), flush=True)


def lanes(path):
    with path.open(newline='') as file:
        result = list(csv.DictReader(file, delimiter='\t'))
    for lane in result:
        lane['floor'] = int(lane['floor'])
        lane['fleets'] = lane['fleets'].split(',')
        if (lane['floor'] < 1 or not re.fullmatch(r'[a-z0-9_]+', lane['lane'])
                or any(not re.fullmatch(r'[a-z0-9_-]+\*?', f) for f in lane['fleets'])
                or lane['fleets'][0].endswith('*')):
            raise ValueError('invalid lane: ' + str(lane))
    if len({lane['lane'] for lane in result}) != len(result):
        raise ValueError('duplicate lanes')
    return result


def members(lane):
    found, first = {}, []
    for selector in lane['fleets']:
        values = json.loads(run('ahra', 'ai', 'fleet', selector, '--json', '--live'))
        if not isinstance(values, list):
            raise ValueError('fleet response must be a list')
        for member in values:
            fleet = member.get('fleet', selector)
            matches = (fleet.startswith(selector[:-1]) if selector.endswith('*') else fleet == selector)
            if not matches or member.get('provider', 'Codex').lower() != 'codex':
                continue
            session = member['session']
            if member.get('pollError') or member.get('statusStale'):
                raise ValueError('fleet live status unavailable')
            if session['status'] not in {'Running', 'Idle', 'Completed', 'Failed', 'Cancelled'}:
                raise ValueError('unknown session status')
            key = session['sessionId']
            found[key] = member
            if selector == lane['fleets'][0] and key not in {m['session']['sessionId'] for m in first}:
                first.append(member)
    return sum(m['session']['status'] == 'Running' for m in found.values()), first


def credit_block(floor):
    usage = json.loads(run('ahra', 'ai', 'usage', '--provider', 'codex', '--json'))
    threshold = number(usage['rules']['creditThresholdPercent'])
    accounts = [a for a in usage['accounts'] if a['provider'].lower() == 'codex']
    if not accounts:
        raise ValueError('no codex usage accounts')
    for account in accounts:
        if account.get('error'):
            raise ValueError('usage unavailable')
        meters = account['meters']
        if not any(m['unit'] == 'Credits' for m in meters):
            raise ValueError('credits meter unavailable')
        for meter in meters:
            if meter['unit'] == 'Credits' and number(meter['remaining']) < floor:
                return 'credit_floor'
            if meter['unit'] == 'Percent' and number(meter['remaining']) < threshold:
                return 'percent_floor'
    return ''


def number(value):
    if isinstance(value, bool) or value is None:
        raise ValueError('missing numeric usage')
    result = float(value)
    if not math.isfinite(result):
        raise ValueError('invalid numeric usage')
    return result


def queue(state, lane):
    path = state / 'queues' / lane['lane']
    return sorted((p for p in path.iterdir() if p.is_file() and not p.is_symlink()),
                  key=lambda p: p.name) if path.exists() else []


def notify(state, lane, running, count, dry, records):
    path = state / 'notifications.json'
    all_flags = load(path, {})
    flags = all_flags.setdefault(lane['lane'], {})
    if count > lane['floor'] / 2:
        flags['low'] = False
    if count > 0:
        flags['empty'] = False
    for key, condition, text in [
            ('low', count < lane['floor'] / 2,
             f"Fleet floor {lane['lane']}: queue has {count} briefs, below half its floor of {lane['floor']}."),
            ('empty', count == 0 and running < lane['floor'],
             f"Fleet floor {lane['lane']}: queue is empty (0 briefs), running {running} of {lane['floor']}.")]:
        if condition and not flags.get(key):
            text = re.sub(r'[A-Z]{3,}', lambda m: m.group().lower(), text)
            if dry:
                print('would message ' + lane['owner'] + ': ' + text, flush=True)
            else:
                try:
                    run('ahra', 'os', 'send', lane['owner'], text,
                        '--from', 'system_adamic_developer_tools')
                except (OSError, subprocess.SubprocessError) as error:
                    records.row(lane, running, count, outcome='notification_failed: ' + str(error))
                    continue
            flags[key] = True
    if not dry:
        save(path, all_flags)


def one_pass(state, configured, checkout, credit_floor, dry=False, records=None):
    records = records or Records(state, dry=dry)
    # Intents survive a killed process: the brief is never offered again automatically.
    intents = state / 'inflight'
    if not dry and intents.exists():
        for path in sorted(intents.glob('*.json')):
            entry = load(path, {})
            brief = Path(entry['brief'])
            used = Path(entry['used'])
            if brief.exists() and not used.exists():
                used.parent.mkdir(parents=True, exist_ok=True)
                os.rename(brief, used)
            records.row(entry['lane'], entry['running'], entry['queued'],
                        brief.name, entry['session'], 'dispatch_uncertain')
            path.unlink()
    try:
        blocked = credit_block(credit_floor)
    except (OSError, ValueError, KeyError, TypeError, subprocess.SubprocessError) as error:
        blocked = 'usage_unavailable: ' + str(error)
    for lane in configured:
        briefs = queue(state, lane)
        try:
            running, first = members(lane)
        except (OSError, ValueError, KeyError, TypeError, subprocess.SubprocessError) as error:
            records.row(lane, '', len(briefs), outcome='fleet_unavailable: ' + str(error))
            continue
        notify(state, lane, running, len(briefs), dry, records)
        reason = ('off' if (state / 'off').exists() else blocked if blocked else
                  'floor_met' if running >= lane['floor'] else 'queue_empty' if not briefs else '')
        if reason:
            records.row(lane, running, len(briefs), outcome=reason)
            continue
        available = [m for m in first if m['session']['status'] in {'Idle', 'Completed'}]
        dispatched = 0
        for brief in briefs:
            if running >= lane['floor'] or dispatched >= 3:
                break
            if (state / 'off').exists():
                records.row(lane, running, len(queue(state, lane)), outcome='off')
                break
            count = len(briefs) - dispatched if dry else len(queue(state, lane))
            used = state / 'used' / lane['lane'] / brief.name
            if used.exists():
                records.row(lane, running, count, brief.name, outcome='already_used')
                continue
            try:
                with brief.open() as file:
                    header = file.readline().rstrip('\r\n')
                label = header[len('Label: '):].strip() if header.startswith('Label: ') else brief.name
                label = label or brief.name
                member = next((m for m in available if m.get('label') == label),
                              available[0] if available else None)
                session = member['session']['sessionId'] if member else ''
                command = (['ahra', 'ai', 'send', session, '--message-file', str(brief)] if member else
                           ['ahra', 'ai', 'start', 'codex', '--directory', str(checkout), '--label', label,
                            '--fleet', lane['fleets'][0], '--prompt-file', str(brief)])
                if dry:
                    print('would dispatch ' + json.dumps(command), flush=True)
                    records.row(lane, running, count, brief.name, session, 'dry_run')
                else:
                    intent = intents / (uuid.uuid4().hex + '.json')
                    save(intent, {'lane': lane, 'running': running, 'queued': count,
                                  'brief': str(brief), 'used': str(used), 'session': session})
                    used.parent.mkdir(parents=True, exist_ok=True)
                    os.rename(brief, used)
                    # The retained copy supplies the identical prompt bytes after claiming it.
                    command[-1] = str(used)
                    try:
                        output = run(*command)
                        if not member:
                            match = re.search(r'^([0-9a-fA-F]{8}-[^\s]+)', output, re.MULTILINE)
                            if not match:
                                raise ValueError('start returned no session id')
                            session = match.group(1)
                        records.row(lane, running, count, brief.name, session,
                                    'sent' if member else 'started')
                    except (OSError, ValueError, subprocess.SubprocessError) as error:
                        records.row(lane, running, count, brief.name, session,
                                    'dispatch_uncertain: ' + str(error))
                        # Stop this lane: do not amplify a possibly failing provider.
                        intent.unlink()
                        break
                    intent.unlink()
                if member:
                    available.remove(member)
                dispatched += 1
                running += 1
            except OSError as error:
                records.row(lane, running, count, brief.name, outcome='brief_unavailable: ' + str(error))
                break
        remaining = len(briefs) - dispatched if dry else len(queue(state, lane))
        notify(state, lane, running, remaining, dry, records)
        if running < lane['floor']:
            records.row(lane, running, remaining,
                        outcome='queue_empty' if not remaining else 'pass_cap' if dispatched == 3 else 'dispatch_stopped')
    records.flush()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--once', action='store_true')
    parser.add_argument('--dry-run', action='store_true')
    args = parser.parse_args()
    state = Path(os.environ.get('ADAMIC_FLEET_FLOOR_STATE', str(Path.home() / '.adamic-fleet-floor'))).expanduser().resolve()
    checkout = Path(os.environ.get('ADAMIC_FLEET_FLOOR_CHECKOUT', '/Users/kirkouimet/Projects/system/adamic'))
    credit_floor = number(os.environ.get('ADAMIC_FLEET_FLOOR_CREDIT_FLOOR', '5000'))
    if credit_floor < 0:
        parser.error('credit floor must be nonnegative')
    configured = lanes(HERE / 'cloud/fleet-floor/lanes.tsv')
    if args.dry_run:
        one_pass(state, configured, checkout, credit_floor, dry=True)
        return
    state.mkdir(parents=True, exist_ok=True)
    with (state / 'loop.lock').open('a') as lock:
        try:
            fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError:
            raise SystemExit('fleet floor already running')
        while True:
            started = time.monotonic()
            one_pass(state, configured, checkout, credit_floor)
            if args.once:
                return
            time.sleep(max(0, 300 - (time.monotonic() - started)))


if __name__ == '__main__':
    main()

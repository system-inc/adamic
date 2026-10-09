#!/usr/bin/env python3
"""The 90-second probe's pure parts (#c89weq5, for Outcome #awn479j): its branch name, the skip-globs check, and one
published fast-gate record read as five phases and a status line. cloud/probe-90.sh does the Git and the waiting and
calls this; cloud/probe-90-test.py tests it offline.

    record.py branch <UTC stamp> <leaf|all>
    record.py skipped <branch> <skip-globs file>          prints the glob that skips it, exit 1 when none does
    record.py line <record dir> --mode M --sha S --ref R --stamp-epoch N --commit-epoch N --pushed-epoch N --tools T

The phases, from whichever route answered:

  Loom's pool (fast.json runner "pool"; the normal route): no steps_seconds, so they come from record.jsonl, the run's
  unit events, and the record's two clocks: the ref's stamp (fast.sh's serve start) and the record commit's time.
    select   serve start to the first unit started: the gate's selection run, the job's planning, Loom placing units
             (unknown, with publish, when the record carries an earlier attempt's run, its units started before the stamp)
    fetch    the longest unit setup, "setup N s" (the tree and its build cache onto an instance)
    build    the build-vet unit's exit wallSeconds less its setup (go build ./... and go vet ./...)
    test     the longest test unit, "tests took N s" (else its exit wallSeconds less its setup)
    publish  the last unit finished to the record commit
  A box (fast.json steps_seconds, written by run.py; a race's other route). Its steps run at once, each its seconds
  from run.py's start, and the box's own fetch happens before run.py:
    select   run.py's wall less its longest step: the time before its stages start, mostly selection
    fetch    the record commit less the stamp less run.py's wall: ssh, the tools and tree fetch, copying back
    build    steps_seconds.build
    test     steps_seconds.tests
    publish  not separable on a box (inside fetch), shown as "-"
  wall is the stamp to the record commit on both; queue is the push to the stamp (the watcher noticing, the pool job
  waiting for a slot). Phases overlap (tests run beside the build), so they don't sum to the wall.
"""
import argparse
import fnmatch
import gzip
import json
import os
import re
import sys

MaximumLine = 160


def branchName(stamp, mode):
    """devtools/probe-90-<stamp>, and -all for the forced-wide mutant: a devtools/* tip the watcher gates as a
    candidate, against main, and never the tools branch devtools/probe-90 itself."""
    if not re.fullmatch(r'\d{8}T\d{6}Z', stamp):
        raise ValueError('stamp must be a UTC stamp like 20261009T170000Z: %r' % stamp)
    if mode not in ('leaf', 'all'):
        raise ValueError('mode must be leaf or all: %r' % mode)
    return 'devtools/probe-90-%s%s' % (stamp, '-all' if mode == 'all' else '')


def skippedBy(branch, text):
    """The skip-globs line's glob that takes the branch out, as the watcher reads it (bash [[ == ]], where * crosses
    slashes, as fnmatchcase's does), or None."""
    for line in text.splitlines():
        fields = line.split()
        if not fields or fields[0].startswith('#'):
            continue
        if fnmatch.fnmatchcase(branch, fields[0]):
            return fields[0]
    return None


def epoch(text):
    """An event's ISO time (2026-10-09T15:49:44.948Z) as epoch seconds."""
    from datetime import datetime, timezone
    text = text.rstrip('Z')
    whole, _, fraction = text.partition('.')
    moment = datetime.strptime(whole, '%Y-%m-%dT%H:%M:%S').replace(tzinfo=timezone.utc).timestamp()
    return moment + (float('0.' + fraction) if fraction else 0.0)


def readEvents(directory):
    for name in ('record.jsonl', 'record.jsonl.gz'):
        path = os.path.join(directory, name)
        if os.path.exists(path):
            opener = gzip.open if name.endswith('.gz') else open
            with opener(path, 'rt') as handle:
                return [json.loads(line) for line in handle if line.strip()]
    return []


def poolPhases(events, stampEpoch, commitEpoch):
    units = {}
    for event in events:
        unit = event.get('unit')
        if not unit:
            continue
        row = units.setdefault(unit, {})
        kind = event.get('type')
        if kind == 'started':
            row['started'] = min(row.get('started', float('inf')), epoch(event['time']))
        elif kind == 'finished':
            row['finished'] = max(row.get('finished', 0.0), epoch(event['time']))
        elif kind == 'exit' and 'wallSeconds' in event:
            row['wall'] = float(event['wallSeconds'])
        elif kind == 'output':
            text = event.get('text') or ''
            setup = re.search(r'\bsetup (\d+(?:\.\d+)?) s\b', text)
            if setup:
                row['setup'] = float(setup.group(1))
            took = re.search(r'\btests took (\d+(?:\.\d+)?) s\b', text)
            if took:
                row['tests'] = float(took.group(1))
    started = [row['started'] for row in units.values() if 'started' in row]
    finished = [row['finished'] for row in units.values() if 'finished' in row]
    setups = [row['setup'] for row in units.values() if 'setup' in row]
    build = units.get('build-vet', {})
    tests = []
    for name, row in units.items():
        if name == 'build-vet':
            continue
        if 'tests' in row:
            tests.append(row['tests'])
        elif 'wall' in row:
            tests.append(row['wall'] - row.get('setup', 0.0))
    # A retried job can publish an earlier attempt's run (its units started before this record's stamp: 1e8eff51's
    # 14:37Z record carried its 13:04Z run, Oct 9). Its unit seconds stand, but nothing measured against this stamp does.
    earlier = bool(started) and min(started) < stampEpoch
    return {
        'select': min(started) - stampEpoch if started and not earlier else None,
        'fetch': max(setups) if setups else None,
        'build': build['wall'] - build.get('setup', 0.0) if 'wall' in build else None,
        'test': max(tests) if tests else None,
        'publish': commitEpoch - max(finished) if finished and not earlier else None,
    }


def boxPhases(fast, stampEpoch, commitEpoch):
    steps = fast.get('steps_seconds') or {}
    run = fast.get('wall_seconds')
    return {
        'select': max(run - max(steps.values()), 0.0) if run is not None and steps else None,
        'fetch': commitEpoch - stampEpoch - run if run is not None else None,
        'build': steps.get('build'),
        'test': steps.get('tests'),
        'publish': None,
    }


def phases(directory, stampEpoch, commitEpoch):
    """The record's route and its five phases, with its wall (stamp to record commit)."""
    # A box run cut short can publish a red with no fast.json (seven such records on Oct 9): its phases read unknown.
    try:
        with open(os.path.join(directory, 'fast.json')) as handle:
            fast = json.load(handle)
    except (OSError, ValueError):
        fast = {}
    if fast.get('runner') == 'pool':
        route, mapped = 'pool', poolPhases(readEvents(directory), stampEpoch, commitEpoch)
    else:
        route, mapped = 'box', boxPhases(fast, stampEpoch, commitEpoch)
    mapped['wall'] = commitEpoch - stampEpoch
    return route, fast, mapped


def seconds(value):
    return '-' if value is None else '%d' % round(value)


def statusLine(mode, verdict, route, mapped, queue, tools, runId):
    """One line of at most 160 characters with no all-caps word: the wall, its phases, the tools and the run."""
    head = 'probe-90 %s %s: wall %ss = select %s fetch %s build %s test %s publish %s' % (
        mode, verdict, seconds(mapped['wall']), seconds(mapped['select']), seconds(mapped['fetch']),
        seconds(mapped['build']), seconds(mapped['test']), seconds(mapped['publish']))
    tail = '; %s, tools %s, run %s' % (route, (tools or 'unknown')[:9], runId)
    line = head + ('; queue %ss' % seconds(queue) if queue is not None else '') + tail
    if len(line) > MaximumLine:
        line = head + tail
    if len(line) > MaximumLine:
        line = line[:MaximumLine]
    return line


def runIdentifier(fast, reference):
    """The pool's run, or the box record's own stamp (gate-logs/<sha12>/<stamp>/fast)."""
    if fast.get('pool_run'):
        return fast['pool_run']
    parts = reference.split('/')
    return 'record %s' % (parts[-2] if len(parts) >= 2 else reference)


def main(argv):
    if argv[:1] == ['branch'] and len(argv) == 3:
        print(branchName(argv[1], argv[2]))
        return 0
    if argv[:1] == ['skipped'] and len(argv) == 3:
        text = open(argv[2]).read() if os.path.exists(argv[2]) else ''
        glob = skippedBy(argv[1], text)
        if glob:
            print(glob)
            return 0
        return 1
    if argv[:1] == ['line']:
        parser = argparse.ArgumentParser(prog='record.py line')
        parser.add_argument('directory')
        for name in ('--mode', '--sha', '--ref', '--tools'):
            parser.add_argument(name, required=True)
        for name in ('--stamp-epoch', '--commit-epoch', '--pushed-epoch'):
            parser.add_argument(name, type=float, required=True)
        arguments = parser.parse_args(argv[1:])
        route, fast, mapped = phases(arguments.directory, arguments.stamp_epoch, arguments.commit_epoch)
        with open(os.path.join(arguments.directory, 'status.txt')) as handle:
            verdict = handle.readline().split(':', 1)[0].strip()
        print(statusLine(arguments.mode, verdict, route, mapped, arguments.stamp_epoch - arguments.pushed_epoch,
                         arguments.tools, runIdentifier(fast, arguments.ref)))
        return 0
    print(__doc__.split('\n\n')[1], file=sys.stderr)
    return 2


if __name__ == '__main__':
    sys.exit(main(sys.argv[1:]))

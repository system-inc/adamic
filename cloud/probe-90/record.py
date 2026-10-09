#!/usr/bin/env python3
"""The 90-second probe's pure parts (#c89weq5, for Outcome #awn479j): its branch name, the skip-globs check, and one
published fast-gate record read as five phases and a status line. cloud/probe-90.sh does the Git and the waiting and
calls this; cloud/probe-90-test.py tests it offline.

    record.py branch <UTC stamp> <leaf|all>
    record.py skipped <branch> <skip-globs file>          prints the glob that skips it, exit 1 when none does
    record.py line <record dir> --mode M --sha S --ref R --tools T --jobs J --pushed-epoch N --commit-epoch N
                   --stamp-epoch N [--enqueue-epoch N] [--running-epoch N]
    record.py pending --mode M --sha S --jobs J --pushed-epoch N --now-epoch N [--enqueue-epoch N] [--running-epoch N]
                      [--note TEXT]                       the line for a probe that got no green or red record

The line's wall is the push to the record commit: what a change sees. Its phases, from whichever route answered:

  queue    A+B (system_adamic's ruling, Oct 9: the queue is its own phase): A is Loom's job file written (enqueue,
           cloud/pool-job.sh) to its .running (fast.sh serves it); B is .running to the job's first unit placed (the
           selection unit's "started" in <sha>.work/select-record.jsonl). ">N" when it never placed.
  Loom's pool (fast.json runner "pool"; the normal route): no steps_seconds, so the rest come from record.jsonl, the
  run's unit events, and the record commit's time:
    select   the selection unit started to the first test or build unit started: the selection run and Loom placing
             the job's units (from the record's stamp, fast.sh's serve start, when the selection's start is unknown;
             unknown, with publish, when the record carries an earlier attempt's run, its units started before it)
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
  Phases overlap (tests run beside the build), so they don't sum to the wall. The line ends with the pool's depth when
  it was read: jobs running by tier, then queued by tier ("pool 40:8 30:4 queued 30:10 10:31").
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


def poolPhases(events, stampEpoch, commitEpoch, placeEpoch=None):
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
    origin = placeEpoch if placeEpoch is not None and placeEpoch >= stampEpoch else stampEpoch
    return {
        'select': min(started) - origin if started and not earlier else None,
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


def phases(directory, stampEpoch, commitEpoch, placeEpoch=None, pushedEpoch=None):
    """The record's route and its five phases, with its wall (the push to the record commit, or the stamp to it when
    the push time is unknown)."""
    # A box run cut short can publish a red with no fast.json (seven such records on Oct 9): its phases read unknown.
    try:
        with open(os.path.join(directory, 'fast.json')) as handle:
            fast = json.load(handle)
    except (OSError, ValueError):
        fast = {}
    if fast.get('runner') == 'pool':
        route, mapped = 'pool', poolPhases(readEvents(directory), stampEpoch, commitEpoch, placeEpoch)
    else:
        route, mapped = 'box', boxPhases(fast, stampEpoch, commitEpoch)
    mapped['wall'] = commitEpoch - (pushedEpoch if pushedEpoch is not None else stampEpoch)
    return route, fast, mapped


def seconds(value):
    return '-' if value is None else '%d' % round(value)


def placedEpoch(jobs, sha, runningEpoch=None):
    """When the job's selection unit started on the pool (its first unit placed), from <sha>.work/select-record.jsonl,
    or None. Only a start at or after this attempt's .running counts: the work directory outlives a void."""
    path = os.path.join(jobs, sha + '.work', 'select-record.jsonl')
    try:
        with open(path) as handle:
            times = [epoch(event['time']) for event in (json.loads(line) for line in handle if line.strip())
                     if event.get('type') == 'started' and event.get('unit') == 'select']
    except (OSError, ValueError, KeyError):
        return None
    if runningEpoch is not None:
        times = [moment for moment in times if moment >= runningEpoch - 5]
    return min(times) if times else None


def poolDepth(jobs):
    """Loom's fast jobs right now: running by tier, then queued (written, not decided, running or cancelled) by tier."""
    running, queued = {}, {}
    try:
        names = os.listdir(jobs)
    except OSError:
        return 'pool depth unknown'
    present = set(names)
    for name in names:
        sha = name[:-5] if name.endswith('.json') else ''
        if not re.fullmatch(r'[0-9a-f]{40}', sha) or sha + '.verdict' in present or sha + '.cancelled' in present:
            continue
        try:
            with open(os.path.join(jobs, name)) as handle:
                tier = int(json.load(handle).get('priority', 0))
        except (OSError, ValueError, TypeError, AttributeError):
            tier = 0
        side = running if sha + '.running' in present else queued
        side[tier] = side.get(tier, 0) + 1

    def tiers(counts):
        return ' '.join('%d:%d' % (tier, counts[tier]) for tier in sorted(counts, reverse=True)) or '0'
    return 'pool %s queued %s' % (tiers(running), tiers(queued))


def queueText(enqueueEpoch, runningEpoch, placeEpoch, nowEpoch=None):
    """A+B: enqueue to .running, .running to the first unit placed (">N" while it hasn't placed)."""
    first = seconds(runningEpoch - enqueueEpoch) if enqueueEpoch is not None and runningEpoch is not None else '-'
    if runningEpoch is None:
        second = '-'
    elif placeEpoch is not None:
        second = seconds(placeEpoch - runningEpoch)
    elif nowEpoch is not None:
        second = '>' + seconds(nowEpoch - runningEpoch)
    else:
        second = '-'
    return '%s+%s' % (first, second)


def fit(candidates):
    """The first candidate within 160 characters, else the last one cut to 160."""
    for line in candidates:
        if len(line) <= MaximumLine:
            return line
    return candidates[-1][:MaximumLine]


def statusLine(mode, verdict, route, mapped, queue, depth, tools, runId):
    """One line of at most 160 characters with no all-caps word: the wall, the queue and the gate's phases, the pool's
    depth, the tools and the run."""
    head = 'probe-90 %s %s %ss: queue %s select %s fetch %s build %s test %s publish %s' % (
        mode, verdict, seconds(mapped['wall']), queue, seconds(mapped['select']), seconds(mapped['fetch']),
        seconds(mapped['build']), seconds(mapped['test']), seconds(mapped['publish']))
    tools = (tools or 'unknown')[:9]
    short = runId.rsplit('-', 1)[-1] if runId.startswith('adamic-') else runId.replace('record ', '')
    return fit(['%s; %s; %s, tools %s, run %s' % (head, depth, route, tools, runId),
                '%s; %s; %s, tools %s, run %s' % (head, depth, route, tools, short),
                '%s; %s, tools %s, run %s; %s' % (head, route, tools, short, depth)])


def pendingLine(mode, sha, wall, queue, depth, note):
    """The line for a probe that got no green or red record: how long it waited, its queue, why, and the depth."""
    head = 'probe-90 %s no verdict in %ss: queue %s' % (mode, seconds(wall), queue)
    return fit(['%s%s; %s; sha %s' % (head, (', ' + note) if note else '', depth, sha[:9]),
                '%s; %s; sha %s' % (head, depth, sha[:9])])


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
    if argv[:1] in (['line'], ['pending']):
        parser = argparse.ArgumentParser(prog='record.py ' + argv[0])
        if argv[0] == 'line':
            parser.add_argument('directory')
            for name in ('--ref', '--tools'):
                parser.add_argument(name, required=True)
            for name in ('--stamp-epoch', '--commit-epoch'):
                parser.add_argument(name, type=float, required=True)
        else:
            parser.add_argument('--now-epoch', type=float, required=True)
            parser.add_argument('--note', default='')
        for name in ('--mode', '--sha', '--jobs'):
            parser.add_argument(name, required=True)
        parser.add_argument('--pushed-epoch', type=float, required=True)
        for name in ('--enqueue-epoch', '--running-epoch'):
            parser.add_argument(name, type=float)
        arguments = parser.parse_args(argv[1:])
        depth = poolDepth(arguments.jobs)
        place = placedEpoch(arguments.jobs, arguments.sha, arguments.running_epoch)
        if argv[0] == 'pending':
            queue = queueText(arguments.enqueue_epoch, arguments.running_epoch, place, arguments.now_epoch)
            print(pendingLine(arguments.mode, arguments.sha, arguments.now_epoch - arguments.pushed_epoch, queue, depth,
                              arguments.note))
            return 0
        # Never seen running: the record's stamp is fast.sh's serve start.
        running = arguments.running_epoch if arguments.running_epoch is not None else arguments.stamp_epoch
        place = place if place is not None else placedEpoch(arguments.jobs, arguments.sha, running)
        route, fast, mapped = phases(arguments.directory, arguments.stamp_epoch, arguments.commit_epoch, place,
                                     arguments.pushed_epoch)
        queue = queueText(arguments.enqueue_epoch, running, place) if route == 'pool' else '-'
        with open(os.path.join(arguments.directory, 'status.txt')) as handle:
            verdict = handle.readline().split(':', 1)[0].strip()
        print(statusLine(arguments.mode, verdict, route, mapped, queue, depth, arguments.tools,
                         runIdentifier(fast, arguments.ref)))
        return 0
    print(__doc__.split('\n\n')[1], file=sys.stderr)
    return 2


if __name__ == '__main__':
    sys.exit(main(sys.argv[1:]))

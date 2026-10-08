#!/usr/bin/env python3
"""The star-idle alarm: the critical path's wave-0 step (the ★ on the waterfall) must always have a live
turn in the gate. Kirk and @system_adamic, Oct 8: the star sat eighth in the queue unseen, and its first
red sat until a tick read the log. Every 15 s this reads the ★ step's Branches globs and the watcher's own
files, and pages the step's owner and @system_adamic the first time it finds the star

  - queued, not running, for over a minute (a lower-ranked tip holds the slot it needs), or
  - not running and not queued, its newest verdict red with no newer push since.

A gate running on any of its branches is a live turn and re-arms the alarm. It reads only files and
`ahra tasks`, so a broken watcher can't silence it.

  cloud/star-idle-alarm.py          # the loop (launchd: com.adamic.star-idle-alarm)
  cloud/star-idle-alarm.py --once   # one check: prints the state, sends what the loop would
"""
import fnmatch
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import time

state = Path(os.environ.get('ADAMIC_FAST_GATE_WATCH_STATE', os.path.expanduser('~/.adamic-fast-gate-watch')))
watchLog = Path(os.environ.get('ADAMIC_FAST_GATE_WATCH_LOG', os.path.expanduser('~/Projects/system/adamic-gate-logs/fast-gate-watch.log')))
ahraDirectory = os.environ.get('ADAMIC_FAST_GATE_AHRA_DIR', '/Users/kirkouimet/Projects/ahra')
queuedLimit = int(os.environ.get('ADAMIC_STAR_QUEUED_SECONDS', '60'))
fullLog = Path(os.environ.get('ADAMIC_FULL_GATE_LOG', os.path.expanduser('~/Projects/system/adamic-gate-logs/full-gate-main.log')))
mainReds = os.environ.get('ADAMIC_MAIN_REDS', '')  # a file standing in for cloud/merge-tree's, in tests
verdictLine = re.compile(r'^\S+ done (\S+): (green|red): ([0-9a-f]{40})\b(.*)$')


def ahra(*arguments):
    return subprocess.run(['ahra', *arguments], cwd=ahraDirectory, capture_output=True, text=True, check=True).stdout


def star():
    """The critical path's wave-0 step: its id, owner and Branches globs (none if it declares none)."""
    waterfall = json.loads(ahra('tasks', 'waterfall', 'system_adamic', '--json'))
    waves = {node['id']: node.get('wave') for node in waterfall['nodes']}
    path = [identifier for identifier in waterfall.get('criticalPath', []) if waves.get(identifier) == 0]
    if not path:
        return None
    shown = ahra('tasks', 'show', path[0])
    owner = re.search(r'^owner\s+@(\S+)', shown, re.M)
    branches = re.search(r'^\s*Branches:\s*(.+)$', shown, re.M)
    return {'id': path[0], 'owner': owner.group(1) if owner else '', 'globs': branches.group(1).split() if branches else []}


def lines(path):
    try:
        return path.read_text(errors='replace').splitlines()
    except OSError:
        return []


def matches(branch, globs):
    return any(fnmatch.fnmatchcase(branch, glob) for glob in globs)


def check(step, now):
    """The star's state and, when it has no live turn, the alarm's key and text."""
    globs = step['globs']
    for entry in sorted(state.glob('running/*')):
        fields = (lines(entry) or [''])[0].split()
        if fields and matches(fields[0], globs):
            return 'running %s %s' % (fields[0], fields[1][:12]), None, None
    queued = [line.split() for line in lines(state / 'queue') if len(line.split()) >= 4 and matches(line.split()[2], globs)]
    if queued:
        _, since, branch, sha = min(queued, key=lambda fields: int(fields[1]))[:4]
        waited = now - int(since)
        ranked = [line.split()[3] for line in lines(state / 'queue.ranked') if len(line.split()) >= 4]
        rank = ('%d of %d in the ranked queue' % (ranked.index(branch) + 1, len(ranked))) if branch in ranked else 'not yet ranked'
        if waited < queuedLimit:
            return 'queued %s %s, %d s' % (branch, sha[:12], waited), None, None
        return ('queued %s %s, %d s' % (branch, sha[:12], waited), 'queued:' + sha,
                '★%s has no live turn in the gate: %s %s has waited %d s queued, not running (%s); every slot it could take is busy. '
                'The watcher stops a lower-ranked big gate on request.' % (step['id'], branch, sha[:12], waited, rank))
    seen = {}
    for line in lines(state / 'seen'):
        fields = line.split()
        if len(fields) == 2:
            seen[fields[0]] = fields[1]
    last = None
    for line in lines(watchLog):
        found = verdictLine.match(line)
        if found and matches(found.group(1), globs):
            last = found
    if last is None:
        return 'idle, no verdict on its branches', None, None
    branch, verdict, sha, rest = last.groups()
    if verdict == 'red' and seen.get(branch) == sha:
        where = re.search(r'first failure at (\S+)', rest)
        return ('red %s %s, no newer push' % (branch, sha[:12]), 'red:' + sha,
                '★%s has no live turn in the gate: its newest verdict is red (%s %s, first failure at %s) and nothing newer has been pushed '
                'on its branches (%s).' % (step['id'], branch, sha[:12], where.group(1) if where else 'unknown', ' '.join(step['globs'])))
    return '%s %s %s' % (verdict, branch, sha[:12]), None, None


def mainRed():
    """Main's newest whole-gate verdict when it's red, from the full gate's log: (sha, step, gate-log ref)."""
    verdict, published = None, {}
    for line in lines(fullLog):
        found = re.match(r'^\S+ (red|green): ([0-9a-f]{40})\b(.*)$', line)
        if found:
            verdict = found
        ref = re.match(r'^published (gate-logs/([0-9a-f]{12})/\S+/full-main)', line)
        if ref:
            published[ref.group(2)] = ref.group(1)
    if verdict is None or verdict.group(1) != 'red':
        return None
    where = re.search(r'first failure at (\S+)', verdict.group(3))
    return verdict.group(2), where.group(1) if where else 'unknown', published.get(verdict.group(2)[:12], '')


def mainRedRow(sha):
    """main-reds.tsv's row for this main, from cloud/merge-tree (integration's): its owner and fix-forward branch."""
    try:
        text = Path(mainReds).read_text() if mainReds else subprocess.run(
            ['git', 'show', 'refs/remotes/origin/cloud/merge-tree:cloud/integration/main-reds.tsv'],
            capture_output=True, text=True, check=True, cwd=os.path.dirname(os.path.abspath(__file__))).stdout
    except (OSError, subprocess.CalledProcessError):
        return None
    for line in text.splitlines():
        if line.startswith('gate-logs/' + sha[:12] + '/'):
            owner = re.search(r'owner @(\w+)', line)
            fix = re.search(r'fix-forward (\S+)', line)
            closed = re.search(r'\b(LIFTED|FIXED FORWARD|CLOSED)\b', line)
            return {'owner': owner.group(1) if owner else '', 'fix': fix.group(1).rstrip('.,;') if fix else '', 'closed': bool(closed)}
    return {'owner': '', 'fix': '', 'closed': False}


def checkMain(now):
    """While main is red, the star is main's fix-forward (@system_adamic, Oct 8): it must be gating."""
    red = mainRed()
    if red is None:
        return 'main green', None, None, []
    sha, where, ref = red
    row = mainRedRow(sha) or {'owner': '', 'fix': '', 'closed': False}
    if row['closed']:
        return 'main red %s, explained in main-reds.tsv' % sha[:12], None, None, []
    owner = row['owner'] or 'system_adamic_integration'
    if not row['fix']:
        return ('main red %s, no fix-forward named' % sha[:12], 'main:' + sha,
                'Main %s is red at %s (%s) and main-reds.tsv names no fix-forward for it (a row with "owner @<name>" and '
                '"fix-forward <branch>"), so nothing lands and nothing is visibly fixing it.' % (sha[:12], where, ref or 'full-main'), [owner])
    fixStep = {'id': 'main-fix', 'globs': [row['fix']], 'owner': owner}
    summary, key, text = check(fixStep, now)
    if summary.startswith('running') or (summary.startswith('queued') and key is None):
        return 'main red %s, fix-forward %s' % (sha[:12], summary), None, None, []
    return ('main red %s, fix-forward %s' % (sha[:12], summary), 'main:%s:%s' % (sha, key or 'idle'),
            'Main %s is red at %s and its fix-forward %s has no live turn in the gate (%s). Nothing lands until it does.'
            % (sha[:12], where, row['fix'], summary), [owner])


def page(step, text):
    for recipient in dict.fromkeys(filter(None, step['owner'].split(',') + ['system_adamic'])):
        try:
            ahra('os', 'send', recipient, text, '--from', 'system_adamic_developer_tools')
        except (subprocess.CalledProcessError, OSError) as error:
            print('%s could not page %s: %s' % (time.strftime('%H:%M:%S', time.gmtime()), recipient, error), flush=True)


def once(step):
    now = int(time.time())
    summary, key, text = check(step, now) if step and step['globs'] else ('no star with Branches', None, None)
    mainSummary, mainKey, mainText, mainOwners = checkMain(now)
    print('%s ★%s %s; %s' % (time.strftime('%H:%M:%S', time.gmtime()), step['id'] if step else '-', summary, mainSummary), flush=True)
    for name, alarmKey, alarmText, owners in (('star-idle-alarmed', key, text, [step['owner']] if step else []),
                                              ('main-red-alarmed', mainKey, mainText, mainOwners)):
        alarmed = state / name
        previous = alarmed.read_text().strip() if alarmed.exists() else ''
        if alarmKey is None:
            alarmed.unlink(missing_ok=True)
        elif alarmKey != previous:
            page({'owner': ','.join(owners)}, alarmText)
            alarmed.write_text(alarmKey + '\n')
            print('%s paged %s and system_adamic' % (time.strftime('%H:%M:%S', time.gmtime()), ', '.join(owners) or 'no owner'), flush=True)


def main():
    step, refreshed = None, 0
    while True:
        if step is None or time.time() - refreshed >= 60:
            try:
                step, refreshed = star(), time.time()
            except (subprocess.CalledProcessError, OSError, ValueError, KeyError) as error:
                print('%s could not read the star: %s' % (time.strftime('%H:%M:%S', time.gmtime()), error), flush=True)
        once(step)
        if '--once' in sys.argv:
            return
        time.sleep(15)


if __name__ == '__main__':
    main()

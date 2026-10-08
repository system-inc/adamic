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


def page(step, text):
    for recipient in dict.fromkeys(filter(None, [step['owner'], 'system_adamic'])):
        try:
            ahra('os', 'send', recipient, text, '--from', 'system_adamic_developer_tools')
        except (subprocess.CalledProcessError, OSError) as error:
            print('%s could not page %s: %s' % (time.strftime('%H:%M:%S', time.gmtime()), recipient, error), flush=True)


def once(step):
    summary, key, text = check(step, int(time.time()))
    print('%s ★%s %s' % (time.strftime('%H:%M:%S', time.gmtime()), step['id'], summary), flush=True)
    alarmed = state / 'star-idle-alarmed'
    previous = alarmed.read_text().strip() if alarmed.exists() else ''
    if key is None:
        alarmed.unlink(missing_ok=True)
    elif key != previous:
        page(step, text)
        alarmed.write_text(key + '\n')
        print('%s paged %s and system_adamic' % (time.strftime('%H:%M:%S', time.gmtime()), step['owner'] or 'no owner'), flush=True)


def main():
    step, refreshed = None, 0
    while True:
        if step is None or time.time() - refreshed >= 60:
            try:
                step, refreshed = star(), time.time()
            except (subprocess.CalledProcessError, OSError, ValueError, KeyError) as error:
                print('%s could not read the star: %s' % (time.strftime('%H:%M:%S', time.gmtime()), error), flush=True)
        if step and step['globs']:
            once(step)
        if '--once' in sys.argv:
            return
        time.sleep(15)


if __name__ == '__main__':
    main()

#!/usr/bin/env python3
"""The star-idle alarm: the critical path's wave-0 step (the ★ on the waterfall) must always have a live
turn in the gate. Kirk and @system_adamic, Oct 8: the star sat eighth in the queue unseen, and its first
red sat until a tick read the log. Every 15 s this reads the ★ step's Branches globs and the watcher's own
files, and pages the step's owner and @system_adamic the first time it finds the star

  - queued, not running, for over a minute (a lower-ranked tip holds the slot it needs), or
  - not running and not queued, its newest verdict red with no newer push since.

A gate running on any of its branches is a live turn and re-arms the alarm. The chain's quiet worker too
(@system_adamic, Oct 8: V1's worker sat quiet two hours on the critical path and nobody knew): each of the
critical path's first three steps with no push to any of its branches for 20 minutes, and nothing of it
running or queued, pages its owner and @system_adamic, once a quiet spell. It reads only files and
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
# A whole gate of main runs about 50 minutes on Home; past an hour the loop is stuck, not busy.
busyLimit = int(os.environ.get('ADAMIC_MAIN_BUSY_SECONDS', '3600'))
# Once the pool is promoted it gates main first, and the box loop waits its grace before taking a main itself
# (full-gate-main.sh mainTurn, ADAMIC_POOL_GRACE); the page waits that out too.
poolPromoted = Path(os.environ.get('ADAMIC_POOL_PROMOTED', os.path.expanduser('~/.adamic-full-gate/pool-promoted')))
poolGrace = int(os.environ.get('ADAMIC_POOL_GRACE', '600'))
fullLog = Path(os.environ.get('ADAMIC_FULL_GATE_LOG', os.path.expanduser('~/Projects/system/adamic-gate-logs/full-gate-main.log')))
mainReds = os.environ.get('ADAMIC_MAIN_REDS', '')  # a file standing in for cloud/merge-tree's, in tests
verdictLine = re.compile(r'^(\d\d:\d\d:\d\d) done (\S+): (green|red): ([0-9a-f]{40})\b(.*)$')
pushLine = re.compile(r'^(\d\d:\d\d:\d\d) queued (\S+) ([0-9a-f]{40})\b')
quietLimit = int(os.environ.get('ADAMIC_CHAIN_QUIET_SECONDS', '1200'))
# Kirk, Oct 9 04:44Z: nothing in wave 0 sits untouched for 10 minutes, the star or not.
waveQuietLimit = int(os.environ.get('ADAMIC_WAVE_QUIET_SECONDS', '600'))
requestsFile = Path(os.environ.get('ADAMIC_FULL_GATE_REQUESTS', os.path.expanduser('~/.adamic-full-gate/requests')))


def superseded(branch, sha):
    """A train candidate integration has dropped from its requests file (cloud/integration/star-train.py: superseded
    slices leave it). Its verdict is history, never the star's state (@system_adamic, Oct 8 23:20Z)."""
    if not fnmatch.fnmatchcase(branch, 'cloud/land-train-*'):
        return False
    live = [line.strip() for line in lines(requestsFile) if line.strip()]
    return bool(live) and sha not in live


def ahra(*arguments):
    return subprocess.run(['ahra', *arguments], cwd=ahraDirectory, capture_output=True, text=True, check=True).stdout


def step(identifier, node=None):
    """A roadmap step's id, owner and branch globs (none if it declares none). The globs are the task's branches field
    (Oct 9: a real field, read from the waterfall's node), else a 'Branches:' line where a node doesn't carry one."""
    shown = ahra('tasks', 'show', identifier)
    owner = re.search(r'^owner\s+@(\S+)', shown, re.M)
    branches = re.search(r'^\s*Branches:\s*(.+)$', shown, re.M)
    globs = list((node or {}).get('branches') or []) or (branches.group(1).split() if branches else [])
    return {'id': identifier, 'owner': owner.group(1) if owner else '', 'globs': globs}


def star():
    """The critical path's wave-0 step, and its first three steps (the chain), as step() reads them."""
    waterfall = json.loads(ahra('tasks', 'waterfall', 'system_adamic', '--json'))
    waves = {node['id']: node.get('wave') for node in waterfall['nodes']}
    nodes = {node['id']: node for node in waterfall['nodes']}
    path = waterfall.get('criticalPath', [])
    # A step the waterfall doesn't list as ready waits on steps before it, not on its worker (compiler, Oct 9: V5 paged
    # quiet while it waited on V4, which waited on V2 and V3). A waterfall without a ready list leaves every step ready.
    ready = {entry if isinstance(entry, str) else entry.get('id') for entry in waterfall['ready']} if 'ready' in waterfall else None
    chain = [step(identifier, nodes.get(identifier)) for identifier in path[:3]]
    for entry in chain:
        entry['ready'] = ready is None or entry['id'] in ready
    first = [entry for entry in chain if waves.get(entry['id']) == 0]
    global waveZero
    waveZero = [node for node in waterfall['nodes']
                if node.get('wave') == 0 and node.get('kind', 'Task') == 'Task' and node.get('status') not in ('Done', 'Cancelled', 'Failed')]
    return (first[0] if first else None), chain


# Every open wave-0 task, as the waterfall's nodes: refreshed with the star, once a minute.
waveZero = []
waveOwners = {}


def waveOwner(identifier):
    """A task's owner username, read once (ownership rarely moves) from ahra tasks show."""
    if identifier not in waveOwners:
        try:
            waveOwners[identifier] = step(identifier)['owner']
        except (subprocess.CalledProcessError, OSError):
            return ''
    return waveOwners[identifier]


def checkWaveZero(now):
    """(name, key, text, owners, louder) for each wave-0 task that needs a page: no motion for waveQuietLimit seconds
    (its owner), or a new status saying exactly what the last one said (its owner and the parent, louder: a repeat is a
    stall wearing a status). Keys change with the task's clock, so each quiet spell or repeat pages once."""
    alarms = []
    for node in waveZero:
        identifier = node['id']
        touched = int(node.get('lastTouchedAt') or 0) // 1000
        quiet = now - touched
        key = 'quiet:%s:%d' % (identifier, touched) if touched and quiet >= waveQuietLimit else None
        text = ("#%s is in wave 0 and hasn't moved for %d minutes (%s, last status: %s). Kirk's rule: nothing in wave 0 "
                "sits untouched for 10. Post what is happening this second with an ETA, or reap, split or ask up."
                % (identifier, quiet // 60, node.get('status'), node.get('statusText') or 'none')) if key else None
        alarms.append(('wave-quiet-alarmed-' + identifier, key, text, [waveOwner(identifier)] if key else [], False))
        statusAt, statusText = str(node.get('statusAt') or ''), (node.get('statusText') or '').strip()
        seen = state / ('wave-status-' + identifier)
        previous = seen.read_text().split('\t', 1) if seen.exists() else ['', '']
        repeat = None
        if statusAt and statusText and statusAt != previous[0]:
            if statusText == previous[1].strip():
                repeat = 'repeat:%s:%s' % (identifier, statusAt)
            seen.write_text(statusAt + '\t' + statusText + '\n')
        if repeat:
            alarms.append(('wave-repeat-alarmed-' + identifier, repeat,
                           '#%s posted the same status twice in a row: "%s". A repeated status is a stall: what changed, '
                           'and what is the next move this tick?' % (identifier, statusText), [waveOwner(identifier)], True))
    return alarms


def lines(path):
    try:
        return path.read_text(errors='replace').splitlines()
    except OSError:
        return []


landedCache = {'fetched': 0}


def landed(sha):
    """Whether main already contains this sha (a test may name landed shas in ADAMIC_LANDED_SHAS)."""
    stand = os.environ.get('ADAMIC_LANDED_SHAS')
    if stand is not None:
        return sha in lines(Path(stand))
    repository = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    if time.time() - landedCache['fetched'] > 60:
        subprocess.run(['git', '-C', repository, 'fetch', '-q', 'origin', '+refs/heads/main:refs/remotes/origin/main'], capture_output=True)
        landedCache['fetched'] = time.time()
    subprocess.run(['git', '-C', repository, 'fetch', '-q', 'origin', sha], capture_output=True)
    return subprocess.run(['git', '-C', repository, 'merge-base', '--is-ancestor', sha, 'refs/remotes/origin/main'], capture_output=True).returncode == 0


def ageOf(clock, now):
    """Seconds since an HH:MM:SS UTC stamp from the watcher's log, taken as today's (yesterday's if later)."""
    hours, minutes, seconds = (int(part) for part in clock.split(':'))
    day = now - now % 86400
    stamp = day + hours * 3600 + minutes * 60 + seconds
    return now - (stamp if stamp <= now else stamp - 86400)


def matches(branch, globs):
    return any(fnmatch.fnmatchcase(branch, glob) for glob in globs)


def fullRuns():
    """The commits a one-off whole gate is running on now (cloud/full-gate-main.sh <sha>, started by hand on a
    star candidate); a test names them in ADAMIC_FULL_GATE_RUNNING."""
    stand = os.environ.get('ADAMIC_FULL_GATE_RUNNING')
    if stand is not None:
        return [line.strip() for line in lines(Path(stand)) if line.strip()]
    listing = subprocess.run(['ps', '-axo', 'command'], capture_output=True, text=True).stdout
    return re.findall(r'full-gate-main\.sh ([0-9a-f]{40})\b', listing)


def runningTurn(globs):
    """The one probe for a running turn, shared by the star and the chain's quiet check so they can't disagree
    (@system_adamic, Oct 8): a fast gate on any of the branches, or a whole gate (cloud/full-gate-main.sh <sha>)
    of one of their tips. None when neither runs."""
    for entry in sorted(state.glob('running/*')):
        fields = (lines(entry) or [''])[0].split()
        if fields and matches(fields[0], globs):
            return 'running %s %s' % (fields[0], fields[1][:12])
    tips = {}
    for line in lines(state / 'seen'):
        fields = line.split()
        if len(fields) == 2 and matches(fields[0], globs):
            tips[fields[1]] = fields[0]
    for sha in fullRuns():
        if sha in tips:
            return 'running full-main %s %s' % (tips[sha], sha[:12])
    # The loops' own whole gates carry no sha on their command line: their published record says running
    # (@system_adamic, Oct 9 00:14Z: V1 paged quiet while its whole gate ran). A superseded candidate's record is
    # history, even one a stop left reading running.
    for sha, branch in tips.items():
        if not superseded(branch, sha) and wholeGate(sha)[0] == 'running':
            return 'running whole gate %s %s' % (branch, sha[:12])
    return None


def check(step, now):
    """The star's state and, when it has no live turn, the alarm's key and text."""
    globs = step['globs']
    running = runningTurn(globs)
    if running:
        return running, None, None
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
        if found and matches(found.group(2), globs) and not superseded(found.group(2), found.group(4)):
            last = found
    if last is None:
        return 'idle, no verdict on its branches', None, None
    clock, branch, verdict, sha, rest = last.groups()
    if verdict == 'green' and seen.get(branch) == sha:
        # Green and not yet on main is a handoff waiting (@system_adamic, Oct 8): after a minute it pages the
        # lander, integration, not the author.
        if landed(sha):
            return 'green %s %s, landed' % (branch, sha[:12]), None, None
        if fnmatch.fnmatchcase(branch, 'cloud/land-train-*'):
            # A train slice lands on its whole gate, not its fast gate (Kirk and @system_adamic, Oct 8: each carries
            # Gate-runs: deferred and lands via push-main --full-gate). Its fast green is an early signal; the
            # handoff starts when the whole-gate record is green.
            record, since = wholeGate(sha)
            if record == 'red':
                return ('whole gate red %s %s' % (branch, sha[:12]), 'whole-red:' + sha,
                        '★%s: %s %s is green on its fast gate but its whole gate is red, so it can\'t land.%s' % (step['id'], branch, sha[:12], loadNote(wholeGateStatus.get(sha))))
            if record != 'green':
                return 'green %s %s, whole gate %s' % (branch, sha[:12], record), None, None
            clock = time.strftime('%H:%M:%S', time.gmtime(since))
        waited = ageOf(clock, now)
        if waited < queuedLimit:
            return 'green %s %s, %d s' % (branch, sha[:12], waited), None, None
        return ('green %s %s, unlanded %d s' % (branch, sha[:12], waited), 'green:' + sha,
                '★%s is green and not landed: %s %s went green %d s ago and nothing has pushed it to main. '
                'The handoff is the constraint now.' % (step['id'], branch, sha[:12], waited))
    if verdict == 'red' and seen.get(branch) == sha:
        where = re.search(r'first failure at (\S+)', rest)
        return ('red %s %s, no newer push' % (branch, sha[:12]), 'red:' + sha,
                '★%s has no live turn in the gate: its newest verdict is red (%s %s, first failure at %s) and nothing newer has been pushed '
                'on its branches (%s).%s' % (step['id'], branch, sha[:12], where.group(1) if where else 'unknown', ' '.join(step['globs']), loadNote(rest)))
    return '%s %s %s' % (verdict, branch, sha[:12]), None, None


def checkQuiet(chain, now):
    """Each chain step whose worker is quiet: no push to its branches for quietLimit seconds (the watcher logs
    every push it sees as 'queued'), and nothing of it running or queued. A step's time starts no earlier than
    when this alarm first saw it on the chain. Returns (summary, key, text, owners) for each step."""
    seenFile = state / 'chain-first-seen.json'
    try:
        firstSeen = json.loads(seenFile.read_text())
    except (OSError, ValueError):
        firstSeen = {}
    firstSeen = {entry['id']: firstSeen.get(entry['id'], now) for entry in chain}
    seenFile.write_text(json.dumps(firstSeen) + '\n')
    queued = [line.split()[2] for line in lines(state / 'queue') if len(line.split()) >= 4]
    pushes = [(now - ageOf(found.group(1), now), found.group(2), found.group(3)) for found in map(pushLine.match, lines(watchLog)) if found]
    # Globs more than one chain step names (compiler/views-rehearsal for V1 to V6) count a push for a step only
    # through commits whose Train-slice trailer names it.
    counts = {}
    for entry in chain:
        for glob in entry['globs']:
            counts[glob] = counts.get(glob, 0) + 1
    results = []
    for entry in chain:
        if not entry['globs']:
            results.append(('%s: no Branches' % entry['id'], None, None, []))
            continue
        if runningTurn(entry['globs']) or any(matches(branch, entry['globs']) for branch in queued):
            results.append(('%s: in the gate' % entry['id'], None, None, []))
            continue
        if not entry.get('ready', True):
            results.append(('%s: waits on the steps before it' % entry['id'], None, None, []))
            continue
        own = [glob for glob in entry['globs'] if counts[glob] == 1]
        if pregating(entry['globs'], own):
            results.append(('%s: in its pre-gate' % entry['id'], None, None, []))
            continue
        last = sorted([push for push in pushes if matches(push[1], own)] + unwatchedPushes(entry['globs'], own))
        pushedAgo = now - last[-1][0] if last else None
        # Quiet since the later of its last push and its arrival on the chain.
        onChain = now - firstSeen[entry['id']]
        quiet = onChain if pushedAgo is None else min(pushedAgo, onChain)
        if quiet < quietLimit:
            results.append(('%s: pushed %s' % (entry['id'], '%d s ago' % pushedAgo if pushedAgo is not None else 'never, on the chain %d s' % quiet), None, None, []))
            continue
        lastPush = '%s %s, %d min ago' % (last[-1][1], last[-1][2][:12], pushedAgo // 60) if last else 'none seen'
        results.append(('%s: quiet %d min' % (entry['id'], quiet // 60),
                        'quiet:%s:%s' % (entry['id'], last[-1][2] if last else 'none'),
                        '#%s is on the critical path and its worker is quiet: no push to its branches (%s) for %d minutes, and nothing of it '
                        'in the gate. Last push: %s.' % (entry['id'], ' '.join(entry['globs']), quiet // 60, lastPush),
                        [entry['owner']] if entry['owner'] else []))
    return results


watchedPrefixes = ('codex/', 'area/', 'devtools/', 'cloud/land-')
branchCache = {}


def branchCommits(glob):
    """Recent commits (committer time, branch, sha, Train-slice trailer) on origin's branches matching a glob the
    watcher doesn't watch, read from git at most once a minute (ADAMIC_BRANCH_COMMITS stands in, in tests: one
    'time branch sha slice' per line)."""
    stand = os.environ.get('ADAMIC_BRANCH_COMMITS')
    if stand is not None:
        rows = [line.split() + [''] for line in lines(Path(stand)) if line.strip()]
        return [(int(row[0]), row[1], row[2], row[3]) for row in rows if fnmatch.fnmatchcase(row[1], glob)]
    cached = branchCache.get(glob)
    if cached and time.time() - cached[0] < 60:
        return cached[1]
    repository = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    listing = subprocess.run(['git', '-C', repository, 'ls-remote', 'origin', 'refs/heads/' + glob], capture_output=True, text=True).stdout
    rows = []
    for line in listing.splitlines():
        sha, ref = line.split()
        branch = ref[len('refs/heads/'):]
        subprocess.run(['git', '-C', repository, 'fetch', '-q', 'origin', sha], capture_output=True)
        log = subprocess.run(['git', '-C', repository, 'log', sha, '--since=12 hours ago', '-n', '200',
                              '--format=%ct %H %(trailers:key=Train-slice,valueonly,separator=%x2C)'], capture_output=True, text=True).stdout
        for entry in log.splitlines():
            fields = entry.split()
            if len(fields) >= 2:
                rows.append((int(fields[0]), branch, fields[1], fields[2] if len(fields) > 2 else ''))
    branchCache[glob] = (time.time(), rows)
    return rows


pregates = Path(os.environ.get('ADAMIC_FULL_GATE_PREGATES', os.path.expanduser('~/.adamic-full-gate/pregate')))


def pregating(globs, own):
    """A pre-gate on Loom's pool running or queued on one of a step's commits is its live turn: the candidate waits on the
    pool, not on its worker (@system_adamic, Oct 9 01:45Z: V2 paged quiet while d7c6d796's pre-gate waited behind V1's).
    Loom writes ${pregates}/<sha>, first word running, queued, green or red. The step's commits are its own pushes, as the
    quiet check counts them, and the watched tips of its globs."""
    active = {path.name for path in pregates.glob('*') if (lines(path) or [''])[0].split()[:1] in (['running'], ['queued'])}
    if not active:
        return None
    shas = {sha for _, _, sha in unwatchedPushes(globs, own)}
    shas |= {line.split()[1] for line in lines(state / 'seen') if len(line.split()) == 2 and matches(line.split()[0], own)}
    found = sorted(active & shas)
    return ('pre-gate %s' % found[0][:12]) if found else None


def unwatchedPushes(globs, own):
    """(time, branch, sha) pushes on a step's globs the watcher doesn't watch: every commit on a branch only this
    step names, and on a shared one only commits whose Train-slice trailer the step's own globs name."""
    pushes = []
    for glob in globs:
        if glob.startswith(watchedPrefixes):
            continue
        for when, branch, sha, slices in branchCommits(glob):
            if glob in own or any(name and any(name in mine for mine in own) for name in slices.split(',')):
                pushes.append((when, branch, sha))
    return pushes


wholeGateCache = {}
# Each record's status line, by sha, as wholeGate last read it: a red page quotes its load label.
wholeGateStatus = {}


def loadNote(status):
    """What a page adds when the red's own record says its first failure ran above the box's core count (run.py's
    'under load' label, @system_adamic from the witness, Oct 9: about six of Oct 8's reds were load, each found by hand)."""
    found = re.search(r'under load \((load [\d.]+ on \d+ cores) at first failure\)', status or '')
    return (' Its record says the first failure ran under load (%s): read it as load before reading it as a bug.' % found.group(1)) if found else ''


def wholeGate(sha):
    """A commit's newest whole-gate record (gate-logs/<sha12>/<stamp>/full-main, a box's or the pool's): its state
    (none, running, green, red, void) and when it was published, read from git at most once a minute.
    ADAMIC_WHOLE_GATES stands in, in tests: 'sha state epoch' lines."""
    stand = os.environ.get('ADAMIC_WHOLE_GATES')
    if stand is not None:
        for line in lines(Path(stand)):
            fields = line.split(None, 3)
            if len(fields) >= 3 and fields[0] == sha:
                wholeGateStatus[sha] = fields[3] if len(fields) > 3 else ''
                return fields[1], int(fields[2])
        return 'none', 0
    cached = wholeGateCache.get(sha)
    if cached and time.time() - cached[0] < 60:
        return cached[1]
    repository = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    listing = subprocess.run(['git', '-C', repository, 'ls-remote', 'origin', 'refs/heads/gate-logs/%s/*' % sha[:12]],
                             capture_output=True, text=True).stdout
    refs = sorted(line.split()[1] for line in listing.splitlines() if line.split()[1].endswith('/full-main'))
    result = ('none', 0)
    if refs:
        subprocess.run(['git', '-C', repository, 'fetch', '-q', 'origin', refs[-1]], capture_output=True)
        status = subprocess.run(['git', '-C', repository, 'show', 'FETCH_HEAD:status.txt'], capture_output=True, text=True).stdout
        full = subprocess.run(['git', '-C', repository, 'show', 'FETCH_HEAD:full.json'], capture_output=True, text=True).stdout
        when = subprocess.run(['git', '-C', repository, 'log', '-1', '--format=%ct', 'FETCH_HEAD'], capture_output=True, text=True).stdout.strip()
        finished = '"finished": true' in full
        wholeGateStatus[sha] = status
        state = ('void' if status.startswith('void:') else 'green' if status.startswith('green:') and finished else
                 'red' if status.startswith('red:') and finished else 'running')
        result = (state, int(when) if when.isdigit() else 0)
    wholeGateCache[sha] = (time.time(), result)
    return result


mainHeadCache = {'checked': 0, 'sha': ''}


def mainHead():
    """Origin's main, asked at most every 30 s (a test names it in ADAMIC_MAIN_HEAD)."""
    stand = os.environ.get('ADAMIC_MAIN_HEAD')
    if stand is not None:
        return (lines(Path(stand)) or [''])[0].strip()
    if time.time() - mainHeadCache['checked'] >= 30:
        repository = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
        answer = subprocess.run(['git', '-C', repository, 'ls-remote', 'origin', 'refs/heads/main'], capture_output=True, text=True)
        if answer.returncode == 0 and answer.stdout.split():
            mainHeadCache['sha'] = answer.stdout.split()[0]
        mainHeadCache['checked'] = time.time()
    return mainHeadCache['sha']


def busyWith(log, now):
    """The main the full-gate loop is gating right now ('<sha12> on <box> since <hh:mm:ss>'): its latest 'full gate of main <sha>' line with no 'finished:'
    after it, started (by the line's UTC clock) less than busyLimit ago. Empty when the loop is idle or stuck."""
    for index in range(len(log) - 1, -1, -1):
        match = re.match(r'^(\d\d):(\d\d):(\d\d) full gate of main ([0-9a-f]{40})\b.*? on (\S+)$', log[index])
        if not match:
            continue
        if any(re.match(r'^\d\d:\d\d:\d\d finished:', line) for line in log[index + 1:]):
            return ''
        day = now - now % 86400
        started = day + int(match.group(1)) * 3600 + int(match.group(2)) * 60 + int(match.group(3))
        if started > now:
            started -= 86400
        return '%s on %s since %s' % (match.group(4)[:12], match.group(5), log[index][:8]) if now - started < busyLimit else ''
    return ''


def checkConfirmation(now):
    """A main that moved must have its whole gate started within a minute (@system_adamic, Oct 8: 54cbc125
    landed at about 20:59Z with the full-gate loop still paused, and nothing confirmed it until a hand found it).
    The full gate's log names every main it starts."""
    head = mainHead()
    if not head:
        return 'main head unknown', None, None, []
    log = lines(fullLog)
    if any(head in line for line in log if 'full gate of' in line):
        return 'main %s confirming' % head[:12], None, None, []
    # The loop gates one main at a time, so a main that lands mid-run waits for it (Oct 9 03:13Z: 745dc0bb paged 64 s
    # after landing while the loop was three minutes into e69fcba7). Quiet while that run is unfinished and younger than
    # busyLimit; a run past it, or a loop that isn't running anything, still pages.
    busy = busyWith(log, now)
    if busy:
        return 'main %s queued behind a running main gate, %s' % (head[:12], busy), None, None, []
    # The pool's record (or a box's published before its log line) confirms it as well as the loop's log does.
    if wholeGate(head)[0] in ('running', 'green', 'red'):
        return 'main %s confirming (%s record)' % (head[:12], wholeGate(head)[0]), None, None, []
    seen = state / 'main-head-first-seen'
    fields = (lines(seen) or [''])[0].split()
    if len(fields) != 2 or fields[0] != head:
        seen.write_text('%s %d\n' % (head, now))
        fields = [head, str(now)]
    waited = now - int(fields[1])
    limit = queuedLimit + (poolGrace if poolPromoted.exists() else 0)
    if waited < limit:
        return 'main %s not yet confirmed, %d s of %d' % (head[:12], waited, limit), None, None, []
    return ('main %s unconfirmed %d s' % (head[:12], waited), 'unconfirmed:' + head,
            "Main moved to %s %d s ago and no whole gate has started on it: the full-gate loop (com.adamic.full-gate-main, "
            "Home) isn't confirming main, so a landing's red can't show." % (head[:12], waited), ['system_adamic_developer_tools'])


def mainRed():
    """Main's newest whole-gate verdict when it's red, from the full gate's log: (sha, step, gate-log ref). Only main's
    own head counts: the loop also gates the star's train candidates, whose verdicts are their slices' (Oct 8 23:19Z,
    'Main 3e88a5766db1 is red' paged for a superseded train-2)."""
    verdict, published = None, {}
    head = mainHead()
    for line in lines(fullLog):
        found = re.match(r'^\S+ (red|green): ([0-9a-f]{40})\b(.*)$', line)
        if found and found.group(2) == head:
            verdict = found
        ref = re.match(r'^published (gate-logs/([0-9a-f]{12})/\S+/full-main)', line)
        if ref:
            published[ref.group(2)] = ref.group(1)
    if verdict is None or verdict.group(1) != 'red':
        return None
    where = re.search(r'first failure at (\S+)', verdict.group(3))
    return verdict.group(2), where.group(1) if where else 'unknown', published.get(verdict.group(2)[:12], ''), verdict.group(3)


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
    sha, where, ref, rest = red
    row = mainRedRow(sha) or {'owner': '', 'fix': '', 'closed': False}
    if row['closed']:
        return 'main red %s, explained in main-reds.tsv' % sha[:12], None, None, []
    owner = row['owner'] or 'system_adamic_integration'
    if not row['fix']:
        return ('main red %s, no fix-forward named' % sha[:12], 'main:' + sha,
                'Main %s is red at %s (%s) and main-reds.tsv names no fix-forward for it (a row with "owner @<name>" and '
                '"fix-forward <branch>"), so nothing lands and nothing is visibly fixing it.%s' % (sha[:12], where, ref or 'full-main', loadNote(rest)), [owner])
    fixStep = {'id': 'main-fix', 'globs': [row['fix']], 'owner': owner}
    summary, key, text = check(fixStep, now)
    # Running, queued under a minute, green under a minute or landed is a live turn; no verdict at all isn't.
    if summary.startswith('running') or (key is None and summary.startswith(('queued', 'green'))):
        return 'main red %s, fix-forward %s' % (sha[:12], summary), None, None, []
    key = key or 'idle'
    if key.startswith('green:'):
        return ('main red %s, fix-forward %s' % (sha[:12], summary), 'main:%s:%s' % (sha, key),
                'Main %s is red at %s and its fix-forward %s is green but not landed (%s). Nothing lands until it does.'
                % (sha[:12], where, row['fix'], summary), ['system_adamic_integration'])
    return ('main red %s, fix-forward %s' % (sha[:12], summary), 'main:%s:%s' % (sha, key),
            'Main %s is red at %s and its fix-forward %s has no live turn in the gate (%s). Nothing lands until it does.'
            % (sha[:12], where, row['fix'], summary), [owner])


def page(step, text, parent=True):
    for recipient in dict.fromkeys(filter(None, step['owner'].split(',') + (['system_adamic'] if parent else []))):
        try:
            ahra('os', 'send', recipient, text, '--from', 'system_adamic_developer_tools')
        except (subprocess.CalledProcessError, OSError) as error:
            print('%s could not page %s: %s' % (time.strftime('%H:%M:%S', time.gmtime()), recipient, error), flush=True)


def once(step, chain=()):
    now = int(time.time())
    summary, key, text = check(step, now) if step and step['globs'] else ('no star with Branches', None, None)
    mainSummary, mainKey, mainText, mainOwners = checkMain(now)
    confirmSummary, confirmKey, confirmText, confirmOwners = checkConfirmation(now)
    mainSummary += '; ' + confirmSummary
    print('%s ★%s %s; %s' % (time.strftime('%H:%M:%S', time.gmtime()), step['id'] if step else '-', summary, mainSummary), flush=True)
    quiet = checkQuiet(chain, now)
    if chain:
        print('%s chain: %s' % (time.strftime('%H:%M:%S', time.gmtime()), '; '.join(summary for summary, _, _, _ in quiet)), flush=True)
    starOwners = ['system_adamic_integration'] if key and key.startswith('green:') else ([step['owner']] if step else [])
    alarms = [('star-idle-alarmed', key, text, starOwners),
              ('main-red-alarmed', mainKey, mainText, mainOwners),
              ('main-confirm-alarmed', confirmKey, confirmText, confirmOwners)]
    alarms += [('chain-quiet-alarmed-' + entry['id'], quietKey, quietText, owners)
               for entry, (_, quietKey, quietText, owners) in zip(chain, quiet)]
    alarms = [alarm + (True,) for alarm in alarms]
    wave = checkWaveZero(now)
    print('%s wave 0: %d open, %d quiet' % (time.strftime('%H:%M:%S', time.gmtime()), len(waveZero),
                                           sum(1 for _, key, _, _, louder in wave if key and not louder)), flush=True)
    for name, alarmKey, alarmText, owners, parent in alarms + wave:
        alarmed = state / name
        previous = alarmed.read_text().strip() if alarmed.exists() else ''
        if alarmKey is None:
            if not name.startswith('wave-repeat-'):
                alarmed.unlink(missing_ok=True)
        elif alarmKey != previous:
            page({'owner': ','.join(owners)}, alarmText, parent)
            alarmed.write_text(alarmKey + '\n')
            print('%s paged %s%s' % (time.strftime('%H:%M:%S', time.gmtime()), ', '.join(owners) or 'no owner',
                                     ' and system_adamic' if parent else ''), flush=True)


def main():
    step, chain, refreshed = None, [], 0
    while True:
        if step is None or time.time() - refreshed >= 60:
            try:
                (step, chain), refreshed = star(), time.time()
            except (subprocess.CalledProcessError, OSError, ValueError, KeyError) as error:
                print('%s could not read the star: %s' % (time.strftime('%H:%M:%S', time.gmtime()), error), flush=True)
        once(step, chain)
        if '--once' in sys.argv:
            return
        time.sleep(15)


if __name__ == '__main__':
    main()

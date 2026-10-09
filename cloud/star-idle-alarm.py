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
fullLog = Path(os.environ.get('ADAMIC_FULL_GATE_LOG', os.path.expanduser('~/Projects/system/adamic-gate-logs/full-gate-main.log')))
mainReds = os.environ.get('ADAMIC_MAIN_REDS', '')  # a file standing in for cloud/merge-tree's, in tests
verdictLine = re.compile(r'^(\d\d:\d\d:\d\d) done (\S+): (green|red): ([0-9a-f]{40})\b(.*)$')
pushLine = re.compile(r'^(\d\d:\d\d:\d\d) queued (\S+) ([0-9a-f]{40})\b')
quietLimit = int(os.environ.get('ADAMIC_CHAIN_QUIET_SECONDS', '1200'))
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


def step(identifier):
    """A roadmap step's id, owner and Branches globs (none if it declares none)."""
    shown = ahra('tasks', 'show', identifier)
    owner = re.search(r'^owner\s+@(\S+)', shown, re.M)
    branches = re.search(r'^\s*Branches:\s*(.+)$', shown, re.M)
    return {'id': identifier, 'owner': owner.group(1) if owner else '', 'globs': branches.group(1).split() if branches else []}


def star():
    """The critical path's wave-0 step, and its first three steps (the chain), as step() reads them."""
    waterfall = json.loads(ahra('tasks', 'waterfall', 'system_adamic', '--json'))
    waves = {node['id']: node.get('wave') for node in waterfall['nodes']}
    path = waterfall.get('criticalPath', [])
    chain = [step(identifier) for identifier in path[:3]]
    first = [entry for entry in chain if waves.get(entry['id']) == 0]
    return (first[0] if first else None), chain


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
                        '★%s: %s %s is green on its fast gate but its whole gate is red, so it can\'t land.' % (step['id'], branch, sha[:12]))
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
                'on its branches (%s).' % (step['id'], branch, sha[:12], where.group(1) if where else 'unknown', ' '.join(step['globs'])))
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
        own = [glob for glob in entry['globs'] if counts[glob] == 1]
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


def wholeGate(sha):
    """A commit's newest whole-gate record (gate-logs/<sha12>/<stamp>/full-main, a box's or the pool's): its state
    (none, running, green, red, void) and when it was published, read from git at most once a minute.
    ADAMIC_WHOLE_GATES stands in, in tests: 'sha state epoch' lines."""
    stand = os.environ.get('ADAMIC_WHOLE_GATES')
    if stand is not None:
        for line in lines(Path(stand)):
            fields = line.split()
            if len(fields) == 3 and fields[0] == sha:
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


def checkConfirmation(now):
    """A main that moved must have its whole gate started within a minute (@system_adamic, Oct 8: 54cbc125
    landed at about 20:59Z with the full-gate loop still paused, and nothing confirmed it until a hand found it).
    The full gate's log names every main it starts."""
    head = mainHead()
    if not head:
        return 'main head unknown', None, None, []
    if any(head in line for line in lines(fullLog) if 'full gate of' in line):
        return 'main %s confirming' % head[:12], None, None, []
    seen = state / 'main-head-first-seen'
    fields = (lines(seen) or [''])[0].split()
    if len(fields) != 2 or fields[0] != head:
        seen.write_text('%s %d\n' % (head, now))
        fields = [head, str(now)]
    waited = now - int(fields[1])
    if waited < queuedLimit:
        return 'main %s not yet confirmed, %d s' % (head[:12], waited), None, None, []
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


def page(step, text):
    for recipient in dict.fromkeys(filter(None, step['owner'].split(',') + ['system_adamic'])):
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
    for name, alarmKey, alarmText, owners in alarms:
        alarmed = state / name
        previous = alarmed.read_text().strip() if alarmed.exists() else ''
        if alarmKey is None:
            alarmed.unlink(missing_ok=True)
        elif alarmKey != previous:
            page({'owner': ','.join(owners)}, alarmText)
            alarmed.write_text(alarmKey + '\n')
            print('%s paged %s and system_adamic' % (time.strftime('%H:%M:%S', time.gmtime()), ', '.join(owners) or 'no owner'), flush=True)


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

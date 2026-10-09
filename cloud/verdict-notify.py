#!/usr/bin/env python3
"""Sends a finished fast gate's verdict to whoever waits on it, the moment the watcher reaps the gate.

    cloud/verdict-notify.py <branch> <sha> <gate log> [--early <gate-logs ref>]

Who hears it, from integration's own tables on cloud/merge-tree (cloud/integration/areas.tsv and its
router, area-route.py, the one auto-area-merge trusts):
  area/<name>    the area's owner, and integration
  cloud/land-*   integration, and the owner of the area the landing merges (its subject's 'Merge <area>/...')
                 or names (area/<name>)
  codex/*        the owner of the area integration's router sends the branch to; a held branch goes
                 to every @name the hold's note names, or to integration when it names nobody
  devtools/*     developer tools, except devtools/fast-gate: the gate's own tools branch, gated against main
                 with the whole tools delta, reads red at darwin on every push and is checked by its canaries
Integration hears every verdict, not only landings and areas, while the watcher's state directory holds a
file named verdicts-to-integration. The message: green or red, the base it was gated against, the first
failure's step and its package and test with the failure's last lines, the gate-logs ref, and the worker's
session when the gate knew it. Green or red only: a void gate isn't a verdict, and the watcher reports
those itself. --early sends a complete run's red at its first failure (the watcher publishes that ref
the moment the failure is in the log); the final verdict follows when the run ends.
"""
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import tempfile
import time

here = Path(__file__).resolve().parent.parent
state = Path(os.environ.get('ADAMIC_FAST_GATE_WATCH_STATE', str(Path.home() / '.adamic-fast-gate-watch')))
ahra = Path(os.environ.get('ADAMIC_FAST_GATE_AHRA_DIR', '/Users/kirkouimet/Projects/ahra'))
integration = 'system_adamic_integration'
tables = ('areas.tsv', 'area-routes.tsv', 'fleet-areas.tsv', 'census-owners.tsv', 'area-route.py')


def git(*args):
    return subprocess.run(['git', '-C', str(here), *args], capture_output=True, text=True).stdout


def routerDirectory():
    """Integration's tables, from cloud/merge-tree fetched at most every five minutes (or a fixed
    directory for tests). A table that can't be read leaves the router unavailable."""
    fixed = os.environ.get('ADAMIC_VERDICT_ROUTES_DIR')
    if fixed:
        return Path(fixed)
    ref = 'refs/remotes/origin/cloud/merge-tree'
    stamp = state / 'verdict-routes-fetched'
    state.mkdir(parents=True, exist_ok=True)
    if not stamp.exists() or time.time() - stamp.stat().st_mtime > 300:
        subprocess.run(['git', '-C', str(here), 'fetch', '-q', 'origin', '+refs/heads/cloud/merge-tree:' + ref], capture_output=True)
        stamp.touch()
    directory = Path(tempfile.mkdtemp())
    for name in tables:
        text = git('show', '%s:cloud/integration/%s' % (ref, name))
        if not text:
            raise RuntimeError('cloud/merge-tree has no cloud/integration/%s' % name)
        (directory / name).write_text(text)
    return directory


def owners(directory):
    rows = [line.split('\t') for line in (directory / 'areas.tsv').read_text().splitlines() if line.strip() and not line.startswith('#')]
    return {row[0]: row[1] for row in rows if len(row) > 1}


def router(directory):
    """Integration's router module, loaded exactly as auto-area-merge loads it."""
    source = directory / 'area-route.py'
    module = {'__file__': str(source), '__name__': 'trusted_area_route'}
    exec(compile(source.read_text(), str(source), 'exec'), module)
    return module


def labelArea(branch, module):
    """The area of the fleet whose session is labeled as the branch is named (codex/<label>), through
    integration's fleet-areas.tsv. A verdict often lands before any reply naming the branch is fetched,
    so the router's own fleet lookup can't see it yet; a scout's label is its branch's name."""
    label = branch.split('/', 1)[1]
    try:
        roster = json.loads(Path(os.environ.get('ADAMIC_FLEET_ROSTER', str(ahra / 'modules/ai/data/fleet-roster.json'))).read_text())
    except (OSError, ValueError):
        return None
    fleets = {member.get('fleet') for member in roster if member.get('label') == label and member.get('fleet')}
    areas = {module['fleetArea'](fleet) for fleet in fleets} - {None}
    return areas.pop() if len(areas) == 1 else None


def recipients(branch, sha):
    """The members to tell, in order, and a note when routing couldn't name an owner."""
    names, note = [], ''
    directory = None
    if branch.startswith('devtools/'):
        names.append('system_adamic_developer_tools')
    else:
        try:
            directory = routerDirectory()
            areaOwners = owners(directory)
            if branch.startswith('area/'):
                names += [areaOwners.get(branch[len('area/'):], integration), integration]
            elif branch.startswith('cloud/land-'):
                names.append(integration)
                message = git('log', '-1', '--format=%B', sha)
                # "Merge compiler/area-next-fixtures at ... onto main": the merged branch's area, then any area/<name>.
                for area in re.findall(r'^Merge (?:area/)?([a-z0-9-]+)/', message) + re.findall(r'\barea/([a-z0-9-]+)', message):
                    if area in areaOwners:
                        names.append(areaOwners[area])
                        break
            else:
                module = router(directory)
                area, why = module['route'](branch)
                if area != 'hold' and area not in areaOwners:
                    labeled = labelArea(branch, module)
                    if labeled in areaOwners:
                        area, why = labeled, 'its session label'
                if area == 'hold' or area not in areaOwners:
                    # A hold whose note names Circles goes to them alone: integration can't act on those
                    # (codex/views-*, codex/scout-*). An unrouted branch, or a hold naming nobody, is integration's.
                    named = re.findall(r'@([a-z0-9_]+)', why)
                    names += named or [integration]
                    note = ' Routed by the hold: %s.' % why.rstrip('.') if named else ' Routed to integration: %s.' % why.rstrip('.')
                else:
                    names.append(areaOwners[area])
        except Exception as error:
            names.append(integration)
            note = ' Routed to integration: the area router is unavailable (%s).' % error
        finally:
            if directory is not None and not os.environ.get('ADAMIC_VERDICT_ROUTES_DIR'):
                for name in tables:
                    (directory / name).unlink(missing_ok=True)
                directory.rmdir()
    if (state / 'verdicts-to-integration').exists():
        names.append(integration)
    seen = []
    for name in names:
        if name not in seen:
            seen.append(name)
    return seen, note


def excerpt(text):
    """The first failure's test and the last lines of its output, short enough for a message: the block
    the gate prints after 'FIRST FAILURE (...):', up to its verdict line, where the diagnostic is last."""
    match = re.search(r'^FIRST FAILURE \([^\n]*\n(.*?)(?=^(?:green|red|void): |\Z)', text, re.M | re.S)
    if not match:
        return ''
    lines = [line.rstrip()[:160] for line in match.group(1).splitlines()
             if line.strip() and not line.startswith(('=== RUN', '=== PAUSE', '=== CONT'))]
    if len(lines) > 9:
        lines = lines[:1] + ['...'] + lines[-8:]
    return '\n' + '\n'.join('    ' + line for line in lines)


def message(branch, sha, log, early=None):
    text = Path(log).read_text(errors='replace')
    verdicts = re.findall(r'^(green|red): ' + sha + r' (.*)$', text, re.M)
    if early is not None:
        # A complete run's first failure, before its verdict line exists.
        if not re.search(r'^FIRST FAILURE ', text, re.M):
            return None
        verdicts = [('red', '')]
    if not verdicts:
        return None
    verdict, rest = verdicts[-1]
    base = re.search(r'^fast gate: \S+ against (\S+) ([0-9a-f]{7})', text, re.M)
    published = re.findall(r'^published (gate-logs/\S+)', text, re.M)
    session = re.search(r'session ([0-9a-f]{8}-[0-9a-f-]{27})', rest)
    parts = ['Fast gate %s%s: %s %s' % (verdict, ' at its first failure (the complete run goes on)' if early is not None else '', branch, sha[:12])]
    if base:
        parts.append(' against %s %s' % (base.group(1), base.group(2)))
    if verdict == 'red':
        first = re.search(r'^FIRST FAILURE \((\S+?),[^\n]*\n(.*)$', text, re.M)
        if first:
            test = first.group(2).strip()
            where = test.split('/adamic/', 1)[-1] if re.match(r'^\S+ Test\S*$', test) else ''
            parts.append(', first failure at %s%s' % (first.group(1), (': ' + where) if where else ''))
        failed = re.search(r'(\d+) fail, (\d+) pass\s*$', rest)
        if failed and failed.group(1) != '0':
            parts.append(' (%s failed, %s passed)' % failed.groups())
        # The gate prints the box's load under its first failure; above the core count, the page says so first
        # (@system_adamic from the witness, Oct 9: about six of Oct 8's reds were load, each found by hand).
        box = re.search(r'^\(box load ([\d.]+) on (\d+) cores at first failure\)$', text, re.M)
        if box and float(box.group(1)) > int(box.group(2)):
            parts.append(', under load (load %s on %s cores at first failure)' % box.groups())
    else:
        took = re.search(r'fast gate in ([\d.]+ s)', rest)
        if took:
            parts.append(' in %s' % took.group(1))
    parts.append('.')
    if early is not None:
        parts.append(' Log: %s.' % early)
    elif published:
        parts.append(' Log: %s.' % published[-1])
    if session:
        parts.append(' Worker session %s.' % session.group(1))
    if verdict == 'red':
        parts.append(excerpt(text))
    return ''.join(parts)


def main():
    branch, sha, log = sys.argv[1:4]
    early = sys.argv[5] if len(sys.argv) > 5 and sys.argv[4] == '--early' else None
    if branch == 'devtools/fast-gate':
        return
    text = message(branch, sha, log, early)
    if text is None:
        return
    names, note = recipients(branch, sha)
    # ahra refuses all-caps words.
    text = re.sub(r'[A-Z]{3,}', lambda match: match.group(0).lower(), text + note)
    for name in names:
        sent = subprocess.run(['ahra', 'os', 'send', name, text, '--from', 'system_adamic_developer_tools'],
                              cwd=str(ahra), capture_output=True, text=True)
        print('%s %s %s: %s' % (time.strftime('%H:%M:%S', time.gmtime()), 'told' if sent.returncode == 0 else 'could not tell', name, text))


if __name__ == '__main__':
    main()

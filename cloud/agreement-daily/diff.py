#!/usr/bin/env python3
"""The agreement run's daily diff: a day's shards read out of their Loom job logs, and the day held against the one before.

    python3 -I diff.py compact <report.json>             the compact report job.sh prints for one shard
    python3 -I diff.py extract <job log>...              one day, merged from each shard's log, as JSON on stdout
    python3 -I diff.py compare <yesterday.json> <today.json>

compact keeps of the tool's --json report its head, shard, verdict, void, void_reason, counts and each program's path,
judgment, outcome_hash and task, the last two only where the tool wrote them. job.sh gzips and base64s it onto one
line between AGREEMENT_REPORT_BEGIN and AGREEMENT_REPORT_END.

extract refuses, exiting 2 and naming the file, a log with no report or with two, two logs of one shard, shards of
different counts or a path in two shards, and logs of different heads.

compare prints today's head and counts, then each section: moved (a judgment changed), output changed (the judgment
held and outcome_hash changed), new, gone, void and pending (today's, with the task each awaits, for its reader to
check the task is still open). A program judged void either day is listed under void alone, never as moved or changed,
since a void proves nothing. It exits 0 when moved, output changed, new and gone are all empty, else 1; anything it
can't read exits 2. A move or an output change opens a task: a human reads it, and nothing here decides it.
"""
import base64
import binascii
import gzip
import json
import re
import sys
import zlib

BEGIN = 'AGREEMENT_REPORT_BEGIN'
END = 'AGREEMENT_REPORT_END'
# The report's count fields, in the order the tool's summary line says them.
COUNTS = ('compared', 'agreeing', 'ruled', 'checked_stops', 'pending', 'disagreeing', 'memory_unsafe', 'refused')
SHARD = re.compile(r'^([1-9][0-9]*)/([1-9][0-9]*)$')


class Refusal(Exception):
    """A file this can't read as what it should be; main says it and exits 2."""


def compact(report):
    """The compact form of the tool's --json report: the fields diff.py reads and nothing else."""
    if not isinstance(report, dict) or not isinstance(report.get('programs'), list):
        raise Refusal('it holds no programs list')
    programs = []
    for program in report['programs']:
        if not isinstance(program, dict) or not isinstance(program.get('path'), str):
            raise Refusal('a program has no path')
        kept = {'path': program['path'], 'judgment': program.get('judgment')}
        for name in ('outcome_hash', 'task'):
            if name in program:
                kept[name] = program[name]
        programs.append(kept)
    return {
        'head': report.get('head'),
        'shard': report.get('shard'),
        'verdict': report.get('verdict'),
        'void': report.get('void', False),
        'void_reason': report.get('void_reason', ''),
        'counts': {name: report.get(name, 0) for name in COUNTS},
        'programs': programs,
    }


def blob(log):
    """The one compact report between the markers in a job log's text, decoded. A Loom log may put a stamp before each
    line, so a marker is a line's last word, and it keeps a line to 49,152 bytes, so the report job.sh printed on one
    line comes back on as many as it takes (J-z6bxcxr's on two): the report is every line's last word between the
    markers, joined."""
    lines = log.splitlines()
    found = []
    for index, line in enumerate(lines):
        words = line.split()
        if words and words[-1] == BEGIN:
            pieces = []
            for following in lines[index + 1:]:
                last = (following.split() or [''])[-1]
                if last == END:
                    break
                pieces.append(last)
            else:
                raise Refusal(f'its {BEGIN} has no {END} after it')
            if not pieces:
                raise Refusal(f'its {BEGIN} has no report line before {END}')
            found.append(''.join(pieces))
    if not found:
        raise Refusal(f'it holds no report between {BEGIN} and {END}')
    if len(found) > 1:
        raise Refusal(f'it holds {len(found)} reports, not one')
    try:
        report = json.loads(gzip.decompress(base64.b64decode(found[0], validate=True)))
    except (binascii.Error, OSError, EOFError, zlib.error, ValueError) as error:
        raise Refusal(f"its report doesn't decode: {error}")
    if not isinstance(report, dict) or not isinstance(report.get('programs'), list) or not isinstance(report.get('counts'), dict):
        raise Refusal('its report holds no counts and programs')
    if not isinstance(report.get('head'), str) or not SHARD.match(str(report.get('shard'))):
        raise Refusal('its report names no head and k/n shard')
    return report


def extract(paths):
    """One day from its shards' logs: the head, the shards in order, the counts summed and every program by path."""
    day = {'head': None, 'shards': [], 'counts': {name: 0 for name in COUNTS}, 'programs': {}}
    count = None
    for path in paths:
        try:
            with open(path, encoding='utf-8', errors='replace') as source:
                report = blob(source.read())
        except OSError as error:
            raise Refusal(f"{path}: it can't be read: {error}")
        except Refusal as refusal:
            raise Refusal(f'{path}: {refusal}')
        if day['head'] is None:
            day['head'] = report['head']
        elif report['head'] != day['head']:
            raise Refusal(f"{path}: its head {report['head']} isn't {day['head']}, the other logs' head")
        if report['shard'] in day['shards']:
            raise Refusal(f"{path}: shard {report['shard']} is in another log too")
        # Shards of different counts overlap, so they'd count a program twice
        shardCount = report['shard'].split('/')[1]
        if count is not None and shardCount != count:
            raise Refusal(f"{path}: shard {report['shard']} isn't of /{count}, as the other logs' are")
        count = shardCount
        day['shards'].append(report['shard'])
        for name in COUNTS:
            value = report['counts'].get(name, 0)
            if not isinstance(value, int):
                raise Refusal(f'{path}: its count {name} is {value!r}, no number')
            day['counts'][name] += value
        for program in report['programs']:
            if not isinstance(program, dict) or not isinstance(program.get('path'), str):
                raise Refusal(f'{path}: a program has no path')
            if program['path'] in day['programs']:
                raise Refusal(f"{path}: {program['path']} is in another shard too")
            day['programs'][program['path']] = {name: program[name] for name in ('judgment', 'outcome_hash', 'task') if name in program}
    day['shards'].sort(key=lambda shard: int(shard.split('/')[0]))
    return day


def load(path):
    """A day extract wrote."""
    try:
        with open(path, encoding='utf-8') as source:
            day = json.load(source)
    except (OSError, ValueError) as error:
        raise Refusal(f"{path}: it doesn't read as a day: {error}")
    if not isinstance(day, dict) or not isinstance(day.get('programs'), dict) or not isinstance(day.get('counts'), dict):
        raise Refusal(f'{path}: it holds no counts and programs')
    if not all(isinstance(program, dict) for program in day['programs'].values()):
        raise Refusal(f'{path}: a program is no object')
    return day


def compare(yesterday, today):
    """The report's lines, and whether the day holds: nothing moved, changed, came or went."""
    before, after = yesterday['programs'], today['programs']
    moved, changed, new, gone, void, pending = [], [], [], [], [], []
    for path in sorted(set(before) | set(after)):
        if path not in before:
            new.append(f"{path}: {after[path].get('judgment')}")
            continue
        if path not in after:
            gone.append(f"{path}: {before[path].get('judgment')}")
            continue
        was, now = before[path].get('judgment'), after[path].get('judgment')
        if was == 'void' or now == 'void':
            void.append(f'{path}: {was} -> {now}')
        elif was != now:
            moved.append(f'{path}: {was} -> {now}')
        elif 'outcome_hash' in before[path] and 'outcome_hash' in after[path] and before[path]['outcome_hash'] != after[path]['outcome_hash']:
            changed.append(path)
    for path in sorted(after):
        if after[path].get('judgment') == 'pending':
            pending.append(f"{path}: #{after[path].get('task', '?')}")
    counts = today['counts']
    lines = [
        f"head {today.get('head')}, shards {' '.join(today.get('shards') or [])}; yesterday {yesterday.get('head')}",
        'counts: ' + ', '.join(f"{name.replace('_', ' ')} {counts.get(name, 0)}" for name in COUNTS),
    ]
    for name, section in (('moved', moved), ('output changed', changed), ('new', new), ('gone', gone), ('void', void), ('pending', pending)):
        lines.append(f'{name} ({len(section)})')
        lines.extend('  ' + line for line in section)
    return lines, not (moved or changed or new or gone)


def main(arguments):
    try:
        if len(arguments) == 2 and arguments[0] == 'compact':
            try:
                with open(arguments[1], encoding='utf-8') as source:
                    report = json.load(source)
            except (OSError, ValueError) as error:
                raise Refusal(f"{arguments[1]}: it doesn't parse: {error}")
            try:
                print(json.dumps(compact(report), separators=(',', ':')))
            except Refusal as refusal:
                raise Refusal(f'{arguments[1]}: {refusal}')
            return 0
        if len(arguments) >= 2 and arguments[0] == 'extract':
            print(json.dumps(extract(arguments[1:]), indent=1, sort_keys=True))
            return 0
        if len(arguments) == 3 and arguments[0] == 'compare':
            lines, holds = compare(load(arguments[1]), load(arguments[2]))
            print('\n'.join(lines))
            return 0 if holds else 1
    except Refusal as refusal:
        print(f'diff.py: {refusal}', file=sys.stderr)
        return 2
    print('usage: diff.py compact <report.json> | diff.py extract <job log>... | diff.py compare <yesterday.json> <today.json>', file=sys.stderr)
    return 2


if __name__ == '__main__':
    sys.exit(main(sys.argv[1:]))

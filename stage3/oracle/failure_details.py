"""Observational failure details, independent of every acceptance predicate."""
import collections
import json
from pathlib import Path
import re

CLASSES = ('timeout', 'baseline-content', 'baseline-missing', 'exception', 'other')
ANSI = re.compile(r'\x1b\[[0-?]*[ -/]*[@-~]')


def reporter_failures(log):
    log = ANSI.sub('', log)
    summaries = list(re.finditer(r'^\s*\d+ failing\b', log, re.MULTILINE))
    if not summaries:
        return []
    tail = log[summaries[-1].end():]
    starts = list(re.finditer(r'^[ \t]*\d+\) (.*)', tail, re.MULTILINE))
    rows = []
    for index, start in enumerate(starts):
        block = tail[start.start():starts[index + 1].start() if index + 1 < len(starts) else len(tail)]
        lines = block.splitlines()
        parts, body = [], []
        in_title = True
        for number, line in enumerate(lines):
            text = re.sub(r'^\s*\d+\) ', '', line).strip() if number == 0 else line.strip()
            if in_title:
                if not text:
                    continue
                parts.append(text.removesuffix(':'))
                if text.endswith(':'):
                    in_title = False
            else:
                if re.match(r'^(Error in |Completed |Failed tasks:|npm error)', text):
                    break
                body.append(line)
        error = '\n'.join(body).strip()
        rows.append(dict(title=' '.join(parts), message=error, stack=error,
                         stack_top=next((line.strip() for line in body if re.match(r'\s*at ', line)), None),
                         elapsed_ms=None, timeout_ms=None, source='reporter'))
    return rows


def diff_previews(diff):
    files, current = {}, None
    for line in diff.splitlines(True):
        if line.startswith('--- reference/'):
            current = line[len('--- reference/'):].rstrip('\r\n')
            files[current] = []
        if current is not None:
            files[current].append(line)
    return {name: dict(path=name, preview=''.join(lines[:40]), total_lines=len(lines),
                       truncated=len(lines) > 40) for name, lines in files.items()}


def read_failures(log, diff='', observation_dir=None):
    previews = diff_previews(diff)
    events = collections.defaultdict(list)
    if observation_dir is not None and Path(observation_dir).exists():
        for file in sorted(Path(observation_dir).glob('*.jsonl')):
            for line in file.read_text().splitlines():
                row = json.loads(line)
                events[row['title'].strip()].append(row)
    captured = []
    for fallback in reporter_failures(log):
        observed = events[fallback['title']]
        row = dict(observed.pop(0), source='mocha') if observed else fallback
        row['title'] = fallback['title']
        captured.append(row)
    # A later worker crash can prevent the host from printing its failure summary.
    for remaining in events.values():
        captured.extend(dict(row, source='mocha', title=row['title'].strip()) for row in remaining)
    rows = []
    for row in captured:
        message = ANSI.sub('', row['message'])
        # Associate by explicit baseline paths, never by the run's global diff count.
        paths = [name for name in previews if name in message or
                 ('baseline' in message.lower() and Path(name).name in message)]
        baseline = bool(re.search(r'baseline|generated content', message, re.I) or
                        re.search(r'baseline', row['title'], re.I))
        limit = re.search(r'Timeout of (\d+)ms exceeded', message, re.I)
        if limit:
            cause = 'timeout'
            row['timeout_ms'] = int(limit[1])
        elif baseline and paths:
            cause = 'baseline-content'
        elif baseline:
            cause = 'baseline-missing'
        elif row.get('stack_top') or re.match(r'(?:\w*Error|Exception):', message):
            cause = 'exception'
        else:
            cause = 'other'
        row['cause'] = cause
        row['baselines'] = [previews[name] for name in paths]
        rows.append(row)
    return rows


def cause_counts(rows):
    counts = dict.fromkeys(CLASSES, 0)
    for row in rows:
        counts[row['cause']] += 1
    return counts


def markdown(rows):
    parts = []
    for row in rows:
        parts += [f"### {row['title']} ({row['cause']})\n", '```text\n' + row['message'] + '\n```\n']
        if row.get('elapsed_ms') is not None:
            parts.append(f"Elapsed: {row['elapsed_ms']:.3f} ms; timeout: {row['timeout_ms']} ms.\n")
        if row.get('stack_top'):
            parts.append('Stack top: ' + row['stack_top'] + '\n')
        for baseline in row['baselines']:
            parts += [baseline['path'] + ' (first 40 lines)\n', '```diff\n' + baseline['preview'] + '\n```\n']
    return '\n'.join(parts)

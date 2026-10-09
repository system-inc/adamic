#!/usr/bin/env python3
"""Summarize JSON Callgrind Ir; reject incomplete or inconsistent accounting."""
import argparse
import collections
import json
import hashlib
from pathlib import Path
import re


def summarize(path):
    # DWARF inline records can split one function among main.c and adamic.h,
    # or add a same-name call edge. Collapse those records, not their costs twice.
    symbols = {}
    files = {}
    source_file = "???"
    source_line = 0
    line_self = collections.Counter()
    own = collections.Counter()
    edges = collections.Counter()
    calls = collections.Counter()
    function = callee = None
    pending = False
    count = 0
    total = None
    for line in path.read_text().splitlines():
        if line.startswith('summary:'):
            total = int(line.split()[1])
        elif line.startswith(('fl=', 'fi=', 'fe=', 'cfl=', 'cfi=')):
            value = line.split('=', 1)[1]
            match = re.match(r'\((\d+)\)(?: (.*))?$', value)
            if match:
                identifier, name = match.groups()
                if name is not None:
                    files[identifier] = name
                value = files[identifier]
            if not line.startswith('c'):
                source_file = Path(value).name
        elif line.startswith(('fn=', 'cfn=')):
            value = line.split('=', 1)[1]
            match = re.match(r'\((\d+)\)(?: (.*))?$', value)
            if match:
                identifier, name = match.groups()
                if name is not None:
                    symbols[identifier] = name
                value = symbols[identifier]
            if line.startswith('fn='):
                function = value
            else:
                callee = value
        elif line.startswith('calls='):
            pending = True
            count = int(line[6:].split()[0])
        elif function and line and line[0] in '0123456789+-*':
            position = line.split()[0]
            if position != '*':
                source_line = source_line + int(position) if position[0] in '+-' else int(position)
            cost = int(line.split()[-1])  # The recorded event is Ir alone.
            if pending:
                edges[function, callee] += cost
                calls[function, callee] += count
                pending = False
            else:
                own[function] += cost
                line_self[source_file, source_line, function] += cost
    if total is None or sum(own.values()) != total:
        raise RuntimeError("callgrind self costs do not sum to summary")
    inclusive = own.copy()
    for (caller, target), cost in edges.items():
        if caller != target:
            inclusive[caller] += cost
    def rows(counter):
        return [dict(function=name, instructions=value,
                     percent=100 * value / total,
                     self=own[name], inclusive=inclusive[name])
                for name, value in counter.most_common()
                if name not in ('(below main)',) and not name.startswith('0x')]
    result = dict(total=total, inclusive=rows(inclusive), self=rows(own), self_all=dict(own),
                  line_self=[dict(file=f, line=n, function=fn, instructions=value)
                             for (f, n, fn), value in line_self.most_common()],
                  edges=[dict(caller=a, callee=b, instructions=value, calls=calls[a, b])
                         for (a, b), value in edges.most_common()])
    path.with_suffix('.json').write_text(json.dumps(result, indent=2) + '\n')
    return result


def prepare(artifacts, directory):
    cases = json.loads((artifacts / 'cases.json').read_text())
    answers = json.loads((artifacts / 'go.json').read_text())
    selected = [i for i, case in enumerate(cases)
                if case['name'].startswith('generated/') or
                (i % 16 == 0 and len(case['text'].encode()) <= 131072)]
    for i in sorted(range(len(cases)), key=lambda i: len(cases[i]['text']), reverse=True):
        if len(cases[i]['text'].encode()) <= 131072 and i not in selected:
            selected.append(i)
            break
    selected.sort()
    def escape(text):
        return text.replace('\\', '\\\\').replace('\n', '\\n').replace('\r', '\\r').replace('\t', '\\t')
    def answer(i):
        value = answers[i]
        return ('error\t' + escape(value['error']) if value['error'] else
                'ok\t' + escape(value['output'])) + '\n'
    directory.mkdir(parents=True, exist_ok=True)
    (directory / 'profile-cases.txt').write_text(''.join(
        cases[i]['name'] + '\t' + escape(cases[i]['text']) + '\n' for i in selected))
    (directory / 'profile-expected.txt').write_text(''.join(answer(i) for i in selected))
    (directory / 'expected.txt').write_text(''.join(answer(i) for i in range(len(cases))))
    (directory / 'manifest.txt').write_text(''.join(cases[i]['name'] + '\n' for i in selected))
    print(f'{len(selected)} texts, {sum(len(cases[i]["text"].encode()) for i in selected)} source bytes')
    print('profile input SHA256', hashlib.sha256((directory / 'profile-cases.txt').read_bytes()).hexdigest())


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('profile', type=Path)
    parser.add_argument('--prepare-to', type=Path, help='treat profile as the saved corpus artifacts directory')
    args = parser.parse_args()
    if args.prepare_to:
        prepare(args.profile, args.prepare_to)
        return
    result = summarize(args.profile)
    for mode in ('inclusive', 'self'):
        print(f'Top 20 {mode} instructions:')
        for row in result[mode][:20]:
            print(f'{row["instructions"]:>14,} {row["percent"]:6.2f}% {row["function"]}')


if __name__ == '__main__':
    main()

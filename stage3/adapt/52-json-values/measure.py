#!/usr/bin/env python3
"""Check the upstream projects and compare every native diagnostic at unchanged source locations."""
import argparse
from collections import Counter
import difflib
import json
from pathlib import Path
import re
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('before', type=Path)
parser.add_argument('after', type=Path)
parser.add_argument('output', type=Path)
parser.add_argument('--adamic', required=True, type=Path)
args = parser.parse_args()
repo = Path(__file__).resolve().parents[3]
args.output.mkdir(parents=True, exist_ok=True)
checker = repo / 'stage3/api/node_modules/typescript/lib/tsc.js'
commands = []
for label, tree in [('before', args.before.resolve()), ('after', args.after.resolve())]:
    for project in ['compiler', 'jsTyping', 'services', 'server', 'harness', 'testRunner']:
        command = ['node', str(checker), '-p', 'src/' + project, '--pretty', 'false']
        with (args.output / f'stock-{label}-{project}.log').open('w') as log:
            result = subprocess.run(command, cwd=tree, stdout=log, stderr=subprocess.STDOUT)
        commands.append(dict(tree=label, project=project, command=command, exit=result.returncode))
        assert result.returncode == 0, (label, project)
    roots = sorted(str(p) for p in (tree / 'src/compiler').rglob('*') if p.suffix in ['.ts', '.a'])
    command = [str(args.adamic.resolve()), 'types', *roots]
    with (args.output / f'types-{label}.stdout').open('w') as out, (args.output / f'types-{label}.stderr').open('w') as err:
        result = subprocess.run(command, stdout=out, stderr=err)
    commands.append(dict(tree=label, roots=len(roots), command=command, exit=result.returncode))
    assert result.returncode == 1, (label, result.returncode)

pattern = re.compile(r'^(.*?):(\d+):(\d+): error TS\d+:', re.M)
line_maps = {}
def old_line(relative, line):
    if relative not in line_maps:
        old = (args.before / relative).read_text().splitlines()
        new = (args.after / relative).read_text().splitlines()
        mapping = {}
        for block in difflib.SequenceMatcher(a=old, b=new, autojunk=False).get_matching_blocks():
            mapping.update((block.b+i+1, block.a+i+1) for i in range(block.size))
        line_maps[relative] = mapping
    return line_maps[relative].get(line)

def entries(label, tree):
    text = (args.output / f'types-{label}.stderr').read_text()
    assert 'panic:' not in text
    matches = list(pattern.finditer(text)); result = []
    for index, match in enumerate(matches):
        end = matches[index+1].start() if index+1 < len(matches) else len(text)
        entry = text[match.start():end]
        file = Path(match.group(1)); line = int(match.group(2)); column = int(match.group(3))
        if file.is_relative_to(tree):
            relative = file.relative_to(tree)
            if label == 'after':
                mapped = old_line(relative, line)
                if mapped is None:
                    result.append('NEW_CHANGED_LOCATION:' + entry)
                    continue
                line = mapped
            head = str(file) + ':' + match.group(2) + ':' + match.group(3)
            entry = str(relative) + f':{line}:{column}' + entry[len(head):]
        result.append(entry.replace(str(args.before.resolve()), '<tree>').replace(str(args.after.resolve()), '<tree>'))
    return result
old = entries('before', args.before.resolve()); new = entries('after', args.after.resolve())
added = list((Counter(new) - Counter(old)).elements()); removed = list((Counter(old) - Counter(new)).elements())
report = dict(commands=commands, native=dict(before=len(old), after=len(new), added=added, removed=removed,
    mapping='Exact diagnostic text and columns; map only byte-equal source lines with SequenceMatcher; preserve multiplicity.'))
(args.output / 'measurement.json').write_text(json.dumps(report, indent=2) + '\n')
assert not added, added
print(json.dumps({'native': report['native'], 'stock_projects': 6}, indent=2))

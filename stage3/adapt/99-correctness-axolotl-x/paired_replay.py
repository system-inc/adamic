#!/usr/bin/env python3
"""Compare ordered replay outcomes on independent main and main-plus-99x trees."""
import difflib
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import sys
from concurrent.futures import ThreadPoolExecutor

repo, baseline, accepted, output = map(lambda s: Path(s).resolve(), sys.argv[1:])
output.mkdir(parents=True, exist_ok=False)
env = dict(os.environ, NODE_PATH=str(repo / 'stage3/api/node_modules'))
owned = ['builder.ts', 'checker.ts', 'program.ts', 'tsbuildPublic.ts', 'watch.ts', 'watchPublic.ts']
pins = json.loads((repo / 'stage3/adapt/99-correctness-axolotl-x/owners.json').read_text())
owner_names = sorted({p['owner'].split('.')[-1] for p in pins})

def snapshot(tree):
    # Retain bytes so actual adapter edits can be measured independently of its report.
    return {str(p.relative_to(tree)): (p.read_bytes(), p.stat().st_mtime_ns)
            for p in tree.rglob('*') if p.is_file() and '.git' not in p.relative_to(tree).parts}

def run(label, tree, directory):
    before = snapshot(tree)
    logdir = output / label
    logdir.mkdir(exist_ok=True)
    command = ['node', str(directory / 'adapt.cjs'), str(tree)]
    with (logdir / (directory.name + '.log')).open('w') as log:
        result = subprocess.run(command, env=env, stdout=log, stderr=subprocess.STDOUT)
    after = snapshot(tree)
    writes, edits = [], []
    for name in sorted(before.keys() | after.keys()):
        b, a = before.get(name), after.get(name)
        if b != a:
            writes.append(name)
        if (b[0] if b else None) != (a[0] if a else None):
            old = b[0] if b else b''
            new = a[0] if a else b''
            diff = list(difflib.unified_diff(old.decode(errors='backslashreplace').splitlines(True), new.decode(errors='backslashreplace').splitlines(True), fromfile='before/'+name, tofile='after/'+name))
            edits.append(dict(file=name, added=sum(s.startswith('+') and not s.startswith('+++') for s in diff), removed=sum(s.startswith('-') and not s.startswith('---') for s in diff)))
            with (logdir / (directory.name + '.patch')).open('a') as patch:
                patch.write(''.join(diff))
    text = (logdir / (directory.name + '.log')).read_text()
    refusal = text.strip().replace(str(tree), '<tree>') if result.returncode else None
    return dict(adapter=directory.name, exit=result.returncode, files_written=writes, edits=edits, refusal=refusal)

records = {'before': [], 'after': []}
failures = []
for directory in sorted((repo / 'stage3/adapt').iterdir()):
    if not directory.is_dir():
        continue
    if directory.name == '99-correctness-axolotl-x':
        records['before'].append(dict(adapter=directory.name, excluded='99x is absent from main'))
        records['after'].append(run('after', accepted, directory))
    else:
        with ThreadPoolExecutor(max_workers=2) as pool:
            futures = [pool.submit(run, label, tree, directory) for label, tree in [('before', baseline), ('after', accepted)]]
            b, a = [f.result() for f in futures]
        records['before'].append(b)
        records['after'].append(a)
        if b != a:
            failures.append(dict(adapter=directory.name, reason='replay outcome differs', before=b, after=a))
        for label, record in [('before', b), ('after', a)]:
            refusal = record['refusal'] or ''
            files = [name for name in owned if re.search(r'(?<![A-Za-z0-9_])'+re.escape(name)+r'(?=[:\s\"\x27]|$)', refusal)]
            owners = [name for name in owner_names if re.search(r'\b'+re.escape(name)+r'\b', refusal)]
            if files or owners:
                failures.append(dict(adapter=directory.name, reason='refusal names a 99x file or owner', label=label, files=files, owners=owners, refusal=refusal))
    for label, rows in records.items():
        (output / (label+'.json')).write_text(json.dumps(rows, indent=2)+'\n')
    # Exclude only 99x itself from the equality comparison.
    rendered = {label: json.dumps([r for r in rows if r['adapter']!='99-correctness-axolotl-x'], indent=2).splitlines(True) for label, rows in records.items()}
    (output/'records.diff').write_text(''.join(difflib.unified_diff(rendered['before'], rendered['after'], fromfile='before.json',tofile='after.json')))
    (output/'status.json').write_text(json.dumps(dict(failures=failures, adapters_recorded=len(records['after']), complete=False),indent=2)+'\n')
    print(json.dumps(dict(adapter=directory.name, before=records['before'][-1], after=records['after'][-1], failures=failures)), flush=True)
    if failures:
        sys.exit(1)
(output/'status.json').write_text(json.dumps(dict(failures=[], adapters_recorded=len(records['after']), complete=True),indent=2)+'\n')

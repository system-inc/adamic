#!/usr/bin/env python3
"""Compare ordered replay outcomes on independent main and main-plus-99x trees."""
import difflib
import json
import os
from pathlib import Path
import subprocess
import sys
from concurrent.futures import ThreadPoolExecutor

repo, baseline, accepted, output = map(lambda s: Path(s).resolve(), sys.argv[1:5])
resume = '--resume' in sys.argv[5:]
output.mkdir(parents=True, exist_ok=resume)
env = dict(os.environ, NODE_PATH=str(repo / 'stage3/api/node_modules'))

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
            changes = [dict(before=''.join(old.decode(errors='backslashreplace').splitlines(True)[i:j]), after=''.join(new.decode(errors='backslashreplace').splitlines(True)[k:l])) for op,i,j,k,l in difflib.SequenceMatcher(None,old.decode(errors='backslashreplace').splitlines(True),new.decode(errors='backslashreplace').splitlines(True)).get_opcodes() if op!='equal']
            edits.append(dict(file=name, changes=changes, added=sum(s.startswith('+') and not s.startswith('+++') for s in diff), removed=sum(s.startswith('-') and not s.startswith('---') for s in diff)))
            with (logdir / (directory.name + '.patch')).open('a') as patch:
                patch.write(''.join(diff))
    text = (logdir / (directory.name + '.log')).read_text()
    refusal = text.strip().replace(str(tree), '<tree>') if result.returncode else None
    return dict(adapter=directory.name, exit=result.returncode, files_written=writes, edits=edits, refusal=refusal)

records = {label: json.loads((output/(label+'.json')).read_text()) if resume else [] for label in ['before','after']}
assert [r['adapter'] for r in records['before']] == [r['adapter'] for r in records['after']]
done = {r['adapter'] for r in records['after']}
failures = []
for directory in sorted((repo / 'stage3/adapt').iterdir()):
    if not directory.is_dir() or directory.name in done:
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

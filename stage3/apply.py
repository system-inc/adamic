#!/usr/bin/env python3
"""Create a fresh pinned tree and measure each successive adaptation."""
import json
import os
import shutil
from pathlib import Path
import subprocess
import sys
import tempfile

stage = Path(__file__).resolve().parent
pin = json.loads((stage / 'source.json').read_text())
if len(sys.argv) != 2:
    sys.exit('usage: stage3/apply.sh <new-output-directory>')
out = Path(sys.argv[1]).resolve()
if out.exists():
    sys.exit(f'refusing to replace existing output: {out}')
cache = Path(os.environ.get('STAGE3_CACHE', str(Path.home() / '.cache/adamic-stage3'))).resolve()
cache.mkdir(parents=True, exist_ok=True)

def run(*args, **kwargs):
    return subprocess.run(args, check=True, **kwargs)

mirror = cache / 'typescript.git'
if not mirror.exists():
    run('git', 'clone', '--bare', '--depth', '1', '--branch', pin['tag'], pin['repository'], str(mirror))
head = subprocess.check_output(['git', '--git-dir', str(mirror), 'rev-parse', pin['tag'] + '^{commit}'], text=True).strip()
if head != pin['commit']:
    sys.exit(f'pin mismatch: {head}')
run('git', 'clone', '--no-hardlinks', str(mirror), str(out))
run('git', '-C', str(out), 'checkout', '--detach', pin['commit'])
# The pristine measurement tree includes upstream's normal generated inputs.
run('node', str(stage / 'adapt/00-setup/adapt.cjs'), str(out))
with tempfile.TemporaryDirectory(prefix='stage3-index-') as scratch:
    env = dict(os.environ, GIT_INDEX_FILE=str(Path(scratch) / 'index'))
    def snapshot():
        run('git', '-C', str(out), 'add', '--all', env=env, stdout=subprocess.DEVNULL)
        run('git', '-C', str(out), 'add', '--all', '--force', 'src', env=env, stdout=subprocess.DEVNULL)
        return subprocess.check_output(['git', '-C', str(out), 'write-tree'], env=env, text=True).strip()
    # Include tracked files outside src, too; generated build inputs are only in src.
    run('git', '-C', str(out), 'read-tree', 'HEAD', env=env)
    pristine = previous = snapshot()
    rows = []
    def counts(before, after):
        changes = subprocess.check_output(['git', '-C', str(out), 'diff', '--numstat', before, after], text=True).splitlines()
        added = removed = 0
        for change in changes:
            a, r, name = change.split('\t', 2)
            if a == '-' or r == '-':
                raise RuntimeError(f'binary adaptation cannot be measured in lines: {name}')
            added += int(a)
            removed += int(r)
        return len(changes), added, removed
    adaptations = sorted(directory for directory in (stage / 'adapt').iterdir() if directory.is_dir())
    adaptation_env = dict(os.environ)
    if any(directory.name != '00-setup' for directory in adaptations):
        api = cache / 'api'
        api.mkdir(exist_ok=True)
        for manifest in ['package.json', 'package-lock.json']:
            shutil.copyfile(stage / 'api' / manifest, api / manifest)
        run('npm', 'ci', '--prefix', str(api), '--ignore-scripts', '--no-audit', '--no-fund')
        adaptation_env['NODE_PATH'] = str(api / 'node_modules')
    for directory in adaptations:
        if not directory.is_dir():
            continue
        if len(directory.name) < 4 or not directory.name[:2].isdigit() or directory.name[2] != '-':
            raise RuntimeError(f'invalid adaptation directory: {directory}')
        run('node', str(directory / 'adapt.cjs'), str(out), env=adaptation_env)
        current = snapshot()
        rows.append((directory.name, *counts(previous, current)))
        previous = current
    total = counts(pristine, previous)
    text = '# Patch set\n\nCompared with the pinned pristine tree after upstream diagnostic generation.\n'
    text += 'Rows measure incremental edits; total measures the final tree against pristine.\n\n'
    text += '| Adaptation | Files | Lines added | Lines removed |\n|---|---:|---:|---:|\n'
    for name, files, added, removed in rows:
        text += f'| {name} | {files} | {added} | {removed} |\n'
    text += f'| **Total** | {total[0]} | {total[1]} | {total[2]} |\n'
    (stage / 'patch-set.md').write_text(text)
print(out)

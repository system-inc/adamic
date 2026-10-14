#!/usr/bin/env python3
"""Offline lane product steps; each successor owns an independent mutable copy."""
import argparse
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys

sys.dont_write_bytecode = True
LANE = Path(__file__).resolve().parent
ROOT = LANE.parents[1]
sys.path.insert(0, str(ROOT / 'stage3'))
import apply
from artifact import input_key, payload_digest, upstream_packages_digest


def packages():
    source = apply.typescript_source(json.loads((ROOT / 'stage3/source.json').read_text()))
    modules = apply.installed_packages(source)
    actual, wanted = apply.packages_digest(modules), upstream_packages_digest(ROOT)
    if actual != wanted:
        raise ValueError(f'supplied upstream packages hash to {actual}, not pinned {wanted}')
    return source, modules


def run(command, tree):
    print(json.dumps(command), flush=True)
    subprocess.run(command, cwd=tree, check=True, timeout=55, env=dict(os.environ, npm_config_offline='true', npm_config_update_notifier='false'))


def regular_bins(tree):
    # buildcache publishes regular files. npm's only links are executable aliases;
    # portable shell trampolines preserve their target's own module resolution.
    for directory, directories, files in os.walk(tree, followlinks=False):
        for name in directories + files:
            path = Path(directory) / name
            if path.is_symlink():
                if path.parent.name != '.bin' or not path.resolve().is_file():
                    raise ValueError(f'unexpected installed package link: {path}')
                target = os.readlink(path)
                if target.startswith('/') or '"' in target or '$' in target or '`' in target:
                    raise ValueError('nonrelative executable alias')
                path.unlink()
                path.write_text('#!/bin/sh\nexec "$(dirname "$0")/' + target + '" "$@"\n')
                path.chmod(0o755)


def step(directory, phase, previous):
    key, inputs = input_key()
    os.environ['npm_config_cache'] = str(directory / 'npm-cache')
    os.environ['npm_config_userconfig'] = '/dev/null'
    (directory / 'npm-global.rc').write_text('')
    os.environ['npm_config_globalconfig'] = str(directory / 'npm-global.rc')
    if phase == 'install':
        source, modules = packages()
        if (source / 'package-lock.json').read_bytes() != (previous / 'tree/package-lock.json').read_bytes():
            raise ValueError('supplied packages lock differs from adapted tree lock')
        shutil.copytree(previous / 'tree', directory / 'tree', symlinks=True)
        shutil.copytree(modules, directory / 'tree/node_modules', symlinks=True)
        regular_bins(directory / 'tree/node_modules')
        phases = dict(apply=dict(exit=0), install=dict(exit=0))
    else:
        shutil.copytree(previous / 'tree', directory / 'tree', symlinks=True)
        phases = json.loads((previous / 'phases.json').read_text())
        command = ['npm', 'run', 'build'] if phase == 'compiler' else ['npm', 'exec', '--offline', 'hereby', '--', 'tests']
        run(command, directory / 'tree')
        phases['build' if phase == 'compiler' else 'harness'] = dict(exit=0)
    (directory / 'phases.json').write_text(json.dumps(phases, sort_keys=True) + '\n')
    if phase == 'harness':
        shutil.copytree(apply.pinned_packages(ROOT / 'stage3/api') / 'typescript', directory / 'api', symlinks=True)
        ready = dict(key=key, inputs=inputs, phases=phases, payload=payload_digest(directory / 'tree'))
        (directory / 'ready.json').write_text(json.dumps(ready, sort_keys=True) + '\n')
        (directory / 'shards-ready.json').write_text(json.dumps(dict(artifact=key, api_digest=payload_digest(directory / 'api')), sort_keys=True) + '\n')
        # This archive travels as an ordinary buildcache blob beside the directly usable layout.
        archive = directory / 'artifact.tgz'
        run(['tar', '--sort=name', '--mtime=@0', '--owner=0', '--group=0', '--numeric-owner',
             '-C', str(directory), '-I', 'gzip -1n', '-cf', str(archive),
             'ready.json', 'shards-ready.json', 'tree', 'api'], ROOT)
    shutil.rmtree(directory / 'npm-cache', ignore_errors=True)
    if input_key()[0] != key:
        raise ValueError('artifact inputs changed while building')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--packages-key', action='store_true')
    parser.add_argument('directory', nargs='?', type=Path)
    parser.add_argument('phase', nargs='?')
    parser.add_argument('previous', nargs='?', type=Path)
    args = parser.parse_args()
    if args.packages_key:
        print(apply.packages_digest(packages()[1]))
    else:
        step(args.directory, args.phase, args.previous)

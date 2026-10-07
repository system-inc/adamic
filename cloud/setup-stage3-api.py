#!/usr/bin/env python3
"""Install the API seat's own npm lock, validating its installed bytes on warm runs."""
import importlib.util
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import time

sys.dont_write_bytecode = True
SHARED = Path(__file__).with_name('setup-markdown-width.py')
spec = importlib.util.spec_from_file_location('npm_setup', SHARED)
npm = importlib.util.module_from_spec(spec)
spec.loader.exec_module(npm)


def tree_digest(directory):
    entries = []
    for path in sorted(directory.rglob('*')):
        relative = str(path.relative_to(directory))
        mode = path.lstat().st_mode & 0o777
        if path.is_symlink():
            # npm's tsc/tsserver links point inside this installed tree. Hash the
            # spelling, and the target's bytes via its ordinary entry in the tree.
            if not path.resolve().is_relative_to(directory.resolve()) or not path.exists():
                raise ValueError('npm symlink escapes or loses its target: ' + relative)
            entries.append([relative, mode, 'link', os.readlink(path)])
        elif path.is_file():
            entries.append([relative, mode, npm.digest(path.read_bytes())])
        elif path.is_dir():
            entries.append([relative, mode, 'directory'])
        else:
            raise ValueError('unexpected npm entry: ' + relative)
    return npm.digest(json.dumps([directory.stat().st_mode & 0o777, entries]).encode())


def stamp_key(lock, manifest, bootstrap, helper, node):
    return npm.installation_key(npm.digest(lock), npm.digest(manifest), npm.digest(bootstrap),
                                npm.digest(helper), node)


def prepare(repository, tools, node):
    repository, tools = Path(repository), Path(tools)
    api = repository / 'stage3/api'
    lock = api / 'package-lock.json'
    if not lock.exists():
        return 'skipped (stage3/api/package-lock.json absent)'
    bootstrap = Path(__file__).with_name('markdown-width') / 'npm-bootstrap.json'
    inputs = {name: (api / name).read_bytes() for name in ['package-lock.json', 'package.json']}
    bootstrap_bytes = bootstrap.read_bytes()
    key = stamp_key(inputs['package-lock.json'], inputs['package.json'],
                               bootstrap_bytes, Path(__file__).read_bytes() + SHARED.read_bytes(),
                               subprocess.check_output([node, '--version']).decode().strip())
    tools.mkdir(parents=True, exist_ok=True)
    stamp = tools / ('stage3-api-' + npm.digest(str(api.resolve()).encode()))
    modules = api / 'node_modules'
    if os.environ.get('ADAMIC_GATE_UNCACHED') != '1':
        try:
            saved = json.loads(stamp.read_text())
            if saved['key'] == key and saved['tree'] == tree_digest(modules):
                return 'skipped (validated API lock and installed bytes)'
        except (OSError, ValueError, KeyError):
            pass
    # npm ci runs against the actual seat, preserving its default bin links and
    # package layout. An edit during installation cannot publish a stale stamp.
    with tempfile.TemporaryDirectory(prefix='stage3-npm-', dir=tools) as temporary:
        scratch = Path(temporary)
        metadata = json.loads(bootstrap_bytes)
        archive = scratch / 'npm.tgz'
        subprocess.run(['curl', '-fsSL', metadata['url'], '-o', str(archive)], check=True)
        npm.verified_archive(archive.read_bytes(), metadata['integrity'])
        subprocess.run(['tar', '--no-same-owner', '-xzf', str(archive), '-C', str(scratch)], check=True)
        subprocess.run([node, str(scratch / 'package/bin/npm-cli.js'), 'ci', '--prefix', str(api),
                        '--cache', str(scratch / 'cache'), '--ignore-scripts', '--no-audit', '--no-fund',
                        '--fetch-retries=0', '--global=false', '--dry-run=false', '--install-strategy=hoisted',
                        '--registry=https://registry.npmjs.org'], check=True,
                       env=dict(os.environ, npm_config_update_notifier='false'))
        if any((api / name).read_bytes() != data for name, data in inputs.items()):
            raise RuntimeError('API manifests changed during npm ci; rerun setup')
        (scratch / 'stamp').write_text(json.dumps(dict(key=key, tree=tree_digest(modules))))
        (scratch / 'stamp').replace(stamp)
    return 'installed (npm ci --prefix stage3/api, integrity verified)'


if __name__ == '__main__':
    started = time.monotonic()
    answer = prepare(*sys.argv[1:])
    print(f'setup: stage3 API dependencies {answer}; step-duration={time.monotonic() - started:.3f}s')

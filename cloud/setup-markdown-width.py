#!/usr/bin/env python3
"""Install the locked Node oracle dependencies; validate their bytes on every hit."""
import base64
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import time


def digest(data):
    return hashlib.sha256(data).hexdigest()


def installation_key(lock, manifest, bootstrap, helper, node):
    return digest(json.dumps(dict(lock=lock, manifest=manifest, bootstrap=bootstrap,
                                  helper=helper, node=node), sort_keys=True).encode())


def tree_digest(directory):
    files = []
    for path in sorted(directory.rglob('*')):
        if path == directory / '.adamic-stamp':
            continue
        relative = str(path.relative_to(directory))
        mode = path.lstat().st_mode & 0o777
        if path.is_symlink():
            raise ValueError('unexpected installed symlink: ' + relative)
        elif path.is_file():
            files.append([relative, mode, digest(path.read_bytes())])
        elif path.is_dir():
            files.append([relative, mode, 'directory'])
        else:
            raise ValueError('unexpected installed entry: ' + relative)
    return digest(json.dumps([directory.stat().st_mode & 0o777, files]).encode())


def verified_archive(data, integrity):
    algorithm, expected = integrity.split('-', 1)
    if algorithm != 'sha512' or base64.b64encode(hashlib.sha512(data).digest()).decode() != expected:
        raise ValueError('npm bootstrap integrity mismatch')


def prepare(source, destination, node):
    source, destination = Path(source), Path(destination)
    if destination.resolve().is_relative_to('/root'):
        raise ValueError('Markdown dependencies must be outside /root')
    # Use the same immutable inputs for the key and npm, even if a checkout changes
    # during a download. The next invocation will validate the new checkout's key.
    inputs = {name: (source / name).read_bytes() for name in
              ['package-lock.json', 'package.json', 'npm-bootstrap.json']}
    key = installation_key(*(digest(inputs[name]) for name in
                             ['package-lock.json', 'package.json', 'npm-bootstrap.json']),
                           digest(Path(__file__).read_bytes()),
                           subprocess.check_output([node, '--version']).decode().strip())
    stamp = destination / '.adamic-stamp'
    if os.environ.get('ADAMIC_GATE_UNCACHED') != '1':
        try:
            saved = json.loads(stamp.read_text())
            if saved['key'] == key and saved['tree'] == tree_digest(destination):
                return 'skipped (validated lock and installed bytes)'
        except (OSError, ValueError, KeyError):
            pass
    destination.parent.mkdir(parents=True, exist_ok=True)
    # No globally installed npm or npm cache is assumed, even when Node was copied alone.
    with tempfile.TemporaryDirectory(prefix='markdown-install-', dir=destination.parent) as temporary:
        scratch = Path(temporary)
        bootstrap = json.loads(inputs['npm-bootstrap.json'])
        archive = scratch / 'npm.tgz'
        subprocess.run(['curl', '-fsSL', bootstrap['url'], '-o', str(archive)], check=True)
        verified_archive(archive.read_bytes(), bootstrap['integrity'])
        subprocess.run(['tar', '--no-same-owner', '-xzf', str(archive), '-C', str(scratch)], check=True)
        install = scratch / 'dependencies'
        install.mkdir(mode=0o755)
        for name in ['package.json', 'package-lock.json']:
            (install / name).write_bytes(inputs[name])
        subprocess.run([node, str(scratch / 'package/bin/npm-cli.js'), 'ci',
                        '--prefix', str(install), '--cache', str(scratch / 'cache'),
                        '--ignore-scripts', '--no-audit', '--no-fund', '--fetch-retries=0',
                        '--global=false', '--dry-run=false', '--install-strategy=hoisted',
                        '--registry=https://registry.npmjs.org'], check=True,
                       env=dict(os.environ, npm_config_update_notifier='false'))
        for path in install.rglob('*'):
            path.chmod(0o755 if path.is_dir() else 0o644)
        (install / '.adamic-stamp').write_text(json.dumps(dict(key=key, tree=tree_digest(install))))
        backup = scratch / 'previous'
        if destination.exists():
            destination.rename(backup)
        try:
            install.rename(destination)
        except BaseException:
            if backup.exists():
                backup.rename(destination)
            raise
    return 'installed (npm ci, integrity verified)'


if __name__ == '__main__':
    started = time.monotonic()
    answer = prepare(*sys.argv[1:])
    print(f'setup: markdown dependencies {answer}; step-duration={time.monotonic() - started:.3f}s')

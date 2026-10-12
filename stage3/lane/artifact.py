#!/usr/bin/env python3
"""Prepare a shared upstream build by input hash, outside the test-unit budget."""
import argparse
import errno
import fcntl
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import time

ROOT = Path(__file__).resolve().parents[2]
# Importing apply.py must not leave stage3/__pycache__ in the checkout, which nothing ignores.
sys.dont_write_bytecode = True
sys.path.append(str(Path(__file__).resolve().parents[1]))
from apply import installed_packages


def digest_files(root, files):
    digest = hashlib.sha256()
    for file in sorted(files):
        relative = str(file.relative_to(root)).encode()
        data = file.read_bytes()
        digest.update(len(relative).to_bytes(8, 'big'))
        digest.update(relative)
        digest.update(len(data).to_bytes(8, 'big'))
        digest.update(data)
    return digest.hexdigest()


def input_key(root=ROOT):
    # Broad by design: adapters consume each other's proof files as well as code.
    stage = root / 'stage3'
    # Installed packages are their lockfiles' (keyed here, and checked where they're used), not inputs of their own.
    files = [file for file in stage.rglob('*') if file.is_file()
             and file.relative_to(stage).parts[0] not in ('lane', 'oracle')
             and '__pycache__' not in file.parts and 'node_modules' not in file.parts]
    tools = {name: subprocess.check_output([name, '--version'], text=True).strip()
             for name in ('node', 'npm', 'git', 'python3')}
    # An arbitrary compiler API override is an additional untracked input.
    if os.environ.get('TSC_ADAPT_TYPESCRIPT'):
        raise ValueError('unset TSC_ADAPT_TYPESCRIPT for a reproducible artifact')
    inputs = dict(files=digest_files(root, files), tools=tools,
                  preparation=hashlib.sha256(Path(__file__).read_bytes()).hexdigest(),
                  platform=subprocess.check_output(['node', '-p',
                      'process.platform+"/"+process.arch'], text=True).strip())
    key = hashlib.sha256(json.dumps(inputs, sort_keys=True).encode()).hexdigest()
    return key, inputs


def payload_digest(tree):
    return digest_files(tree, [file for file in tree.rglob('*') if file.is_file()])


def link_or_copy(source, destination):
    try:
        os.link(source, destination)
    except OSError as error:
        if error.errno != errno.EXDEV:
            raise
        shutil.copy2(source, destination)


def link_packages(root, tree):
    """Give the adapted tree its devDependencies from Adamic's tree, never from npm ci.

    The tree's build installs stage3/typescript's lockfile once (npm ci --ignore-scripts), and that lockfile must be
    the adapted tree's own. The packages are linked in file by file, as prepare.sh links stage3/api's: the adapted
    tree gets a node_modules directory of its own, which its .gitignore's node_modules/ matches where a symlink
    wouldn't, and a file the build adds stays in this copy. The build writes nothing inside node_modules (measured).
    """
    project = root / 'stage3/typescript'
    adapted = hashlib.sha256((tree / 'package-lock.json').read_bytes()).hexdigest()
    pinned = hashlib.sha256((project / 'package-lock.json').read_bytes()).hexdigest()
    if adapted != pinned:
        raise ValueError(f'the adapted tree\'s package-lock.json is {adapted}, not stage3/typescript\'s {pinned}; '
                         'commit the adapted lockfile there')
    shutil.copytree(installed_packages(project), tree / 'node_modules', symlinks=True, copy_function=link_or_copy)


def prepare(cache, root=ROOT):
    started = time.monotonic()
    key, inputs = input_key(root)
    cache.mkdir(parents=True, exist_ok=True)
    artifact = cache / key
    with (cache / (key + '.lock')).open('a') as lock:
        fcntl.flock(lock, fcntl.LOCK_EX)
        if artifact.exists():
            ready = json.loads((artifact / 'ready.json').read_text())
            if ready['inputs'] != inputs or ready['payload'] != payload_digest(artifact / 'tree'):
                raise ValueError('artifact content or input identity changed; refusing reuse')
            return dict(key=key, tree=str(artifact / 'tree'), cache_hit=True,
                        seconds=round(time.monotonic() - started, 3))
        scratch = Path(tempfile.mkdtemp(prefix=key + '.', dir=cache))
        try:
            phases = {}
            commands = [('apply', ['bash', str(root / 'stage3/apply.sh'), str(scratch / 'tree')]),
                        ('install', None),
                        ('build', ['npm', 'run', 'build']),
                        ('harness', ['npm', 'exec', 'hereby', '--', 'tests'])]
            for name, command in commands:
                begin = time.monotonic()
                if command is None:
                    link_packages(root, scratch / 'tree')
                    phases[name] = dict(exit=0, seconds=round(time.monotonic() - begin, 3))
                    continue
                with (scratch / (name + '.log')).open('w') as log:
                    result = subprocess.run(command, cwd=root if name == 'apply' else scratch / 'tree',
                                            stdout=log, stderr=subprocess.STDOUT)
                phases[name] = dict(exit=result.returncode, seconds=round(time.monotonic() - begin, 3))
                if result.returncode:
                    raise RuntimeError(f'{name} exit {result.returncode}; see {scratch}/{name}.log')
            if input_key(root)[0] != key:
                raise ValueError('inputs changed while preparing the artifact')
            ready = dict(key=key, inputs=inputs, payload=payload_digest(scratch / 'tree'),
                         phases=phases, seconds=round(time.monotonic() - started, 3))
            (scratch / 'ready.json').write_text(json.dumps(ready, indent=2) + '\n')
            scratch.rename(artifact)
        except Exception:
            # Preserve failed preparation logs for diagnosis; never publish a partial artifact.
            raise
    return dict(key=key, tree=str(artifact / 'tree'), cache_hit=False,
                seconds=round(time.monotonic() - started, 3))


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('cache', type=Path)
    args = parser.parse_args()
    print(json.dumps(prepare(args.cache.resolve()), indent=2))

#!/usr/bin/env python3
"""Fetch the keyed adapted tree; --build-product prepares it outside test units."""
import argparse
import hashlib
from concurrent.futures import ThreadPoolExecutor
import tarfile
import urllib.request
import urllib.error
import json
import os
import shutil
from pathlib import Path
import subprocess
import sys
import tempfile

stage = Path(__file__).resolve().parent
FORMAT = 'adamic-stage3-tree-v1'


def product_key():
    # Names, modes, empty directories and bytes all participate. Tables outside
    # adapt are generated outputs, not inputs. API locks pin the adapter parser,
    # and the installed parser's bytes by their pinned digest (a build checks it).
    digest = hashlib.sha256(FORMAT.encode())
    digest.update(subprocess.check_output(['node', '--version']).strip())
    inputs = [stage / 'source.json', stage / 'apply.py', stage / 'api/package.json',
              stage / 'api/package-lock.json', stage / 'api/node_modules.sha256']
    inputs += sorted(path for path in (stage / 'adapt').rglob('*')
                     if '__pycache__' not in path.parts)
    for path in inputs:
        if path.is_symlink() and path.is_dir():
            raise RuntimeError(f'unkeyed symlink directory input: {path}')
        name = path.relative_to(stage).as_posix()
        kind = ('symlink:' + os.readlink(path)) if path.is_symlink() else (
            'directory' if path.is_dir() else 'file')
        data = b'' if path.is_dir() else path.read_bytes()
        record = json.dumps([name, kind, path.stat().st_mode & 0o777,
                             hashlib.sha256(data).hexdigest()], separators=(',', ':'))
        digest.update(record.encode() + b'\n')
    return digest.hexdigest()


def fetch(store, name, destination):
    if store.startswith(('https://', 'http://', 'file://')):
        with urllib.request.urlopen(store.rstrip('/') + '/' + name, timeout=20) as response:
            with destination.open('wb') as target:
                shutil.copyfileobj(response, target)
    else:
        shutil.copyfile(Path(store) / name, destination)


class ProductMissing(RuntimeError):
    """Only an absent manifest permits a local fallback build."""


def restore_product(store, key, out):
    # Verify before extracting and expose the output only after the complete archive has been restored.
    with tempfile.TemporaryDirectory(prefix='stage3-product-', dir=out.parent) as scratch:
        scratch = Path(scratch)
        manifest_path = scratch / 'manifest.json'
        archive = scratch / 'tree.tgz'
        try:
            try:
                fetch(store, key + '.manifest', manifest_path)
            except FileNotFoundError as error:
                raise ProductMissing(f'product {key} missing from {store}') from error
            except urllib.error.HTTPError as error:
                if error.code == 404:
                    raise ProductMissing(f'product {key} missing from {store}') from error
                raise
            manifest = json.loads(manifest_path.read_text())
            if manifest['format'] != FORMAT or manifest['key'] != key:
                raise RuntimeError('adapted product input key mismatch')
            parts = manifest['parts']
            if not parts or any(part['suffix'] != f'p{index:03d}'
                                for index, part in enumerate(parts)):
                raise RuntimeError('adapted product part manifest mismatch')
            def download(part):
                target = scratch / part['suffix']
                fetch(store, key + '.' + part['suffix'], target)
                with target.open('rb') as source:
                    actual = hashlib.file_digest(source, 'sha256').hexdigest()
                if target.stat().st_size != part['size'] or actual != part['sha256']:
                    raise RuntimeError('adapted product part hash mismatch')
                return target
            with ThreadPoolExecutor(max_workers=8) as pool:
                paths = list(pool.map(download, parts))
            with archive.open('wb') as target:
                for path in paths:
                    with path.open('rb') as source:
                        shutil.copyfileobj(source, target)
        except (OSError, KeyError, ValueError) as error:
            raise RuntimeError(f'product {key} unavailable; run apply.sh --build-product '
                               f'outside tests and publish it to {store}') from error
        with archive.open('rb') as source:
            actual = hashlib.file_digest(source, 'sha256').hexdigest()
        if archive.stat().st_size != manifest['size'] or actual != manifest['sha256']:
            raise RuntimeError('adapted product payload hash mismatch')
        tree = scratch / 'tree'
        tree.mkdir()
        with tarfile.open(archive) as payload:
            payload.extractall(tree, filter='data')
        os.rename(tree, out)


def save_product(store, key, out):
    if store.startswith(('https://', 'http://', 'file://')):
        raise RuntimeError('--build-product requires a local product store; publish the keyed parts and manifest')
    store = Path(store)
    store.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix='stage3-publish-', dir=store) as scratch:
        archive = Path(scratch) / 'tree.tgz'
        # Keep .git: downstream proofs use git show and diff against the pin.
        run('tar', '-C', str(out), '-czf', str(archive), '.')
        with archive.open('rb') as source:
            manifest = dict(format=FORMAT, key=key, size=archive.stat().st_size,
                            sha256=hashlib.file_digest(source, 'sha256').hexdigest())
        manifest['parts'] = []
        with archive.open('rb') as source:
            for index in range((archive.stat().st_size + 16 * 1024 * 1024 - 1) // (16 * 1024 * 1024)):
                data = source.read(16 * 1024 * 1024)
                suffix = f'p{index:03d}'
                part = Path(scratch) / suffix
                part.write_bytes(data)
                manifest['parts'].append(dict(suffix=suffix, size=len(data),
                                              sha256=hashlib.sha256(data).hexdigest()))
                os.replace(part, store / (key + '.' + suffix))
        manifest_path = Path(scratch) / 'manifest.json'
        manifest_path.write_text(json.dumps(manifest, sort_keys=True) + '\n')
        os.replace(manifest_path, store / (key + '.manifest'))


def publish_hook(store, key):
    """Describe local payloads for developer tools; never read credentials/upload."""
    store = Path(store).resolve()
    manifest_path = store / (key + '.manifest')
    manifest = json.loads(manifest_path.read_text())
    if manifest['format'] != FORMAT or manifest['key'] != key:
        raise RuntimeError('adapted product input key mismatch')
    payloads = []
    for index, part in enumerate(manifest['parts']):
        if part['suffix'] != f'p{index:03d}':
            raise RuntimeError('adapted product part manifest mismatch')
        payload = store / (key + '.' + part['suffix'])
        with payload.open('rb') as source:
            actual = hashlib.file_digest(source, 'sha256').hexdigest()
        if payload.stat().st_size != part['size'] or actual != part['sha256']:
            raise RuntimeError('adapted product part hash mismatch')
        payloads.append(dict(path=str(payload), asset='adamic/build-cache/' + payload.name))
    if not payloads:
        raise RuntimeError('adapted product part manifest mismatch')
    payloads.append(dict(path=str(manifest_path), asset='adamic/build-cache/' + manifest_path.name))
    return dict(format=FORMAT, key=key,
                inputs=['stage3/source.json', 'stage3/apply.py', 'stage3/api/package.json',
                        'stage3/api/package-lock.json', 'stage3/api/node_modules.sha256',
                        'stage3/adapt/** (except __pycache__)',
                        'node --version'],
                pinned_source=json.loads((stage / 'source.json').read_text()),
                payloads=payloads, publish_order='parts first, manifest last')


def run(*args, **kwargs):
    return subprocess.run(args, check=True, **kwargs)

def typescript_source(pin):
    """The pinned upstream checkout is a gate input; apply never clones it from the network or keeps one in HOME."""
    source = os.environ.get('ADAMIC_TYPESCRIPT_SOURCE')
    if not source:
        raise RuntimeError(f'set ADAMIC_TYPESCRIPT_SOURCE to the gate inputs\' TypeScript {pin["tag"]} '
                           f'checkout at {pin["commit"]}')
    head = subprocess.run(['git', '-C', source, 'rev-parse', 'HEAD'], capture_output=True, text=True).stdout.strip()
    if head != pin['commit']:
        raise RuntimeError(f'ADAMIC_TYPESCRIPT_SOURCE {source} is at {head or "no commit"}, '
                           f'not the pinned {pin["commit"]}')
    return Path(source)


def installed_packages(project):
    """An npm project's packages, installed from its lockfile by the tree's build or the runner, never by apply.

    npm's hidden lockfile records what was installed, in a prebuilt tree and in a runner's linked copy alike: each
    package it names must be the lockfile's entry, a package it leaves out must be one the lockfile marks optional
    (another platform's), and each package must be the version the lockfile pins.
    """
    modules = project / 'node_modules'
    lockfile = project / 'package-lock.json'
    try:
        installed = json.loads((modules / '.package-lock.json').read_text())['packages']
        wanted = {name: entry for name, entry in json.loads(lockfile.read_text())['packages'].items() if name}
        if any(wanted.get(name) != entry for name, entry in installed.items()):
            raise ValueError('its hidden lockfile names other packages')
        absent = [name for name, entry in wanted.items() if name not in installed and not entry.get('optional')]
        if absent:
            raise ValueError(f'{absent[0]} is not installed')
        for name in installed:
            if json.loads((project / name / 'package.json').read_text())['version'] != wanted[name]['version']:
                raise ValueError(f'{name} is not version {wanted[name]["version"]}')
    except (OSError, KeyError, ValueError) as error:
        raise RuntimeError(f'{modules} is not an install of {lockfile} ({error}); stage3 never runs npm, '
                           f'so the tree must provide them') from error
    return modules


# prepare.sh writes this into a checkout's linked stage3/api/node_modules; a prebuilt tree's has none.
runnerMarker = '.fast-gate-lockfile-sha256'


def packages_digest(modules):
    """sha256 over an install's every entry in sorted order: a directory's name, a regular file's bytes and executable
    bit, a symlink's target text. Only the runner's lockfile marker is left out."""
    digest = hashlib.sha256()
    for path in sorted(modules.rglob('*'), key=lambda path: path.relative_to(modules).as_posix()):
        name = path.relative_to(modules).as_posix()
        if name == runnerMarker:
            continue
        if path.is_symlink():
            record = [name, 'symlink', os.readlink(path)]
        elif path.is_dir():
            record = [name, 'directory']
        elif path.is_file():
            with path.open('rb') as source:
                record = [name, 'file', bool(path.stat().st_mode & 0o111),
                          hashlib.file_digest(source, 'sha256').hexdigest()]
        else:
            raise RuntimeError(f'{path} is neither a file, a directory nor a symlink')
        digest.update(json.dumps(record, separators=(',', ':')).encode() + b'\n')
    return digest.hexdigest()


def pinned_packages(project):
    """installed_packages, whose bytes must also be the ones <project>/node_modules.sha256 pins, since a version
    check reads no bytes. Only a platform-independent install can be pinned this way."""
    modules = installed_packages(project)
    pin = project / 'node_modules.sha256'
    actual = packages_digest(modules)
    if actual != pin.read_text().strip():
        raise RuntimeError(f'{modules} hashes to {actual}, not the {pin.read_text().strip()} that {pin} pins; '
                           f'its bytes are not the lockfile\'s install')
    return modules


def build_product(out):
    if any(os.environ.get(name) for name in ['CENSUS_TYPESCRIPT', 'TSC_ADAPT_TYPESCRIPT']):
        raise RuntimeError('keyed builds require the pinned stage3/api parser, without overrides')
    pin = json.loads((stage / 'source.json').read_text())
    source = typescript_source(pin)
    adaptations = sorted(directory for directory in (stage / 'adapt').iterdir() if directory.is_dir())
    adaptation_env = dict(os.environ)
    if any(directory.name != '00-setup' for directory in adaptations):
        adaptation_env['NODE_PATH'] = str(pinned_packages(stage / 'api'))
    # A file:// clone, since a local path ignores --depth: the pinned commit alone, not the checkout's history.
    run('git', 'clone', '--depth', '1', source.resolve().as_uri(), str(out))
    # The product names no path on the machine that built it.
    run('git', '-C', str(out), 'remote', 'remove', 'origin')
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
        for directory in adaptations:
            if not directory.is_dir():
                continue
            if len(directory.name) < 4 or not directory.name[:2].isdigit() or directory.name[2] != '-':
                raise RuntimeError(f'invalid adaptation directory: {directory}')
            run('git', '-C', str(out), 'update-ref', 'refs/stage3/' + directory.name + '-before', previous)
            run('node', str(directory / 'adapt.cjs'), str(out), env=adaptation_env)
            current = snapshot()
            run('git', '-C', str(out), 'update-ref', 'refs/stage3/' + directory.name + '-after', current)
            rows.append((directory.name, *counts(previous, current)))
            previous = current
        total = counts(pristine, previous)
        text = '# Patch set\n\nCompared with the pinned pristine tree after upstream diagnostic generation.\n'
        text += 'Rows measure incremental edits; total measures the final tree against pristine.\n\n'
        text += '| Adaptation | Files | Lines added | Lines removed |\n|---|---:|---:|---:|\n'
        for name, files, added, removed in rows:
            text += f'| {name} | {files} | {added} | {removed} |\n'
        text += f'| **Total** | {total[0]} | {total[1]} | {total[2]} |\n'
        (out / 'patch-set.md').write_text(text)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('output', type=Path, nargs='?', help='new adapted-tree directory')
    parser.add_argument('--write-table', action='store_true',
                        help='also update the checked-in stage3/patch-set.md')
    parser.add_argument('--build-product', action='store_true',
                        help='prepare keyed product outside test units')
    parser.add_argument('--publish-hook', action='store_true',
                        help='print local payload paths and key for developer tools publisher')
    parser.add_argument('--key', action='store_true', help='print the input hash only')
    parser.add_argument('--packages-digest', action='store_true',
                        help='print the digest of stage3/api/node_modules, for stage3/api/node_modules.sha256')
    args = parser.parse_args()
    if args.packages_digest:
        print(packages_digest(installed_packages(stage / 'api')))
        return
    key = product_key()
    if args.key:
        print(key)
        return
    if args.publish_hook:
        cache = Path(os.environ.get('STAGE3_CACHE', str(Path.home() / '.cache/adamic-stage3')))
        store = os.environ.get('STAGE3_PRODUCT_STORE', str(cache / 'products'))
        if store.startswith(('http://', 'https://', 'file://')):
            parser.error('--publish-hook requires the local product store')
        print(json.dumps(publish_hook(store, key), indent=2))
        return
    if args.output is None:
        parser.error('output is required')
    out = args.output.resolve()
    if out.exists():
        sys.exit(f'refusing to replace existing output: {out}')
    cache = Path(os.environ.get('STAGE3_CACHE', str(Path.home() / '.cache/adamic-stage3'))).resolve()
    store = os.environ.get('STAGE3_PRODUCT_STORE', str(cache / 'products'))
    if not args.build_product and 'STAGE3_PRODUCT_STORE' not in os.environ:
        if not (Path(store) / (key + '.manifest')).exists():
            # Remote reads are opt-in: the gate supplies its public asset base.
            # An unconfigured cold worker must use the local miss/build path.
            store = os.environ.get('ADAMIC_BUILD_CACHE_URL', store)
    out.parent.mkdir(parents=True, exist_ok=True)
    if args.build_product:
        build_product(out)
        save_product(store, key, out)
    else:
        try:
            restore_product(store, key, out)
        except ProductMissing:
            print(f'stage3 product cache miss: {key}; building adapted tree locally', file=sys.stderr)
            build_product(out)
            local_store = store if not store.startswith(('http://', 'https://', 'file://')) else str(cache / 'products')
            save_product(local_store, key, out)
    if args.write_table:
        (stage / 'patch-set.md').write_bytes((out / 'patch-set.md').read_bytes())
    print(out)


if __name__ == '__main__':
    main()

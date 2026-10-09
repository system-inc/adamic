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
    # adapt are generated outputs, not inputs. API locks pin the adapter parser.
    digest = hashlib.sha256(FORMAT.encode())
    digest.update(subprocess.check_output(['node', '--version']).strip())
    inputs = [stage / 'source.json', stage / 'apply.py',
              stage / 'api/package.json', stage / 'api/package-lock.json']
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
                        'stage3/api/package-lock.json', 'stage3/adapt/** (except __pycache__)',
                        'node --version'],
                pinned_source=json.loads((stage / 'source.json').read_text()),
                payloads=payloads, publish_order='parts first, manifest last')


def run(*args, **kwargs):
    return subprocess.run(args, check=True, **kwargs)

def build_product(out, cache):
    if any(os.environ.get(name) for name in ['CENSUS_TYPESCRIPT', 'TSC_ADAPT_TYPESCRIPT']):
        raise RuntimeError('keyed builds require the pinned stage3/api parser, without overrides')
    pin = json.loads((stage / 'source.json').read_text())
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
    args = parser.parse_args()
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
        cache.mkdir(parents=True, exist_ok=True)
        build_product(out, cache)
        save_product(store, key, out)
    else:
        try:
            restore_product(store, key, out)
        except ProductMissing:
            print(f'stage3 product cache miss: {key}; building adapted tree locally', file=sys.stderr)
            cache.mkdir(parents=True, exist_ok=True)
            build_product(out, cache)
            local_store = store if not store.startswith(('http://', 'https://', 'file://')) else str(cache / 'products')
            save_product(local_store, key, out)
    if args.write_table:
        (stage / 'patch-set.md').write_bytes((out / 'patch-set.md').read_bytes())
    print(out)


if __name__ == '__main__':
    main()

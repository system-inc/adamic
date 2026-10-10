#!/usr/bin/env python3
"""Prepare once, then fetch the oracle test hook by its checked content hash.

The key covers Go's effective dependency file lists, dirty sources, embedded
runtime files and toolchain flags. Discovery does not compile dependencies. The local store can be copied between workers; no
successful test observation is cached here.
"""
import argparse
import fcntl
import hashlib
import json
import os
from pathlib import Path
import subprocess
import tempfile


def records(text):
    decoder = json.JSONDecoder()
    offset = 0
    while offset < len(text):
        if text[offset].isspace():
            offset += 1
            continue
        value, offset = decoder.raw_decode(text, offset)
        yield value


def action_key(repository):
    packages = subprocess.check_output(
        ['go', 'list', '-deps', '-test', '-json', './internal/oracle'],
        cwd=repository, text=True)
    inputs = []
    manifests = {repository / 'go.mod', repository / 'go.sum', repository / 'go.work',
                 repository / 'go.work.sum', Path(__file__).resolve()}
    fields = ['GoFiles', 'CgoFiles', 'CFiles', 'CXXFiles', 'MFiles', 'HFiles',
              'FFiles', 'SFiles', 'SysoFiles', 'EmbedFiles']
    for package in records(packages):
        if package.get('Error') or package.get('DepsErrors'):
            raise RuntimeError('oracle dependency discovery failed')
        directory = Path(package['Dir'])
        files = []
        for field in fields:
            for name in package.get(field, []):
                path = directory / name
                files.append((field, name, hashlib.sha256(path.read_bytes()).hexdigest()))
        inputs.append((package['ImportPath'], str(directory), sorted(files)))
        module = package.get('Module', {})
        for dependency in [module, module.get('Replace', {})]:
            if dependency.get('GoMod'):
                path = Path(dependency['GoMod'])
                manifests.update([path, path.with_name('go.sum')])
    settings = json.loads(subprocess.check_output(
        ['go', 'env', '-json', 'GOOS', 'GOARCH', 'GOAMD64', 'GOARM64', 'GOFLAGS',
         'CGO_ENABLED', 'GOEXPERIMENT', 'GOWORK', 'GOTOOLDIR', 'CC', 'CXX',
         'CGO_CFLAGS', 'CGO_CPPFLAGS', 'CGO_CXXFLAGS', 'CGO_FFLAGS', 'CGO_LDFLAGS'],
        text=True))
    if settings['GOWORK'] not in ['', 'off']:
        workspace = Path(settings['GOWORK'])
        manifests.update([workspace, Path(str(workspace) + '.sum')])
    for name in ['compile', 'link']:
        path = Path(settings['GOTOOLDIR']) / name
        manifests.add(path)
    manifest_hashes = [(str(path), hashlib.sha256(path.read_bytes()).hexdigest()
                        if path.exists() else None) for path in sorted(manifests)]
    version = subprocess.check_output(['go', 'version'], text=True)
    key = [sorted(inputs), manifest_hashes, settings, version, str(repository), '-buildvcs=false']
    return hashlib.sha256(json.dumps(key, sort_keys=True).encode()).hexdigest()


def manifest_product(store, manifest, key):
    data = json.loads(manifest.read_text())
    digest = data['sha256']
    if data['action'] != key or len(digest) != 64 or any(c not in '0123456789abcdef' for c in digest):
        raise RuntimeError('invalid oracle build manifest')
    return store / digest


def fetch(repository, store, prepare):
    store.mkdir(parents=True, exist_ok=True)
    key = action_key(repository)
    with (store / (key + '.lock')).open('a') as lock:
        fcntl.flock(lock, fcntl.LOCK_EX)
        manifest = store / (key + '.json')
        product = manifest_product(store, manifest, key) if manifest.exists() else None
        if product is None or not product.is_file():
            if not prepare:
                raise RuntimeError('oracle hook not prepared; run python3 stage3/fixtures/build-hook.py --prepare')
            with tempfile.TemporaryDirectory(dir=store) as scratch:
                binary = Path(scratch) / 'oracle.test'
                with (store / (key + '.build.log')).open('w') as log:
                    result = subprocess.run(
                        ['go', 'test', '-buildvcs=false', '-c', '-o', str(binary), './internal/oracle'],
                        cwd=repository, stdout=log, stderr=subprocess.STDOUT)
                if result.returncode:
                    raise RuntimeError('oracle hook preparation failed (exit %d):\n%s' % (
                        result.returncode, (store / (key + '.build.log')).read_text()))
                if action_key(repository) != key:
                    raise RuntimeError('oracle inputs changed during preparation')
                digest = hashlib.sha256(binary.read_bytes()).hexdigest()
                product = store / digest
                os.replace(binary, product)
                temporary = Path(scratch) / 'manifest.json'
                temporary.write_text(json.dumps({'action': key, 'sha256': digest}) + '\n')
                os.replace(temporary, manifest)
        product = manifest_product(store, manifest, key)
        if hashlib.sha256(product.read_bytes()).hexdigest() != product.name:
            raise RuntimeError('oracle build product hash mismatch')
        return product


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--prepare', action='store_true')
    parser.add_argument('--store', type=Path, default=Path(os.environ.get(
        'ADAMIC_STAGE3_BUILD_STORE', str(Path.home() / '.cache/adamic-stage3/build-products'))))
    args = parser.parse_args()
    print(fetch(Path(__file__).resolve().parents[2], args.store.resolve(), args.prepare))

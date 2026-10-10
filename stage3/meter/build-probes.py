#!/usr/bin/env python3
"""Prepare stage 3 Go tests and meter/census executables once; fetch only verified hash products."""
import argparse
import fcntl
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[2]
STORE = Path(os.environ.get('ADAMIC_STAGE3_BUILD_STORE',
                          str(Path.home() / '.cache/adamic-stage3/build-products'))) / 'probes'
spec = importlib.util.spec_from_file_location('hook_products', ROOT / 'stage3/fixtures/build-hook.py')
hook = importlib.util.module_from_spec(spec)
spec.loader.exec_module(hook)


def key():
    # Include all local Go and overlay sources in addition to the compiler's
    # effective dependencies, module hashes, toolchain contents and flags.
    paths = set(ROOT.rglob('*.go')) | set((ROOT / 'stage3/census').rglob('*.txt'))
    paths.update([Path(__file__), ROOT / 'stage3/census/latent/make_overlay.py',
                  ROOT / 'stage3/meter/entry_overlay.py'])
    files = [(str(p.relative_to(ROOT)), hashlib.sha256(p.read_bytes()).hexdigest())
             for p in sorted(paths) if '.git' not in p.parts and 'node_modules' not in p.parts]
    return hashlib.sha256(json.dumps([hook.action_key(ROOT), files]).encode()).hexdigest()


def fetch(prepare=False):
    if prepare:
        hook.fetch(ROOT, STORE.parent, True)
    STORE.mkdir(parents=True, exist_ok=True)
    action = key()
    with (STORE / (action + '.lock')).open('a') as lock:
        fcntl.flock(lock, fcntl.LOCK_EX)
        manifest = STORE / (action + '.json')
        if not manifest.exists():
            if not prepare:
                raise RuntimeError('probes not prepared; run python3 stage3/meter/build-probes.py --prepare')
            with tempfile.TemporaryDirectory(dir=STORE) as directory:
                scratch = Path(directory)
                overlay, entry = scratch / 'overlay', scratch / 'entry-overlay'
                commands = [ ['python3', 'stage3/census/latent/make_overlay.py', str(ROOT), str(overlay)],
                             ['python3', 'stage3/meter/entry_overlay.py', str(overlay), str(entry)] ]
                products = [('census', './stage3/census/tool', None),
                            ('latent', './stage3/census/latent/tool', overlay),
                            ('entry', './stage3/census/latent/tool', entry),
                            ('replay', './stage3/census/latent/replay/worker', overlay)]
                test_packages = [('fixtures-test', './stage3/fixtures'),
                                 ('refusalrewrite-test', './stage3/census/latent/refusalrewrite'),
                                 ('statecopy-test', './stage3/census/latent/statecopy'),
                                 ('statementrewrite-test', './stage3/census/latent/statementrewrite')]
                for name, package, source in products:
                    command = ['go', 'build', '-buildvcs=false', '-o', str(scratch / name)]
                    if source: command += ['-overlay=' + str(source / 'overlay.json')]
                    commands.append(command + [package])
                for name, package in test_packages:
                    commands.append(['go', 'test', '-buildvcs=false', '-c', '-o',
                                     str(scratch / name), package])
                products += [(name, package, None) for name, package in test_packages]
                with (STORE / (action + '.build.log')).open('w') as log:
                    for command in commands:
                        subprocess.run(command, cwd=ROOT, stdout=log, stderr=subprocess.STDOUT, check=True)
                if key() != action:
                    raise RuntimeError('probe inputs changed during preparation')
                hashes = {}
                for name, _, _ in products:
                    binary = scratch / name
                    digest = hashlib.sha256(binary.read_bytes()).hexdigest()
                    os.replace(binary, STORE / digest)
                    hashes[name] = digest
                temporary = scratch / 'manifest'
                temporary.write_text(json.dumps(dict(action=action, products=hashes)) + '\n')
                os.replace(temporary, manifest)
        data = json.loads(manifest.read_text())
        if data['action'] != action or set(data['products']) != {'census', 'latent', 'entry', 'replay',
                'fixtures-test', 'refusalrewrite-test', 'statecopy-test', 'statementrewrite-test'}:
            raise RuntimeError('invalid probe manifest')
        paths = {}
        for name, digest in data['products'].items():
            if len(digest) != 64 or any(c not in '0123456789abcdef' for c in digest):
                raise RuntimeError('invalid probe hash')
            product = STORE / digest
            if hashlib.sha256(product.read_bytes()).hexdigest() != digest:
                raise RuntimeError('probe product hash mismatch')
            paths[name] = str(product)
        return paths


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--prepare', action='store_true')
    print(json.dumps(fetch(parser.parse_args().prepare), sort_keys=True))

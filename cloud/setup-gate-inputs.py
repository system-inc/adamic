#!/usr/bin/env python3
"""Opt-in gate corpora and the checker archive, validated against their exact inputs."""
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile
import time

sys.dont_write_bytecode = True
SOURCE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location('gate_npm', SOURCE / 'setup-gate-npm.py')
npm = importlib.util.module_from_spec(spec)
spec.loader.exec_module(npm)
TS_COMMIT = '050880ce59e30b356b686bd3144efe24f875ebc8'
TS_URL = 'https://github.com/microsoft/TypeScript.git'
LEDGER_SCRIPT = 'scripts/processDiagnosticMessages.mjs'
LEDGER_DIAGNOSTICS = 'src/compiler/diagnosticMessages.json'
LEDGER_GENERATED = ['src/compiler/diagnosticInformationMap.generated.ts',
                    'src/compiler/diagnosticMessages.generated.json']
CSS_COMMIT = 'cb4b33fba24a8428d00e54be85fc886288a374ea'
CSS_URL = 'https://github.com/system-inc/prettier.git'
CSS_SPARSE = ['tests/format/css', 'tests/format/scss', 'tests/format/less',
              'tests/format/js/multiparser-css', 'tests/format/js/template-literals']
CSS_COUNTS = {'.css': 157, '.scss': 90, '.less': 43}
PRETTIER_VERSION = '3.9.6'
SIZE = 100 << 20
VARIABLES = {'ADAMIC_TYPESCRIPT_SOURCE': 'typescript',
             'ADAMIC_CYCLE_LEDGER_ROOT': 'cycle-ledger',
             'ADAMIC_CYCLE_LEDGER_OUTPUT': 'cycle-ledger-output.json', 'ADAMIC_CSS_FIXTURES': 'css-fixtures',
             'ADAMIC_CSSNUMBERS_LIBRARY': 'css-printer', 'ADAMIC_CSSSTRINGS_LIBRARY': 'css-printer',
             'ADAMIC_MARKDOWNINLINE_LIBRARY': 'css-printer/node_modules/prettier',
             'ADAMIC_CSS_LIBRARY': 'css',
             'ADAMIC_GRAPHQL_LIBRARY': 'graphql', 'ADAMIC_MEDIA_QUERY_LIBRARY': 'media-query',
             'ADAMIC_SELECTOR_LIBRARY': 'selector', 'ADAMIC_VALUES_LIBRARY': 'values',
             'ADAMIC_GRAPHQL_PRETTIER': 'css-printer', 'ADAMIC_JSON_PRETTIER': 'json-prettier', 'ADAMIC_CSS_PRINTER_LIBRARY': 'css-printer',
             'ADAMIC_ESTREE_LIBRARY': 'css-printer', 'ADAMIC_TS_PRETTIER': 'css-printer', 'ADAMIC_YAML_LIBRARY': 'css-printer',
             'ADAMIC_GITIGNORE_LARGEST': 'gitignore/.gitignore',
             'ADAMIC_CLANG_TSGO_ARCHIVE': 'checker/tsgo.a'}
ARCHIVE_FLAGS = ['-trimpath', '-buildvcs=false', '-buildmode=c-archive']
# Go list with c-archive emits a stray .h in its working directory. Validate
# the same source closure in exe mode; stamp both flag sets for actual linking.
VALIDATION_FLAGS = ['-trimpath', '-buildvcs=false', '-buildmode=exe']


def command(args, cwd=None):
    return subprocess.check_output(args, cwd=cwd, timeout=600).decode().strip()


def cache_key(**inputs):
    return npm.digest(json.dumps(inputs, sort_keys=True).encode())


def source_hash():
    return npm.digest(Path(__file__).read_bytes() + (SOURCE / 'setup-gate-npm.py').read_bytes())


def file_hash(path):
    answer = hashlib.sha256()
    with path.open('rb') as stream:
        for chunk in iter(lambda: stream.read(1 << 20), b''):
            answer.update(chunk)
    return answer.hexdigest()


def artifact_digest(directory):
    entries = []
    for path in sorted(directory.rglob('*')):
        relative = path.relative_to(directory)
        if '.git' in relative.parts or path == directory / '.adamic-stamp':
            continue
        mode = path.lstat().st_mode & 0o777
        if path.is_symlink():
            raise ValueError('unexpected corpus symlink: ' + str(relative))
        if path.is_file():
            entries.append([str(relative), mode, file_hash(path)])
        elif path.is_dir():
            entries.append([str(relative), mode, 'directory'])
        else:
            raise ValueError('unexpected corpus entry: ' + str(relative))
    return cache_key(root_mode=directory.stat().st_mode & 0o777, entries=entries)


def hit(destination, key):
    if os.environ.get('ADAMIC_GATE_UNCACHED') == '1':
        return False
    try:
        saved = json.loads((destination / '.adamic-stamp').read_text())
        return saved['key'] == key and saved['tree'] == artifact_digest(destination)
    except (OSError, ValueError, KeyError):
        return False


def publish(install, destination, key):
    install.chmod(0o755)
    (install / '.adamic-stamp').write_text(json.dumps(dict(key=key, tree=artifact_digest(install))))
    backup = install.parent / 'previous'
    if destination.exists():
        destination.rename(backup)
    try:
        install.rename(destination)
    except BaseException:
        if backup.exists():
            backup.rename(destination)
        raise


def typescript(root):
    destination = root / 'typescript'
    key = cache_key(kind='typescript', commit=TS_COMMIT, url=TS_URL, helper=source_hash())
    if hit(destination, key) and command(['git', '-C', str(destination), 'rev-parse', 'HEAD']) == TS_COMMIT:
        return 'skipped (validated exact commit and checkout bytes)'
    with tempfile.TemporaryDirectory(prefix='typescript-', dir=root) as temporary:
        install = Path(temporary) / 'checkout'
        subprocess.run(['git', 'init', str(install)], check=True, timeout=30)
        subprocess.run(['git', '-C', str(install), 'fetch', '--depth=1', TS_URL, TS_COMMIT], check=True, timeout=600)
        subprocess.run(['git', '-C', str(install), 'checkout', '--detach', 'FETCH_HEAD'], check=True, timeout=60)
        if command(['git', '-C', str(install), 'rev-parse', 'HEAD']) != TS_COMMIT:
            raise ValueError('TypeScript commit integrity mismatch')
        subprocess.run(['git', '-C', str(install), 'fsck', '--full', '--no-reflogs'], check=True, timeout=120)
        publish(install, destination, key)
    return 'installed (depth-one exact commit, git object integrity checked)'



def ledger_key(checkout, node_version):
    return cache_key(kind='cycle-ledger', commit=TS_COMMIT, url=TS_URL,
                     script=file_hash(checkout / LEDGER_SCRIPT),
                     diagnostics=file_hash(checkout / LEDGER_DIAGNOSTICS),
                     node=node_version, helper=source_hash())


def validate_ledger(checkout):
    actual = command(['git', '-C', str(checkout), 'rev-parse', 'HEAD'])
    if actual != TS_COMMIT:
        raise ValueError(f'cycle ledger source pin: expected {TS_COMMIT}, got {actual}')
    # Generated diagnostics and our stamp are ignored/untracked. Tracked source
    # must remain upstream's original bytes, not the parser's adapted corpus.
    subprocess.run(['git', '-C', str(checkout), 'diff', '--exit-code', 'HEAD', '--'],
                   check=True, timeout=60)
    for name in LEDGER_GENERATED:
        file = checkout / name
        if not file.is_file() or file.is_symlink() or not file.stat().st_size:
            raise ValueError('cycle ledger missing generated diagnostics: ' + name)


def cycle_ledger(root, node):
    node_version = command([node, '--version'])
    expected = json.loads((SOURCE / 'node-pin.json').read_text())['version']
    if node_version != expected:
        raise ValueError(f'cycle ledger Node pin: expected {expected}, got {node_version}')
    destination = root / 'cycle-ledger'
    output = root / VARIABLES['ADAMIC_CYCLE_LEDGER_OUTPUT']
    # Mutable proof output must never invalidate the immutable input stamp.
    if output.is_symlink():
        raise ValueError('cycle ledger output must not be a symlink: ' + str(output))
    with output.open('ab'):
        pass
    if not output.is_file() or not os.access(output, os.W_OK):
        raise ValueError('cycle ledger output is not writable: ' + str(output))
    try:
        key = ledger_key(destination, node_version)
    except OSError:
        key = None
    if key is not None and hit(destination, key):
        validate_ledger(destination)
        return 'skipped (validated pristine commit, generator, Node and diagnostic bytes)'
    with tempfile.TemporaryDirectory(prefix='cycle-ledger-', dir=root) as temporary:
        install = Path(temporary) / 'checkout'
        subprocess.run(['git', 'init', str(install)], check=True, timeout=30)
        subprocess.run(['git', '-C', str(install), 'fetch', '--depth=1', TS_URL, TS_COMMIT],
                       check=True, timeout=600)
        subprocess.run(['git', '-C', str(install), 'checkout', '--detach', 'FETCH_HEAD'],
                       check=True, timeout=60)
        subprocess.run(['git', '-C', str(install), 'fsck', '--full', '--no-reflogs'],
                       check=True, timeout=120)
        key = ledger_key(install, node_version)
        # Relative input keeps the generated comment identical across installs.
        subprocess.run([node, LEDGER_SCRIPT, LEDGER_DIAGNOSTICS], cwd=install,
                       check=True, timeout=120)
        validate_ledger(install)
        if ledger_key(install, command([node, '--version'])) != key:
            raise ValueError('cycle ledger generator inputs changed; rerun setup')
        publish(install, destination, key)
    return 'installed (pristine exact commit and both generated diagnostics verified)'


def css_fixture_key():
    return cache_key(kind='css-fixtures', commit=CSS_COMMIT, url=CSS_URL,
                     sparse=CSS_SPARSE, counts=CSS_COUNTS, helper=source_hash())


def css_counts(directory):
    counts = {extension: 0 for extension in CSS_COUNTS}
    for path in directory.rglob('*'):
        if '.git' not in path.relative_to(directory).parts and path.is_file():
            extension = path.suffix.lower()
            if extension in counts:
                counts[extension] += 1
    if counts != CSS_COUNTS:
        raise ValueError(f'CSS fixture counts differ: expected {CSS_COUNTS}, got {counts}')
    return counts


def validate_css_checkout(directory):
    actual = command(['git', '-C', str(directory), 'rev-parse', 'HEAD'])
    if actual != CSS_COMMIT:
        raise ValueError(f'CSS fixture commit integrity mismatch: expected {CSS_COMMIT}, got {actual}')
    sparse = command(['git', '-C', str(directory), 'sparse-checkout', 'list']).splitlines()
    if sorted(sparse) != sorted(CSS_SPARSE):
        raise ValueError(f'CSS sparse checkout differs: expected {CSS_SPARSE}, got {sparse}')
    return css_counts(directory)


def css_fixtures(root):
    destination = root / 'css-fixtures'
    key = css_fixture_key()
    if hit(destination, key):
        # Git metadata is excluded from the content stamp, so verify it explicitly.
        validate_css_checkout(destination)
        return 'skipped (validated exact commit, sparse set, counts and checkout bytes)'
    with tempfile.TemporaryDirectory(prefix='css-fixtures-', dir=root) as temporary:
        install = Path(temporary) / 'checkout'
        subprocess.run(['git', 'init', str(install)], check=True, timeout=30)
        subprocess.run(['git', '-C', str(install), 'remote', 'add', 'origin', CSS_URL], check=True, timeout=30)
        subprocess.run(['git', '-C', str(install), 'fetch', '--depth=1', '--filter=blob:none',
                        'origin', CSS_COMMIT], check=True, timeout=600)
        subprocess.run(['git', '-C', str(install), 'sparse-checkout', 'init', '--cone'], check=True, timeout=30)
        subprocess.run(['git', '-C', str(install), 'sparse-checkout', 'set', *CSS_SPARSE], check=True, timeout=60)
        subprocess.run(['git', '-C', str(install), 'checkout', '--detach', 'FETCH_HEAD'], check=True, timeout=600)
        validate_css_checkout(install)
        subprocess.run(['git', '-C', str(install), 'fsck', '--full', '--no-reflogs'], check=True, timeout=120)
        publish(install, destination, key)
    return f'installed (exact commit, sparse checkout, Git integrity and counts verified: {CSS_COUNTS})'


def shared_prettier(root, node):
    # Match both createRequire(prefix/package.json) and direct package/plugin loads.
    script = """
const {createRequire} = require('node:module');
const prefix = process.argv[1];
const expected = process.argv[2];
const resolve = createRequire(prefix + '/package.json');
for (const name of ['ADAMIC_CSSNUMBERS_LIBRARY', 'ADAMIC_CSSSTRINGS_LIBRARY']) {
    const actual = resolve('prettier/package.json').version;
    if (actual !== expected) throw Error(`${name}: expected prettier ${expected}, got ${actual}`);
    resolve('prettier/plugins/postcss');
}
const library = prefix + '/node_modules/prettier';
const actual = require(library).version;
if (actual !== expected) throw Error(`ADAMIC_MARKDOWNINLINE_LIBRARY: expected prettier ${expected}, got ${actual}`);
require(library + '/plugins/markdown');
console.log('shared prettier ' + actual + ': CSS numbers, CSS strings and Markdown inline paths verified');
"""
    return command([node, '-e', script, str(root / 'css-printer'), PRETTIER_VERSION])


def ignore_contents(size):
    yield b'big\n#'
    remaining = size - len(b'big\n#') - 1
    if remaining < 0:
        raise ValueError('ignore corpus size too small')
    while remaining:
        chunk = min(remaining, 1 << 20)
        yield b'x' * chunk
        remaining -= chunk
    yield b'\n'


def gitignore(root, size=SIZE):
    destination = root / 'gitignore'
    expected = hashlib.sha256()
    for chunk in ignore_contents(size):
        expected.update(chunk)
    key = cache_key(kind='gitignore', size=size, content=expected.hexdigest(), helper=source_hash())
    if hit(destination, key):
        return 'skipped (validated deterministic 100 MiB bytes)'
    with tempfile.TemporaryDirectory(prefix='gitignore-', dir=root) as temporary:
        install = Path(temporary) / 'corpus'
        install.mkdir(mode=0o755)
        file = install / '.gitignore'
        with file.open('wb') as output:
            for chunk in ignore_contents(size):
                output.write(chunk)
        file.chmod(0o644)
        if file.stat().st_size != size or file_hash(file) != expected.hexdigest():
            raise ValueError('generated ignore corpus integrity mismatch')
        publish(install, destination, key)
    return 'generated (deterministic 100 MiB, SHA256 verified)'


def archive_inputs(repository):
    # Go validates source, C headers, embeds, replacements and compiler flags before
    # the action IDs can replace repeated linking. Dirty sources invalidate too.
    data = command(['go', 'list', '-deps', '-export', '-json', *VALIDATION_FLAGS, './bridge/tsgo/archive'], repository)
    decoder = json.JSONDecoder()
    packages = []
    while data.strip():
        package, end = decoder.raw_decode(data.lstrip())
        data = data.lstrip()[end:]
        if package.get('Error') or package.get('DepsErrors'):
            raise ValueError('Go could not validate checker archive dependencies')
        artifact = package.get('Export')
        if artifact and not Path(artifact).is_file():
            raise ValueError('missing validated Go export')
        packages.append([package['ImportPath'], package.get('BuildID')])
    environment = json.loads(command(['go', 'env', '-json'], repository))
    environment['GOGCCFLAGS'] = re.sub(r'-f(debug|file)-prefix-map=\S*/go-build\d+=/tmp/go-build',
                                     r'-f\1-prefix-map=<temporary>=/tmp/go-build', environment['GOGCCFLAGS'])
    cc = environment['CC']
    return dict(kind='checker', head=command(['git', 'rev-parse', 'HEAD'], repository),
                packages=sorted(packages), environment=environment, version=command(['go', 'version'], repository),
                cc_version=command([cc, '--version']).splitlines()[0], flags=ARCHIVE_FLAGS, validation_flags=VALIDATION_FLAGS, helper=source_hash())


def archive(repository, root):
    destination = root / 'checker'
    key = cache_key(**archive_inputs(repository))
    if hit(destination, key):
        return 'skipped (validated Go actions and archive bytes)'
    with tempfile.TemporaryDirectory(prefix='checker-', dir=root) as temporary:
        install = Path(temporary) / 'archive'
        install.mkdir(mode=0o755)
        subprocess.run(['go', 'build', *ARCHIVE_FLAGS, '-o', str(install / 'tsgo.a'), './bridge/tsgo/archive'],
                       cwd=repository, check=True, timeout=600)
        for file in install.iterdir():
            file.chmod(0o644)
        if not (install / 'tsgo.a').read_bytes().startswith(b'!<arch>\n'):
            raise ValueError('checker build did not produce a C archive')
        if cache_key(**archive_inputs(repository)) != key:
            raise ValueError('checker inputs changed during build; rerun setup')
        publish(install, destination, key)
    return 'built (Go content-addressed inputs, archive bytes stamped)'


def main():
    phase, repository, root, node = sys.argv[1:]
    root = Path(root)
    if root.resolve().is_relative_to('/root'):
        raise ValueError('gate inputs must be outside /root')
    root.mkdir(parents=True, exist_ok=True, mode=0o755)
    if phase in {'env', 'env-no-archive', 'env-archive'}:
        for name, relative in VARIABLES.items():
            archive_input = name == 'ADAMIC_CLANG_TSGO_ARCHIVE'
            selected = phase == 'env' or (phase == 'env-archive') == archive_input
            if selected:
                print('export ' + name + '=' + __import__('shlex').quote(str(root / relative)))
            else:
                print('unset ' + name)
        return
    tasks = {'corpora': [('CSS fixtures', lambda: css_fixtures(root)), ('TypeScript source', lambda: typescript(root)), ('100 MiB gitignore', lambda: gitignore(root))],
             'archive': [('checker archive', lambda: archive(repository, root))],
             'npm': [(name, lambda name=name: npm.prepare(SOURCE / 'gate-inputs' / name, root / name, node))
                     for name in dict.fromkeys(VARIABLES.values()) if (SOURCE / 'gate-inputs' / name).is_dir()]}
    if phase == 'npm':
        tasks['npm'].append(('shared Prettier paths', lambda: shared_prettier(root, node)))
        # This parallel preparation lane runs only after pinned Node is ready.
        tasks['npm'].append(('cycle ledger', lambda: cycle_ledger(root, node)))
    for name, action in tasks[phase]:
        started = time.monotonic()
        answer = action()
        print(f'setup: gate input {name} {answer}; step-duration={time.monotonic() - started:.3f}s', flush=True)


if __name__ == '__main__':
    main()

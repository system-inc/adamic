#!/usr/bin/env python3
"""Download the gate package closure; recursive module graphs are explicit opt-in."""
import base64
import hashlib
import json
import os
from pathlib import Path
import stat
import subprocess
import sys
import tempfile
import time
import zipfile


def digest(value):
    return hashlib.sha256(json.dumps(value, sort_keys=True).encode()).hexdigest()


def manifests(repository):
    modules, files = set(), set()
    for directory, children, names in os.walk(repository):
        children[:] = [name for name in children if name not in {'.git', 'node_modules'}]
        root = Path(directory)
        if 'go.mod' in names:
            modules.add(root)
            files.update([root / 'go.mod', root / 'go.sum'])
        if 'go.work' in names:
            files.update([root / 'go.work', root / 'go.work.sum'])
    work = os.environ.get('GOWORK')
    if work and work != 'off':
        files.update([Path(work), Path(work + '.sum')])
    return sorted(modules), [[str(path), hashlib.sha256(path.read_bytes()).hexdigest() if path.exists() else None]
                             for path in sorted(files)]


def cache_state(paths):
    # ctime catches content edits even when mtime is restored; include all extracted
    # files, not merely a .ziphash marker. A removal or chmod invalidates too.
    entries = []
    for path in sorted(set(Path(path) for path in paths)):
        selected = [path] + sorted(path.rglob('*')) if path.is_dir() else [path]
        for file in selected:
            info = file.lstat()
            if stat.S_ISLNK(info.st_mode):
                raise ValueError('unexpected Go module cache symlink: ' + str(file))
            entries.append([str(file), info.st_mode, info.st_size, info.st_mtime_ns, info.st_ctime_ns, info.st_ino])
    return digest(entries)


def records(data):
    decoder = json.JSONDecoder()
    while data.strip():
        value, end = decoder.raw_decode(data.lstrip())
        data = data.lstrip()[end:]
        yield value


def key(manifests, version, environment, helper):
    return digest(dict(manifests=manifests, version=version, environment=environment, helper=helper))


TARGETS = Path(__file__).with_name('cohere-module-targets.json')
ENVIRONMENT = ['GOMODCACHE', 'GOPROXY', 'GOSUMDB', 'GONOSUMDB', 'GONOPROXY', 'GOPRIVATE',
               'GOWORK', 'GOTOOLCHAIN', 'GOFLAGS', 'GOOS', 'GOARCH', 'CGO_ENABLED', 'GOEXPERIMENT']


def go(command, directory):
    result = subprocess.run(['go', *command], cwd=directory, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    # A denied proxy redirect must not strand the fleet. Direct mode retains
    # Go's go.sum/sumdb checks and the same exact requested versions.
    problem = (result.stdout + result.stderr).decode(errors='replace').lower()
    if result.returncode and os.environ.get('GOPROXY') != 'off' and any(
            word in problem for word in ['403 forbidden', '502 bad gateway', 'proxyconnect', 'connection', 'timeout', 'denied']):
        print('setup: Go proxy failed; retrying configured proxy', flush=True)
        result = subprocess.run(['go', *command], cwd=directory, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        if result.returncode:
            print('setup: retrying direct with checksum verification', flush=True)
            result = subprocess.run(['go', *command], cwd=directory, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                                    env=dict(os.environ, GOPROXY='direct'))
    if result.stderr:
        sys.stderr.write(result.stderr.decode(errors='replace'))
    if result.returncode:
        raise RuntimeError('go ' + ' '.join(command) + ': ' + result.stdout.decode(errors='replace'))
    return result.stdout.decode()


def gate_closure(repository):
    catalog = json.loads(TARGETS.read_text())
    groups = [(repository, ['./...'])]
    if (repository / 'cohere/go.mod').is_file():
        groups.append((repository / 'cohere', catalog['packages']))
    modules, paths, contexts, imported, definitions = {}, set(), [], {}, {}
    for directory, targets in groups:
        environment = json.loads(go(['env', '-json', *ENVIRONMENT], directory))
        contexts.append([str(directory), environment])
        paths.update([directory / 'go.mod', directory / 'go.sum'])
        workspace = environment.get('GOWORK')
        if workspace and workspace != 'off':
            paths.update([Path(workspace), Path(workspace + '.sum')])
        for package in records(go(['list', '-deps', '-test', '-json', *targets], directory)):
            if package.get('Error') or package.get('DepsErrors'):
                raise RuntimeError('Go could not resolve the gate package closure')
            module = package.get('Module', {})
            for item in [module, module.get('Replace', {})]:
                if item.get('GoMod'):
                    file = Path(item['GoMod'])
                    # External cached manifests are validated as artifacts below.
                    if file.is_relative_to(repository):
                        paths.update([file, file.with_name('go.sum')])
                    definitions[str(file)] = directory
            selected = module.get('Replace', module)
            if package.get('ImportPath') and selected.get('Path'):
                imported[package['ImportPath'].split(' [')[0]] = selected['Path']
            if selected.get('Version'):
                identity = selected['Path'] + '@' + selected['Version']
                modules[identity] = str(directory)
    # Go checks shorter module prefixes for ambiguous imports. The local shims' unused
    # regexp2 v1 requirement must be present to resolve regexp2/v2 without a
    # download line in the lint oracle, although no compiled package uses v1.
    for definition, directory in definitions.items():
        requirements = json.loads(go(['mod', 'edit', '-json', definition], directory)).get('Require', [])
        for requirement in requirements:
            candidate = requirement['Path']
            if any(name.startswith(candidate + '/') and owner.startswith(candidate + '/')
                   for name, owner in imported.items()):
                modules[candidate + '@' + requirement['Version']] = str(directory)
    paths.add(TARGETS)
    return modules, paths, contexts


def manifest_inputs(paths):
    return [[str(path), hashlib.sha256(path.read_bytes()).hexdigest() if path.exists() else None]
            for path in sorted(paths)]


def content_sum(files):
    # Go's sumdb/dirhash.Hash1: SHA256 of sorted filename and content-hash lines.
    result = hashlib.sha256()
    for name, contents in sorted(files):
        result.update((hashlib.sha256(contents).hexdigest() + '  ' + name + '\n').encode())
    return 'h1:' + base64.b64encode(result.digest()).decode()


def verify_selected(record):
    archive, directory = Path(record['Zip']), Path(record['Dir'])
    with zipfile.ZipFile(archive) as stream:
        zipped = content_sum([(name, stream.read(name)) for name in stream.namelist()])
    prefix = record['Path'] + '@' + record['Version'] + '/'
    files = []
    for file in directory.rglob('*'):
        if file.is_symlink():
            raise ValueError('unexpected module cache symlink: ' + str(file))
        if file.is_file():
            files.append((prefix + file.relative_to(directory).as_posix(), file.read_bytes()))
    if zipped != record['Sum'] or content_sum(files) != record['Sum']:
        raise ValueError('module archive/extracted checksum mismatch: ' + record['Path'])
    if content_sum([('go.mod', Path(record['GoMod']).read_bytes())]) != record['GoModSum']:
        raise ValueError('module go.mod checksum mismatch: ' + record['Path'])


def prepare(repository, tools, all_modules=False):
    if all_modules:
        return prepare_all(repository, tools)
    repository, tools = Path(repository).resolve(), Path(tools)
    _, preliminary = manifests(repository)
    preliminary.append([str(TARGETS), hashlib.sha256(TARGETS.read_bytes()).hexdigest()])
    selected, files, contexts = gate_closure(repository)
    inputs = manifest_inputs(files)
    old = dict(preliminary)
    if any(old.get(name) != value for name, value in inputs if name in old and not name.endswith('.sum')):
        raise RuntimeError('gate module/workspace definitions changed during resolution; rerun setup')
    version = go(['version'], repository).strip()
    helper = hashlib.sha256(Path(__file__).read_bytes()).hexdigest()
    environment = dict(contexts=contexts, closure=sorted(selected))
    current = key(inputs, version, environment, helper)
    tools.mkdir(parents=True, exist_ok=True)
    stamp = tools / ('modules-closure-' + digest(str(repository)))
    if os.environ.get('ADAMIC_GATE_UNCACHED') != '1':
        try:
            saved = json.loads(stamp.read_text())
            if saved['key'] == current and saved['cache'] == cache_state(saved['paths']):
                return f'skipped (validated gate closure: {len(selected)} downloaded modules)'
        except (OSError, ValueError, KeyError):
            pass
    paths = set()
    for directory in sorted(set(selected.values())):
        identifiers = sorted(name for name, cwd in selected.items() if cwd == directory)
        for record in records(go(['mod', 'download', '-json', *identifiers], directory)):
            if record.get('Error'):
                raise RuntimeError(record['Error'])
            verify_selected(record)
            for field in ['Dir', 'Zip', 'GoMod', 'Info']:
                paths.add(record[field])
            paths.add(record['Zip'] + 'hash')
    completed_inputs = manifest_inputs(files)
    definitions = lambda values: [item for item in values if not item[0].endswith('.sum')]
    if definitions(inputs) != definitions(completed_inputs):
        raise RuntimeError('gate module/workspace definitions changed during download; rerun setup')
    completed = dict(key=key(completed_inputs, version, environment, helper), paths=sorted(paths),
                     cache=cache_state(paths), modules=sorted(selected))
    with tempfile.NamedTemporaryFile(mode='w', dir=tools, delete=False) as file:
        json.dump(completed, file)
        temporary = Path(file.name)
    temporary.replace(stamp)
    return f'downloaded and verified gate closure ({len(selected)} downloaded modules)'


def prepare_all(repository, tools):
    repository, tools = Path(repository).resolve(), Path(tools)
    modules, inputs = manifests(repository)
    environment = json.loads(subprocess.check_output(['go', 'env', '-json', 'GOMODCACHE', 'GOPROXY', 'GOSUMDB',
                               'GONOSUMDB', 'GONOPROXY', 'GOPRIVATE', 'GOWORK', 'GOTOOLCHAIN', 'GOFLAGS'], cwd=repository).decode())
    version = subprocess.check_output(['go', 'version'], cwd=repository).decode().strip()
    helper = hashlib.sha256(Path(__file__).read_bytes()).hexdigest()
    current = key(inputs, version, environment, helper)
    tools.mkdir(parents=True, exist_ok=True)
    stamp = tools / ('modules-' + digest(str(repository)))
    if os.environ.get('ADAMIC_GATE_UNCACHED') != '1':
        try:
            saved = json.loads(stamp.read_text())
            if saved['key'] == current and saved['cache'] == cache_state(saved['paths']):
                return 'skipped (validated manifests and downloaded module cache)'
        except (OSError, ValueError, KeyError):
            pass
    paths = set()
    for module in modules:
        print('setup: downloading module dependencies in ' + str(module), flush=True)
        answer = go(['mod', 'download', '-json'], module)
        for record in records(answer):
            if record.get('Error'):
                raise RuntimeError(record['Error'])
            for field in ['Dir', 'Zip', 'GoMod', 'Info']:
                if record.get(field):
                    paths.add(record[field])
            if record.get('Zip'):
                paths.add(record['Zip'] + 'hash')
        # Download checks go.sum/sumdb; verify also rejects modified extracted bytes.
        print(go(['mod', 'verify'], module), flush=True)
    completed_modules, completed_inputs = manifests(repository)
    # Go may complete sum files; changes to module/workspace definitions require a retry.
    definitions = lambda values: [item for item in values if not item[0].endswith('.sum')]
    if modules != completed_modules or definitions(inputs) != definitions(completed_inputs):
        raise RuntimeError('module/workspace definitions changed during download; rerun setup')
    inputs = completed_inputs
    completed = dict(key=key(inputs, version, environment, helper), paths=sorted(paths), cache=cache_state(paths))
    with tempfile.NamedTemporaryFile(mode='w', dir=tools, delete=False) as file:
        json.dump(completed, file)
        temporary = Path(file.name)
    temporary.replace(stamp)
    return f'downloaded and verified ({len(modules)} module directories)'


if __name__ == '__main__':
    started = time.monotonic()
    arguments = sys.argv[1:]
    all_modules = '--all-modules' in arguments
    if all_modules: arguments.remove('--all-modules')
    print('setup: module dependencies ' + prepare(*arguments, all_modules=all_modules) + f'; step-duration={time.monotonic() - started:.3f}s')

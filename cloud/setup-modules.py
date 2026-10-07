#!/usr/bin/env python3
"""Warm every checkout module, including dependency graphs absent from root builds."""
import hashlib
import json
import os
from pathlib import Path
import stat
import subprocess
import sys
import tempfile
import time


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


def prepare(repository, tools):
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
    # Shim manifests rely on workspace-local modules at v0.0.0. Use private
    # modfiles with those same local resolutions, rather than modifying sources
    # or letting an inherited workspace hide the current manifest's graph.
    download_environment = dict(os.environ, GOWORK='off')
    local_modules = {}
    for module in modules:
        metadata = json.loads(subprocess.check_output(['go', 'mod', 'edit', '-json'],
                              cwd=module, env=download_environment))
        local_modules[metadata['Module']['Path']] = module
    paths = set()
    for module in modules:
        print('setup: downloading module dependencies in ' + str(module), flush=True)
        with tempfile.TemporaryDirectory(prefix='setup-module-') as temporary:
            modfile = Path(temporary) / 'download.mod'
            modfile.write_bytes((module / 'go.mod').read_bytes())
            if (module / 'go.sum').exists():
                modfile.with_suffix('.sum').write_bytes((module / 'go.sum').read_bytes())
            metadata = json.loads(subprocess.check_output(['go', 'mod', 'edit', '-json'],
                                  cwd=module, env=download_environment))
            replaced = {item['Old']['Path'] for item in metadata.get('Replace', [])}
            replacements = ['-replace=' + name + '=' + str(directory)
                            for name, directory in sorted(local_modules.items()) if name not in replaced]
            subprocess.run(['go', 'mod', 'edit', '-modfile=' + str(modfile), *replacements],
                           cwd=module, env=download_environment, check=True)
            result = subprocess.run(['go', 'mod', 'download', '-modfile=' + str(modfile), '-json', 'all'],
                                    cwd=module, env=download_environment, stdout=subprocess.PIPE)
            answer = result.stdout.decode()
            for record in records(answer):
                if record.get('Error'):
                    raise RuntimeError(record['Error'])
                for field in ['Dir', 'Zip', 'GoMod', 'Info']:
                    if record.get(field):
                        paths.add(record[field])
                if record.get('Zip'):
                    paths.add(record['Zip'] + 'hash')
            result.check_returncode()
            # Download checks the copied go.sum/sumdb; verify rejects edited bytes.
            subprocess.run(['go', 'mod', 'verify', '-modfile=' + str(modfile)],
                           cwd=module, env=download_environment, check=True)
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
    print('setup: module dependencies ' + prepare(*sys.argv[1:]) + f'; step-duration={time.monotonic() - started:.3f}s')

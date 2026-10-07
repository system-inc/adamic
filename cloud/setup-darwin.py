#!/usr/bin/env python3
"""Portable setup lane for macOS, which lacks Linux's GNU shell utilities."""
import argparse
import fcntl
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import shlex
import shutil
import subprocess
import sys
import tempfile
import time

SOURCE = Path(__file__).resolve().parent
sys.dont_write_bytecode = True
spec = importlib.util.spec_from_file_location('node_setup', SOURCE / 'setup-node.py')
node = importlib.util.module_from_spec(spec)
spec.loader.exec_module(node)


def output(command, **kwargs):
    return subprocess.run(command, check=True, capture_output=True, text=True, timeout=600, **kwargs).stdout.strip()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--warm-tests', action='store_true')
    parser.add_argument('--gate-inputs', action='store_true')
    parser.add_argument('--gate-inputs-no-archive', action='store_true')
    parser.add_argument('--gate-archive', action='store_true')
    flags = parser.parse_args()
    if flags.gate_inputs_no_archive and (flags.gate_inputs or flags.gate_archive):
        parser.error('--gate-inputs-no-archive conflicts with --gate-inputs or --gate-archive')
    gate_inputs = flags.gate_inputs or flags.gate_inputs_no_archive
    gate_archive = flags.gate_inputs or flags.gate_archive
    repository = Path(os.environ.get('ADAMIC_SETUP_REPOSITORY', SOURCE.parent))
    tools = Path(os.environ.get('ADAMIC_TOOLS', str(Path.home() / '.adamic-tools'))).resolve()
    if tools.is_relative_to('/root'):
        tools = Path('/tmp/adamic-gate') / ('tools-' + hashlib.sha256(str(tools).encode()).hexdigest())
    tools.mkdir(parents=True, exist_ok=True)
    gate = Path('/tmp/adamic-gate')
    gate.mkdir(exist_ok=True)
    gate.chmod(0o1777)
    started = time.monotonic()
    load_before = os.getloadavg()
    def step(message):
        print(f'setup: {message} ({time.monotonic() - started:.3f}s)', flush=True)
    def helper(name, *arguments):
        subprocess.run([sys.executable, str(SOURCE / name), *map(str, arguments)], check=True, timeout=1800)
    with (tools / 'setup.lock').open('w') as lock:
        lock_deadline = time.monotonic() + 600
        while True:
            try:
                fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
                break
            except BlockingIOError:
                if time.monotonic() >= lock_deadline:
                    raise TimeoutError('setup lock: deadline 600s exceeded')
                time.sleep(0.05)
        step(node.prepare(tools))
        os.environ['PATH'] = str(tools / 'bin') + ':' + str(tools / 'go/bin') + ':' + os.environ['PATH']
        os.environ.update(GOTOOLCHAIN='auto', TMPDIR=str(gate))
        # Xcode supplies clang and leaks. Preserve any installed Go with toolchain auto-selection.
        if shutil.which('go') is None:
            version = output(['curl', '-fsSL', '--max-time', '600', 'https://go.dev/VERSION?m=text']).splitlines()[0]
            architecture = 'amd64' if node.asset()[1] == 'x64' else 'arm64'
            with tempfile.TemporaryDirectory(dir=gate) as temporary:
                archive = Path(temporary) / 'go.tar.gz'
                node.download(f'https://dl.google.com/go/{version}.darwin-{architecture}.tar.gz', archive)
                subprocess.run(['tar', '-xzf', str(archive), '-C', str(tools)], check=True, timeout=600)
        go_version = output(['go', 'version'], cwd=repository)
        clang_version = output(['clang', '--version']).splitlines()[0]
        step('go and Xcode clang ready')
        subprocess.run(['git', '-C', str(repository), 'config', 'submodule.cohere.url', 'https://github.com/system-inc/cohere.git'], check=True, timeout=30)
        subprocess.run(['git', '-C', str(repository), 'submodule', 'update', '--init', '--recursive', '--depth', '1', '--filter=blob:none'], check=True, timeout=600)
        helper('setup-modules.py', repository, tools)
        helper('setup-markdown-width.py', SOURCE / 'markdown-width', tools / 'markdown-width', tools / 'bin/node')
        helper('setup-stage3-api.py', repository, tools, tools / 'bin/node')
        inputs = tools / 'gate-inputs'
        if gate_inputs:
            for phase in ['npm', 'corpora']:
                helper('setup-gate-inputs.py', phase, repository, inputs, tools / 'bin/node')
        if gate_archive:
            helper('setup-gate-inputs.py', 'archive', repository, inputs, tools / 'bin/node')
        environment = ('export PATH=' + shlex.quote(str(tools / 'bin') + ':' + str(tools / 'go/bin')) + ':"$PATH"\n'
                       'export GOTOOLCHAIN=auto\nexport TMPDIR=' + shlex.quote(str(gate)) + '\n'
                       'export ADAMIC_MARKDOWNWIDTH_DEPS=' + shlex.quote(str(tools / 'markdown-width')) + '\n')
        if gate_inputs or gate_archive:
            phase = 'env' if gate_inputs and gate_archive else 'env-no-archive' if gate_inputs else 'env-archive'
            environment += output([sys.executable, str(SOURCE / 'setup-gate-inputs.py'), phase, str(repository), str(inputs), str(tools / 'bin/node')]) + '\n'
        else:
            # Use the same variable catalog as the Linux installer; never carry a prior opt-in seat.
            spec = importlib.util.spec_from_file_location('gate_inputs', SOURCE / 'setup-gate-inputs.py')
            catalog = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(catalog)
            environment += 'unset ' + ' '.join(catalog.VARIABLES) + '\n'
        (tools / 'env.sh').write_text(environment)
        with tempfile.TemporaryDirectory(dir=gate) as temporary:
            packages = Path(temporary) / 'packages.json'
            uncached = os.environ.get('ADAMIC_GATE_UNCACHED') == '1'
            key = ''
            stamp = tools / ('warm-' + hashlib.sha256(str(repository).encode()).hexdigest() + '-' + str(flags.warm_tests).lower())
            if not uncached:
                with packages.open('wb') as log:
                    subprocess.run(['go', 'list', '-deps', '-export', *(['-test'] if flags.warm_tests else []), '-json', './...'], cwd=repository, stdout=log, check=True, timeout=600)
                key = output([sys.executable, str(SOURCE / 'setup-key.py'), str(repository), str(packages), str(flags.warm_tests).lower()])
            if key and stamp.exists() and stamp.read_text().strip() == key:
                step('go build skipped (validated warming stamp)')
            else:
                subprocess.run(['go', 'build', *(['-a'] if uncached else []), './...'], cwd=repository, check=True, timeout=600)
                if flags.warm_tests:
                    subprocess.run(['go', 'test', *(['-a'] if uncached else []), '-count=1', '-run', '^$', './...'], cwd=repository, check=True, timeout=600)
                if key:
                    completed = output([sys.executable, str(SOURCE / 'setup-key.py'), str(repository), str(packages), str(flags.warm_tests).lower()])
                    temporary_stamp = Path(temporary) / 'stamp'
                    temporary_stamp.write_text(completed + '\n')
                    os.replace(temporary_stamp, stamp)
                step('build cache warm')
        actual = output([str(tools / 'bin/node'), '--version'])
        expected = json.loads(node.PIN.read_text())['version']
        if actual != expected:
            raise ValueError(f'node: got {actual}, want {expected}')
        print(f'setup: build-flags commit={output(["git", "-C", str(repository), "rev-parse", "HEAD"])} nproc={os.cpu_count()} cpu.max=unavailable go={go_version} clang={clang_version} node={actual} cached={"no" if uncached else "yes"} warm-tests={flags.warm_tests} gate-inputs={gate_inputs} gate-archive={gate_archive} load-before={load_before} load-after={os.getloadavg()}')
        step('done on macOS')
        print(f'setup: source {tools / "env.sh"}')
        print('setup: add the source command to your shell configuration')
        print(f'setup: node {actual}')


if __name__ == '__main__':
    main()

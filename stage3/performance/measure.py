#!/usr/bin/env python3
"""Pinned, fresh-process checker baseline. No timing before output agreement."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import random
import re
import shutil
import statistics
import subprocess
import sys
import time
import tarfile
import urllib.request

ROOT = Path(__file__).resolve().parent
REPO = ROOT.parents[1]
sys.path.insert(0, str(REPO / 'stage3/drivers/tsc'))
from corpus import materialize


def save(path, value):
    path.write_text(json.dumps(value, indent=2) + '\n')


def digest(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def setup(argv, cwd, log):
    with log.open('ab') as stream:
        subprocess.run(argv, cwd=cwd, stdout=stream, stderr=stream, check=True, timeout=600)


def capture(argv, cwd, prefix):
    argv = ['/usr/bin/timeout', '--signal=TERM', '--kill-after=5s', '300s', *argv]
    save(prefix.with_suffix('.command.json'), {'argv': argv, 'cwd': str(cwd)})
    with prefix.with_suffix('.stdout').open('wb') as stdout, prefix.with_suffix('.stderr').open('wb') as stderr:
        start = time.perf_counter()
        result = subprocess.run(argv, cwd=cwd, stdout=stdout, stderr=stderr)
        elapsed = time.perf_counter() - start
    prefix.with_suffix('.exit').write_text(str(result.returncode) + '\n')
    return (prefix.with_suffix('.stdout').read_bytes(), prefix.with_suffix('.stderr').read_bytes(), result.returncode), elapsed



def comparison_table(output, rows):
    labels = ['node', 'go-single', 'go-default']
    if any(row['compiler'] == 'native' for row in rows):
        labels.append('native')
    lines = ['| Input | ' + ' | '.join(labels) + ' |', '|---|' + '---:|' * len(labels)]
    by_input = {}
    for row in rows:
        by_input.setdefault(row['input'], {})[row['compiler']] = row
    for name, modes in by_input.items():
        cells = []
        for label in labels:
            row = modes[label]
            cells.append(f"{row['mean']:.4f} ± {row['stdev']:.4f}")
        lines.append('| ' + name + ' | ' + ' | '.join(cells) + ' |')
    (output / 'comparison.md').write_text('\n'.join(lines) + '\n')


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('output', type=Path)
    parser.add_argument('--native', type=Path)
    args = parser.parse_args()
    out = args.output.resolve()
    out.mkdir(parents=True, exist_ok=False)
    pins = json.loads((ROOT / 'pins.json').read_text())
    node = str(Path(shutil.which('node')).resolve())
    if subprocess.check_output([node, '--version'], text=True).strip() != pins['node']:
        raise RuntimeError('Node must be ' + pins['node'] + '; source the setup env.sh')
    deps = out / 'dependencies'
    deps.mkdir()
    for name in ('package.json', 'package-lock.json'):
        shutil.copyfile(ROOT / name, deps / name)
    log = out / 'setup.log'
    time_binary = Path('/usr/bin/time')
    if not time_binary.is_file():
        archive = out / 'time-1.9.tar.gz'
        urllib.request.urlretrieve('https://ftp.gnu.org/gnu/time/time-1.9.tar.gz', archive)
        if digest(archive) != 'fbacf0c81e62429df3e33bda4cee38756604f18e01d977338e23306a3e3b521e':
            raise RuntimeError('GNU time archive hash mismatch')
        with tarfile.open(archive) as tar:
            tar.extractall(out, filter='data')
        build = out / 'time-1.9'
        setup(['./configure'], build, log)
        setup(['make', '-j2'], build, log)
        time_binary = build / 'time'
    save(out / 'time-tool.json', {'path': str(time_binary), 'sha256': digest(time_binary),
                                'version': subprocess.check_output([str(time_binary), '--version'], text=True)})
    setup(['npm', 'ci', '--ignore-scripts', '--no-audit', '--no-fund'], deps, log)
    lib = deps / 'node_modules/typescript/lib'
    go_lib = deps / 'node_modules/@typescript/native-preview-linux-x64/lib'
    if not go_lib.is_dir():
        raise RuntimeError('This baseline currently supports Linux x64')
    go = str(go_lib / 'tsgo')
    # Explicit noLib roots below make every compiler read the stock library bytes.
    commands = {'node': [node, str(lib / 'tsc.js')], 'go-single': [go, '--singleThreaded'], 'go-default': [go]}
    if args.native:
        commands['native'] = [str(args.native.resolve())]
    versions = {'node_version': pins['node'], 'node_sha256': digest(Path(node)),
                'tsc_sha256': digest(lib / '_tsc.js'), 'typescript_api_sha256': digest(lib / 'typescript.js'),
                'tsgo_sha256': digest(Path(go)), 'pins': pins,
                'versions': {name: subprocess.check_output(cmd + ['--version'], text=True).strip() for name, cmd in commands.items()}}
    if args.native:
        versions['native_sha256'] = digest(args.native.resolve())
    save(out / 'versions.json', versions)
    save(out / 'machine.json', {'nproc': subprocess.check_output(['nproc'], text=True).strip(),
                               'cpu': Path('/proc/cpuinfo').read_text(),
                               'cpu_max': Path('/sys/fs/cgroup/cpu.max').read_text(),
                               'uname': subprocess.check_output(['uname', '-a'], text=True).strip(),
                               'load_start': Path('/proc/loadavg').read_text(),
                               'gomaxprocs': os.environ.get('GOMAXPROCS'), 'commands': commands})
    inputs = out / 'inputs'
    inputs.mkdir()
    trees = {}
    for name, (url, commit) in pins['repositories'].items():
        tree = inputs / name
        setup(['git', 'init', str(tree)], out, log)
        setup(['git', 'fetch', '--depth', '1', url, commit], tree, log)
        if name == 'typescript':
            setup(['git', 'sparse-checkout', 'set', 'src/compiler', 'scripts'], tree, log)
        setup(['git', 'checkout', '--detach', 'FETCH_HEAD'], tree, log)
        observed = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=tree, text=True).strip()
        if observed != commit:
            raise RuntimeError('input pin mismatch: ' + name)
        trees[name] = tree
    # TypeScript's own generator supplies its required generated source.
    tree = trees['typescript']
    (tree / 'node_modules').symlink_to(deps / 'node_modules', target_is_directory=True)
    setup([node, 'scripts/processDiagnosticMessages.mjs', 'src/compiler/diagnosticMessages.json'], tree, log)
    projects = []
    def project(name, cwd, config, options, files):
        # Resolve a fixed stock library closure through triple-slash lib references.
        targets = options.pop('lib', ['es2020', 'dom'])
        pending = list(targets)
        libraries = set()
        while pending:
            target = pending.pop().lower()
            filename = 'lib.' + target + '.d.ts'
            if filename in libraries:
                continue
            libraries.add(filename)
            pending.extend(re.findall(r'<reference lib="([^"]+)"', (lib / filename).read_text()))
        options.update(noLib=True, incremental=False, composite=False)
        config_data = {'compilerOptions': options, 'files': files + [str(lib / f) for f in sorted(libraries)]}
        save(cwd / config, config_data)
        projects.append((name, cwd, ['-p', config, '--noEmit', '--pretty', 'false']))
    selected = json.loads((REPO / 'stage3/drivers/tsc/selection.json').read_text())['cases']
    for index in pins['driver_cases']:
        row = selected[index - 1]
        cwd = inputs / row['id']
        cwd.mkdir()
        source = REPO / 'stage3/drivers/tsc' / row['path']
        content, options = materialize(source.read_bytes())
        (cwd / source.name).write_text(content)
        project(row['id'], cwd, 'tsconfig.json', {'types': [], 'skipDefaultLibCheck': True, 'noErrorTruncation': True,
                                               'ignoreDeprecations': '6.0', 'target': 'es2020', **options}, [source.name])
    for name in ('mitt', 'zod'):
        cwd = trees[name]
        sources = sorted(cwd.glob('src/**/*.ts')) if name == 'mitt' else sorted((cwd / 'packages/zod/src').glob('**/*.ts'))
        files = [str(p.relative_to(cwd)) for p in sources if '/tests/' not in str(p) and '/bench/' not in str(p) and '/benchmarks/' not in str(p) and p.name != 'compile.ts']
        project(name, cwd, 'performance.json', {'target': 'es2020', 'module': 'nodenext', 'strict': True,
                                              'types': [], 'skipLibCheck': True, 'allowImportingTsExtensions': True}, files)
    # Keep the upstream compiler project and its options. Only replace implicit libs
    # with the same explicit stock declaration roots and disable build-state writes.
    compiler = tree / 'src/compiler'
    original = compiler / 'tsconfig.json'
    shutil.copyfile(original, compiler / 'tsconfig.upstream.json')
    data = json.loads(original.read_text())
    data['compilerOptions'].update(noLib=True, lib=None, incremental=False, composite=False)
    data['files'] = [str(lib / 'lib.es2020.d.ts')]
    # TypeScript follows the stock library's references; list closure explicitly.
    pending = ['es2020']; libraries = set()
    while pending:
        target = pending.pop(); filename = 'lib.' + target + '.d.ts'
        if filename in libraries: continue
        libraries.add(filename)
        pending.extend(re.findall(r'<reference lib="([^"]+)"', (lib / filename).read_text()))
    data['files'] = [str(lib / f) for f in sorted(libraries)]
    save(original, data)
    projects.append(('typescript-compiler', tree, ['-p', 'src/compiler', '--noEmit', '--pretty', 'false']))
    hashes = {}
    for path in inputs.rglob('*'):
        if path.is_file() and '.git' not in path.parts and 'node_modules' not in path.parts:
            hashes[str(path.relative_to(inputs))] = digest(path)
    hashes.update({'stock-lib/' + p.name: digest(p) for p in lib.glob('lib*.d.ts')})
    save(out / 'input-hashes.json', hashes)
    evidence = out / 'raw'; evidence.mkdir()
    accepted = []; excluded = []; native_bad = []
    # Global preflight: a bad native binary receives no timed or profile invocation.
    for name, cwd, flags in projects:
        folder = evidence / name; folder.mkdir()
        reference, _ = capture(commands['node'] + flags, cwd, folder / 'node-preflight')
        differing = []
        for label, command in commands.items():
            if label == 'node': continue
            result, _ = capture(command + flags, cwd, folder / (label + '-preflight'))
            if result != reference:
                differing.append(label)
                if label == 'native': native_bad.append(name)
        if differing or reference[2] not in (0, 2):
            excluded.append({'input': name, 'differing': differing, 'node_exit': reference[2]})
        else:
            accepted.append((name, cwd, flags, reference))
    save(out / 'preflight.json', {'accepted': [p[0] for p in accepted], 'excluded': excluded, 'native_mismatches': native_bad})
    if native_bad:
        print('REFUSED: native diagnostics differ on ' + ', '.join(native_bad) + '; no timing performed', flush=True)
        return 1
    results = []
    for name, cwd, flags, reference in accepted:
        folder = evidence / name
        samples = {label: [] for label in commands}
        for label, command in commands.items():
            for warmup in range(2):
                actual, _ = capture(command + flags, cwd, folder / f'{label}-warmup-{warmup}')
                if actual != reference: raise RuntimeError('warmup diagnostics changed')
        randomizer = random.Random(0)
        for run in range(10):
            labels = list(commands); randomizer.shuffle(labels)
            for label in labels:
                actual, elapsed = capture(commands[label] + flags, cwd, folder / f'{label}-run-{run:02}')
                if actual != reference: raise RuntimeError('timed diagnostics changed')
                samples[label].append(elapsed)
        for label, command in commands.items():
            memory = folder / (label + '-memory.txt')
            actual, _ = capture([str(time_binary), '-v', '-o', str(memory)] + command + flags, cwd, folder / (label + '-memory'))
            if actual != reference: raise RuntimeError('memory diagnostics changed')
            peak = int(re.search(r'Maximum resident set size \(kbytes\): (\d+)', memory.read_text())[1])
            values = samples[label]
            results.append({'input': name, 'compiler': label, 'seconds': values, 'mean': statistics.mean(values),
                            'stdev': statistics.stdev(values), 'min': min(values), 'max': max(values), 'peak_rss_kib': peak})
        save(out / 'results.json', results)
        print('measured ' + name, flush=True)
    # Largest selected input by total checked source bytes, rather than elapsed noise.
    largest = max(accepted, key=lambda item: sum(p.stat().st_size for p in item[1].rglob('*.ts') if '.git' not in p.parts and 'node_modules' not in p.parts))
    name, cwd, flags, _ = largest
    folder = evidence / name
    for label, command in commands.items():
        capture(command + flags + ['--extendedDiagnostics'], cwd, folder / (label + '-extended'))
    capture([node, '--cpu-prof', '--cpu-prof-dir=' + str(out), '--cpu-prof-name=tsc.cpuprofile', str(lib / 'tsc.js')] + flags, cwd, folder / 'node-profile')
    profile = json.loads((out / 'tsc.cpuprofile').read_text())
    frames = {n['id']: n['callFrame'] for n in profile['nodes']}
    totals = {}
    for sample, delta in zip(profile['samples'], profile['timeDeltas']):
        function = frames[sample]['functionName'] or '(anonymous)'
        totals[function] = totals.get(function, 0) + delta
    save(out / 'profile-summary.json', {'input': name, 'total_sample_us': sum(totals.values()),
                                      'top_self_time_us': sorted(totals.items(), key=lambda x: -x[1])[:30]})
    lines = ['| Input | Compiler | Mean s | SD s | Min s | Max s | Peak KiB |', '|---|---|---:|---:|---:|---:|---:|']
    for r in results:
        lines.append(f"| {r['input']} | {r['compiler']} | {r['mean']:.4f} | {r['stdev']:.4f} | {r['min']:.4f} | {r['max']:.4f} | {r['peak_rss_kib']} |")
    (out / 'table.md').write_text('\n'.join(lines) + '\n')
    comparison_table(out, results)
    machine = json.loads((out / 'machine.json').read_text())
    machine['load_end'] = Path('/proc/loadavg').read_text()
    save(out / 'machine.json', machine)
    return 0


if __name__ == '__main__':
    raise SystemExit(main())

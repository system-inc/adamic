#!/usr/bin/env python3
"""Build, verify and measure; no compiler source changes and no invented data."""
import argparse
import csv
import gzip
import hashlib
import json
import os
import shutil
import statistics
import subprocess
import time
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
HERE = ROOT / 'bench/wasm'
OUT = HERE / 'out'
RESULTS = HERE / 'results'
SDK = Path(os.environ['WASI_SYSROOT']).resolve().parents[1]
TOOLS = Path(os.environ.get('WASM_SIZE_TOOLS', '/workspace/wasm-size-tools'))
OPT = TOOLS / 'binaryen-version_126/bin/wasm-opt'
ESBUILD = TOOLS / 'node_modules/.bin/esbuild'
BROTLI = TOOLS / 'brotli-build/brotli'
PROGRAMS = {
    'hello': 'internal/load/testdata/0.1/compile/01_hello.ts',
    'dedication': 'dedication/dedication.a',
    'collections': 'internal/oracle/testdata/collections.a',
    'request': 'internal/native/wasm/request.a',
}
CONFIGS = {
    'O2': {}, 'Oz': {'optimization': '-Oz'},
    'gc': {'link_flags': ['-Wl,--gc-sections']},
    'sections': {'compile_flags': ['-ffunction-sections', '-fdata-sections']},
    'lto': {'compile_flags': ['-flto']},
    'no-whole': {'drop_archive': True},
    'Oz-opt': {'optimization': '-Oz', 'wasm_opt': str(OPT)},
    'O2-opt': {'wasm_opt': str(OPT)},
    'Oz-strip-opt': {'optimization': '-Oz', 'llvm_strip': str(SDK / 'bin/llvm-strip'), 'wasm_opt': str(OPT)},
}


def run(command, env=None, check=True, input_data=None):
    with (RESULTS / 'commands.jsonl').open('a') as log:
        log.write(json.dumps([str(value) for value in command]) + '\n')
    result = subprocess.run([str(value) for value in command], cwd=ROOT, env=env, capture_output=True, input=input_data)
    if check and result.returncode:
        raise RuntimeError(f'{command}: exit {result.returncode}\n{result.stderr.decode(errors="replace")}')
    return result


def environment(name):
    directory = OUT / 'sdk' / name
    (directory / 'bin').mkdir(parents=True, exist_ok=True)
    (directory / 'share').mkdir(exist_ok=True)
    sysroot = directory / 'share/wasi-sysroot'
    if not sysroot.exists():
        sysroot.symlink_to(SDK / 'share/wasi-sysroot', target_is_directory=True)
    wrapper = directory / 'bin/clang'
    shutil.copy2(HERE / 'clang.py', wrapper)
    wrapper.chmod(0o755)
    archiver = directory / 'bin/llvm-ar'
    if not archiver.exists():
        archiver.symlink_to(SDK / 'bin/llvm-ar')
    configuration = {'clang': str(SDK / 'bin/clang'), 'optimization': '-O2',
                     'commands': str(RESULTS / f'clang-{name}.jsonl'), **CONFIGS[name]}
    (directory / 'bin/configuration.json').write_text(json.dumps(configuration))
    env = os.environ.copy()
    env['WASI_SYSROOT'] = str(sysroot)
    env['ADAMIC_ORACLE_WASI'] = '1'
    env['ADAMIC_GATE_UNCACHED'] = '1'
    return env


def observation(command):
    result = run(command, check=False)
    return result.returncode, result.stdout, result.stderr


def equivalent(expected, actual):
    if expected != actual:
        raise AssertionError('stdout, stderr or exit differs')


def node(command):
    return ['node', '--disable-warning=ExperimentalWarning', *command]


def wasm_command(path, request=False, check=False):
    return node([HERE / 'reactor.mjs' if request else ROOT / 'oracle/wasi.mjs', path,
                 *(['--check'] if check else [])])


def measure_sizes():
    rows = []
    run(['go', 'build', '-o', OUT / 'adamic', './cmd/adamic'])
    package = OUT / 'node_modules/adamic'
    package.mkdir(parents=True, exist_ok=True)
    (package / 'package.json').write_text(json.dumps({'type': 'module', 'exports': './index.mjs'}))
    shutil.copy2(ROOT / 'oracle/adamic.mjs', package / 'index.mjs')
    rows.append(size_row('shared', 'js-runtime', package / 'index.mjs'))
    candidates = sorted([(p.stat().st_size, str(p.relative_to(ROOT)))
                         for p in (ROOT / 'internal/oracle/testdata').rglob('*')
                         if p.suffix in ['.a', '.ts']
                         if 'new Map' in p.read_text()], reverse=True)
    (RESULTS / 'map-candidates.json').write_text(json.dumps(candidates, indent=2) + '\n')
    for name, source in PROGRAMS.items():
        directory = OUT / name
        directory.mkdir(exist_ok=True)
        # .mts is a scratch copy of the unmodified source, for Node's built-in type stripping.
        source_copy = directory / 'source.mts'
        shutil.copy2(ROOT / source, source_copy)
        js = directory / 'program.mjs'
        js.write_bytes(run([OUT / 'adamic', 'js', source]).stdout)
        mini = directory / 'program.min.mjs'
        run([ESBUILD, js, '--minify', '--bundle', '--platform=node', '--format=esm', f'--outfile={mini}'])
        expected = observation(node([source_copy]))
        equivalent(expected, observation(node([js])))
        equivalent(expected, observation(node([mini])))
        for variant, path in [('source', source_copy), ('js', js), ('js-min', mini)]:
            rows.append(size_row(name, variant, path))
        for config in CONFIGS:
            path = directory / f'{config}.wasm'
            result = run([OUT / 'adamic', 'build', '--target', 'wasm32-wasi', source, '-o', path],
                         environment(config), check=False)
            (RESULTS / f'build-{name}-{config}.log').write_bytes(result.stdout + result.stderr)
            if result.returncode:
                rows.append({'program': name, 'variant': config, 'error': result.stderr.decode()})
                continue
            if name == 'request':
                actual = observation(wasm_command(path, True, True))
                equivalent((0, b'', b''), actual)
            else:
                equivalent(expected, observation(wasm_command(path)))
            for treatment in ['raw', 'strip', 'opt']:
                target = path
                if treatment == 'strip':
                    target = directory / f'{config}.strip.wasm'
                    shutil.copy2(path, target)
                    run([SDK / 'bin/llvm-strip', target])
                if treatment == 'opt':
                    target = directory / f'{config}.opt.wasm'
                    run([OPT, '-Oz', path, '-o', target])
                if name == 'request':
                    equivalent((0, b'', b''), observation(wasm_command(target, True, True)))
                else:
                    equivalent(expected, observation(wasm_command(target)))
                rows.append(size_row(name, f'{config}-{treatment}', target))
    (RESULTS / 'sizes.json').write_text(json.dumps(rows, indent=2) + '\n')
    with (RESULTS / 'sizes.csv').open('w') as output:
        writer = csv.DictWriter(output, fieldnames=['program', 'variant', 'raw', 'gzip9', 'brotli11', 'sha256', 'error'])
        writer.writeheader()
        writer.writerows(rows)


def size_row(name, variant, path):
    data = path.read_bytes()
    compressed = run(['gzip', '-9', '-n', '-c', path]).stdout
    assert gzip.decompress(compressed) == data
    brotli = run([BROTLI, '-q', '11', '-c', path]).stdout
    run(node([HERE / 'compress.mjs', path]), input_data=brotli)
    return {'program': name, 'variant': variant, 'raw': len(data), 'gzip9': len(compressed),
            'brotli11': len(brotli), 'sha256': hashlib.sha256(data).hexdigest()}


def startup():
    results = {'before': machine(), 'programs': {}}
    for name in PROGRAMS:
        directory = OUT / name
        wasm = directory / 'O2.wasm'
        tasks = {'wasm': wasm_command(wasm, name == 'request'),
                 'js': node([directory / 'program.mjs']), 'source': node([directory / 'source.mts'])}
        samples = {key: [] for key in tasks}
        # One untimed warmup per form, then rotate order each round.
        for command in tasks.values():
            run(command)
        keys = list(tasks)
        for round in range(30):
            offset = round % len(keys)
            for key in keys[offset:] + keys[:offset]:
                started = time.perf_counter_ns()
                result = subprocess.run([str(v) for v in tasks[key]], stdout=subprocess.DEVNULL,
                                        stderr=subprocess.DEVNULL, cwd=ROOT)
                elapsed = (time.perf_counter_ns() - started) / 1e6
                if result.returncode:
                    raise RuntimeError(f'timed {key} failed')
                samples[key].append(elapsed)
        stages = json.loads(run(node([HERE / 'startup.mjs', wasm, directory / 'program.mjs', directory / 'source.mts'])).stdout)
        results['programs'][name] = {'process_ms': samples, 'stages_ms': stages}
    results['after'] = machine()
    (RESULTS / 'startup.json').write_text(json.dumps(results, indent=2) + '\n')


def machine():
    return {'nproc': run(['nproc']).stdout.decode().strip(),
            'loadavg': Path('/proc/loadavg').read_text().strip(),
            'cpu_max': Path('/sys/fs/cgroup/cpu.max').read_text().strip(),
            'time': time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime())}


def oracle(config):
    command = ['go', 'test', './internal/oracle', '-run', '^TestWASIAgreesWithNode$', '-count=1', '-v', '-timeout', '30m']
    env = environment(config)
    revision = run(['git', 'rev-parse', 'HEAD']).stdout.decode().strip()
    with (RESULTS / f'oracle-{config}.log').open('wb') as log:
        code = subprocess.run(command, env=env, cwd=ROOT, stdout=log, stderr=subprocess.STDOUT).returncode
    lines = (RESULTS / f'oracle-{config}.log').read_text().splitlines()
    counts = oracle_counts(lines)
    (RESULTS / f'oracle-{config}.json').write_text(json.dumps({'revision': revision, 'exit': code, **counts}, indent=2) + '\n')
    print(config, code, counts, flush=True)


def oracle_counts(lines):
    counts = {status: sum(line.startswith(f'    --- {status}:') for line in lines)
              for status in ['PASS', 'FAIL', 'SKIP']}
    if counts['PASS'] + counts['FAIL'] == 0:
        raise AssertionError('oracle selected no executable subtests')
    return counts


def mutants():
    expected = observation(node([OUT / 'hello/source.mts']))
    # Mutate emitted JS only, after a known-good control. Exact comparison must reject it.
    control = (OUT / 'hello/program.mjs').read_text()
    path = OUT / 'hello/mutant.mjs'
    outcomes = {}
    for name, text in [('stdout', control.replace('Kenneth Lane Thompson', 'Kenneth Lane Thompson!')),
                       ('exit', control + '\nprocess.exitCode = 23;\n'),
                       ('stderr', control + '\nprocess.stderr.write("mutant\\n");\n')]:
        path.write_text(text)
        try:
            equivalent(expected, observation(node([path])))
        except AssertionError:
            outcomes[name] = 'caught by exact stdout/stderr/exit comparison'
        else:
            raise AssertionError(f'{name} mutant survived')
    data = (OUT / 'hello/program.mjs').read_bytes()
    try:
        assert gzip.decompress(gzip.compress(data + b'!')) == data
    except AssertionError:
        outcomes['gzip-roundtrip'] = 'caught by decompressed-byte equality'
    else:
        raise AssertionError('gzip mutant survived')
    wrong_compression = run([BROTLI, '-q', '11', '-c'], input_data=data + b'!').stdout
    result = run(node([HERE / 'compress.mjs', OUT / 'hello/program.mjs']), check=False, input_data=wrong_compression)
    assert result.returncode != 0 and b'Brotli round trip failed' in result.stderr
    (RESULTS / 'brotli-mutant.log').write_bytes(result.stdout + result.stderr)
    outcomes['brotli-roundtrip'] = 'caught by Brotli round trip failed after compressing input plus !'
    reactor_mutant = OUT / 'reactor-mutant.mjs'
    reactor_mutant.write_text((HERE / 'reactor.mjs').read_text().replace('actual !== source.handleRequest(request)', "actual + '!' !== source.handleRequest(request)"))
    result = run(node([reactor_mutant, OUT / 'request/O2.wasm', '--check']), check=False)
    assert result.returncode != 0 and b'response differs' in result.stderr
    (RESULTS / 'reactor-mutant.log').write_bytes(result.stdout + result.stderr)
    outcomes['reactor-response'] = 'caught by source response comparison after adding ! to the observed response'
    empty_log = RESULTS / 'empty-selection-mutant.log'
    with empty_log.open('wb') as log:
        code = subprocess.run(['go', 'test', './internal/oracle', '-run', '^NoSuchWASITest$', '-count=1', '-v'], cwd=ROOT, stdout=log, stderr=subprocess.STDOUT).returncode
    assert code == 0
    try:
        oracle_counts(empty_log.read_text().splitlines())
    except AssertionError:
        outcomes['empty-selection'] = 'caught by nonzero oracle subtest guard'
    else:
        raise AssertionError('empty-selection mutant survived')
    (RESULTS / 'mutants.json').write_text(json.dumps(outcomes, indent=2) + '\n')


def versions():
    commands = {'node': ['node', '--version'], 'go': ['go', 'version'],
                'clang': [SDK / 'bin/clang', '--version'], 'strip': [SDK / 'bin/llvm-strip', '--version'],
                'binaryen': [OPT, '--version'], 'esbuild': [ESBUILD, '--version'], 'gzip': ['gzip', '--version']}
    values = {key: run(value).stdout.decode().strip() for key, value in commands.items()}
    brotli_version = run([BROTLI, '--version'])
    values['brotli'] = (brotli_version.stdout + brotli_version.stderr).decode().strip()
    values['python'] = run(['python3', '--version']).stdout.decode().strip()
    values['brotli_build_compiler'] = run(['cc', '--version']).stdout.decode().splitlines()[0]
    values['hyperfine'] = shutil.which('hyperfine')
    values['revision'] = run(['git', 'rev-parse', 'HEAD']).stdout.decode().strip()
    values['configurations'] = CONFIGS
    (RESULTS / 'versions.json').write_text(json.dumps(values, indent=2) + '\n')


if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument('mode', choices=['measure', 'archive', 'compress', 'strip-opt', 'startup', 'oracle', 'mutants'])
    parser.add_argument('--configuration', choices=list(CONFIGS), default='O2')
    args = parser.parse_args()
    OUT.mkdir(exist_ok=True)
    RESULTS.mkdir(exist_ok=True)
    if args.mode == 'measure':
        versions()
        measure_sizes()
        mutants()
    elif args.mode == 'archive':
        for source, destination in [('wasm-size-setup.log', 'setup.log'), ('wasm-size-brotli-build.log', 'brotli-cmake-failure.log'), ('wasm-size-brotli-cc.log', 'brotli-build.log'), ('wasm-size-npm.log', 'esbuild-install.log')]:
            path = Path('/tmp') / source
            if path.exists():
                shutil.copy2(path, RESULTS / destination)
        build_rows = [{'program': row['program'], 'configuration': row['variant'][:-4], 'exit': 0}
                      for row in json.loads((RESULTS / 'sizes.json').read_text())
                      if row['variant'].endswith('-raw') and 'error' not in row]
        (RESULTS / 'build-status.json').write_text(json.dumps(build_rows, indent=2) + '\n')
        for path in list(RESULTS.glob('*.log')) + list(RESULTS.glob('*.jsonl')):
            if path.stat().st_size:
                path.with_suffix(path.suffix + '.gz').write_bytes(gzip.compress(path.read_bytes(), compresslevel=9, mtime=0))
            path.unlink()
    elif args.mode == 'compress':
        rows = json.loads((RESULTS / 'sizes.json').read_text())
        refreshed = []
        for row in rows:
            name, variant = row['program'], row['variant']
            if name == 'shared':
                path = OUT / 'node_modules/adamic/index.mjs'
            elif variant in ['source', 'js', 'js-min']:
                path = OUT / name / {'source': 'source.mts', 'js': 'program.mjs', 'js-min': 'program.min.mjs'}[variant]
            else:
                treatment = next(suffix for suffix in ['strip-opt', 'raw', 'strip', 'opt'] if variant.endswith('-' + suffix))
                configuration = variant[:-len(treatment) - 1]
                path = OUT / name / (configuration + ('' if treatment == 'raw' else '.' + treatment) + '.wasm')
            assert hashlib.sha256(path.read_bytes()).hexdigest() == row['sha256']
            refreshed.append(size_row(name, variant, path))
        (RESULTS / 'sizes.json').write_text(json.dumps(refreshed, indent=2) + '\n')
        with (RESULTS / 'sizes.csv').open('w') as output:
            writer = csv.DictWriter(output, fieldnames=['program', 'variant', 'raw', 'gzip9', 'brotli11', 'sha256', 'error'])
            writer.writeheader()
            writer.writerows(refreshed)
        metadata = json.loads((RESULTS / 'versions.json').read_text())
        brotli_version = run([BROTLI, '--version'])
        metadata['brotli'] = (brotli_version.stdout + brotli_version.stderr).decode().strip()
        (RESULTS / 'versions.json').write_text(json.dumps(metadata, indent=2) + '\n')
    elif args.mode == 'strip-opt':
        rows = json.loads((RESULTS / 'sizes.json').read_text())
        rows = [row for row in rows if not row['variant'].endswith('-strip-opt')]
        for name in PROGRAMS:
            for config in CONFIGS:
                original = OUT / name / f'{config}.strip.wasm'
                if not original.exists():
                    continue
                target = OUT / name / f'{config}.strip-opt.wasm'
                run([OPT, '-Oz', original, '-o', target])
                expected = (0, b'', b'') if name == 'request' else observation(node([OUT / name / 'source.mts']))
                equivalent(expected, observation(wasm_command(target, name == 'request', name == 'request')))
                rows.append(size_row(name, f'{config}-strip-opt', target))
        (RESULTS / 'sizes.json').write_text(json.dumps(rows, indent=2) + '\n')
        with (RESULTS / 'sizes.csv').open('w') as output:
            writer = csv.DictWriter(output, fieldnames=['program', 'variant', 'raw', 'gzip9', 'brotli11', 'sha256', 'error'])
            writer.writeheader()
            writer.writerows(rows)
    elif args.mode == 'startup':
        startup()
    elif args.mode == 'oracle':
        oracle(args.configuration)
    else:
        mutants()

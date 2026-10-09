#!/usr/bin/env python3
"""Bootstrap pinned inputs, prove the fixture/mutant, then profile two CLIs."""
import argparse
import gzip
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import sys

ROOT = Path(__file__).resolve().parent


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def save(path, value):
    path.write_text(json.dumps(value, indent=2) + '\n')


def run(argv, cwd, prefix, expected=0):
    save(prefix.with_suffix('.command.json'), {'argv': argv, 'cwd': str(cwd)})
    with prefix.with_suffix('.stdout').open('wb') as stdout, prefix.with_suffix('.stderr').open('wb') as stderr:
        result = subprocess.run(argv, cwd=cwd, stdout=stdout, stderr=stderr, timeout=1800)
    prefix.with_suffix('.exit').write_text(str(result.returncode) + '\n')
    if result.returncode != expected:
        raise RuntimeError(f'{prefix}: expected {expected}, got {result.returncode}; see log files')
    return tuple(prefix.with_suffix('.' + suffix).read_bytes() for suffix in ('stdout', 'stderr', 'exit'))


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('output', type=Path, help='new output directory')
    args = parser.parse_args()
    out = args.output.resolve(); out.mkdir(parents=True, exist_ok=False)
    node = str(Path(shutil.which('node')).resolve())
    pins = json.loads((ROOT / 'pins.json').read_text())
    if subprocess.check_output([node, '--version'], text=True).strip() != pins['node']:
        raise RuntimeError('source toolchain env.sh; Node must be ' + pins['node'])
    fixture = out / 'fixture'; fixture.mkdir()
    observe = str(ROOT / 'observe.cjs')
    def observed(mode, folder, stock, flags=None, target='-', cwd=None):
        folder.mkdir(exist_ok=True)
        return run([node, '--expose-gc', observe, mode, str(folder), str(stock), str(target), *(flags or [])],
                   cwd or stock.parent, folder / 'cli')
    observed('fixture', fixture, ROOT / 'fixture.cjs')
    run([sys.executable, str(ROOT / 'parse.py'), str(fixture), '--fixture'], ROOT, fixture / 'verify')
    run([sys.executable, str(ROOT / 'parse.py'), str(fixture), '--fixture', '--drop-constructor', 'Node'],
        ROOT, fixture / 'mutant', expected=1)
    error = (fixture / 'mutant.stderr').read_text()
    if 'fixture parse/Node: expected 6 live, got 0' not in error:
        raise RuntimeError('mutant failed for an unrelated reason')
    print('PASS predictable fixture; dropped-Node parser mutant caught', flush=True)
    deps = out / 'dependencies'; deps.mkdir()
    for name in ('package.json', 'package-lock.json'): shutil.copyfile(ROOT / name, deps / name)
    run(['npm', 'ci', '--ignore-scripts', '--no-audit', '--no-fund'], deps, out / 'npm')
    stock = deps / 'node_modules/typescript/lib/_tsc.js'
    loader = stock.parent / 'tsc.js'
    if sha(stock) != pins['typescript_cli_sha256']: raise RuntimeError('CLI hash mismatch')
    projects = []
    for name, url, commit in [('compiler', 'https://github.com/microsoft/TypeScript.git', pins['typescript_commit']),
                              ('mitt', 'https://github.com/developit/mitt.git', pins['mitt_commit'])]:
        tree = out / (name + '-source'); tree.mkdir()
        run(['git', 'init'], tree, out / (name + '-init'))
        run(['git', 'fetch', '--depth', '1', url, commit], tree, out / (name + '-fetch'))
        if name == 'compiler': run(['git', 'sparse-checkout', 'set', 'src/compiler', 'scripts'], tree, out / 'compiler-sparse')
        run(['git', 'checkout', '--detach', 'FETCH_HEAD'], tree, out / (name + '-checkout'))
        if subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=tree, text=True).strip() != commit:
            raise RuntimeError('repository pin mismatch')
        if name == 'compiler':
            (tree / 'node_modules').symlink_to(deps / 'node_modules', target_is_directory=True)
            run([node, 'scripts/processDiagnosticMessages.mjs', 'src/compiler/diagnosticMessages.json'], tree, out / 'generate')
            flags = ['-p', str(tree / 'src/compiler'), '--noEmit', '--pretty', 'false', '--incremental', 'false', '--composite', 'false']
        else:
            config = {'compilerOptions': {'target': 'es2020', 'module': 'nodenext', 'strict': True, 'types': [], 'skipLibCheck': True},
                      'files': ['src/index.ts']}
            save(tree / 'allocation.json', config)
            flags = ['-p', str(tree / 'allocation.json'), '--noEmit', '--pretty', 'false']
        projects.append((name, tree, flags))
    input_hashes = {}
    for name, tree, flags in projects:
        for item in tree.rglob('*'):
            if item.is_file() and '.git' not in item.parts and 'node_modules' not in item.parts:
                input_hashes[name + '/' + str(item.relative_to(tree))] = sha(item)
    save(out / 'input-hashes.json', input_hashes)
    save(out / 'versions.json', {'pins': pins, 'node_sha256': sha(Path(node)), 'cli_sha256': sha(stock),
                               'tsc_version': subprocess.check_output([node, str(loader), '--version'], text=True).strip(),
                               'base_commit': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, text=True).strip(),
                               'nproc': subprocess.check_output(['nproc'], text=True).strip(),
                               'cpu_max': Path('/sys/fs/cgroup/cpu.max').read_text(),
                               'cpu_model': next(line for line in Path('/proc/cpuinfo').read_text().splitlines() if line.startswith('model name')),
                               'harness_hashes': {p.name: sha(p) for p in ROOT.glob('*') if p.is_file()}})
    for name, tree, flags in projects:
        folder = out / name; folder.mkdir()
        pure = folder / 'stock'; pure.mkdir()
        oracle = run([node, str(loader), *flags], tree, pure / 'cli')
        run([node, '--heap-prof', '--heap-prof-interval=' + str(pins['sampling_interval']),
             '--heap-prof-dir=' + str(pure), '--heap-prof-name=stock.heapprofile', str(loader), *flags], tree, pure / 'heap')
        run([node, str(loader), *flags, '--extendedDiagnostics'], tree, pure / 'extended')
        for mode in ('observe', 'peak', 'lifetime'):
            destination = folder / mode; destination.mkdir()
            target = folder / 'observe/observations.json' if mode == 'peak' else '-'
            actual = observed(mode, destination, stock, flags, target, tree)
            if actual != oracle: raise RuntimeError(f'{name}/{mode}: observed CLI changed stdout/stderr/status')
            print('PASS ' + name + '/' + mode + ': output matches stock', flush=True)
        run([sys.executable, str(ROOT / 'parse.py'), str(folder / 'lifetime')], ROOT, folder / 'parse')
        from parse import sampling, snapshot
        save(folder / 'stock-profile-summary.json', sampling(pure / 'stock.heapprofile'))
        save(folder / 'observe-profile-summary.json', sampling(folder / 'observe/collected.heapprofile'))
        peak_groups = snapshot(folder / 'peak/peak.heapsnapshot')
        save(folder / 'peak-summary.json', {category: {'count': len(values), 'self_bytes': sum(values.values())}
                                          for category, values in peak_groups.items()})
        observations = [json.loads((folder / mode / 'observations.json').read_text()) for mode in ('observe', 'peak', 'lifetime')]
        if not all(item['counters'] == observations[0]['counters'] for item in observations):
            raise RuntimeError(f'{name}: constructor counts changed between observation modes')
        print('PASS ' + name + ': all constructor counts agree across modes', flush=True)
    from report import write_report
    write_report(out)
    # Keep exact raw snapshots compact. Compression is after every observation.
    manifest = {}
    for path in out.rglob('*.heapsnapshot'):
        manifest[str(path.relative_to(out))] = {'sha256': sha(path), 'bytes': path.stat().st_size}
        with path.open('rb') as source, path.with_suffix('.heapsnapshot.gz').open('wb') as dest:
            with gzip.GzipFile(fileobj=dest, mode='wb', mtime=0) as compressor: shutil.copyfileobj(source, compressor)
        path.unlink()
    save(out / 'snapshot-hashes.json', manifest)
    print('PASS full allocation profile; report: ' + str(out / 'REPORT.md'), flush=True)


if __name__ == '__main__':
    main()

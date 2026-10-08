#!/usr/bin/env python3
"""Run recorded-outcome checks in candidate/base trees, independently of this tools tree."""
import argparse
import concurrent.futures
import contextlib
import fcntl
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

HERE = Path(__file__).resolve().parent
spec = importlib.util.spec_from_file_location('feedback', HERE / 'darwin-feedback.py')
feedback = importlib.util.module_from_spec(spec)
spec.loader.exec_module(feedback)
MODULE = 'github.com/system-inc/adamic/'
SCHEMA = 1


def command(args, cwd, **kwargs):
    return subprocess.run(args, cwd=cwd, text=True, stdout=subprocess.PIPE,
                          stderr=subprocess.PIPE, **kwargs)


def git(root, *args):
    result = command(['git', *args], root)
    if result.returncode:
        raise RuntimeError(result.stderr.strip())
    return result.stdout.strip()


def sha(root, ref):
    return git(root, 'rev-parse', '--verify', ref + '^{commit}')


def inventory(root):
    """Reviewed checks; discover new gap witnesses in each input tree too."""
    rows = json.loads((HERE / 'records-check-inventory.json').read_text())
    known = {(r['package'], r['test']) for r in rows}
    for source in sorted(root.glob('stage1/**/*gap*test.go')):
        relative = source.relative_to(root).as_posix()
        package = source.parent.relative_to(root).as_posix()
        for test in re.findall(r'^func (Test\w+)\(t \*testing.T\)', source.read_text(), re.M):
            if (package, test) not in known:
                records = [package + '/GAPS.md', relative]
                rows.append(dict(package=package, test=test, source=relative, records=records,
                                 refresh='${EDITOR:-vi} ' + ' '.join(records) +
                                 '; go test ./' + package + ' -count=1 -run ^' + test + '$'))
                known.add((package, test))
    return [r for r in rows if (root / r['source']).exists() and
            re.search(r'^func ' + re.escape(r['test']) + r'\(t \*testing.T\)',
                      (root / r['source']).read_text(), re.M)]


@contextlib.contextmanager
def tree(repository, target):
    """Use a supplied checkout, or a temporary worktree of a commit. Never checkout tools."""
    path = Path(target).expanduser()
    if path.is_dir():
        yield path.resolve(), None
        return
    commit = sha(repository, target)
    temporary = Path(tempfile.mkdtemp(prefix='adamic-records-'))
    root = temporary / 'tree'
    try:
        git(repository, 'worktree', 'add', '--detach', str(root), commit)
        # Borrow only a clean, exactly pinned dependency. No network, floating pin, or fallback.
        for line in git(repository, 'ls-tree', commit).splitlines():
            fields = line.split()
            if fields[:2] != ['160000', 'commit']:
                continue
            pin, name = fields[2:4]
            dependency = repository / name
            if not (dependency / '.git').exists() or git(dependency, 'rev-parse', 'HEAD') != pin:
                raise RuntimeError('dependency gap: ' + name + ' needs initialized commit ' + pin)
            if git(dependency, 'status', '--porcelain', '--untracked-files=no'):
                raise RuntimeError('dependency gap: ' + name + ' has tracked changes')
            destination = root / name
            destination.rmdir()
            destination.symlink_to(dependency, target_is_directory=True)
        # Optional installed npm inputs can be borrowed only with the exact lockfile.
        api = root / 'stage3/api'
        existing = repository / 'stage3/api'
        if (api / 'package-lock.json').exists() and (existing / 'node_modules').exists():
            if (api / 'package-lock.json').read_bytes() == (existing / 'package-lock.json').read_bytes():
                (api / 'node_modules').symlink_to(existing / 'node_modules', target_is_directory=True)
        yield root, commit
    finally:
        if root.exists():
            command(['git', 'worktree', 'remove', '--force', str(root)], repository)
        shutil.rmtree(temporary, ignore_errors=True)


def platform_overlay(root, directory):
    """Lift Go test platform skips without changing fixture files or simulating another GOOS."""
    replacements = {}
    pattern = re.compile(r'(if\s+)([^\n{]*runtime\.GOOS[^\n{]*)(\s*\{\s*\n\s*t\.Skip(?:f)?\()')
    for path in root.rglob('*_test.go'):
        if 'cohere' in path.relative_to(root).parts:
            continue
        source = path.read_text()
        lifted = pattern.sub(lambda m: m[1] + 'false && (' + m[2].strip() + ')' + m[3], source)
        if lifted != source:
            destination = directory / (hashlib.sha256(str(path).encode()).hexdigest() + '.go')
            destination.write_text(lifted)
            replacements[str(path)] = str(destination)
    overlay = directory / 'overlay.json'
    overlay.write_text(json.dumps({'Replace': replacements}))
    return overlay, list(replacements)


def parse_run(stdout, stderr, row, code):
    outcomes, output, elapsed = {}, {}, {}
    for line in stdout.splitlines():
        try:
            event = json.loads(line)
        except ValueError:
            continue
        test = event.get('Test')
        if not test:
            continue
        if event.get('Action') in feedback.Outcomes:
            outcomes[test] = event['Action']
            elapsed[test] = event.get('Elapsed', 0)
        if event.get('Action') == 'output':
            output.setdefault(test, []).append(event.get('Output', ''))
    if row['test'] not in outcomes:
        reason = stderr.strip() or stdout.strip() or 'go test emitted no named outcome'
        return {'outcomes': {}, 'gaps': {row['test']: reason}, 'output': output, 'elapsed': elapsed}
    gaps = {test: ''.join(output.get(test, [])) for test, outcome in outcomes.items() if outcome == 'skip'}
    if code and all(value == 'pass' for value in outcomes.values()):
        gaps[row['test']] = stderr or 'process failed after passing test outcomes'
    return dict(outcomes=outcomes, gaps=gaps, output=output, elapsed=elapsed)


def run_one(root, row, env, timeout):
    started = time.monotonic()
    args = ['go', 'test', '-count=1', '-json', '-p=1', '-parallel=1',
            '-timeout=' + str(timeout) + 's', '-run=^' + row['test'] + '$', './' + row['package']]
    try:
        result = command(args, root, env=env, timeout=timeout + 30)
        parsed = parse_run(result.stdout, result.stderr, row, result.returncode)
    except (subprocess.TimeoutExpired, OSError) as error:
        parsed = dict(outcomes={}, gaps={row['test']: str(error)}, output={}, elapsed={})
    parsed.update(row=row, seconds=time.monotonic() - started)
    return parsed


def run_tree(root, jobs, timeout):
    rows = inventory(root)
    with tempfile.TemporaryDirectory(prefix='adamic-records-overlay-') as scratch:
        overlay, lifted = platform_overlay(root, Path(scratch))
        env = dict(os.environ)
        env['GOFLAGS'] = (env.get('GOFLAGS', '') + ' -overlay=' + str(overlay)).strip()
        env['GOMAXPROCS'] = '1'
        with concurrent.futures.ThreadPoolExecutor(max_workers=jobs) as executor:
            runs = list(executor.map(lambda row: run_one(root, row, env, timeout), rows))
    # The tsc driver is a pinned check too. A SHA does not identify a runnable tsc entry.
    if (root / 'stage3/drivers/tsc/driver.py').exists():
        runs.append(dict(row=dict(package='stage3/drivers/tsc', test='driver.py',
            records=['stage3/drivers/tsc/{tiny,corpus/*}/golden.{stdout,stderr,exit}'],
            refresh='bash stage3/drivers/tsc/run.sh --record -- <tsc command>'),
            outcomes={}, output={}, elapsed={}, seconds=0,
            gaps={'driver.py': 'requires an explicit runnable tsc command; candidate compiler is not a tsc driver'}))
    return dict(runs=runs, lifted=lifted)


def declarations(repository, base, candidate, root):
    trailers = '\n'.join(line[len('Moved-result:'):].strip()
        for line in git(repository, 'log', '--format=%B', base + '..' + candidate).splitlines()
        if line.startswith('Moved-result:'))
    file = root / 'stage3/fixtures/moved-results.txt'
    return trailers, file.read_text() if file.exists() else ''


def record_files(row, test, root):
    if row['package'] == 'stage3/fixtures' and test.startswith('TestFixtures/'):
        bucket = test.split('/')[1]
        return ['stage3/fixtures/' + bucket + '/status.json']
    files = []
    for record in row['records']:
        if '*' in record:
            files.extend(path.relative_to(root).as_posix() for path in root.glob(record))
        else:
            files.append(record)
    return files or row['records']


def compare(base, candidate, trailers, file, base_root, candidate_root):
    old, new, metadata = {}, {}, {}
    gaps = []
    for side, data, final in [('base', base, old), ('candidate', candidate, new)]:
        for run in data['runs']:
            row = run['row']
            for test, outcome in run['outcomes'].items():
                key = (row['package'], test)
                final[key] = outcome
                metadata[key] = row
            for test, reason in sorted(run['gaps'].items()):
                gaps.append(dict(side=side, package=row['package'], test=test,
                                 records=row['records'], refresh=row['refresh'], reason=reason))
    lines, undeclared = [], 0
    changes = feedback.moved(old, new)
    # Report leaves; parents merely propagate their children's failures.
    for (package, test), before, after in changes:
        if any(p == package and child.startswith(test + '/') for (p, child), _, _ in changes):
            continue
        row = metadata[(package, test)]
        key = MODULE + package + ' ' + test
        trailer_declared = not feedback.undeclared(key, trailers)
        file_declared = not feedback.undeclared(key, file)
        declared = trailer_declared or file_declared
        undeclared += not declared
        declaration = ('Moved-result trailer' if trailer_declared else
                       'moved-results.txt' if file_declared else 'undeclared')
        root = base_root if after == 'absent' else candidate_root
        for record in record_files(row, test, root):
            lines.append(dict(package=package, test=test, record=record, old=before, new=after,
                              declaration=declaration, refresh=row['refresh']))
    return dict(moved=lines, gaps=gaps, undeclared=undeclared)


def version(name, *args):
    result = command([name, *args], HERE)
    if result.returncode:
        raise RuntimeError(name + ' version unavailable: ' + result.stderr + result.stdout)
    return result.stdout.strip()


def cache_key(commit, go, node):
    content = (HERE / 'records-check-inventory.json').read_bytes()
    content += Path(__file__).read_bytes() + (HERE / 'darwin-feedback.py').read_bytes()
    # Host-dependent builds, environment oracle inputs and compiler versions must not cross caches.
    env = {k: v for k, v in os.environ.items() if k.startswith(('ADAMIC_', 'OPTIONAL_', 'GO', 'CC', 'CFLAGS'))}
    identity = [SCHEMA, commit, go, node, sys.platform, os.uname().machine,
                hashlib.sha256(content).hexdigest(), env]
    return hashlib.sha256(json.dumps(identity, sort_keys=True).encode()).hexdigest()


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('candidate', nargs='?', default='.', help='commit/ref or checkout path')
    parser.add_argument('--base', help='commit/ref or checkout path; default merge-base with origin/main')
    parser.add_argument('--repository', type=Path, default=HERE.parent)
    parser.add_argument('--cache', type=Path, default=Path.home() / '.cache/adamic-records')
    parser.add_argument('--jobs', type=int, default=len(os.sched_getaffinity(0)) if hasattr(os, 'sched_getaffinity') else os.cpu_count())
    parser.add_argument('--timeout', type=int, default=900)
    parser.add_argument('--json', action='store_true')
    args = parser.parse_args(argv)
    if args.jobs < 1 or args.timeout < 1:
        parser.error('jobs and timeout must be positive')
    repository = args.repository.resolve()
    started = time.monotonic()
    go, node = version('go', 'version'), version('node', '--version')
    with tree(repository, args.candidate) as (candidate_root, candidate_commit):
        candidate_commit = candidate_commit or git(candidate_root, 'rev-parse', 'HEAD')
        base_ref = args.base or git(repository, 'merge-base', candidate_commit, 'origin/main')
        # Only immutable commit inputs are cached. Supplied checkouts may be dirty.
        base_path = Path(base_ref).is_dir()
        base_commit = git(Path(base_ref), 'rev-parse', 'HEAD') if base_path else sha(repository, base_ref)
        args.cache.mkdir(parents=True, exist_ok=True)
        key = cache_key(base_commit, go, node)
        cached = args.cache / (key + '.json')
        with tree(repository, base_ref) as (base_root, _):
            with (args.cache / (key + '.lock')).open('w') as lock:
                fcntl.flock(lock, fcntl.LOCK_EX)
                hit = cached.exists() and not base_path
                if hit:
                    base = json.loads(cached.read_text())
                else:
                    base = run_tree(base_root, args.jobs, args.timeout)
                    # Retain the complete baseline, including named gaps; cache hits still report those gaps.
                    if not base_path:
                        partial = cached.with_suffix('.partial')
                        partial.write_text(json.dumps(base))
                        partial.replace(cached)
            candidate = run_tree(candidate_root, args.jobs, args.timeout)
            trailer, file = declarations(repository, base_commit, candidate_commit, candidate_root)
            report = compare(base, candidate, trailer, file, base_root, candidate_root)
            report.update(base=base_commit, candidate=candidate_commit, base_cached=hit,
                          seconds=time.monotonic() - started, cpus=args.jobs, go=go, node=node,
                          timings={side: {r['row']['package'] + '/' + r['row']['test']: r['seconds']
                                          for r in data['runs']}
                                   for side, data in [('base', base), ('candidate', candidate)]})
            if args.json:
                print(json.dumps(report, indent=2))
            else:
                for move in report['moved']:
                    print('{package} {test}\t{record}\t{old} -> {new}\t{declaration}\t{refresh}'.format(**move))
                for gap in report['gaps']:
                    reason = ' '.join(gap['reason'].split())[:500]
                    print('gap: {side} {package} {test}\t{records}\t{refresh}\t'.format(**gap) + reason)
                print('records-check: %.3fs; %d CPUs; base cache %s; %d undeclared moves; %d gaps' %
                      (report['seconds'], args.jobs, 'hit' if hit else 'miss', report['undeclared'], len(report['gaps'])), file=sys.stderr)
            return 2 if report['gaps'] else 1 if report['undeclared'] else 0


if __name__ == '__main__':
    try:
        sys.exit(main())
    except (RuntimeError, OSError, ValueError) as error:
        print('records-check gap: ' + str(error), file=sys.stderr)
        sys.exit(2)

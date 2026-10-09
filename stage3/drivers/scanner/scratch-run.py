#!/usr/bin/env python3
"""Fetch, merge and measure in a retained, never-pushed scratch worktree."""
import argparse
import hashlib
import importlib.util
import json
import os
from pathlib import Path
import re
import shutil
import signal
import subprocess
import sys
import time
import uuid

HERE = Path(__file__).resolve().parent
REPO = HERE.parents[2]
sys.dont_write_bytecode = True


def load_module(name, file):
    spec = importlib.util.spec_from_file_location(name, file)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('output', type=Path, help='new output directory, retained with scratch worktree')
    parser.add_argument('refs', nargs='*', help='refs merged into fetched origin/main in this exact order')
    args = parser.parse_args()
    out = args.output.resolve()
    out.mkdir(parents=True, exist_ok=False)
    summary = {'version': 1, 'status': 'failed', 'requested_refs': args.refs,
               'output': str(out), 'phases': [], 'merges': [], 'scanner': [], 'ordered_stops': []}
    started = time.monotonic()
    environment_file = None
    env = dict(os.environ, GOPROXY='https://proxy.golang.org|direct')
    cache = Path(env.get('STAGE3_CACHE', str(Path.home() / '.cache/adamic-stage3'))).resolve()
    env['STAGE3_CACHE'] = str(cache)

    def save():
        (out / 'summary.json').write_text(json.dumps(summary, indent=2) + '\n')

    def run(command, name, cwd=REPO, extra=None):
        begin = time.monotonic()
        command = [str(x) for x in command]
        execution = command
        if environment_file:
            execution = ['bash', '-c', 'source "$1"; shift; exec "$@"', 'scanner-tools', str(environment_file), *command]
        row = {'name': name, 'command': command, 'cwd': str(cwd), 'stdout': name + '.stdout', 'stderr': name + '.stderr'}
        with (out / row['stdout']).open('wb') as stdout, (out / row['stderr']).open('wb') as stderr:
            child = subprocess.Popen(execution, cwd=cwd, env=dict(env, **(extra or {})),
                                     stdout=stdout, stderr=stderr, start_new_session=True)
            try:
                code = child.wait(timeout=1800)
            except subprocess.TimeoutExpired:
                os.killpg(child.pid, signal.SIGTERM)
                child.wait()
                code = 124
        row.update(exit=code, wall_seconds=round(time.monotonic()-begin, 3))
        summary['phases'].append(row)
        save()
        print(f'{name}: exit {code}', flush=True)
        return code

    def require(command, name, cwd=REPO, extra=None):
        code = run(command, name, cwd, extra)
        if code != 0:
            raise RuntimeError(f'{name} exited {code}; see {name}.stderr/.stdout')

    def git_text(arguments, name, cwd=REPO):
        require(['git', *arguments], name, cwd)
        return (out / (name + '.stdout')).read_text().strip()

    try:
        # Fetch all heads, even when the caller checkout's fetch refspec is narrow.
        require(['git', 'fetch', 'origin', '+refs/heads/*:refs/remotes/origin/*'], 'fetch')
        base = git_text(['rev-parse', '--verify', 'origin/main^{commit}'], 'main-ref')
        resolved = []
        for i, ref in enumerate(args.refs):
            if ref.startswith('-'):
                raise ValueError('ref cannot start with an option prefix')
            resolved.append({'requested': ref, 'sha': git_text(['rev-parse', '--verify', ref + '^{commit}'], f'ref-{i:02}')})
        branch = 'scratch/scanner-run-' + uuid.uuid4().hex[:12]
        compiler_tree = out / 'compiler-tree'
        require(['git', 'worktree', 'add', '-b', branch, str(compiler_tree), base], 'worktree')
        summary.update(base=base, scratch_branch=branch, scratch_tree=str(compiler_tree), resolved_refs=resolved)
        for i, ref in enumerate(resolved):
            name = f'merge-{i:02}'
            code = run(['git', 'merge', '--no-edit', ref['sha']], name, compiler_tree)
            conflicts = []
            if code:
                conflicts = git_text(['diff', '--name-only', '--diff-filter=U'], name + '-conflicts', compiler_tree).splitlines()
            summary['merges'].append(dict(ref, exit=code, conflicts=conflicts))
            if code:
                if conflicts:
                    summary['status'] = 'merge-conflict'
                    require(['git', 'merge', '--abort'], name + '-abort', compiler_tree)
                    print('Conflicting paths:\n' + '\n'.join(conflicts), flush=True)
                    return 2
                raise RuntimeError('merge failed without conflict paths; see merge log')
        summary['integrated_sha'] = git_text(['rev-parse', 'HEAD'], 'integrated-ref', compiler_tree)
        # Isolated Git metadata borrows local objects, never shared submodule configuration.
        dependency = git_text(['ls-tree', 'HEAD', 'cohere'], 'cohere-pin', compiler_tree).split()[2]
        shared = REPO / 'cohere'
        def initialize_dependency(local, shared_tree, pin, remote, label):
            require(['git', 'init', str(local)], label + '-init-local')
            if (shared_tree / '.git').exists():
                shared_sha = git_text(['rev-parse', 'HEAD'], label + '-shared-ref', shared_tree)
                if shared_sha == pin:
                    shared_git = Path(git_text(['rev-parse', '--absolute-git-dir'], label + '-shared-git', shared_tree))
                    (local / '.git/objects/info/alternates').write_text(str(shared_git / 'objects') + '\n')
                else:
                    require(['git', 'fetch', '--depth=1', remote, pin], label + '-fetch', local)
            else:
                require(['git', 'fetch', '--depth=1', remote, pin], label + '-fetch', local)
            require(['git', 'checkout', '--detach', pin], label + '-checkout', local)

        initialize_dependency(compiler_tree / 'cohere', shared, dependency,
                              'https://github.com/system-inc/cohere.git', 'cohere')
        nested_pin = git_text(['ls-tree', 'HEAD', 'TypeScript'], 'cohere-typescript-pin', compiler_tree / 'cohere').split()[2]
        initialize_dependency(compiler_tree / 'cohere/TypeScript', shared / 'TypeScript', nested_pin,
                              'https://github.com/system-inc/TypeScript.git', 'cohere-typescript')
        summary['cohere_sha'] = dependency
        require(['bash', 'cloud/setup.sh'], 'setup', compiler_tree)
        setup = (out / 'setup.stdout').read_text()
        match = re.search(r'^setup: source (.+)$', setup, re.MULTILINE)
        if not match:
            raise RuntimeError('setup did not report its environment file')
        environment_file = Path(match.group(1))
        summary['toolchain_environment'] = str(environment_file)
        summary['setup_timing_lines'] = [line for line in setup.splitlines() if line.startswith('setup:')]
        require(['nproc'], 'nproc', compiler_tree)
        summary['nproc'] = int((out / 'nproc.stdout').read_text())
        require(['npm', 'ci', '--prefix', str(compiler_tree / 'stage3/api'), '--ignore-scripts', '--no-audit', '--no-fund'], 'compiler-node-types')
        compiler = out / 'adamic'
        code = run(['go', 'build', '-buildvcs=false', '-o', compiler, './cmd/adamic'], 'compiler-build', compiler_tree)
        summary['compiler'] = {'exit': code, 'sha': summary['integrated_sha']}
        if code:
            summary['status'] = 'compiler-failed'
            return 1
        summary['compiler']['binary_sha256'] = hashlib.sha256(compiler.read_bytes()).hexdigest()
        # Use the driver's documented small apply profile; no global adaptations are edited.
        pipeline = out / 'pipeline'
        pipeline.mkdir()
        for name in ['apply.sh', 'apply.py', 'source.json']:
            shutil.copyfile(REPO / 'stage3' / name, pipeline / name)
        shutil.copytree(REPO / 'stage3/api', pipeline / 'api', ignore=shutil.ignore_patterns('node_modules'))
        (pipeline / 'adapt').mkdir()
        for name in ['00-setup', '10-type-imports', '42-scanner-any', '50-temporary-scanner-implicit-returns', '51-temporary-scanner-fallthrough']:
            shutil.copytree(REPO / 'stage3/adapt' / name, pipeline / 'adapt' / name)
        full = out / 'full-tree'
        require(['bash', pipeline / 'apply.sh', full], 'apply')
        # A fixed pre-50 corpus reproduces the measured 509,014-token reference.
        inputs = out / 'inputs'
        shutil.copytree(full / 'src/compiler', inputs / 'src/compiler')
        require(['git', '-C', full, 'show', 'HEAD:src/compiler/scanner.ts'], 'corpus-scanner')
        (inputs / 'src/compiler/scanner.ts').write_bytes((out / 'corpus-scanner.stdout').read_bytes())
        # Reapply type-import rewriting to that one pristine corpus source, as in the measured corpus.
        corpus_adapter = HERE / 'scratch-corpus.cjs'
        require(['node', corpus_adapter, inputs], 'corpus-type-imports', extra={'NODE_PATH': str(cache / 'api/node_modules')})
        reference = out / 'reference'
        require(['bash', HERE / 'run.sh', reference, '--tree', full, '--inputs', inputs, '--node-only'], 'reference')
        raw_slice = out / 'slice'
        api = cache / 'api/node_modules/typescript/lib/typescript.js'
        require(['bash', REPO / 'stage3/slice/run.sh', full, raw_slice,
                 'src/compiler/scanner.ts:createScanner', 'src/compiler/types.ts:ScriptTarget',
                 'src/compiler/types.ts:SyntaxKind'], 'slice', extra={'SLICE_TYPESCRIPT': str(api)})
        failed_modes = []
        for split in [0, 1]:
            directory = out / ('split-' + str(split))
            code = run(['bash', HERE / 'run.sh', directory, '--tree', raw_slice, '--inputs', inputs,
                        '--oracle', reference / 'node.stdout', '--compiler', compiler, '--compiler-cwd', compiler_tree],
                       'scanner-' + str(split), extra={'ADAMIC_NATIVE_SPLIT': str(split), 'ADAMIC_NATIVE_JOBS': str(summary['nproc'])})
            report = json.loads((directory / 'report.json').read_text()) if (directory / 'report.json').exists() else None
            diff = run(['diff', '-u', reference / 'node.stdout', directory / 'node.stdout'], 'full-tree-check-' + str(split)) if (directory / 'node.stdout').exists() else None
            summary['scanner'].append({'split': split, 'exit': code, 'report': report, 'full_tree_diff_exit': diff})
            if code:
                failed_modes.append(split)
            elif report:
                require(['python3', HERE / 'measure.py', directory], 'measurement-' + str(split))
        if not failed_modes:
            summary['status'] = 'pass'
            return 0
        if any(not mode['report'] or mode['report'].get('failure') or mode['full_tree_diff_exit'] != 0 for mode in summary['scanner']):
            raise RuntimeError('scanner comparison/prerequisite failure; no discovery walk')
        summary['status'] = 'scanner-blocked'
        discovery = load_module('scanner_discovery', HERE / 'scratch-stops.py')
        summary['ordered_stops'] = discovery.walk(out, raw_slice, compiler, compiler_tree, cache, run, failed_modes[0])
        return 1
    except (OSError, ValueError, RuntimeError, subprocess.SubprocessError) as error:
        summary['error'] = str(error)
        if summary['status'] not in {'merge-conflict', 'compiler-failed', 'scanner-blocked'}:
            summary['status'] = 'failed'
        print(str(error), file=sys.stderr, flush=True)
        return 1
    finally:
        summary['wall_seconds'] = round(time.monotonic()-started, 3)
        checker = load_module('scanner_summary', HERE / 'scratch-summary.py')
        errors = checker.validate(summary)
        summary['validation'] = {'valid': not errors, 'errors': errors}
        if errors:
            summary['status'] = 'failed'
            summary['error'] = 'summary validation failed: ' + '; '.join(errors)
        save()
        print(out / 'summary.json', flush=True)
        if errors:
            return 1


if __name__ == '__main__':
    raise SystemExit(main())

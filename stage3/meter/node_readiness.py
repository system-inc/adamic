#!/usr/bin/env python3
"""Probe root order with TypeScript's ES-module emit and the installed pinned Node."""
import argparse
import json
from pathlib import Path
import re
import subprocess
import tempfile


# Mirror compilerOptions() in internal/load/load.go. All checks share that loader's
# ES2024, ESNext/bundler, forced-module graph, including value imports used only
# as types. In particular, verbatimModuleSyntax also preserves const enums.
LOADER_OPTIONS = {
    'strict': True, 'noUncheckedIndexedAccess': True, 'exactOptionalPropertyTypes': True,
    'erasableSyntaxOnly': False, 'verbatimModuleSyntax': True,
    'allowImportingTsExtensions': True, 'noEmit': True, 'module': 'ESNext',
    'moduleDetection': 'force', 'moduleResolution': 'bundler', 'target': 'ES2024',
    'lib': ['es2024'], 'types': [],
}


def command(args, log, timeout):
    with log.open('w') as output:
        try:
            result = subprocess.run(args, stdout=output, stderr=output, timeout=timeout)
            return result.returncode, False
        except subprocess.TimeoutExpired:
            return None, True
        except OSError as error:
            output.write(str(error) + '\n')
            return None, False


def run_roots(tree, roots, logs, timeout=5, build_timeout=60):
    """Build once, run each root separately. Tool errors and timeouts never prove a throw."""
    tree, logs = Path(tree).resolve(), Path(logs)
    roots = sorted({str(Path(root).resolve()) for root in roots})
    if not roots:
        return {}
    logs.mkdir(parents=True, exist_ok=True)
    relative = {root: Path(root).relative_to(tree).with_suffix('.js') for root in roots}
    version_exit, _ = command(['node', '--version'], logs / 'node-version.log', timeout)
    version = (logs / 'node-version.log').read_text().strip() if version_exit == 0 else None
    with tempfile.TemporaryDirectory(prefix='meter-node-') as temporary:
        output = Path(temporary)
        # The census already checked the program. Override only emission settings:
        # noCheck avoids a second checker; rewriting TS extensions lets emitted JS
        # resolve those imports. Neither setting elides a module evaluation edge.
        options = dict(LOADER_OPTIONS, noEmit=False, noCheck=True,
                       rewriteRelativeImportExtensions=True, sourceMap=True,
                       rootDir=str(tree), outDir=str(output))
        config = output / 'tsconfig.json'
        config.write_text(json.dumps({'compilerOptions': options, 'files': roots}))
        build_exit, build_timeout_hit = command(
            ['tsc', '--project', str(config)], logs / 'build.log', build_timeout)
        (output / 'package.json').write_text('{"type":"module"}\n')
        results = {}
        for index, root in enumerate(roots):
            result = {'node_version': version, 'build_exit': build_exit, 'exit': None,
                      'timeout': build_timeout_hit, 'first_error': '', 'throws': False}
            if build_exit == 0 and version and version.startswith('v24.'):
                log = logs / f'{index}.log'
                code, timed_out = command(['node', '--enable-source-maps', str(output / relative[root])],
                                          log, timeout)
                error = next((line.strip() for line in log.read_text().splitlines()
                              if re.match(r'^(?:[A-Za-z]*Error|Error):', line)), '')
                result.update(exit=code, timeout=timed_out, first_error=error,
                              throws=bool(code is not None and code > 0 and error and not timed_out))
            else:
                result['first_error'] = ('Node v24 unavailable' if not version or not version.startswith('v24.')
                                         else 'TypeScript emit failed or timed out')
            results[root] = result
    (logs / 'compiler-options.json').write_text(json.dumps(options, indent=2) + '\n')
    (logs / 'results.json').write_text(json.dumps(results, indent=2) + '\n')
    return results


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('tree', type=Path)
    parser.add_argument('logs', type=Path)
    parser.add_argument('roots', nargs='+')
    args = parser.parse_args()
    print(json.dumps(run_roots(args.tree, args.roots, args.logs), indent=2))

#!/usr/bin/env python3
"""Probe root order with TypeScript's ES-module emit and the installed pinned Node."""
import argparse
import json
from pathlib import Path
import re
import subprocess
import tempfile


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
        # Do not bundle or preserve const enums: use tsc's own import elision and module order.
        build_exit, build_timeout_hit = command([
            'tsc', '--ignoreConfig', '--noCheck', '--module', 'esnext', '--moduleResolution', 'bundler',
            '--target', 'es2022', '--sourceMap', '--rootDir', str(tree), '--outDir', str(output),
            *roots], logs / 'build.log', build_timeout)
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
    (logs / 'results.json').write_text(json.dumps(results, indent=2) + '\n')
    return results


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('tree', type=Path)
    parser.add_argument('logs', type=Path)
    parser.add_argument('roots', nargs='+')
    args = parser.parse_args()
    print(json.dumps(run_roots(args.tree, args.roots, args.logs), indent=2))

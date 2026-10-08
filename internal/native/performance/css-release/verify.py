#!/usr/bin/env python3
"""Check full CSS bytes, empty stderr, ASan/UBSan/LSan and unchanged workload counts."""
import argparse
import os
from pathlib import Path
import subprocess


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('snapshots', type=Path)
    parser.add_argument('css_fixture_directory', type=Path)
    parser.add_argument('--final-only', action='store_true')
    args = parser.parse_args()
    root = args.snapshots.resolve()
    corpus = args.css_fixture_directory.resolve()
    repo = Path(__file__).resolve().parents[4]
    commands = {'final': [str(root / 'final/css')], 'sanitized': [str(root / 'final/css-sanitized')]}
    if not args.final_only:
        commands['baseline'] = [str(root / 'baseline/css')]
        commands['Node'] = ['node', '--disable-warning=ExperimentalWarning', str(repo / 'oracle/node.mjs'), str(corpus / 'source/css/print_main.ts')]
    for name, command in commands.items():
        for narrow in (False, True):
            stem = root / ('verified-' + name + ('-narrow' if narrow else ''))
            with Path(str(stem) + '.stdout').open('wb') as out, Path(str(stem) + '.stderr').open('wb') as err:
                subprocess.run(command + [str(corpus / 'cases.txt'), 'print', 'once', 'narrow' if narrow else 'default'], stdout=out, stderr=err, check=True, env=dict(os.environ, ASAN_OPTIONS='detect_leaks=1', UBSAN_OPTIONS='halt_on_error=1'))
            answer = Path(str(stem) + '.stdout').read_bytes()
            if Path(str(stem) + '.stderr').read_bytes() or answer != (corpus / ('narrow-expected.txt' if narrow else 'expected.txt')).read_bytes():
                raise RuntimeError(str(stem) + ' bytes/stderr differ')
            print(name, 'narrow' if narrow else 'default', len(answer), 'bytes identical', flush=True)
    for workload in ('css', 'scanner'):
        answers = []
        for side in ('baseline', 'final'):
            binary = root / side / ('css-counted' if workload == 'css' else 'scanner/scanner-counted')
            arguments = [str(corpus / 'sample.txt'), 'count'] if workload == 'css' else ['--manifest', str(root / side / 'scanner/compiler.txt'), '--count']
            result = subprocess.run([str(binary)] + arguments, capture_output=True, check=True)
            stem = root / ('verified-counts-' + side + '-' + workload)
            Path(str(stem) + '.stdout').write_bytes(result.stdout)
            Path(str(stem) + '.stderr').write_bytes(result.stderr)
            answers.append((result.stdout, result.stderr))
            print(workload, side, result.stdout.decode().strip(), result.stderr.decode().strip(), flush=True)
        if answers[0] != answers[1]:
            raise RuntimeError(workload + ' counts differ')


if __name__ == '__main__':
    main()

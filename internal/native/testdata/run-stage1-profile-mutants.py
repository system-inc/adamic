#!/usr/bin/env python3
"""Prove source-identity and profile-binding checks with compiling Go overlays."""
import argparse
import json
from pathlib import Path
import subprocess

REPO = Path(__file__).resolve().parents[3]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--work', type=Path, required=True)
    parser.add_argument('--only', help='run one named mutant')
    args = parser.parse_args()
    work = args.work.resolve()
    work.mkdir(parents=True, exist_ok=True)
    binding = REPO / 'internal/native/stage1_profile.go'
    build = REPO / 'internal/native/native.go'
    mutations = [('main-source-path', build, 'command.Args[index] = "main.c"',
                  'command.Args[index] = argument', 'emitted profile source identity changed')]
    mutations += [
        ('main-use-source-path', build, 'command.Args[index] = "main.c"',
         'if options.Profile != "" { command.Args[index] = argument } else { command.Args[index] = "main.c" }',
         'emitted profile source identity changed'),
        ('runtime-source-path', REPO / 'internal/native/library.go',
         'command.Args[len(flags)+2] = file.name',
         'if file.name == "dtoa.c" { command.Args[len(flags)+2] = filepath.Join(temporary, file.name) } else { command.Args[len(flags)+2] = file.name }',
         'runtime profile source identity changed or duplicated'),
    ]
    guards = {
        'manifest-version': 'record.Version == current.Version',
        'emitted-units': 'slices.Equal(record.Units, current.Units)',
        'runtime-bytes': 'record.Runtime == current.Runtime',
        'profile-bytes': 'record.Profile == current.Profile',
        'compiler-bytes': 'record.CompilerSHA256 == current.CompilerSHA256',
        'compiler-version': 'record.Compiler == current.Compiler',
        'compile-flags': 'slices.Equal(record.Flags, current.Flags)',
        'link-flags': 'slices.Equal(record.LinkFlags, current.LinkFlags)',
        'training-compile-flags': 'slices.Equal(record.TrainingFlags, current.TrainingFlags)',
        'training-link-flags': 'slices.Equal(record.TrainingLinkFlags, current.TrainingLinkFlags)',
        'target': 'record.Target == current.Target',
    }
    mutations += [(name, binding, expression, 'true',
                   'stale/profile flag reached wrong clang invocation')
                  for name, expression in guards.items()]
    mutations.append(('runtime-fingerprint-file', binding, 'for _, f := range files {',
                      'for _, f := range files { if f.name == "ieee754.c" { continue }',
                      'changed runtime byte did not invalidate ieee754.c'))
    if args.only:
        mutations = [m for m in mutations if m[0] == args.only]
        if not mutations:
            raise RuntimeError('unknown mutant: ' + args.only)
    for name, source, before, after, catcher in mutations:
        text = source.read_text()
        if text.count(before) != 1:
            raise RuntimeError('mutation site must be unique: ' + name)
        changed = work / (name + '.go')
        changed.write_text(text.replace(before, after, 1))
        overlay = work / (name + '.json')
        overlay.write_text(json.dumps({'Replace': {str(source): str(changed)}}))
        log = work / (name + '.log')
        with log.open('wb') as output:
            result = subprocess.run(['go', 'test', '-overlay=' + str(overlay),
                '-v', '-count=1', '-timeout', '2m', './internal/native',
                '-run', '^Test(Stage1ProfileStalenessAndDeterminism|RuntimeFingerprintCoversEveryFile)$'], cwd=REPO,
                stdout=output, stderr=subprocess.STDOUT, timeout=150)
        body = log.read_text()
        if result.returncode == 0 or catcher not in body or '[build failed]' in body:
            raise RuntimeError('mutant survived or wrong failure: ' + name + '; see ' + str(log))
        print(name + ': caught by ' + catcher, flush=True)


if __name__ == '__main__':
    main()

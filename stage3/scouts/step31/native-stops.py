#!/usr/bin/env python3
"""Build gathered component roots and preserve the first stop without claiming runtime parity."""
import argparse
import hashlib
import json
from pathlib import Path
import subprocess
import time

parser = argparse.ArgumentParser()
parser.add_argument('compiler', type=Path)
parser.add_argument('label', choices=['main', 'train'])
parser.add_argument('slices', type=Path, help='parent containing LABEL-PIECE-slice directories')
parser.add_argument('prefix', help='slice directory prefix, such as step31-train')
parser.add_argument('output', type=Path)
args = parser.parse_args()
args.output.mkdir(parents=True, exist_ok=False)
compiler = args.compiler.resolve()
records = []
for piece in ['binder', 'checker', 'emitter']:
    tree = args.slices / (args.prefix + '-' + piece + '-slice')
    entry = tree / 'native-entry.a'
    if not entry.exists():
        entry.write_text('import type {} from "node:util";\nimport "./slice-entry.a";\n')
    manifest = json.loads((tree / 'slice.json').read_text())
    for profile, source in [('raw', tree / 'slice-entry.a'), ('node-declarations', entry)]:
        directory = args.output / (piece + '-' + profile)
        directory.mkdir()
        binary = directory / 'component.bin'
        command = [str(compiler), 'build', str(source.resolve()), '-o', str(binary.resolve())]
        (directory / 'command.json').write_text(json.dumps(command) + '\n')
        started = time.perf_counter()
        try:
            with (directory / 'stdout').open('wb') as out, (directory / 'stderr').open('wb') as err:
                result = subprocess.run(command, stdout=out, stderr=err, timeout=180)
            code, timed_out = result.returncode, False
        except subprocess.TimeoutExpired:
            code, timed_out = 124, True
        text = (directory / 'stderr').read_text()
        (directory / 'exit').write_text(str(code) + '\n')
        row = {'piece': piece, 'compiler': args.label, 'profile': profile, 'command': command,
            'exit': code, 'timeout': timed_out, 'seconds': time.perf_counter() - started,
            'first_stop': next((line for line in text.split('\n') if line.strip()), None),
            'diagnostic_lines': len([line for line in text.split('\n') if line]),
            'binary_exists': binary.exists(), 'entry_sha256': hashlib.sha256(source.read_bytes()).hexdigest(),
            'slice_sha256': hashlib.sha256((tree / 'slice.json').read_bytes()).hexdigest(),
            'slice_summary': manifest['summary'],
            'status': 'pending' if code == 0 else 'blocked', 'native_differential_run': False}
        if code == 0 and not binary.is_file():
            raise RuntimeError('successful build produced no binary')
        (directory / 'report.json').write_text(json.dumps(row, indent=2) + '\n')
        records.append(row)
        print(json.dumps({key: row[key] for key in ['piece', 'compiler', 'profile', 'exit', 'first_stop', 'status']}), flush=True)
report = {'compiler_sha256': hashlib.sha256(compiler.read_bytes()).hexdigest(), 'results': records}
(args.output / 'report.json').write_text(json.dumps(report, indent=2) + '\n')

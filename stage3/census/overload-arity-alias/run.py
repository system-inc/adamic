#!/usr/bin/env python3
"""Run from the repository root; compiler binary is an explicit argument."""
import json, os, pathlib, subprocess, sys
root = pathlib.Path('stage3/census/overload-arity-alias')
out = root / 'outputs'
out.mkdir(exist_ok=True)
binroot = pathlib.Path('/workspace/scratch/overload-arity-alias-binaries')
binroot.mkdir(exist_ok=True)
def run(argv, stem):
    proc = subprocess.run(argv, capture_output=True)
    (out / (stem + '.stdout')).write_bytes(proc.stdout)
    (out / (stem + '.stderr')).write_bytes(proc.stderr)
    (out / (stem + '.exit')).write_text(str(proc.returncode) + '\n')
    return proc
rows = []
for source in sorted((root / 'programs').glob('*.ts')):
    name = source.stem
    node = run(['node', '--no-warnings', str(source)], name + '.node')
    binary = binroot / name
    build = run([sys.argv[1], 'build', str(source), '-o', str(binary), '--explain-checks'], name + '.build')
    native = run([str(binary)], name + '.native') if build.returncode == 0 else None
    row = dict(program=source.name, node_exit=node.returncode, build_exit=build.returncode,
               native_exit=native.returncode if native else None,
               stdout_equal=(node.stdout == native.stdout) if native else None,
               stderr_equal=(node.stderr == native.stderr) if native else None)
    rows.append(row)
    print(json.dumps(row), flush=True)
(out / 'summary.json').write_text(json.dumps(rows, indent=2) + '\n')

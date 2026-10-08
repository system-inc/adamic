#!/usr/bin/env python3
"""Remove individual runtime guards, run package checks and new witnesses.
Run after all unmutated tests finish; mutations are sequential and restored.
"""
import json
import os
from pathlib import Path
import subprocess

root = Path.cwd()
unit = root / 'notes/coverage-oct8/nodehost'
logs = Path('/tmp/nodehost-mutants')
logs.mkdir(exist_ok=True)
pattern = 'TestNodeFSFile|TestNodeBuffer|TestNativeAgreesWithNode/internal/oracle/testdata/(node_buffer_|closure_convention_host24)'
mutants = [
    ('read-zero-return', 'internal/native/runtime/node_fs_file.c', 'if (length == 0) { return 0; }', '', 'fs_zero_read.a'),
    ('owned-buffer-close', 'internal/native/runtime/node_fs_file.c', 'if (owned) { close(descriptor); }', '', 'fs_fd_ownership.a'),
    ('cwd-invalidate', 'internal/native/runtime/node_host.c', 'if (current_directory != NULL) { adamic_release(current_directory); current_directory = NULL; }', '', 'host_directory.a'),
    ('short-write-loop', 'internal/native/runtime/node_fs_file.c', 'used += (size_t)count;', 'used += (size_t)count; break;', 'fs_large_roundtrip.a'),
    ('utf16-pair-guard', 'internal/native/runtime/node_buffer.c', 'if (low >= 0xdc00 && low <= 0xdfff) {', 'if (false) {', 'buffer_windows.a'),
]
summary = {}
for name, file, before, after, program in mutants:
    path = root / file
    original = path.read_text()
    # Isolate owned-buffer-close from the text reader's identical guard.
    start = original.index('static adamic_array *read_buffer(') if name == 'owned-buffer-close' else original.index('static double write_data(') if name == 'short-write-loop' else 0
    at = original.index(before, start)
    try:
        path.write_text(original[:at] + after + original[at + len(before):])
        with (logs / (name + '.packages.log')).open('wb') as log:
            result = subprocess.run(['go', 'test', './internal/lower', './internal/native', './internal/oracle', '-run', pattern, '-count=1', '-timeout', '30m', '-v'], env=dict(os.environ, ADAMIC_GATE_UNCACHED='1'), stdout=log, stderr=subprocess.STDOUT)
        compiler = logs / (name + '.compiler')
        with (logs / (name + '.compiler.log')).open('wb') as log:
            built = subprocess.run(['go', 'build', '-o', str(compiler), './cmd/adamic'], stdout=log, stderr=subprocess.STDOUT)
        if built.returncode != 0:
            raise RuntimeError('mutant compiler did not build: ' + name)
        with (logs / (name + '.witness.log')).open('wb') as log:
            subprocess.run(['python3', str(unit / 'run.py'), '--compiler', str(compiler), '--logs', str(logs / name), *(['--io-shim', '/tmp/nodehost-short-io.so'] if name == 'short-write-loop' else []), program], stdout=log, stderr=subprocess.STDOUT, check=True)
        observed = json.loads((logs / name / 'results.json').read_text())
        summary[name] = {'file': file, 'removed': before, 'replacement': after, 'packages_exit': result.returncode, 'witness': observed}
        print(name, 'packages', result.returncode, 'new-agrees', observed[Path(program).stem]['agree'], flush=True)
        (logs / 'summary.json').write_text(json.dumps(summary, indent=2) + '\n')
    finally:
        path.write_text(original)

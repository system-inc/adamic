#!/usr/bin/env python3
"""Prove both code-point refusal boundaries with compilable source overlays."""
import json
from pathlib import Path
import subprocess
import tempfile

root = Path(__file__).resolve().parents[1]
logs = Path('/tmp/adamic-catchability-source-mutants')
logs.mkdir(exist_ok=True)
mutants = [
    ('code-point-list-omitted', 'if node.CodePoints {', 'if false && node.CodePoints {', '9984394_lib_codepoint.a'),
    ('interface-method-targets-omitted', 'targets[method.Function] = true', 'targets[method.Function] = false', '9984394_lib_dispatch_codepoint.a'),
]
source_path = root / 'internal/lower/exceptions.go'
source = source_path.read_text()
with tempfile.TemporaryDirectory(prefix='adamic-catchability-overlay-') as directory:
    scratch = Path(directory)
    for name, old, new, fixture in mutants:
        if source.count(old) != 1:
            raise RuntimeError(f'{name}: expected exactly one source target')
        variant = scratch / f'{name}.go'
        variant.write_text(source.replace(old, new))
        overlay = scratch / f'{name}.json'
        overlay.write_text(json.dumps({'Replace': {str(source_path): str(variant)}}))
        log = logs / f'{name}.log'
        with log.open('w') as output:
            result = subprocess.run(['go', 'test', '-overlay', str(overlay), './internal/oracle',
                                     '-run', f'TestCodePointCatchabilityBoundary/{fixture}$', '-count=1', '-v'],
                                    cwd=root, stdout=output, stderr=subprocess.STDOUT)
        observation = log.read_text()
        caught = result.returncode != 0 and 'want String.fromCodePoint catchability refusal' in observation and '[build failed]' not in observation
        print(f'{name}: exit {result.returncode}, assertion caught={caught}', flush=True)
        if not caught:
            raise RuntimeError(f'{name} survived or failed before its assertion: {log}')

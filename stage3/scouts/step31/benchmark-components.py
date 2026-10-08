#!/usr/bin/env python3
"""Measure full, unchanged and one-project-change Node runs with process wall time."""
import argparse
import hashlib
import json
from pathlib import Path
import subprocess
import time

ROOT = Path(__file__).resolve().parent
parser = argparse.ArgumentParser()
parser.add_argument('request', type=Path)
parser.add_argument('output', type=Path)
args = parser.parse_args()
args.output.mkdir(parents=True, exist_ok=False)
request = json.loads(args.request.read_text())
assert len(request['projects']) == 6563
assert request['projects'][300]['id'] == 'tiny'
changed = json.loads(args.request.read_text())
changed['projects'][300]['files'][0]['text'] = '// changed input\n' + changed['projects'][300]['files'][0]['text']
changed_path = args.output / 'changed.json'
changed_path.write_text(json.dumps(changed) + '\n')
report = {'projects': 6563, 'acceptance': 301, 'upstream': 6262, 'runs': []}
for mode in ['checker', 'emitter']:
    for label, source, previous in [('full', args.request, None), ('unchanged', args.request, 'full'), ('one-change', changed_path, 'unchanged')]:
        output = args.output / (mode + '-' + label)
        command = ['node', str(ROOT / (mode + '-dump.cjs')), str(source), str(output)]
        if previous:
            command.extend(['--previous', str(args.output / (mode + '-' + previous)), '--changed-only'])
        started = time.perf_counter()
        with (args.output / (mode + '-' + label + '.stdout')).open('wb') as out, (args.output / (mode + '-' + label + '.stderr')).open('wb') as err:
            result = subprocess.run(command, stdout=out, stderr=err, timeout=180)
        wall = time.perf_counter() - started
        assert result.returncode == 0
        assert (args.output / (mode + '-' + label + '.stderr')).read_bytes() == b''
        manifest = json.loads((output / 'manifest.json').read_text())
        assert len(manifest['projects']) == 6563
        assert len(manifest['selected']) == {'full': 6563, 'unchanged': 0, 'one-change': 1}[label]
        if label == 'one-change':
            assert manifest['selected'] == ['tiny']
            original = json.loads((args.output / (mode + '-full/manifest.json')).read_text())['projects'][300]
            assert manifest['projects'][300]['sha256'] != original['sha256']
        rows = [json.loads(line) for line in (output / 'golden.stdout').read_text().split('\n') if line]
        row = {'mode': mode, 'run': label, 'command': command, 'wall_seconds': wall,
            'driver_seconds': manifest['milliseconds'] / 1000, 'selected': len(manifest['selected']),
            'reused': manifest['reused'], 'sha256': manifest['sha256'],
            'bytes': (output / 'golden.stdout').stat().st_size,
            'records': len(rows), 'diagnostics': sum(len(item.get('diagnostics', [])) for item in rows),
            'output_files': sum(item['record'] == 'output' for item in rows),
            'emit_skipped': sum(item.get('emitSkipped', False) for item in rows),
            'acceptance_compile_seconds': sum(item['milliseconds'] for item in manifest['projects'][:301]) / 1000,
            'upstream_compile_seconds': sum(item['milliseconds'] for item in manifest['projects'][301:]) / 1000}
        report['runs'].append(row)
        print(json.dumps(row), flush=True)
(args.output / 'report.json').write_text(json.dumps(report, indent=2) + '\n')

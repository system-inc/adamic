#!/usr/bin/env python3
"""Measure a pinned compiler against the unchanged tree and saved stop snapshots."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('checkout', type=Path)
parser.add_argument('compiler', type=Path)
parser.add_argument('tree', type=Path)
parser.add_argument('output', type=Path)
args = parser.parse_args()
here = Path(__file__).resolve().parent
baseline = here / 'evidence/main-efe9f404'
records = json.loads((baseline / 'stops.json').read_text())
provenance = json.loads((baseline / 'provenance.json').read_text())
tree = args.tree.resolve()
compiler = args.compiler.resolve()
out = args.output.resolve()
out.mkdir()

def hashes():
    return {file: hashlib.sha256((tree / file).read_bytes()).hexdigest()
            for file in provenance['sources']}

if hashes() != provenance['sources']:
    raise RuntimeError('adapted tree differs from the recorded main experiment')

def run(stem, command, split=None):
    env = dict(os.environ)
    if split is not None:
        env.update(ADAMIC_NATIVE_SPLIT=str(split),
                   ADAMIC_NATIVE_JOBS=subprocess.check_output(['nproc'], text=True).strip())
    with (out / (stem + '.stdout')).open('wb') as stdout, (out / (stem + '.stderr')).open('wb') as stderr:
        result = subprocess.run(command, env=env, stdout=stdout, stderr=stderr)
    (out / (stem + '.exit')).write_text(str(result.returncode) + '\n')
    return result.returncode, (out / (stem + '.stderr')).read_text()

def diagnostics_at(text, file, line, column):
    prefix = f'{file}:{line}:{column}: '
    return [row[len(prefix):] for row in text.splitlines() if row.startswith(prefix)]

commit = subprocess.check_output(['git', '-C', str(args.checkout), 'rev-parse', 'HEAD'], text=True).strip()
binary = subprocess.check_output(['go', 'version', '-m', str(compiler)], text=True)
if f'vcs.revision={commit}\n' not in binary or 'vcs.modified=false\n' not in binary:
    raise RuntimeError('compiler binary provenance differs from the clean pinned checkout')
(out / 'binary.log').write_text(binary)
submodules = subprocess.check_output(['git', '-C', str(args.checkout), 'submodule', 'status', '--recursive'], text=True)
(out / 'submodules.log').write_text(submodules)
summary = {'compiler_commit': commit, 'compiler_binary_sha256': hashlib.sha256(compiler.read_bytes()).hexdigest(),
           'tree': str(tree), 'baseline': str(baseline.relative_to(here)),
           'source_hashes': hashes(), 'node': subprocess.check_output(['node', '--version'], text=True).strip(),
           'nproc': int(subprocess.check_output(['nproc'], text=True)), 'stops': []}
streams = []
for split in (0, 1):
    code, text = run(f'build-{split}', [str(compiler), 'build', str(tree / 'src/tsc/tsc.ts'), '-o', str(out / f'tsc-{split}')], split)
    summary[f'split_{split}_exit'] = code
    streams.append(text)
summary['first_stop'] = streams[0].splitlines()[0] if streams[0] else None
summary['split_streams_equal'] = streams[0] == streams[1]
print(json.dumps({key: summary[key] for key in ('compiler_commit', 'first_stop', 'split_0_exit', 'split_1_exit')}), flush=True)
with tempfile.TemporaryDirectory(prefix='tsc-entry-tip-snapshots-') as temporary:
    scratch = Path(temporary) / 'src'
    shutil.copytree(tree / 'src', scratch)
    for record in records:
        ordinal = record['ordinal']
        snapshot_exit, snapshot_text = run(f'{ordinal:02d}-snapshot-types', [str(compiler), 'types', str(scratch / 'tsc/tsc.ts')])
        file = scratch / Path(record['file']).relative_to('src')
        snapshot_diagnostics = diagnostics_at(snapshot_text, file, record['line'], record['column'])
        pristine_diagnostics = diagnostics_at(streams[0], tree / record['file'], record['original_line'], record['original_column'])
        probe = here / record['probe']
        probe_exit, probe_text = run(f'{ordinal:02d}-probe-types', [str(compiler), 'types', str(probe)])
        build_exit, build_text = run(f'{ordinal:02d}-probe-build', [str(compiler), 'build', str(probe), '-o', str(out / f'probe-{ordinal:02d}')], 0)
        node_exit, _ = run(f'{ordinal:02d}-probe-node', ['node', str(here / 'node.mjs'), str(probe)])
        row = {'ordinal': ordinal, 'file': record['file'], 'line': record['original_line'], 'column': record['original_column'],
               'old_message': record['message'], 'probe': record['probe'],
               'pristine_diagnostics': pristine_diagnostics,
               'snapshot_diagnostics': snapshot_diagnostics, 'snapshot_types_exit': snapshot_exit,
               'status': 'remain' if record['message'] in snapshot_diagnostics else ('changed' if snapshot_diagnostics else 'disappear'),
               'placeholder_induced': not record['message_already_in_pristine_diagnostics'],
               'probe_types_exit': probe_exit, 'probe_build_exit': build_exit,
               'probe_first_stop': build_text.splitlines()[0] if build_text else None,
               'node_exit': node_exit, 'node_stdout': (out / f'{ordinal:02d}-probe-node.stdout').read_text()}
        summary['stops'].append(row)
        print(json.dumps({'ordinal': ordinal, 'status': row['status'], 'probe_types_exit': probe_exit}), flush=True)
        replacement = baseline / f'{ordinal:02d}-replacement.json'
        if replacement.exists():
            data = json.loads(replacement.read_text())
            raw = file.read_bytes().decode('utf8').encode('utf-16-le')
            start, end = data['start'] * 2, data['end'] * 2
            if raw[start:end].decode('utf-16-le') != data['original']:
                raise RuntimeError(f'{ordinal}: snapshot replacement does not match')
            file.write_bytes((raw[:start] + data['replacement'].encode('utf-16-le') + raw[end:]).decode('utf-16-le').encode('utf8'))
if hashes() != provenance['sources']:
    raise RuntimeError('original adapted tree changed during measurement')
(out / 'comparison.json').write_text(json.dumps(summary, indent=2) + '\n')

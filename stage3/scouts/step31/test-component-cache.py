#!/usr/bin/env python3
"""Exercise changed-only selection, cache integrity and native comparison fields."""
import argparse
import copy
import hashlib
import json
from pathlib import Path
import shutil
import subprocess
import sys

ROOT = Path(__file__).resolve().parent
parser = argparse.ArgumentParser()
parser.add_argument('output', type=Path)
args = parser.parse_args()
args.output.mkdir(parents=True, exist_ok=False)
request = ROOT / 'fixtures/components.json'

def run(command, name, expected=0, contains=None):
    log = args.output / (name + '.log')
    with log.open('wb') as handle:
        result = subprocess.run(command, stdout=handle, stderr=handle, timeout=120)
    assert result.returncode == expected, (name, result.returncode, log.read_text())
    if contains:
        assert contains in log.read_text(), (name, log.read_text())
    return log

for mode in ['checker', 'emitter']:
    driver = ROOT / (mode + '-dump.cjs')
    baseline = args.output / (mode + '-base')
    run(['node', str(driver), str(request), str(baseline)], mode + '-base')
    clean = args.output / (mode + '-clean')
    run(['node', str(driver), str(request), str(clean), '--previous', str(baseline), '--changed-only'], mode + '-clean')
    manifest = json.loads((clean / 'manifest.json').read_text())
    assert manifest['changed'] == 0 and manifest['reused'] == 2 and manifest['selected'] == []
    assert (clean / 'golden.stdout').read_bytes() == b''
    for before, after in zip(json.loads((baseline / 'manifest.json').read_text())['projects'], manifest['projects']):
        assert before['sha256'] == after['sha256']
    changed = json.loads(request.read_text())
    index = 0 if mode == 'checker' else 1
    file_index = 1 if mode == 'checker' else 0
    needle, replacement = ('"wrong"', '42') if mode == 'checker' else ('41 + 1', '40 + 2')
    changed['projects'][index]['files'][file_index]['text'] = changed['projects'][index]['files'][file_index]['text'].replace(needle, replacement)
    changed_request = args.output / (mode + '-changed.json')
    changed_request.write_text(json.dumps(changed))
    delta = args.output / (mode + '-delta')
    run(['node', str(driver), str(changed_request), str(delta), '--previous', str(clean), '--changed-only'], mode + '-delta')
    manifest = json.loads((delta / 'manifest.json').read_text())
    assert manifest['changed'] == 1 and manifest['reused'] == 1
    assert manifest['selected'] == [changed['projects'][index]['id']]
    assert len(json.loads((delta / 'request.json').read_text())['projects']) == 1
    fresh = args.output / (mode + '-fresh')
    ids = args.output / (mode + '-ids.json')
    ids.write_text(json.dumps(manifest['selected']))
    run(['node', str(driver), str(changed_request), str(fresh), '--ids', str(ids), '--fresh-libs'], mode + '-fresh')
    assert (delta / 'golden.stdout').read_bytes() == (fresh / 'golden.stdout').read_bytes()
    invalid = args.output / (mode + '-invalid-environment')
    shutil.copytree(baseline, invalid)
    old = json.loads((invalid / 'manifest.json').read_text())
    old['environment']['libraries'] = 'mutant'
    (invalid / 'manifest.json').write_text(json.dumps(old))
    new = args.output / (mode + '-invalidated')
    run(['node', str(driver), str(request), str(new), '--previous', str(invalid), '--changed-only'], mode + '-invalidated')
    assert json.loads((new / 'manifest.json').read_text())['changed'] == 2
    corrupted = args.output / (mode + '-corrupted')
    shutil.copytree(baseline, corrupted)
    shard = corrupted / old['projects'][0]['shard']
    shard.write_bytes(shard.read_bytes() + b'x')
    run(['node', str(driver), str(request), str(args.output / (mode + '-reject')), '--previous', str(corrupted)],
        mode + '-reject', 1, 'cached output hash mismatch')
    # Test doubles only: they verify the native protocol; no native implementation is claimed.
    echo = 'import pathlib,sys;sys.stdout.buffer.write(pathlib.Path(sys.argv[1]).read_bytes())'
    for label, source in [('baseline', baseline / 'golden.stdout'), ('delta', delta / 'golden.stdout')]:
        output = args.output / (mode + '-compare-' + label)
        run([sys.executable, str(ROOT / 'component-compare.py'), str(baseline if label == 'baseline' else delta), str(output), '--',
            sys.executable, '-c', echo, str(source)], mode + '-compare-' + label)
    rows = [json.loads(line) for line in (baseline / 'golden.stdout').read_text().split('\n') if line]
    fields = ['code', 'start', 'length', 'message-chain'] if mode == 'checker' else ['text-byte', 'missing-output', 'bom', 'emitSkipped']
    for field in fields:
        mutated = copy.deepcopy(rows)
        if mode == 'checker':
            diagnostic = next(row for row in mutated if row.get('diagnostics'))['diagnostics'][0]
            if field == 'message-chain':
                diagnostic['message']['next'][0]['next'][0]['message'] += ' mutant'
            else:
                diagnostic[field] += 1
        else:
            row = next(row for row in mutated if row.get('record') == 'output')
            if field == 'text-byte':
                row['text'] = 'X' + row['text'][1:]
            elif field == 'missing-output':
                mutated.remove(row)
            elif field == 'bom':
                row['bom'] = not row['bom']
            else:
                next(row for row in mutated if row.get('record') == 'result')['emitSkipped'] = True
        wire = args.output / (mode + '-' + field + '.stdout')
        wire.write_text(''.join(json.dumps(row, separators=(',', ':'), ensure_ascii=False) + '\n' for row in mutated))
        output = args.output / (mode + '-compare-' + field)
        run([sys.executable, str(ROOT / 'component-compare.py'), str(baseline), str(output), '--',
            sys.executable, '-c', echo, str(wire)], mode + '-compare-' + field, 1)
        report = json.loads((output / 'report.json').read_text())
        assert sorted(report['differences']) == ['stdout'] and not report['timeout']
        assert (output / 'actual.exit').read_text() == '0\n'
        assert (output / 'actual.stderr').read_bytes() == b''
        print(mode + ' ' + field + ': caught by bytes, exit 0, empty stderr')
    for name in ['golden.stdout', 'request.json']:
        broken = args.output / (mode + '-broken-' + name)
        shutil.copytree(baseline, broken)
        (broken / name).write_bytes((broken / name).read_bytes() + b'x')
        run([sys.executable, str(ROOT / 'component-compare.py'), str(broken), str(args.output / (mode + '-blocked-' + name)), '--',
            sys.executable, '-c', echo, str(baseline / 'golden.stdout')], mode + '-integrity-' + name, 1, 'cache integrity failure: ' + name)
    print(mode + ': unchanged selection empty; one changed project equals fresh; library identity and output/request integrity mutants caught')

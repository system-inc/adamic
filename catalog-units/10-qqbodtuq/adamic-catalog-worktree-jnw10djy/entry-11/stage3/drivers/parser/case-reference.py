#!/usr/bin/env python3
"""Hash single-file pinned compiler cases with the same Node/native dump driver."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess

p = argparse.ArgumentParser()
p.add_argument('checkout', type=Path)
p.add_argument('full', type=Path)
p.add_argument('slice', type=Path)
p.add_argument('output', type=Path)
p.add_argument('--native', type=Path)
p.add_argument('--reference', type=Path)
a = p.parse_args()
for name in ['checkout', 'full', 'slice', 'output']:
    setattr(a, name, getattr(a, name).resolve())
a.output.mkdir(exist_ok=False)
here = Path(__file__).resolve().parent
pin = json.loads((here.parents[1] / 'source.json').read_text())['commit']
assert subprocess.check_output(['git', '-C', str(a.checkout), 'rev-parse', 'HEAD'], text=True).strip() == pin
entries = subprocess.check_output(['git', '-C', str(a.checkout), 'ls-tree', '-r', pin,
                                  'tests/cases/compiler', 'tests/cases/conformance'], text=True).splitlines()
blobs = {line.split('\t', 1)[1]: line.split('\t', 1)[0].split()[2] for line in entries}
paths = list(blobs)
# No compiler option matrix: each raw single-file case is one input, Latest target.
# Metadata stays as comments. A sole @filename supplies its effective language suffix.
filename = re.compile(r'^//\s*@filename\s*:\s*([^\r\n]*)', re.M | re.I)
def bucket(name):
    for b in ['parser', 'jsx', 'salsa']:
        if name.startswith('tests/cases/conformance/' + b + '/'):
            return b
    return 'remaining'
order = {'parser': 0, 'jsx': 1, 'salsa': 2, 'remaining': 3}
paths.sort(key=lambda name: (order[bucket(name)], name))
inputs = a.output / 'inputs'
records, excluded = [], []
for name in paths:
    if Path(name).suffix not in ['.ts', '.tsx', '.js', '.jsx']:
        continue
    raw = (a.checkout / name).read_bytes()
    # Verify each source blob against the pin, including line endings and BOM.
    oid = hashlib.sha1(b'blob ' + str(len(raw)).encode() + b'\0' + raw).hexdigest()
    assert oid == blobs[name], name
    # Match ts.sys.readFile BOM handling before the parser sees source text.
    if raw.startswith(b'\xfe\xff'):
        text = raw[2:].decode('utf-16-be', 'replace')
    elif raw.startswith(b'\xff\xfe'):
        text = raw[2:].decode('utf-16-le', 'replace')
    else:
        text = raw.decode('utf-8-sig', 'replace')
    content = text.encode('utf-8')
    names = filename.findall(text)
    if len(names) > 1:
        excluded.append({'path': name, 'filename_directives': len(names), 'reason': 'multi-file case'})
        continue
    effective = names[0].strip() if names else Path(name).name
    suffix = Path(effective).suffix.lower()
    # Controlled path isolates cases and retains the effective parser mode.
    driver_path = name + '/input' + (suffix or '.ts')
    target = inputs / 'src/compiler' / driver_path
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_bytes(content)
    records.append({'path': name, 'effective_filename': effective, 'input_path': driver_path,
                    'source_sha256': hashlib.sha256(raw).hexdigest(), 'input_sha256': hashlib.sha256(content).hexdigest(), 'bucket': bucket(name)})
manifest = a.output / 'input-manifest'
manifest.write_text(''.join(r['input_path'] + '\n' for r in records))
def written(name):
    data = name.encode('utf-16-le', 'surrogatepass')
    return ''.join(chr(n) if 32 <= n <= 126 and n != 92 else '\\u' + format(n, '04x')
                   for n in (int.from_bytes(data[i:i+2], 'little') for i in range(0, len(data), 2)))
lookup = {written(r['input_path']): r for r in records}
def observations(dump):
    result = {}
    current = None
    with dump.open('rb') as stream:
        for line in stream:
            if line.startswith(b'file\t'):
                if current is not None:
                    row['sha256'] = digest.hexdigest()
                key = line[5:].decode('ascii').rstrip('\n')
                assert key in lookup and key not in result, key
                current = key
                row = {'bytes': 0, 'node_count': 0, 'diagnostic_rows': 0, 'jsdoc_diagnostic_rows': 0}
                result[key] = row
                digest = hashlib.sha256()
            assert current is not None, 'output before first file'
            digest.update(line)
            row['bytes'] += len(line)
            if line.startswith(b'diagnostic '): row['diagnostic_rows'] += 1
            elif line.startswith(b'jsDocDiagnostic '): row['jsdoc_diagnostic_rows'] += 1
            elif not line.startswith((b'file\t', b'diagnostics ', b'jsDocDiagnostics ')): row['node_count'] += 1
        if current is not None: row['sha256'] = digest.hexdigest()
    assert len(result) == len(records), (len(result), len(records))
    return result

env = dict(os.environ, PARSER_RUNTIME=str(here.parents[2] / 'oracle/adamic.mjs'))
assert env.get('PARSER_TYPESCRIPT'), 'set PARSER_TYPESCRIPT to stock 6.0.3'
def run(tree, label, native=None):
    shutil.copyfile(here / 'main.a', tree / 'parser-proof-main.a')
    shutil.copyfile(here / 'kinds.a', tree / 'kinds.a')
    command = [str(native.resolve()), str(inputs), str(manifest)] if native else [
        'node', '--disable-warning=ExperimentalWarning', str(here / 'node.mjs'),
        str(tree / 'parser-proof-main.a'), str(inputs), str(manifest)]
    dump, stderr = a.output / (label + '.dump'), a.output / (label + '.stderr')
    with dump.open('wb') as stdout, stderr.open('wb') as error:
        process = subprocess.run(command, env=env, stdout=stdout, stderr=error)
    assert process.returncode == 0 and stderr.stat().st_size == 0, (label, process.returncode)
    return observations(dump)
full = run(a.full, 'full-node')
sliced = run(a.slice, 'slice-node')
for r in records:
    key = written(r['input_path'])
    r['node'] = full[key]
    r['slice_node'] = sliced[key]
    r['identical'] = full[key] == sliced[key]
    assert r['identical'], r['path']
summary = {}
for b in ['parser', 'jsx', 'salsa', 'remaining', 'all']:
    rows = [r for r in records if b == 'all' or r['bucket'] == b]
    summary[b] = {'cases': len(rows), 'diagnostic_rows': sum(r['node']['diagnostic_rows'] for r in rows),
                  'jsdoc_diagnostic_rows': sum(r['node']['jsdoc_diagnostic_rows'] for r in rows),
                  'node_count': sum(r['node']['node_count'] for r in rows),
                  'cases_with_parse_diagnostics': sum(r['node']['diagnostic_rows'] > 0 for r in rows)}
report = {'schema': 'parser-preorder-with-jsdoc-v2', 'pinned_commit': pin,
          'selection': 'All tracked single-file compiler/conformance cases (zero or one @filename); source text retained after upstream BOM decoding, metadata comments preserved; sole filename selects language suffix; Latest target, no compiler-option permutations',
          'summary': summary, 'excluded_multi_file_cases': excluded, 'cases': records}
if a.reference:
    expected = json.loads(a.reference.read_text())
    assert expected['pinned_commit'] == pin and expected['cases'] == records
if a.native:
    native = run(a.slice, 'native', a.native)
    for r in records:
        assert native[written(r['input_path'])] == r['node'], ('native', r['path'])
    report['native_identical_cases'] = len(records)
(a.output / 'reference.json').write_text(json.dumps(report, indent=2) + '\n')
print(json.dumps(summary, indent=2))

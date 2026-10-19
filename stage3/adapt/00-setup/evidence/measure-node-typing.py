"""Measure the ordinary Adamic CLI, retaining its complete checker output."""
import collections
import gzip
import hashlib
import json
import pathlib
import re
import subprocess
import sys

HERE = pathlib.Path(__file__).resolve().parent
REPO = HERE.parents[3]
CLASSES = {'node-sys-ts2591', 'node-host-surface', 'optional-host-module'}


def digest(data):
    return hashlib.sha256(data).hexdigest()


def identity(d):
    return 'D' + digest(json.dumps([d[k] for k in ('file', 'line', 'column', 'code', 'message')]).encode())[:12]


def parse(text, tree):
    result = []
    for line in text.splitlines():
        match = re.fullmatch(r'(.+?):(\d+):(\d+): error (TS\d+): (.*)', line)
        if match:
            file, row, col, code, message = match.groups()
            file = file.removeprefix(str(tree) + '/')
            result.append(dict(file=file, line=int(row), column=int(col), code=code, message=message))
        elif line.startswith('  ') and result:
            result[-1]['message'] += '\n' + line
        else:
            raise AssertionError('Unparsed checker output: ' + line)
    for d in result:
        d['id'] = identity(d)
    return result


def compare(rows, actual):
    ids = {d['id'] for d in actual}
    results = []
    for row in rows:
        same = [d for d in actual if (d['file'], d['line'], d['column']) == (row['file'], row['line'], row['column'])]
        status = 'reported' if row['id'] in ids else 'changed' if same else 'absent'
        name = re.search(r"Cannot find (?:name|namespace|module) '([^']+)'", row['message'])
        if name:
            symbol = name[1]
        elif 'captureStackTrace' in row['message']:
            symbol = 'ErrorConstructor.captureStackTrace'
        elif 'stackTraceLimit' in row['message']:
            symbol = 'ErrorConstructor.stackTraceLimit'
        elif row['code'] in ('TS7006', 'TS7031'):
            symbol = 'inspector.Session.post callback (err, profile)'
        else:
            symbol = 'process.versions contextual typing at Object.entries'
        third_party = symbol == 'source-map-support'
        results.append(dict(stock=row, status=status, adamicDiagnostics=same,
                            owner=dict(category='other-host-dependency' if third_party else 'node-library-activation',
                                       name=symbol,
                                       reason='Third-party module, outside bundled @types/node; adaptation 65 host-dependency territory.' if third_party else
                                       'Bundled @types/node is present but not activated for require/import-type-only use. Loader activation gap; adaptation 65 is the temporary host-view territory, not a missing published Node declaration.')))
    return results


def main():
    binary, tree, scratch = map(pathlib.Path, sys.argv[1:])
    binary, tree, scratch = binary.resolve(), tree.resolve(), scratch.resolve()
    scratch.mkdir(parents=True, exist_ok=True)
    stock = json.loads((HERE / 'checker-stops-2026-10-19.json').read_text())
    accepted = json.loads(gzip.decompress((HERE / 'optional-contracts-wave2/after.json.gz').read_bytes()))
    roots = accepted['roots']
    assert len(roots) == 79 and roots == stock['roots']
    hashes = {d['file']: d['sha256'] for d in accepted['runs'][0]['sourceFiles']}
    root_hashes = {r: digest((tree / r).read_bytes()) for r in roots}
    assert all(root_hashes[r] == hashes[r] for r in roots)
    rows = [d for d in stock['diagnostics'] if d['classification'] in CLASSES]
    assert len(rows) == 62
    command = [str(binary), 'types'] + [str(tree / r) for r in roots]
    def run(args, name):
        with (scratch / name).open('w') as log:
            result = subprocess.run(args, stdout=log, stderr=subprocess.STDOUT, timeout=180, cwd=REPO)
        assert result.returncode == 1
        text = (scratch / name).read_text()
        return dict(command=args, exit=result.returncode, log=text, logSha256=digest(text.encode()), diagnostics=parse(text, tree))
    run_result = run(command, 'checker.log')
    comparison = compare(rows, run_result['diagnostics'])
    files = sorted({d['file'] for d in rows})
    focused = [d for d in run_result['diagnostics'] if d['file'] in files]
    stock_ids = {identity(d) for d in accepted['runs'][0]['diagnostics']}
    additional = [d for d in focused if d['id'] not in stock_ids]
    trigger = scratch / 'node-trigger.a'
    trigger.write_text('import "node:fs";\n')
    control = run(command + [str(trigger)], 'node-trigger.log')
    control_comparison = compare(rows, control['diagnostics'])
    # Actual checker control must invalidate the assertion that all 62 survive.
    assert all(d['status'] == 'reported' for d in comparison)
    assert any(d['status'] != 'reported' for d in control_comparison)
    # Completeness mutant: a dropped real finding must be detected by comparison.
    omitted = compare(rows, [d for d in run_result['diagnostics'] if d['id'] != rows[0]['id']])
    assert omitted[0]['status'] == 'absent'
    output = dict(schema=1, date='2026-10-19', base='eceb12092c3dfebe81000bb792cbcbf8908744e1',
                  scope='Measurement only. Ordinary CLI load.Load with unchanged options and prelude; no lowering.',
                  comparisonIdentity='Original census SHA256 id over file, UTF-16 line/column, code, full message. Changed means a different diagnostic at the same exact location.',
                  roots=roots, rootSha256=root_hashes, binarySha256=digest(binary.read_bytes()),
                  setupLog=pathlib.Path('/tmp/node-typing-setup.log').read_text(),
                  counts=dict(status=dict(collections.Counter(d['status'] for d in comparison)),
                              byClass=dict(collections.Counter(d['stock']['classification'] for d in comparison)),
                              allAdamicDiagnostics=len(run_result['diagnostics']), focusedDiagnostics=len(focused), additionalFocusedDiagnostics=len(additional)),
                  nodeTouchingFiles=files, rows=comparison, focusedDiagnostics=focused,
                  additionalFocusedDiagnostics=additional, run=run_result,
                  control=dict(purpose='One extra root activates bundled Node declarations. This is a control, not the 79-root result or a proposed repair.',
                               source=trigger.read_text(), run=control,
                               originalRows=dict(collections.Counter(d['status'] for d in control_comparison)),
                               rows=control_comparison),
                  mutants=[dict(name='Node hook activation', caught=True, observation='Actual extra node:fs root invalidates all-62-present assertion.'),
                           dict(name='Dropped diagnostic', caught=True, id=rows[0]['id'], observation='Removing a real checker finding changes its status to absent.')])
    (HERE / 'node-typing-2026-10-19.json').write_text(json.dumps(output, indent=2) + '\n')
    print(json.dumps(output['counts']))
    print('Node activation control:', output['control']['originalRows'])


if __name__ == '__main__':
    main()

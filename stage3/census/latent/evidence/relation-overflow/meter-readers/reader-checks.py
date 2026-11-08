"""Exercise the other changed readers with a trailer and a program finding."""
import ast
import gzip
import hashlib
import json
from pathlib import Path
import subprocess
import sys
import tempfile

repository = Path(__file__).resolve().parents[6]
with tempfile.TemporaryDirectory(prefix='latent-reader-checks-') as temporary:
    scratch = Path(temporary)
    source = scratch / 'tree/src/compiler/only.a'
    source.parent.mkdir(parents=True)
    source.write_text('\n')
    finding = dict(kind='NotYet', where='program', reason='relation limit', text='relation limit')
    rows = [dict(status='measurement', checker_rejected=True, latent_mode='full',
                 root_files=[str(source)], diagnostic_sites=[]),
            dict(file=str(source), units=[], findings=[]),
            dict(status='relation_report', relation_overflow=dict(after_load='', end='overflow'),
                 findings=[finding])]
    stream = scratch / 'latent.jsonl'
    stream.write_text(''.join(json.dumps(row) + '\n' for row in rows))
    split = repository / 'stage3/adapt/40-explicit-any/split-previous-133.py'
    function = next(node for node in ast.parse(split.read_text()).body
                    if isinstance(node, ast.FunctionDef) and node.name == 'read')
    namespace = dict(gzip=gzip, json=json)
    exec(compile(ast.Module(body=[function], type_ignores=[]), str(split), 'exec'), namespace)
    assert namespace['read'](stream) == [finding]
    print('PASS: split reader derives prefix from files and retains program NotYet')

    relative = 'src/compiler/only.a'
    identity = dict(file=relative, bytes=1, sha256=hashlib.sha256(b'\n').hexdigest())
    (scratch / 'manifest.json').write_text(json.dumps(dict(root=str(scratch / 'tree'), files=[identity])))
    (scratch / 'stock.json').write_text(json.dumps({relative: dict(identity, bodies=[], units=[])}))
    closure = repository / 'stage3/census/tsc-closure'
    with (scratch / 'measure.log').open('w') as log:
        result = subprocess.run([sys.executable, '-B', str(closure / 'measure.py'),
                                 str(scratch / 'manifest.json'), str(stream), str(stream),
                                 str(scratch / 'stock.json'), str(closure / 'hidden.py'),
                                 str(scratch / 'result.json')], stdout=log, stderr=subprocess.STDOUT)
    print(f'closure_measure_exit={result.returncode}')
    assert result.returncode == 0, (scratch / 'measure.log').read_text()
    result = json.loads((scratch / 'result.json').read_text())
    for mode in ('latent', 'full'):
        assert result[mode]['all']['counts'] == {'NotYet': 1}
        assert result[mode]['unattributed']['counts'] == {'NotYet': 1}
        assert result[mode]['per_file'][relative]['counts'] == {}
    assert result['hidden']['groups']['all']['files'] == 1
    print('PASS: closure file coverage excludes trailer; all/unattributed counts retain program NotYet')

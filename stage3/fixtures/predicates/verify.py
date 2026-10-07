"""Check recorded Node results and census coverage; run dedicated mutants on request."""
import json
import pathlib
import subprocess
import sys
import tempfile

ROOT = pathlib.Path(__file__).resolve().parents[3]
BUCKET = pathlib.Path(__file__).resolve().parent
rows = json.loads((BUCKET / 'status.json').read_text())
ledger = json.loads((BUCKET / 'ledger.json').read_text())
sites = [s for s in json.loads(pathlib.Path(sys.argv[1]).read_text()) if s['reason'] == 'a type predicate']


def census_check(entries):
    expected = {(s['file'], s['start'], s['end']) for s in sites}
    actual = {(s['file'], s['start'], s['end']) for s in entries}
    assert len(entries) == len(actual) == 651 and len({s['file'] for s in entries}) == 31
    assert actual == expected, 'Census locations differ'


def node_check(file, expected, scratch):
    with (scratch / 'stdout').open('wb') as stdout, (scratch / 'stderr').open('wb') as stderr:
        result = subprocess.run(['node', '--disable-warning=ExperimentalWarning', 'oracle/node.mjs', str(file)], cwd=ROOT, stdout=stdout, stderr=stderr)
    actual = {'stdout': (scratch / 'stdout').read_text(), 'stderr': (scratch / 'stderr').read_text(), 'exit': result.returncode}
    assert actual == expected, f'Node observation differs: {file.name}: {actual!r}'


census_check(ledger)
with tempfile.TemporaryDirectory(prefix='predicates-verify-') as directory:
    scratch = pathlib.Path(directory)
    for row in rows:
        assert set(row) == {'file', 'tsc', 'reason', 'node', 'stage0'}
        assert row['stage0']['outcome'] in {'NotYet', 'Refused', 'Checker', 'Compiles'}
        node_check(BUCKET / row['file'], row['node'], scratch)
    print('PASS: all 651 census locations and 11 recorded Node observations')
    if '--mutants' in sys.argv:
        try:
            census_check(ledger[:-1])
        except AssertionError:
            print('CAUGHT: dropped ledger node, by census coverage')
        else:
            raise AssertionError('Ledger mutant survived')
        mutant = json.loads((BUCKET / 'mutant.json').read_text())
        source = (BUCKET / mutant['fixture']).read_text()
        assert source.count(mutant['change']['from']) == 1
        mutated = scratch / 'wrong_kind.a'
        mutated.write_text(source.replace(mutant['change']['from'], mutant['change']['to']))
        try:
            node_check(mutated, mutant['baseline'], scratch)
        except AssertionError as error:
            print('CAUGHT: wrong kind, by exact caller stdout:', error)
        else:
            raise AssertionError('Wrong-kind mutant survived')

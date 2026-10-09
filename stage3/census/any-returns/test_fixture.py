"""Hold declaration provenance to stock TypeScript's three known callees."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
import report

ROOT = Path(__file__).resolve().parent


def inspect(directory, mutant=False):
    source = directory / 'three-callees.a'
    source.write_bytes((ROOT / 'testdata/three-callees.a.txt').read_bytes())
    output = directory / ('mutant.json' if mutant else 'normal.json')
    environment = dict(os.environ)
    if mutant:
        environment['ANY_RETURNS_MUTANT'] = 'lib-to-tsc'
    with (directory / ('mutant.log' if mutant else 'normal.log')).open('w') as log:
        subprocess.run(['node', str(ROOT / 'inspect.cjs'), 'fixture', str(directory), str(source), str(output)],
            check=True, env=environment, stdout=log, stderr=subprocess.STDOUT)
    return json.loads(output.read_text())


def check(result):
    observed = {row['expression']: result['callees'][row['callee']]['declaration_owner']
        for row in result['observations']}
    assert observed == {'own': 'tsc', 'JSON.parse': 'lib.d.ts', 'require': 'Node typings'}, \
        f'three-callee declaration provenance: {observed}'
    own = next(c for c in result['callees'].values() if c['callee'] == 'own')
    assert own['stock_return'] == 'any'
    assert own['body_return_candidate'] == 'number'
    assert own['declared_return'] == 'any'
    assert len(result['observations']) == 3
    totals = report.calculate(result)
    assert (totals['boundaries'], totals['raw_observations'], totals['hidden_bytes']) == (3,3,41)
    assert {owner: (row['boundaries'], row['hidden_bytes']) for owner, row in totals['declaration_owners'].items()} == {
        'tsc': (1,5), 'lib.d.ts': (1,16), 'Node typings': (1,20)}, 'known per-callee call bytes'
    own_candidate = next(row for row in totals['ranked_callees'] if row['callee'] == 'own')['return_candidate']['type']
    assert own_candidate == 'number'


def main():
    with tempfile.TemporaryDirectory(prefix='any-returns-fixture-') as scratch:
        directory = Path(scratch)
        normal = inspect(directory)
        check(normal)
        assert normal['diagnostics'] == 0, 'fixture must be accepted by stock TypeScript'
        (ROOT / 'evidence/fixture.json').write_text(json.dumps(normal, indent=2)+'\n')
        print('PASS: own -> tsc (body number, 5 bytes), JSON.parse -> lib.d.ts (16 bytes), require -> Node typings (20 bytes); one boundary each, total 41')
        mutant = inspect(directory, True)
        assert mutant['diagnostics'] == 0, 'mutant must reach the attribution assertion'
        try:
            check(mutant)
        except AssertionError as error:
            assert str(error).startswith('three-callee declaration provenance:')
            print(f'CAUGHT lib-to-tsc mutant by declaration provenance assertion: {error}')
        else:
            raise AssertionError('lib-to-tsc mutant survived')
        (ROOT / 'evidence/mutant.json').write_text(json.dumps(mutant, indent=2)+'\n')


if __name__ == '__main__':
    main()

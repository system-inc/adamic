#!/usr/bin/env python3
"""Check fresh census membership, attribution, history, and source identity."""
import argparse
import copy
import csv
import hashlib
import gzip
import re
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent
OPTIONS = ('noUncheckedIndexedAccess', 'exactOptionalPropertyTypes', 'useUnknownInCatchVariables', 'strictBindCallApply')
FIELDS = ('file', 'line', 'column', 'code', 'message')


def read(path):
    with path.open(newline='') as source:
        return list(csv.DictReader(source))


def key(row):
    return row['file'], int(row['line']), int(row['column']), row['code']


def validate(rows, changes):
    census = read(ROOT / 'evidence/stage0-census.csv')
    assert len(rows) == len(census) == 173, 'fresh census coverage count'
    assert len({key(row) for row in rows}) == len(rows), 'duplicate diagnostic site'
    assert [tuple(str(row[f]) for f in FIELDS) for row in rows] == [tuple(row[f] for f in FIELDS) for row in census], 'fresh census row or ordering changed'
    own = {key(row) for row in read(ROOT / 'evidence/stock-own.csv')}
    stock = {key(row) for row in read(ROOT / 'evidence/stock-project-stricter.csv')}
    historical_options = {key(row) for row in read(ROOT / 'evidence/stock-census-inputs.csv')}
    assert stock == historical_options, 'project and historical strict profiles differ'
    ablations = {option: {key(row) for row in read(ROOT / ('evidence/stock-project-without-' + option + '.csv'))} for option in OPTIONS}
    for row in rows:
        identity = key(row)
        expected = '1' if identity in own else '2' if identity in stock else '3'
        assert row['class'] == expected, 'checker membership at ' + row['id']
        assert row['owner'] and row['source_expression'] and row['evidence'], 'missing source evidence'
        assert not row['family'], 'unexpected adaptation family'
        if expected == '2':
            removed = [option for option in OPTIONS if identity not in ablations[option]]
            assert row['option'] == (removed[0] if removed else 'other'), 'option attribution at ' + row['id']
            assert row['removed_by'] == ';'.join(removed), 'ablation evidence at ' + row['id']
        if expected == '3':
            assert (ROOT / row['witness']).is_file(), 'missing checker-difference witness'
    extras = stock - {key(row) for row in census}
    assert {key(row) for row in read(ROOT.parent / 'stock-only-2026-10-08.csv')} == extras, 'stock-only coverage'
    previous = read(ROOT.parent / 'rows.csv')
    maps = json.loads((ROOT / 'evidence/line-map.json').read_text())
    current = {row['id']: row for row in rows}
    expected_changes = {}
    for row in previous:
        new = current.get(row['id'])
        if new is None:
            expected_changes[row['id']] = 'removed'
        else:
            predicted = maps.get(row['file'], {}).get(str(row['line']), int(row['line']))
            assert int(new['line']) == predicted and new['column'] == row['column'] and new['code'] == row['code'], 'source location mapping at ' + row['id']
            if new['class'] != row['class']:
                expected_changes[row['id']] = 'moved column'
            elif key(new) != key(row):
                expected_changes[row['id']] = 'location shifted'
    assert len({r['id'] for r in changes}) == len(changes), 'duplicate history row'
    assert {r['id']: r['change'] for r in changes} == expected_changes, 'complete diagnostic change history'
    removed = [r for r in previous if r['id'] not in current]
    assert len(removed) == 86, 'expected removals'
    assert len([r for r in removed if r['class'] == '2' and r['code'] != 'TS2740']) == 84, 'iterator removals'
    assert next(r for r in removed if r['code'] == 'TS2740')['id'] == 'D115', 'Set removal'
    assert next(r for r in removed if r['class'] == '3')['id'] == 'D003', 'repeated tuple removal'


def validate_hashes(tree, hashes):
    assert len(hashes) == 79, 'source input count'
    for name, expected in hashes.items():
        assert hashlib.sha256((tree / name).read_bytes()).hexdigest() == expected, 'source hash at ' + name



def validate_per_roots(observations):
    expected = set(json.loads((ROOT / 'evidence/source-hashes.json').read_text()))
    assert len(observations) == 79 and {r['file'] for r in observations} == expected, 'per-root census coverage'
    baseline = {key(r) for r in read(ROOT / 'evidence/stage0-census.csv')}
    clean = {'src/compiler/corePublic.ts', 'src/compiler/hostErrors.ts'}
    for result in observations:
        assert not result.get('error'), 'per-root loader error'
        sites = set()
        for diagnostic in result['diagnostics']:
            match = re.match(r'(.+):(\d+):(\d+): error (TS\d+):', diagnostic)
            assert match, 'malformed per-root diagnostic'
            file, line, column, code = match.groups()
            file = 'src/compiler/' + file.split('/src/compiler/', 1)[1]
            sites.add((file, int(line), int(column), code))
        assert sites == (set() if result['file'] in clean else baseline), 'per-root site membership'


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--tree', type=Path, required=True)
    parser.add_argument('--mutants', action='store_true')
    args = parser.parse_args()
    rows = read(ROOT.parent / 'rows-2026-10-08.csv')
    changes = read(ROOT.parent / 'changes-2026-10-08.csv')
    hashes = json.loads((ROOT / 'evidence/source-hashes.json').read_text())
    validate(rows, changes)
    validate_hashes(args.tree, hashes)
    observations = [json.loads(line) for line in gzip.open(ROOT / 'evidence/stage0-per-root.jsonl.gz', 'rt')]
    validate_per_roots(observations)
    print('PASS: 173 fresh census rows, checker membership, option ablations, complete history, 79 source hashes, all 79 per-root census loads')
    if args.mutants:
        mutations = [('delete-diagnostic', rows[:-1], changes)]
        changed = copy.deepcopy(rows)
        changed[0]['class'] = '1'
        mutations.append(('misclassify-runtime-row', changed, changes))
        changed = copy.deepcopy(rows)
        next(r for r in changed if r['option'] == 'noUncheckedIndexedAccess')['option'] = 'exactOptionalPropertyTypes'
        mutations.append(('misattribute-indexed-read', changed, changes))
        changed = copy.deepcopy(rows)
        changed[-1] = copy.deepcopy(changed[0])
        mutations.append(('duplicate-site', changed, changes))
        mutations.append(('omit-history-change', rows, changes[:-1]))
        for name, mutated_rows, mutated_changes in mutations:
            try:
                validate(mutated_rows, mutated_changes)
            except AssertionError as error:
                print('CAUGHT:', name, 'by', error)
            else:
                raise AssertionError('surviving mutant: ' + name)
        try:
            validate_per_roots(observations[:-1])
        except AssertionError as error:
            print('CAUGHT: omit-census-root by', error)
        else:
            raise AssertionError('surviving census-root mutant')
        changed_hashes = dict(hashes)
        changed_hashes[next(iter(changed_hashes))] = '0' * 64
        try:
            validate_hashes(args.tree, changed_hashes)
        except AssertionError as error:
            print('CAUGHT: wrong-source-hash by', error)
        else:
            raise AssertionError('surviving source-hash mutant')


if __name__ == '__main__':
    main()

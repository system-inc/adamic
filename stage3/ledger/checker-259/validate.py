#!/usr/bin/env python3
"""Validate census coverage against saved independent checker observations."""
import argparse
import copy
import csv
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent
FIELDS = ('file', 'line', 'column', 'code', 'message')
OPTIONS = ('noUncheckedIndexedAccess', 'exactOptionalPropertyTypes', 'useUnknownInCatchVariables', 'strictBindCallApply')


def read(name):
    with (ROOT / name).open(newline='') as source:
        return list(csv.DictReader(source))


def key(row):
    return row['file'], int(row['line']), int(row['column']), row['code']


def validate(rows):
    census = read('evidence/stage0-census.csv')
    assert len(rows) == len(census) == 259, 'census coverage count'
    assert len({key(row) for row in rows}) == len(rows), 'duplicate diagnostic site'
    assert [tuple(row[f] for f in FIELDS) for row in rows] == [tuple(row[f] for f in FIELDS) for row in census], 'original census row or ordering changed'
    own = {key(row) for row in read('evidence/stock-own.csv')}
    stock = {key(row) for row in read('evidence/stock-census-inputs.csv')}
    ablations = {option: {key(row) for row in read('evidence/stock-census-without-' + option + '.csv')} for option in OPTIONS}
    for row in rows:
        identity = key(row)
        expected = '1' if identity in own else '2' if identity in stock else '3'
        assert row['class'] == expected, 'checker membership classification at ' + row['id']
        assert row['owner'] and row['source_expression'] and row['evidence'], 'missing source evidence'
        if expected == '1':
            assert row['family'], 'missing adaptation family'
        else:
            assert not row['family'], 'adaptation family on a runtime-check or checker-difference row'
        if expected == '2':
            removed = [option for option in OPTIONS if identity not in ablations[option]]
            primary = removed[0] if removed else 'other'
            assert row['option'] == primary, 'option attribution at ' + row['id']
            assert row['removed_by'] == ';'.join(removed), 'ablation evidence at ' + row['id']
        if expected == '3':
            assert (ROOT / row['witness']).is_file(), 'missing checker-difference witness'
        if row['witness_for_input_difference']:
            assert (ROOT / row['witness_for_input_difference']).is_file(), 'missing declaration witness'
    extras = stock - {key(row) for row in census}
    assert {key(row) for row in read('stock-only.csv')} == extras, 'stock-only coverage'


def validate_hashes(tree, hashes):
    assert len(hashes) == 79, 'source input count'
    for name, expected in hashes.items():
        assert hashlib.sha256((tree / name).read_bytes()).hexdigest() == expected, 'adapted source hash at ' + name


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--tree', type=Path)
    parser.add_argument('--mutants', action='store_true')
    args = parser.parse_args()
    rows = read('rows.csv')
    validate(rows)
    print('PASS: 259 original rows, unique sites, checker classes, option ablations, evidence and stock-only coverage')
    hashes = json.loads((ROOT / 'evidence/source-hashes.json').read_text())
    if args.tree:
        validate_hashes(args.tree, hashes)
        print('PASS: 79 adapted source hashes')
    if args.mutants:
        mutants = []
        mutants.append(('delete-diagnostic', rows[:-1]))
        changed = copy.deepcopy(rows)
        changed[0]['class'] = '1'
        mutants.append(('misclassify-runtime-row', changed))
        changed = copy.deepcopy(rows)
        target = next(row for row in changed if row['option'] == 'noUncheckedIndexedAccess')
        target['option'] = 'exactOptionalPropertyTypes'
        mutants.append(('misattribute-indexed-read', changed))
        changed = copy.deepcopy(rows)
        changed[-1] = copy.deepcopy(changed[0])
        mutants.append(('duplicate-site', changed))
        for name, changed in mutants:
            try:
                validate(changed)
            except AssertionError as error:
                print('CAUGHT:', name, 'by', error)
            else:
                raise AssertionError('surviving mutant: ' + name)
        if args.tree:
            changed = dict(hashes)
            changed[next(iter(changed))] = '0' * 64
            try:
                validate_hashes(args.tree, changed)
            except AssertionError as error:
                print('CAUGHT: wrong-source-hash by', error)
            else:
                raise AssertionError('surviving source-hash mutant')


if __name__ == '__main__':
    main()

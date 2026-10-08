"""Independent SQL accounting audit, with corrupt-report and source-hash mutants."""
import collections
import copy
import csv
import gzip
import hashlib
import json
from pathlib import Path
import sqlite3
import sys

import compare as measured

HERE = Path(__file__).resolve().parent
FIELDS = ('kind', 'where', 'reason', 'text')
PREFIX = '/tmp/stage3-notyet-adapted/'


def sql_roots():
    database = sqlite3.connect(':memory:')
    database.execute('CREATE TABLE observations (run TEXT, signature TEXT, blocked INT)')
    summaries = {}
    for name in ('before', 'after'):
        with gzip.open(HERE / name / 'full.jsonl.gz', 'rt') as stream:
            records = [json.loads(line) for line in stream]
        for record in records[1:]:
            for finding in record['findings']:
                if finding['phase'] == 'lowering' and finding['kind'] == 'NotYet':
                    value = json.dumps([finding[field].replace(PREFIX, '') for field in FIELDS])
                    database.execute('INSERT INTO observations VALUES (?, ?, ?)', (name, value, int(bool(finding.get('blocked_by')))))
        summaries[name] = json.loads((HERE / name / 'summary.json').read_text())
    result = {}
    for name in ('before', 'after'):
        result[name] = {tuple(json.loads(row[0])) for row in database.execute('SELECT signature FROM observations WHERE run = ? GROUP BY signature HAVING MIN(blocked) = 0', (name,))}
        total = database.execute('SELECT COUNT(DISTINCT signature) FROM observations WHERE run = ?', (name,)).fetchone()[0]
        echoes = total - len(result[name])
        assert (total, echoes, len(result[name])) == (summaries[name]['notyet_sites'], summaries[name]['echo_only_sites'], summaries[name]['root_sites']), 'SQL root and echo totals'
        with (HERE / name / 'roots.csv').open(newline='') as stream:
            recorded = {tuple(row[field] for field in FIELDS) for row in csv.DictReader(stream)}
        assert recorded == result[name], 'run roots match SQL'
    return result


def validate(rows, totals, roots, reasons):
    before, after = roots['before'], roots['after']
    reported = {}
    for row in rows:
        key = tuple(row[field] for field in FIELDS)
        assert key not in reported, 'duplicate signature'
        flags = tuple(int(row[field]) for field in ('retired', 'newly_exposed', 'surviving'))
        expected = (int(key in before and key not in after), int(key in after and key not in before), int(key in before and key in after))
        assert flags == expected, 'SQL membership flags'
        reported[key] = flags
    assert set(reported) == before | after, 'complete union of observed SQL roots'
    assert totals['before'] == len(before) and totals['after'] == len(after), 'recorded root sizes'
    sums = tuple(sum(row[index] for row in reported.values()) for index in range(3))
    assert sums == tuple(totals[field] for field in ('retired', 'newly_exposed', 'surviving')), 'recorded partition totals'
    assert totals['net_retirement'] == len(before) - len(after), 'net difference'
    grouped = collections.defaultdict(lambda: [0, 0, 0])
    for key, flags in reported.items():
        for index in range(3):
            grouped[key[2]][index] += flags[index]
    assert len(reasons) == len(grouped), 'complete reason groups'
    for row in reasons:
        expected = grouped[row['reason']]
        assert expected == [int(row[field]) for field in ('retired', 'newly_exposed', 'surviving')], 'reason partition totals'
        assert int(row['net_retirement']) == expected[0] - expected[1], 'reason net'


def must_fail(name, check, message):
    try:
        check()
    except AssertionError as failure:
        assert message in str(failure), (name, str(failure))
        print(name + ' mutant caught: ' + str(failure))
    else:
        raise AssertionError(name + ' mutant survived')


def main(source):
    source = Path(source)
    manifest = json.loads((HERE / 'source-manifest.json').read_text())
    assert (HERE / 'source-manifest.json').read_bytes() == (HERE.parent / 'source-manifest.json').read_bytes(), 'original manifest identical'
    measured.verify_manifest(manifest, source)
    roots = sql_roots()
    with (HERE / 'before-to-after-signatures.csv').open(newline='') as stream:
        rows = list(csv.DictReader(stream))
    with (HERE / 'before-to-after-reasons.csv').open(newline='') as stream:
        reasons = list(csv.DictReader(stream))
    totals = json.loads((HERE / 'comparison.json').read_text())['before-to-after']['totals']
    validate(rows, totals, roots, reasons)
    assert (HERE / 'before/metadata.json').read_bytes() == (HERE / 'after/metadata.json').read_bytes(), 'identical checker roots, sources, diagnostics and measurement status'
    assert json.loads((HERE / 'before-to-after-unit-changes.json').read_text()) == [], 'no unit eligibility changes'
    print('PASS: exact 81-file manifest, identical checker metadata, unchanged unit eligibility; independent SQL validates every root signature, root/echo totals, reason group, and net retirement')

    swapped = copy.deepcopy(rows)
    for row in swapped:
        row['retired'], row['newly_exposed'] = row['newly_exposed'], row['retired']
    must_fail('swap retired and exposed', lambda: validate(swapped, totals, roots, reasons), 'SQL membership flags')

    survivor = next(index for index, row in enumerate(rows) if int(row['surviving']))
    missing = rows[:survivor] + rows[survivor + 1:]
    must_fail('drop surviving signature', lambda: validate(missing, totals, roots, reasons), 'complete union')

    with (HERE / 'before/echoes.csv').open(newline='') as stream:
        echo = next(row for row in csv.DictReader(stream) if row['echo_only'] == 'True')
    fake = {field: echo['echo_' + field] for field in FIELDS}
    fake.update(retired='1', newly_exposed='0', surviving='0')
    must_fail('count rollback echo as root', lambda: validate(rows + [fake], totals, roots, reasons), 'SQL membership flags')

    wrong_net = dict(totals, net_retirement=totals['net_retirement'] + 1)
    must_fail('inflate net retirement', lambda: validate(rows, wrong_net, roots, reasons), 'net difference')

    wrong_hash = copy.deepcopy(manifest)
    wrong_hash[0]['sha256'] = '0' * 64
    must_fail('change frozen hash', lambda: measured.verify_manifest(wrong_hash, source), 'source hash')

    before_records = measured.read_records(HERE / 'before/full.jsonl.gz')
    after_records = measured.read_records(HERE / 'after/full.jsonl.gz')
    changed_source_set = copy.deepcopy(after_records)
    changed_source_set.pop()
    must_fail('drop source result', lambda: measured.census(changed_source_set, manifest), 'exact frozen source set')

    missing_root = copy.deepcopy(before_records)
    record = next(record for record in missing_root[1:] if any(f.get('blocked_by') for f in record['findings']))
    echo = next(f for f in record['findings'] if f.get('blocked_by'))
    cause = echo['blocked_by']
    record['findings'] = [f for f in record['findings'] if any(f.get(k) != cause[k] for k in measured.CAUSE)]
    must_fail('remove governing root', lambda: measured.census(missing_root, manifest), 'actual root exists')

    missing_symbol = copy.deepcopy(before_records)
    echo = next(f for r in missing_symbol[1:] for f in r['findings'] if f.get('blocked_by') and f['kind'] == 'NotYet')
    echo.pop('blocked_symbol_declaration')
    must_fail('remove declaration symbol', lambda: measured.census(missing_symbol, manifest), 'declaration symbol provenance')

    def same_checker_metadata(left, right):
        assert left == right, 'identical checker metadata'
    changed_checker = copy.deepcopy(after_records[0])
    changed_checker['diagnostic_sites'].pop()
    must_fail('change checker diagnostics', lambda: same_checker_metadata(before_records[0], changed_checker), 'identical checker metadata')

    baseline = measured.census(before_records, manifest)
    changed_unit = copy.deepcopy(after_records)
    changed_unit[1]['units'][0]['status'] = 'skipped_checker_body'
    def same_units():
        result = measured.compare(baseline, measured.census(changed_unit, manifest))
        assert result[0]['unit_eligibility_changes'] == 0, 'unchanged unit eligibility'
    must_fail('change unit eligibility', same_units, 'unchanged unit eligibility')

    for name in ('before', 'after'):
        raw = gzip.decompress((HERE / name / 'full.jsonl.gz').read_bytes())
        saved = json.loads((HERE / name / 'run.json').read_text())
        assert hashlib.sha256(raw).hexdigest() == saved['raw_jsonl_sha256'] and len(raw) == saved['raw_jsonl_bytes'], 'raw output identity'
    print('PASS: both saved compressed outputs reproduce their raw SHA-256 and byte counts; all ten accounting and provenance-structure mutants caught')


if __name__ == '__main__':
    main(*sys.argv[1:])

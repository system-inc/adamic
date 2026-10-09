#!/usr/bin/env python3
"""Assemble dated rows from fresh checker observations, retaining first-ledger IDs."""
import argparse
import csv
import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parent
OPTIONS = ('noUncheckedIndexedAccess', 'exactOptionalPropertyTypes', 'useUnknownInCatchVariables', 'strictBindCallApply')


def read_csv(path):
    with path.open(newline='') as source:
        return list(csv.DictReader(source))


def key(row):
    return row['file'], int(row['line']), int(row['column']), row['code']


def write_csv(path, rows, fields):
    with path.open('w', newline='') as output:
        writer = csv.DictWriter(output, fieldnames=fields, lineterminator='\n')
        writer.writeheader()
        writer.writerows(rows)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--stock', type=Path, required=True)
    parser.add_argument('--stage0', type=Path, required=True)
    parser.add_argument('--tree', type=Path, required=True)
    args = parser.parse_args()
    modes = {p.stem: json.loads(p.read_text()) for p in args.stock.glob('*.json')}
    basic = ['file', 'line', 'column', 'code', 'message']
    for name, result in modes.items():
        write_csv(ROOT / 'evidence' / ('stock-' + name + '.csv'),
                  [{f: row[f] for f in basic} for row in result['rows']], basic)
    (ROOT / 'evidence/options.json').write_text(json.dumps(
        {name: {f: result[f] for f in ['version', 'options', 'roots', 'config_errors']} for name, result in modes.items()}, indent=2) + '\n')
    stage0 = json.loads(args.stage0.read_text())
    assert not stage0.get('error'), stage0
    census = []
    for diagnostic in stage0['diagnostics']:
        match = re.fullmatch(r'(.+):(\d+):(\d+): error (TS\d+): ([\s\S]*)', diagnostic)
        assert match, diagnostic
        file, line, column, code, message = match.groups()
        census.append(dict(file=str(Path(file).relative_to(args.tree)), line=int(line), column=int(column), code=code, message=message))
    census.sort(key=key)
    write_csv(ROOT / 'evidence/stage0-census.csv', census, basic)
    previous = read_csv(ROOT.parent / 'rows.csv')
    maps = json.loads((ROOT / 'evidence/line-map.json').read_text())
    def shifted(row):
        return row['file'], maps.get(row['file'], {}).get(str(row['line']), int(row['line'])), int(row['column']), row['code']
    prior = {shifted(row): row for row in previous}
    own = {key(row): row for row in modes['own']['rows']}
    stock = {key(row): row for row in modes['project-stricter']['rows']}
    ablations = {option: {key(row) for row in modes['project-without-' + option]['rows']} for option in OPTIONS}
    rows = []
    for row in census:
        identity = key(row)
        assert identity in prior, 'new site needs evidence review: ' + str(identity)
        result = dict(prior[identity])
        result.update(row)
        result['class'] = '1' if identity in own else '2' if identity in stock else '3'
        if result['class'] == '2':
            removed = [option for option in OPTIONS if identity not in ablations[option]]
            result['option'] = removed[0] if removed else 'other'
            result['removed_by'] = ';'.join(removed)
        result['stock_message'] = stock.get(identity, {}).get('message', '')
        result['stock_unmodified_reports'] = str(identity in {key(r) for r in modes['effective']['rows']}).lower()
        old_owner_line = str(result['owner_line'])
        result['owner_line'] = maps.get(row['file'], {}).get(old_owner_line, int(old_owner_line))
        result['evidence'] = result['evidence'].replace('Census-input stock option ablation', '2026-10-08 project-input stock option ablation')
        if result['witness_for_input_difference']:
            result['witness_for_input_difference'] = '../' + result['witness_for_input_difference']
        if result['witness']:
            result['witness'] = '../' + result['witness']
        rows.append(result)
    fields = list(previous[0])
    write_csv(ROOT.parent / 'rows-2026-10-08.csv', rows, fields)
    write_csv(ROOT.parent / 'indexed-reads-2026-10-08.csv', [r for r in rows if r['class'] == '2' and r['kind'] == 'indexed read'], fields)
    current = {key(row): row for row in rows}
    changes = []
    for row in previous:
        identity = shifted(row)
        new = current.get(identity)
        if new is None:
            reason = 'iterator done-flag discrimination'
            if row['code'] == 'TS2740':
                reason = 'project library omits ES2025 Set additions'
            changes.append(dict(id=row['id'],change='removed',file=row['file'],old_line=row['line'],old_column=row['column'],new_line='',new_column='',code=row['code'],old_class=row['class'],new_class='',reason=reason))
        elif row['class'] != new['class'] or key(row) != key(new):
            changes.append(dict(id=row['id'],change='moved column' if row['class'] != new['class'] else 'location shifted',file=row['file'],old_line=row['line'],old_column=row['column'],new_line=new['line'],new_column=new['column'],code=row['code'],old_class=row['class'],new_class=new['class'],reason='updated adaptations shift source lines' if row['class'] == new['class'] else 'fresh checker membership'))
    change_fields = ['id','change','file','old_line','old_column','new_line','new_column','code','old_class','new_class','reason']
    write_csv(ROOT.parent / 'changes-2026-10-08.csv', changes, change_fields)
    extras = [row for identity, row in stock.items() if identity not in current]
    write_csv(ROOT.parent / 'stock-only-2026-10-08.csv', [{f: row[f] for f in basic} for row in extras], basic)
    print('assembled', len(rows), 'rows;', len(changes), 'changes;', len(extras), 'stock-only rows')


if __name__ == '__main__':
    main()

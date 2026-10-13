"""Real call witnesses, coverage gaps, unit counts and independent byte-mask mutants."""
import copy
import importlib.util
import json
import os
from pathlib import Path
import subprocess
import sys
from unit_status import DEFERRED, audit_deferred, body_eligible, unit_counts

binary, output = map(lambda p: Path(p).resolve(), sys.argv[1:])
output.mkdir(parents=True, exist_ok=True)
source = output/'source'
source.mkdir(exist_ok=True)
(source/'a.a').write_text('''// a-check: type error TS2322
export function unused<T>(value: T): T { function unusedNested(): T { return value; } return value; }
export function called<T>(value: T): T { return value; }
export function diagnosed<T>(value: T): T { const bad: number = "wrong"; return value; }
export function outer<T>(value: T): T { function nested(): T { return value; } return value; }
export class Box<T> { method<U>(value: U): U { return value; } }
export const arrow = <T>(value: T): T => value;
''')
(source/'z.a').write_text('''import { called, outer } from "./a.a";
export function caller(): number { return called(7) + outer(8); }
''')
spec = importlib.util.spec_from_file_location('hidden', Path(__file__).resolve().parents[1]/'tsc-closure/hidden.py')
hidden = importlib.util.module_from_spec(spec)
spec.loader.exec_module(hidden)


def check_counts(report, units):
    # Independent status membership oracle, not the report helper.
    assert report['attempted_functions'] == sum(u['kind']=='KindFunctionDeclaration' and u['status'] in ('attempted', 'panic', 'speculative') for u in units), 'deferred counted as attempted'
    assert report['attempted_statements'] == sum(u['kind']!='KindFunctionDeclaration' and u['status'] in ('attempted', 'panic', 'speculative') for u in units), 'deferred counted as attempted'
    audit_deferred(report, units)


def killed(name, check):
    try:
        check()
    except AssertionError as error:
        print(name+': caught by '+str(error))
    else:
        raise AssertionError(name+' survived')


for mode in ('legacy', 'full', 'speculative'):
    destination = output/(mode+'.jsonl')
    env = dict(os.environ, LATENT_ASSERT_NO_OUTPUT='1', LATENT_FULL='0' if mode=='legacy' else '1', LATENT_SPECULATIVE='1' if mode=='speculative' else '0')
    with (output/(mode+'.log')).open('w') as log:
        subprocess.run([str(binary), str(source), str(destination)], env=env, stdout=log, stderr=subprocess.STDOUT, timeout=45, check=True)
    rows = hidden.read_rows(destination)
    units = [u for row in rows[1:] for u in row['units']]
    named = {u['name']: u for u in units if u['kind']=='KindFunctionDeclaration'}
    for name in ('unused', 'called', 'diagnosed', 'outer'):
        assert named[name]['status']==DEFERRED, (mode, name, named[name])
    assert named['called']['instantiated'], (mode, named['called'])
    assert named['outer']['instantiated'], (mode, named['outer'])
    assert all(u['status'] != DEFERRED for u in units if u['kind'] in ('KindClassDeclaration', 'KindVariableStatement')), mode
    assert not named['unused']['instantiated'], (mode, named['unused'])
    assert named['diagnosed']['checker_diagnostics'] and body_eligible(named['diagnosed'], named['diagnosed']['checker_diagnostics'])
    if mode!='legacy':
        assert named['nested']['status']==DEFERRED and named['nested']['instantiated'], named['nested']
        assert named['unusedNested']['status']==DEFERRED and not named['unusedNested']['instantiated'], named['unusedNested']
    summary = unit_counts(units)
    check_counts(summary, units)
    assert any(u['where']==named['unused']['where'] for u in summary['never_reached'])
    mutant = copy.deepcopy(summary)
    mutant.pop('never_reached')
    killed(mode+' missing never-reached list', lambda: check_counts(mutant, units))
    mutant = copy.deepcopy(summary)
    mutant['attempted_functions'] += 1
    killed(mode+' deferred as attempted', lambda: audit_deferred(mutant, units))
    (output/(mode+'-summary.json')).write_text(json.dumps(summary, indent=2)+'\n')
    if mode=='speculative':
        continue
    # Stock-shaped spans checked directly against the source's braces. These
    # small functions have no statement braces except outer's nested function.
    stock = {}
    for row in rows[1:]:
        path = Path(row['file'])
        data = path.read_bytes()
        entries = []
        for unit in row['units']:
            line = int(unit['where'].rsplit(':',2)[1])
            start = sum(len(text) for text in data.splitlines(keepends=True)[:line-1])
            end = start + len(data.splitlines(keepends=True)[line-1].rstrip(b'\n'))
            where = path.name + ':' + ':'.join(unit['where'].rsplit(':',2)[1:])
            entries.append(dict(where=where, start=start, end=end))
            if unit['status']==DEFERRED and unit.get('depth',0)==0:
                left = data.index(b'{', start)
                assert data[unit['body_start']:unit['body_end']].strip().startswith(b'{')
                assert unit['body_end']==end
        stock[path.name] = dict(bytes=len(data), units=entries, bodies=[])
    result = hidden.calculate(rows, stock, source)
    assert result['deferred_body_bytes']>0
    stock_path = output/(mode+'-stock.json')
    stock_path.write_text(json.dumps(stock))
    report = dict(closure=dict(root=str(source)), hidden=result)
    report_path = output/(mode+'-hidden.json')
    report_path.write_text(json.dumps(report))
    command = ['python3', str(Path(__file__).resolve().parents[1]/'tsc-closure/rerun-2026-10-09/audit.py'), str(destination), str(stock_path), str(report_path)]
    with (output/(mode+'-byte-audit.log')).open('w') as log:
        subprocess.run(command, stdout=log, stderr=subprocess.STDOUT, timeout=30, check=True)
    # Fold deferred bytes into examined, keeping subtraction arithmetically
    # consistent. Only the independent byte masks expose the false coverage.
    mutant = copy.deepcopy(report)
    for entry in mutant['hidden']['files'].values():
        amount = entry['deferred_body_bytes']
        entry['independently_examined_bytes'] += amount
        entry['blocked_union_bytes'] += amount
        entry['deferred_body_bytes'] = 0
    amount = mutant['hidden']['deferred_body_bytes']
    mutant['hidden']['independently_examined_bytes'] += amount
    mutant['hidden']['blocked_union_bytes'] += amount
    mutant['hidden']['deferred_body_bytes'] = 0
    report_path.write_text(json.dumps(mutant))
    with (output/(mode+'-byte-mutant.log')).open('w') as log:
        process = subprocess.run(command, stdout=log, stderr=subprocess.STDOUT, timeout=30)
    assert process.returncode!=0 and 'file deferred bytes' in (output/(mode+'-byte-mutant.log')).read_text()
    report_path.write_text(json.dumps(report))
    print(mode+' deferred bytes as examined: caught by independent byte-mask audit')
print('PASS: never reached, cross-file concrete caller, diagnosed deferral, nested gap, explicit counts and bytes')

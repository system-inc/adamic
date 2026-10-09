"""Audit and summarize the same-input, no-output 26-root measurement."""
import copy
import gzip
import hashlib
import json
import subprocess
from pathlib import Path
import sys

HERE = Path(__file__).resolve().parent
# HERE is review/compiler/generic-values; the historical ledger is its sibling.
LEDGER = HERE.parent / 'lowering-chain/source-member-4/step-16-generics/baseline.json.gz'
REASON = 'a generic function as a value'

def short(path):
    return path.split('/src/compiler/')[-1]

def read(path):
    opener = gzip.open if str(path).endswith('.gz') else open
    with opener(path, 'rt') as stream:
        return [json.loads(line) for line in stream]

def collect(records):
    units = {}
    findings = []
    for record in records[1:]:
        for unit in record['units']:
            key = short(unit['where'])
            assert key not in units, ('duplicate root', key)
            units[key] = unit
        findings.extend(record['findings'])
    return units, findings

def validate(ledger, base, head, hashes):
    assert hashes == ledger['source_sha256'], 'source bytes differ from historical ledger'
    assert len(hashes) == 79, 'source coverage changed'
    assert base[0] == head[0], 'checker diagnostics or measurement mode changed'
    assert len(base[0]['diagnostics']) == 324 and base[0]['latent_mode'] == 'full'
    roots = {x['unit'] for x in ledger['findings'] if x['reason'] == REASON}
    b_units, b_findings = collect(base)
    h_units, h_findings = collect(head)
    assert len(roots) == 26 and set(b_units) == roots == set(h_units), 'root coverage changed'
    assert b_units == h_units, 'root status or body range changed'
    panic_key = lambda x: (short(x['unit']), short(x['where']), x['phase'], x['reason'])
    assert sorted(panic_key(x) for x in b_findings if x['kind']=='panic') == sorted(panic_key(x) for x in h_findings if x['kind']=='panic'), 'recovered panic boundaries changed'
    historical = {x['where']: x for x in ledger['root_units']}
    for root, unit in b_units.items():
        for key in ['body_start', 'body_end', 'status', 'kind', 'name', 'depth']:
            assert unit.get(key) == historical[root].get(key), ('historical root changed', root, key)
    return roots, b_units, b_findings, h_findings

def generic(finding):
    reason = finding['reason']
    return reason == REASON or 'generic function ' in reason and 'as a value' in reason or 'higher-rank callable slot' in reason or 'a callable slot whose type graph exceeds' in reason

def main():
    corpus, base_path, head_path, destination = map(Path, sys.argv[1:])
    ledger = json.load(gzip.open(LEDGER, 'rt'))
    base, head = read(base_path), read(head_path)
    hashes = {name: hashlib.sha256((corpus/'src/compiler'/name).read_bytes()).hexdigest() for name in ledger['source_sha256']}
    roots, units, b_findings, h_findings = validate(ledger, base, head, hashes)
    mutants = []
    bad = dict(hashes); bad[next(iter(bad))] = '0'*64
    try:
        validate(ledger, base, head, bad)
    except AssertionError as error:
        mutants.append({'mutant': 'changed source hash', 'caught': str(error)})
    else:
        raise AssertionError('source hash mutant escaped')
    missing = copy.deepcopy(head)
    next(record for record in missing[1:] if record['units'])['units'].pop()
    try:
        validate(ledger, base, missing, hashes)
    except AssertionError as error:
        mutants.append({'mutant': 'missing root', 'caught': str(error)})
    else:
        raise AssertionError('missing root mutant escaped')
    def selected(items, root):
        return [dict(x, unit=short(x['unit']), where=short(x['where'])) for x in items if x['phase']=='lowering' and short(x['unit'])==root]
    rows = []
    for root in sorted(roots):
        original = [x for x in ledger['findings'] if x['reason']==REASON and x['unit']==root]
        before, after = selected(b_findings, root), selected(h_findings, root)
        rows.append({'root':root, 'name':units[root].get('name'), 'body_bytes':units[root]['body_end']-units[root]['body_start'] if 'body_end' in units[root] else None, 'original_sites':sorted({x['where'] for x in original}), 'base_generic_blockers':[x for x in before if generic(x)], 'head_generic_blockers':[x for x in after if generic(x)], 'base_all_blockers':before, 'head_all_blockers':after})
    summary = {'roots':len(rows), 'original_distinct_sites':len({site for row in rows for site in row['original_sites']}), 'source_files':len(hashes), 'checker_diagnostics':324, 'base_generic_root_count':sum(bool(row['base_generic_blockers']) for row in rows), 'head_generic_root_count':sum(bool(row['head_generic_blockers']) for row in rows), 'base_generic_site_count':len({x['where'] for row in rows for x in row['base_generic_blockers']}), 'head_generic_site_count':len({x['where'] for row in rows for x in row['head_generic_blockers']}), 'head_roots_without_observed_lowering_blockers':sum(not row['head_all_blockers'] for row in rows), 'base_scan_findings':sum(x['phase']!='lowering' for x in b_findings), 'head_scan_findings':sum(x['phase']!='lowering' for x in h_findings), 'base_lowering_panics':sum(x['phase']=='lowering' and x['kind']=='panic' for x in b_findings), 'head_lowering_panics':sum(x['phase']=='lowering' and x['kind']=='panic' for x in h_findings)}
    destination.write_text(json.dumps({'measurement':'isolated lowering on a checker-rejected program; no IR or backend output', 'base':'50654a407f9bcc3a7fb821da88088b9d99e6a728', 'head':subprocess.check_output(['git', '-C', str(HERE.parents[2]), 'rev-parse', 'HEAD'], text=True).strip(), 'summary':summary, 'source_sha256':hashes, 'audit_mutants':mutants, 'rows':rows}, indent=2)+'\n')
    lines = ['# Step 16 generic-value remeasurement', '', 'Same source bytes, checker diagnostics, root identities, statuses and body spans. Full isolated lowering with production Load/Lower disabled and no-output guard enabled. This is a blocker observation, not successful compilation or a claim of retired hidden bytes.', '', '| Root | Historical value sites | Base generic blockers | Head generic blockers | Other head lowering blockers |', '| --- | --- | ---: | ---: | ---: |']
    for row in rows:
        lines.append('| '+row['root']+' | '+', '.join(row['original_sites'])+' | '+str(len(row['base_generic_blockers']))+' | '+str(len(row['head_generic_blockers']))+' | '+str(len(row['head_all_blockers'])-len(row['head_generic_blockers']))+' |')
    lines += ['', 'Summary: `'+json.dumps(summary,sort_keys=True)+'`', '', 'The JSON companion retains every selected-root finding, module-scan findings, and recovered panic counts. Both runs have ten unchanged computed-property module-scan panics and 29 unchanged isolated-statement panics. Module scan failures and other lowering boundaries can hide children. A root without an observed lowering blocker is not certified accepted.', '', 'Audit mutants: '+ '; '.join(x['mutant']+' rejected: '+x['caught'] for x in mutants)+'.', '']
    destination.with_suffix('.md').write_text('\n'.join(lines))
    print(json.dumps(summary,sort_keys=True))
    print(json.dumps(mutants))

if __name__ == '__main__':
    main()

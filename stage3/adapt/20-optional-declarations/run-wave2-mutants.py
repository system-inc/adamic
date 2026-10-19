#!/usr/bin/env python3
"""Revert each resolver mechanism, then run real full apply and census."""
import argparse
import gzip
import hashlib
import json
import os
from pathlib import Path
import signal
import subprocess
import sys

REPO = Path(__file__).resolve().parents[3]
HELPER = REPO / 'stage3/adapt/20-optional-declarations/resolve-internal.cjs'
MUTATIONS = {
    'nullable-argument': ("return contextual(diagnostic, finding, 'nullable-argument');", 'return false;', 1),
    'nullable-assignment': ("return contextual(diagnostic, finding, 'nullable-assignment');", 'return false;', 1),
    'generic-owner': ("return select(property, value, { ...finding, resolution_group: 'generic-owner' }, true);", 'return false;', 1),
    'unsupported-code': ("return contextual(diagnostic, finding, 'unsupported-code');", 'return false;', 1),
    'nested-expression': ("return contextual(diagnostic, finding, 'nested-expression');", 'return false;', 1),
    'nested-generic-result': ("return contextual(diagnostic, finding, 'nested-generic-result');", 'return false;', 2),
    'public-blocker': ('const isBlocked = owner => blockers().has(owner);', 'const isBlocked = owner => false;', 1),
    'late-exposure': ('function maskedConstructions(program, diagnostics) {', 'function maskedConstructions(program, diagnostics) { return false;', 1),
}


def identity(row):
    return tuple(row[key] for key in ('file', 'line', 'column', 'code', 'message'))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('scratch', type=Path)
    parser.add_argument('repaired_census', type=Path)
    parser.add_argument('--groups', nargs='+', choices=MUTATIONS, default=list(MUTATIONS))
    args = parser.parse_args()
    args.scratch.mkdir(parents=True, exist_ok=True)
    baseline = json.loads(args.repaired_census.read_text())
    table = json.loads((REPO / 'stage3/adapt/00-setup/evidence/checker-stops-2026-10-19.json').read_text())
    split = json.loads((REPO / 'stage3/adapt/00-setup/evidence/optional-contracts-2026-10-19.json').read_text())
    ids = {identity(row): row['id'] for row in table['diagnostics']}
    seed = Path(os.environ['ADAMIC_TYPESCRIPT_SOURCE']).resolve()
    original = HELPER.read_text()
    original_hash = hashlib.sha256(original.encode()).hexdigest()
    signal.signal(signal.SIGTERM, lambda *_: sys.exit('terminated; restoring resolver'))

    def run(command, log, seconds):
        print('start', log.name, flush=True)
        with log.open('w') as output:
            result = subprocess.run(['timeout', str(seconds)] + command, cwd=REPO, stdout=output, stderr=output, timeout=seconds + 10)
        if result.returncode:
            raise RuntimeError(f'{command}: exit {result.returncode}; {log}')
        print('passed', log.name, flush=True)

    for group in args.groups:
        directory = args.scratch / group
        if (directory / 'result.json').exists():
            print('already completed', group, flush=True)
            continue
        directory.mkdir(exist_ok=True)
        tree = directory / 'tree'
        if tree.exists():
            raise RuntimeError(f'partial tree exists: {tree}; preserve it and use a fresh scratch directory')
        run(['git', 'clone', '--shared', str(seed), str(tree)], directory / 'clone.log', 90)
        old, new, count = MUTATIONS[group]
        assert original.count(old) == count, (group, original.count(old), count)
        mutant = original.replace(old, new)
        try:
            HELPER.write_text(mutant)
            witness_results = []
            for witness in ('check-internal.cjs', 'check-nullable-return.cjs'):
                log = directory / (witness + '.log')
                with log.open('w') as output:
                    observed = subprocess.run(['timeout', '120', 'node', str(HELPER.parent / witness)], cwd=REPO,
                                              stdout=output, stderr=output, timeout=130)
                witness_results.append(dict(witness=witness, exit=observed.returncode))
            if group == 'nested-expression':
                assert all(row['exit'] == 0 for row in witness_results), (group, witness_results)
            else:
                assert any(row['exit'] == 1 for row in witness_results), (group, witness_results)
            assert all(row['exit'] in (0, 1) for row in witness_results), (group, witness_results)
            run(['python3', 'stage3/apply.py', '--adapt-checkout', str(tree)], directory / 'apply.log', 300)
        finally:
            HELPER.write_text(original)
        assert hashlib.sha256(HELPER.read_bytes()).hexdigest() == original_hash
        census = directory / 'census'
        run(['node', 'stage3/adapt/00-setup/evidence/checker-census.cjs', str(REPO), str(tree), str(census)], directory / 'census.log', 360)
        observed = json.loads((census / 'raw.json').read_text())
        wanted = {row['id'] for row in split['rows'] if row['reasonGroup'] == group and not any(owner['publicApi'] for owner in row['owners'])}
        runs = []
        for fixed, reverted in zip(baseline['runs'], observed['runs']):
            assert fixed['name'] == reverted['name']
            fixed_rows = {identity(row) for row in fixed['diagnostics']}
            reverted_rows = {identity(row) for row in reverted['diagnostics']}
            returned = sorted(ids[row] for row in reverted_rows - fixed_rows if row in ids)
            runs.append(dict(run=fixed['name'], count=len(reverted_rows), returned_ids=returned,
                             other_new_diagnostics=[list(row) for row in reverted_rows - fixed_rows if row not in ids],
                             removed_from_fixed=[ids.get(row, list(row)) for row in fixed_rows - reverted_rows]))
        primary = runs[0]
        if group == 'nested-expression':
            # Its measured construction is deferred because of public project.
            assert all(not row['returned_ids'] and not row['other_new_diagnostics'] and not row['removed_from_fixed'] for row in runs), runs
        elif group == 'public-blocker':
            assert len(primary['other_new_diagnostics']) == 3, primary
        else:
            assert wanted.intersection(primary['returned_ids']), (group, primary)
        result = dict(group=group, status='unchanged: scoped group declined; witness also reached by construction scan' if group == 'nested-expression' else 'caught',
                      witnesses=witness_results, original_helper_sha256=original_hash,
                      mutant_helper_sha256=hashlib.sha256(mutant.encode()).hexdigest(), runs=runs,
                      group_ids_returned=sorted(wanted.intersection(primary['returned_ids'])))
        (directory / 'result.json').write_text(json.dumps(result, indent=2) + '\n')
        print(json.dumps(result), flush=True)


if __name__ == '__main__':
    main()

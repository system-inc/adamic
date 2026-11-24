"""Drive the real case comparator on selected rows and their filename collisions."""
import argparse
import json
import os
from pathlib import Path
import sys

MUTANTS = Path(__file__).resolve().parent
sys.path.insert(0, str(MUTANTS.parent))
from run import ROOT, baseline_inputs, compare_case, measure_cases, upstream_tree
from driver import project_inputs, project_rows

VARIANCE = '001_varianceCantBeStrictWhileStructureIsnt'
EXPANDO = '012_expandoFunctionExpressionsWithDynamicNames2'
NEIGHBOR = '002_conditionalEqualityTestingNullability'
MODES = {
    'extra-file': (VARIANCE, 3545, 'unexpected.js', 'extra'),
    'missing-file': (VARIANCE, 3545, 'varianceCantBeStrictWhileStructureIsnt.js', 'missing'),
    'javascript-byte': (VARIANCE, 3545, 'varianceCantBeStrictWhileStructureIsnt.js', 'bytes'),
    'no-emit-write': (EXPANDO, 1226, 'unexpected.js', 'extra'),
    'timeout': (VARIANCE, 3545, None, 'native'),
    'node-timeout': (VARIANCE, 3545, None, 'Node'),
}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('output', type=Path)
    args = parser.parse_args()
    output = args.output.resolve()
    output.mkdir()
    upstream = upstream_tree()
    oracle = ['node', os.environ['STAGE3_VERDICT_NODE_TSC']]
    acceptance = {item[0]: item[1:] for item in project_rows(False)}
    baselines = json.loads((ROOT / 'selection.json').read_text())['cases']
    unexpected = []
    for mode, (target, index, path, expected) in MODES.items():
        candidate = MUTANTS / (mode + '.sh')
        reference = oracle
        if mode == 'node-timeout':
            candidate = ROOT / 'standins/node.sh'
            reference = ['node', str(MUTANTS / 'node-timeout.js')]

        def acceptance_inputs(case):
            row, source = acceptance[case]
            files, options = project_inputs(case, row, source)
            legacy = {stream: (source / ('golden.' + stream)).read_bytes()
                      for stream in ('stdout', 'stderr', 'exit')}
            return 'acceptance', case, files, options, legacy

        row, content, options, legacy = baseline_inputs(upstream, baselines[index - 1])
        shadow = f'{index:05d}_{Path(row["source"]).stem}'
        target_inputs = acceptance_inputs(target)
        assert list(target_inputs[2]) == [row['name']], (target, shadow)
        entries = [target_inputs, ('baselines', shadow, {row['name']: content}, options, legacy),
                   acceptance_inputs(NEIGHBOR)]

        def execute(entry):
            suite, case, files, config, golden = entry
            folder = output / mode / suite / case
            diff = compare_case(candidate, reference, folder, files, config, golden,
                                normalize_legacy=suite == 'baselines', suite=suite)
            return ({'case': case, 'differences': diff} if diff else None), config.get('noEmit') is True

        report = measure_cases(entries, execute)
        wanted = [{"case": target, "side": expected, "seconds": float(os.environ.get('TSC_TIMEOUT', '180'))}]
        if path:
            failures = report['failures']
            checks = {'files'}
            if mode == 'no-emit-write':
                checks.add('candidate_files')
            ok = (len(failures) == 1 and failures[0]['case'] == target
                  and set(failures[0]['differences']) == checks
                  and set(failures[0]['differences']['files']) == {path}
                  and failures[0]['differences']['files'][path]['kind'] == expected
                  and not report['timeouts'])
        else:
            ok = report['timeouts'] == wanted and not report['failures']
            project = output / mode / 'acceptance' / target
            timed_side = 'reference' if expected == 'Node' else 'candidate'
            ok &= (project / timed_side / 'actual.exit').read_text() == '124\n'
            if expected == 'Node':
                ok &= not (project / 'candidate').exists()
        print(json.dumps({'standin': mode, 'target': target, 'shadow': shadow,
                          'passing_neighbor': NEIGHBOR, 'proof_passed': ok, **report}), flush=True)
        if not ok:
            unexpected.append(mode)
    if unexpected:
        raise SystemExit('unexpected verdicts: ' + ', '.join(unexpected))


if __name__ == '__main__':
    main()

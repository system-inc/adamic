"""Compare captured verdicts without replacing their independent Node expectations."""
from collections import Counter
import json
from pathlib import Path
import re
import subprocess
import sys

ANSI = re.compile(rb'\x1b\[[0-9;]*m')
HEADER = re.compile(rb'^(.*?)\b(error|message|warning|suggestion) TS(\d+): (.*)$', re.M | re.I)


def diagnostic_records(stdout):
    plain = ANSI.sub(b'', stdout)
    plain = re.sub(rb'^([^\n]+?):(\d+):(\d+) - ', rb'\1(\2,\3): ', plain, flags=re.M)
    matches = list(HEADER.finditer(plain))
    records = []
    for index, match in enumerate(matches):
        end = matches[index + 1].start() if index + 1 < len(matches) else len(plain)
        records.append(((match[1].rstrip(b': -'), match[2].lower(), int(match[3])),
                        plain[match.start():end]))
    return records


def classify(left, right):
    """One primary cause per disagreement; exit/stderr differences remain separate tags."""
    if left['exit'] == b'124\n' or right['exit'] == b'124\n':
        return 'timeout'
    if left['stdout'] == right['stdout']:
        return 'stderr' if left['stderr'] != right['stderr'] else 'exit status'
    a, b = diagnostic_records(left['stdout']), diagnostic_records(right['stdout'])
    option_text = b'\n'.join(record[1] for record in a + b if 5000 <= record[0][2] < 7000)
    if re.search(rb'(Unknown compiler option|has been removed|not supported|not implemented|must be:)', option_text, re.I):
        return 'unsupported CLI option/value'
    if not a or not b:
        return 'checker diagnostic set' if a or b else 'non-diagnostic output'
    if Counter(record[1] for record in a) == Counter(record[1] for record in b):
        return 'diagnostic ordering'
    if Counter(record[0] for record in a) == Counter(record[0] for record in b):
        return 'message wording/presentation'
    if Counter((record[0][1], record[0][2]) for record in a) == Counter((record[0][1], record[0][2]) for record in b):
        return 'diagnostic locations'
    # Codes, categories or multiplicities differ. This is an observed
    # diagnostic-set difference, not proof which checker's implementation is wrong.
    return 'checker diagnostic set'


def captures(directory, suite, summary):
    if suite == 'baselines':
        rows = json.loads((directory / suite / 'selection.json').read_text())
        return [((row['source'], row.get('configuration', '')),
                 directory / suite / f'{index:05d}_{Path(row["source"]).stem}',
                 {key: row.get(key) for key in ('source_sha256', 'baseline_sha256', 'expected_sha256', 'options')})
                for index, row in enumerate(rows, 1)]
    folders = sorted(path for path in (directory / suite).iterdir()
                     if path.is_dir() and (path / 'actual.exit').exists())
    return [((path.name, ''), path, None) for path in folders]


def streams(folder, suite):
    return {name: (folder / ('actual.diagnostics' if suite == 'baselines' and name == 'stdout'
                            else 'actual.' + name)).read_bytes()
            for name in ('stdout', 'stderr', 'exit')}


def compare_results(left_dir, right_dir, output):
    from run import first_difference, write_json
    output.mkdir(parents=True, exist_ok=True)
    summaries = [json.loads((directory / 'summary.json').read_text()) for directory in (left_dir, right_dir)]
    if any(summary['harness_errors'] for summary in summaries):
        raise RuntimeError('cannot compare an incomplete verdict with harness errors')
    if summaries[0]['upstream_commit'] != summaries[1]['upstream_commit']:
        raise RuntimeError('comparison input pins differ')
    tables, disagreements = {}, []
    for suite in ('acceptance', 'tiny', 'baselines'):
        reports = [summary['suites'][suite] for summary in summaries]
        populations = [captures(directory, suite, summary) for directory, summary in zip((left_dir, right_dir), summaries)]
        identities = [[(row[0], row[2]) for row in population] for population in populations]
        if identities[0] != identities[1] or len(identities[0]) != reports[0]['total'] or reports[0]['total'] != reports[1]['total']:
            raise RuntimeError(f'{suite}: comparison populations or input hashes differ')
        fail_sets = [{(row['case'], row.get('configuration', '')) for row in report['failures']} for report in reports]
        if any(len(fail_set) != report['failed'] for fail_set, report in zip(fail_sets, reports)):
            raise RuntimeError(f'{suite}: duplicate failure identities')
        agreed = both_pass = both_fail = 0
        causes = Counter()
        for (identity, a_folder, _), (_, b_folder, _) in zip(*populations):
            a, b = [streams(folder, suite) for folder in (a_folder, b_folder)]
            differences = {name: first_difference(a[name], b[name]) for name in a if a[name] != b[name]}
            timed_out = a['exit'] == b'124\n' or b['exit'] == b'124\n'
            if not differences and not timed_out:
                agreed += 1
                both_pass += identity not in fail_sets[0] and identity not in fail_sets[1]
                both_fail += identity in fail_sets[0] and identity in fail_sets[1]
            else:
                cause = classify(a, b)
                causes[cause] += 1
                disagreements.append({'suite': suite, 'case': identity[0], 'configuration': identity[1],
                                      'cause': cause, 'streams': list(differences), 'differences': differences,
                                      'left_oracle_pass': identity not in fail_sets[0],
                                      'right_oracle_pass': identity not in fail_sets[1],
                                      'left_capture': str(a_folder), 'right_capture': str(b_folder)})
        tables[suite] = {'total': reports[0]['total'], 'agreement': agreed, 'disagreement': reports[0]['total'] - agreed,
                         'left_passed': reports[0]['passed'], 'right_passed': reports[1]['passed'],
                         'both_oracle_passed': both_pass, 'agreed_but_both_oracle_failed': both_fail,
                         'causes': dict(causes), 'excluded_inputs': reports[0]['excluded'],
                         'deferred': reports[0].get('deferred', 0)}
    grouped = Counter(row['cause'] for row in disagreements)
    examples = {cause: [row for row in disagreements if row['cause'] == cause][:3] for cause in grouped}
    result = {'schema_version': 1, 'left': summaries[0]['tsc'], 'right': summaries[1]['tsc'],
              'upstream_commit': summaries[0]['upstream_commit'], 'suites': tables,
              'cause_counts': {cause: grouped[cause] for cause in ('unsupported CLI option/value', 'checker diagnostic set', 'diagnostic locations', 'message wording/presentation', 'diagnostic ordering', 'exit status', 'stderr', 'timeout', 'non-diagnostic output')}, 'examples': examples, 'disagreements': disagreements,
              'success': all(row['disagreement'] == 0 and row['both_oracle_passed'] == row['total'] for row in tables.values())}
    write_json(output / 'comparison.json', result)
    lines = ['# Compiler agreement', '', '| Suite | Left oracle passes | Right oracle passes | Agreement | Disagreement |', '|---|---:|---:|---:|---:|']
    for suite, row in tables.items():
        lines.append(f'| {suite} | {row["left_passed"]}/{row["total"]} | {row["right_passed"]}/{row["total"]} | {row["agreement"]} | {row["disagreement"]} |')
    lines += ['', 'Agreement uses the verdict diagnostic projection for baselines and raw streams for driver suites. Both compilers are also independently judged against unchanged Node expectations. Two identical wrong outputs do not produce success. Tiny repeats one acceptance project.', '', '| Primary disagreement cause | Runs |', '|---|---:|']
    for cause, count in sorted(result['cause_counts'].items(), key=lambda item: -item[1]):
        lines.append(f'| {cause} | {count} |')
    lines += ['', 'Cause classification describes observed outputs. A changed diagnostic code/location/set can reflect checker, library, version, option-default or resolver behavior; it does not alone establish a checker bug. Secondary exit and stderr differences remain recorded. Ordering requires equal multisets of complete diagnostic blocks; wording requires equal header multisets. Unsupported options take precedence because they can prevent checking. Exact comparisons never use these classification relaxations.', '']
    for cause, rows in examples.items():
        lines += [f'## {cause}', '']
        for row in rows:
            lines.append(f'- {row["suite"]}: `{row["case"]}` `{row["configuration"]}`; streams: {", ".join(row["streams"])}')
        lines.append('')
    (output / 'comparison.md').write_text('\n'.join(lines) + '\n')
    return result


def compare_compilers(left, right, output, limit):
    from run import ROOT
    for label, binary in [('left', left), ('right', right)]:
        argv = [sys.executable, str(ROOT / 'run.py'), '--tsc', str(binary)]
        if limit is not None:
            argv += ['--baseline-limit', str(limit)]
        with (output / (label + '.log')).open('wb') as log:
            completed = subprocess.run([*argv, str(output / label)], stdout=log, stderr=log)
        if completed.returncode not in (0, 1):
            raise RuntimeError(f'{label} verdict has infrastructure errors; see {output / (label + ".log")}')
    result = compare_results(output / 'left', output / 'right', output)
    print((output / 'comparison.md').read_text(), end='')
    return 0 if result['success'] else 1


if __name__ == '__main__':
    import argparse
    parser = argparse.ArgumentParser(description='Compare two existing captured verdicts without rerunning them')
    parser.add_argument('left', type=Path)
    parser.add_argument('right', type=Path)
    parser.add_argument('output', type=Path)
    args = parser.parse_args()
    if args.output.exists():
        parser.error('output directory must be new')
    result = compare_results(args.left.resolve(), args.right.resolve(), args.output.resolve())
    print((args.output / 'comparison.md').read_text(), end='')
    raise SystemExit(0 if result['success'] else 1)

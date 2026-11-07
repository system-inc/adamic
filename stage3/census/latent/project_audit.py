"""Bounded entry-project audit used by audit_deferred.py --project."""
from collections import Counter
import json
import os
from pathlib import Path
import subprocess
import time
from unit_status import audit_deferred, unit_counts


def run_project(binary, output, project, baseline, seconds):
    assert 0 < seconds < 1500, 'each run must be under 25 minutes'
    assert project.is_file(), project
    output.mkdir(parents=True, exist_ok=True)

    deadline = time.monotonic() + seconds

    def run(name, executable, speculative):
        started = time.monotonic()
        destination = output / (name + '.jsonl')
        destination.unlink(missing_ok=True)
        env = dict(os.environ, LATENT_ASSERT_NO_OUTPUT='1', LATENT_FULL='0',
                   LATENT_SPECULATIVE='1' if speculative else '0')
        with (output / (name + '.log')).open('w') as log:
            try:
                result = subprocess.run([str(executable), str(project), str(destination)],
                                        env=env, stdout=log, stderr=subprocess.STDOUT,
                                        timeout=max(0.1, deadline - time.monotonic()))
                code = result.returncode
            except subprocess.TimeoutExpired:
                code = 'timeout'
        rows, partial_line = [], False
        lines = destination.read_text().splitlines() if destination.exists() else []
        for index, line in enumerate(lines):
            try:
                rows.append(json.loads(line))
            except json.JSONDecodeError:
                if code != 0 and index == len(lines) - 1:
                    partial_line = True
                    break
                raise
        files = rows[1:]
        units = [unit for row in files for unit in row['units']]
        counts = unit_counts(units)
        audit_deferred(counts, units)
        findings = [finding for row in files for finding in row['findings']
                    if 'latent state copy:' in finding.get('text', '') or
                    'latent state copy:' in finding.get('reason', '')]
        unhandled = sum(line.startswith('panic: latent state copy:') for line in
                        (output / (name + '.log')).read_text().splitlines())
        report = dict(exit=code, wall_seconds=time.monotonic() - started, complete=code == 0 and not partial_line, partial_line=partial_line, file_records=len(files),
                      units=counts, units_by_kind=dict(Counter(u['kind'] for u in units)),
                      state_copy_panic_findings=len(findings),
                      unhandled_state_copy_panics=unhandled)
        return report, files

    # All children share this invocation's deadline; each retains its raw ledger.
    result, fixed = run('speculative', binary, True)
    report = dict(project=str(project), scope='one entry root, imported program retained by loader',
                  timeout_seconds_total=seconds, speculative=result)
    if baseline:
        base_report, base = run('main-speculative', baseline, True)
        report['main_speculative'] = base_report
        shared = {u['where'] for r in fixed for u in r['units']} & {u['where'] for r in base for u in r['units']}
        def records(files):
            return {u['where']: dict(unit=u, findings=[f for f in row['findings'] if f.get('unit') == u['where']])
                    for row in files for u in row['units']}
        left, right = records(fixed), records(base)
        report['shared_speculative_units'] = len(shared)
        report['different_shared_unit_records'] = sorted(k for k in shared if left[k] != right[k])
        report['speculative_comparison_available'] = bool(shared)
        first_report, first = run('first-error', binary, False)
        base_first_report, base_first = run('main-first-error', baseline, False)
        report['first_error'] = first_report
        report['main_first_error'] = base_first_report
        report['first_error_byte_identical'] = (first_report['complete'] and base_first_report['complete'] and
            (output / 'first-error.jsonl').read_bytes() == (output / 'main-first-error.jsonl').read_bytes())
    report['passed'] = (result['complete'] and result['file_records'] > 0 and
                        result['state_copy_panic_findings'] == 0 and result['unhandled_state_copy_panics'] == 0)
    if baseline:
        report['passed'] = report['passed'] and report['first_error_byte_identical']
    (output / 'project-summary.json').write_text(json.dumps(report, indent=2) + '\n')
    print(json.dumps(report, indent=2))
    return 0 if report['passed'] else 1

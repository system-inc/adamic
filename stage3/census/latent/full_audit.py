"""Exercise real lowering continuation, diagnosed nested exclusions, and rollback."""
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile


def check(rows):
    units = [unit for row in rows[1:] for unit in row['units']]
    findings = [finding for row in rows[1:] for finding in row['findings']]
    target = next(unit for unit in units if unit.get('name') == 'three')
    assert target['status'] == 'attempted', target
    actual = {(finding['where'], finding['reason']) for finding in findings
              if finding['unit'] == target['where'] and finding['phase'] == 'lowering'
              and finding['kind'] == 'Refused'}
    # String and number conditions now lower with JavaScript truthiness. Keep
    # three separate refusal sites so the first-error mutant still loses two.
    assert {reason for _, reason in actual} == {'var'} and len(actual) == 3, actual
    excluded = next(unit for unit in units if unit.get('name') == 'badNested')
    assert excluded['depth'] == 1 and excluded['parent'] and excluded['body_end'] > excluded['body_start'], excluded
    assert excluded['status'] == 'split_checker_body' and excluded['checker_diagnostics'], excluded
    assert not any(finding['unit'] == excluded['where'] for finding in findings), excluded
    good = next(unit for unit in units if unit.get('name') == 'cleanNested')
    assert good['status'] == 'attempted' and good['parent'] == excluded['parent'], good
    assert any(finding['kind'] == 'Refused' and finding['reason'] == 'var'
               and finding['unit'] == good['where'] for finding in findings), findings
    rollback = next(unit for unit in units if unit.get('name') == 'rollback')
    assert any(finding['kind'] == 'NotYet' and 'reading poison' in finding['reason']
               and finding['unit'] == rollback['where'] for finding in findings), findings
    middle = next(unit for unit in units if unit.get('name') == 'middle')
    assert middle['status'] == 'attempted', middle
    actual = {(finding['where'], finding['reason']) for finding in findings
              if finding['unit'] == middle['where'] and finding['phase'] == 'lowering'
              and finding['kind'] == 'Refused'}
    assert {reason for _, reason in actual} == {'var'} and len(actual) == 3, actual
    boundaries = [finding for finding in findings if finding['kind'] == 'Boundary']
    assert boundaries and all(finding.get('end', 0) > finding.get('start', 0) for finding in boundaries), boundaries


def audit(binary, output):
    output.mkdir(parents=True, exist_ok=True)
    source = output / 'source'
    source.mkdir(exist_ok=True)
    (source / 'probe.a').write_text('''declare function external(): number;
export function three(text: string, count: number): void {
    if (text) { }
    if (count) { }
    var old = 1;
    var second = 2;
    var third = 3;
    const okay: number = 2;
}
export function diagnosed(): void {
    const impossible: number = "wrong";
    function cleanNested(): void { var x = 2; }
    function badNested(): number { return "wrong"; }
}
export function rollback(): void {
    const poison: number = external();
    if (poison) { }
}
// A generic child is not instantiated in the enclosing prologue. Its diagnosed
// body stays separately excluded while middle reaches its continuation witness.
export function outerOnly(): void {
    function middle(text: string, count: number): void {
        function grandBad<Unused>(): number { return "wrong"; }
        if (text) { }
        if (count) { }
        var x = 1;
        var second = 2;
        var third = 3;
    }
}
''')
    observations = {}
    for name, extra in [('full', {}), ('first-error-mutant', {'LATENT_MUTANT_FIRST_ERROR_ONLY': '1'}), ('rollback-mutant', {'LATENT_MUTANT_KEEP_FAILED_STATE': '1'})]:
        destination = output / (name + '.jsonl')
        with (output / (name + '.log')).open('w') as log:
            subprocess.run([str(binary), str(source), str(destination)],
                           env=dict(os.environ, LATENT_FULL='1', LATENT_ASSERT_NO_OUTPUT='1', **extra),
                           stdout=log, stderr=subprocess.STDOUT, check=True)
        observations[name] = [json.loads(line) for line in destination.read_text().splitlines()]
    check(observations['full'])
    try:
        check(observations['first-error-mutant'])
    except AssertionError as failure:
        print('first-error-only mutant caught by all-three-refusal-sites assertion:', failure)
    else:
        raise AssertionError('first-error-only mutant survived')
    try:
        check(observations['rollback-mutant'])
    except AssertionError:
        print('failed-state-retention mutant caught by poison binding rollback assertion')
    else:
        raise AssertionError('failed-state-retention mutant survived')
    print('PASS: three distinct lowering refusal sites; accepted string and number conditions; named checker-diagnosed nested exclusion; eligible sibling; failed declaration rolled back; explicit byte-span boundaries; no-output guards')


if __name__ == '__main__':
    binary = Path(sys.argv[1]).resolve()
    if len(sys.argv) > 2:
        audit(binary, Path(sys.argv[2]).resolve())
    else:
        with tempfile.TemporaryDirectory(prefix='latent-full-audit-') as temporary:
            audit(binary, Path(temporary))

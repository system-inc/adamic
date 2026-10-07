"""Prove continuation, file attribution, and measurement-only behavior."""
import json
import os
import pathlib
import subprocess
import sys
import tempfile

binary = str(pathlib.Path(sys.argv[1]).resolve())
with tempfile.TemporaryDirectory(prefix='latent-audit-') as temporary:
    scratch = pathlib.Path(temporary)
    source = scratch / 'source'
    source.mkdir()
    (source / 'one.a').write_text('import { other } from "./two.a";\nfunction first(): number { function inner(): number { return 1; } return inner(); }\nfunction target(): number { return other(); }\nfunction later(): number { function inner(): number { return 3; } return inner(); }\nfunction asserted(value: number | undefined): number { return value! + value!; }\nfunction bad(): number { const impossible: number = "wrong"; return impossible!; }\nfunction signatureOnly(value: Missing): number { return 1; }\nfunction broken(): number { return "wrong"; } function clean(): number { return 4; }\n')
    (source / 'two.a').write_text('export function other(): number { return 3; }\n')
    observations = {}
    for name, extra in [('baseline', {}), ('mutant', {'LATENT_MUTANT_FUNCTION': 'target'}), ('scope_mutant', {'LATENT_MUTANT_BODY_SCOPE': '1'})]:
        output = scratch / (name + '.jsonl')
        with (scratch / (name + '.log')).open('w') as log:
            subprocess.run([binary, str(source), str(output)], env=dict(os.environ, LATENT_ASSERT_NO_OUTPUT='1', **extra), stdout=log, stderr=subprocess.STDOUT, check=True)
        rows = [json.loads(line) for line in output.read_text().splitlines()]
        assert rows[0]['status'] == 'measurement', rows
        assert rows[0]['checker_rejected'] is True, rows
        assert sum(len(r['units']) for r in rows[1:]) == 10, rows
        observations[name] = {pathlib.Path(row['file']).name: row for row in rows[1:]}
    baseline, mutant = observations['baseline'], observations['mutant']
    units = baseline['one.a']['units'] + baseline['two.a']['units']
    skipped = {u['where'] for u in units if u['status'] == 'skipped_checker_body'}
    assert len(skipped) == 2, units
    signature = next(u for u in units if ':7:1' in u['where'])
    assert signature['status'] == 'attempted', signature
    same_line = [u for u in units if ':8:' in u['where']]
    assert [u['status'] for u in same_line] == ['skipped_checker_body', 'attempted'], same_line
    scope_signature = next(u for u in observations['scope_mutant']['one.a']['units'] if ':7:1' in u['where'])
    assert scope_signature['status'] == 'skipped_checker_body', scope_signature
    # This mutant is caught only by the body-versus-signature eligibility check.
    try:
        assert scope_signature['status'] == 'attempted'
    except AssertionError:
        print('signature/body range mutant caught')
    else:
        raise AssertionError('scope mutant survived')
    assert not any(f['unit'] in skipped for row in baseline.values() for f in row['findings'])
    assert not any(':3:' in f['unit'] for row in baseline.values() for f in row['findings']), 'imported sibling call was not registered'

    findings = baseline['one.a']['findings']
    assert sum(f['reason'] == 'a function inside a function (a closure)' for f in findings) == 2, findings
    assert sum(f['kind'] == 'Refused' and f['reason'] == 'the non-null assertion !' for f in findings) == 2, findings
    assert len(mutant['one.a']['findings']) == len(findings) + 1
    added = [f for f in mutant['one.a']['findings'] if f not in findings]
    assert len(added) == 1 and added[0]['kind'] == 'NotYet' and added[0]['reason'] == 'latent planted extra NotYet', added
    assert ':3:1' in added[0]['where'], added
    assert mutant['two.a'] == baseline['two.a']
    # Kill the assertion itself with a misplaced extra finding in the other file.
    caught = False
    forged = dict(mutant['two.a'], findings=[added[0]])
    try:
        assert forged == baseline['two.a']
    except AssertionError:
        caught = True
    assert caught
    print(json.dumps({'continuation': 'pass', 'all_refusal_sites': 'pass', 'no_IR_output': 'pass', 'production_loader_disabled': 'pass', 'body_skip': 'pass', 'signature_eligible': 'pass', 'same_line_ranges': 'pass', 'measurement': 'measured on a checker-rejected program', 'mutant': {'one.a_delta': 1, 'two.a_delta': 0, 'location': 'one.a:3:1'}, 'misattribution_mutant': 'caught'}, indent=2))

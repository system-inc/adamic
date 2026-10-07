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
    (source / 'one.a').write_text('function first(): number { function inner(): number { return 1; } return inner(); }\nfunction target(): number { return 2; }\nfunction later(): number { function inner(): number { return 3; } return inner(); }\nfunction asserted(value: number | undefined): number { return value! + value!; }\n')
    (source / 'two.a').write_text('function other(): number { return 3; }\n')
    observations = {}
    for name, extra in [('baseline', {}), ('mutant', {'LATENT_MUTANT_FUNCTION': 'target'})]:
        output = scratch / (name + '.jsonl')
        with (scratch / (name + '.log')).open('w') as log:
            subprocess.run([binary, str(source), str(output)], env=dict(os.environ, LATENT_ASSERT_NO_OUTPUT='1', **extra), stdout=log, stderr=subprocess.STDOUT, check=True)
        rows = [json.loads(line) for line in output.read_text().splitlines()]
        assert rows[0]['status'] == 'measurement', rows
        assert rows[0]['units'] == 5, rows
        observations[name] = {pathlib.Path(row['file']).name: row for row in rows[1:]}
    baseline, mutant = observations['baseline'], observations['mutant']
    assert all(row['status'] == 'measured' for row in baseline.values())
    findings = baseline['one.a']['findings']
    assert sum(f['reason'] == 'a function inside a function (a closure)' for f in findings) == 2, findings
    assert sum(f['kind'] == 'Refused' and f['reason'] == 'the non-null assertion !' for f in findings) == 2, findings
    assert len(mutant['one.a']['findings']) == len(findings) + 1
    added = [f for f in mutant['one.a']['findings'] if f not in findings]
    assert len(added) == 1 and added[0]['kind'] == 'NotYet' and added[0]['reason'] == 'latent planted extra NotYet', added
    assert ':2:1' in added[0]['where'], added
    assert mutant['two.a'] == baseline['two.a']
    # Kill the assertion itself with a misplaced extra finding in the other file.
    caught = False
    forged = dict(mutant['two.a'], findings=[added[0]])
    try:
        assert forged == baseline['two.a']
    except AssertionError:
        caught = True
    assert caught
    print(json.dumps({'continuation': 'pass', 'all_refusal_sites': 'pass', 'no_IR_output': 'pass', 'mutant': {'one.a_delta': 1, 'two.a_delta': 0, 'location': 'one.a:2:1'}, 'misattribution_mutant': 'caught'}, indent=2))

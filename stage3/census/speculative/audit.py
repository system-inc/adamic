"""Focused real-lowering controls and mutants. Every invocation writes a log."""
import json
import os
from pathlib import Path
import subprocess
import sys

binary = Path(sys.argv[1]).resolve()
output = Path(sys.argv[2]).resolve()
output.mkdir(parents=True, exist_ok=True)
source = output / 'source'
source.mkdir(exist_ok=True)
(source / 'nested.a').write_bytes((Path(__file__).parent / 'control/nested.a').read_bytes())


def run(name, **extra):
    destination = output / (name + '.jsonl')
    env = dict(os.environ, LATENT_FULL='1', LATENT_ASSERT_NO_OUTPUT='1')
    env.pop('LATENT_SPECULATIVE', None)
    env.pop('LATENT_MUTANT_NO_STUBS', None)
    env.pop('LATENT_MUTANT_DEPTH_ZERO', None)
    env.update(extra)
    with (output / (name + '.log')).open('w') as log:
        subprocess.run([str(binary), str(source), str(destination)], env=env,
                       stdout=log, stderr=subprocess.STDOUT, check=True)
    return [json.loads(line) for line in destination.read_text().splitlines()]


def check(rows):
    assert rows[0]['latent_mode'] == 'speculative'
    findings = [f for r in rows[1:] for f in r['findings']]
    for kind, reason in [('NotYet', 'a WithStatement'), ('Refused', 'with')]:
        actual = [(int(f['where'].rsplit(':', 2)[1]), f['depth']) for f in findings
                  if f['kind'] == kind and f['reason'] == reason]
        assert actual == [(3, 0), (4, 1)], actual
    for kind, reason in [('NotYet', 'a DebuggerStatement'), ('Refused', 'debugger')]:
        actual = [(int(f['where'].rsplit(':', 2)[1]), f['depth']) for f in findings
                  if f['kind'] == kind and f['reason'] == reason]
        assert actual == [(5, 2), (8, 0)], actual
    assert all(f['root_kind'] == f['reason'] for f in findings)
    assert all(r['speculative_coverage']['unvisited_nodes'] == 0 for r in rows[1:])
    return sum(f['kind'] in ('NotYet', 'Refused') for f in findings)


speculative = run('speculative', LATENT_SPECULATIVE='1')
count = check(speculative)
baseline = run('full')
no_stubs = run('no-stubs-mutant', LATENT_SPECULATIVE='1', LATENT_MUTANT_NO_STUBS='1')
assert no_stubs == baseline, 'no-stubs mutant did not restore the exact current full census'
baseline_count = sum(f['kind'] in ('NotYet', 'Refused') for r in baseline[1:] for f in r['findings'])
assert count > baseline_count, (count, baseline_count)
for name, mutant in [('no-stubs', no_stubs), ('depth-zero', run('depth-zero-mutant', LATENT_SPECULATIVE='1', LATENT_MUTANT_DEPTH_ZERO='1'))]:
    try:
        check(mutant)
    except AssertionError as failure:
        print(name + ' mutant caught by nested depth/continuation control:', failure)
    else:
        raise AssertionError(name + ' mutant survived')
print(f'PASS: exact depths 0,1,2; sibling depth 0; {count} speculative sites versus {baseline_count} full sites; no-stubs byte-identical full JSON; all AST nodes visited; no-output guards')

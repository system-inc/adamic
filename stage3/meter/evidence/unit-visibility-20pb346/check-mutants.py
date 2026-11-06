"""Run named regression tests against three temporary source mutants; always restore."""
from pathlib import Path
import difflib
import json
import subprocess

repo = Path(__file__).resolve().parents[4]
output = Path(__file__).resolve().parent
cases = [
    ('entry-times-mutant', 'stage3/meter/twice-daily.sh',
     'LATENT_TIMES="$run/$label/tsc/latent-times.jsonl" ', '',
     ['python3', '-E', '-S', '-B', '-m', 'unittest',
      'compiler_test.CompilerSelectionTests.test_entry_times_sidecar_preserves_report_20pb346'],
     repo / 'stage3/meter'),
    ('extra-skip-mutant', 'stage3/census/latent/privatefields/plan.go',
     'func (plan Plan) Skip(value reflect.Value, index int) bool {',
     'func (plan Plan) Skip(value reflect.Value, index int) bool {\n if index == 0 { return true }',
     ['go', 'test', '-count=1', './stage3/census/latent/privatefields', '-run',
      '^TestLatentPrivateIRPlanPreservesUnlistedFields20pb346$'], repo),
    ('private-nil-mutant', 'stage3/census/latent/privatefields/plan.go',
     'if f.policy == Nil && !value.Field(index).IsNil() {',
     'if false && f.policy == Nil && !value.Field(index).IsNil() {',
     ['go', 'test', '-count=1', './stage3/census/latent/privatefields', '-run',
      '^TestLatentPrivateIRPlanRejectsPopulatedNilCache20pb346$'], repo),
]
results = []
for name, relative, old, new, command, cwd in cases:
    path = repo / relative
    original = path.read_text()
    assert original.count(old) == 1, name
    changed = original.replace(old, new, 1)
    (output / (name + '.patch')).write_text(''.join(difflib.unified_diff(
        original.splitlines(True), changed.splitlines(True),
        fromfile='a/' + relative, tofile='b/' + relative)))
    try:
        path.write_text(changed)
        with (output / (name + '.txt')).open('w') as log:
            result = subprocess.run(command, cwd=cwd, stdout=log,
                                    stderr=subprocess.STDOUT, timeout=180)
        results.append(dict(mutant=name, command=command, exit_code=result.returncode,
                            killed=result.returncode != 0))
        assert result.returncode != 0, name + ' survived'
    finally:
        path.write_text(original)
(output / 'mutant-results.json').write_text(json.dumps(results, indent=2) + '\n')

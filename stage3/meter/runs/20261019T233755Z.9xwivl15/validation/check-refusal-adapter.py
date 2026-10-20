#!/usr/bin/env python3
"""Validate exact-shape preflight support and fail-closed visitor checks."""
from pathlib import Path
import subprocess
import tempfile

run = Path(__file__).resolve().parents[1]
scratch = Path(tempfile.mkdtemp(prefix='meter-refusal-'))
program = scratch / 'rewrite.go'
program.write_text((run / 'validation/refusalrewrite-run.go.txt').read_text())
binary = scratch / 'rewrite'
with (run / 'validation/refusal-adapter-build.log').open('w') as log:
    subprocess.run(['go', 'build', '-o', str(binary), str(program)], stdout=log, stderr=log, check=True, timeout=120)
original = Path('/workspace/adamic/internal/lower/refusals.go').read_text()
for name, source, expected in [
    ('current', original, 0),
    ('unknown-preflight', original.replace('l.referenceRefusal(module)', 'l.unknownPreflight(module)'), 1),
    ('missing-child-walk', original.replace('node.ForEachChild(visit)', '_ = node'), 1),
]:
    source_path = scratch / (name + '.go')
    output = scratch / (name + '-out.go')
    source_path.write_text(source)
    log_path = run / 'validation' / ('refusal-adapter-' + name + '.log')
    with log_path.open('w') as log:
        result = subprocess.run([str(binary), '-input', str(source_path), '-output', str(output)], stdout=log, stderr=log, timeout=120)
    assert result.returncode == expected, (name, log_path.read_text())
    if expected:
        assert not output.exists()
        assert 'latent refusal rewrite: lowering.refuse:' in log_path.read_text()
    else:
        text = output.read_text()
        start = text.index('func (l *lowering) latentRefuse(')
        assert text[start:].count('latentRecord(err)') == 2
        assert text[:start].count('l.referenceRefusal(module)') == 1
        assert text[:start].count('l.parallelPreflight(module)') == 1
    print(name + ': ' + ('accepted with two collected preflight errors' if not expected else 'rejected before output'))

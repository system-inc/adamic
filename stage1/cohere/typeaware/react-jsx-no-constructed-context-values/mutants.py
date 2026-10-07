#!/usr/bin/env python3
"""Prove the memo factory and both escape checks can fail only the byte oracle."""
import argparse
from pathlib import Path
import re
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('--artifacts', type=Path, required=True)
parser.add_argument('--controls', type=Path, required=True)
parser.add_argument('--compiler', type=Path, required=True)
parser.add_argument('--checker', type=Path, required=True)
args = parser.parse_args()
own = Path(__file__).resolve().parent
out = args.artifacts.resolve()
out.mkdir(parents=True, exist_ok=True)
sequence = 0

def run(label, command):
    global sequence
    sequence += 1
    stem = out / f'{sequence:03d}-{label}'
    with stem.with_suffix('.stdout').open('wb') as stdout, stem.with_suffix('.stderr').open('wb') as stderr:
        result = subprocess.run(list(map(str, command)), stdout=stdout, stderr=stderr, timeout=600)
    assert result.returncode == 0, (label, result.returncode)
    return stem.with_suffix('.stdout').read_bytes(), stem.with_suffix('.stderr').read_bytes()

def imports(source, changed=None):
    return re.sub(r"from '([^']+)'", lambda match: "from '" + str(changed.get(match[1], (own / match[1]).resolve())) + "'" if match[1].startswith('.') else match[0], source)

source = (own / 'stability.a').read_text()
for label, before, after in [
    ('missing-list', 'if(args.length === 1)', 'if(args.length === 0)'),
    ('render-escape', '!this.escapes(result.holders, true)', 'true'),
    ('helper-escape', 'this.escapes(own, false)', 'false'),
    ('factory-identity', 'result.fresh === undefined ? new MemoFinding()', 'false ? new MemoFinding()'),
]:
    assert source.count(before) == 1, (label, source.count(before))
    stability = out / (label + '-stability.a')
    stability.write_text(imports(source.replace(before, after), {}))
    analysis = out / (label + '-analysis.a')
    analysis.write_text(imports((own / 'analysis.a').read_text(), {'./stability.a': stability}))
    entry = out / (label + '-suite.a')
    entry.write_text(imports((own / 'suite.a').read_text(), {'./analysis.a': analysis}))
    binary = out / label
    run(label + '-build', [args.compiler, 'build', entry, '-o', binary, '--tsgo', args.checker])
    for project in sorted(args.controls.glob('case-*')):
        command = [project / 'tsconfig.json', project / 'manifest']
        truth, _ = run(label + '-go', [args.controls / 'oracle', *command])
        got, stderr = run(label + '-native', [binary, *command])
        assert not stderr, (label, project)
        if got != truth:
            print('KILLED', label, project.name, 'compiled, exit 0, empty stderr; only Go finding bytes differ', flush=True)
            break
    else:
        raise AssertionError('memo check mutant survived: ' + label)
    binary.unlink()
print('PASS four compiling memo/escape mutants', flush=True)

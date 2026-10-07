#!/usr/bin/env python3
"""Prove the inherited generator accepts TS and rejects identical .a source."""
from pathlib import Path
import shutil
import subprocess
import tempfile

repository = Path(__file__).resolve().parents[5]
with tempfile.TemporaryDirectory(prefix='lint-wave1-09-a-') as directory:
    root = Path(directory)
    shutil.copytree(repository / 'stage1/cohere/lint/rules', root / 'rules')
    def generate():
        return subprocess.run(['go', 'run', './cmd/lint-registry', '-root', str(root)],
                              cwd=repository, text=True, capture_output=True)
    baseline = generate()
    print('baseline exit', baseline.returncode)
    print(baseline.stdout, end='')
    print(baseline.stderr, end='')
    assert baseline.returncode == 0
    assert set(baseline.stdout.splitlines()) == {
        'no-debugger', 'no-empty', 'eqeqeq', 'no-var', 'no-duplicate-case'}
    old = root / 'rules/no-debugger/rule.ts'
    data = old.read_bytes()
    target = old.with_suffix('.a')
    old.rename(target)
    assert data == target.read_bytes()
    changed = generate()
    print('identical module renamed .a exit', changed.returncode)
    print(changed.stdout, end='')
    print(changed.stderr, end='')
    assert changed.returncode != 0
    assert 'rule.ts: no such file or directory' in changed.stderr
    print('Observed: .a registration is blocked before rule compilation or execution.')

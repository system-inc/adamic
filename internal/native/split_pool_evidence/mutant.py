#!/usr/bin/env python3
"""The eight-build check must fail when token acquisition/release is bypassed."""
import json
import pathlib
import subprocess
import tempfile

root = pathlib.Path(__file__).resolve().parents[3]
source = root / 'internal/native/clang.go'
with tempfile.TemporaryDirectory(prefix='adamic-clang-mutant-') as scratch:
    scratch = pathlib.Path(scratch)
    mutant = scratch / 'clang.go'
    text = source.read_text()
    assert 'clangTokens <- struct{}{}' in text and '<-clangTokens' in text
    mutant.write_text(text.replace('clangTokens <- struct{}{}', '// mutant: bypass acquisition')
                     .replace('<-clangTokens', '/* mutant: bypass release */'))
    overlay = scratch / 'overlay.json'
    overlay.write_text(json.dumps({'Replace': {str(source): str(mutant)}}))
    result = subprocess.run(['go', 'test', '-overlay=' + str(overlay), './internal/native',
                             '-run=^TestClangPoolConcurrentBuilds$', '-count=1', '-v'],
                            cwd=root, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
    print(result.stdout, end='')
    if result.returncode == 0 or 'peak clangs=' not in result.stdout or '--- FAIL: TestClangPoolConcurrentBuilds' not in result.stdout:
        raise SystemExit('mutant was not caught by the concurrency assertion')
    print('PASS: semaphore-bypass mutant caught by concurrency assertion')

#!/usr/bin/env python3
"""Audit tracked documentation paths; --rewrite applies the folder migration."""
import pathlib
import re
import subprocess
import sys

root = pathlib.Path(__file__).resolve().parent.parent
old = 'docs' + '/'
pattern = re.compile(r'[A-Za-z0-9_./:@+-]*' + old + r'[A-Za-z0-9_./@+-]*')
fixtures = {
    'stage1/cohere/gitignore/sample-cases.txt',
    'stage1/cohere/gitignore/testdata/cohere_side_test.go',
}


def unrelated(path, token):
    token = token.rstrip('.')
    # These are upstream paths, synthetic gitignore inputs, and throughput units.
    return (
        token.startswith('cohere/' + old)
        or token.startswith('compiler/packages/babel-plugin-react-compiler/' + old)
        or (path in fixtures and token in {old, old + 'generated', 'n' + old + 'generated', 'sub/' + old + 'generated'})
        or (path == 'stage1/cohere/markdownblocks/REPORT.txt' and token == old + 's')
    )


rewrite = sys.argv[1:] == ['--rewrite']
if sys.argv[1:] and not rewrite:
    raise SystemExit('usage: check-documentation-paths.py [--rewrite]')
paths = subprocess.check_output(['git', 'ls-files', '-z'], cwd=root).decode().split('\0')
matched = changed = 0
for path in paths:
    if not path or path == 'cohere' or path.startswith('cohere/'):
        continue
    file = root / path
    data = file.readlink().as_posix().encode('utf-8') if file.is_symlink() else file.read_bytes()
    if b'\0' in data:
        continue
    try:
        text = data.decode('utf-8')
    except UnicodeDecodeError:
        continue
    matches = [m for m in pattern.finditer(text) if not unrelated(path, m.group())]
    if not matches:
        continue
    matched += len(matches)
    changed += 1
    for match in matches:
        line = text.count('\n', 0, match.start()) + 1
        print(f'{path}:{line}:{match.group()}')
    if rewrite:
        for match in reversed(matches):
            text = text[:match.start()] + match.group().replace(old, 'documentation/') + text[match.end():]
        file.write_bytes(text.encode('utf-8'))
if rewrite:
    print(f'replaced {matched} repo paths in {changed} files')
else:
    raise SystemExit(1 if matched else 0)

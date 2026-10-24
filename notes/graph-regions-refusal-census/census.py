#!/usr/bin/env python3
"""Compare file-backed refusal promises; run the Go harnesses for inline cases too."""
import argparse
import concurrent.futures
import json
from pathlib import Path
import subprocess

parser = argparse.ArgumentParser(__doc__)
parser.add_argument('--before', required=True, type=Path)
parser.add_argument('--after', required=True, type=Path)
parser.add_argument('--root', type=Path, default=Path('.'))
parser.add_argument('--records-ref', default='749e668e1d7aa32b0ed716b3f17ab6cba94b6f13')
parser.add_argument('--workers', type=int, default=2)
parser.add_argument('--output', type=Path, required=True)
args = parser.parse_args()
root = args.root.resolve()
compilers = [args.before.resolve(), args.after.resolve()]


def git(*arguments):
    return subprocess.check_output(['git', '-C', str(root), *arguments], timeout=30)


def recorded(path):
    return git('show', args.records_ref + ':' + path).decode()


files = git('ls-tree', '-r', '--name-only', args.records_ref).decode().splitlines()
programs = set()
for path in files:
    p = Path(path)
    if path.startswith('internal/fuzz/'):
        continue
    if p.suffix in ('.pending', '.what') and (root / p.with_suffix('.a')).is_file():
        programs.add(str(p.with_suffix('.a')))
    if path.startswith('internal/oracle/testdata/') and p.suffix == '.a' and any(
            'refus' in part or 'notyet' in part.lower() for part in p.parts):
        programs.add(path)
    if p.name == 'status.json' and p.parent.parent == Path('stage3/fixtures'):
        for entry in json.loads(recorded(path)):
            if entry['stage0']['outcome'] in ('Refused', 'NotYet'):
                programs.add(str(p.parent / entry['file']))

for line in recorded('internal/acheck/known.txt').splitlines():
    if line.strip() and not line.startswith('#'):
        programs.add(line.split(': ', 1)[0])

# Include explicit header promises even inside Go-owned fixtures, where acheck
# delegates to the owning package. Dated evidence keeps acheck's exact exemption.
headers = subprocess.run(
    ['git', '-C', str(root), 'grep', '-l', '-z', '-E',
     '^// (a-check: refused|Expect: refused)', args.records_ref, '--', '*.a', '*.ts'],
    capture_output=True, timeout=30, check=False)
if headers.returncode not in (0, 1):
    raise RuntimeError(headers.stderr.decode())
for name in headers.stdout.decode().split('\0'):
    if not name:
        continue
    path = name.split(':', 1)[1]
    if not path.startswith(('internal/fuzz/', 'notes/', 'review/', 'docs/verification/')):
        programs.add(path)
programs = sorted(path for path in programs if (root / path).is_file())


def check(path):
    observations = []
    for compiler in compilers:
        result = subprocess.run([str(compiler), 'c', str(root / path)], cwd=root,
                                stdout=subprocess.DEVNULL, stderr=subprocess.PIPE,
                                timeout=60, check=False)
        observations.append({'exit': result.returncode,
                             'diagnostic': result.stderr.decode(errors='replace')})
    return {'path': path, 'before': observations[0], 'after': observations[1]}


print(f'Checking {len(programs)} recorded file-backed promises', flush=True)
with args.output.open('w') as output:
    with concurrent.futures.ThreadPoolExecutor(max_workers=args.workers) as pool:
        for row in pool.map(check, programs):
            output.write(json.dumps(row) + '\n')
            output.flush()
            if row['before']['exit'] and row['after']['exit'] == 0:
                print('ADMITTED', row['path'], flush=True)

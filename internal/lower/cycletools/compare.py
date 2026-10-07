"""Compare complete lowering decisions on pinned, unmodified input bytes."""
import argparse
import concurrent.futures
import hashlib
import json
import pathlib
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument('old')
parser.add_argument('new')
parser.add_argument('census', type=pathlib.Path)
parser.add_argument('output', type=pathlib.Path)
parser.add_argument('--jobs', type=int, default=4)
parser.add_argument('--manifest', type=pathlib.Path, help='explicit path-to-category JSON map')
parser.add_argument('--baseline', type=pathlib.Path, help='reuse old decisions with verified input hashes')
args = parser.parse_args()
repository = pathlib.Path(__file__).resolve().parents[3]
paths = {}
for category, directory in [
    ('oracle', repository / 'internal/oracle/testdata'),
    ('stage1', repository / 'stage1'),
    ('census-repro', repository / 'stage3/census'),
    ('load', repository / 'internal/load/testdata'),
    ('census', args.census / 'src/compiler'),
]:
    for path in directory.rglob('*'):
        if path.is_file() and path.suffix in ('.a', '.ts'):
            paths[str(path)] = category
if args.manifest:
    paths = json.loads(args.manifest.read_text())
if not args.manifest and not any(category == 'census' for category in paths.values()):
    raise RuntimeError('census source corpus is missing')
args.output.mkdir(parents=True, exist_ok=True)
baseline = {}
if args.baseline:
    baseline = {r['path']: r for r in map(json.loads, args.baseline.read_text().splitlines())}


def inspect(item):
    path, category = item
    record = dict(path=path, category=category,
                  sha256=hashlib.sha256(pathlib.Path(path).read_bytes()).hexdigest())
    for name, binary in [('old', args.old), ('new', args.new)]:
        if name == 'old' and args.baseline:
            earlier = baseline[path]
            if record['sha256'] != earlier['sha256']:
                raise RuntimeError(f'input changed: {path}')
            record[name] = earlier['old']
            continue
        result = subprocess.run([binary, path], capture_output=True, text=True,
                                cwd=repository, timeout=600)
        if result.returncode:
            raise RuntimeError(f'{name}: {path}: exit {result.returncode}: {result.stderr}')
        record[name] = json.loads(result.stdout)
    record['different'] = record['old']['decision'] != record['new']['decision']
    return record


records = []
with concurrent.futures.ThreadPoolExecutor(max_workers=args.jobs) as workers:
    for index, record in enumerate(workers.map(inspect, sorted(paths.items())), 1):
        records.append(record)
        print(f"{index}/{len(paths)} {record['category']} different={record['different']} {record['path']}", flush=True)
with (args.output / 'decisions.jsonl').open('w') as output:
    for record in records:
        output.write(json.dumps(record, sort_keys=True) + '\n')
summary = dict(inputs=len(records), differences=sum(r['different'] for r in records))
summary['categories'] = {c: sum(r['category'] == c for r in records)
                         for c in sorted(set(paths.values()))}
summary['old_decisions'] = {}
for record in records:
    kind = record['old']['decision'].split(':', 1)[0]
    summary['old_decisions'][kind] = summary['old_decisions'].get(kind, 0) + 1
(args.output / 'summary.json').write_text(json.dumps(summary, indent=2) + '\n')
print(json.dumps(summary), flush=True)
raise SystemExit(bool(summary['differences']))

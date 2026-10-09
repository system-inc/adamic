"""Resumable per-file walks, each loaded as the same whole compiler project.

Usage: stream.py BINARY COMPILER_ROOT OUTPUT [SECONDS=45] [RSS_MIB=6144] [MODE=speculative]
Each completed file is atomically published with a checksum. Raw file depths
are provisional until finalize_stream.py combines all failed boundaries.
"""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import time

binary, root, output = (Path(x).resolve() for x in sys.argv[1:4])
seconds = float(sys.argv[4]) if len(sys.argv) > 4 else 45
rss_mib = int(sys.argv[5]) if len(sys.argv) > 5 else 6144
mode = sys.argv[6] if len(sys.argv) > 6 else 'speculative'
assert mode in ('speculative', 'full', 'no-stubs')
output.mkdir(parents=True, exist_ok=True)
records = output / 'records'
records.mkdir(exist_ok=True)
inventory = [dict(file=str(p.relative_to(root)), bytes=p.stat().st_size,
                  sha256=hashlib.sha256(p.read_bytes()).hexdigest())
             for p in sorted(root.rglob('*')) if p.is_file() and p.suffix in ('.ts', '.a')]
identity = dict(binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(),
                root=str(root), files=inventory, seconds=seconds, rss_mib=rss_mib,
                project='whole compiler; only the census walk is filtered')
if mode != 'speculative': identity['mode'] = mode
manifest = output / 'INPUT.json'
if manifest.exists():
    assert json.loads(manifest.read_text()) == identity, 'resume input or limits changed'
else:
    manifest.write_text(json.dumps(identity, indent=2) + '\n')
# Measure the largest sources first, then keep the inventory ordering stable.
for item in sorted(inventory, key=lambda x: (-x['bytes'], x['file'])):
    selected = os.environ.get('LATENT_RUN_ONLY')
    if selected and item['file'] != selected:
        continue
    name = item['file'].replace('/', '__')
    record = records / (name + '.jsonl')
    checksum = records / (name + '.sha256')
    if record.exists():
        assert checksum.read_text().strip() == hashlib.sha256(record.read_bytes()).hexdigest()
        rows = [json.loads(line) for line in record.read_text().splitlines()]
        assert len(rows) == 2 and rows[1]['file'] == str(root / item['file'])
        print('resume skip ' + item['file'], flush=True)
        continue
    pending = records / (name + '.pending')
    metrics = output / (name + '.metrics.json')
    log = output / (name + '.log')
    env = dict(os.environ, LATENT_ONLY_FILE=str(root / item['file']),
               LATENT_SPECULATIVE='1', LATENT_FULL='1', LATENT_ASSERT_NO_OUTPUT='1',
               GOMEMLIMIT=os.environ.get('LATENT_FILE_MEMORY', '1GiB'), GOGC='200', GOMAXPROCS=os.environ.get('LATENT_FILE_CPUS', '4'))
    env.pop('LATENT_MUTANT_NO_STUBS', None)
    if mode == 'full': env.pop('LATENT_SPECULATIVE', None)
    if mode == 'no-stubs': env['LATENT_MUTANT_NO_STUBS'] = '1'
    deadline = float(os.environ.get('LATENT_FILE_DEADLINE', 'inf'))
    allowance = min(seconds, deadline - time.time() - 2)
    if allowance <= 0:
        print('unit deadline skip ' + item['file'], flush=True)
        continue
    print('start ' + item['file'], flush=True)
    process = subprocess.run([sys.executable, str(Path(__file__).with_name('timed_run.py')),
                              str(allowance), str(rss_mib), str(metrics), str(log),
                              str(binary), str(root), str(pending)], env=env,
                             timeout=seconds + 10)
    measurement = json.loads(metrics.read_text())
    measurement['unit_deadline_limited'] = allowance < seconds
    measurement['go_env'] = {key: env[key] for key in ('GOMEMLIMIT','GOGC','GOMAXPROCS')}
    metrics.write_text(json.dumps(measurement, indent=2) + '\n')
    if process.returncode == 0:
        rows = [json.loads(line) for line in pending.read_text().splitlines()]
        assert len(rows) == 2 and rows[1]['file'] == str(root / item['file'])
        if mode == 'speculative':
            assert rows[1]['speculative_coverage']['unvisited_nodes'] == 0
            assert rows[1]['speculative_coverage']['source_bytes'] == item['bytes']
        checksum.write_text(hashlib.sha256(pending.read_bytes()).hexdigest() + '\n')
        os.replace(pending, record)
        print('complete ' + item['file'], flush=True)
    else:
        print('incomplete ' + item['file'] + ' exit=' + str(process.returncode), flush=True)
print('walk finished: ' + str(len(list(records.glob('*.jsonl')))) + '/' + str(len(inventory)), flush=True)

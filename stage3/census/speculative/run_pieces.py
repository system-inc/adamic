"""Bounded, resumable piece measurements: BINARY ROOT PLAN OUTPUT SECONDS RSS_MIB WORKERS.

Each child loads the whole project. timed_run.py enforces resource limits. Only
validated, completed records are atomically persisted; failed attempts are excluded.
"""
from concurrent.futures import ThreadPoolExecutor
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import time

binary, root, plan_path, output = (Path(p).resolve() for p in sys.argv[1:5])
seconds, rss_mib, workers = map(int, sys.argv[5:8])
assert seconds > 0 and rss_mib > 0 and 1 <= workers <= 64
output.mkdir(parents=True, exist_ok=True)
plan = json.loads(plan_path.read_text())
assert Path(plan['file']).is_relative_to(root)
assert hashlib.sha256(Path(plan['file']).read_bytes()).hexdigest() == plan['sha256']
identity = dict(binary_sha256=hashlib.sha256(binary.read_bytes()).hexdigest(),
                plan_sha256=hashlib.sha256(plan_path.read_bytes()).hexdigest(), seconds=seconds, rss_mib=rss_mib,
                sources=[dict(file=str(p.relative_to(root)), sha256=hashlib.sha256(p.read_bytes()).hexdigest())
                         for p in sorted(root.rglob('*')) if p.is_file()])
manifest = output / 'INPUT.json'
if manifest.exists():
    assert json.loads(manifest.read_text()) == identity, 'piece run input changed'
else:
    manifest.write_text(json.dumps(identity, indent=2) + '\n')
records = output / 'records'
records.mkdir(exist_ok=True)
scripts = Path(__file__).resolve().parent

def validate(piece, path):
    rows = [json.loads(line) for line in path.read_text().splitlines()]
    assert len(rows) == 2 and rows[1]['file'] == plan['file']
    record = rows[1]
    assert record['piece'] == dict(id=piece['id'], plan_sha256=identity['plan_sha256'], atoms=piece['atoms'])
    assert record['speculative_coverage']['unvisited_nodes'] == 0, 'piece AST traversal incomplete'
    allowed = {'prepass', *piece['atoms']}
    assert all(finding['piece_atom'] in allowed for finding in record['findings']), 'site moved between pieces'
    return rows

def run(piece):
    name = f"piece-{piece['id']:03d}"
    record = records / (name + '.jsonl')
    checksum = records / (name + '.sha256')
    metrics = output / (name + '.metrics.json')
    if record.exists():
        assert checksum.read_text().strip() == hashlib.sha256(record.read_bytes()).hexdigest(), 'piece checksum changed'
        validate(piece, record)
        print(f"RESUME {name}", flush=True)
        return
    assert not metrics.exists() or json.loads(metrics.read_text())['exit'] != 0, 'completed piece dropped'
    pending = output / (name + '.pending.jsonl')
    env = dict(os.environ, LATENT_SPECULATIVE='1', LATENT_FULL='1', LATENT_ASSERT_NO_OUTPUT='1',
               LATENT_ONLY_FILE=plan['file'], LATENT_PIECE_PLAN=str(plan_path), LATENT_PIECE_ID=str(piece['id']),
               GOMAXPROCS='1', GOMEMLIMIT='1GiB', GOGC='200',
               LATENT_PROGRESS_FILE=str(output / (name + '.progress.json')))
    Path(env["LATENT_PROGRESS_FILE"]).unlink(missing_ok=True)
    print(f"START {name} bytes={piece['bytes']} epoch={time.time():.3f}", flush=True)
    with (output / (name + '.supervisor.log')).open('w') as log:
        result = subprocess.run([sys.executable, str(scripts / 'timed_run.py'), str(seconds), str(rss_mib),
                                 str(metrics), str(output / (name + '.log')), str(binary), str(root), str(pending)],
                                env=env, stdout=log, stderr=subprocess.STDOUT, timeout=seconds+15)
    if result.returncode == 0:
        validate(piece, pending)
        pending.replace(record)
        checksum.write_text(hashlib.sha256(record.read_bytes()).hexdigest() + '\n')
    print(f"DONE {name} {metrics.read_text().strip()}", flush=True)

selection = os.environ.get('LATENT_PIECES')
pieces = plan['pieces'] if not selection else [p for p in plan['pieces'] if p['id'] in set(map(int, selection.split(',')))]
assert not selection or len(pieces) == len(set(map(int, selection.split(',')))), 'unknown requested pieces'
with ThreadPoolExecutor(max_workers=workers) as pool:
    list(pool.map(run, sorted(pieces, key=lambda p: (-p['bytes'], p['id']))))

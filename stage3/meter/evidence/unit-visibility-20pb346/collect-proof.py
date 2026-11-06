"""Collect exact full-run identity, prefix identities, growth and profile evidence.

Usage: python3 collect-proof.py RUN_DIRECTORY
Collect completed outputs when available; preserve incomplete proof attempts explicitly.
"""
from pathlib import Path
import gzip
import hashlib
import json
import shutil
import sys

source = Path(sys.argv[1]).resolve()
output = Path(__file__).resolve().parent

def save_gzip(src, dest):
    with src.open('rb') as reader, dest.open('wb') as raw:
        with gzip.GzipFile(fileobj=raw, mode='wb', mtime=0) as writer:
            shutil.copyfileobj(reader, writer)

full = {}
for side in ('before', 'after'):
    folder = source / (side + '-compiler-complete')
    if not (folder / 'run.json').exists():
        full[side] = {'completed': False, 'reason': 'controlled retry stopped, or after side not started'}
        if folder.exists():
            save_gzip(folder / 'latent-times.jsonl', output / (side + '-compiler-incomplete-times.jsonl.gz'))
            for name in ('supervision.json', 'stopped-proof.json'):
                if (folder / name).exists():
                    shutil.copyfile(folder / name, output / (side + '-' + name))
        continue
    run = json.loads((folder / 'run.json').read_text())
    assert run['exit_code'] == 0, run
    data = (folder / 'latent.jsonl').read_bytes()
    rows = [json.loads(line) for line in data.splitlines()]
    identities = [(u['where'], u['kind']) for row in rows[1:] for u in row['units']]
    assert len(identities) == len(set(identities))
    full[side] = dict(run=run, bytes=len(data), files=len(rows)-1,
                      units=len(identities), sha256=hashlib.sha256(data).hexdigest())
    save_gzip(folder / 'latent.jsonl', output / (side + '-compiler-latent.jsonl.gz'))
    save_gzip(folder / 'latent-times.jsonl', output / (side + '-compiler-times.jsonl.gz'))
    shutil.copyfile(folder / 'run.json', output / (side + '-compiler-complete.json'))
    for name in ('watchdog-extension.json', 'run-original-watchdog.json', 'supervision.json'):
        if (folder / name).exists():
            shutil.copyfile(folder / name, output / (side + '-' + name))
if all('sha256' in full[side] for side in ('before', 'after')):
    assert full['before']['sha256'] == full['after']['sha256'], 'completed latent streams differ'

capped = {}
for mode in ('entry', 'compiler'):
    names = {}
    raw = {}
    for side in ('before', 'after'):
        label = side + '-' + mode + '-cap'
        folder = source / label
        times = [json.loads(line) for line in (folder / 'latent-times.jsonl').read_text().splitlines()]
        names[side] = [(r['file'], r['where']) for r in times]
        raw[side] = (folder / 'latent.jsonl').read_bytes()
        shutil.copyfile(folder / 'run.json', output / (label + '.json'))
        save_gzip(folder / 'latent-times.jsonl', output / (label + '-times.jsonl.gz'))
        if mode == 'entry':
            shutil.copyfile(folder / 'growth.jsonl', output / (side + '-entry-growth.jsonl'))
            shutil.copyfile(folder / 'cpu.pprof', output / (side + '-entry-initial-cpu.pprof'))
    count = min(len(names['before']), len(names['after']))
    assert names['before'][:count] == names['after'][:count], mode + ' attempted-unit prefix differs'
    assert raw['before'] == raw['after'], mode + ' capped latent output differs'
    capped[mode] = dict(before_units=len(names['before']), after_units=len(names['after']),
                        common_prefix_units=count, capped_latent_bytes=len(raw['before']),
                        capped_sha256=hashlib.sha256(raw['before']).hexdigest())

for name in ('environment.json', 'build.py', 'run.py', 'complete.sh', 'extend-completion-guard.py', 'resume-completion.py', 'exit-status-probe.json', 'run-completion-limited.py'):
    shutil.copyfile(source / name, output / name)
for side in ('before', 'after'):
    quiet = source / (side + '-entry-profile-quiet')
    shutil.copyfile(quiet / 'cpu.pprof', output / (side + '-entry-cpu.pprof'))
    shutil.copyfile(quiet / 'run.json', output / (side + '-entry-profile-quiet.json'))
    shutil.copyfile(quiet / 'profile-top.txt', output / (side + '-profile-top.txt'))
    failed = source / (side + '-compiler-default-' + ('oom' if side == 'before' else 'interrupted'))
    shutil.copyfile(failed / 'run.json', output / (side + '-compiler-default-attempt.json'))
    save_gzip(failed / 'latent-times.jsonl', output / (side + '-compiler-default-times.jsonl.gz'))
for name in ('default-memory.events.txt', 'default-memory.max.txt', 'default-memory.peak.txt'):
    shutil.copyfile(source / name, output / name)
shutil.copyfile(source / 'completed-fixture/proof.json', output / 'completed-fixture-proof.json')
(output / 'identity-proof.json').write_text(json.dumps(dict(full=full, capped=capped), indent=2)+'\n')

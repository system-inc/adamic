"""Save a successfully completed head stream before any pairing work."""
import gzip
import hashlib
import json
from pathlib import Path
import subprocess
import sys

folder = Path(__file__).resolve().parent
root = folder.parents[4]
status = json.loads((folder / 'first-error-head.status.json').read_text())
assert status['child_exit'] == 0, status
stream = folder / 'head.jsonl'
last = None
with stream.open() as source:
    header = json.loads(source.readline())
    for line in source:
        last = json.loads(line)
with gzip.open(folder / 'base.jsonl.gz', 'rt') as source:
    assert header == json.loads(source.readline()), 'head and baseline headers differ'
assert last and last.get('status') == 'relation_report', 'head stream has no completion trailer'
provenance = json.loads((folder / 'provenance.json').read_text())
provenance['binaries']['resumed_head'] = hashlib.sha256(Path('/tmp/zb0q0zd-resume-head').read_bytes()).hexdigest()
provenance['streams']['head.jsonl'] = {'sha256_uncompressed': hashlib.sha256(stream.read_bytes()).hexdigest(), 'bytes_uncompressed': stream.stat().st_size}
provenance['final_head_exit'] = 0
provenance['paired_measurement'] = 'head complete; checkpoint saved before pairing'
(folder / 'provenance.json').write_text(json.dumps(provenance, indent=2) + '\n')
validation = json.loads((folder / 'validation.json').read_text())
validation['final_head_census'] = 'completed with child exit 0; pairing pending'
validation['resumed_head_exit'] = 0
(folder / 'validation.json').write_text(json.dumps(validation, indent=2) + '\n')
old_status = folder / 'STATUS.md'
previous_status = folder / 'previous-stop.md'
if old_status.exists() and not previous_status.exists():
    previous_status.write_bytes(old_status.read_bytes())
old_status.write_text('# Head complete, pairing pending\n\nThe resumed head exited successfully. Its complete stream and trailer are saved\nand checkpointed before pairing. The verified 018b7c21 baseline is reused.\nRun finish_pair.py next, inspect all movements and both relation trailers, then\nsquash code and evidence into one commit on 018b7c21 and push normally to\ntypescript/census-deferred. The previous incomplete-turn notes are retained in\nprevious-stop.md. No paired count has yet been claimed.\n')
paths = [folder / name for name in ['STATUS.md', 'previous-stop.md', 'validation.json', 'provenance.json', 'first-error-head.status.json', 'first-error-head.log.txt', 'resume-verification.json', 'resume-overlay.log.txt', 'resume-entry.log.txt', 'resume-build.log.txt', 'checkpoint_head.py']]
for name in ['head.jsonl', 'head-times.jsonl']:
    path = folder / name
    if path.stat().st_size > 1000000:
        archive = path.with_name(path.name + '.gz')
        archive.write_bytes(gzip.compress(path.read_bytes(), compresslevel=9, mtime=0))
        path.unlink()
        path = archive
    paths.append(path)
message = Path('/tmp/zb0q0zd-head-complete-message.txt')
message.write_text('never cut alone: lands with chg_pavttxhpyc992z320pnwrjtj2m, because both edit the latent driver templates\n\nCheckpoint the successful complete head census before pairing. The baseline and adapted source hashes were verified for reuse. Head child exit is zero and the complete relation trailer is retained. Paired overstatement counts and the final one-commit publication are pending.\n')
for command in [['git', 'add', *map(str, paths)], ['git', 'commit', '-F', str(message)], ['git', 'push', 'origin', 'HEAD:refs/heads/typescript/census-deferred-checkpoint']]:
    result = subprocess.run(command, cwd=root)
    print('checkpoint child exit:', result.returncode, flush=True)
    if result.returncode:
        raise SystemExit(result.returncode)

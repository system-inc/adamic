#!/usr/bin/env python3
"""Record exact pinned sources, input bytes, executable hashes and paired agreement."""
import gzip
import hashlib
import json
import os
from pathlib import Path
import subprocess

run = Path(__file__).resolve().parents[1]
scratch = Path('/tmp/adamic-gate/stage3-meter.RXmuND')
pin = '295df3bc6aa3d47933ccf675da1c33e745856313'
report = json.loads((run / 'report.json').read_text())
def sha(path):
    with path.open('rb') as source:
        return hashlib.file_digest(source, 'sha256').hexdigest()
def command(args, cwd):
    return subprocess.check_output(args, cwd=cwd, text=True).strip()
def stream(path):
    with gzip.open(path, 'rt') as source:
        return [json.loads(line) for line in source]
def normalize(value, tree):
    if isinstance(value, str):
        return value.replace(str(tree), '<adapted>')
    if isinstance(value, list):
        return [normalize(item, tree) for item in value]
    if isinstance(value, dict):
        return {key: normalize(item, tree) for key, item in value.items() if key not in ('seconds',)}
    return value
result = dict(measured_commit=pin, nproc=command(['nproc'], scratch),
              cpu_max=Path('/sys/fs/cgroup/cpu.max').read_text().strip(), trees={}, paired_agreement={})
for label in ['main', 'area']:
    source = scratch / (label + '-source')
    tree = scratch / (label + '-adapted')
    binaries = scratch / (label + '-binaries')
    assert command(['git', 'rev-parse', 'HEAD'], source) == pin
    # Untracked node_modules is an input link, not a tracked source change.
    subprocess.run(['git', 'diff', '--exit-code', 'HEAD'], cwd=source, check=True, stdout=subprocess.DEVNULL)
    manifest = {str(path.relative_to(tree)): sha(path) for path in sorted((tree / 'src/compiler').rglob('*')) if path.is_file()}
    assert len([name for name in manifest if name.endswith(('.a', '.ts'))]) == 79
    executables = {}
    for name in ['census', 'entry-census', 'latent-census']:
        path = binaries / name
        executables[name] = sha(path)
        with (run / 'validation' / (label + '-' + name + '-buildinfo.log')).open('w') as log:
            subprocess.run(['go', 'version', '-m', str(path)], stdout=log, stderr=log, check=True)
    result['trees'][label] = dict(compiler_commit=pin, tracked_compiler_sources_clean=True,
                                  cohere_commit=command(['git', 'rev-parse', 'HEAD'], source/'cohere'),
                                  checker_commit=command(['git', 'rev-parse', 'HEAD'], source/'cohere/TypeScript'),
                                  goflags=command(['go', 'env', 'GOFLAGS'], source),
                                  gowork=command(['go', 'env', 'GOWORK'], source),
                                  adapted_compiler_files=manifest, executable_sha256=executables,
                                  adapted_entry_sha256=sha(tree/'src/tsc/tsc.ts'))
for name in ['census.jsonl.gz', 'tsc/census.jsonl.gz']:
    left = normalize(stream(run/'main'/name), scratch/'main-adapted')
    right = normalize(stream(run/'area'/name), scratch/'area-adapted')
    assert left == right, name
    result['paired_agreement'][name] = 'identical after replacing tree prefix and excluding seconds'
for name in ['latent.jsonl.gz', 'tsc/latent.jsonl.gz']:
    left = normalize(stream(run/'main'/name), scratch/'main-adapted')
    right = normalize(stream(run/'area'/name), scratch/'area-adapted')
    result['paired_agreement'][name] = ('identical after replacing tree prefix and excluding seconds' if left == right else 'event records differ; aggregate counts compared separately')
assert result['trees']['main']['adapted_compiler_files'] == result['trees']['area']['adapted_compiler_files']
assert result['trees']['main']['adapted_entry_sha256'] == result['trees']['area']['adapted_entry_sha256']
assert report['trees']['main']['totals'] == report['trees']['area']['totals']
main_reasons = {key.replace(str(scratch/'main-adapted'), '<adapted>'): value for key, value in report['trees']['main']['latent_lowering']['per_reason'].items()}
area_reasons = {key.replace(str(scratch/'area-adapted'), '<adapted>'): value for key, value in report['trees']['area']['latent_lowering']['per_reason'].items()}
assert main_reasons == area_reasons
result['upstream_source_commit'] = command(['git', 'rev-parse', 'HEAD'], Path('/tmp/badrnrf-inputs/TypeScript'))
assert result['upstream_source_commit'] == '050880ce59e30b356b686bd3144efe24f875ebc8'
result['paired_agreement']['adapted_source_bytes'] = 'all compiler inputs and tsc entry identical'
result['paired_agreement']['aggregate_checker_lowering_and_latent_reasons'] = 'identical after normalizing embedded adapted-root prefixes in reason keys'
result['retained_stream_sha256'] = {str(path.relative_to(run)): sha(path) for path in sorted(run.rglob('*.jsonl.gz'))}
(run/'validation/provenance.json').write_text(json.dumps(result, indent=2)+'\n')
print(json.dumps(result['paired_agreement'], indent=2))
print('ordinary binaries identical:', result['trees']['main']['executable_sha256']['census'] == result['trees']['area']['executable_sha256']['census'])
print('tracked compiler sources unchanged at', pin)

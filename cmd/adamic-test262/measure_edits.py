#!/usr/bin/env python3
"""Interleave library edits and cold runs; preserve raw output only in scratch space."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import shlex
import subprocess
import sys
sys.path.insert(0, str(__import__("pathlib").Path(__file__).resolve().parents[2] / "internal" / "boundedrun"))
from python import run as bounded_run
import time

parser = argparse.ArgumentParser()
parser.add_argument('--before-root', required=True)
parser.add_argument('--after-root', required=True)
parser.add_argument('--test262', required=True)
parser.add_argument('--output', required=True)
args = parser.parse_args()
output = Path(args.output).resolve()
output.mkdir(parents=True, exist_ok=True)
roots = {name: Path(path).resolve() for name, path in [('before', args.before_root), ('after', args.after_root)]}
loops = [('Math', 'built-ins/Math', 'math.c', 'library_math_number.go'),
         ('padStart', 'built-ins/String/prototype/padStart', 'string.c', 'library_string.go'),
         ('sort', 'built-ins/Array/prototype/sort', 'array.c', 'library_array.go')]
records = []
references = {}

def capture(command, cwd=None):
    return bounded_run(command, cwd=cwd, stdout=subprocess.PIPE, text=True, check=True).stdout.strip()

identity = {'nproc': capture(['nproc']), 'cpu.max': Path('/sys/fs/cgroup/cpu.max').read_text().strip(),
            'go': capture(['go', 'version']), 'clang': capture(['clang', '--version']).splitlines()[0],
            'node': capture(['node', '--version']), 'GOFLAGS': os.environ.get('GOFLAGS', ''),
            'test262': capture(['git', '-C', args.test262, 'rev-parse', 'HEAD']),
            'native_flags': '-std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable '
                            '-Wno-unused-but-set-variable -Wno-unused-function -Wno-unused-parameter '
                            '-Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls '
                            '-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all'}
commits = {name: capture(['git', 'rev-parse', 'HEAD'], root) for name, root in roots.items()}
go_caches = {variant: str(output / ('go-build-' + variant)) for variant in roots}

def environment(cache):
    variant = cache.name.rsplit('-', 1)[1]
    return dict(os.environ, XDG_CACHE_HOME=str(cache), GOCACHE=go_caches[variant], GOMAXPROCS='4', ADAMIC_GATE_UNCACHED='0')

def instrument(command, env):
    return ' '.join(k + '=' + shlex.quote(env[k]) for k in ['XDG_CACHE_HOME', 'GOCACHE', 'GOMAXPROCS', 'ADAMIC_GATE_UNCACHED']) + ' ' + shlex.join(command)

def build(variant, env, label):
    binary = output / ('runner-' + variant)
    command = ['go', 'build', '-o', str(binary), './cmd/adamic-test262']
    with (output / (label + '-build.log')).open('w') as log:
        bounded_run(command, cwd=roots[variant], env=env, stdout=log, stderr=log, check=True)
    return binary, instrument(command, env)

def summarize(profile):
    tests = profile['tests'] or []
    stages = {}
    per_test = []
    for test in tests:
        phases = {}
        for phase in test['phases']:
            name = phase['stage']
            aggregate = stages.setdefault(name, {'calls': 0, 'hits': 0, 'seconds': 0})
            aggregate['calls'] += 1
            aggregate['hits'] += int(phase.get('hit', False))
            aggregate['seconds'] += phase['end'] - phase['start']
            phases[name] = {'seconds': phase['end'] - phase['start'], 'hit': phase.get('hit', False)}
        per_test.append({'path': test['path'], 'worker': test['worker'], 'start': test['start'],
                         'end': test['end'], 'phases': phases})
    workers = []
    for worker in sorted(set(t['worker'] for t in tests)):
        tasks = [t for t in tests if t['worker'] == worker]
        workers.append({'worker': worker, 'tests': len(tasks), 'busy_wall_seconds': sum(t['end'] - t['start'] for t in tasks),
                        'first': min(t['start'] for t in tasks), 'last': max(t['end'] for t in tasks)})
    events = sorted([(t['start'], 1) for t in tests] + [(t['end'], -1) for t in tests])
    active = peak = 0
    for _, change in events:
        active += change
        peak = max(peak, active)
    return {'stages': stages, 'workers': workers, 'peak_overlapping_workers': peak,
            'prepare': profile['prepare'], 'per_test': per_test}

def run(variant, label, directory, kind, round_number, env, binary, build_command, edit=None, rebuild_seconds=0, load_before=None):
    profile_path = output / (label + '-profile.json')
    command = [str(binary), '-jobs', '4', '-adapt', '-json', '-test262', args.test262,
               '-profile', str(profile_path), directory]
    load_before = load_before or Path('/proc/loadavg').read_text().strip()
    start = time.monotonic()
    with (output / (label + '.json')).open('w') as stdout, (output / (label + '.log')).open('w') as stderr:
        result = bounded_run(command, cwd=roots[variant], env=env, stdout=stdout, stderr=stderr)
    elapsed = time.monotonic() - start
    if result.returncode:
        raise RuntimeError(label + ' failed; see scratch log')
    report = (output / (label + '.json')).read_bytes()
    log = (output / (label + '.log')).read_bytes()
    loop = directory
    if loop in references and references[loop] != (report, log):
        raise AssertionError(label + ' verdict JSON or ordered progress/table changed')
    references[loop] = (report, log)
    summary = summarize(json.loads(profile_path.read_text()))
    record = dict(identity, commit=commits[variant], root=str(roots[variant]), variant=variant, loop=directory,
                  case=kind, round=round_number, cache=('cold' if kind == 'cold' else 'primed before edit'),
                  rebuild_seconds=rebuild_seconds, runner_seconds=elapsed, seconds=elapsed + rebuild_seconds,
                  build_command=build_command, instrument=instrument(command, env), exit=result.returncode,
                  load_before=load_before, load_after=Path('/proc/loadavg').read_text().strip(), edit=edit,
                  report_sha256=hashlib.sha256(report).hexdigest(), table_progress_sha256=hashlib.sha256(log).hexdigest(),
                  stages=summary['stages'], workers=summary['workers'], peak_overlapping_workers=summary['peak_overlapping_workers'],
                  prepare=summary['prepare'])
    if kind != 'prime':
        native = record['stages']['native-observation']
        node = record['stages']['node']
        if kind == 'cold' or variant == 'after':
            assert native['hits'] == 0, (label, native)
        if kind != 'cold' and variant == 'after':
            assert node['hits'] == node['calls'], (label, node)
        records.append(record)
        # Keep per-test detail only for the fastest cold sample of each directory/variant.
        if kind == 'cold':
            key = directory + ':' + variant
            existing = compact['cold_profiles'].get(key)
            if not existing or elapsed < existing['runner_seconds']:
                compact['cold_profiles'][key] = dict(runner_seconds=elapsed, record_label=label, **summary)
    print(label, f'{record["seconds"]:.3f}s', 'native', record['stages'].get('native-observation'), 'Node', record['stages'].get('node'), flush=True)
    compact['runs'] = records
    (output / 'measurements.json').write_text(json.dumps(compact, indent=2) + '\n')

compact = {'method': 'Same box, paired before/after for three rounds; edits include rebuilding runner then invoking it. Prime each loop, mutate one runtime comment byte or one lowering whitespace byte, restore after every run. Cold runs use fresh result/runtime cache roots; Private Go build caches for each runner warmed by the untimed original-source primes, preventing artifacts from earlier experiments from biasing rebuild costs.',
           'identity': identity, 'commits': commits, 'cold_profiles': {}, 'runs': []}

# Original-program caches are shared between edit kinds; every changed byte is unique and
# changes the runner/compiler identity. Thus native observations must miss on every edit.
for name, directory, runtime_file, lowering_file in loops:
    primed = set()
    for kind, relative in [('runtime-edit', 'internal/native/runtime/' + runtime_file),
                           ('lowering-edit', 'internal/lower/' + lowering_file)]:
        originals = {variant: (root / relative).read_bytes() for variant, root in roots.items()}
        assert originals['before'] == originals['after']
        original = originals['before']
        comment = original.index(b'//') + 2
        while not (65 <= original[comment] <= 90 or 97 <= original[comment] <= 122):
            comment += 1
        edit_hashes = set()
        for round_number in range(3):
            for variant in ['before', 'after']:
                root = roots[variant]
                cache = output / ('cache-edit-' + name + '-' + variant)
                env = environment(cache)
                label = f'{name}-{kind}-{round_number}-{variant}'
                if variant not in primed:
                    binary, command = build(variant, env, label + '-prime')
                    run(variant, label + '-prime', directory, 'prime', -1, env, binary, command)
                    primed.add(variant)
                offset = comment
                replacement = [value for value in range(75, 91) if value != original[comment]][round_number]
                if kind == 'lowering-edit':
                    # Removing a blank-line byte preserves Go semantics but moves different
                    # source regions in each round, including compiler debug line tables.
                    blanks = [index + 1 for index in range(len(original) - 1) if original[index:index + 2] == b'\n\n']
                    for number in range(round_number + 1):
                        offset = min(blanks, key=lambda index: abs(index - len(original) * (number + 1) / 5))
                        blanks.remove(offset)
                    replacement = 32
                changed = original[:offset] + bytes([replacement]) + original[offset + 1:]
                assert sum(a != b for a, b in zip(original, changed)) == 1
                if variant == 'before':
                    digest = hashlib.sha256(changed).hexdigest()
                    assert digest not in edit_hashes, 'repeated edit input'
                    edit_hashes.add(digest)
                path = root / relative
                try:
                    path.write_bytes(changed)
                    load_before = Path('/proc/loadavg').read_text().strip()
                    start = time.monotonic()
                    binary, build_command = build(variant, env, label)
                    rebuild_seconds = time.monotonic() - start
                    edit = {'path': relative, 'offset': offset, 'before_byte': original[offset], 'after_byte': changed[offset],
                            'before_sha256': hashlib.sha256(original).hexdigest(), 'after_sha256': hashlib.sha256(changed).hexdigest()}
                    run(variant, label, directory, kind, round_number, env, binary, build_command, edit, rebuild_seconds, load_before)
                finally:
                    path.write_bytes(original)
    for round_number in range(3):
        for variant in ['before', 'after']:
            label = f'{name}-cold-{round_number}-{variant}'
            env = environment(output / ('cache-' + label))
            binary, build_command = build(variant, env, label)
            run(variant, label, directory, 'cold', round_number, env, binary, build_command)
print('all after-edit native observations missed; all after-edit Node observations hit; all verdicts, tables and progress byte-identical', flush=True)

#!/usr/bin/env bash
# Usage: bash stage3/census/speculative/server-run.sh WORKDIR DEADLINE_EPOCH [WORKERS=14]
set -euo pipefail
repo=$(cd "$(dirname "$0")/../../.." && pwd)
exec python3 - "$repo" "$@" <<'PY'
import fcntl
import gzip
import hashlib
import json
import math
import os
from pathlib import Path
import subprocess
import shutil
import sys
import tarfile
import threading
import time

REPO = Path(sys.argv[1])
SCRIPTS = REPO / 'stage3/census/speculative'
ARCHIVE = SCRIPTS / 'evidence/continuation'
BASE = ARCHIVE / 'retry-priority'
ADAPTER = 'f6bb0b41fb00b69a0a33db5551290b62a06e3cc7'
CONTROL = '_namespaces/ts.moduleSpecifiers.ts'


def digest(data):
    return hashlib.sha256(data).hexdigest()


def atomic(path, data):
    pending = path.with_name(path.name + '.pending')
    pending.write_text(data)
    os.replace(pending, path)


def save(path, value):
    atomic(path, json.dumps(value, indent=2) + '\n')


def read(path):
    return json.loads(path.read_text())


def source_checks(root, identity):
    # The original retry_stream source checks are retained without relaxation.
    for item in identity['files']:
        source = root / item['file']
        assert source.stat().st_size == item['bytes'] and digest(source.read_bytes()) == item['sha256'], item['file']
    inventory = {str(p.relative_to(root)) for p in root.rglob('*') if p.is_file() and p.suffix in ('.ts', '.a')}
    assert inventory == {item['file'] for item in identity['files']}, 'source inventory changed'


def archived_record(name, root):
    path = BASE / 'records' / (name.replace('/', '__') + '.jsonl.gz')
    checks = read(ARCHIVE / 'SHA256.json')
    assert digest(path.read_bytes()) == checks[str(path.relative_to(ARCHIVE))], 'archive checksum'
    raw = gzip.decompress(path.read_bytes())
    checksum = path.with_suffix('').with_suffix('.sha256')
    assert digest(checksum.read_bytes()) == checks[str(checksum.relative_to(ARCHIVE))], 'checksum archive'
    assert digest(raw) == checksum.read_text().strip(), 'completed record checksum differs'
    old = read(BASE / 'INPUT.json')
    # Literal UTF-8 replacement preserves every other byte, including JSON order.
    return raw, raw.replace(old['root'].encode(), str(root).encode())


def command(args, log, seconds, env=None, check=True):
    # GNU timeout kills the complete inherited process group, including compilers.
    with log.open('wb') as output:
        result = subprocess.run(['timeout', '-k', '10s', str(max(1, seconds)) + 's', *map(str, args)],
                                cwd=REPO, env=env, stdout=output, stderr=subprocess.STDOUT,
                                timeout=max(1, seconds) + 15)
    if check and result.returncode:
        raise RuntimeError(f'{args[0]} exited {result.returncode}; see {log}')
    return result.returncode


def prove(binary, root, out, seconds):
    out.mkdir(parents=True, exist_ok=True)
    original = read(BASE / 'INPUT.json')
    source_checks(root, original)
    raw, expected = archived_record(CONTROL, root)
    (out / 'expected.jsonl').write_bytes(expected)
    # A new proof must never consume the previous invocation's output.
    (out / 'actual.jsonl').unlink(missing_ok=True)
    env = dict(os.environ, LATENT_ONLY_FILE=str(root / CONTROL), LATENT_SPECULATIVE='1',
               LATENT_FULL='1', LATENT_ASSERT_NO_OUTPUT='1', GOMAXPROCS='1', GOMEMLIMIT='3GiB')
    for key in ('LATENT_MUTANT_NO_STUBS', 'LATENT_PROGRESS_FILE'):
        env.pop(key, None)
    code = command([sys.executable, SCRIPTS / 'timed_run.py', seconds, 3072,
                    out / 'metrics.json', out / 'compiler.log', binary, root, out / 'actual.jsonl'],
                   out / 'supervisor.log', seconds + 10, env, check=False)
    actual = out / 'actual.jsonl'
    passed = code == 0 and actual.exists() and actual.read_bytes() == expected
    proof = dict(file=CONTROL, original_binary_sha256=original['binary_sha256'],
                 rebuild_binary_sha256=digest(binary.read_bytes()), original_record_sha256=digest(raw),
                 relocated_record_sha256=digest(expected), actual_record_sha256=digest(actual.read_bytes()) if actual.exists() else None,
                 original_root=original['root'], relocated_root=str(root), exit=code, byte_identical=passed)
    save(out / 'PROOF.json', proof)
    return proof


if len(sys.argv) > 2 and sys.argv[2] == '--prove-producer':
    if len(sys.argv) != 7:
        raise SystemExit('usage: server-run.sh --prove-producer BINARY COMPILER_ROOT OUTPUT SECONDS')
    proof = prove(Path(sys.argv[3]).resolve(), Path(sys.argv[4]).resolve(), Path(sys.argv[5]).resolve(), float(sys.argv[6]))
    print(json.dumps(proof))
    raise SystemExit(0 if proof['byte_identical'] else 2)

if len(sys.argv) not in (4, 5):
    raise SystemExit('usage: server-run.sh WORKDIR DEADLINE_EPOCH [WORKERS=14]')
WORK = Path(sys.argv[2]).resolve()
DEADLINE = float(sys.argv[3])
WORKERS = int(sys.argv[4]) if len(sys.argv) == 5 else 14
assert math.isfinite(DEADLINE) and WORKERS > 0
assert not any(c in str(WORK) for c in ('\\', '"', '\n')), 'work directory must be safe for literal JSON path relocation'
WORK.mkdir(parents=True, exist_ok=True)
lock = (WORK / 'server-run.lock').open('w')
fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
os.sched_setaffinity(0, sorted(os.sched_getaffinity(0))[:WORKERS])
os.environ.update(GOMEMLIMIT='3GiB', LATENT_FILE_MEMORY='3GiB', GOMAXPROCS=str(WORKERS))
os.environ.setdefault('GOPROXY', 'https://proxy.golang.org|direct')
for key in ('LATENT_RETRY_FILES', 'LATENT_RUN_ONLY', 'LATENT_ONLY_FILE', 'LATENT_MUTANT_NO_STUBS'):
    os.environ.pop(key, None)
ROOT = (WORK / 'tree/src/compiler').resolve()
RUN = WORK / 'run'
RUN.mkdir(exist_ok=True)
(RUN / 'records').mkdir(exist_ok=True)
state = {'stage': 'starting', 'reuse': 'not yet proved'}
stop = threading.Event()
progress_lock = threading.Lock()
watcher = None


def progress():
    lines = [time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime()),
             f"stage: {state['stage']}", f"baseline reuse: {state['reuse']}",
             f"files done: {len(list((RUN / 'records').glob('*.jsonl'))) if state['reuse'] != 'not yet proved' else 0}/{len(read(BASE / 'INPUT.json')['files'])}",
             f'workers requested: {WORKERS}; deadline epoch: {DEADLINE}']
    watch = WORK / 'watch.jsonl'
    if watch.exists():
        rows = watch.read_text().splitlines()
        if rows:
            try:
                snapshot = json.loads(rows[-1])
                lines.append('watch snapshot: ' + snapshot['utc'])
                for worker in snapshot['workers']:
                    output = Path(worker['output'])
                    if output.parent == RUN / 'records' and Path('/proc', str(worker['pid'])).exists():
                        name = output.name.removesuffix('.pending').removesuffix('.jsonl')
                        lines.append(f"worker pid {worker['pid']}: {name.replace('__', '/')}; RSS {worker['rss_kib']} KiB")
            except (ValueError, KeyError):
                pass
    for checkpoint in sorted(RUN.glob('*.progress.json')):
        try:
            value = read(checkpoint)
            lines.append('last checkpoint: ' + json.dumps(value, sort_keys=True))
        except (OSError, ValueError):
            pass
    with progress_lock:
        atomic(WORK / 'progress.txt', '\n'.join(lines) + '\n')


def monitor():
    while not stop.wait(60):
        progress()


def stage(name):
    state['stage'] = name
    progress()


def empty_result():
    original = read(BASE / 'INPUT.json')
    result = dict(status='nothing completed', files=[], examined_bytes=0,
                  counts={'NotYet': [0] * 5, 'Refused': [0] * 5}, top20=[],
                  deadline_epoch=DEADLINE, baseline_reuse=state['reuse'],
                  unexamined=original['files'])
    save(WORK / 'RESULT.json', result)
    atomic(WORK / 'README.md', 'Nothing completed. No incomplete observations count as coverage.\n')


def metrics_table():
    lines = ['file\twall_seconds\tpeak_rss_kib\texit_status']
    for item in read(BASE / 'INPUT.json')['files']:
        path = RUN / (item['file'].replace('/', '__') + '.metrics.json')
        value = read(path) if path.exists() else {}
        lines.append(f"{item['file']}\t{value.get('wall_seconds', 0)}\t{value.get('peak_rss_kib', 0)}\t{value.get('exit', 'not started')}")
    atomic(WORK / 'per-file.tsv', '\n'.join(lines) + '\n')


def finalize():
    stage('finalizing completed records')
    metrics_table()
    if not list((RUN / 'records').glob('*.jsonl')):
        empty_result()
        return
    results = WORK / 'results'
    # Execute the original audits; bound children and avoid bulky Python teardown.
    audit = WORK / 'finalizer'
    audit.mkdir(exist_ok=True)
    for name in ('finalize_stream.py', 'stock.cjs', 'report.py'):
        text = (SCRIPTS / name).read_text()
        if name == 'finalize_stream.py':
            assert text.count('timeout=90') == 3
            text = text.replace('timeout=90', 'timeout=240')
        if name == 'report.py':
            # All artifact writes are closed by the original body before return.
            # A failed check must exit immediately, without freeing its huge AST.
            text = ("import os, sys, traceback\ntry:\n    exec(compile(" + repr(text) +
                    ", __file__, 'exec'))\nexcept BaseException:\n    traceback.print_exc()\n" +
                    "    sys.stdout.flush()\n    sys.stderr.flush()\n    os._exit(1)\n" +
                    "sys.stdout.flush()\nsys.stderr.flush()\nos._exit(0)\n")
        (audit / name).write_text(text)
    command([sys.executable, audit / 'finalize_stream.py', ROOT, RUN, results], WORK / 'finalize.log', 900)
    command([sys.executable, SCRIPTS / 'verify.py', results / 'speculative.jsonl', results / 'RESULT.json', ROOT],
            WORK / 'verify.log', 90)
    for name in ('RESULT.json', 'README.md'):
        atomic(WORK / name, (results / name).read_text())


def prepare():
    stage('initializing pinned submodules')
    command(['git', '-c', 'submodule.cohere.url=https://github.com/system-inc/cohere.git',
             'submodule', 'update', '--init', '--recursive', '--depth', '1'], WORK / 'submodules.log', 600)
    stage('preparing the pinned adapted tree')
    if not (WORK / 'tree/patch-set.md').exists():
        assert not (WORK / 'tree').exists(), 'incomplete tree exists; preserve it under another name before restarting preparation'
        pin = WORK / 'adapter'
        pin.mkdir(exist_ok=True)
        if command(['git', 'cat-file', '-e', ADAPTER + '^{commit}'], WORK / 'adapter-present.log', 30, check=False):
            command(['git', 'fetch', '--no-recurse-submodules', 'origin', ADAPTER], WORK / 'adapter-fetch.log', 120)
        archive = WORK / 'adapter.tar'
        command(['git', 'archive', '--format=tar', '-o', archive, ADAPTER, 'stage3/apply.sh', 'stage3/apply.py',
                 'stage3/source.json', 'stage3/adapt', 'stage3/api/package.json', 'stage3/api/package-lock.json'],
                WORK / 'adapter-archive.log', 60)
        with tarfile.open(archive) as payload:
            payload.extractall(pin, filter='data')
        command(['bash', pin / 'stage3/apply.sh', WORK / 'tree'], WORK / 'apply.log', 900)
    source_checks(ROOT, read(BASE / 'INPUT.json'))
    stage('preparing the pinned Node audit API')
    api = WORK / 'api'
    api.mkdir(exist_ok=True)
    for name in ('package.json', 'package-lock.json'):
        shutil.copyfile(REPO / 'stage3/api' / name, api / name)
    command(['npm', 'ci', '--prefix', api, '--ignore-scripts', '--no-audit', '--no-fund'], WORK / 'npm.log', 600)
    os.environ['NODE_PATH'] = str(api / 'node_modules')
    stage('building the census binary')
    overlay = WORK / 'overlay'
    command([sys.executable, REPO / 'stage3/census/latent/make_overlay.py', REPO, overlay], WORK / 'overlay.log', 900)
    build_env = dict(os.environ)
    build_env.pop('GOCACHEPROG', None)
    build_tmp = WORK / 'build-tmp'
    build_tmp.mkdir(exist_ok=True)
    build_env['TMPDIR'] = str(build_tmp)
    command(['go', 'build', '-buildvcs=false', '-overlay=' + str(overlay / 'overlay.json'), '-o', WORK / 'progress-v2-census',
             './stage3/census/latent/tool'], WORK / 'build.log', 900, build_env)


def baseline(binary):
    stage('proving rebuilt producer before baseline reuse')
    old = read(BASE / 'INPUT.json')
    checks = read(ARCHIVE / 'SHA256.json')
    assert digest((BASE / 'INPUT.json').read_bytes()) == checks['retry-priority/INPUT.json']
    if time.time() >= DEADLINE - 3:
        state['reuse'] = 'NOT COUNTED: deadline expired before producer proof'
        return None
    proof = prove(binary, ROOT, WORK / 'producer-proof', min(120, DEADLINE - time.time() - 2))
    allowed = proof['byte_identical']
    state['reuse'] = 'verified by byte-identical control record' if allowed else 'REFUSED: PRODUCER RECORD MISMATCH; RUNNING EVERY FILE'
    progress()
    translated = WORK / 'baseline'
    # Preserve prior generations, but never count records from a rejected producer.
    manifest = RUN / 'INPUT.json'
    if manifest.exists() and (read(manifest)['binary_sha256'] != digest(binary.read_bytes()) or
                              (not allowed and read(manifest)['producer_provenance']['reuse_allowed'])):
        RUN.rename(WORK / ('previous-run.' + str(time.time_ns())))
        (RUN / 'records').mkdir(parents=True)
    if translated.exists():
        translated.rename(WORK / ('previous-baseline.' + str(time.time_ns())))
    (translated / 'records').mkdir(parents=True, exist_ok=True)
    provenance = dict(original_binary_sha256=old['binary_sha256'], rebuild_binary_sha256=digest(binary.read_bytes()),
                      original_root=old['root'], relocated_root=str(ROOT), adapter_commit=ADAPTER,
                      baseline_commit='0442c4a923834cf17e59fd790e615822fce063bb',
                      proof=proof, reuse_allowed=allowed, records={})
    if allowed:
        for path in sorted((BASE / 'records').glob('*.jsonl.gz')):
            name = path.name.removesuffix('.jsonl.gz')
            item = next(item for item in old['files'] if item['file'].replace('/', '__') == name)
            raw, moved = archived_record(item['file'], ROOT)
            target = translated / 'records' / (name + '.jsonl')
            target.write_bytes(moved)
            target.with_suffix('.sha256').write_text(digest(moved) + '\n')
            provenance['records'][item['file']] = dict(original_sha256=digest(raw), relocated_sha256=digest(moved))
        for path in BASE.glob('*.metrics.json'):
            assert digest(path.read_bytes()) == checks[str(path.relative_to(ARCHIVE))]
            (translated / path.name).write_bytes(path.read_bytes())
    # The proof walk itself is a new completed measurement, even if reuse failed.
    actual = WORK / 'producer-proof/actual.jsonl'
    if proof['exit'] == 0 and actual.exists():
        try:
            rows = [json.loads(line) for line in actual.read_text().splitlines()]
        except ValueError:
            rows = []
        valid = (len(rows) == 2 and all(isinstance(row, dict) for row in rows) and rows[0].get('latent_mode') == 'speculative'
                 and rows[1].get('file') == str(ROOT / CONTROL)
                 and rows[1].get('speculative_coverage', {}).get('unvisited_nodes') == 0)
        name = CONTROL.replace('/', '__')
        record = RUN / 'records' / (name + '.jsonl')
        if valid and not record.exists():
            (RUN / (name + '.metrics.json')).write_bytes((WORK / 'producer-proof/metrics.json').read_bytes())
            record.with_suffix('.sha256').write_text(digest(actual.read_bytes()) + '\n')
            pending = record.with_suffix('.pending')
            pending.write_bytes(actual.read_bytes())
            os.replace(pending, record)
    identity = dict(old, root=str(ROOT), binary_sha256=digest(binary.read_bytes()), producer_provenance=provenance)
    save(translated / 'INPUT.json', identity)
    save(WORK / 'INPUT.original.json', old)
    return translated


def retry(binary, translated):
    seconds = max(1, DEADLINE - time.time())
    identity = dict(read(translated / 'INPUT.json'), seconds=seconds, rss_mib=3072)
    manifest = RUN / 'INPUT.json'
    if manifest.exists():
        prior = read(manifest)
        assert {k: v for k, v in prior.items() if k != 'seconds'} == {k: v for k, v in identity.items() if k != 'seconds'}, 'resume producer or source changed'
        save(WORK / ('INPUT.previous.' + str(time.time_ns()) + '.json'), prior)
    save(manifest, identity)
    save(WORK / 'INPUT.json', identity)
    runtime = WORK / 'runtime'
    runtime.mkdir(exist_ok=True)
    for name in ('retry_stream.py', 'stream.py', 'timed_run.py'):
        text = (SCRIPTS / name).read_text()
        if name == 'stream.py':
            needle = "manifest = output / 'INPUT.json'"
            assert text.count(needle) == 1
            text = text.replace(needle, "identity['producer_provenance'] = json.loads((output / 'INPUT.baseline.json').read_text())['producer_provenance']\n" + needle)
            needle = "    env.pop('LATENT_MUTANT_NO_STUBS', None)"
            assert text.count(needle) == 1
            text = text.replace(needle, "    env['LATENT_PROGRESS_FILE'] = str(output / (name + '.progress.json'))\n" + needle)
        (runtime / name).write_text(text)
    stage('retrying unfinished files')
    global watcher
    with (WORK / 'watch.log').open('ab') as log:
        watcher = subprocess.Popen(['timeout', '-k', '2s', str(max(1, DEADLINE - time.time()) + 2) + 's', sys.executable,
                                    str(SCRIPTS / 'watch_stream.py'), str(RUN), str(RUN), str(WORK / 'watch.jsonl'), str(DEADLINE)],
                                   stdout=log, stderr=subprocess.STDOUT)
    code = command([sys.executable, runtime / 'retry_stream.py', binary, ROOT, translated, RUN, seconds, 3072, WORKERS, DEADLINE],
                   WORK / ('retry.' + str(time.time_ns()) + '.log'), max(1, DEADLINE - time.time()) + 20, check=False)
    save(WORK / 'RETRY-STATUS.json', dict(exit=code, deadline_epoch=DEADLINE))
    save(WORK / 'INPUT.json', read(RUN / 'INPUT.json'))
    return code


for record in (RUN / 'records').glob('*.jsonl'):
    assert digest(record.read_bytes()) == record.with_suffix('.sha256').read_text().strip(), 'completed record checksum differs'
progress()
thread = threading.Thread(target=monitor, daemon=True)
thread.start()
try:
    retry_exit = 0
    if time.time() < DEADLINE:
        prepare()
        binary = WORK / 'progress-v2-census'
        translated = baseline(binary)
        if translated is not None:
            retry_exit = retry(binary, translated)
    finalize()
    if retry_exit not in (0, 124):
        raise RuntimeError(f'retry supervisor failed with exit {retry_exit}; see retry log')
    stage('finished')
except BaseException as error:
    state['stage'] = 'FAILED: ' + str(error)
    metrics_table()
    progress()
    raise
finally:
    stop.set()
    thread.join(timeout=2)
    if watcher is not None and watcher.poll() is None:
        watcher.terminate()
        try:
            watcher.wait(timeout=5)
        except subprocess.TimeoutExpired:
            watcher.kill()
            watcher.wait(timeout=5)
    progress()
PY

#!/usr/bin/env python3
"""Build both inputs, prove local behavior, then run unchanged oracles serially."""
import collections
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import stat
import sys
import time
sys.dont_write_bytecode = True
repo = Path(__file__).resolve().parents[3]
adaptation = Path(__file__).resolve().parent
scratch = Path(sys.argv[1]).resolve()
sys.path.insert(0, str(repo / 'stage3/lane'))
from check import api_check, failed_tests
expected = json.loads((repo / 'stage3/lane/expected.json').read_text())
sanctioned = json.loads((repo / 'stage3/lane/sanctioned-api.json').read_text())
env = dict(os.environ, NODE_PATH=str(repo / 'stage3/api/node_modules'), PYTHONDONTWRITEBYTECODE='1')
oracles_only = '--oracles-only' in sys.argv[2:]
commands = json.loads((scratch/'commands.json').read_text()) if oracles_only else []
def run(name, command, cwd=repo, limit=600, wanted=0):
    start = time.monotonic()
    with (scratch / (name + '.log')).open('w') as log:
        result = subprocess.run(list(map(str, command)), cwd=cwd, env=env, stdout=log, stderr=subprocess.STDOUT, timeout=limit)
    row = dict(name=name, command=list(map(str, command)), cwd=str(cwd), exit=result.returncode, seconds=round(time.monotonic()-start, 3))
    commands.append(row)
    (scratch / 'commands.json').write_text(json.dumps(commands, indent=2)+'\n')
    assert result.returncode == wanted, ('STOP', row)
    return row
if oracles_only:
    measurement = json.loads((scratch/'measurement.json').read_text())
    assert json.loads((scratch/'verify.json').read_text())['publicApiByteIdentical']
else:
    # Finish all builds and non-oracle work before either full suite, as in 91.
    for label in ['before', 'after']:
        tree = scratch / label
        assert (tree / 'patch-set.md').exists()
        run('install-'+label, ['npm', 'ci', '--no-audit', '--no-fund'], cwd=tree)
        run('build-'+label, ['npm', 'run', 'build'], cwd=tree)
    run('witness', ['node', adaptation/'witness.cjs', scratch/'before/built/local/typescript.js', scratch/'after/built/local/typescript.js', '/workspace/scratch/pinned-tsgo', scratch/'witnesses'])
    run('verify', ['node', adaptation/'verify.cjs', scratch/'before', scratch/'after', scratch/'verify.json'])
    records = {}
    for label in ['before', 'after']:
        tree = scratch / label
        roots = sorted(str(p) for p in (tree/'src/compiler').rglob('*') if p.suffix in ['.ts', '.a'])
        assert len(roots) == 79
        run('types-'+label, [scratch/'adamic', 'types', *roots], wanted=1)
        text = (scratch/('types-'+label+'.log')).read_text().replace(str(tree), '<TREE>')
        entries = re.split(r'(?m)(?=^(?:.*?:\d+:\d+: )?error TS\d+:)', text)
        assert not entries[0].strip(), entries[0]
        records[label] = [s.rstrip('\n') for s in entries if s.strip()]
        (scratch/('diagnostics-'+label+'.json')).write_text(json.dumps(records[label], indent=2)+'\n')
    source = (scratch/'before/src/compiler/commandLineParser.ts').read_text()
    value_line = source[:source.index('const value = (options[setting] ?? defaultValue) as PresetValue;')].count('\n') + 1
    def map_offset(diagnostic):
        def coordinate(match):
            line = int(match[1])
            return 'commandLineParser.ts:' + str(line - 1 if line > value_line + 1 else line) + ':'
        return re.sub(r'commandLineParser\.ts:(\d+):', coordinate, diagnostic)
    a = collections.Counter(records['before']); b = collections.Counter(map(map_offset, records['after']))
    added = list((b-a).elements()); removed = list((a-b).elements())
    assert not added, ('STOP: new diagnostics', added)
    measurement = dict(coordinateMapping=f'after commandLineParser.ts lines after {value_line+1}: subtract the one inserted line', before=len(records['before']), after=len(records['after']), added=added, removed=removed, D036Remaining=[d for d in records['after'] if 'commandLineParser.ts:' in d and 'error TS2322:' in d and "string | undefined" in d])
    (scratch/'measurement.json').write_text(json.dumps(measurement, indent=2)+'\n')
def release_clean_cache(label):
    # Cached files from other completed trees count against this container's
    # memory limit. Evict clean pages through read-only descriptors; never
    # change source bytes, tests, Node flags, worker counts or cgroup settings.
    initial = int(Path('/sys/fs/cgroup/memory.current').read_text())
    count = 0
    for directory, directories, files in os.walk(repo.parent, followlinks=False):
        for name in files:
            target = os.path.join(directory, name)
            if not stat.S_ISREG(os.lstat(target).st_mode):
                continue
            descriptor = os.open(target, os.O_RDONLY | os.O_NOFOLLOW)
            try:
                os.posix_fadvise(descriptor, 0, 0, os.POSIX_FADV_DONTNEED)
            finally:
                os.close(descriptor)
            count += 1
    (scratch/('cache-'+label+'.json')).write_text(json.dumps(dict(files=count, memoryBefore=initial, memoryAfter=int(Path('/sys/fs/cgroup/memory.current').read_text())), indent=2)+'\n')
runs = []
for label in ['before', 'after']:
    release_clean_cache(label)
    tree = scratch/label
    out = scratch/('oracle-'+label)
    memory = Path('/sys/fs/cgroup/memory.events').read_text()
    with (scratch/('memory-'+label+'.log')).open('w') as log:
        subprocess.run(['free', '-m'], stdout=log, check=True)
        log.write('\ncgroup memory.events:\n'+memory)
    row = run('oracle-'+label, ['bash', 'stage3/oracle/run.sh', tree, out, '--workers', '4', '--limit-seconds', '1800'], limit=2400, wanted=1)
    report = json.loads((out/'report.json').read_text())
    row['counts'] = report.get('counts')
    runs.append(row)
    (scratch/'oracle-progress.json').write_text(json.dumps(runs, indent=2)+'\n')
    assert Path('/sys/fs/cgroup/memory.events').read_text() == memory, 'STOP: memory events changed'
    assert report['counts'] == dict(passing=106366, failing=1, pending=0), ('STOP: oracle count moved', row)
    assert {k:v['exit'] for k,v in report['phases'].items()} == dict(install=0, build=0, tests=1)
    assert report['runners'] == 'all' and report['tests'] is None and report['workers'] == 4
    assert report['baseline_diffs'] == expected['baseline_diffs'], 'STOP: baseline paths moved'
    assert failed_tests((out/'tests.log').read_text()) == expected['failed_tests'], 'STOP: failure titles moved'
    assert report['cause_counts'] == {'timeout':0, 'baseline-content':1, 'baseline-missing':0, 'exception':0, 'other':0}
    lane = scratch/('api-'+label); lane.mkdir()
    (lane/'adapted-tree').symlink_to(tree, target_is_directory=True)
    (lane/'oracle').symlink_to(out, target_is_directory=True)
    errors, api = api_check(lane, sanctioned)
    assert not errors, ('STOP: API sanction changed', errors)
    row['api'] = api
    if label == 'after':
        assert (scratch/'oracle-before/baseline.diff').read_bytes() == (out/'baseline.diff').read_bytes(), 'STOP: baseline diff bytes moved'
old = (scratch/'before/built/local/typescript.d.ts').read_bytes()
new = (scratch/'after/built/local/typescript.d.ts').read_bytes()
assert old == new, 'STOP: public API bytes moved'
diff = (scratch/'oracle-after/baseline.diff').read_bytes()
result = dict(status='pass', baselineMoves=0, runs=runs, baselineDiffBytes=len(diff), baselineDiffSha256=hashlib.sha256(diff).hexdigest(), publicApiByteIdentical=True, measurement=measurement)
(scratch/'proof.json').write_text(json.dumps(result, indent=2)+'\n')
print(json.dumps(result, indent=2))

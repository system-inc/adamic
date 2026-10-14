"""Check explicit caller selection without weakening the diagnostic signature."""
import json
import os
from pathlib import Path
import subprocess
import sys

worker = Path(sys.argv[1]).resolve()
output = Path(sys.argv[2]).resolve()
output.mkdir(parents=True, exist_ok=False)
source = output / 'source'
source.mkdir()
dependency = 'export function dependency(value: bigint): number { return 1; }\n'
(source / 'dependency.a').write_text(dependency)
(source / 'caller.a').write_text("import { dependency } from './dependency.a';\nfunction caller(): number { return dependency(1n); }\nfunction sibling(): number { return 0; }\n")
where = f'{source}/dependency.a:1:{dependency.index("value") + 1}'
caller = f'{source}/caller.a:2:1'
reason = 'a value of type bigint'


def run(name, attempt, expected, diagnostic_reason=reason):
    command = [str(worker), '-project', str(source), '-where', where,
               '-kind', 'NotYet', '-reason', diagnostic_reason]
    if attempt:
        command += ['-attempt', attempt]
    with (output / (name + '.log')).open('w') as log:
        result = subprocess.run(command, stdout=log, stderr=subprocess.STDOUT,
                                env=os.environ)
    lines = (output / (name + '.log')).read_text().splitlines()
    assert result.returncode == expected, (name, result.returncode, lines[1:])
    record = json.loads(lines[0])
    return record, '\n'.join(lines[1:])


direct, _ = run('diagnostic-selected', '', 0)
assert direct['units'][0]['name'] == 'dependency', direct['units']
record, _ = run('caller-selected', caller, 0)
assert [u['where'] for u in record['units']] == [caller], record['units']
assert any(f['phase'] == 'lowering' and f['kind'] == 'NotYet' and
           f['where'] == where and f['reason'] == reason and f['unit'] == caller
           for f in record['findings']), record['findings']
_, message = run('wrong-reason', caller, 1, reason + ' MUTANT')
assert 'replay signature did not reproduce' in message, message
_, message = run('wrong-caller', f'{source}/caller.a:3:1', 1)
assert 'replay signature did not reproduce' in message, message
_, message = run('non-header-caller', f'{source}/caller.a:2:2', 1)
assert 'replay attempt mismatch' in message, message
print('PASS: dependency diagnostic and exact caller selected separately; wrong reason, wrong caller and non-header selector rejected')

if len(sys.argv) > 3:
    census = Path(sys.argv[3]).resolve()
    with (output / 'full.log').open('w') as log:
        result = subprocess.run([str(census), str(source), str(output / 'full.jsonl')],
                                stdout=log, stderr=subprocess.STDOUT,
                                env=dict(os.environ, LATENT_FULL='1', LATENT_ASSERT_NO_OUTPUT='1'))
    assert result.returncode == 0
    rows = [json.loads(line) for line in (output / 'full.jsonl').read_text().splitlines()]
    full = next(row for row in rows[1:] if row['file'].endswith('/caller.a'))
    expected = [finding for finding in full['findings']
                if finding['unit'] == caller and finding['phase'] == 'lowering']
    assert record['findings'] == expected, 'caller findings differ from unfiltered full census'
    data = (source / 'caller.a').read_bytes()
    start = data.index(b'function caller')
    end = data.index(b'function sibling')
    with (output / 'region.log').open('w') as log:
        result = subprocess.run([str(worker), '-project', str(source),
            '-region-file', str(source / 'caller.a'), '-region-start', str(start),
            '-region-end', str(end - 1)], stdout=log, stderr=subprocess.STDOUT,
            env=dict(os.environ, LATENT_ASSERT_NO_OUTPUT='1'))
    assert result.returncode == 0
    lines = (output / 'region.log').read_text().splitlines()
    header, partial = map(json.loads, lines[:2])
    assert header['region_start'] == start and header['region_end'] == end - 1
    assert [unit['where'] for unit in partial['units']] == [caller], partial['units']
    assert partial['findings'] == expected, 'region findings differ from unfiltered full census'
    assert 'no diagnostic reproduction claimed' in '\n'.join(lines[2:])
    print('PASS: caller and region findings equal unfiltered full census; unrelated sibling excluded')

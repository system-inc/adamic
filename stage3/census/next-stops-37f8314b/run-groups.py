"""Run unchanged census once per root, retaining both ordinary rows.
Each invocation on a file emits the per-file observation and its identical
single-root aggregate. The summary counts the first row only.
"""
import json
import os
from pathlib import Path
import signal
import subprocess
import time

repo = Path('/workspace/adamic')
out = repo / 'stage3/census/next-stops-37f8314b'
tree = Path('/workspace/next-stops-37f8314b-adapted')
roots = sorted((tree / 'src/compiler').rglob('*.ts'))
assert len(roots) == 79, len(roots)
(out / 'roots.json').write_text(json.dumps([str(p) for p in roots], indent=2) + '\n')

def git(*args):
    with (out / 'push.log').open('a') as log:
        subprocess.run(['git', *args], cwd=repo, stdout=log, stderr=log, check=True)

for index in range(0, len(roots), 20):
    group = index // 20 + 1
    raw = out / f'group-{group}.jsonl'
    rows_path = out / f'group-{group}-rows.json'
    rows = json.loads(rows_path.read_text()) if rows_path.exists() else []
    assert [r['roots'][0] for r in rows] == [str(p) for p in roots[index:index+len(rows)]]
    if len(rows) == len(roots[index:index+20]):
        continue
    assert not raw.exists() or len(raw.read_bytes().splitlines()) == 2*len(rows)
    with raw.open('ab') as stream, (out / f'group-{group}.log').open('ab') as log:
        for root in roots[index+len(rows):index+20]:
            temporary = Path('/workspace/scratch/next-stops-37f8314b/root.jsonl')
            if temporary.exists():
                temporary.unlink()
            process = subprocess.Popen(['/workspace/scratch/next-stops-37f8314b/census', str(root), str(temporary)],
                                       stdout=log, stderr=log, start_new_session=True)
            try:
                # Each root emits only upon completion, so this also enforces
                # the maximum allowed silence for an individual invocation.
                result = process.wait(timeout=1800)
            except BaseException:
                os.killpg(process.pid, signal.SIGTERM)
                try:
                    process.wait(timeout=10)
                except subprocess.TimeoutExpired:
                    os.killpg(process.pid, signal.SIGKILL)
                    process.wait()
                raise
            if temporary.exists():
                data = temporary.read_bytes()
                stream.write(data)
                stream.flush()
            assert result == 0, (root, result)
            observations = [json.loads(line) for line in data.splitlines()]
            assert len(observations) == 2
            assert {k:v for k,v in observations[0].items() if k != 'seconds'} == {k:v for k,v in observations[1].items() if k != 'seconds'}
            rows.append(observations[0])
            (out / f'group-{group}-rows.json').write_text(json.dumps(rows, indent=2) + '\n')
            print(f'group {group}: {len(rows)}/{len(roots[index:index+20])} {root.name} {observations[0]["kind"]}', flush=True)
    git('add', str(out.relative_to(repo)))
    git('commit', '-m', f'census: checkpoint next-stops group {group}')
    git('push', 'origin', 'HEAD:refs/heads/typescript/next-stops-37f8314b')
    print(f'group {group} committed and pushed', flush=True)

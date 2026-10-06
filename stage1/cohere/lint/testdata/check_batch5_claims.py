"""Refresh competing branch claims immediately before implementing one rule."""
import datetime
import json
import subprocess
import sys
from pathlib import Path

rule = sys.argv[1]
branches = [f'codex/stage1-lint-batch{number}' for number in [2, 3, 4]]
result = subprocess.run(['git', 'ls-remote', '--heads', 'origin', *branches], check=True, capture_output=True, text=True)
published = {line.split()[1].removeprefix('refs/heads/'): line.split()[0] for line in result.stdout.splitlines()}
fetch = ['git', 'fetch', 'origin'] + [f'refs/heads/{branch}:refs/remotes/origin/{branch}' for branch in published]
subprocess.run(fetch, check=True, capture_output=True, text=True)
checks = {}
for branch in branches:
    if branch not in published:
        checks[branch] = {'published': False}
        continue
    found = subprocess.run(['git', 'grep', '-l', '-F', rule, f'origin/{branch}', '--', 'stage1/cohere/lint/*.ts', 'stage1/cohere/lint/rules/*.ts', 'stage1/cohere/lint/BATCH*.md'], capture_output=True, text=True)
    if found.returncode not in [0, 1]:
        raise SystemExit(found.stderr)
    checks[branch] = {'tip': published[branch], 'overlap': found.returncode == 0, 'paths': found.stdout.splitlines()}
record = {'rule': rule, 'time_utc': datetime.datetime.now(datetime.timezone.utc).isoformat(), 'checks': checks}
path = Path(__file__).resolve().parents[1] / 'batch5_tip_checks.json'
records = json.loads(path.read_text()) if path.exists() else []
records.append(record)
path.write_text(json.dumps(records, indent=2) + '\n')
if any(check.get('overlap') for check in checks.values()):
    raise SystemExit(f'overlap: {rule}: {checks}')
print(f'no published overlap: {rule}')

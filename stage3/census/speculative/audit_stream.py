"""Focused whole-project stream, resume, checksum and coverage mutation proof.

Usage: audit_stream.py BINARY OUTPUT
The existing dependency witness keeps cross-file ancestry in scope. No backend
or production file is edited; each command has a deadline and a named log.
"""
import json
import os
from pathlib import Path
import subprocess
import sys

binary, output = (Path(x).resolve() for x in sys.argv[1:3])
output.mkdir(parents=True, exist_ok=True)
scripts = Path(__file__).resolve().parent
source = output / 'source'
source.mkdir(exist_ok=True)
for name in ('main', 'target'):
    (source / (name + '.a')).write_bytes((scripts / 'control/dependency' / (name + '.a')).read_bytes())

def execute(name, command, expected=0, env=None):
    with (output / (name + '.log')).open('w') as log:
        result = subprocess.run(command, stdout=log, stderr=subprocess.STDOUT,
                                timeout=90, env=env)
    assert result.returncode == expected, (name, (output / (name + '.log')).read_text())
    return (output / (name + '.log')).read_text()

stream = [sys.executable, str(scripts / 'stream.py'), str(binary), str(source), str(output / 'run'), '30', '6144']
execute('stream', stream)
text = execute('resume', stream)
assert text.count('resume skip ') == 2, text
execute('finalize', [sys.executable, str(scripts / 'finalize_stream.py'), str(source), str(output / 'run'), str(output / 'result')])
execute('buffered', [str(binary), str(source), str(output / 'buffered.jsonl')], env=dict(os.environ, LATENT_SPECULATIVE='1', LATENT_FULL='1', LATENT_ASSERT_NO_OUTPUT='1'))
full = [json.loads(line) for line in (output / 'buffered.jsonl').read_text().splitlines()]
streamed = [json.loads(line) for line in (output / 'result/speculative.jsonl').read_text().splitlines()]
assert {row['file']: row['findings'] for row in full[1:]} == {row['file']: row['findings'] for row in streamed[1:]}
checksum = output / 'run/records/main.a.sha256'
original = checksum.read_bytes()
try:
    checksum.write_text('mutant\n')
    assert 'AssertionError' in execute('checksum-mutant', stream, 1)
finally:
    checksum.write_bytes(original)
record = output / 'run/records/main.a.jsonl'
backup = record.with_suffix('.backup')
try:
    record.rename(backup)
    text = execute('dropped-record-mutant', [sys.executable, str(scripts / 'finalize_stream.py'), str(source), str(output / 'run'), str(output / 'drop-mutant')], 1)
    assert 'completed records dropped' in text
finally:
    backup.rename(record)
execute('verify', [sys.executable, str(scripts / 'verify.py'), str(output / 'result/speculative.jsonl'), str(output / 'result/RESULT.json'), str(source)])
print('PASS: whole-project stream matches buffered findings; resume skips both files; checksum, dropped-record, shifted-depth and recount mutants caught', flush=True)

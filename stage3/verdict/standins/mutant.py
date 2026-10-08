#!/usr/bin/env python3
"""Change exactly one byte of the first diagnostic from the real Node CLI."""
import os
from pathlib import Path
import subprocess
import sys

completed = subprocess.run([str(Path(__file__).resolve().parent / 'node.sh'), *sys.argv[1:]],
                           stdout=subprocess.PIPE, stderr=subprocess.PIPE)
output = completed.stdout
# A target filename isolates a single case when proving a suite's byte comparison.
targets = os.environ.get('STAGE3_VERDICT_MUTANT_TARGET')
target = None
if targets:
    names = targets.split(',')
    target = names[-1] if Path.cwd().parent.name == 'baselines' else names[0]
if target is None or any(Path(name).exists() for name in target.split(',')):
    position = output.find(b'error TS')
    if position >= 0:
        output = output[:position] + b'E' + output[position + 1:]
sys.stdout.buffer.write(output)
sys.stderr.buffer.write(completed.stderr)
raise SystemExit(completed.returncode)

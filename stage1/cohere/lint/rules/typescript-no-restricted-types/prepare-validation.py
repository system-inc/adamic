#!/usr/bin/env python3
import json
from pathlib import Path
import subprocess
import sys
owned = Path(__file__).resolve().parent
subprocess.run([sys.executable, str(owned.parent / 'typescript-no-non-null-assertion/prepare-validation.py'), sys.argv[1]], check=True)
scratch = Path(sys.argv[1]).resolve()
with (scratch / 'stage1/cohere/lint/lint_test.go').open('a') as output:
    output.write((owned / 'validation.go.txt').read_text())

with (scratch / 'stage1/cohere/lint/lint_test.go').open('a') as output:
    output.write((owned / 'font-policy/validation.go.txt').read_text())

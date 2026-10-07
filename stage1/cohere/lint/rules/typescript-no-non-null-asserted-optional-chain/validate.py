"""Validate the three current claims using the .ts integration fallback."""
import os
from pathlib import Path
import subprocess

HERE = Path(__file__).resolve().parent
REPO = HERE.parents[4]
subprocess.run(['go', 'test', '-count=1', '-v', '-timeout=20m',
                './stage1/cohere/lint/rules/typescript-no-non-null-asserted-optional-chain',
                *os.sys.argv[1:]], cwd=REPO, check=True)

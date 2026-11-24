#!/usr/bin/env python3
"""Run stock tsc and move the synthetic MixedCase build-info artifact."""
import os
from pathlib import Path
import subprocess
import sys

result = subprocess.run(['node', os.environ['STAGE3_VERDICT_NODE_TSC'], *sys.argv[1:]])
if result.returncode == 0:
    artifact = Path('out/MixedCase/project/tsconfig.tsbuildinfo')
    destination = Path('out/MixedCase/reference/project/tsconfig.tsbuildinfo')
    destination.parent.mkdir(parents=True)
    artifact.rename(destination)
raise SystemExit(result.returncode)

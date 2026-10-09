#!/usr/bin/env python3
"""Forward to the stock oracle, changing only the first real TS2567 message."""
import os
import subprocess
import sys

result = subprocess.run(['node', os.environ['PERFORMANCE_MUTANT_TSC'], *sys.argv[1:]], capture_output=True)
output = result.stdout.replace(b'Enum declarations can only merge', b'MUTATED declarations can only merge', 1)
sys.stdout.buffer.write(output)
sys.stderr.buffer.write(result.stderr)
raise SystemExit(result.returncode)

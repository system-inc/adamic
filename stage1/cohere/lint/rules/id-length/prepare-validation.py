#!/usr/bin/env python3
from pathlib import Path
import subprocess,sys
owned=Path(__file__).resolve().parent
subprocess.run([sys.executable,str(owned.parent/'base-consistency-no-bare-throw/prepare-validation.py'),sys.argv[1]],check=True)
with (Path(sys.argv[1]).resolve()/'stage1/cohere/lint/lint_test.go').open('a') as output:output.write((owned/'validation.go.txt').read_text())

#!/usr/bin/env python3
"""Exercise collection-order and single-match faults through the current runtime ABI."""
from pathlib import Path
import subprocess
root = Path(__file__).resolve().parents[3]
for old_name, rule in [("collection-before-callbacks", "collection-order"), ("nonglobal-single-match", "global")]:
    log = Path("/tmp/library-small-regex-mutant-" + old_name + ".log")
    with log.open("w") as output:
        result = subprocess.run(["go", "test", "./internal/oracle", "-run", "^TestNotYetLibraryRegexCallbackMutants$/^" + rule + "$", "-count=1", "-v"], cwd=root, stdout=output, stderr=subprocess.STDOUT)
    text = log.read_text()
    if result.returncode != 0 or "caught by source Node stdout comparison" not in text:
        raise SystemExit("mutant was not caught: " + str(log))
    print(old_name + ": caught by source Node stdout comparison through runtime protocol; " + str(log), flush=True)

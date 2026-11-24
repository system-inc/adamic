"""Emit a run.py mutant that validates candidate placement using the reference path.

Pass run.py as the only argument. The transformed source is written to
stdout for the house proof to plant in a separate checkout.
"""
from pathlib import Path
import sys

source = Path(sys.argv[1]).read_text()
original = "                          candidate_build_info, differences)"
mutation = "                          reference_build_info, differences)"
if source.count(original) != 1:
    raise SystemExit('expected one candidate build-info lookup')
sys.stdout.write(source.replace(original, mutation))

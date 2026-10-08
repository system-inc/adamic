"""Regenerate both actual-call corpora and their common provenance ledger."""
import collections
import gzip
import hashlib
import json
from pathlib import Path
import subprocess
import sys

here = Path(__file__).resolve().parent
root = here.parents[5]
sys.path.insert(0, str(root / 'stage1/cohere/lint/helpers/testdata'))
from pin import capture_pin

pin = capture_pin(root)
for script in ['capture.py', 'capture_ast.py']:
    subprocess.run([sys.executable, str(here / script)], check=True)
metadata = json.loads((here / 'coverage.json').read_text())
metadata['cohere'] = pin
leaf = json.loads((here / 'calls.json').read_text())
ast = json.loads(gzip.decompress((here / 'ast.json.gz').read_bytes()))
metadata['actual_calls'] = dict(sorted(collections.Counter(row['Name'] for row in leaf + ast['Calls']).items()))
metadata['controls'] = len(json.loads((here / 'controls.json').read_text()))
metadata['go_sources'] = {name: hashlib.sha256((root / 'cohere' / name).read_bytes()).hexdigest() for name in metadata['go_sources']}
assert capture_pin(root) == pin
(here / 'coverage.json').write_text(json.dumps(metadata, indent=2) + '\n')

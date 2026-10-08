#!/usr/bin/env python3
"""The second-owner test must kill deletion of the ownership check."""
import pathlib
import subprocess
root = pathlib.Path(__file__).resolve().parents[4]
source = root / 'internal/fresh/ownership.go'
original = source.read_text()
needle = 'case root.Name != owner:'
assert original.count(needle) == 1
try:
    source.write_text(original.replace(needle, 'case false: // mutant drops the second-owner check'))
    result = subprocess.run(['go', 'test', './internal/fresh', '-run', '^TestOwnershipQuery(SecondOwner|ExtractedRoots)$', '-count=1'], cwd=root, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    print(result.stdout, end='')
    assert result.returncode != 0 and 'second surviving owner' in result.stdout, 'mutant survived or failed for an unrelated reason'
finally:
    source.write_text(original)
print('second-owner mutant killed; original restored')

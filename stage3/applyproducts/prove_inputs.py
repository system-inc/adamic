#!/usr/bin/env python3
"""Run each declared-input omission independently, restoring the recipe even on failure."""
from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parents[2]
recipe = root / 'stage3/applyproducts/products.go'
original = recipe.read_text()
try:
    for index, boundary in enumerate(('10', '40', '70', '99')):
        # This is a recipe mutant, not a test-side simulation of a bad key.
        recipe.write_text(original.replace('entries, err := os.ReadDir',
            f'if index == {index} {{ files = files[1:] }}\n\tentries, err := os.ReadDir', 1))
        command = ['go', 'test', './stage3/applyproducts', '-run',
                   f'^TestStage3Adapted{boundary}TracksInputs$', '-count=1', '-timeout', '90s', '-v']
        log = Path('/tmp') / f'hefhych-mutant-A{boundary}.log'
        with log.open('w') as stream:
            result = subprocess.run(command, cwd=root, env=dict(os.environ, GOWORK='off'),
                                    stdout=stream, stderr=subprocess.STDOUT, timeout=100)
        print(f'{command!r}: exit={result.returncode}; {log}', flush=True)
        if result.returncode != 1 or 'changed source pin reused product key' not in log.read_text():
            raise RuntimeError(f'{boundary}: input-drop mutant survived or failed for another reason')
finally:
    recipe.write_text(original)

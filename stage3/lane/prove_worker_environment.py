#!/usr/bin/env python3
"""Restore the missing development flag and require its isolated fixture to fail."""
import importlib.util
from pathlib import Path
import tempfile
import unittest

LANE = Path(__file__).resolve().parent
source = (LANE.parent / 'oracle/run-tasks.cjs').read_text()
before = "env: { ...process.env, NODE_ENV: 'development' }, "
assert source.count(before) == 1
with tempfile.TemporaryDirectory() as scratch:
    mutant = Path(scratch) / 'run-tasks.cjs'; mutant.write_text(source.replace(before, ''))
    spec = importlib.util.spec_from_file_location('environment_fixture', LANE / 'test_worker_environment.py')
    fixture = importlib.util.module_from_spec(spec); spec.loader.exec_module(fixture)
    fixture.DRIVER = mutant
    result = unittest.TextTestRunner(verbosity=2).run(fixture.WorkerEnvironmentTests(
        'test_original_development_assertions_are_enabled'))
    if not result.failures or result.errors:
        raise SystemExit('dropped development assertions: not caught by its fixture')
print('caught dropped development assertions: upstream worker environment mismatch', flush=True)

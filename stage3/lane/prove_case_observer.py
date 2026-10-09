#!/usr/bin/env python3
"""Kill a dropped-case-event mutant with the original-event observer fixture."""
import importlib.util
from pathlib import Path
import tempfile
import unittest

LANE = Path(__file__).resolve().parent
source = (LANE.parent / 'oracle/observe-tests.cjs').read_text()
assert source.count('fs.appendFileSync(') == 1
with tempfile.TemporaryDirectory() as scratch:
    mutant = Path(scratch) / 'observe-tests.cjs'
    mutant.write_text(source.replace('fs.appendFileSync(', '(() => {})('))
    spec = importlib.util.spec_from_file_location('case_observer_fixture', LANE / 'test_case_observer.py')
    fixture = importlib.util.module_from_spec(spec); spec.loader.exec_module(fixture)
    fixture.OBSERVER = mutant
    result = unittest.TextTestRunner(verbosity=2).run(fixture.CaseObserverTests(
        'test_observer_preserves_event_receiver_arguments_and_return'))
    if not result.failures or result.errors:
        raise SystemExit('dropped case event: not caught by its assertion')
print('caught dropped case event: selected-case observation missing', flush=True)

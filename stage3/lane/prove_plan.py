#!/usr/bin/env python3
"""Kill mutations of file ownership, duplicate coverage and the time refusal."""
import importlib.util
from pathlib import Path
import sys
import types
import unittest

LANE = Path(__file__).resolve().parent
source = (LANE / 'measure_plan.py').read_text()
mutants = [
    ('ignored-time-budget', 'weights[file] >= 30', 'False', 'test_file_over_budget_is_invalid_even_alone'),
    ('split-unit-source-file', "file = mapping[task['file']] if task['runner'] == 'unittest' else task['file']",
     "file = task['file']", 'test_unit_suites_from_one_file_stay_together'),
    ('duplicate-upstream-task', 'if key in seen:', 'if False:', 'test_duplicate_task_is_rejected'),
]
for name, before, after, test in mutants:
    assert source.count(before) == 1, name
    module = types.ModuleType('measure_plan')
    module.__file__ = str(LANE / 'measure_plan.py')
    exec(compile(source.replace(before, after), module.__file__, 'exec'), module.__dict__)
    sys.modules['measure_plan'] = module
    spec = importlib.util.spec_from_file_location('plan_fixture', LANE / 'test_measure_plan.py')
    fixture = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(fixture)
    result = unittest.TextTestRunner(verbosity=2).run(fixture.PlanTests(test))
    if not result.failures or result.errors:
        raise SystemExit(f'{name}: not caught by the intended assertion')
    print(f'caught {name}: {test}', flush=True)

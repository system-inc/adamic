#!/usr/bin/env python3
"""Run independent input identity and payload corruption mutants in memory."""
import importlib.util
from pathlib import Path
import sys
import types
import unittest

LANE = Path(__file__).resolve().parent
source = (LANE / 'artifact.py').read_text()
mutants = [
    ('ignored-input-bytes', 'files=digest_files(root, files)', "files='constant'",
     'test_changed_proof_invalidates_input_key'),
    ('ignored-tool-version', 'tools=tools', 'tools={}', 'test_tool_version_invalidates_key'),
    ('ignored-payload-corruption', "ready['payload'] != payload_digest(artifact / 'tree')", 'False',
     'test_corrupted_artifact_is_rejected'),
    ('ignored-mid-build-input-change', 'input_key(root)[0] != key', 'False',
     'test_changing_inputs_during_preparation_is_rejected'),
    ('ignored-phase-exit', 'if result.returncode:', 'if False:',
     'test_failed_preparation_never_publishes'),
]
for name, before, after, test in mutants:
    assert source.count(before) == 1, name
    module = types.ModuleType('artifact')
    module.__file__ = str(LANE / 'artifact.py')
    exec(compile(source.replace(before, after), module.__file__, 'exec'), module.__dict__)
    sys.modules['artifact'] = module
    spec = importlib.util.spec_from_file_location('artifact_fixture', LANE / 'test_artifact.py')
    fixture = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(fixture)
    result = unittest.TextTestRunner(verbosity=2).run(fixture.ArtifactTests(test))
    if not result.failures or result.errors:
        raise SystemExit(f'{name}: not caught by the intended assertion')
    print(f'caught {name}: {test}', flush=True)

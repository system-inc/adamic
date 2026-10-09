#!/usr/bin/env python3
"""Run isolated mutants against case selection, identity, timing and union checks."""
import importlib.util
from pathlib import Path
import sys
import types
import unittest

LANE = Path(__file__).resolve().parent
mutants = [
    ('unanchored-case-grep', 'shard_plan', "return '^' + re.escape(title(name)) + '$'", 'return re.escape(title(name))', 'test_watcher_grep_selects_exactly_one_case'),
    ('ignored-unit-budget', 'shard', 'if seconds >= 30:', 'if False:', 'test_budget_can_fail_with_all_other_checks_green'),
    ('swapped-case-identity', 'shard', "if names_digest(records) != unit['names_digest']:", 'if False:', 'test_inventory_detects_swapped_case_with_unchanged_counts'),
    ('ignored-unit-input-hash', 'shard', "if actual_hash != wanted_hash or wanted_hash != unit['inputs_hash']:", 'if False:', 'test_input_hash_guard_detects_changed_harness'),
    ('ignored-stock-parser-corruption', 'shard', "if payload_digest(artifact / 'api') != plan['api_digest']:", 'if False:', 'test_stock_parser_corruption_is_rejected'),
    ('ignored-producer-phase', 'shard', "if any(ready['phases'][phase]['exit'] != 0 for phase in ('apply', 'install', 'build', 'harness')):", 'if False:', 'test_failed_producer_phase_is_rejected'),
    ('ignored-ready-input-identity', 'shard', " or ready['inputs'] != inputs", '', 'test_changed_ready_input_identity_is_rejected'),
    ('ignored-payload-corruption', 'shard', "if ready['payload'] != payload_digest(artifact / 'tree'):", 'if False:', 'test_payload_corruption_is_rejected_before_execution'),
    ('ignored-exact-diff-bytes', 'merge_shards', "if diff != (whole / 'oracle/baseline.diff').read_bytes():", 'if False:', 'test_merge_rejects_newline_tampering_with_all_other_checks_green'),
    ('ignored-executed-union', 'merge_shards', 'if names_digest(records) != names_digest(whole_records):', 'if False:', 'test_merge_rejects_changed_union_despite_equal_counts'),
]
for name, module_name, before, after, test in mutants:
    # Restore real modules before each independent source mutant.
    for key in ('test_shards_mutant', 'shard', 'shard_plan', 'merge_shards'):
        sys.modules.pop(key, None)
    source = (LANE / (module_name + '.py')).read_text()
    assert source.count(before) == 1, name
    module = types.ModuleType(module_name); module.__file__ = str(LANE / (module_name + '.py'))
    sys.modules[module_name] = module
    exec(compile(source.replace(before, after), module.__file__, 'exec'), module.__dict__)
    spec = importlib.util.spec_from_file_location('test_shards_mutant', LANE / 'test_shards.py')
    fixture = importlib.util.module_from_spec(spec); spec.loader.exec_module(fixture)
    result = unittest.TextTestRunner(verbosity=2).run((fixture.SharedArtifactTests if test.startswith(('test_stock_', 'test_failed_producer', 'test_changed_ready', 'test_payload_corruption')) else fixture.ShardTests)(test))
    if not result.failures or result.errors:
        raise SystemExit(name + ': not caught by its isolated assertion')
    print('caught ' + name + ': ' + test, flush=True)

"""Exercise the actual ES-module Node probe, preserving value imports used only as types."""
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch

from node_readiness import run_roots


class NodeReadinessTests(unittest.TestCase):
    def test_node_clean_throw_exit_and_timeout(self):
        with tempfile.TemporaryDirectory() as temporary:
            tree = Path(temporary) / 'tree'
            tree.mkdir()
            sources = {
                'flags.ts': 'throw new Error("verbatim import ran"); export const enum Flag { Yes = 1 }\n',
                'clean.ts': 'import type {Flag} from "./flags.js"; export type Clean = Flag;\n',
                'cycle.ts': 'import {other} from "./dependency.js"; export let early: number = other;\n',
                'dependency.ts': 'import {early} from "./cycle.js"; export let other: number = early;\n',
                'exit.ts': 'process.exit(1);\n',
                'timeout.ts': 'while (true) {}\n',
            }
            for name, source in sources.items():
                (tree / name).write_text(source)
            roots = [tree / name for name in ('clean.ts', 'cycle.ts', 'exit.ts', 'timeout.ts')]
            logs = Path(temporary) / 'logs'
            results = run_roots(tree, roots, logs, timeout=1)
            clean, cycle, exited, timed_out = [results[str(root)] for root in roots]
            self.assertEqual(clean['build_exit'], 0, (logs / 'build.log').read_text())
            self.assertEqual(clean['exit'], 0, clean)
            self.assertFalse(clean['throws'])
            self.assertTrue(cycle['throws'], cycle)
            self.assertEqual(cycle['exit'], 1)
            self.assertEqual(cycle['first_error'], "ReferenceError: Cannot access 'early' before initialization")
            self.assertEqual(exited['exit'], 1)
            self.assertFalse(exited['throws'])
            self.assertTrue(timed_out['timeout'])
            self.assertIsNone(timed_out['exit'])
            self.assertFalse(timed_out['throws'])
            self.assertTrue((logs / 'results.json').is_file())

    def test_value_import_used_only_as_type_keeps_its_evaluation_edge(self):
        with tempfile.TemporaryDirectory() as temporary:
            tree = Path(temporary)
            (tree / 'flags.ts').write_text(
                'throw new Error("verbatim import ran"); export const enum Flag { Yes = 1 }\n')
            root = tree / 'main.ts'
            root.write_text('import {Flag} from "./flags.js"; export type OnlyAType = Flag;\n')
            result = run_roots(tree, [root], tree / 'logs')[str(root)]
            self.assertEqual(result['build_exit'], 0)
            self.assertTrue(result['throws'], result)
            self.assertEqual(result['exit'], 1)
            self.assertEqual(result['first_error'], 'Error: verbatim import ran')

    def test_build_failure_cannot_prove_node_throws(self):
        with tempfile.TemporaryDirectory() as temporary:
            tree = Path(temporary)
            root = tree / 'main.ts'
            root.write_text('const = broken;\n')
            result = run_roots(tree, [root], tree / 'logs')[str(root)]
            self.assertNotEqual(result['build_exit'], 0)
            self.assertFalse(result['throws'])
            self.assertIsNone(result['exit'])

    def test_no_candidates_runs_no_commands(self):
        with patch('node_readiness.command') as command:
            self.assertEqual(run_roots(Path('/tree'), [], Path('/logs')), {})
        command.assert_not_called()

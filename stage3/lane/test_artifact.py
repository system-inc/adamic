"""Check artifact input invalidation and rejection of corrupted cached bytes."""
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch
import artifact


class ArtifactTests(unittest.TestCase):
    def setUp(self):
        self.scratch = tempfile.TemporaryDirectory()
        self.addCleanup(self.scratch.cleanup)
        self.root = Path(self.scratch.name)
        (self.root / 'stage3/adapt/00-setup').mkdir(parents=True)
        self.input = self.root / 'stage3/adapt/00-setup/proof.json'
        self.input.write_text('{"proof": 1}\n')
        self.probe = patch.object(artifact.subprocess, 'check_output', return_value='pinned version\n')
        probe = self.probe.start()
        probe.side_effect = lambda command, **kwargs: 'linux/x64\n' if '-p' in command else probe.return_value
        self.addCleanup(self.probe.stop)

    def test_changed_proof_invalidates_input_key(self):
        original = artifact.input_key(self.root)[0]
        self.input.write_text('{"proof": 2}\n')
        self.assertNotEqual(original, artifact.input_key(self.root)[0])

    def test_new_input_invalidates_key(self):
        original = artifact.input_key(self.root)[0]
        (self.input.parent / 'new-proof.json').write_text('{}\n')
        self.assertNotEqual(original, artifact.input_key(self.root)[0])

    def test_tool_version_invalidates_key(self):
        original = artifact.input_key(self.root)[0]
        artifact.subprocess.check_output.return_value = 'different version\n'
        self.assertNotEqual(original, artifact.input_key(self.root)[0])

    def test_corrupted_artifact_is_rejected(self):
        cache = self.root / 'cache'
        key, inputs = artifact.input_key(self.root)
        output = cache / key
        tree = output / 'tree'
        tree.mkdir(parents=True)
        payload = tree / 'compiler.js'
        payload.write_text('original output\n')
        (output / 'ready.json').write_text(json.dumps(dict(
            inputs=inputs, payload=artifact.payload_digest(tree))))
        self.assertTrue(artifact.prepare(cache, self.root)['cache_hit'])
        payload.write_text('wrong output\n')
        with self.assertRaisesRegex(ValueError, 'artifact content'):
            artifact.prepare(cache, self.root)

    def test_failed_preparation_never_publishes(self):
        cache = self.root / 'cache'
        with patch.object(artifact.subprocess, 'run', return_value=type('Result', (), {'returncode': 7})()):
            with self.assertRaisesRegex(RuntimeError, 'apply exit 7'):
                artifact.prepare(cache, self.root)
        self.assertFalse(any(cache.glob('*/ready.json')))

    def test_changing_inputs_during_preparation_is_rejected(self):
        def command(arguments, **kwargs):
            if arguments[0] == 'bash':
                Path(arguments[2]).mkdir()
            else:
                self.input.write_text('changed while building')
            return type('Result', (), {'returncode': 0})()
        with patch.object(artifact.subprocess, 'run', side_effect=command):
            with self.assertRaisesRegex(ValueError, 'inputs changed'):
                artifact.prepare(self.root / 'cache', self.root)
        self.assertFalse(any((self.root / 'cache').glob('*/ready.json')))

    def test_file_boundaries_are_part_of_digest(self):
        first = self.root / 'first'
        second = self.root / 'second'
        first.write_text('a'); second.write_text('bc')
        before = artifact.digest_files(self.root, [first, second])
        first.write_text('ab'); second.write_text('c')
        self.assertNotEqual(before, artifact.digest_files(self.root, [second, first]))


if __name__ == '__main__':
    unittest.main()

#!/usr/bin/env python3
"""Exercise scratch-run.sh with executable stand-ins and independent Node bytes."""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

HERE = Path(__file__).resolve().parent
REPO = HERE.parents[2]


class ScratchEvidenceTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.root = Path(self.temp.name)
        self.node = self.root / 'node.stdout'
        with self.node.open('wb') as output:
            subprocess.run(['node', '-e', 'process.stdout.write("Identifier\\t0\\t1\\nEOF\\t1\\t1\\n")'],
                           stdout=output, check=True)
        self.data = self.node.read_bytes()
        self.sha = subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=REPO, text=True).strip()

    def tearDown(self):
        self.temp.cleanup()

    def command(self, name, data, exit_code=0):
        payload = self.root / (name + '.bytes')
        payload.write_bytes(data)
        binary = self.root / (name + '-native')
        # A real executable process, not a mocked subprocess result.
        binary.write_text('#!/usr/bin/env python3\nfrom pathlib import Path\nimport sys\nsys.stdout.buffer.write(Path(' + repr(str(payload)) + ').read_bytes())\nsys.exit(' + str(exit_code) + ')\n')
        binary.chmod(0o755)
        out = self.root / name
        result = subprocess.run(['bash', HERE / 'scratch-run.sh', out,
                                 '--stand-in-native', binary, '--stand-in-node', self.node,
                                 '--adamic-sha', self.sha, '--source-sha', self.sha],
                                cwd=REPO, capture_output=True, text=True)
        print(' '.join(str(x) for x in [name, 'exit', result.returncode, result.stdout.strip(), result.stderr.strip()] if str(x)))
        return out, result

    def test_matching_binary_writes_complete_evidence_and_kills_native_mutant(self):
        out, result = self.command('matching', self.data)
        self.assertEqual(result.returncode, 0, result.stderr)
        evidence = json.loads((out / 'scanner-native-evidence.json').read_text())
        self.assertTrue(evidence['stand_in'])
        self.assertFalse(evidence['milestone_eligible'])
        block = evidence['evidence']['scanner_native']
        schema = json.loads((REPO / 'stage3/progress.json').read_text())
        for description in schema['evidence_required']:
            fields = description.split(':', 1)[0].split(' and ')
            for field in fields:
                self.assertIn(field, block)
        self.assertEqual(block['adamic_sha'], self.sha)
        self.assertEqual(block['source_sha'], self.sha)
        digest = hashlib.sha256(self.data).hexdigest()
        self.assertEqual(block['node_sha256'], digest)
        self.assertEqual(block['native_sha256'], digest)
        directory = Path(block['run_directory'])
        self.assertEqual((directory / 'node.stdout').read_bytes(), self.data)
        native = (directory / 'native.stdout').read_bytes()
        mutant = (directory / block['mutant']['output']).read_bytes()
        self.assertEqual(len(mutant), len(native))
        self.assertEqual(sum(a != b for a, b in zip(native, mutant)), 1)
        comparison = json.loads((directory / block['mutant']['comparison']).read_text())
        self.assertFalse(comparison['equal'])
        self.assertEqual(comparison['first_difference']['offset'], 0)
        self.assertEqual(block['mutant']['comparison_exit'], 1)
        print('evidence fields complete; one-byte native mutant caught at offset 0')

    def test_one_byte_off_fails_with_exact_first_difference(self):
        data = bytearray(self.data)
        offset = self.data.index(b'EOF') + 1
        data[offset] ^= 1
        out, result = self.command('one-byte-off', data)
        self.assertEqual(result.returncode, 1)
        self.assertFalse((out / 'scanner-native-evidence.json').exists())
        comparison = json.loads((out / 'evidence/comparison.stdout').read_text())
        expected = dict(offset=offset, line=2, column=2,
                        left=self.data[offset], right=data[offset])
        self.assertEqual(comparison['first_difference'], expected)
        self.assertIn(json.dumps(expected), result.stderr)
        print('one-byte stand-in mutant caught:', expected)

    def test_failed_binary_and_empty_output_cannot_claim_milestone(self):
        for name, data, code in [('failed', self.data, 7), ('empty', b'', 0)]:
            if name == 'empty':
                self.node.write_bytes(b'')
            out, result = self.command(name, data, code)
            self.assertEqual(result.returncode, 1)
            self.assertFalse((out / 'scanner-native-evidence.json').exists())


if __name__ == '__main__':
    unittest.main()

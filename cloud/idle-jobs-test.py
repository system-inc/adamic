#!/usr/bin/env python3
"""Durable retry and collection tests without contacting gate boxes."""
import contextlib
import importlib.util
import io
import json
import os
from pathlib import Path
import sys
import tarfile
import tempfile
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location('idle', Path(__file__).with_name('idle-jobs.py'))
idle = importlib.util.module_from_spec(spec)
spec.loader.exec_module(idle)
SHA = 'a' * 40


class IdleTests(unittest.TestCase):
    def test_retry_and_completion(self):
        with tempfile.TemporaryDirectory() as tmp:
            state = Path(tmp)
            (state / 'enabled').touch()
            initial = {SHA: {'next': 2001, 'completed': 0, 'pending': {'home': {'seed': 1, 'count': 2000}}}}
            (state / 'state.json').write_text(json.dumps(initial))
            launches = []
            result = 'unfinished'

            def ssh(box, script, *args):
                if args == ('probe',):
                    return 'idle'
                if args[0] == 'start':
                    launches.append(args)
                    return 'started'
                return result

            with patch.dict(os.environ, {'ADAMIC_IDLE_STATE': tmp, 'ADAMIC_IDLE_BOXES': 'home'}), \
                    patch.object(sys, 'argv', ['idle-jobs.py', '--once']), \
                    patch.object(idle, 'ssh', ssh), patch.object(idle, 'collect'), \
                    patch.object(idle.subprocess, 'check_output', return_value=SHA + '\trefs/heads/main\n'), \
                    contextlib.redirect_stdout(io.StringIO()):
                idle.main()
                self.assertEqual(launches[-1][2:4], ('1', '2000'))
                self.assertEqual(json.loads((state / 'state.json').read_text()), initial)
                result = 'done'
                (state / 'enabled').unlink()
                idle.main()
                saved = json.loads((state / 'state.json').read_text())[SHA]
                self.assertEqual(saved['pending'], {})
                self.assertEqual(saved['by_box']['home'], 2000)
                self.assertEqual(saved['next'], 2001)
                self.assertEqual(len(launches), 1)

    def test_collection_dedupes_bytes_and_signature(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            raw = root / 'raw'
            raw.mkdir()
            for key, box in [('b' * 64, 'home'), ('c' * 64, 'cloud')]:
                folder = raw / key
                folder.mkdir()
                for name, text in {'program.a': 'original', 'reduced.a': 'minimal', 'signature.txt': 'crash:a store\n', 'seed': '1', 'main': SHA, 'box': box}.items():
                    (folder / name).write_text(text)
            archive = root / 'all.tar'
            with tarfile.open(archive, 'w') as tar:
                tar.add(raw, arcname='.')

            def run(*args, **kwargs):
                kwargs['stdout'].write(archive.read_bytes())

            output = io.StringIO()
            with patch.object(idle.subprocess, 'run', run), contextlib.redirect_stdout(output):
                idle.collect('home', root / 'collected')
                idle.collect('cloud', root / 'collected')
            self.assertEqual(len(list((root / 'collected').iterdir())), 1)
            self.assertEqual(len(output.getvalue().splitlines()), 1)
            self.assertIn('crash:a store', output.getvalue())

    def test_full_gate_embeds_same_preemption(self):
        full = Path(__file__).with_name('full-gate-main.sh').read_text()
        start = full.index('# BEGIN idle preemption')
        end = full.index('# END idle preemption') + len('# END idle preemption')
        block = full[start:end].replace('\\\\', '\\').replace('\\$', '$').replace('\\`', '`')
        expected = Path(__file__).with_name('idle-preempt.sh').read_text()
        self.assertEqual(block, expected[expected.index('# BEGIN idle preemption'):].strip())


if __name__ == '__main__':
    unittest.main()

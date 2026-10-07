#!/usr/bin/env python3
"""Run intended failing cache/integrity tests against isolated installer mutants."""
import os
from pathlib import Path
import subprocess
import sys
import tempfile

SOURCE = Path(__file__).resolve().parent
original = (SOURCE / 'setup-node.py').read_text()
output = Path(sys.argv[1])
output.mkdir(parents=True, exist_ok=True)
mutants = {}
for component in ['version', 'checksum', 'system', 'architecture', 'pin', 'helper']:
    mutants['drop-' + component] = (original.replace(f'{component}={component}', f'{component}=None', 1), 'test_every_key_component')
mutants['corrupt-archive'] = (original.replace('if digest(archive.read_bytes()) != expected:', 'if False:'), 'test_corrupted_archive_refused_before_install')
mutants['published-checksum'] = (original.replace('if published != [checksum]:', 'if False:'), 'test_published_checksum_must_match_pin')
mutants['installed-bytes'] = (original.replace("saved['binary'] == digest(destination.read_bytes())", 'True'), 'test_warm_corruption_modes_uncached_and_pin_bytes')
mutants['installed-mode'] = (original.replace('destination.stat().st_mode & 0o777 == 0o755', 'True'), 'test_warm_corruption_modes_uncached_and_pin_bytes')
for name, (mutant, test) in mutants.items():
    assert mutant != original, name
    with tempfile.TemporaryDirectory(prefix='node-mutant-', dir='/tmp/adamic-gate') as temporary:
        path = Path(temporary) / 'setup-node.py'
        path.write_text(mutant)
        with (output / (name + '.log')).open('wb') as log:
            result = subprocess.run([sys.executable, '-m', 'unittest', 'cloud.test_node_setup.NodeSetup.' + test],
                                    cwd=SOURCE.parent, env=dict(os.environ, ADAMIC_NODE_SETUP_MODULE=str(path), PYTHONDONTWRITEBYTECODE='1'),
                                    stdout=log, stderr=subprocess.STDOUT, timeout=60)
        text = (output / (name + '.log')).read_text()
        assert result.returncode != 0 and 'FAIL:' in text and 'ERROR:' not in text, (name, text)
        print(name + ': caught by ' + test, flush=True)

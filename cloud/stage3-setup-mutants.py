#!/usr/bin/env python3
"""Run one omission for each API cache input and retain the intended failures."""
from pathlib import Path
import os
import shutil
import subprocess
import tempfile

source = Path(__file__).resolve().parent
text = (source / 'setup-stage3-api.py').read_text()
scratch = Path(tempfile.mkdtemp(prefix='stage3-mutants-', dir='/tmp/adamic-gate'))
print('mutant logs:', scratch, flush=True)
shutil.copyfile(source / 'setup-markdown-width.py', scratch / 'setup-markdown-width.py')
shutil.copytree(source / 'markdown-width', scratch / 'markdown-width')
variants = {('drop-' + name): text.replace('npm.digest(' + name + ')', "npm.digest(b'ignored')")
            for name in ['lock', 'manifest', 'bootstrap', 'helper']}
variants['drop-node'] = text.replace('npm.digest(helper), node)', "npm.digest(helper), 'ignored')")
variants['drop-tree-bytes'] = text.replace('npm.digest(path.read_bytes())', "'ignored'")
variants['drop-tree-modes'] = text.replace('mode = path.lstat().st_mode & 0o777', 'mode = 0o644')
variants['drop-link-target'] = text.replace('os.readlink(path)', "'ignored'")
for name, variant in variants.items():
    assert variant != text, name
    module = scratch / (name + '.py')
    module.write_text(variant)
    environment = dict(os.environ, ADAMIC_STAGE3_SETUP_MODULE=str(module))
    environment.pop('ADAMIC_SETUP_INTEGRATION', None)
    log = scratch / (name + '.log')
    with log.open('wb') as output:
        result = subprocess.run(['python3', str(source / 'test_stage3_setup.py')], env=environment,
                                stdout=output, stderr=subprocess.STDOUT)
    answer = log.read_text()
    assert result.returncode == 1 and 'AssertionError' in answer and 'FAILED (failures=' in answer, answer
    print(name, 'killed at intended assertion', flush=True)
    if name == 'drop-lock':
        environment['ADAMIC_SETUP_INTEGRATION'] = '1'
        log = scratch / 'real-drop-lock.log'
        with log.open('wb') as output:
            result = subprocess.run(['python3', str(source / 'test_stage3_setup.py'),
                                     'Stage3Integration'], env=environment,
                                    stdout=output, stderr=subprocess.STDOUT)
        answer = log.read_text()
        assert result.returncode == 1 and 'AssertionError' in answer and 'skipped (validated API lock' in answer, answer
        print('drop-lock also killed by real installation', flush=True)
# The existing Go warming stamp independently collects the API inputs.
text = (source / 'setup-key.py').read_text()
for filename in ['cloud/setup-stage3-api.py', 'stage3/api/package.json', 'stage3/api/package-lock.json']:
    anchor = '"' + filename + '"'
    variant = text.replace(anchor + ',', '') if anchor + ',' in text else text.replace(', ' + anchor, '')
    assert variant != text
    name = 'omit-go-' + Path(filename).name
    module = scratch / (name + '.py')
    module.write_text(variant)
    environment = dict(os.environ, ADAMIC_SETUP_KEY_MODULE=str(module))
    environment.pop('ADAMIC_SETUP_INTEGRATION', None)
    log = scratch / (name + '.log')
    with log.open('wb') as output:
        result = subprocess.run(['python3', str(source / 'test_setup.py')], env=environment,
                                stdout=output, stderr=subprocess.STDOUT)
    answer = log.read_text()
    assert result.returncode == 1 and 'test_collected_manifests_invalidate' in answer and 'AssertionError' in answer, answer
    print(name, 'killed by collected manifest assertion', flush=True)

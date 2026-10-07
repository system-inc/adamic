#!/usr/bin/env python3
"""Kill independent omissions from the Markdown installer, retaining each failure log."""
from pathlib import Path
import os
import re
import subprocess
import tempfile

source = Path(__file__).resolve().parent
text = (source / 'setup-markdown-width.py').read_text()
scratch = Path(tempfile.mkdtemp(prefix='markdown-mutants-', dir='/tmp/adamic-gate'))
print('mutant logs:', scratch, flush=True)
variants = {}
for component in ['lock', 'manifest', 'bootstrap', 'helper', 'node']:
    # Remove just this named input from the canonical key; keep every other input.
    anchor = f'{component}={component}'
    assert text.count(anchor) == 1, anchor
    variants['drop-' + component] = re.sub(anchor + r',?\s*', '', text)
variants['drop-tree-bytes'] = text.replace('digest(path.read_bytes())', "'ignored'")
variants['drop-tree-mode'] = text.replace('mode = path.lstat().st_mode & 0o777', 'mode = 0o644')
variants['drop-bootstrap-integrity'] = text.replace("raise ValueError('npm bootstrap integrity mismatch')", 'return')
for name, variant in variants.items():
    assert variant != text, name
    module = scratch / (name + '.py')
    module.write_text(variant)
    environment = dict(os.environ, ADAMIC_MARKDOWN_SETUP_MODULE=str(module))
    environment.pop('ADAMIC_SETUP_INTEGRATION', None)
    log = scratch / (name + '.log')
    with log.open('wb') as output:
        result = subprocess.run(['python3', str(source / 'test_markdown_setup.py')], env=environment,
                                stdout=output, stderr=subprocess.STDOUT)
    answer = log.read_text()
    assert result.returncode == 1 and 'AssertionError' in answer and 'FAILED (failures=' in answer, answer
    print(name, 'killed by intended assertion', flush=True)
    if name == 'drop-lock':
        environment['ADAMIC_SETUP_INTEGRATION'] = '1'
        log = scratch / 'real-drop-lock.log'
        with log.open('wb') as output:
            result = subprocess.run(['python3', str(source / 'test_markdown_setup.py'),
                                     'InstallationIntegration'], env=environment,
                                    stdout=output, stderr=subprocess.STDOUT)
        answer = log.read_text()
        assert result.returncode == 1 and 'AssertionError' in answer and 'skipped (validated lock' in answer, answer
        print('drop-lock also killed by real npm invalidation proof', flush=True)
# The outer Go warming stamp must independently collect all new installation inputs.
key_source = (source / 'setup-key.py').read_text()
for filename in ['cloud/markdown-width/package.json', 'cloud/markdown-width/package-lock.json',
                 'cloud/markdown-width/npm-bootstrap.json', 'cloud/setup-markdown-width.py']:
    anchor = '"' + filename + '"'
    assert key_source.count(anchor) == 1, filename
    variant = re.sub(re.escape(anchor) + r',?\s*', '', key_source)
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

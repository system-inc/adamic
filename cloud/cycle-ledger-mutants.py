#!/usr/bin/env python3
"""Drop each ledger key component and prove the cache checks fail."""
import os
from pathlib import Path
import shutil
import subprocess
import tempfile

source = Path(__file__).resolve().parent
scratch = Path(tempfile.mkdtemp(prefix='cycle-ledger-mutants-', dir='/tmp/adamic-gate'))
for name in ['setup-gate-npm.py', 'node-pin.json', 'setup.sh']:
    shutil.copyfile(source / name, scratch / name)
text = (source / 'setup-gate-inputs.py').read_text()
variants = {'drop-' + name: text.replace('def cache_key(**inputs):',
            'def cache_key(**inputs):\n    inputs.pop(' + repr(name) + ', None)')
            for name in ['kind', 'commit', 'url', 'script', 'diagnostics', 'node', 'helper', 'sparse']}
variants['drop-checkout-status'] = text.replace('def checkout_pristine(directory, generated=()):', 'def checkout_pristine(directory, generated=()):\n    return True')
variants['drop-generated-validation'] = text.replace('for name in LEDGER_GENERATED:', 'for name in []:')
variants['drop-head-validation'] = text.replace('if actual != TS_COMMIT:', 'if False:')
variants['drop-node-pin'] = text.replace('if node_version != expected:', 'if False:')
print('mutant logs:', scratch, flush=True)
for name, variant in variants.items():
    assert variant != text
    module = scratch / (name + '.py')
    module.write_text(variant)
    environment = dict(os.environ, ADAMIC_GATE_INPUTS_MODULE=str(module), PYTHONDONTWRITEBYTECODE='1')
    log = scratch / (name + '.log')
    tests = ['Ledger.test_every_ledger_key_component'] if name.startswith('drop-') and name[5:] in ['commit', 'url', 'script', 'diagnostics', 'node', 'helper', 'sparse'] else ['Ledger.test_missing_generated_file_and_wrong_head_refused']
    if name == 'drop-checkout-status':
        tests = ['Ledger.test_ignored_and_untracked_files_force_pristine_reinstall']
    if name == 'drop-kind':
        command = ['python3', str(source / 'test_gate_inputs.py'), 'Inputs.test_every_key_component']
    else:
        command = ['python3', str(source / 'test_cycle_ledger_setup.py'), *tests]
    with log.open('wb') as output:
        result = subprocess.run(command, env=environment, stdout=output, stderr=subprocess.STDOUT, timeout=60)
    answer = log.read_text()
    assert result.returncode == 1 and 'AssertionError' in answer and 'FAILED (failures=' in answer, answer
    print(name + ': caught by intended assertion', flush=True)

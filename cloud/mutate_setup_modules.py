#!/usr/bin/env python3
"""Require each dropped cache key/check to be caught by the focused tests."""
import os
from pathlib import Path
import subprocess
import tempfile

cloud = Path(__file__).resolve().parent
source = (cloud / 'setup-modules.py').read_text()
mutants = {}
for name in ['manifests', 'version', 'environment', 'helper']:
    token = name + '=' + name
    mutants[name] = source.replace(token, token + ' if False else None', 1)
for name in ['go.mod', 'go.sum', 'go.work', 'go.work.sum']:
    mutants[name] = source.replace('for path in sorted(files)]',
                                  f"for path in sorted(files) if path.name != {name!r}]")
for name in ['st_mode', 'st_size', 'st_mtime_ns', 'st_ctime_ns', 'st_ino']:
    mutants[name] = source.replace(', info.' + name, '', 1)
mutants['artifact-path'] = source.replace('[str(file), info.st_mode', '[info.st_mode', 1)
mutants['changed-definition'] = source.replace('if modules != completed_modules or definitions(inputs) != definitions(completed_inputs):', 'if False:')
logs = Path(tempfile.mkdtemp(prefix='setup-modules-mutants-'))
print('mutant logs:', logs, flush=True)
for name, mutant in mutants.items():
    assert source != mutant, name
    file = logs / (name + '.py')
    file.write_text(mutant)
    environment = dict(os.environ, ADAMIC_MODULES_HELPER=str(file))
    log = logs / (name + '.log')
    with log.open('w') as output:
        result = subprocess.run(['python3', str(cloud / 'test_setup_modules.py')],
                                env=environment, stdout=output, stderr=subprocess.STDOUT)
    assert result.returncode != 0 and 'AssertionError' in log.read_text(), str(log)
    print(name + ': caught by assertion', flush=True)

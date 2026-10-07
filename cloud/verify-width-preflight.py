#!/usr/bin/env python3
"""Check each missing module and kill removal of its early harness check via Go overlays."""
from pathlib import Path
import json
import os
import subprocess
import tempfile

repository = Path(__file__).resolve().parent.parent
source = repository / 'stage1/cohere/markdownblocks/width_test.go'
text = source.read_text()
scratch = Path(tempfile.mkdtemp(prefix='width-preflight-', dir='/tmp/adamic-gate'))
print('preflight logs:', scratch, flush=True)
modules = ['emoji-regex', 'get-east-asian-width', 'narrow-emojis']
for module in modules:
    fixture = scratch / module
    for present in modules:
        if present != module:
            file = fixture / 'node_modules' / present / 'index.js'
            file.parent.mkdir(parents=True, exist_ok=True)
            file.write_text('// file-presence fixture; missing module must fail before this is used\n')
    environment = dict(os.environ, ADAMIC_MARKDOWNWIDTH_DEPS=str(fixture), ADAMIC_GATE_UNCACHED='1')
    command = ['/tmp/markdown-after.test', '-test.run=^TestMarkdownUnicodeWidths$', '-test.timeout=30m', '-test.v']
    with (scratch / (module + '-present-check.log')).open('wb') as output:
        result = subprocess.run(command, cwd=source.parent, env=environment, stdout=output, stderr=subprocess.STDOUT)
    expected = 'missing Markdown width oracle module ' + module
    assert result.returncode == 1 and expected in (scratch / (module + '-present-check.log')).read_text()
    quoted = '"' + module + '"'
    variant = text.replace(quoted + ', ', '') if quoted + ', ' in text else text.replace(', ' + quoted, '')
    assert variant != text
    replacement = scratch / (module + '.go')
    replacement.write_text(variant)
    overlay = scratch / (module + '.json')
    overlay.write_text(json.dumps({'Replace': {str(source): str(replacement)}}))
    binary = scratch / (module + '.test')
    with (scratch / (module + '-build.log')).open('wb') as output:
        result = subprocess.run(['go', 'test', '-overlay=' + str(overlay), '-c', '-o', str(binary),
                                 './stage1/cohere/markdownblocks'], cwd=repository,
                                stdout=output, stderr=subprocess.STDOUT)
    assert result.returncode == 0, (scratch / (module + '-build.log')).read_text()
    log = scratch / (module + '-mutant.log')
    with log.open('wb') as output:
        result = subprocess.run([str(binary)] + command[1:], cwd=source.parent, env=environment,
                                stdout=output, stderr=subprocess.STDOUT)
    answer = log.read_text()
    assert result.returncode == 1 and expected not in answer and 'ERR_MODULE_NOT_FOUND' in answer, answer
    print(module, 'guard omission caught by loss of required early diagnostic', flush=True)

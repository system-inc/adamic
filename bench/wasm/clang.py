#!/usr/bin/env python3
"""Benchmark-only SDK compiler wrapper; also used by the unchanged WASI oracle."""
import json
import os
import subprocess
import sys
from pathlib import Path

configuration = json.loads((Path(sys.argv[0]).resolve().parent / 'configuration.json').read_text())
arguments = sys.argv[1:]
if arguments == ['--version']:
    subprocess.run([configuration['clang'], '--version'], check=True)
    print('wasm-size configuration: ' + json.dumps(configuration, sort_keys=True))
    sys.exit(0)
arguments = [configuration['optimization'] if value == '-O2' else value for value in arguments]
if configuration.get('drop_archive'):
    for value in ['--whole-archive', '--no-whole-archive']:
        if value in arguments:
            index = arguments.index(value)
            assert arguments[index - 1] == '-Xlinker'
            del arguments[index - 1:index + 1]
arguments += configuration.get('compile_flags', [])
linking = '-c' not in arguments
if linking:
    arguments += configuration.get('link_flags', [])
record = {'command': [configuration['clang'], *arguments]}
with open(configuration['commands'], 'a') as log:
    log.write(json.dumps(record) + '\n')
code = subprocess.run(record['command']).returncode
if code == 0 and linking and configuration.get('llvm_strip'):
    output = arguments[arguments.index('-o') + 1]
    code = subprocess.run([configuration['llvm_strip'], output]).returncode
if code == 0 and linking and configuration.get('wasm_opt'):
    output = arguments[arguments.index('-o') + 1]
    code = subprocess.run([configuration['wasm_opt'], '-Oz', output, '-o', output]).returncode
sys.exit(code)

#!/usr/bin/env python3
"""Planted emitted-code mutant for the admission gate, never a delivery compiler."""
import os
import re
import subprocess
import sys

compiler = os.environ['GENERIC_VALUES_HEAD_COMPILER']
args = sys.argv[1:]
if len(args) == 2 and args[0] == 'js' and args[1].endswith('/01_comparer.a'):
    result = subprocess.run([compiler, *args], capture_output=True)
    if result.returncode:
        sys.stdout.buffer.write(result.stdout)
        sys.stderr.buffer.write(result.stderr)
        sys.exit(result.returncode)
    emitted = result.stdout.decode()
    pattern = r'(function function_\d+_equateValues_\d+\([^)]*\) \{\n)\treturn adamicEqual\([^;\n]*\);'
    changed, count = re.subn(pattern, r'\1\treturn false;', emitted)
    if count != 2:
        raise RuntimeError('comparer mutant did not match two concrete instances')
    sys.stdout.write(changed)
    sys.stderr.buffer.write(result.stderr)
    sys.exit(0)
os.execv(compiler, [compiler, *args])

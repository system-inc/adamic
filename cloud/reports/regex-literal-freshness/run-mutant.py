#!/usr/bin/env python3
"""Hoist one retained RegExp object per compiled program, then restore the runtime."""
import os
import pathlib
import subprocess

root = pathlib.Path(__file__).resolve().parents[3]
target = root / 'internal/native/runtime/regexp.c'
original = target.read_text()
site = '\tadamic_object *result = adamic_object_new(&regex_shape);'
cache = '''\tstatic const adamic_regex_program *keys[64];
\tstatic adamic_object *objects[64];
\tstatic size_t count;
\tfor (size_t k = 0; k < count; k++)
\t\tif (keys[k] == program) return adamic_retain(objects[k]);
\tadamic_object *result = adamic_object_new(&regex_shape);'''
finish = '\tresult->slots[7].boolean = (program->flags & 4) != 0 && !(program->flags & 64);\n\treturn result;'
remember = '''\tresult->slots[7].boolean = (program->flags & 4) != 0 && !(program->flags & 64);
\tkeys[count] = program;
\tobjects[count++] = adamic_retain(result);
\treturn result;'''
assert original.count(site) == original.count(finish) == 1
log = pathlib.Path('/tmp/regex-literal-freshness-mutant.log')
try:
    target.write_text(original.replace(site, cache).replace(finish, remember))
    with log.open('w') as output:
        result = subprocess.run(['go', 'test', './internal/oracle', '-run', 'TestNativeAgreesWithNode/internal/oracle/testdata/regexp_literal_freshness', '-count=1', '-v', '-timeout', '15m'], cwd=root, env=dict(os.environ, GOFLAGS='-buildvcs=false'), stdout=output, stderr=subprocess.STDOUT)
    if result.returncode == 0 or 'stdout differs' not in log.read_text():
        raise SystemExit(f'Hoisting mutant NOT CAUGHT: exit={result.returncode}; {log}')
    print(f'Hoisting mutant caught by Node stdout comparison: exit={result.returncode}; {log}')
finally:
    target.write_text(original)

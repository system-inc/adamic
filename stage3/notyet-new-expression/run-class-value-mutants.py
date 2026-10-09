#!/usr/bin/env python3
"""Run constructor dispatch/order mutants and restore the lowering helper."""
from pathlib import Path
import os
import subprocess

root = Path(__file__).resolve().parents[2]
path = root / 'internal/lower/new_class_value.go'
logs = Path('/tmp/new-expression-class-mutants')
logs.mkdir(exist_ok=True)
mutants = [
    ('cache-always-initialize',
     'return ir.Coalesce{Value: left, Fallback: right, Of: ir.Object}, nil',
     '_ = left; return right, nil'),
    ('cache-do-not-store',
     'b.body = statements', 'b.body = statements[:0]'),
    ('static-allocator',
     'Condition: ir.InstanceOf{Value: receiver, Class: candidate.identity, Exact: true}',
     'Condition: ir.Binary{Operator: ir.Or, Left: ir.BooleanConstant{Value: true}, Right: ir.InstanceOf{Value: receiver, Class: candidate.identity, Exact: true}}'),
    ('reread-after-arguments',
     'receiver := b.read(b.parameters[0])',
     'receiver := b.read(b.parameters[0])\n'
     '\tif read, ok := constructor.(ir.Read); ok && l.result.Locals[read.Local].Global { receiver = constructor }'),
    ('repeat-constructor',
     'receiver := b.read(b.parameters[0])',
     'receiver := b.read(b.parameters[0])\n'
     '\tif _, call := constructor.(ir.Call); call { b.body = append(b.body, ir.Evaluate{Value: constructor}) }'),
    ('repeat-argument',
     'receiver := b.read(b.parameters[0])',
     'receiver := b.read(b.parameters[0])\n'
     '\tif len(arguments) > 1 { if _, call := arguments[1].(ir.Call); call { b.body = append(b.body, ir.Evaluate{Value: arguments[1]}) } }'),
]
command = ['go', 'test', './internal/oracle', '-run',
           'TestNativeAgreesWithNode/internal/oracle/testdata/new_expression_class_value',
           '-count=1', '-timeout', '10m']
env = dict(os.environ, ADAMIC_GATE_UNCACHED='1')
for name, before, after in mutants:
    original = path.read_bytes()
    source = original.decode()
    assert source.count(before) == 1, (name, 'mutant anchor changed')
    try:
        path.write_text(source.replace(before, after))
        with (logs / (name + '.log')).open('w') as output:
            result = subprocess.run(command, cwd=root, env=env, stdout=output, stderr=subprocess.STDOUT)
        log = (logs / (name + '.log')).read_text()
        assert result.returncode != 0, (name, 'survived')
        if name == 'cache-do-not-store':
            assert 'exit codes differ' in log and 'a constructor value without a registered class allocator' in log, (name, 'not killed by constructor trap', log)
        else:
            assert 'stdout differs' in log, (name, 'not killed by output comparison', log)
        assert 'clang failed' not in log, (name, 'invalid build kill', log)
        catcher = 'registered allocator trap and Node exit comparison' if name == 'cache-do-not-store' else 'Node stdout comparison'
        print(name + ': killed by ' + catcher, flush=True)
    finally:
        path.write_bytes(original)

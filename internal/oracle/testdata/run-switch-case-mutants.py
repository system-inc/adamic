#!/usr/bin/env python3
"""Run independent switch declaration mutants; always restore the compiler."""
from pathlib import Path
import os
import subprocess

repository = Path(__file__).resolve().parents[3]
logs = Path('/tmp/adamic-switch-case-mutants')
logs.mkdir(exist_ok=True)
mutants = [
    ('native-read', 'internal/native/emit_locals.go', 'if read.Checked {', 'if read.Checked && e.program.Locals[read.Local].Ready == 0 {', 'switch_case_read_tdz'),
    ('native-write', 'internal/native/emit_statements.go', 'if statement.Checked {', 'if statement.Checked && e.program.Locals[statement.Local].Ready == 0 {', 'switch_case_write_tdz'),
    ('javascript-read', 'internal/javascript/javascript.go', 'if expression.Checked {', 'if expression.Checked && e.program.Locals[expression.Local].Ready == 0 {', 'switch_case_read_tdz'),
    ('javascript-write', 'internal/javascript/javascript.go', 'if statement.Checked {', 'if statement.Checked && e.program.Locals[statement.Local].Ready == 0 {', 'switch_case_write_tdz'),
    ('early-ready', 'internal/lower/switch_declarations.go', 'ir.Declare{Local: ready, Value: ir.BooleanConstant{}}', 'ir.Declare{Local: ready, Value: ir.BooleanConstant{Value: true}}', 'switch_case_capture_tdz'),
    ('entry-ready', 'internal/lower/switch_declarations.go', 'ir.Declare{Local: ready, Value: ir.BooleanConstant{}}', 'ir.Declare{Local: ready, Value: ir.BooleanConstant{Value: true}}', 'switch_case_reentry_tdz'),
    ('never-ready', 'internal/lower/switch_declarations.go', 'ir.Assign{Local: ready - 1, Value: ir.BooleanConstant{Value: true}}', 'ir.Assign{Local: ready - 1, Value: ir.BooleanConstant{}}', 'switch_case_scope'),
    ('initializer-twice', 'internal/lower/switch_declarations.go', 'ir.Assign{Local: declared.Local, Value: declared.Value})', 'ir.Assign{Local: declared.Local, Value: declared.Value}, ir.Assign{Local: declared.Local, Value: declared.Value})', 'switch_case_scope'),
    ('oracle-transform', 'oracle/node.mjs', "mode: 'strip'", "mode: 'transform'", 'switch_case_scope'),
]
for name, relative, before, after, fixture in mutants:
    path = repository / relative
    original = path.read_text()
    assert before in original, name
    try:
        path.write_text(original.replace(before, after, 1))
        command = ['go', 'test', './internal/oracle', '-run', 'TestNativeAgreesWithNode/internal/oracle/testdata/' + fixture, '-count=1', '-v', '-timeout', '10m']
        with (logs / (name + '.log')).open('w') as output:
            result = subprocess.run(command, cwd=repository, env={**os.environ, 'ADAMIC_GATE_UNCACHED': '1'}, stdout=output, stderr=subprocess.STDOUT)
        log = (logs / (name + '.log')).read_text()
        assert result.returncode != 0, name + ' survived'
        assert 'exit codes differ' in log or 'stdout differs' in log or 'stderr differs' in log, name + ' did not fail an oracle comparison'
        assert 'build failed' not in log and 'Lower:' not in log and 'clang:' not in log, name + ' failed outside the oracle'
        print(name + ': caught by Node comparison; log ' + str(logs / (name + '.log')), flush=True)
    finally:
        path.write_text(original)

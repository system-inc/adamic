#!/usr/bin/env python3
"""Run independent lowering mutants, restoring every source before continuing."""
import os
from pathlib import Path
import subprocess
import sys

root = Path(__file__).resolve().parents[3]
logs = Path("/tmp/statements-small-mutants")
logs.mkdir(exist_ok=True)
mutants = [
    ("prefix-write", "internal/lower/expression.go", "l.result.Functions[index].Body = append(body, ir.Return{Value: value})", "l.result.Functions[index].Body = []ir.Statement{ir.Return{Value: value}}; _ = body", "TestNativeAgreesWithNode/internal/oracle/testdata/statements_small_prefix", "stdout"),
    ("prefix-result", "internal/lower/expression.go", "l.result.Functions[index].Body = append(body, ir.Return{Value: value})", "l.result.Functions[index].Body = append(body, ir.Return{Value: ir.Binary{Operator: ir.Subtract, Left: value, Right: ir.NumberConstant{Value: 1}}})", "TestNativeAgreesWithNode/internal/oracle/testdata/statements_small_prefix", "stdout"),
]
mutants.extend([
    ("nonnull-check", "internal/lower/assignments.go", "current, err = l.nonNullValue(assertion, current)", "current, err = l.nonNullValue(assertion, current); current = ir.NumberConstant{Value: 0}", "TestStatementsSmallTypeScriptIncrement/missing", "exit codes differ"),
    ("nonnull-write", "internal/lower/assignments.go", "Value: fit(updated, of), Class: l.classOf(target)", "Value: fit(updated.Left, of), Class: l.classOf(target)", "TestStatementsSmallTypeScriptIncrement/nonnull", "stdout differs"),
    ("nonnull-receiver", "internal/lower/assignments.go", "ir.Declare{Local: held, Value: object},", "ir.Evaluate{Value: object}, ir.Declare{Local: held, Value: object},", "TestStatementsSmallTypeScriptIncrement/nonnull", "stdout differs"),
    ("throw-identity", "internal/lower/exceptions.go", "return []ir.Statement{ir.Throw{Value: value}}, nil", 'return []ir.Statement{ir.Throw{Value: ir.MakeError{Message: ir.Property{Object: value, Name: "message", Of: ir.String}}}}, nil', "TestNativeAgreesWithNode/internal/oracle/testdata/statements_small_throw", "stdout"),
    ("error-provenance", "internal/lower/exceptions.go", "&& madeError(thrown)", "&& true", "TestStatementsSmallRulings/structural_error", "stdout differs"),
    ("parameter-order", "internal/lower/functions.go", "(defaultIndex == len(defaults) || patterns[patternIndex].incoming < defaults[defaultIndex].incoming)", "true", "TestNativeAgreesWithNode/internal/oracle/testdata/statements_small_parameters", "stdout"),
    ("template-object", "internal/lower/expression.go", 'return nil, l.notYet(span, "a template interpolating an object, an array, a map, a function or undefined")', 'value = ir.StringConstant{Index: l.constant("[object Object]")}', "TestStatementsSmallRulings/template", "want preserved stop"),
    ("nonnull-refusal", "internal/lower/refusals.go", 'ast.KindNonNullExpression: {"the non-null assertion !", "write ?? panic(\'why it can\'t be missing\'), or narrow and handle the missing case"},', '', "TestStatementsSmallRulings/nonnull", "want preserved stop"),
    ("initializing-capture", "internal/lower/locals.go", 'l.unlowerable = l.notYet(name, "a function value that captures the variable its own initializer declares")', '_ = name', "TestStatementsSmallRulings/capture", "want preserved stop"),
])
if len(sys.argv) > 1:
    mutants = [mutant for mutant in mutants if mutant[0] in sys.argv[1:]]
for name, relative, original, replacement, pattern, catcher in mutants:
    path = root / relative
    clean = path.read_text()
    assert clean.count(original) == 1, name
    try:
        path.write_text(clean.replace(original, replacement))
        with (logs / (name + ".log")).open("w") as log:
            result = subprocess.run(["go", "test", "./internal/oracle", "-run", pattern, "-count=1", "-timeout", "10m"], cwd=root, env=dict(os.environ, ADAMIC_GATE_UNCACHED="1"), stdout=log, stderr=subprocess.STDOUT)
        output = (logs / (name + ".log")).read_text()
        assert result.returncode != 0 and catcher in output and "FAIL" in output, name + ": intended assertion did not catch it"
        print(name + ": caught by " + catcher)
    finally:
        path.write_text(clean)

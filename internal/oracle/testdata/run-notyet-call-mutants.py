#!/usr/bin/env python3
"""Run independent caller mutants; restore source even when a check fails."""
import os
from pathlib import Path
import subprocess

repository = Path(__file__).resolve().parents[3]
source = repository / "internal/lower/expression.go"
original = source.read_text()
helper = repository / "internal/lower/optional_void.go"
original_helper = helper.read_text()
native = repository / "internal/native/emit_objects.go"
original_native = native.read_text()
logs = Path("/tmp/notyet-call-mutants")
logs.mkdir(exist_ok=True)
void_fixtures = "TestNativeAgreesWithNode/internal/oracle/testdata/notyet_void_undefined_"
cases = [
    ("reject-void-union", " && !optionalVoidResult(result)", "", void_fixtures, "Lower:"),
    ("drop-void-effects", "return ir.CallClosure{Closure: closure, Arguments: arguments, Returns: returns}, nil",
     "if optionalVoidResult(l.checker.GetTypeAtLocation(node)) { return ir.Undefined{}, nil }; return ir.CallClosure{Closure: closure, Arguments: arguments, Returns: returns}, nil",
     void_fixtures, "stdout differs"),
    ("accept-void-value", 'return nil, l.notYet(node, "a void call used as a value")',
     "return ir.Undefined{}, nil", "TestOptionalVoidValueStillNotYet", "want a named value-use stop"),
    ("reject-union-argument", "if censusCallableSlotless(argument.Type()) && !boxedParameters[index] {",
     "if _ = index; censusCallableSlotless(argument.Type()) {",
     "TestNativeAgreesWithNode/internal/oracle/testdata/notyet_union_callable_adapter", "Lower:"),
    ("omit-union-box", "arguments[index] = fit(arguments[index], takes)",
     "if takes != ir.Union { arguments[index] = fit(arguments[index], takes) }",
     "TestNativeAgreesWithNode/internal/oracle/testdata/notyet_union_callable_adapter_location", "exit codes differ"),
    ("nullable-signature", "l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node.AsCallExpression().Expression))",
     "l.checker.GetTypeAtLocation(node.AsCallExpression().Expression)",
     "TestNativeAgreesWithNode/internal/oracle/testdata/notyet_union_callable_adapter", "Lower:"),
    ("ignore-generic-substitution", "if returns, isKnown = l.representation(result); !isKnown {",
     "if returns, isKnown = l.representation(result); !isKnown || result.Flags()&checker.TypeFlagsTypeParameter != 0 {",
     "TestNativeAgreesWithNode/internal/oracle/testdata/notyet_callable_type_parameter", "a call returning T"),
]
try:
    for name, before, after, selector, catcher in cases:
        if before not in original:
            raise RuntimeError(f"{name}: mutation target absent")
        source.write_text(original.replace(before, after))
        log = logs / (name + ".log")
        with log.open("w") as output:
            result = subprocess.run(
                ["go", "test", "./internal/oracle", "-run", selector, "-count=1", "-v"],
                cwd=repository, stdout=output, stderr=subprocess.STDOUT,
                env=dict(os.environ, ADAMIC_GATE_UNCACHED="1"),
            )
        observation = log.read_text()
        if result.returncode == 0 or catcher not in observation or "[build failed]" in observation:
            raise RuntimeError(f"{name}: intended catcher absent; see {log}")
        print(f"{name}: exit {result.returncode}, caught by {catcher}; {log}", flush=True)
    source.write_text(original)
    for name, before, after, selector, catcher in [
        ("accept-mixed-void", "if member.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsUndefined) == 0 {", "if false {",
         "TestNativeAgreesWithNode/internal/oracle/testdata/notyet_void_undefined_mixed", "want stage 0 to refuse"),
    ]:
        if before not in original_helper:
            raise RuntimeError(f"{name}: mutation target absent")
        helper.write_text(original_helper.replace(before, after))
        log = logs / (name + ".log")
        with log.open("w") as output:
            result = subprocess.run(
                ["go", "test", "./internal/oracle", "-run", selector, "-count=1", "-v"],
                cwd=repository, stdout=output, stderr=subprocess.STDOUT,
                env=dict(os.environ, ADAMIC_GATE_UNCACHED="1"),
            )
        observation = log.read_text()
        if result.returncode == 0 or catcher not in observation or "[build failed]" in observation:
            raise RuntimeError(f"{name}: intended catcher absent; see {log}")
        print(f"{name}: exit {result.returncode}, caught by {catcher}; {log}", flush=True)
    helper.write_text(original_helper)
    before = "func (e *emitter) dispatchable(function int) bool {"
    after = before + "\n for index, parameter := range e.program.Functions[function].Parameters { if index > 0 && e.program.Locals[parameter].Type == ir.Union { return false } }"
    native.write_text(original_native.replace(before, after))
    log = logs / "omit-union-adapter.log"
    with log.open("w") as output:
        result = subprocess.run(
            ["go", "test", "./internal/oracle", "-run", "TestNativeAgreesWithNode/internal/oracle/testdata/notyet_union_callable_adapter", "-count=1", "-v"],
            cwd=repository, stdout=output, stderr=subprocess.STDOUT,
            env=dict(os.environ, ADAMIC_GATE_UNCACHED="1"),
        )
    observation = log.read_text()
    if result.returncode == 0 or "a method the checker proved is there is missing" not in observation or "[build failed]" in observation:
        raise RuntimeError(f"omit-union-adapter: intended catcher absent; see {log}")
    print(f"omit-union-adapter: exit {result.returncode}, caught by native/Node disagreement; {log}", flush=True)
finally:
    source.write_text(original)
    helper.write_text(original_helper)
    native.write_text(original_native)


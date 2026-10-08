"""Run semantic mutants against generic-value boundaries and Node fixtures."""
from pathlib import Path
import os
import subprocess
import tempfile

os.chdir(Path(__file__).resolve().parents[3])
logs = Path(tempfile.mkdtemp(prefix="generic-function-value-mutants-"))
mutants = [
    ("identity-cache", "internal/lower/generic_function_value.go",
     "\tif existing, known := l.genericInstances[key]; known {\n\t\treturn l.functionValue(node, existing)\n\t}\n", "", "oracle"),
    ("multiple-value-specializations", "internal/lower/generic_function_value.go",
     "existingKey != key {", "false {", "lower"),
    ("parameter-proof", "internal/lower/generic_function_value.go",
     "if !identicalTypes(l.checker, from, to) || !l.sameKeeping(from, to, map[[2]*checker.Type]bool{}) {",
     "if !l.sameKeeping(from, to, map[[2]*checker.Type]bool{}) {", "lower"),
    ("result-proof", "internal/lower/generic_function_value.go",
     "if !identicalTypes(l.checker, result, givenResult) && !stringResult {",
     "if !stringResult && result == givenResult && false {", "lower"),
    ("extra-context-parameters", "internal/lower/generic_function_value.go",
     "len(target.Parameters()) > len(given.Parameters())",
     "len(target.Parameters()) != len(given.Parameters())", "oracle"),
    ("phantom-binder", "internal/lower/generic_function_value.go",
     "concrete := concreteTypes[binder]",
     "concrete := concreteTypes[binder]\n\t\tif concrete == nil && len(targets) > 0 { concrete = targets[0] }", "lower"),
    ("string-result-widening", "internal/lower/generic_function_value.go",
     "stringResult := l.withoutUndefined(givenResult).Flags()&checker.TypeFlagsString != 0 && resultKnown && resultKind == ir.String",
     "stringResult := l.withoutUndefined(givenResult).Flags()&checker.TypeFlagsString != 0 && resultKnown && resultKind == ir.String && false", "oracle"),
    ("old-generic-refusal", "internal/lower/expression.go",
     "return l.genericFunctionValue(node, declaration)",
     '_ = declaration; return nil, l.notYet(node, "a generic function as a value")', "oracle"),
]
for name, file, old, new, package in mutants:
    path = Path(file)
    original = path.read_text()
    assert old in original, name
    try:
        path.write_text(original.replace(old, new, 1))
        target = "TestGenericFunctionValueBoundaries" if package == "lower" else "TestNativeAgreesWithNode/internal/oracle/testdata/generic_function_(value|census)"
        log = logs / (name + ".log")
        with log.open("w") as output:
            result = subprocess.run(["go", "test", "./internal/" + package, "-run", target, "-count=1", "-timeout", "10m"], stdout=output, stderr=subprocess.STDOUT)
        text = log.read_text()
        assert result.returncode != 0, name + " survived"
        assert "build failed" not in text and "panic:" not in text, name + " invalid kill"
        print(name, "caught", log, flush=True)
    finally:
        path.write_text(original)

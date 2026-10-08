"""Mutate only source keys and identity guarding; restore each file on exit."""
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
mutants = [
    ("reuse-one-instance", "internal/lower/generic.go", 'key += "," + l.genericTypeKey(concrete)', 'key += ",shared"', "TestGenericFunctionValueDifferentReferenceRepresentations", "got 1 instances for two reference representations"),
    ("erase-reference-identity", "internal/lower/generic.go", 'key += "," + l.genericTypeKey(concrete)', 'key += "," + typeName(held)', "TestGenericFunctionValueReferenceInstances", "got 1 instances for two reference types"),
    ("miss-earlier-comparison", "internal/lower/expression.go", "if l.hasGenericFunctionValues() {", "if len(l.forwarders) > 0 && l.hasGenericFunctionValues() {", "TestGenericFunctionValueEarlierIdentityComparisonIsNotYet", "want earlier identity comparison NotYet"),
    ("observe-specialized-pointers", "internal/lower/generic_function_value.go", "if observes && l.hasGenericFunctionValues() {", "if false && observes && l.hasGenericFunctionValues() {", "TestGenericFunctionValueIdentityCallsAreNotYet", "want identity observation NotYet"),
    ("deduplicate-specialized-pointers", "internal/lower/generic_function_value.go", "if len(arguments) > 0 && l.functionIdentityType(arguments[0]) && l.hasGenericFunctionValues() {", "if false && len(arguments) > 0 && l.functionIdentityType(arguments[0]) && l.hasGenericFunctionValues() {", "TestGenericFunctionValueIdentityCallsAreNotYet", "want identity observation NotYet"),
    ("compare-boxed-functions", "internal/lower/expression.go", "if l.functionIdentityType(l.checker.GetTypeAtLocation(node.AsBinaryExpression().Left)) || l.functionIdentityType(l.checker.GetTypeAtLocation(node.AsBinaryExpression().Right)) {", "if both(ir.Closure) {", "TestGenericFunctionValueUnionIdentityIsNotYet", "want union identity comparison NotYet"),
    ("compare-specialized-pointers", "internal/lower/expression.go", "if l.hasGenericFunctionValues() {", "if false && l.hasGenericFunctionValues() {", "TestGenericFunctionValueIdentityComparisonIsNotYet", "want identity comparison NotYet"),
]
for name, relative, old, new, test, catcher in mutants:
    source = root / relative
    original = source.read_text()
    assert original.count(old) == 1, (name, original.count(old))
    try:
        source.write_text(original.replace(old, new))
        log = Path("/tmp") / ("generic-function-value-" + name + ".log")
        with log.open("w") as output:
            result = subprocess.run(["go", "test", "./internal/lower", "-run", "^" + test + "$", "-count=1"], cwd=root, stdout=output, stderr=subprocess.STDOUT)
        assert result.returncode != 0 and catcher in log.read_text(), (name, log.read_text())
        print(name + ": caught by " + test)
    finally:
        source.write_text(original)

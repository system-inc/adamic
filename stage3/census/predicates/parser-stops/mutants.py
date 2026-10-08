#!/usr/bin/env python3
"""Erase each independent proof and require its pinned negative control to fail."""
from pathlib import Path
import subprocess
import sys

root = Path(__file__).resolve().parents[4]
evidence = Path(__file__).resolve().parent / "evidence"
mutants = {
    "callback": ("internal/lower/predicates_proof.go", "func (l *lowering) proveInferredPredicate(function *ast.Node, target *checker.Type) bool {", "\n\treturn true\n", "TestParserCallbackLie", False),
    "assertion": ("internal/lower/predicates.go", "func (p *predicateFlowProof) normalReturn(node *ast.Node, path predicateFlowPath) error {", "\n\treturn nil\n", "TestParserAssertionProofErasure", False),
    "failure-helper": ("internal/lower/predicates.go", "if verifier.neverCall(expression, map[*ast.Node]bool{}) {", "if true || verifier.neverCall(expression, map[*ast.Node]bool{}) {", "TestParserAssertionProofErasure", True),
    "specialized-null": ("internal/lower/expression.go", "AlwaysFalse: !value.Type().IsReference() || !l.includesNull(l.concrete(l.checker.GetTypeAtLocation(operand)))", "AlwaysFalse: !value.Type().IsReference() || !l.includesNull(l.checker.GetTypeAtLocation(operand))", "TestParserAssertionBackends", True),
    "specialized-undefined": ("internal/lower/expression.go", "if l.includesNull(l.concrete(l.checker.GetTypeAtLocation(operand))) {", "if l.includesNull(l.checker.GetTypeAtLocation(operand)) {", "TestParserAssertionBackends", True),
    "scalar-evaluation": ("internal/lower/expression.go", "\t\t\toperand := node.AsBinaryExpression().Left\n\t\t\tif leftNull {", "\t\t\tif !value.Type().IsReference() { return ir.BooleanConstant{Value: operator == ast.KindExclamationEqualsEqualsToken}, nil }\n\t\t\toperand := node.AsBinaryExpression().Left\n\t\t\tif leftNull {", "TestParserAssertionBackends", True),
}
for name in sys.argv[1:] or mutants:
    relative, signature, replacement, test, replace = mutants[name]
    path = root / relative
    original = path.read_text()
    assert original.count(signature) == 1
    start = original.index(signature) + len(signature)
    # Insert an unconditional return. Go permits unreachable statements, so this
    # removes only the verifier, with no build-warning failure as a substitute.
    try:
        path.write_text(original.replace(signature, replacement, 1) if replace else original[:start] + replacement + original[start:])
        with (evidence / ("mutant-" + name + ".log")).open("w") as output:
            result = subprocess.run(["go", "test", "./internal/lower", "-run", "^" + test + "$", "-count=1", "-timeout", "10m"], cwd=root, stdout=output, stderr=subprocess.STDOUT)
        text = (evidence / ("mutant-" + name + ".log")).read_text()
        assert result.returncode != 0 and "--- FAIL: " + test in text and "[build failed]" not in text, text
        if test == "TestParserAssertionBackends" and name != "specialized-undefined":
            assert "--- FAIL: " + test + "/native" in text and "--- FAIL: " + test + "/javascript" in text, text
        if name == "specialized-undefined":
            assert "--- FAIL: " + test + "/native" in text, text
        print(name + ": caught by " + test)
    finally:
        path.write_text(original)

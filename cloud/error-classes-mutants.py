#!/usr/bin/env python3
"""Run the error unit's 21 source-check mutants without changing the working tree.

Source /workspace/adamic-tools/env.sh first. Each mutant must build, reach its
assertion, and fail there. A compilation failure never counts as a kill.
"""
import argparse
import json
from pathlib import Path
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[1]
LOWER = "./internal/lower"
PROBE = "TestErrorBoundariesAreExplicit/"
MUTANTS = [
 ("mayThrow-overwritten", [("internal/lower/exceptions.go", 'found = found || node.Checked', 'found = node.Checked')], "./internal/oracle", "TestNativeAgreesWithNode/internal/oracle/testdata/error_checks.a$", "differs"),
 ("derived-error-descriptors", [("internal/lower/library_object.go", 'name == "hasOwn" && l.errorType(l.checker.GetTypeAtLocation(written[0]))', 'false && name == "hasOwn" && l.errorType(l.checker.GetTypeAtLocation(written[0]))'), ("internal/lower/prototype.go", 'name != "valueOf" && l.errorType(l.checker.GetTypeAtLocation(receiver))', 'name != "valueOf" && l.isLibraryType(l.checker.GetTypeAtLocation(receiver), "Error")'), ("internal/lower/prototype.go", 'name != "valueOf" && l.errorType(shape)', 'name != "valueOf" && l.isLibraryType(shape, "Error")')], LOWER, PROBE+"derived_error_(hasOwn|enumerable|locale|static_own)$", "want refusal"),
 ("ready-read-flow-edge", [("internal/flow/build.go", 'if value.Interface().(ir.Read).Checked {', 'if false && value.Interface().(ir.Read).Checked {')], "./internal/flow", "TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/error_checks.a$", "the call ended"),
 ("ready-write-flow-edge", [("internal/flow/build.go", 'ok && statement.Checked && instruction.Part == 0', 'ok && false && statement.Checked && instruction.Part == 0')], "./internal/flow", "TestEveryPathNodeTakesIsInTheGraph/../oracle/testdata/error_checks.a$", "the call ended"),
 ("inherited-constructor-field", [("internal/lower/object.go", 'if l.inheritedLibraryMember(node)', 'if false && l.inheritedLibraryMember(node)'), ("internal/lower/object.go", 'name == "constructor" && field != nil', 'false && name == "constructor" && field != nil')], LOWER, PROBE+"inherited_constructor_value$", "want refusal"),
 ("nullable-generic-receiver", [("internal/lower/error_classes.go", 'if value.Type() == ir.Object && l.includesUndefined(l.checker.GetTypeAtLocation(args[0])) {', 'if false && value.Type() == ir.Object && l.includesUndefined(l.checker.GetTypeAtLocation(args[0])) {')], LOWER, PROBE+"nullable_generic_receiver$", "want refusal"),
 ("stack-read", [("internal/lower/object.go", 'if l.errorType(l.checker.GetTypeAtLocation(access.Expression)) && name == "stack" {', 'if false && l.errorType(l.checker.GetTypeAtLocation(access.Expression)) && name == "stack" {')], LOWER, PROBE+"stack_read$", "want refusal"),
 ("stack-write", [("internal/lower/class.go", 'if target.Name().Text() == "stack" && l.errorType(l.checker.GetTypeAtLocation(target.AsPropertyAccessExpression().Expression)) {', 'if false && target.Name().Text() == "stack" && l.errorType(l.checker.GetTypeAtLocation(target.AsPropertyAccessExpression().Expression)) {')], LOWER, PROBE+"stack_write$", "want refusal"),
 ("cause-cycle", [("internal/lower/cycles.go", 'if flags&checker.TypeFlagsUnknown != 0 {', 'if false && flags&checker.TypeFlagsUnknown != 0 {')], LOWER, PROBE+"cause_closes_a_cycle$", "want refusal"),
 ("custom-toString", [("internal/lower/class.go", 'if methodName == "toString" && l.errorType(classType) {', 'if false && methodName == "toString" && l.errorType(classType) {')], LOWER, PROBE+"override$", "want refusal"),
 ("optional-generic-field", [("internal/lower/error_classes.go", 'if !known || symbol.Flags&ast.SymbolFlagsOptional != 0 {', 'if !known {')], LOWER, PROBE+"optional_generic_field$", "want refusal"),
 ("erased-cause", [("internal/lower/error_classes.go", 'if proven.Flags()&checker.TypeFlagsUnknown != 0 {', 'if false && proven.Flags()&checker.TypeFlagsUnknown != 0 {')], LOWER, PROBE+"erased_cause$", "want refusal"),
 ("nonliteral-options", [("internal/lower/error_classes.go", "return nil, l.notYet(options, \"ErrorOptions not written as a literal: the erased cause's representation and ownership must be known where it is boxed\")", 'return ir.Undefined{Of: ir.Union}, nil')], LOWER, PROBE+"nonliteral_options$", "want refusal"),
 ("normalization-expansion", [("internal/lower/error_library.go", 'if !constant || len(l.result.Strings[text.Index]) > 536870888/36 {', 'if false && (!constant || len(l.result.Strings[text.Index]) > 536870888/36) {'), ("internal/lower/exceptions.go", 'if !bounded || len(l.result.Strings[text.Index]) > 536870888/36 {', 'if false && (!bounded || len(l.result.Strings[text.Index]) > 536870888/36) {')], LOWER, PROBE+"normalization_expansion$", "want refusal"),
 ("structural-view", [("internal/lower/class_inheritance.go", 'if l.errorType(l.checker.GetTypeAtLocation(node)) && target.Flags()&checker.TypeFlagsObject != 0 && !l.errorType(target) {', 'if false && l.errorType(l.checker.GetTypeAtLocation(node)) && target.Flags()&checker.TypeFlagsObject != 0 && !l.errorType(target) {'), ("internal/lower/class_inheritance.go", 'if l.errorType(from) && to.Flags()&checker.TypeFlagsObject != 0 && !l.errorType(to) {', 'if false && l.errorType(from) && to.Flags()&checker.TypeFlagsObject != 0 && !l.errorType(to) {')], LOWER, PROBE+"(structural_view|nested_structural_view)$", "want refusal"),
 ("fake-error", [("internal/lower/class_inheritance.go", 'if isClassInstance(to) || l.errorType(to) {', 'if isClassInstance(to) {')], LOWER, PROBE+"fake_error$", "want refusal"),
 ("spread-error", [("internal/lower/object.go", 'if l.errorType(l.checker.GetTypeAtLocation(property.AsSpreadAssignment().Expression)) {', 'if false && l.errorType(l.checker.GetTypeAtLocation(property.AsSpreadAssignment().Expression)) {')], LOWER, PROBE+"spread$", "want refusal"),
 ("inherited-own-field", [("internal/lower/object.go", 'if l.inheritedLibraryMember(node)', 'if false && l.inheritedLibraryMember(node)'), ("internal/lower/refusals.go", '&& !isRandom && !l.errorPrototypeRead(node) {', '&& !isRandom && !l.errorPrototypeRead(node) && false {'), ("internal/lower/object.go", 'field.Declarations[0].Kind == ast.KindMethodSignature && load.IsLibrary', 'field.Declarations[0].Kind == ast.KindMethodSignature && false && load.IsLibrary')], LOWER, PROBE+"inherited_method_value$", "want refusal"),
 ("mutable-unknown-alias", [("internal/lower/expression.go", ' && symbol.Declarations[0].Parent.Flags&ast.NodeFlagsConst != 0', '')], LOWER, PROBE+"mutable_unknown_alias$", "want refusal"),
 ("unrecorded-initializer", [("internal/lower/error_classes.go", 'Site: write()', 'Site: 0 * write()')], "./internal/fresh", "TestEveryWriteIsRecordedAndKnown$", "a write lowering didn't record"),
 ("mutable-cause", [("internal/lower/class.go", 'return nil, l.notYet(target, "assigning Error.cause after construction: its erased reference ownership must be proven before mutable cause links are allowed")', 'object, err := l.expression(target.AsPropertyAccessExpression().Expression); if err != nil {return nil,err}; value, err := l.expression(valueNode); if err != nil {return nil,err}; return []ir.Statement{ir.SetProperty{Object:object,Name:"cause",Value:fit(value,ir.Union),Site:l.writeSite(target.AsPropertyAccessExpression().Expression)}},nil')], LOWER, PROBE+"mutable_cause$", "want refusal"),
]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--logs", type=Path, default=Path("/tmp/adamic-error-source-mutants"))
    parser.add_argument("--only", choices=[mutant[0] for mutant in MUTANTS])
    args = parser.parse_args()
    args.logs.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix="adamic-error-overlay-") as directory:
        scratch = Path(directory)
        with (args.logs / "summary.log").open("w") as summary:
            for name, edits, package, probe, assertion in MUTANTS:
                if args.only and name != args.only:
                    continue
                sources = {}
                for relative, old, new in edits:
                    original = sources.get(relative, (ROOT / relative).read_text())
                    if old not in original:
                        raise RuntimeError(f"{name}: source target no longer exists")
                    sources[relative] = original.replace(old, new)
                replacements = {}
                for index, (relative, source) in enumerate(sources.items()):
                    variant = scratch / f"{name}-{index}.go"
                    variant.write_text(source)
                    replacements[str(ROOT / relative)] = str(variant)
                overlay = scratch / f"{name}.json"
                overlay.write_text(json.dumps({"Replace": replacements}))
                log = args.logs / f"{name}.log"
                with log.open("w") as output:
                    result = subprocess.run(["go", "test", "-overlay", str(overlay), package,
                                             "-run", probe, "-count=1", "-v"], cwd=ROOT,
                                            stdout=output, stderr=subprocess.STDOUT)
                observation = log.read_text()
                caught = result.returncode != 0 and assertion in observation and "[build failed]" not in observation
                line = f"{name}: exit {result.returncode}, assertion caught={caught}"
                print(line, flush=True)
                summary.write(line + "\n")
                summary.flush()
                if not caught:
                    raise RuntimeError(f"{name} survived or failed before its assertion: {log}")


if __name__ == "__main__":
    main()

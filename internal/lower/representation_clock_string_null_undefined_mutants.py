"""Run clock admission mutants after sourcing the repository setup environment.

Run without other compiler builds: source files are changed one at a time and restored.
All test output is written directly to the selected log directory.
"""
import argparse
import pathlib
import subprocess

parser = argparse.ArgumentParser()
parser.add_argument("--logs", type=pathlib.Path, default=pathlib.Path("/tmp/clock-string-null-undefined-mutants"))
args = parser.parse_args()
args.logs.mkdir(parents=True, exist_ok=True)
repository = pathlib.Path(__file__).resolve().parents[2]
helper = repository / "internal/lower/representation_clock_string_null_undefined.go"
expression = repository / "internal/lower/expression.go"
original_helper, original_expression = helper.read_text(), expression.read_text()
mutants = [
    ("clock-string-null-undefined-admit-without-null", helper,
     original_helper.replace("!l.includesNull(proven) || ", ""),
     "./internal/lower", "TestRepresentationClockStringNullUndefinedAdmission"),
    ("clock-string-null-undefined-admit-without-undefined", helper,
     original_helper.replace(" || !l.includesUndefined(proven)", ""),
     "./internal/lower", "TestRepresentationClockStringNullUndefinedAdmission"),
    ("clock-string-null-undefined-admit-without-string", helper,
     original_helper.replace("return stringMember", "return stringMember || true"),
     "./internal/lower", "TestRepresentationClockStringNullUndefinedAdmission"),
    ("clock-string-null-undefined-admit-mixed", helper,
     original_helper.replace("if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) == 0 {\n\t\t\treturn false",
                             "if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) == 0 {\n\t\t\tcontinue"),
     "./internal/lower", "TestRepresentationClockStringNullUndefinedAdmission"),
    ("clock-string-null-undefined-typeof-admission", expression,
     original_expression.replace("if operand.Type() == ir.NullishString {\n\t\t\tnull = false\n\t\t}", ""),
     "./internal/oracle", "^TestRepresentationClockStringNullUndefined$"),
]
try:
    for name, path, mutant, package, test in mutants:
        original = original_helper if path == helper else original_expression
        if mutant == original:
            raise RuntimeError("missing mutation site: " + name)
        path.write_text(mutant)
        log_path = args.logs / (name + ".log")
        with log_path.open("w") as log:
            result = subprocess.run(["go", "test", package, "-run", test, "-count=1"],
                                    cwd=repository, stdout=log, stderr=subprocess.STDOUT)
        path.write_text(original)
        output = log_path.read_text()
        if result.returncode == 0 or "--- FAIL:" not in output or "[build failed]" in output:
            raise RuntimeError("mutant did not fail its targeted test: " + name)
        print(name + ": killed by targeted test (exit " + str(result.returncode) + ")", flush=True)
finally:
    helper.write_text(original_helper)
    expression.write_text(original_expression)

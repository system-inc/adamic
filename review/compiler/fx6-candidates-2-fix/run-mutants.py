import difflib
import pathlib
import subprocess

root = pathlib.Path(__file__).resolve().parents[3]
evidence = pathlib.Path(__file__).resolve().parent
source = root / "internal/lower/view_member_read.go"
original = source.read_text()
mutants = [
    ("nullable-receiver", original.replace("return int(l.checker.GetNonNullableType(receiver).Id())", "return int(receiver.Id())"), "./internal/oracle", "TestCheckedViewOptionalReadBoundary|TestCheckedViewObjectPrimitiveSource/optional-receiver", "optional checked read escaped"),
    ("generic-receiver", original.replace("receiver = l.concrete(receiver)", "// mutant: leave the receiver binder rigid", 1).replace("if receiver.Flags()&checker.TypeFlagsTypeParameter != 0 {", "if false {", 1), "./internal/oracle", "TestCheckedViewUntaggedSourceFlows/generic/wrong", "exit codes differ"),
    ("union-receiver", original.replace("if receiver.Flags()&checker.TypeFlagsUnion != 0 {", "if false {", 1), "./stage3/fixtures", "TestFixturesAssertions/assertions/19_identifier_kind.a/stage0", "gap changed"),
]
try:
    for name, mutant, package, test, catcher in mutants:
        assert mutant != original
        source.write_text(mutant)
        patch = difflib.unified_diff(original.splitlines(True), mutant.splitlines(True), fromfile="a/internal/lower/view_member_read.go", tofile="b/internal/lower/view_member_read.go")
        (evidence / (name + ".diff")).write_text("".join(patch))
        command = ["go", "test", package, "-run", test, "-count=1", "-timeout", "90s"]
        with (evidence / (name + ".log")).open("w") as output:
            result = subprocess.run(command, cwd=root, stdout=output, stderr=subprocess.STDOUT, timeout=120)
        log = (evidence / (name + ".log")).read_text()
        assert result.returncode != 0 and catcher in log, (name, result.returncode, log)
        print(name + ": caught by " + catcher, flush=True)
        source.write_text(original)
finally:
    source.write_text(original)

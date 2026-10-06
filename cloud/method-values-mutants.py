#!/usr/bin/env python3
"""Prove the refusal checks admit real unsafe source when removed. Restore every hunk."""
import difflib
import pathlib
import subprocess
import sys

ROOT = pathlib.Path(__file__).resolve().parents[1]
LOWER = "internal/lower/library_method_values.go"
REFUSALS = "internal/lower/refusals.go"
MUTANTS = [
    ("mutable alias", LOWER, [("ast.NodeFlagsConst", "ast.NodeFlagsBlockScoped")]),
    ("opaque field", LOWER, [("if !constMethodInitializer(node) {", "if false {")]),
    ("opaque return", LOWER, [("if !constMethodInitializer(node) {", "if false {")]),
    ("opaque identity", LOWER, [("if !constMethodInitializer(node) {", "if false {")]),
    ("opaque shorthand", REFUSALS, [("if node.Kind == ast.KindShorthandPropertyAssignment {", "if false {")]),
    ("absent array", LOWER, [("if l.mayBeUndefined(receiver) {", "if false {")]),
    ("array element erasure", LOWER, [("if from == nil || to == nil || from != to || !l.sameKeeping(from, to, map[[2]*checker.Type]bool{}) {", "if from == nil || to == nil {")]),
    ("apply spread", LOWER, [("return nil, true, l.notYet(node, \"apply with holes or spread (argument presence is not proven)\")", "written = written[:1]\n\t\t\t\tbreak")]),
    ("call spread", LOWER, [("if hasSpread(node) {", "if false {"), ("written := call.Arguments.Nodes", "written := call.Arguments.Nodes\n\tif hasSpread(node) { written = written[:2] }")]),
    ("string receiver", LOWER, [("of != ir.String {", "of == 0 {")]),
    ("number receiver", LOWER, [("if value.Type() != ir.Number {", "if false {")]),
    ("bind arity", LOWER, [("default:\n\t\treturn nil, l.notYet(node, \"bind of a method with optional, required or partial arguments (its argument adapter is not lowered); use an arrow\")", "default:")]),
    ("bind absent receiver", LOWER, [("if receiver.Type() != ir.String || l.mayBeUndefined(written[0]) {", "if receiver.Type() != ir.String {")]),
]


def patch_for(path, before, after):
    return "".join(difflib.unified_diff(before.splitlines(True), after.splitlines(True),
                                        fromfile="a/" + path, tofile="b/" + path))


def apply(patch, reverse=False):
    command = ["git", "apply"] + (["-R"] if reverse else [])
    subprocess.run(command, input=patch, text=True, cwd=ROOT, check=True)


def main():
    logs = pathlib.Path(sys.argv[1] if len(sys.argv) > 1 else "/tmp/adamic-method-values-source-mutants")
    logs.mkdir(parents=True, exist_ok=True)
    failures = []
    selected = [m for m in MUTANTS if len(sys.argv) < 3 or m[0] == sys.argv[2]]
    for name, path, changes in selected:
        before = (ROOT / path).read_text()
        after = before
        for old, new in changes:
            if old not in after:
                raise RuntimeError("mutation no longer matches: " + name)
            after = after.replace(old, new)
        patch = patch_for(path, before, after)
        apply(patch)
        try:
            with (logs / (name.replace(" ", "-") + ".log")).open("w") as log:
                result = subprocess.run(["go", "test", "./internal/lower", "-count=1", "-run",
                                         "^TestLibraryMethodValueSafety$/^" + name.replace(" ", "_") + "$", "-v"],
                                        cwd=ROOT, stdout=log, stderr=subprocess.STDOUT)
            output = (logs / (name.replace(" ", "-") + ".log")).read_text()
            killed = result.returncode != 0 and "unsafe method value was admitted" in output
            print(name + ": " + ("caught by admission check" if killed else "NOT CAUGHT"), flush=True)
            if not killed:
                failures.append(name)
        finally:
            apply(patch, reverse=True)
            if (ROOT / path).read_text() != before:
                raise RuntimeError("restore changed original: " + path)
    if failures:
        raise SystemExit("surviving or invalid mutants: " + ", ".join(failures))
    print(str(len(selected)) + " source mutants caught; original sources restored")


if __name__ == "__main__":
    main()

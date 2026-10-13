import difflib
import os
from pathlib import Path
import subprocess

root = Path(__file__).resolve().parents[3]
evidence = root / "review/compiler/cycle-closure-targets"
mutants = [
    ("omit-boundary-inputs", "internal/ir/call_targets.go", "if unknownCaller {", "if false && unknownCaller {", "./internal/ir", "^TestClosureTargetsKnownCallerDoesNotHideUnknownCaller$", "known caller hid unbounded caller"),
    ("unknown-is-empty", "internal/ir/call_targets.go", "targets.Unknown = true", "targets.Unknown = false", "./internal/ir", "^TestClosureTargetsBoundOnlyProvenValues$", "bounded without proof"),
    ("omit-captures", "internal/lower/cycles.go", "f.l.result.Functions[function].Environment", "f.l.result.Functions[function].Environment[:0]", "./internal/oracle", "^TestClosureCycleDirect$", "want cycle refusal"),
    ("drop-return-target", "internal/ir/call_targets.go", "for _, function := range from.Functions {", "for _, function := range from.Functions[:min(len(from.Functions), 1)] {", "./internal/ir", "^TestClosureTargetsFollowParametersReturnsAndMapValues$", "want both returned targets"),
    ("omit-map-captures", "internal/lower/cycles.go", "targets := f.l.result.StoredClosureTargets(node.slot)", 'if node.slot == "map" { continue }; targets := f.l.result.StoredClosureTargets(node.slot)', "./internal/oracle", "^TestClosureCycleMap$", "want cycle refusal"),
]
for name, relative, before, after, package, selector, catcher in mutants:
    path = root / relative
    original = path.read_text()
    if before not in original:
        raise RuntimeError("mutation site absent: " + name)
    mutated = original.replace(before, after)
    evidence.joinpath(name + ".patch").write_text("".join(difflib.unified_diff(original.splitlines(True), mutated.splitlines(True), fromfile=relative, tofile=relative)))
    try:
        path.write_text(mutated)
        with evidence.joinpath(name + ".log").open("w") as log:
            result = subprocess.run(["go", "test", package, "-run", selector, "-count=1", "-timeout", "90s", "-v"], cwd=root, stdout=log, stderr=subprocess.STDOUT, timeout=120)
        output = evidence.joinpath(name + ".log").read_text()
        if result.returncode != 1 or catcher not in output or "build failed" in output:
            raise RuntimeError("mutant not caught by intended assertion: " + name)
        print(name + ": caught by " + selector, flush=True)
    finally:
        path.write_text(original)

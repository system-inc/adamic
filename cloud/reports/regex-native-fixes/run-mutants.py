#!/usr/bin/env python3
"""Reproduce the six mutants in a disposable checkout of this branch's tip.

Source /workspace/adamic-tools/env.sh first. Run with the repository path as
argument. Logs go to the directory printed at startup; no existing checkout is
mutated. Requires Linux for LeakSanitizer. The long leak mutant takes several minutes.
"""
import os
from pathlib import Path
import subprocess
import sys
import tempfile

if sys.platform != "linux":
    raise SystemExit("The frame-free proof requires Linux LeakSanitizer")
root = Path(sys.argv[1]).resolve()
scratch = Path(tempfile.mkdtemp(prefix="regex-native-mutants-"))
checkout = scratch / "checkout"
print("logs:", scratch, flush=True)
subprocess.run(["git", "worktree", "add", "--detach", str(checkout), "HEAD"], cwd=root, check=True)
cohere = checkout / "cohere"
cohere.rmdir()  # Newly created, uninitialized submodule directory only.
cohere.symlink_to(root / "cohere", target_is_directory=True)
# Use the existing checker source path, so its warmed Go cache is reusable.
(checkout / "go.work").write_text('go 1.27\nuse (\n.\n"' + str(root / "cohere/TypeScript/tsc") + '"\n)\n')

mutants = [
    ("class-octal", "internal/regexp/parser.go", "if c >= '1' && c <= '9' {",
     "if !inClass && c >= '1' && c <= '9' {", "./internal/regexp", "^TestLegacyClassOctalNode$", "DISAGREEMENT"),
    ("quantifier-order", "internal/regexp/parser.go", "quantifierOrderBound(min).Cmp(quantifierOrderBound(max))",
     "min.Cmp(max)", "./internal/regexp", "^TestLargeQuantifierBoundsNode$", "DISAGREEMENT"),
    ("class-duplicates", "internal/native/runtime/regexp.c", "count = unique;",
     "(void)unique;", "./internal/oracle", "^TestRegExpNativeTiming$/class_string_duplicates", "native execution exceeded 3s"),
    ("input-cache", "internal/native/runtime/regexp.c", "if (cached != NULL) {",
     "if (cached != NULL && false) {", "./internal/oracle", "^TestRegExpNativeTiming$/quadratic", "native execution exceeded 3s"),
    ("compound-write", "internal/lower/regexp.go", "ast.IsAssignmentOperator(assignment.OperatorToken.Kind)",
     "assignment.OperatorToken.Kind == ast.KindEqualsToken", "./internal/lower", "^TestRegExpGroupCompoundRefusals$", "want regex assignment NotYet for g.y +="),
    ("frame-free", "internal/native/runtime/regexp.c", "free(frame);",
     "(void)frame;", "./internal/oracle", "^TestRegExpLongBacktrackNode$", "LeakSanitizer: detected memory leaks"),
]
try:
    for name, filename, old, new, package, test, marker in mutants:
        path = checkout / filename
        original = path.read_text()
        assert original.count(old) == 1, (name, "mutation site changed")
        path.write_text(original.replace(old, new, 1))
        log = scratch / (name + ".txt")
        try:
            with log.open("w") as output:
                result = subprocess.run(["go", "test", package, "-run", test, "-count=1", "-v", "-timeout=15m"],
                                        cwd=checkout, env=dict(os.environ, ADAMIC_GATE_UNCACHED="1"),
                                        stdout=output, stderr=subprocess.STDOUT, timeout=1000)
            observation = log.read_text()
            assert result.returncode != 0 and marker in observation, (name, "not caught by its check", log)
            if name == "frame-free":
                assert "stdout differs" not in observation and "exit codes differ" not in observation
                assert "release build:" not in observation
            print(name, "caught", flush=True)
        finally:
            path.write_text(original)
finally:
    subprocess.run(["git", "worktree", "remove", "--force", str(checkout)], cwd=root, check=True)

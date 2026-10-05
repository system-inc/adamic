# drive.py <worktree> [name...]: applies each named mutant (or all) to the worktree's invariance.go, runs
# run.sh, compares with the baseline, restores the file, and appends one line per mutant to results.txt.
import os, subprocess, sys
from mutants import mutants
worktree = sys.argv[1]
wanted = set(sys.argv[2:])
here = os.path.dirname(os.path.abspath(__file__))
path = os.path.join(worktree, "internal/lower/invariance.go")
original = open(path).read()
base = open(os.path.join(here, "base.verdicts")).read().splitlines()
for name, old, new in mutants:
    if wanted and name not in wanted:
        continue
    if original.count(old) != 1:
        print(name, "does not apply:", original.count(old), "matches")
        continue
    open(path, "w").write(original.replace(old, new))
    try:
        subprocess.run([os.path.join(here, "run.sh"), worktree, name], check=False)
    finally:
        open(path, "w").write(original)
    tests = open(os.path.join(here, name + ".tests")).read()
    testsCaught = "FAIL" in tests or "build failed" in tests or "ok  \tgithub.com/system-inc/adamic/internal/oracle" not in tests
    verdicts = open(os.path.join(here, name + ".verdicts")).read().splitlines()
    changed = [after for before, after in zip(base, verdicts) if before != after]
    line = "%s: branch tests %s; probes changed %d%s" % (name, "CATCH" if testsCaught else "pass", len(changed), (" (" + "; ".join(c.split(":")[0] for c in changed) + ")") if changed else "")
    print(line, flush=True)
    with open(os.path.join(here, "results.txt"), "a") as results:
        results.write(line + "\n")

#!/usr/bin/env python3
"""The test-only lane's checks at the landing: the ones that take seconds, never minutes (@system_adamic, Oct 9
03:51Z, after two test-only landings reddened every gate on main with nothing in front of them). Kirk's rule
stands, no tests run here; these refuse a violator at its own landing, with the reason, instead.

- gofmt -l on every changed Go file.
- The declared-tools check: a command a changed Go file runs by literal name must be in developer tools'
  cloud/fast-gate/tools.txt (origin/devtools/fast-gate), the same pattern the gate's tools stage reads.
- The t.Parallel analyzer (cmd/adamic-gate's TestEveryTestIsParallelOrSaysWhy, which only reads the tree): a
  changed test file's top-level test either calls Parallel() first or says why not.
- go vet on the changed test packages, when it finishes inside 10 s; past that it's skipped and said so.
- a-check on each added or changed .a outside a Go package's tree (and not a-check-exempt): stage 0's front end,
  as the gate runs it, since a-check type-checks every such .a in the tree.

usage: lane-checks.py <old main> <tree> <commit holding the tree>   (push-main, from the merge tree)
Prints one summary line and exits 0, or prints each violation and exits 1.

A worker runs the same checks on its own committed branch before pushing, from the repository's root (in
its own checkout, nothing moved), so a change never travels just to be refused (@system_adamic, Oct 9):
  git fetch -q origin main devtools/fast-gate cloud/merge-tree && git show origin/cloud/merge-tree:cloud/integration/lane-checks.py | python3 -
"""
import os, re, subprocess, sys, tempfile, time


def run(*arguments, **options):
    return subprocess.run(arguments, capture_output=True, text=True, **options)


started = time.monotonic()
# With no arguments it checks the worker's HEAD against main, in place.
inPlace = len(sys.argv) < 4
if inPlace:
    commit = run("git", "rev-parse", "HEAD").stdout.strip()
    old = run("git", "merge-base", commit, "origin/main").stdout.strip()
    tree = run("git", "rev-parse", f"{commit}^{{tree}}").stdout.strip()
    laneTree = run("git", "rev-parse", "--show-toplevel").stdout.strip()
    if not old:
        sys.exit("lane-checks: no merge base with origin/main; fetch it first")
else:
    old, tree, commit = sys.argv[1:4]
    laneTree = os.path.expanduser(os.environ.get("ADAMIC_LANE_TREE", "~/.adamic-lane-tree"))
toolLiteral = re.compile(r'exec\.(?:Command\(|CommandContext\([A-Za-z_.()]+, |LookPath\()"([^"]+)"')
vetSeconds = 10
# The submodules (cohere, and its TypeScript) come from a local checkout's module store when it holds the
# pinned commit, in seconds; else from their remotes. Without them most packages can't load, so vet would
# miss even a compile error (Oct 9: stage1/cohere/json's missing sync import reached main this way).
modules = os.path.expanduser(os.environ.get("ADAMIC_LANE_MODULES", "~/Projects/system/adamic/.git/modules"))
toolsRef = os.environ.get("ADAMIC_LANE_TOOLS_REF", "origin/devtools/fast-gate")


def packageOf(finding):
    # "vet: stage1/x/y_test.go:9:2: message" (or without the prefix) -> "./stage1/x"
    return "./" + os.path.dirname(re.sub(r"^(vet: )?", "", finding.strip()).split(":")[0])


changed = run("git", "diff", "--name-only", "--diff-filter=AM", old, tree).stdout.split()
goFiles = [path for path in changed if path.endswith(".go") and not path.startswith(("cohere/", "stage3/upstream/"))]
problems, notes = [], []

# gofmt and the tools, read from the blobs.
with tempfile.TemporaryDirectory() as scratch:
    for path in goFiles:
        target = os.path.join(scratch, path)
        os.makedirs(os.path.dirname(target), exist_ok=True)
        with open(target, "w") as handle:
            handle.write(run("git", "show", f"{tree}:{path}").stdout)
    unformatted = run("gofmt", "-l", ".", cwd=scratch).stdout.split() if goFiles else []
problems += [f"{path.removeprefix('./')} isn't gofmt-formatted" for path in unformatted]
run("git", "fetch", "-q", "origin", "devtools/fast-gate")
declared = {line.split("\t")[0] for line in run("git", "show", f"{toolsRef}:cloud/fast-gate/tools.txt").stdout.splitlines()
            if line.strip() and not line.startswith("#")}
if not declared:
    problems.append("can't read cloud/fast-gate/tools.txt from origin/devtools/fast-gate")
for path in goFiles:
    for number, line in enumerate(run("git", "show", f"{tree}:{path}").stdout.splitlines(), 1):
        for name in toolLiteral.findall(line):
            if declared and name not in declared:
                problems.append(f"{path}:{number} runs {name}, which cloud/fast-gate/tools.txt doesn't declare (declare it on devtools/fast-gate, or don't shell out)")

# a-check type-checks every .a outside a Go package's tree, so a .a is never test-only to it (@system_adamic, Oct 9
# 08:22: a review/ witness that didn't type-check reddened main's canary through this lane). The same selection as the
# gate's: no package directory at or above it, and no a-check-exempt glob in developer tools' executors.txt.
aFiles = [path for path in changed if path.endswith(".a")]
aChecked = []
if aFiles:
    import fnmatch
    goDirectories = {os.path.dirname(path) for path in run("git", "ls-tree", "-r", "--name-only", tree).stdout.split() if path.endswith(".go")}

    def owned(path):
        directory = os.path.dirname(path)
        while directory:
            if directory in goDirectories:
                return True
            directory = os.path.dirname(directory)
        return False
    exemptions = [line.split()[1] for line in run("git", "show", f"{toolsRef}:cloud/fast-gate/executors.txt").stdout.splitlines()
                  if line.startswith("a-check-exempt") and len(line.split()) >= 2]
    aChecked = [path for path in aFiles if not owned(path) and not any(fnmatch.fnmatchcase(path, glob) for glob in exemptions)]
    # A tree with no stage 0 to run (a scratch repository) can't be a-checked here; say so instead of refusing.
    if aChecked and not run("git", "cat-file", "-e", f"{tree}:cmd/adamic").returncode == 0:
        notes.append(f"no cmd/adamic in this tree to a-check {len(aChecked)} .a files")
        aChecked = []

# The analyzer, a-check and vet need the tree on disk: the lane's own worktree, moved to this commit.
testPackages = sorted({"./" + os.path.dirname(path) for path in goFiles if path.endswith("_test.go")})
if (testPackages or aChecked) and not problems:
    if not inPlace and not os.path.isdir(laneTree):
        run("git", "worktree", "add", "-q", "--detach", laneTree, commit)
    checkedOut = run("true") if inPlace else run("git", "-C", laneTree, "checkout", "-q", "--detach", "--force", commit)
    if checkedOut.returncode != 0:
        problems.append(f"the lane's worktree couldn't check out {commit[:8]}: {checkedOut.stderr.strip()[:200]}")
    else:
        for directory, name, store in (() if inPlace else ((laneTree, "cohere", os.path.join(modules, "cohere")),
                                       (os.path.join(laneTree, "cohere"), "TypeScript", os.path.join(modules, "cohere/modules/TypeScript")))):
            if not os.path.exists(os.path.join(directory, ".gitmodules")):
                break
            local = run("git", "-c", "protocol.file.allow=always", "-c", f"submodule.{name}.url={store}", "submodule", "update", "--init", name, cwd=directory, timeout=300) if os.path.isdir(store) else None
            if local is None or local.returncode != 0:
                if run("git", "submodule", "update", "--init", name, cwd=directory, timeout=300).returncode != 0:
                    notes.append(f"no {name} submodule here")
                    break
        # go.work names cohere's TypeScript module; without it, the repository's own module alone.
        goEnvironment = dict(os.environ) if os.path.exists(os.path.join(laneTree, "cohere/TypeScript/tsc/go.mod")) else dict(os.environ, GOWORK="off")
        analyzer = os.path.join(laneTree, "cmd/adamic-gate/parallel_test.go")
        if os.path.exists(analyzer):
            ran = run("go", "test", "-count=1", "-run", "^TestEveryTestIsParallelOrSaysWhy$", "-v", "./cmd/adamic-gate", cwd=laneTree, timeout=120, env=goEnvironment)
            if ran.returncode != 0:
                problems.append("the t.Parallel analyzer failed: " + (ran.stdout + ran.stderr).strip().splitlines()[-1][:300])
            for line in ran.stdout.splitlines():
                key, _, rest = line.strip().partition(": call the test parameter's Parallel()")
                if rest:
                    path, _, test = key.rpartition("\t") if "\t" in key else key.rpartition(" ")
                    path = path.split()[-1] if path else path
                    if path in changed:
                        problems.append(f"{path} {test}: call the test parameter's Parallel() as the first statement, or add // Not parallel: <shared state> above it")
        else:
            notes.append("no t.Parallel analyzer on this tree")
        # a-check, as the gate runs it (run.py aCheck): stage 0's front end on each file; a clean result or a
        # "can't lower ... yet" stop passes, and a first line "// a-check: refused <rule>" or "// a-check: type
        # error <code>" must fail exactly that way.
        if aChecked:
            with tempfile.TemporaryDirectory() as binaries:
                adamic = os.path.join(binaries, "adamic")
                built = run("go", "build", "-o", adamic, "./cmd/adamic", cwd=laneTree, timeout=300, env=goEnvironment)
                if built.returncode != 0:
                    problems.append("a-check couldn't build cmd/adamic: " + (built.stdout + built.stderr).strip().splitlines()[-1][:200])
                for path in aChecked if built.returncode == 0 else []:
                    with open(os.path.join(laneTree, path), errors="replace") as handle:
                        header = handle.readline().strip()
                    expect = header[len("// a-check:"):].strip() if header.startswith("// a-check:") else ""
                    errors = run(adamic, "c", path, cwd=laneTree, timeout=120)
                    text = errors.stderr
                    if errors.returncode == 0 or ("can't lower" in text and " yet" in text):
                        outcome = "checked"
                    elif "Adamic 0.1 refuses" in text:
                        outcome = "refused"
                    elif " error TS" in text:
                        outcome = "type error"
                    else:
                        outcome = "failed"
                    if expect.startswith("refused"):
                        ok = outcome == "refused" and expect[len("refused"):].strip() in text
                    elif expect.startswith("type error"):
                        code = expect[len("type error"):].strip()
                        ok = outcome == "type error" and (not code or ("error " + code) in text)
                    else:
                        ok = outcome == "checked"
                    if not ok:
                        first = text.strip().splitlines()[0][:200] if text.strip() else ""
                        problems.append(f"a-check: {path}: {outcome}, expected {expect or 'checked'} ({first})")
            notes.append(f"a-check {len(aChecked)} .a files")
        try:
            if not testPackages:
                # Only a .a brought the tree here; there's nothing to vet.
                raise LookupError
            vetted = run("go", "vet", *testPackages, cwd=laneTree, timeout=vetSeconds, env=goEnvironment)
            # A finding names a file and position; a package that can't load here (it needs the cohere
            # submodule, say) is the gate's to vet, not a reason to refuse.
            loading = re.compile(r"replacement directory|no required module|cannot find module|does not exist|no such file or directory|go\.mod|missing go\.sum|is not in std|cannot load")
            findings = [line for line in (vetted.stdout + vetted.stderr).splitlines() if re.search(r"\.go:\d+:\d+: ", line) and not loading.search(line)]
            if findings and inPlace:
                notes.append("if main already fails vet in one of these packages, the lane won't refuse it there")
            if findings and not inPlace:
                # vet stops at a package's first type error, so a package main already breaks can't be judged
                # here: only packages that vet clean on main refuse, or a fix into a broken package never lands.
                broken = sorted({packageOf(line) for line in findings})
                run("git", "-C", laneTree, "checkout", "-q", "--detach", "--force", old)
                run("git", "submodule", "update", "-q", cwd=laneTree, timeout=300)
                before = run("go", "vet", *broken, cwd=laneTree, timeout=vetSeconds * 3, env=goEnvironment)
                alreadyBroken = {packageOf(line)
                                 for line in (before.stdout + before.stderr).splitlines() if re.search(r"\.go:\d+:\d+: ", line) and not loading.search(line)}
                findings = [line for line in findings if packageOf(line) not in alreadyBroken]
                if alreadyBroken:
                    notes.append("vet already fails on main in " + " ".join(sorted(alreadyBroken)))
            if findings:
                problems.append("go vet: " + " | ".join(findings[:5])[:600])
            elif vetted.returncode != 0:
                notes.append("vet couldn't load every package here: " + (vetted.stdout + vetted.stderr).strip().splitlines()[-1][:120])
            else:
                notes.append(f"vet {len(testPackages)} packages")
        except LookupError:
            pass
        except subprocess.TimeoutExpired:
            notes.append(f"vet skipped, over {vetSeconds} s")

seconds = time.monotonic() - started
if problems:
    print("\n".join(problems + notes))
    sys.exit(1)
print(f"lane checks {seconds:.1f} s: gofmt and tools on {len(goFiles)} Go files, t.Parallel on {len(testPackages)} test packages" + ("; " + "; ".join(notes) if notes else ""))

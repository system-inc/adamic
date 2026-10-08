#!/usr/bin/env python3
"""Census: no test or build reads push-main's record paths.

usage (from the repository): python3 cloud/integration/record-paths-test.py [<commit>]   (default origin/main)

push-main lands a gated tree over a main that moved only by record commits, and star-train keeps a
slice's gates when main moves past its base only by those paths, and takes the landed tree's whole
gate as main's confirmation. That is honest only while nothing a gate runs reads a record: a test
that read the velocity table, a meter run or stage 3's progress record could change its verdict on a
record-only commit the kept gate never saw (@system_adamic, October 8). So every line of code or
data that names a record path is listed here with why it can't move a verdict, keyed on the exact
line, and any other line fails this test. If one is found that a test or build really reads, that
path leaves the record set in recordPaths below, push-main and star-train together.

It can't see a path built from pieces (filepath.Join("meter", "runs")); it reads what is spelled.
"""
import re
import subprocess
import sys

# push-main.sh's record paths, the same three star-train.py's recordsOnly accepts.
recordPaths = re.compile(r"landings\.csv|meter/runs|progress\.json")
# Files that can run in a gate or a build, and data files a test could read a path from.
code = ["*.go", "*.py", "*.sh", "*.mjs", "*.cjs", "*.js", "*.ts", "*.a", "*.json", "Makefile", "*.mk"]
# The records themselves, never code.
records = [":!documentation/velocity/**", ":!stage3/meter/runs/**", ":!stage3/progress.json"]

explained = {
    ("stage3/meter/twice-daily.sh", "runs=${STAGE3_METER_RUNS:-$repository/stage3/meter/runs}"):
        "the meter writes its runs after a landing; no gate runs it",
    ("stage3/meter/compiler_test.py", "run, = (self.repository / 'stage3/meter/runs').iterdir()"):
        "reads a run it made in its own temporary repository (self.root / 'repository')",
    ("stage3/meter/compiler_test.py", "self.assertFalse((self.repository/'stage3/meter/runs').exists())"):
        "its own temporary repository",
    ("stage3/meter/compiler_test.py", "self.assertFalse((self.repository / 'stage3/meter/runs').exists())"):
        "its own temporary repository",
    ("stage3/drivers/tsc-entry/collect.py", "run = sorted((repository / 'stage3/meter/runs').glob('*/report.json'))[-1]"):
        "a hand-run evidence collector (REFRESH.md); no gate runs it",
    ("stage3/drivers/tsc-entry/evidence/meter-comparison.json", '"meter_run": "stage3/meter/runs/20261008T000754Z.hFDR7Q/report.json",'):
        "recorded evidence naming the run it compared, data no test reads",
    ("stage3/drivers/tsc-entry/evidence/provenance.json", '"meter_run": "stage3/meter/runs/20261008T000754Z.hFDR7Q/report.json",'):
        "recorded evidence naming the run it compared, data no test reads",
}


def grep(commit, pattern, paths, *options):
    found = subprocess.run(["git", "grep", *options, "-e", pattern, commit, "--", *paths], capture_output=True, text=True)
    if found.returncode not in (0, 1):
        raise SystemExit(f"git grep failed: {found.stderr.strip()}")
    return found.stdout.splitlines()


def readers(commit):
    unexplained = []
    for line in grep(commit, recordPaths.pattern, code + records, "-n", "-E"):
        path, _, rest = line.removeprefix(f"{commit}:").partition(":")
        text = rest.partition(":")[2].strip()
        if (path, text) in explained:
            continue
        # A JSON file naming a record path is data (evidence a worker recorded). It can move a verdict
        # only through code that reads it, so it's explained when no code names the file.
        if path.endswith(".json"):
            name = path.rsplit("/", 1)[-1]
            if not grep(commit, name, [pattern for pattern in code if pattern != "*.json"] + records, "-l", "-F"):
                continue
        unexplained.append(f"{path}: {text[:160]}")
    return unexplained


if __name__ == "__main__":
    commit = sys.argv[1] if len(sys.argv) > 1 else "origin/main"
    unexplained = readers(commit)
    if unexplained:
        print(f"{len(unexplained)} lines in {commit} name a record path with no reason they can't move a verdict:")
        print("\n".join(unexplained))
        sys.exit(1)
    print(f"{commit}: no test or build reads a record path ({len(explained)} explained lines)")

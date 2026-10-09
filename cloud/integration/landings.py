#!/usr/bin/env python3
"""The velocity table, printed as CSV from git: one row per landing on main, oldest first.

usage (from any clone of the repository): python3 cloud/integration/landings.py [<main ref>]   (default origin/main)

Since Oct 9 each landing is one commit on main carrying its numbers as trailers (push-main.sh), and
documentation/velocity/landings.csv left main. The rows that file held come first, read from the last
commit that had it, then one row per landing commit on main's first-parent line. The columns are the
old file's, so the witness and the waterfall read the same table.
"""
import csv, datetime, subprocess, sys

table = "documentation/velocity/landings.csv"
header = ["pushed_at_utc", "old_main", "new_main", "commits_landed", "branches_landed", "gate_minutes", "pass", "fail", "skip", "backlog_commits"]
trailers = {"Old-main": "old_main", "Landed-commits": "commits_landed", "Branches": "branches_landed", "Gate-minutes": "gate_minutes",
            "Pass": "pass", "Fail": "fail", "Skip": "skip", "Backlog": "backlog_commits"}


def git(*arguments):
    return subprocess.run(["git", *arguments], capture_output=True, text=True, check=True).stdout


main = sys.argv[1] if len(sys.argv) > 1 else "origin/main"
writer = csv.writer(sys.stdout, lineterminator="\n")
writer.writerow(header)

# The file's own rows: from main while it's still there, else from the parent of the commit that deleted it.
historic = subprocess.run(["git", "show", f"{main}:{table}"], capture_output=True, text=True)
if historic.returncode != 0:
    deleted = git("log", "-1", "--format=%H", "--diff-filter=D", main, "--", table).strip()
    historic = subprocess.run(["git", "show", f"{deleted}^:{table}"], capture_output=True, text=True) if deleted else historic
if historic.returncode == 0:
    for line in historic.stdout.splitlines()[1:]:
        if line.strip():
            # Rows were written unquoted with commas kept out of the branches field, so they pass through as they are.
            sys.stdout.write(line + "\n")

# Every landing commit since: the ones carrying Landed-commits, along main's first-parent line.
separator = "\x1e"
log = git("log", "--first-parent", "--reverse", f"--format=%H%x1f%cI%x1f%(trailers:only,unfold){separator}", main)
for entry in log.split(separator):
    if "\x1f" not in entry:
        continue
    sha, committed, block = entry.strip("\n").split("\x1f", 2)
    values = {}
    for line in block.splitlines():
        key, _, value = line.partition(": ")
        if key in trailers:
            values[trailers[key]] = value.strip()
    if "commits_landed" not in values:
        continue
    pushed = datetime.datetime.fromisoformat(committed).astimezone(datetime.timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
    # The old rows kept commas out of the branches field so a plain split reads them; these do the same.
    values.update(pushed_at_utc=pushed, new_main=sha, branches_landed=values.get("branches_landed", "").replace(",", ";"))
    writer.writerow([values.get(column, "") for column in header])

#!/usr/bin/env python3
"""The merge queue (#qth06pv, design in merge-queue.md beside this file): integration's candidates stacked on main's
tip, main+A, main+A+B, main+A+B+C, each pushed as a cut the fast-gate watcher and Loom's pool gate like any other, so
every prefix is judged at once and the longest green one can land.

This step builds and submits the stacks (#dyre8as). Members are the candidates file's lines, ordered by their cut's
Gate-tier (highest first), then by file order. Each stack is a chain of merges made with git merge-tree and
git commit-tree on main's tip, so no checkout is touched. A member that conflicts with the prefix before it is left
out and reported with the conflicting files; one already on main is skipped. A run whose main and members match the
stacks already submitted changes nothing; otherwise the old stacks are withdrawn from the pool and the new ones pushed.

The state file (~/.adamic-merge-queue/state.json) is written whole under the gate lane's lock, the one cut.sh takes.

usage: cloud/integration/merge-queue.py stacks [--dry-run]
"""
import datetime, fcntl, json, os, subprocess, sys, time

candidates = os.environ.get("QUEUE_CANDIDATES", os.path.expanduser("~/.adamic-integration/candidates"))
state = os.environ.get("QUEUE_STATE", os.path.expanduser("~/.adamic-merge-queue"))
lock = os.environ.get("QUEUE_LOCK", os.path.expanduser("~/.adamic-gate-lane/lock"))
withdraw = os.environ.get("QUEUE_WITHDRAW", os.path.expanduser("~/.loom/bin/withdraw.sh"))
depth = int(os.environ.get("QUEUE_DEPTH", "3"))
author = "system_adamic_release_integration"


class GitError(Exception):
    pass


def git(*arguments, check=True):
    result = subprocess.run(["git", *arguments], capture_output=True, text=True, timeout=300)
    if check and result.returncode != 0:
        raise GitError("git %s: %s" % (" ".join(arguments[:2]), result.stderr.strip()))
    return result.stdout.strip()


def tier(sha):
    """The cut's Gate-tier trailer, 30 when it has none."""
    for line in git("log", "-1", "--format=%B", sha).splitlines():
        if line.startswith("Gate-tier:"):
            return int(line.split(":", 1)[1].strip())
    return 30


def members():
    """The candidates file's lines in queue order: tier highest first, then file order."""
    lines = [line.split("\t") for line in open(candidates).read().splitlines() if line and not line.startswith("#")]
    found = []
    for position, fields in enumerate(lines):
        branch, sha, task = fields[0], fields[1], fields[2]
        found.append({"branch": branch, "sha": sha, "task": task, "tier": tier(sha), "position": position})
    return sorted(found, key=lambda member: (-member["tier"], member["position"]))


def build(main, ordered):
    """Stacks on main: each one the stack before it merged with the next member that merges clean."""
    stacks, conflicts, skipped = [], [], []
    prefix, inside = main, []
    for member in ordered:
        if len(stacks) == depth:
            break
        if subprocess.run(["git", "merge-base", "--is-ancestor", member["sha"], prefix], timeout=300).returncode == 0:
            skipped.append(member)
            continue
        merged = subprocess.run(["git", "merge-tree", "--write-tree", "--name-only", prefix, member["sha"]],
                                capture_output=True, text=True, timeout=300)
        if merged.returncode != 0:
            # The first line is the tree; the conflicting paths follow it up to a blank line.
            files = []
            for line in merged.stdout.splitlines()[1:]:
                if not line:
                    break
                files.append(line)
            conflicts.append(dict(member, files=files))
            continue
        tree = merged.stdout.splitlines()[0]
        inside = inside + [member]
        message = "Merge queue stack %d on main %s: + %s %s\n\nGate-runs: deferred\nGate-tier: %d\nQueue-members: %s\nQueue-main: %s\n" % (
            len(stacks) + 1, main[:8], member["branch"], member["sha"][:8], max(m["tier"] for m in inside),
            " ".join(m["sha"][:8] for m in inside), main)
        prefix = git("commit-tree", tree, "-p", prefix, "-p", member["sha"], "-m", message)
        stacks.append({"k": len(stacks) + 1, "sha": prefix, "branch": "cloud/land-queue-%d-%s" % (len(stacks) + 1, prefix[:8]),
                       "members": [m["sha"] for m in inside], "tasks": [m["task"] for m in inside],
                       "state": "gating", "record": None, "voids": 0})
    return stacks, conflicts, skipped


def locked():
    handle = open(lock, "w")
    for _ in range(120):
        try:
            fcntl.flock(handle, fcntl.LOCK_EX | fcntl.LOCK_NB)
            return handle
        except BlockingIOError:
            time.sleep(5)
    sys.exit("merge-queue: the gate lane held its lock for 10 minutes")


def stacksCommand(dryRun):
    git("fetch", "-q", "origin", "main")
    main = git("rev-parse", "origin/main")
    ordered = members()
    for member in ordered:
        if git("cat-file", "-t", member["sha"], check=False) != "commit":
            git("fetch", "-q", "origin", member["branch"], check=False)
    stacks, conflicts, skipped = build(main, ordered)
    path = os.path.join(state, "state.json")
    previous = json.load(open(path)) if os.path.exists(path) else {"main": None, "stacks": []}
    same = previous["main"] == main and [s["members"] for s in previous["stacks"]] == [s["members"] for s in stacks]
    for member in conflicts:
        print("conflict %s %s with the prefix before it: %s" % (member["branch"], member["sha"][:8], ", ".join(member["files"])))
    for member in skipped:
        print("on main already: %s %s" % (member["branch"], member["sha"][:8]))
    if same:
        print("unchanged: %d stacks on main %s" % (len(stacks), main[:8]))
        return 0
    for stack in stacks:
        print("stack %d %s %s: %s" % (stack["k"], stack["branch"], stack["sha"][:8], " ".join(m[:8] for m in stack["members"])))
    if dryRun:
        return 0
    handle = locked()
    try:
        if stacks:
            git("push", "-q", "origin", *["%s:refs/heads/%s" % (stack["sha"], stack["branch"]) for stack in stacks])
        for old in previous["stacks"]:
            if old["sha"] not in {stack["sha"] for stack in stacks}:
                subprocess.run([withdraw, old["sha"], author, "merge queue restacked on main %s" % main[:8]],
                               capture_output=True, text=True, timeout=120)
                git("push", "-q", "origin", ":refs/heads/" + old["branch"], check=False)
        os.makedirs(state, exist_ok=True)
        written = {"main": main, "built": datetime.datetime.now(datetime.timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
                   "stacks": stacks, "conflicts": [{"sha": m["sha"], "task": m["task"], "files": m["files"]} for m in conflicts],
                   "kicked": previous.get("kicked", [])}
        open(path + ".new", "w").write(json.dumps(written, indent=1) + "\n")
        os.replace(path + ".new", path)
    finally:
        handle.close()
    return 0


if __name__ == "__main__":
    if len(sys.argv) < 2 or sys.argv[1] != "stacks":
        sys.exit(__doc__.strip().splitlines()[-1])
    sys.exit(stacksCommand("--dry-run" in sys.argv[2:]))

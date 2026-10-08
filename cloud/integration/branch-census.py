#!/usr/bin/env python3
"""Classifies every remote branch that isn't main or an area by what its patches have reached.

A branch's commits outside main and every area are compared by patch id, not commit id, with
every commit main and the areas gained in the lookback window, so a branch that was rebased or
cherry-picked after its patches landed counts as superseded. The classes:

  merged       its tip is already in main or an area (an ancestor of one of them)
  merges-only  its only commits outside main and the areas are merges (an integration or
               speculative branch, or one whose work landed through other merges); never deleted,
               since a merge can carry a conflict resolution nothing upstream has
  superseded   it has commits of its own, but every patch they carry is already in main or an area
  in-flight    patches not landed, last commit under --fresh-hours old (a worker may be on it)
  unmerged     patches not landed, last commit between --fresh-hours and --stalled-hours old
  stalled      patches not landed, no commit for --stalled-hours (2) up to --stale-hours: work
               that stopped short or finished and was never merged; each run names these to their
               owner with the last commit's subject. (Codex sessions don't record their branch, so
               a branch can't yet be tied to whether its worker is still running.)
  abandoned    patches not landed, last commit older than --stale-hours
  closed       its owner says it's dead or held elsewhere (census-closed.tsv, with the reason)

census-owners.tsv names the owner where the prefix or the messages get it wrong.

--delete-merged deletes the merged branches from origin by name, each pushed with a lease on the
tip the census saw, so a branch that moved since is left alone, and appends every name it deleted
to --deleted-log. Every commit on a merged branch is in main or an area, so nothing is lost.

Integration branches (cloud/integrate-*, cloud/speculate-*) and coverage readers' notes are
counted under integration. The owner comes from the branch's prefix, from the fleet roster's
labels, and, when --messages-db names the kingdom's os.db, from the Circle that first sent
integration a message naming the branch. It deletes nothing and changes no ref.

usage: cloud/integration/branch-census.py [--fresh-hours 1] [--stale-hours 24] [--lookback-days 10]
         [--messages-db ~/Projects/ahra/modules/os/data/os.db] [--tsv out.tsv]
         [--delete-merged --deleted-log deleted.tsv]
"""

import argparse
import collections
import datetime
import os
import sqlite3
import subprocess
import sys

prefixOwners = [
    ("compiler/", "system_adamic_compiler"),
    ("runtime/", "system_adamic_runtime"),
    ("library/", "system_adamic_library"),
    ("apple/", "system_adamic_platforms"),
    ("codex/stage3-", "system_adamic_typescript"),
    ("codex/lint-", "system_cohere_lint"),
    ("codex/typeaware-", "system_cohere_lint"),
    ("codex/stage1-", "system_cohere_format"),
    ("codex/wasm", "system_adamic_platforms"),
    ("codex/host-", "system_adamic_platforms"),
    ("codex/workers-", "system_adamic_platforms"),
    ("codex/library-", "system_adamic_library"),
    ("codex/regex-", "system_adamic_library"),
    ("devtools/", "system_adamic_developer_tools"),
    ("gate-logs/", "system_adamic_developer_tools"),
    ("gate/", "system_adamic_integration"),
    ("cloud/grok-coverage-", "system_adamic_integration"),
    ("cloud/gate-log-", "system_adamic_integration"),
    ("cloud/fixes-3-failures", "system_adamic_integration"),
    ("cloud/integrate-", "system_adamic_integration"),
    ("cloud/speculate-", "system_adamic_integration"),
    ("coverage/", "system_adamic_integration"),
]


def readTable(path):
    table = {}
    if os.path.exists(path):
        for line in open(path):
            if line.strip() and not line.startswith("#"):
                key, value = line.rstrip("\n").split("\t", 1)
                table[key] = value
    return table


def git(*arguments, input_text=None):
    return subprocess.run(["git", *arguments], input=input_text, capture_output=True, text=True, check=True).stdout


def patchIds(revisionArguments):
    """Maps commit to patch id for every non-merge commit the revision arguments select."""
    log = subprocess.run(["git", "log", "--no-merges", "-p", "--format=commit %H", *revisionArguments], capture_output=True, check=True).stdout
    output = subprocess.run(["git", "patch-id", "--stable"], input=log, capture_output=True, check=True).stdout.decode()
    commitToPatch = {}
    for line in output.splitlines():
        patch, commit = line.split()
        commitToPatch[commit] = patch
    return commitToPatch


def ownersFromMessages(databasePath, branches):
    """The first Circle that sent integration a message naming each branch. @system_adamic routes
    other Circles' branches, so it counts only when no Circle named the branch."""
    owners = {}
    routed = {}
    connection = sqlite3.connect(f"file:{os.path.expanduser(databasePath)}?mode=ro", uri=True)
    rows = connection.execute(
        "select s.username, m.body from messages m join profiles s on s.id = m.sender_profile_id "
        "join profiles r on r.id = m.recipient_profile_id "
        "where r.username = 'system_adamic_integration' and s.username like 'system_%' order by m.sequence"
    ).fetchall()
    for username, body in rows:
        for branch in branches:
            if branch.split("/", 1)[-1] in body:
                if username == "system_adamic":
                    routed.setdefault(branch, username)
                else:
                    owners.setdefault(branch, username)
    return {**routed, **owners}


def utc(unixSeconds):
    return datetime.datetime.fromtimestamp(int(unixSeconds), datetime.timezone.utc).strftime("%Y-%m-%dT%H:%MZ")


def areasAgainstMain():
    """Each area's standing against main: commits of its own ahead, main's commits it lacks, how many
    of main's landings those are (record commits, the velocity table and meter runs, aren't landings),
    the main it last took, and since when it has lagged (the oldest landing it lacks). An area that
    never catches up hides every break main brings until a whole stack gate finds it."""
    lines = []
    areas = sorted(name.strip() for name in git("for-each-ref", "--format=%(refname:short)", "refs/remotes/origin/area/").split("\n") if name.strip())
    for area in areas:
        ahead = int(git("rev-list", "--count", "--no-merges", f"origin/main..{area}").strip())
        behind = int(git("rev-list", "--count", f"{area}..origin/main").strip())
        tookSha = git("merge-base", area, "origin/main").strip()
        tookShort, tookTime = git("log", "-1", "--format=%h %ct", tookSha).split()
        took = f"{tookShort} {utc(tookTime)}"
        if behind == 0:
            lines.append(f"  {area.removeprefix('origin/')}: holds main, {ahead} commits ahead")
            continue
        missing = [line.split(" ", 2) for line in git("log", "--first-parent", "--reverse", "--format=%h %ct %s", f"{area}..origin/main").splitlines() if line.strip()]
        landings = [entry for entry in missing if not entry[2].startswith("Record ")]
        since = utc((landings or missing)[0][1])
        lines.append(f"  {area.removeprefix('origin/')}: BEHIND main by {behind} commits ({len(landings)} landings), behind since {since}; last took main at {took}; {ahead} commits ahead")
    return lines


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--fresh-hours", type=float, default=1)
    parser.add_argument("--stale-hours", type=float, default=24)
    parser.add_argument("--stalled-hours", type=float, default=2)
    parser.add_argument("--notify-dir", help="write one stalled-branch message per owner here")
    parser.add_argument("--lookback-days", type=int, default=10)
    parser.add_argument("--messages-db")
    parser.add_argument("--tsv")
    parser.add_argument("--delete-merged", action="store_true")
    parser.add_argument("--deleted-log")
    arguments = parser.parse_args()

    if subprocess.run(["git", "fetch", "-q", "origin"], capture_output=True).returncode != 0:
        # Another fetch in this repository may hold the lock; the remote refs from the last one still count.
        print("warning: git fetch failed; using the remote refs already fetched", file=sys.stderr)
    references = git("for-each-ref", "--format=%(refname:short) %(objectname) %(committerdate:unix)", "refs/remotes/origin").split("\n")
    upstreams = []
    branches = []
    for line in references:
        if not line.strip():
            continue
        name, sha, committed = line.split()
        short = name.removeprefix("origin/")
        if short in ("HEAD", "origin") or short.startswith(("gate-logs/", "locks/")):
            # Gate logs are evidence and locks are merges in progress, not work waiting to land.
            continue
        if short == "main" or short.startswith("area/"):
            upstreams.append(name)
        else:
            branches.append((short, sha, int(committed)))

    since = f"--since={arguments.lookback_days}.days"
    landed = set(patchIds([since, *upstreams]).values())
    outside = patchIds([*(f"origin/{short}" for short, _, _ in branches), "--not", *upstreams])

    owners = {}
    if arguments.messages_db:
        owners = ownersFromMessages(arguments.messages_db, [short for short, _, _ in branches])

    directory = os.path.dirname(os.path.abspath(__file__))
    closed = readTable(os.path.join(directory, "census-closed.tsv"))
    ownerOverrides = readTable(os.path.join(directory, "census-owners.tsv"))
    now = datetime.datetime.now(datetime.timezone.utc).timestamp()
    rows = []
    distinctUnlanded = set()
    distinctSuperseded = set()
    for short, sha, committed in branches:
        commits = git("rev-list", "--no-merges", f"origin/{short}", "--not", *upstreams).split()
        unlanded = [commit for commit in commits if outside.get(commit) not in landed]
        age = (now - committed) / 3600
        tipLanded = any(subprocess.run(["git", "merge-base", "--is-ancestor", sha, upstream]).returncode == 0 for upstream in upstreams)
        if tipLanded:
            category = "merged"
        elif not commits:
            category = "merges-only"
        elif short in closed:
            category = "closed"
        elif not unlanded:
            category = "superseded"
        elif age < arguments.fresh_hours:
            category = "in-flight"
        elif age < arguments.stalled_hours:
            category = "unmerged"
        elif age < arguments.stale_hours:
            category = "stalled"
        else:
            category = "abandoned"
        owner = ownerOverrides.get(short) or next((circle for prefix, circle in prefixOwners if short.startswith(prefix)), None) or owners.get(short, "unknown")
        rows.append((category, owner, short, sha, age, len(commits), len(unlanded)))
        distinctUnlanded.update(unlanded)
        if category == "superseded":
            distinctSuperseded.update(commits)

    if arguments.tsv:
        with open(arguments.tsv, "w") as file:
            file.write("class\towner\tbranch\ttip\tage_hours\tcommits_outside\tpatches_unlanded\n")
            for row in rows:
                file.write("\t".join([row[0], row[1], row[2], row[3][:8], f"{row[4]:.1f}", str(row[5]), str(row[6])]) + "\n")

    byClass = collections.Counter(row[0] for row in rows)
    print(f"{len(rows)} branches outside main and the areas: " + ", ".join(f"{byClass[c]} {c}" for c in ("merged", "merges-only", "superseded", "in-flight", "unmerged", "stalled", "abandoned", "closed")))
    print(f"distinct commits on superseded branches, already landed as patches: {len(distinctSuperseded)}")
    print(f"distinct commits outside main and the areas whose patches haven't landed: {len(distinctUnlanded)}")
    print()
    print("areas against main (ahead: the area's own commits; behind: main's commits the area lacks):")
    for line in areasAgainstMain():
        print(line)
    print()
    byOwner = collections.defaultdict(collections.Counter)
    for row in rows:
        byOwner[row[1]][row[0]] += 1
    for owner in sorted(byOwner):
        counts = byOwner[owner]
        print(f"{owner}: " + ", ".join(f"{counts[c]} {c}" for c in ("merged", "merges-only", "superseded", "in-flight", "unmerged", "stalled", "abandoned", "closed") if counts[c]))
    print()
    if arguments.notify_dir:
        os.makedirs(arguments.notify_dir, exist_ok=True)
        for owner in sorted(byOwner):
            stalled = sorted((row for row in rows if row[1] == owner and row[0] == "stalled"), key=lambda row: -row[4])
            if not stalled or owner == "unknown":
                continue
            with open(os.path.join(arguments.notify_dir, f"{owner}.txt"), "w") as file:
                file.write(f"Stalled branches (branch census): yours, with unlanded work and no commit for {arguments.stalled_hours:g} hours or more, oldest first. Each is either still being worked (say so), finished (land it through your area), or dead (say so and it's closed). Branch, tip, hours since its last commit, unlanded commits, last commit:\n")
                for row in stalled:
                    subject = git("log", "-1", "--format=%s", row[3]).strip()
                    file.write(f"- {row[2]} {row[3][:8]} {row[4]:.1f}h {row[6]} | {subject}\n")
    print("unmerged and stalled, by owner, oldest first (branch, tip, hours since last commit, unlanded patches):")
    for owner in sorted(byOwner):
        unmerged = sorted((row for row in rows if row[1] == owner and row[0] in ("unmerged", "stalled")), key=lambda row: -row[4])
        if unmerged:
            print(f"  {owner}")
            for row in unmerged:
                print(f"    {row[2]} {row[3][:8]} {row[4]:.1f}h {row[6]}")

    if arguments.delete_merged:
        merged = [row for row in rows if row[0] == "merged"]
        deleted = []
        for row in merged:
            result = subprocess.run(["git", "push", "-q", f"--force-with-lease=refs/heads/{row[2]}:{row[3]}", "origin", f":refs/heads/{row[2]}"], capture_output=True, text=True)
            if result.returncode == 0:
                deleted.append(row)
            else:
                print(f"left {row[2]}: {result.stderr.strip().splitlines()[-1] if result.stderr.strip() else 'push refused'}", file=sys.stderr)
        stamp = datetime.datetime.now(datetime.timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
        if arguments.deleted_log:
            with open(arguments.deleted_log, "a") as file:
                for row in deleted:
                    file.write(f"{stamp}\tdeleted\t{row[2]}\t{row[3]}\t{row[1]}\n")
        print(f"deleted {len(deleted)} of {len(merged)} merged branches")


if __name__ == "__main__":
    sys.exit(main())

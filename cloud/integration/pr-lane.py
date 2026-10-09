#!/usr/bin/env python3
"""The test-only lane for pull requests: a pull request into main lands through push-main.sh --test-only
or not at all (@system_adamic, Oct 9 03:28Z, after pull requests 1 to 5 merged straight to main with
nothing checking their paths). Nobody merges a pull request on GitHub by hand: open it, and within a
minute this lands its head when every path it changes against main is one the lane takes, or comments
push-main's refusal on it once and leaves it open. A new push to the branch is tried again.

push-main lands the head itself or a merge holding it, so GitHub marks the pull request merged.

A merge made on GitHub anyway can't be refused (every node pushes as the same admin account, so a
required check binds push-main too or nobody), so each run also reads every commit GitHub put on main's
first-parent line since the last run and tells integration and @system_adamic at once when one changed
anything the lane wouldn't take.

Branches sent to integration by message take the same lane: a line "<branch><TAB><note>" in
~/.adamic-pr-lane/branches lands when it can and leaves the file; a refusal leaves it too and tells
integration. A change push-main holds for the star (exit 3) stays, for either kind, until the star's
gate finishes.

usage (launchd com.adamic.pr-lane runs it every minute, from the merge tree): cloud/integration/pr-lane.py
"""
import json, os, re, subprocess, sys

repository = "system-inc/adamic"
ahra = os.path.expanduser("~/Projects/ahra")
# push-main.sh's testOnlyPattern.
testOnlyPaths = re.compile(r"(_test\.go$|_test\.py$|(^|/)test_[^/]*\.py$|/testdata/|^review/|(^|/)shards\.json$|^stage3/fixtures/|^stage3/meter/)")
directory = os.path.dirname(os.path.abspath(__file__))
state = os.path.expanduser("~/.adamic-pr-lane")
os.makedirs(state, exist_ok=True)


def run(*arguments):
    return subprocess.run(arguments, capture_output=True, text=True)


def audit():
    # Every commit GitHub made on main's first-parent line since the last audit, checked by the lane's rule.
    run("git", "fetch", "-q", "origin", "main")
    main = run("git", "rev-parse", "origin/main").stdout.strip()
    marker = os.path.join(state, "audited")
    since = open(marker).read().strip() if os.path.exists(marker) else ""
    if since and run("git", "merge-base", "--is-ancestor", since, main).returncode == 0:
        for line in run("git", "log", "--first-parent", "--format=%H %cn", f"{since}..{main}").stdout.splitlines():
            sha, committer = line.split(" ", 1)
            if committer != "GitHub":
                continue
            outside = [path for path in run("git", "diff", "--name-only", f"{sha}^1", sha).stdout.split() if not testOnlyPaths.search(path)]
            if outside:
                note = f"A pull request merged on GitHub put non-test paths on main with no gate: {sha[:8]} ({run('git', 'log', '-1', '--format=%s', sha).stdout.strip()[:120]}): {' '.join(outside[:5])}. Revert or gate it; pull requests land only through pr-lane.py."
                for recipient in ("system_adamic_integration", "system_adamic"):
                    subprocess.run(["./node_modules/.bin/ahra", "os", "send", recipient, note], cwd=ahra, capture_output=True)
                print(f"alarm {sha[:8]}: {' '.join(outside[:5])}")
    with open(marker, "w") as handle:
        handle.write(main)


audit()


listed = run("gh", "pr", "list", "--repo", repository, "--base", "main", "--state", "open", "--json", "number,headRefName,headRefOid,title,isDraft")
if listed.returncode != 0:
    sys.exit(f"gh pr list failed: {listed.stderr.strip()}")
for pullRequest in json.loads(listed.stdout):
    if pullRequest["isDraft"]:
        continue
    number, sha = pullRequest["number"], pullRequest["headRefOid"]
    refused = os.path.join(state, f"refused-{number}-{sha[:12]}")
    if os.path.exists(refused):
        continue
    run("git", "fetch", "-q", "origin", f"+refs/pull/{number}/head:refs/remotes/origin/pull/{number}")
    note = f"pull request #{number} {pullRequest['headRefName']} {sha[:8]} ({pullRequest['title'].replace(',', ';')})"
    pushed = run("bash", os.path.join(directory, "push-main.sh"), "--test-only", sha, note)
    with open(os.path.join(state, f"push-{number}-{sha[:12]}.log"), "w") as handle:
        handle.write(pushed.stdout + pushed.stderr)
    if pushed.returncode == 0:
        landed = [line for line in pushed.stdout.splitlines() if line.startswith("Pushed main")]
        run("gh", "pr", "comment", str(number), "--repo", repository, "--body", f"Landed by integration's test-only lane: {landed[0] if landed else sha}")
        print(f"landed #{number} {sha[:8]}")
        continue
    output = pushed.stdout + pushed.stderr
    if pushed.returncode == 3:
        print(f"held #{number} {sha[:8]}: {output.strip().splitlines()[-1]}")
        continue
    # Main moved between the check and the push (another landing won the race): the next minute tries again.
    if "cannot lock ref" in output or "[rejected]" in output or "failed to push" in output:
        print(f"raced #{number} {sha[:8]}, retrying next run")
        continue
    reason = (pushed.stderr.strip() or pushed.stdout.strip()).splitlines()[-1:]
    run("gh", "pr", "comment", str(number), "--repo", repository, "--body",
        f"Not landed: integration's test-only lane takes only tests, testdata, review evidence, shard tables and the ruled stage 3 harness, and push-main said: {reason[0] if reason else 'no output'}. Push a fix to the branch and it's tried again, or send the branch to @system_adamic_integration for a gate.")
    open(refused, "w").close()
    print(f"refused #{number} {sha[:8]}: {reason[0] if reason else 'no output'}")


# Branches sent by message, in the order they were queued.
queued = os.path.join(state, "branches")
if os.path.exists(queued):
    kept = []
    for line in open(queued).read().splitlines():
        if not line.strip():
            continue
        branch, _, note = line.partition("\t")
        run("git", "fetch", "-q", "origin", branch)
        sha = run("git", "rev-parse", f"origin/{branch}").stdout.strip()
        pushed = run("bash", os.path.join(directory, "push-main.sh"), "--test-only", sha, f"{note or branch} ({branch} {sha[:8]})")
        output = pushed.stdout + pushed.stderr
        if pushed.returncode == 0:
            landed = [entry for entry in pushed.stdout.splitlines() if entry.startswith("Pushed main")]
            print(f"landed {branch} {sha[:8]}")
            subprocess.run(["./node_modules/.bin/ahra", "os", "send", "system_adamic_integration", f"pr-lane landed queued {branch}: {landed[0][:300] if landed else sha[:8]}"], cwd=ahra, capture_output=True)
        elif pushed.returncode == 3 or "cannot lock ref" in output or "[rejected]" in output or "failed to push" in output:
            kept.append(line)
        else:
            reason = (pushed.stderr.strip() or pushed.stdout.strip()).splitlines()[-1:]
            print(f"refused {branch} {sha[:8]}: {reason[0] if reason else 'no output'}")
            subprocess.run(["./node_modules/.bin/ahra", "os", "send", "system_adamic_integration", f"pr-lane refused queued {branch} {sha[:8]}: {reason[0][:300] if reason else 'no output'}"], cwd=ahra, capture_output=True)
    with open(queued, "w") as handle:
        handle.write("".join(entry + "\n" for entry in kept))

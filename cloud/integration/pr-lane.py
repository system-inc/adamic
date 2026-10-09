#!/usr/bin/env python3
"""The test-only lane for pull requests: a pull request into main lands through push-main.sh --test-only
or not at all (@system_adamic, Oct 9 03:28Z, after pull requests 1 to 5 merged straight to main with
nothing checking their paths). Nobody merges a pull request on GitHub by hand: open it, and within a
minute this lands its head when every path it changes against main is one the lane takes, or comments
push-main's refusal on it once and leaves it open. A new push to the branch is tried again.

push-main lands the head itself or a merge holding it, so GitHub marks the pull request merged.

usage (launchd com.adamic.pr-lane runs it every minute, from the merge tree): cloud/integration/pr-lane.py
"""
import json, os, subprocess, sys

repository = "system-inc/adamic"
directory = os.path.dirname(os.path.abspath(__file__))
state = os.path.expanduser("~/.adamic-pr-lane")
os.makedirs(state, exist_ok=True)


def run(*arguments):
    return subprocess.run(arguments, capture_output=True, text=True)


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
    # Main moved between the check and the push (another landing won the race): the next minute tries again.
    if "cannot lock ref" in output or "[rejected]" in output or "failed to push" in output:
        print(f"raced #{number} {sha[:8]}, retrying next run")
        continue
    reason = (pushed.stderr.strip() or pushed.stdout.strip()).splitlines()[-1:]
    run("gh", "pr", "comment", str(number), "--repo", repository, "--body",
        f"Not landed: integration's test-only lane takes only tests, testdata, review evidence, shard tables and the ruled stage 3 harness, and push-main said: {reason[0] if reason else 'no output'}. Push a fix to the branch and it's tried again, or send the branch to @system_adamic_integration for a gate.")
    open(refused, "w").close()
    print(f"refused #{number} {sha[:8]}: {reason[0] if reason else 'no output'}")

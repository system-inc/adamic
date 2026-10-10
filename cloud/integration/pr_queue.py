#!/usr/bin/env python3
"""A pull request's way through Loom's front door (#shynynn, Loom's ruling Oct 10 01:25Z): the pr lane submits each
test-only pull request merged onto main's tip, so its future is exactly what the lander fast-forwards main to.

merged() writes that commit: the head itself when it already descends from main, else a merge whose first parent is
main's tip and whose second is the head (git merge-tree and commit-tree, nothing checked out), or nothing on a conflict,
with the paths. publish() puts it on its own branch, cloud/pr-queue-<number>-<sha8>, since the queue's git facts need
the sha on GitHub; one branch per sha, so no push ever rewrites one.

candidate() is the whole decision for one head, with no network: submit the landing commit, or why not. submit() posts
it with a submit token minted for the lane's owner, the way the wire verifies it (loom wire/source/Token.ts).

usage: imported by pr-lane.py
"""
import base64, hashlib, hmac, json, subprocess, time, urllib.error, urllib.request

pipeline = "https://loom-pipeline.kirk-ouimet.workers.dev"


def git(repository, *arguments):
    return subprocess.run(["git", "-C", repository, *arguments], capture_output=True, text=True)


def merged(repository, main, head, message):
    """The commit that lands head on main and the conflicting paths: (sha, []), (None, paths) on a conflict, or
    (None, []) when main already holds head."""
    if git(repository, "merge-base", "--is-ancestor", head, main).returncode == 0:
        return None, []
    if git(repository, "merge-base", "--is-ancestor", main, head).returncode == 0:
        return head, []
    written = git(repository, "merge-tree", "--write-tree", "--name-only", "--no-messages", main, head)
    lines = written.stdout.splitlines()
    if written.returncode == 1:
        return None, [line for line in lines[1:] if line]
    if written.returncode != 0 or not lines:
        raise RuntimeError("git merge-tree %s %s: %s" % (main[:12], head[:12], written.stderr.strip()))
    committed = git(repository, "commit-tree", lines[0], "-p", main, "-p", head, "-m", message)
    if committed.returncode != 0:
        raise RuntimeError("git commit-tree: %s" % committed.stderr.strip())
    return committed.stdout.strip(), []


def candidate(repository, main, number, head, testOnly):
    """What the lane does with pull request number's head on main: {action: submit, sha, body} (body is POST /changes'
    without its owner), {action: held} when main already holds it, {action: conflict, paths}, or {action: outside,
    paths} when the landing commit changes a path testOnly refuses."""
    sha, conflicts = merged(repository, main, head, "Land pull request #%s (%s) over main %s" % (number, head[:8], main[:8]))
    if conflicts:
        return {"action": "conflict", "paths": conflicts}
    if sha is None:
        return {"action": "held"}
    paths = sorted(path for path in git(repository, "diff", "--no-renames", "--name-only", main, sha).stdout.splitlines() if path)
    outside = [path for path in paths if not testOnly(path)]
    if outside or not paths:
        return {"action": "outside", "paths": outside}
    return {"action": "submit", "sha": sha, "body": {"sha": sha, "base": main, "paths": paths}}


def token(secret, owner):
    """A submit token for owner, good for ten minutes."""
    payload = base64.urlsafe_b64encode(json.dumps({"run": owner, "scope": "submit", "expires": int(time.time()) + 600},
                                                  separators=(",", ":")).encode()).rstrip(b"=")
    return (payload + b"." + base64.urlsafe_b64encode(hmac.new(secret.strip().encode(), payload, hashlib.sha256).digest()).rstrip(b"=")).decode()


def submit(body, owner, secret):
    """POST /changes as owner: (status, answer)."""
    request = urllib.request.Request(pipeline + "/changes", data=json.dumps(dict(body, owner=owner)).encode(), method="POST",
                                     headers={"Authorization": "Bearer " + token(secret, owner), "Content-Type": "application/json",
                                              "User-Agent": "adamic-pr-lane"})
    try:
        with urllib.request.urlopen(request, timeout=30) as response:
            return response.status, json.loads(response.read() or b"null")
    except urllib.error.HTTPError as error:
        text = error.read()
        try:
            return error.code, json.loads(text)
        except ValueError:
            return error.code, {"error": text.decode(errors="replace")[:300]}


def throughQueue(mode, repository, main, number, head, testOnly, owner, secret, post=None):
    """One head through the front door. mode "check" decides and touches nothing outside this clone (the proof before
    cutover); "on" publishes the landing commit and submits it. The candidate, plus {status, answer} when it posted."""
    found = candidate(repository, main, number, head, testOnly)
    if mode != "on" or found["action"] != "submit":
        return found
    publish(repository, found["sha"], number)
    status, answer = (post or submit)(found["body"], owner, secret)
    return dict(found, status=status, answer=answer)


def publish(repository, sha, number):
    """Pushes sha to its own branch on origin and returns the branch, or raises with git's words."""
    branch = "cloud/pr-queue-%s-%s" % (number, sha[:8])
    pushed = git(repository, "push", "-q", "origin", "%s:refs/heads/%s" % (sha, branch))
    if pushed.returncode != 0:
        raise RuntimeError("git push %s: %s" % (branch, pushed.stderr.strip()))
    return branch

#!/usr/bin/env python3
"""A pull request's way through Loom's front door (#shynynn, Loom's ruling Oct 10 01:25Z): the pr lane submits each
test-only pull request merged onto main's tip, so its future is exactly what the lander fast-forwards main to.

merged() writes that commit: the head itself when it already descends from main, else a merge whose first parent is
main's tip and whose second is the head (git merge-tree and commit-tree, nothing checked out), or nothing on a conflict,
with the paths. publish() puts it on its own branch, cloud/pr-queue-<number>-<sha8>, since the queue's git facts need
the sha on GitHub; one branch per sha, so no push ever rewrites one.

usage: imported by pr-lane.py
"""
import subprocess


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


def publish(repository, sha, number):
    """Pushes sha to its own branch on origin and returns the branch, or raises with git's words."""
    branch = "cloud/pr-queue-%s-%s" % (number, sha[:8])
    pushed = git(repository, "push", "-q", "origin", "%s:refs/heads/%s" % (sha, branch))
    if pushed.returncode != 0:
        raise RuntimeError("git push %s: %s" % (branch, pushed.stderr.strip()))
    return branch

#!/usr/bin/env python3
"""Checks area-route.py: fixed cases, then every unlanded codex/* branch on origin.

usage (from the repository): python3 cloud/integration/area-route-test.py
A family is the unlanded codex/* branches sharing the name's first word (codex/notyet-*, codex/views-*).
A family with two or more branches that fall through to "no prefix and no fleet" fails the test by
name, so a new family of worker branches can't reach the automatic merges undecided. A first word that
only one undecided branch carries (codex/async-typeof beside a fleet-routed codex/async-*) isn't a
family: that branch is held and reported at merge time.
"""
import collections
import importlib.util
import subprocess
import sys
from pathlib import Path

directory = Path(__file__).resolve().parent
specification = importlib.util.spec_from_file_location("areaRoute", directory / "area-route.py")
areaRoute = importlib.util.module_from_spec(specification)
specification.loader.exec_module(areaRoute)

fixed = [
    ("codex/stage1-lint-batch2", "stage1-lint"),
    ("codex/stage1-markdown-x", "stage1-format"),
    ("codex/lint-rules-x", "stage1-lint"),
    ("codex/notyet-binary", "compiler"),
    ("codex/stricter-indexed-b", "compiler"),
    ("codex/enum-tag-narrowing-2", "compiler"),
    ("codex/host-promises", "runtime"),
    ("codex/host-anything", "library"),
    ("codex/stage3-anything", "stage3"),
    ("codex/views-anything", "hold"),
    ("codex/no-such-family-ever-2f3a", "hold"),
]


def git(*arguments):
    return subprocess.run(["git", *arguments], capture_output=True, text=True, check=True).stdout


def main():
    failures = []
    for branch, want in fixed:
        got, how = areaRoute.route(branch)
        if got != want:
            failures.append(f"fixed case {branch}: want {want}, got {got} ({how})")

    git("fetch", "-q", "origin", "+refs/heads/main:refs/remotes/origin/main", "+refs/heads/codex/*:refs/remotes/origin/codex/*")
    unlanded = []
    for line in git("for-each-ref", "--format=%(refname:short) %(objectname)", "refs/remotes/origin/codex/").splitlines():
        name, sha = line.split()
        if subprocess.run(["git", "merge-base", "--is-ancestor", sha, "refs/remotes/origin/main"]).returncode:
            unlanded.append(name.removeprefix("origin/"))
    families = collections.defaultdict(list)
    for branch in unlanded:
        families[branch.removeprefix("codex/").split("-")[0]].append(branch)

    routed = held = 0
    undecided = collections.defaultdict(list)
    for family, branches in sorted(families.items()):
        for branch in branches:
            area, how = areaRoute.route(branch)
            if area != "hold":
                routed += 1
            elif how.startswith("no prefix"):
                undecided[family].append(branch)
            else:
                held += 1
    for family, branches in sorted(undecided.items()):
        if len(branches) > 1:
            failures.append(f"family codex/{family}-* has {len(branches)} undecided branches ({', '.join(branches)}); add it to area-routes.tsv")
    alone = sum(len(branches) for branches in undecided.values() if len(branches) == 1)
    print(f"{len(unlanded)} unlanded codex/* branches: {routed} routed, {held} held by a named row, {alone} undecided one-offs (held and reported at merge time)")
    for failure in failures:
        print("FAIL " + failure)
    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main())

#!/usr/bin/env python3
"""The star's merge train (Kirk and @system_adamic, October 8).

The star's chain is a list of slices in star-train.tsv, each a source branch its owner pushes. Each
slice gates on top of the one before it, as if that one had already landed: slice 1 merges onto main,
slice n onto slice n-1's candidate. A candidate's name carries its source and base shas
(cloud/land-train-<n>-<slug>-<source8>-<base8>), so a new source commit or a new main builds a new
candidate above it and nothing is ever force-pushed; the superseded ones go on the watcher's skip list.

The bottom candidate lands itself: when its fast gate is green and main's newest whole gate on the
current main is green or explained (main-reds.tsv, rows starting OPEN don't count), this runs
push-main.sh, whose checks are the verdict (deferred tests, records, base still main). Integration
stays on reds and conflicts. One run per invocation; launchd runs it every minute.

usage: cloud/integration/star-train.py [--dry-run]
"""

import json, os, subprocess, sys, time

directory = os.path.dirname(os.path.abspath(__file__))
dryRun = "--dry-run" in sys.argv
watch = os.path.expanduser("~/.adamic-fast-gate-watch")
state = os.path.expanduser("~/.adamic-star-train")
ahra = os.path.expanduser("~/Projects/ahra")
os.makedirs(state, exist_ok=True)


def git(*arguments, check=True):
    return subprocess.run(["git", *arguments], capture_output=True, text=True, check=check).stdout.strip()


def log(line):
    stamp = time.strftime("%H:%M:%SZ", time.gmtime())
    with open(os.path.join(state, "train.log"), "a") as handle:
        handle.write(f"{stamp} {line}\n")
    print(line)


def tell(key, recipient, message):
    # Each distinct event is said once, however many minutes it stays true.
    marker = os.path.join(state, "told-" + key.replace("/", "_"))
    if os.path.exists(marker) or dryRun:
        return
    subprocess.run(["./node_modules/.bin/ahra", "os", "send", recipient, message], cwd=ahra, capture_output=True)
    open(marker, "w").close()


def slices():
    rows = []
    for line in open(os.path.join(directory, "star-train.tsv")):
        if line.strip() and not line.startswith("#"):
            slug, source, owner = line.rstrip("\n").split("\t")[:3]
            rows.append((slug, source, owner))
    return rows


def remoteHeads(pattern):
    heads = {}
    for line in git("ls-remote", "origin", f"refs/heads/{pattern}").splitlines():
        sha, reference = line.split("\t")
        heads[reference.removeprefix("refs/heads/")] = sha
    return heads


def newestLog(sha, kind):
    names = sorted(remoteHeads(f"gate-logs/{sha[:12]}/*/{kind}"))
    for name in reversed(names):
        git("fetch", "-q", "origin", f"+refs/heads/{name}:refs/remotes/origin/{name}")
        status = git("show", f"origin/{name}:status.txt", check=False).split("\n")[0]
        if status:
            return name, status
    return None, ""


def explained(reference):
    for line in open(os.path.join(directory, "main-reds.tsv")):
        if line.startswith(reference + "\t") and not line.split("\t", 1)[1].startswith("OPEN"):
            return True
    return False


def appendOnce(path, line):
    existing = open(path).read().split("\n") if os.path.exists(path) else []
    if line not in existing:
        with open(path, "a") as handle:
            handle.write(line + "\n")


git("fetch", "-q", "origin")
main = git("rev-parse", "origin/main")
base = main
chain = []
trainHeads = remoteHeads("cloud/land-train-*")
for number, (slug, source, owner) in enumerate(slices(), start=1):
    sourceSha = git("rev-parse", "-q", "--verify", f"origin/{source}", check=False)
    if not sourceSha:
        break
    if subprocess.run(["git", "merge-base", "--is-ancestor", sourceSha, base]).returncode == 0:
        continue  # already landed, or carried by the slice below
    name = f"cloud/land-train-{number}-{slug}-{sourceSha[:8]}-{base[:8]}"
    if name in trainHeads:
        candidate = trainHeads[name]
    else:
        merged = subprocess.run(["git", "merge-tree", "--write-tree", base, sourceSha], capture_output=True, text=True)
        if merged.returncode != 0:
            files = [line for line in merged.stdout.splitlines()[1:] if line and not line.startswith(("Auto-merging", "CONFLICT"))]
            log(f"conflict: {source} {sourceSha[:8]} onto {base[:8]}: {' '.join(files[:6])}")
            tell(f"conflict-{name}", owner, f"The star's train can't stack {source} {sourceSha[:8]} onto {base[:8]} (slice {number}, {slug}): it conflicts in {', '.join(files[:6])}. Please merge {base[:8]} into {source}, both sides kept, and push; the train restacks on its own.")
            tell(f"conflict-{name}", "system_adamic_integration", f"Train conflict at slice {number} ({slug}): {source} {sourceSha[:8]} onto {base[:8]}, sent to @{owner}.")
            break
        tree = merged.stdout.splitlines()[0]
        message = (f"Merge {source} at {sourceSha[:8]} onto {base[:8]}, slice {number} ({slug}) of the star's train\n\n"
                   f"Gates as if every slice below had landed; git merge-tree was clean.\n\nGate-runs: deferred")
        candidate = git("commit-tree", tree, "-p", base, "-p", sourceSha, "-m", message)
        if not dryRun:
            git("push", "-q", "origin", f"{candidate}:refs/heads/{name}")
            appendOnce(os.path.join(watch, "priority"), name)
        log(f"built {name} {candidate[:12]}")
        for other, otherSha in trainHeads.items():
            if other.startswith(f"cloud/land-train-{number}-") and other != name and not dryRun:
                appendOnce(os.path.join(watch, "skip"), f"{other} {otherSha}")
    chain.append((number, slug, source, owner, name, candidate, base))
    base = candidate

if not chain:
    sys.exit(0)

number, slug, source, owner, name, candidate, candidateBase = chain[0]
# An owner can hold a source commit from landing (it still gates): a file hold-<source8> in the state
# directory, holding why. A new commit on the source is a new candidate, so the hold lapses with it.
sourceSha = git("rev-parse", f"origin/{source}")
hold = os.path.join(state, f"hold-{sourceSha[:8]}")
if os.path.exists(hold):
    log(f"held: {name} ({open(hold).read().strip()})")
    sys.exit(0)
reference, status = newestLog(candidate, "fast")
if not status.startswith("green"):
    if status.startswith("red"):
        log(f"red: {name} {reference}")
        tell(f"red-{candidate}", "system_adamic_integration", f"Train slice {number} ({slug}) is red: {reference}. Everything above it waits for the fix on {source}.")
    sys.exit(0)

mainLog, mainStatus = newestLog(main, "full-main")
if not mainLog or mainStatus.startswith("running") or not (mainStatus.startswith("green") or explained(mainLog)):
    log(f"waiting: {name} is green, main {main[:8]}'s whole gate is {mainStatus.split(':')[0] or 'not started'}")
    sys.exit(0)

log(f"landing {name} on {reference}")
if dryRun:
    sys.exit(0)
note = f"the star's train, slice {number} ({slug}): {source} {git('rev-parse', 'origin/' + source)[:8]}"
pushed = subprocess.run(["bash", os.path.join(directory, "push-main.sh"), "--fast-gate", reference, candidate, note], capture_output=True, text=True)
with open(os.path.join(state, f"push-{candidate[:12]}.log"), "w") as handle:
    handle.write(pushed.stdout + pushed.stderr)
if pushed.returncode == 0:
    landed = [line for line in pushed.stdout.splitlines() if line.startswith("Pushed main")]
    log(f"landed {name}")
    tell(f"landed-{candidate}", "system_adamic", f"Train landed slice {number} ({slug}) by itself: {landed[0][:400] if landed else candidate[:12]}. Main's whole gate on the new main runs before the next slice lands.")
    tell(f"landed-{candidate}", owner, f"Your slice {number} ({slug}) landed: {landed[0][:300] if landed else candidate[:12]}.")
else:
    refusal = (pushed.stderr.strip() or pushed.stdout.strip()).splitlines()[-1:]
    log(f"refused {name}: {refusal}")
    tell(f"refused-{candidate}", "system_adamic_integration", f"push-main refused train slice {number} ({slug}) {candidate[:12]}: {refusal[0][:400] if refusal else 'no output'}")

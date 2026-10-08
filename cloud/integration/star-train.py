#!/usr/bin/env python3
"""The star's merge train (Kirk and @system_adamic, October 8).

The star's chain is a list of slices in star-train.tsv, each a source branch its owner pushes. Each
slice gates on top of the one before it, as if that one had already landed: slice 1 merges onto main,
slice n onto slice n-1's candidate. A candidate's name carries its source and base shas
(cloud/land-train-<n>-<slug>-<source8>-<base8>), so a new source commit or a new main builds a new
candidate above it and nothing is ever force-pushed; the superseded ones go on the watcher's skip list.

The bottom candidate lands itself: when its own whole gate (full-main on its sha) is green and main's newest whole gate on the
current main (or on the tree main moved past only by record commits) is green or explained (main-reds.tsv, rows starting OPEN don't count), this runs
push-main.sh, whose checks are the verdict (deferred tests, records, base still main). Integration
stays on reds and conflicts. One run per invocation; launchd runs it every minute.

usage: cloud/integration/star-train.py [--dry-run]
"""

import importlib.util, json, os, subprocess, sys, time

directory = os.path.dirname(os.path.abspath(__file__))
dryRun = "--dry-run" in sys.argv
watch = os.path.expanduser("~/.adamic-fast-gate-watch")
state = os.path.expanduser("~/.adamic-star-train")
requests = os.path.expanduser("~/.adamic-full-gate/requests")
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
            fields = line.rstrip("\n").split("\t")
            # A fourth column names riders: test-only branches merged in below the slice's source, so
            # they land on the slice's whole gate instead of restarting it (@system_adamic, October 8).
            rows.append((fields[0], fields[1], fields[2], fields[3].split() if len(fields) > 3 else []))
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


specification = importlib.util.spec_from_file_location("recordReaders", os.path.join(directory, "record-paths-test.py"))
recordReaders = importlib.util.module_from_spec(specification)
specification.loader.exec_module(recordReaders)


def recordsOnly(older, newer):
    # True when newer's tree differs from older's only in record paths: the velocity row push-main
    # commits on every landing, meter runs and stage 3's progress record, the same paths push-main
    # lets a landing ride over (@system_adamic, October 7).
    changed = git("diff", "--name-only", older, newer).split()
    if not all(path == "documentation/velocity/landings.csv" or path == "stage3/progress.json"
               or path.startswith("stage3/meter/runs/") for path in changed):
        return False
    # Honest only while no test or build reads a record (record-paths-test.py, @system_adamic's
    # condition): a reader means the gate has to see the record change, so the slice rebuilds.
    unexplained = recordReaders.readers(newer) if changed else []
    if unexplained:
        log(f"record reader in {newer[:8]}: {unexplained[0]}")
        tell(f"record-reader-{newer[:12]}", "system_adamic_integration", f"record-paths-test.py fails on {newer[:8]}, so the train rebuilds over record commits instead of keeping gates: {'; '.join(unexplained[:3])}")
        return False
    return True


def mainConfirmation(sha):
    # Main's whole gate: its own newest finished record, or, when main moved past a gated tree only by
    # record commits, that tree's record. A slice that landed on its own full gate leaves main one
    # velocity commit past it, and that record is main's confirmation too (--full-gate, @system_adamic,
    # October 8), so the next slice doesn't wait a whole gate for a row in a table no test reads.
    seen, frontier = set(), [sha]
    while frontier:
        current = frontier.pop(0)
        if current in seen:
            continue
        seen.add(current)
        for name in reversed(sorted(remoteHeads(f"gate-logs/{current[:12]}/*/full-main"))):
            git("fetch", "-q", "origin", f"+refs/heads/{name}:refs/remotes/origin/{name}")
            status = git("show", f"origin/{name}:status.txt", check=False).split("\n")[0]
            if status.startswith(("green", "red")):
                return name, status
        frontier += [parent for parent in git("rev-list", "--parents", "-n", "1", current).split()[1:] if recordsOnly(parent, current)]
    return newestLog(sha, "full-main")


def removeLine(path, line):
    if os.path.exists(path):
        existing = open(path).read().split("\n")
        if line in existing:
            with open(path, "w") as handle:
                handle.write("\n".join(kept for kept in existing if kept != line))


def riderGreenAlone(rider, riderSha):
    # The rider's own fast gate on main: the rider itself when it holds main, otherwise main merged
    # with it, pushed once as cloud/land-rider-<sha8> so the watcher gates it, and remembered in a
    # state file so the train reads that candidate's verdict on later runs.
    if subprocess.run(["git", "merge-base", "--is-ancestor", main, riderSha]).returncode == 0:
        gated = riderSha
    else:
        record = os.path.join(state, f"rider-gate-{riderSha[:12]}")
        gated = open(record).read().strip() if os.path.exists(record) else ""
        if not gated:
            merged = subprocess.run(["git", "merge-tree", "--write-tree", "--name-only", main, riderSha], capture_output=True, text=True)
            if merged.returncode != 0:
                tell(f"rider-alone-conflict-{riderSha[:12]}", "system_adamic_integration", f"Rider {rider} {riderSha[:8]} doesn't merge with main {main[:8]}; it can't gate alone, so it doesn't ride.")
                return False
            gated = git("commit-tree", merged.stdout.splitlines()[0], "-p", main, "-p", riderSha, "-m",
                        f"Merge {rider} at {riderSha[:8]} onto main {main[:8]}, gating alone before it rides the star's train")
            if dryRun:
                return False
            git("push", "-q", "origin", f"{gated}:refs/heads/cloud/land-rider-{riderSha[:8]}")
            with open(record, "w") as handle:
                handle.write(gated + "\n")
            log(f"rider gating alone: {rider} {riderSha[:8]} as cloud/land-rider-{riderSha[:8]} {gated[:12]}")
    reference, status = newestLog(gated, "fast")
    if status.startswith("red"):
        log(f"rider red alone: {rider} {riderSha[:8]} {reference}; it doesn't ride")
        tell(f"rider-red-{riderSha[:12]}", "system_adamic_integration", f"Rider {rider} {riderSha[:8]} is red on its own fast gate ({reference}), so it doesn't ride the star.")
        return False
    return status.startswith("green")


def appendOnce(path, line):
    existing = open(path).read().split("\n") if os.path.exists(path) else []
    if line not in existing:
        with open(path, "a") as handle:
            handle.write(line + "\n")


git("fetch", "-q", "origin")
main = git("rev-parse", "origin/main")
base = main
# A landed slice's request is spent: its record published before it landed, so a line left behind
# would spend Home's whole gate on a tree that's already main's past.
if not dryRun and os.path.exists(requests):
    for line in open(requests).read().split():
        if subprocess.run(["git", "merge-base", "--is-ancestor", line, main], capture_output=True).returncode == 0:
            removeLine(requests, line)
chain = []
trainHeads = remoteHeads("cloud/land-train-*")
for number, (slug, source, owner, riderBranches) in enumerate(slices(), start=1):
    # A source is a branch, or branch@trailer: the newest commit on that branch whose message carries
    # "Train-slice: <slug>", so slices built stacked on one rehearsal branch each become their own car.
    if "@" in source:
        branch, _, trailer = source.partition("@")
        sourceSha = git("log", "-1", "--format=%H", f"--grep=^Train-slice: {trailer}$", f"origin/{branch}", check=False)
    else:
        sourceSha = git("rev-parse", "-q", "--verify", f"origin/{source}", check=False)
    if not sourceSha:
        break
    if subprocess.run(["git", "merge-base", "--is-ancestor", sourceSha, base]).returncode == 0:
        continue  # already landed, or carried by the slice below
    # Riders already on main (or carried below) drop out; a rider that moves is a new candidate. A rider
    # that won't merge drops out too and is integration's to fix: riders never hold the star.
    riders = []
    below = base
    for rider in riderBranches:
        riderSha = git("rev-parse", "-q", "--verify", f"origin/{rider}", check=False)
        if not riderSha:
            log(f"rider missing: {rider} for slice {number} ({slug})")
            tell(f"rider-missing-{rider}-{number}", "system_adamic_integration", f"Train slice {number} ({slug}) names rider {rider}, which isn't on origin; it builds without it.")
            continue
        if subprocess.run(["git", "merge-base", "--is-ancestor", riderSha, base]).returncode == 0:
            continue
        # @system_adamic, October 8: a rider joins only after its own fast gate is green alone on main,
        # and at most three ride a slice; the rest wait for the next slice or land on their own.
        if len(riders) >= 3 or not riderGreenAlone(rider, riderSha):
            continue
        merged = subprocess.run(["git", "merge-tree", "--write-tree", "--name-only", below, riderSha], capture_output=True, text=True)
        if merged.returncode != 0:
            files = [line for line in merged.stdout.splitlines()[1:] if line and not line.startswith(("Auto-merging", "CONFLICT"))]
            log(f"rider conflict: {rider} {riderSha[:8]} onto {below[:8]} (slice {number}): {' '.join(files[:6])}; building without it")
            tell(f"rider-conflict-{riderSha[:12]}-{base[:12]}", "system_adamic_integration", f"Rider {rider} {riderSha[:8]} conflicts with slice {number} ({slug})'s base {base[:8]} in {', '.join(files[:6])}; the slice builds without it.")
            continue
        below = git("commit-tree", merged.stdout.splitlines()[0], "-p", below, "-p", riderSha, "-m",
                    f"Merge {rider} at {riderSha[:8]} onto {below[:8]}, riding below slice {number} ({slug}) of the star's train")
        riders.append((rider, riderSha))
    riderTag = "".join(f"-with-{riderSha[:8]}" for _, riderSha in riders)
    # Fixes: commits the owner made for this slice after a later slice was on top of it, so they carry
    # no trailer and sit above the later slice on the rehearsal. The owner sends their shas and
    # integration lists them in a state file fixes-<slug>, in order; each is cherry-picked into the
    # slice's tree, so the slice gets its fix without the later slice's commits. A fix already on the
    # base or in the source drops out.
    fixesFile = os.path.join(state, f"fixes-{slug}")
    fixes = []
    for line in (open(fixesFile).read().split() if os.path.exists(fixesFile) else []):
        fixSha = git("rev-parse", "-q", "--verify", f"{line}^{{commit}}", check=False)
        if not fixSha:
            log(f"fix missing: {line} for slice {number} ({slug})")
            continue
        if any(subprocess.run(["git", "merge-base", "--is-ancestor", fixSha, other]).returncode == 0 for other in (base, sourceSha)):
            continue
        fixes.append(fixSha)
    fixTag = "-fixes-" + subprocess.run(["git", "hash-object", "--stdin"], input=" ".join(fixes), capture_output=True, text=True).stdout[:8] if fixes else ""
    prefix = f"cloud/land-train-{number}-{slug}{riderTag}{fixTag}-{sourceSha[:8]}-"
    name = f"{prefix}{base[:8]}"
    if name not in trainHeads:
        # When a slice lands, main moves to its velocity commit, a row in a table no test reads. The
        # candidate above, built on the landed slice, keeps its gates: push-main lands it over a main
        # that moved only by records, so it isn't rebuilt and gated whole again. Its base is below
        # its rider merges, one first parent each.
        for other, otherSha in trainHeads.items():
            if other.startswith(prefix):
                otherBase = git("rev-parse", otherSha + "^1" * (len(riders) + 1))
                if subprocess.run(["git", "merge-base", "--is-ancestor", otherBase, base]).returncode == 0 and recordsOnly(otherBase, base):
                    name = other
                    break
    if name in trainHeads:
        candidate = trainHeads[name]
    else:
        merged = subprocess.run(["git", "merge-tree", "--write-tree", "--name-only", below, sourceSha], capture_output=True, text=True)
        if merged.returncode != 0:
            files = [line for line in merged.stdout.splitlines()[1:] if line and not line.startswith(("Auto-merging", "CONFLICT"))]
            log(f"conflict: {source} {sourceSha[:8]} onto {below[:8]} (slice {number}): {' '.join(files[:6])}")
            tell(f"conflict-{name}", owner, f"The star's train can't stack {source} {sourceSha[:8]} onto {base[:8]} (slice {number}, {slug}): it conflicts in {', '.join(files[:6])}. Please merge {base[:8]} into {source}, both sides kept, and push; the train restacks on its own.")
            tell(f"conflict-{name}", "system_adamic_integration", f"Train conflict at slice {number} ({slug}): {source} {sourceSha[:8]} onto {below[:8]} in {', '.join(files[:6])}, sent to @{owner}.")
            break
        tree = merged.stdout.splitlines()[0]
        picked = None
        for fixSha in fixes:
            step = git("commit-tree", tree, "-p", below, "-p", sourceSha, "-m", "star-train fix step")
            merged = subprocess.run(["git", "merge-tree", "--write-tree", "--name-only", f"--merge-base={fixSha}^", step, fixSha], capture_output=True, text=True)
            if merged.returncode != 0:
                picked = fixSha
                break
            tree = merged.stdout.splitlines()[0]
        if picked:
            files = [line for line in merged.stdout.splitlines()[1:] if line and not line.startswith(("Auto-merging", "CONFLICT"))]
            log(f"fix conflict: {picked[:8]} into slice {number} ({slug}): {' '.join(files[:6])}")
            tell(f"fix-conflict-{name}", owner, f"The star's train can't take fix {picked[:8]} into slice {number} ({slug}) on {base[:8]}: it conflicts in {', '.join(files[:6])}. Please send a fix that applies on {source} {sourceSha[:8]}.")
            tell(f"fix-conflict-{name}", "system_adamic_integration", f"Train fix {picked[:8]} conflicts in slice {number} ({slug}): {', '.join(files[:6])}, sent to @{owner}.")
            break
        ridden = (f"\nFixes cherry-picked into it: {', '.join(fixSha[:8] for fixSha in fixes)}." if fixes else "")
        ridden += f"\nRiders merged below it: {', '.join(f'{rider} {riderSha[:8]}' for rider, riderSha in riders)}." if riders else ""
        message = (f"Merge {source} at {sourceSha[:8]} onto {below[:8]}, slice {number} ({slug}) of the star's train\n\n"
                   f"Gates as if every slice below had landed; git merge-tree was clean.{ridden}\n\nGate-runs: deferred")
        candidate = git("commit-tree", tree, "-p", below, "-p", sourceSha, "-m", message)
        if not dryRun:
            git("push", "-q", "origin", f"{candidate}:refs/heads/{name}")
            appendOnce(os.path.join(watch, "priority"), name)
        log(f"built {name} {candidate[:12]}")
        for other, otherSha in trainHeads.items():
            if other.startswith(f"cloud/land-train-{number}-") and other != name and not dryRun:
                appendOnce(os.path.join(watch, "skip"), f"{other} {otherSha}")
                # A superseded slice's whole gate would spend Home on a tree that never lands.
                removeLine(requests, otherSha)
    # Every slice's whole gate runs ahead, in train order (Kirk, October 8): the full-gate loop takes
    # the oldest line of its request file between mains and drops it when the record publishes.
    if not dryRun:
        os.makedirs(os.path.dirname(requests), exist_ok=True)
        appendOnce(requests, candidate)
    chain.append((number, slug, source, owner, name, candidate, base, riders))
    base = candidate

# Home takes the oldest request first, so the file is kept in train order: a slice rebuilt below its
# neighbors would otherwise sit behind the slices stacked on it, and the bottom gates last.
if chain and not dryRun and os.path.exists(requests):
    lines = open(requests).read().split()
    ordered = [entry[5] for entry in chain if entry[5] in lines]
    rest = [line for line in lines if line not in ordered]
    if ordered + rest != lines:
        with open(requests, "w") as handle:
            handle.write("\n".join(ordered + rest) + "\n")

if not chain:
    sys.exit(0)

number, slug, source, owner, name, candidate, candidateBase, riders = chain[0]
# An owner can hold a source commit from landing (it still gates): a file hold-<source8> in the state
# directory, holding why. A new commit on the source is a new candidate, so the hold lapses with it.
sourceSha = chain[0][5] and git("rev-parse", f"{candidate}^2")
hold = os.path.join(state, f"hold-{sourceSha[:8]}")
if os.path.exists(hold):
    log(f"held: {name} ({open(hold).read().strip()})")
    sys.exit(0)
# The star's slices land on their own whole gate (Kirk and @system_adamic, October 8): the full gate
# run on the candidate is its verdict and main's confirmation at once. The fast gate stays the early
# first-failure signal; a red on either is the slice's red.
for kind in ("fast", "full-main"):
    early, earlyStatus = newestLog(candidate, kind)
    if earlyStatus.startswith("red"):
        log(f"red: {name} {early}")
        tell(f"red-{candidate}-{kind}", "system_adamic_integration", f"Train slice {number} ({slug}) is red on its {kind} gate: {early}. Everything above it waits for the fix on {source}.")
        sys.exit(0)
reference, status = newestLog(candidate, "full-main")
if not status.startswith("green"):
    sys.exit(0)

mainLog, mainStatus = mainConfirmation(main)
if not mainLog or mainStatus.startswith("running") or not (mainStatus.startswith("green") or explained(mainLog)):
    log(f"waiting: {name} is green, main {main[:8]}'s whole gate is {mainStatus.split(':')[0] or 'not started'}")
    sys.exit(0)

log(f"landing {name} on {reference}")
if dryRun:
    sys.exit(0)
note = f"the star's train, slice {number} ({slug}): {source} {git('rev-parse', candidate + '^2')[:8]}" + "".join(f", rider {rider} {riderSha[:8]}" for rider, riderSha in riders)
pushed = subprocess.run(["bash", os.path.join(directory, "push-main.sh"), "--full-gate", reference, candidate, note], capture_output=True, text=True)
with open(os.path.join(state, f"push-{candidate[:12]}.log"), "w") as handle:
    handle.write(pushed.stdout + pushed.stderr)
if pushed.returncode == 0:
    landed = [line for line in pushed.stdout.splitlines() if line.startswith("Pushed main")]
    log(f"landed {name}")
    tell(f"landed-{candidate}", "system_adamic", f"Train landed slice {number} ({slug}) by itself: {landed[0][:400] if landed else candidate[:12]}. Its whole gate is main's confirmation, so the next slice lands on its own record.")
    tell(f"landed-{candidate}", owner, f"Your slice {number} ({slug}) landed: {landed[0][:300] if landed else candidate[:12]}.")
else:
    refusal = (pushed.stderr.strip() or pushed.stdout.strip()).splitlines()[-1:]
    log(f"refused {name}: {refusal}")
    tell(f"refused-{candidate}", "system_adamic_integration", f"push-main refused train slice {number} ({slug}) {candidate[:12]}: {refusal[0][:400] if refusal else 'no output'}")

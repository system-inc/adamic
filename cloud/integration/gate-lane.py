#!/usr/bin/env python3
"""The gate lane: a candidate's record lands within two minutes of publishing, with no mind in between (@system_adamic,
Oct 9 05:52, #43kay4z: the five-minute rule's "no handoff ever holds a green", made mechanical). Each run reads
integration's candidates, finds each sha's newest finished gate record, and hands it to push-main.sh, which is the only
judge: it lands a record that is complete, zerorun-clean and red only on names ruled infra (--infra-red checks the
class on the output), and refuses anything else. A landing is posted on the candidate's task with the record's publish
time and leaves the file; a refusal is posted once per record; a hold (exit 3) waits. A candidate behind main's tip is
gated as its merge onto main (refs/gate-merges/<merge sha>), so a gate merge's records count as the candidate's and
the merge is what lands.

The candidates file, one line per landing (integration's to edit: a cut adds its line, a landing removes it):
    <landing branch><TAB><full sha><TAB><task id><TAB><infra names, "<package> <Test>=<evidence>" joined by ";", or empty>

usage (launchd com.adamic.gate-lane runs it every minute, from the merge tree): cloud/integration/gate-lane.py
"""
import datetime, json, os, subprocess, sys

directory = os.path.dirname(os.path.abspath(__file__))
candidates = os.environ.get("GATE_LANE_CANDIDATES", os.path.expanduser("~/.adamic-integration/candidates"))
state = os.environ.get("GATE_LANE_STATE", os.path.expanduser("~/.adamic-gate-lane"))
pushMain = os.environ.get("GATE_LANE_PUSH_MAIN", os.path.join(directory, "push-main.sh"))
ahra = os.environ.get("GATE_LANE_AHRA", "cd ~/Projects/ahra && ahra")


def git(*arguments):
    return subprocess.run(["git", *arguments], capture_output=True, text=True).stdout.strip()


def records(sha):
    """The newest finished record set of sha: (stamp, 'full', ref) or (stamp, 'fast', ref, phases ref or None), or None."""
    refs = [line.split("\t")[1].removeprefix("refs/heads/") for line in git("ls-remote", "origin", "refs/heads/gate-logs/%s/*" % sha[:12]).splitlines() if "\t" in line]
    finished = {}
    for ref in refs:
        kind = ref.rsplit("/", 1)[1]
        if kind not in ("fast", "fast-phases", "full-main"):
            continue
        git("fetch", "-q", "origin", "+refs/heads/%s:refs/remotes/origin/%s" % (ref, ref))
        status = git("show", "origin/%s:status.txt" % ref).split("\n")[0]
        if status.startswith(("green", "red")):
            finished[ref] = kind
    stamps = sorted({ref.split("/")[2] for ref in finished}, reverse=True)
    for stamp in stamps:
        at = {kind: ref for ref, kind in finished.items() if ref.split("/")[2] == stamp}
        if "full-main" in at:
            return (stamp, "full", at["full-main"])
        if "fast" in at:
            return (stamp, "fast", at["fast"], at.get("fast-phases"))
    return None


def gateMerges(sha):
    """The shas Loom gated in sha's place: a candidate behind main's tip is gated as a merge onto it, kept at
    refs/gate-merges/<merge sha> with main as first parent and the candidate as second, and that merge is what lands."""
    git("fetch", "-q", "origin", "+refs/gate-merges/*:refs/gate-merges/*")
    merges = []
    for line in git("for-each-ref", "refs/gate-merges", "--format=%(objectname) %(parent)").splitlines():
        fields = line.split()
        if len(fields) == 3 and fields[2] == sha:
            merges.append(fields[0])
    return merges


def newest(sha):
    """The newest finished record set gated for the candidate, itself or a gate merge of it: (sha it gated, set) or None."""
    found = [(record, gated) for gated in [sha] + gateMerges(sha) for record in [records(gated)] if record]
    if not found:
        return None
    record, gated = max(found, key=lambda pair: pair[0][0])
    return gated, record[1:]


def post(task, text):
    # ahra reads a comment from a file, not stdin; a post that fails is logged, never silent.
    import tempfile
    with tempfile.NamedTemporaryFile("w", suffix=".md", delete=False) as handle:
        handle.write(text)
    ran = subprocess.run("%s tasks comment %s --role Agent --text-file %s" % (ahra, task, handle.name), shell=True, text=True, capture_output=True)
    os.unlink(handle.name)
    if ran.returncode != 0 or "Comment added" not in ran.stdout + ran.stderr:
        print("could not post on %s: %s" % (task, (ran.stdout + ran.stderr).strip()[-300:]), flush=True)


def main():
    os.makedirs(state, exist_ok=True)
    # One run at a time: launchd starts one a minute, and a landing can take longer than that.
    import fcntl
    lock = open(os.path.join(state, "lock"), "w")
    try:
        fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
    except BlockingIOError:
        return
    seen = os.path.join(state, "seen")
    told = set(open(seen).read().split("\n")) if os.path.exists(seen) else set()
    if not os.path.exists(candidates):
        return
    landed = set()
    for line in open(candidates).read().splitlines():
        fields = line.split("\t")
        if len(fields) < 3 or line.startswith("#"):
            continue
        branch, sha, task = fields[:3]
        infra = [name for name in (fields[3] if len(fields) > 3 else "").split(";") if name.strip()]
        newestSet = newest(sha)
        if not newestSet:
            continue
        gated, found = newestSet
        # The record says which tree it gated; a set published under the candidate's prefix may have gated a gate
        # merge of it (the star, Oct 9 14:43Z: phases under f9fc14c1478f/ gated c310d512). Land what the record names,
        # when it's the candidate or one of its gate merges; push-main still refuses a set whose records disagree.
        named = git("show", "origin/%s:%s" % (found[1], "full.json" if found[0] == "full" else "fast.json"))
        try:
            named = json.loads(named).get("sha", "") if named else ""
        except ValueError:
            named = ""
        if named and named != gated and named in [sha] + gateMerges(sha):
            gated = named
        arguments = ["--full-gate", found[1]] if found[0] == "full" else ["--fast-gate", found[1]] + (["--also-gate", found[2]] if found[2] else [])
        for name in infra:
            arguments += ["--infra-red", name.strip()]
        key = " ".join([gated] + arguments)
        # A gate merge names its candidate, so the landing and the task read which commit came in.
        label = gated[:8] if gated == sha else "%s (gate merge of %s)" % (gated[:8], sha[:8])
        published = git("log", "-1", "--format=%cI", "origin/" + found[1])
        ran = subprocess.run(["bash", pushMain, *arguments, gated, "%s (gate lane)" % branch], capture_output=True, text=True)
        now = datetime.datetime.now(datetime.timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
        if ran.returncode == 0:
            pushed = [l for l in ran.stdout.splitlines() if l.startswith("Pushed main")]
            post(task, "Gate lane landed %s at %s on %s (record published %s): %s" % (label, now, " ".join(arguments), published, pushed[0] if pushed else ""))
            landed.add(line)
            continue
        if ran.returncode != 3 and key not in told:
            reason = [l for l in ran.stderr.splitlines() if l.startswith("refused")] or ran.stderr.splitlines()[-1:]
            post(task, "Gate lane: %s on %s doesn't land: %s" % (label, " ".join(arguments), (reason[0] if reason else "push-main exit %d" % ran.returncode)[:600]))
            told.add(key)
    open(seen, "w").write("\n".join(sorted(told)))
    # Lines added while this run was landing must survive it: re-read the file and drop only what landed (three cuts
    # appended mid-run were lost when the run wrote back its own earlier read, Oct 9 07:10).
    kept = [line for line in open(candidates).read().splitlines() if line not in landed]
    open(candidates + ".new", "w").write("\n".join(kept) + ("\n" if kept else ""))
    os.replace(candidates + ".new", candidates)


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""The gate lane: a candidate's record lands within two minutes of publishing, with no mind in between (@system_adamic,
Oct 9 05:52, #43kay4z: the five-minute rule's "no handoff ever holds a green", made mechanical). Each run reads
integration's candidates, finds each sha's newest finished gate record, and hands it to push-main.sh, which is the only
judge: it lands a record that is complete, zerorun-clean and red only on names ruled infra (--infra-red checks the
class on the output), and refuses anything else. A landing is posted on the candidate's task with the record's publish
time and leaves the file; a refusal is posted once per record; a hold (exit 3) waits.

The candidates file, one line per landing (integration's to edit: a cut adds its line, a landing removes it):
    <landing branch><TAB><full sha><TAB><task id><TAB><infra names, "<package> <Test>=<evidence>" joined by ";", or empty>

usage (launchd com.adamic.gate-lane runs it every minute, from the merge tree): cloud/integration/gate-lane.py
"""
import datetime, os, subprocess, sys

directory = os.path.dirname(os.path.abspath(__file__))
candidates = os.environ.get("GATE_LANE_CANDIDATES", os.path.expanduser("~/.adamic-integration/candidates"))
state = os.environ.get("GATE_LANE_STATE", os.path.expanduser("~/.adamic-gate-lane"))
pushMain = os.environ.get("GATE_LANE_PUSH_MAIN", os.path.join(directory, "push-main.sh"))
ahra = os.environ.get("GATE_LANE_AHRA", "cd ~/Projects/ahra && ahra")


def git(*arguments):
    return subprocess.run(["git", *arguments], capture_output=True, text=True).stdout.strip()


def records(sha):
    """The newest finished record set of sha: ('full', ref) or ('fast', ref, phases ref or None), or None."""
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
            return ("full", at["full-main"])
        if "fast" in at:
            return ("fast", at["fast"], at.get("fast-phases"))
    return None


def post(task, text):
    subprocess.run("%s tasks comment %s --role Agent --text-file -" % (ahra, task), shell=True, input=text, text=True, capture_output=True)


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
    kept = []
    for line in open(candidates).read().splitlines():
        fields = line.split("\t")
        if len(fields) < 3 or line.startswith("#"):
            kept.append(line)
            continue
        branch, sha, task = fields[:3]
        infra = [name for name in (fields[3] if len(fields) > 3 else "").split(";") if name.strip()]
        found = records(sha)
        if not found:
            kept.append(line)
            continue
        arguments = ["--full-gate", found[1]] if found[0] == "full" else ["--fast-gate", found[1]] + (["--also-gate", found[2]] if found[2] else [])
        for name in infra:
            arguments += ["--infra-red", name.strip()]
        key = " ".join([sha] + arguments)
        published = git("log", "-1", "--format=%cI", "origin/" + found[1])
        ran = subprocess.run(["bash", pushMain, *arguments, sha, "%s (gate lane)" % branch], capture_output=True, text=True)
        now = datetime.datetime.now(datetime.timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
        if ran.returncode == 0:
            pushed = [l for l in ran.stdout.splitlines() if l.startswith("Pushed main")]
            post(task, "Gate lane landed %s at %s on %s (record published %s): %s" % (sha[:8], now, " ".join(arguments), published, pushed[0] if pushed else ""))
            continue
        if ran.returncode != 3 and key not in told:
            reason = [l for l in ran.stderr.splitlines() if l.startswith("refused")] or ran.stderr.splitlines()[-1:]
            post(task, "Gate lane: %s on %s doesn't land: %s" % (sha[:8], " ".join(arguments), (reason[0] if reason else "push-main exit %d" % ran.returncode)[:600]))
            told.add(key)
        kept.append(line)
    open(seen, "w").write("\n".join(sorted(told)))
    open(candidates + ".new", "w").write("\n".join(kept) + ("\n" if kept else ""))
    os.replace(candidates + ".new", candidates)


if __name__ == "__main__":
    main()

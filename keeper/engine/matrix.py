"""matrix.py: the kill matrix, one JSONL per base (#a5wmhm0, design #j3cbpzp section 4, @system_adamic_tests).

	matrix.py add <matrix.jsonl> <mutant.json> <classified.json>   # merge one run (classify.py's output) into its mutant's row
	matrix.py unique <matrix.jsonl>                                # each test's unique kills: mutants only it catches
	matrix.py survivors <matrix.jsonl>                             # mutants no test catches
	matrix.py lost <older.jsonl> <newer.jsonl>                     # tests that held a unique kill and hold none now

A row is one mutant, keyed by its hash: {"hash", "file", "line", "op", "base", "outcomes": {test: outcome}, "catchers"}.
Every run of the same mutant merges into its row, so a mutant replayed twice is never two mutants (the oracle set's
b11-024 and b11-033 were one diff with two different first catchers). In a merge a test's newer outcome wins, except
that "absent" never overwrites what a test actually did, and a catch never yields to a kill or deadline from a busier
run. A catcher is a test whose outcome is fail or panic-in-code, the outcomes classify.py says count.
"""
import json, os, sys

counting = ("fail", "panic-in-code")
weak = ("absent", "kill", "deadline")


def load(path):
    rows = {}
    if os.path.exists(path):
        for line in open(path):
            if line.strip():
                row = json.loads(line)
                rows[row["hash"]] = row
    return rows


def save(path, rows):
    temporary = path + ".tmp"
    with open(temporary, "w") as handle:
        for hash in sorted(rows):
            handle.write(json.dumps(rows[hash], sort_keys=True) + "\n")
    os.replace(temporary, path)


def merge(row, classified):
    for test in classified["tests"]:
        name, outcome = test["test"], test["outcome"]
        before = row["outcomes"].get(name)
        if before is not None and (outcome == "absent" or (outcome in weak and before in counting)):
            continue
        row["outcomes"][name] = outcome
    row["catchers"] = sorted(name for name, outcome in row["outcomes"].items() if outcome in counting)
    return row


def unique(rows):
    kills = {}
    for row in rows.values():
        if len(row["catchers"]) == 1:
            kills.setdefault(row["catchers"][0], []).append(row["hash"])
    return {test: sorted(hashes) for test, hashes in sorted(kills.items())}


def main():
    verb, arguments = sys.argv[1], sys.argv[2:]
    if verb == "add":
        path, mutant, classified = arguments[0], json.load(open(arguments[1])), json.load(open(arguments[2]))
        rows = load(path)
        row = rows.setdefault(mutant["hash"], {key: mutant.get(key) for key in ("hash", "file", "line", "op", "base")} | {"outcomes": {}, "catchers": []})
        merge(row, classified)
        save(path, rows)
        print("%s\t%s" % (mutant["hash"], ",".join(row["catchers"]) or "survives"))
    elif verb == "unique":
        json.dump(unique(load(arguments[0])), sys.stdout, indent=1)
        print()
    elif verb == "survivors":
        for row in load(arguments[0]).values():
            if not row["catchers"]:
                print("%s\t%s:%s\t%s" % (row["hash"], row["file"], row["line"], row["op"]))
    elif verb == "lost":
        older, newer = load(arguments[0]), load(arguments[1])
        before, after = unique(older), unique(newer)
        seen = {test for row in newer.values() for test, outcome in row["outcomes"].items() if outcome not in ("absent", "skip")}
        for test in sorted(before):
            if test in seen and test not in after:
                print("%s\tlost its last unique kill (had %s)" % (test, ",".join(before[test])))
    else:
        sys.exit("verbs: add, unique, survivors, lost")


if __name__ == "__main__":
    main()

"""classify.py: the catch rules in code (#rdrbxnp, design #j3cbpzp section 3, @system_adamic_tests).

	classify.py <mutant go test -json log> [--base <base log>] [--rules rules.json]
	# prints JSON: one row per top-level test, {"test", "package", "outcome", "counts", "why"}, plus "catchers"

What the audit learned to count, by hand, now counted the same way every time. A test catches a mutant when it fails
on an assertion, or panics inside the code under test, and is green at base. Never a catch:
	pin              a hash, census or golden of the inputs or outputs (rules.json "pins"): it trips on any change,
	                 right or wrong. A pin that has a direction-check partner (rules.json "pairs") says so, because
	                 the pair is proven by the mutant plus a regenerated golden, never by the bare mutant.
	witness          a test that plants its own mutants (Mutant, Mutants or Witness in its name, or rules.json "witnesses")
	kill, deadline   killed by a signal, or out of time: under load these were green alone (M06c, M08c), so they're
	                 rerun alone before they're believed ("rerun": true)
	panic-harness    a panic whose deepest frame in the repository is test code, or outside internal/ and stage1/
	red-at-base      failing on the base log too
	absent           never finished: in flight when another test's panic stopped the package
A panic inside the code under test is a catch (D5 at class_static.go:177 panicked in ir and lower on valid input).
"""
import argparse, json, os, re, sys

module = "github.com/system-inc/adamic/"
planting = re.compile(r"Mutants?|Witness")
killed = re.compile(r"signal: killed|signal: terminated|\bKilled\b")
deadline = re.compile(r"panic: test timed out|exceeded its \d+s deadline|deadline exceeded|context deadline exceeded|own work exceeded")
frame = re.compile(r"^\s*(/\S+?\.(?:go|c|h)):\d+")


def tests(log):
    """Top-level test -> {"package", "action", "output"}. Subtests fold into their parent's output."""
    rows = {}
    for line in open(log, errors="replace"):
        try:
            event = json.loads(line)
        except ValueError:
            continue
        name = event.get("Test")
        if not name:
            continue
        top = name.split("/")[0]
        row = rows.setdefault(top, {"package": event.get("Package", "").replace(module, ""), "action": None, "output": []})
        if event.get("Action") == "output":
            row["output"].append(event.get("Output", ""))
        elif event.get("Action") in ("pass", "fail", "skip") and "/" not in name:
            row["action"] = event["Action"]
    return rows


def panic_frame(output):
    """The deepest stack frame inside the repository after the first panic, as a repository-relative path."""
    text = "".join(output)
    at = text.find("panic:")
    if at < 0:
        return None
    for line in text[at:].splitlines():
        match = frame.match(line)
        if not match:
            continue
        path = match.group(1)
        for root in ("/adamic/", "/" + module):
            if root in path:
                relative = path.split(root, 1)[1]
                if not relative.startswith(("go/src/", "pkg/mod/")):
                    return relative
    return None


def classify(row, name, rules, base):
    text = "".join(row["output"])
    if row["action"] == "skip":
        return "skip", "skipped"
    if name in rules.get("pins", []):
        pair = [pair["direction"] for pair in rules.get("pairs", []) if pair["golden"] == name]
        return "pin", "a golden or census" + (" (pair: %s, proven by a regenerated golden)" % pair[0] if pair else "")
    if planting.search(name) or name in rules.get("witnesses", []):
        return "witness", "plants its own mutants"
    if row["action"] == "pass":
        return "pass", ""
    if base.get(name, {}).get("action") == "fail":
        return "red-at-base", "fails at base too"
    if deadline.search(text):
        return "deadline", deadline.search(text).group(0)
    if killed.search(text):
        return "kill", killed.search(text).group(0)
    if row["action"] is None:  # in flight when another test's panic took the binary down (D5's log: 112 of them)
        return "absent", "never finished: the package stopped first"
    if "panic:" in text:
        where = panic_frame(row["output"])
        if where and where.startswith(("internal/", "stage1/")) and not where.endswith("_test.go"):
            return "panic-in-code", where
        return "panic-harness", where or "no frame in the repository"
    return "fail", next((line.strip() for line in row["output"] if re.search(r"_test\.go:\d+:", line)), "failed")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("log")
    parser.add_argument("--base")
    parser.add_argument("--rules", default=os.path.join(os.path.dirname(os.path.abspath(__file__)), "rules.json"))
    options = parser.parse_args()
    rules = json.load(open(options.rules)) if os.path.exists(options.rules) else {}
    base = tests(options.base) if options.base else {}
    rows = []
    for name, row in sorted(tests(options.log).items()):
        outcome, why = classify(row, name, rules, base)
        rows.append({"test": name, "package": row["package"], "outcome": outcome, "counts": outcome in ("fail", "panic-in-code"), "rerun": outcome in ("kill", "deadline"), "why": why})
    json.dump({"tests": rows, "catchers": [row["test"] for row in rows if row["counts"]]}, sys.stdout, indent=1)
    print()


if __name__ == "__main__":
    main()

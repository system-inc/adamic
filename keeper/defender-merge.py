"""defender-merge.py: every defender reply (fanout/replies/du*.md) folded into defender-ledger.json, one row per
defended, undefended or unjudged test, keyed by unit and test so a re-defended row replaces its earlier entry."""
import glob, json, os, re, sys
S = sys.argv[1]
path = S + "/defender-ledger.json"
ledger = json.load(open(path)) if os.path.exists(path) else {}
unreadable = []
# Oldest reply first, so a rerun of a row (a named chunk) replaces the run before it: file names don't sort by time.
for reply in sorted(glob.glob(S + "/fanout/replies/du*.md"), key=os.path.getmtime):
    name = os.path.basename(reply)[:-3]
    unit = re.match(r"du(\d{3}|-fresh)", name)
    if not unit:
        continue
    text = open(reply).read()
    found = re.search(r"```json\n(.*?)\n```", text, re.S)
    try:
        rows = json.loads(found.group(1))
    except Exception:
        unreadable.append(name)
        continue
    for row in rows:
        row["unit"] = "u" + unit.group(1)
        row["defender_run"] = name
        ledger["%s %s" % (row["unit"], row["test"])] = row
# A unique kill inside a bounded matrix (the whole package overran 90 s) keeps the row: defenders word it as defended
# or as cannot-judge. Read the attempts, not the word: an attempt whose only failures are the row itself, or members
# of its own family, defends it, marked bounded so the replay still settles package-wide uniqueness.
def own(row, failed):
    name, failed = row["test"].replace(" family", ""), failed.replace(" family", "")
    return failed == name or failed.startswith(name + "_") or failed.startswith(name + "/")
for row in ledger.values():
    if row.get("defense") == "cannot-judge":
        for attempt in row.get("attempts") or []:
            failed = attempt.get("rows_failed") or []
            if failed and all(own(row, name) for name in failed):
                row["defender_defense"] = "cannot-judge"
                row["defense"], row["bounded"] = "defended", True
                row["unique_mutant"] = row.get("unique_mutant") or "%s %s" % (attempt.get("mutant"), attempt.get("file_line"))
                break
# A family's setup, numbered shard or product row (brief v8's family rule) is judged with its family, so a defender's
# verdict on one alone never makes it a deletion candidate.
for row in ledger.values():
    # A member named by its fixture's encoded path (Test<Family>____oracle_testdata_...) is a family part too.
    if re.search(r"_(Setup|\d{2,3})$", row["test"].split()[0]) or row["test"].startswith("TestProduct_") or "____" in row["test"]:
        row.setdefault("defender_defense", row.get("defense"))
        row["defense"] = "family-part"
# The keeper's own rulings on returned rows (keeper-overrides.json, {"<unit> <test>": {field: value}}) apply after every
# merge, so a re-merge of the replies never undoes them.
overrides = S + "/keeper-overrides.json"
if os.path.exists(overrides):
    for key, fields in json.load(open(overrides)).items():
        if key in ledger:
            ledger[key].update(fields)
json.dump(ledger, open(path + ".partial", "w"), indent=1)
os.rename(path + ".partial", path)
counts = {}
for row in ledger.values():
    counts[row.get("defense")] = counts.get(row.get("defense"), 0) + 1
print("rows", len(ledger), counts, "unreadable", unreadable)

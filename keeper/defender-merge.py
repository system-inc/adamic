"""defender-merge.py: every defender reply (fanout/replies/du*.md) folded into defender-ledger.json, one row per
defended, undefended or unjudged test, keyed by unit and test so a re-defended row replaces its earlier entry."""
import glob, json, os, re, sys
S = sys.argv[1]
path = S + "/defender-ledger.json"
ledger = json.load(open(path)) if os.path.exists(path) else {}
unreadable = []
for reply in sorted(glob.glob(S + "/fanout/replies/du*.md")):
    name = os.path.basename(reply)[:-3]
    unit = re.match(r"du(\d{3})", name)
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
json.dump(ledger, open(path + ".partial", "w"), indent=1)
os.rename(path + ".partial", path)
counts = {}
for row in ledger.values():
    counts[row.get("defense")] = counts.get(row.get("defense"), 0) + 1
print("rows", len(ledger), counts, "unreadable", unreadable)

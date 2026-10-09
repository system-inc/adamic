"""merge.py: fold every fanout reply not yet in the ledger into it, one row per reply row, tagged with its unit."""
import glob, json, os, re, sys
S = sys.argv[1]
ledgerPath = S + "/test-audit-ledger.json"
ledger = json.load(open(ledgerPath))
have = {row["unit"] for row in ledger}
units = {u["unit"]: u for u in json.load(open(S + "/manifest/units.json"))["units"]}
repo = os.path.expanduser("~/Projects/system/adamic")


def fromBranch(unit):
    """A reply that left its rows on the evidence branch: rows.json there, else the first json block of REPORT.md."""
    import subprocess
    slug = units[unit]["slug"]
    branch = "origin/test-audit/" + slug
    if subprocess.run(["git", "-C", repo, "fetch", "-q", "origin", "test-audit/" + slug]).returncode != 0:
        return None
    for name in ("rows.json", "REPORT.md", "report.md"):
        shown = subprocess.run(["git", "-C", repo, "show", "%s:review/test-audit/%s/%s" % (branch, slug, name)], capture_output=True, text=True)
        if shown.returncode != 0:
            continue
        try:
            if name.endswith(".json"):
                return json.loads(shown.stdout)
            found = re.search(r"```json\n(.*?)\n```", shown.stdout, re.S)
            if found:
                return json.loads(found.group(1))
        except Exception:
            continue
    return None
added, unreadable = [], []
for path in sorted(glob.glob(S + "/fanout/replies/u[0-9][0-9][0-9].md")):
    unit = os.path.basename(path)[:-3]
    if unit in have:
        continue
    text = open(path).read()
    match = re.search(r"```json\n(.*?)\n```", text, re.S)
    try:
        rows = json.loads(match.group(1))
    except Exception:
        rows = fromBranch(unit)
        if rows is None:
            unreadable.append(unit)
            continue
    base = (re.search(r"\b([0-9a-f]{40})\b", text) or re.search(r"`([0-9a-f]{8,})`", text))
    for row in rows:
        row.update({"unit": unit, "wave": "main", "brief": "v7", "base": base.group(1)[:10] if base else None})
    ledger += rows
    added.append(unit)
json.dump(ledger, open(ledgerPath + ".partial", "w"), indent=1)
os.rename(ledgerPath + ".partial", ledgerPath)
print("added", added, "unreadable", unreadable, "rows", len(ledger))

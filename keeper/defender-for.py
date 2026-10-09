"""defender-for.py: the defender brief for one unit, its rows filled from the ledger (subsumed and untrue rows)."""
import json, sys
brief, manifest, ledgerPath, unitId = sys.argv[1:5]
# "u031" is every candidate row of the unit; "u031.2" is its second chunk of three, so a defender has the budget for
# three honest attempts per row (du033 showed ten rows can't get them in one window); "u033=TestA,TestB" names rows.
chunk, names = None, None
if "=" in unitId:
    unitId, listed = unitId.split("=", 1)
    names = set(listed.split(","))
elif "." in unitId:
    unitId, chunk = unitId.split(".", 1)
    chunk = int(chunk)
units = {u["unit"]: u for u in json.load(open(manifest))["units"]}
u = units[unitId]
rows = [r for r in json.load(open(ledgerPath)) if r["unit"] == unitId and (r.get("keeper_verdict") or r.get("verdict")) in ("subsumed", "untrue")]
if names is not None:
    rows = [r for r in rows if r["test"] in names]
if chunk is not None:
    rows = rows[(chunk - 1) * 3:chunk * 3]
lines = []
for r in rows:
    verdict = r.get("keeper_verdict") or r.get("verdict")
    by = r.get("subsumed_by")
    by = ", ".join(by) if isinstance(by, list) else (by or "")
    lines.append("- %s (%s%s): kills %s; audit evidence: %s" % (r["test"], verdict, (", subsumed by " + by) if by else "", r.get("kills") or [], str(r.get("evidence") or "")[:200]))
text = open(brief).read().replace("__PACKAGE__", u["package"]).replace("__SLUG__", u["slug"]).replace("__ROWS__", "\n".join(lines) or "(none)")
print(text)

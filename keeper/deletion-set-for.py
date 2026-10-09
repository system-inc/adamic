"""deletion-set-for.py: the deletion-set replay brief for one package, its candidates read from the defender ledger.

	deletion-set-for.py <brief> <manifest> <defender-ledger.json> <package>

A candidate is a row the defenders left "not defended" or ruled "dead". Its evidence branches are its unit's audit
branch (test-audit/<slug>) and every defender branch that judged it (test-defend/<slug>, plus any the reply named).
"""
import json, re, sys

brief, manifestPath, ledgerPath, package = sys.argv[1:5]
units = {u["unit"]: u for u in json.load(open(manifestPath))["units"]}
rows = [row for row in json.load(open(ledgerPath)).values() if row.get("package") == package and row.get("defense") in ("not defended", "dead")]
if not rows:
    sys.exit("deletion-set-for: no candidates in " + package)
branches = set()
for row in rows:
    # Wave A's u-fresh predates the manifest; its branches are named for its package.
    slug = units[row["unit"]]["slug"] if row["unit"] in units else row["package"].replace("/", "-")
    branches |= {"test-audit/" + slug, "test-defend/" + slug}
    for found in re.findall(r"test-defend/[A-Za-z0-9_./-]+", json.dumps(row)):
        branches.add(found.rstrip("/.`"))
# A family row is skipped as its whole family: the name and its numbered or named members.
names = sorted({row["test"].replace(" family", "").split()[0] for row in rows})
skip = "^(" + "|".join(re.escape(name) for name in names) + ")($|_|/)"
candidates = "\n".join("- `%s` (%s%s)" % (row["test"], row["defense"], ", shared with " + (", ".join(row["subsumed_by"]) if isinstance(row.get("subsumed_by"), list) else str(row.get("subsumed_by") or "")) if row.get("subsumed_by") else "") for row in sorted(rows, key=lambda row: row["test"]))
slug = package.replace("/", "-")
text = open(brief).read()
for key, value in (("__PACKAGE_DIR__", package), ("__PACKAGE__", package), ("__CANDIDATES__", candidates), ("__BRANCHES__", "\n".join("- `%s`" % branch for branch in sorted(branches))), ("__SKIP__", skip), ("__SLUG__", slug)):
    text = text.replace(key, value)
print(text)

import json, sys
brief, manifest, unitId = sys.argv[1:4]
# A unit id "d<unit>" is that unit's defender run (#az6b13b): the defender brief, rows from the ledger.
if unitId.startswith("du"):
    import os, subprocess
    here = os.path.dirname(os.path.abspath(__file__))
    text = subprocess.run([sys.executable, "-I", os.path.join(here, "defender-for.py"), os.path.join(here, "defender-brief-v4.md"), manifest, os.path.join(here, "test-audit-ledger.json"), unitId[1:]], check=True, capture_output=True, text=True).stdout
    # A keeper note for a defender run is keyed by its unit (fanout/notes/du052.md), so every chunk of the unit carries it.
    notes = os.path.join(os.path.dirname(os.path.abspath(manifest)), "..", "fanout", "notes", "du" + unitId[2:].split(".")[0].split("=")[0] + ".md")
    if os.path.exists(notes):
        text += "\n\n## A note from the keeper for this unit\n" + open(notes).read()
    sys.stdout.write(text)
    sys.exit(0)
units = {u['unit']: u for u in json.load(open(manifest))['units']}
u = units[unitId]
pkg = u['package']
big = sum(x['package'] == pkg for x in units.values()) > 1
unitText = ("a slice of the package %s: the %d rows named under \"Your rows\" below" % (pkg, u['rows'])) if big else ("the package %s (every top-level Test in it, families as one row)" % pkg)
text = open(brief).read().replace('__UNIT__', unitText).replace('__PACKAGE__', pkg).replace('__SLUG__', u['slug'])
head = ("New unit for you, from @system_adamic_tests (the keeper of Adamic's test audit, #xphstyt), unit %s. Your workspace is warm, so skip setup if /workspace/adamic-tools/env.sh works. Start clean from origin/main as the brief says; anything left from your last unit is irrelevant. Push your evidence (standalone diffs, matrices, probes) on test-audit/%s under review/test-audit/%s/.\n\n" % (unitId, u['slug'], u['slug']))
tail = ""
if big:
    tail = "\n\n## Your rows\nFiles at origin/main %s: %s\nTest functions (%d):\n%s\n" % (json.load(open(manifest))['origin_main'][:10], ", ".join(u['files']), len(u['tests']), "\n".join(u['tests']))
import os
notes = os.path.join(os.path.dirname(os.path.abspath(manifest)), "..", "fanout", "notes", unitId + ".md")
if os.path.exists(notes):
    tail += "\n\n## A note from the keeper for this unit\n" + open(notes).read()
print(head + text + tail)

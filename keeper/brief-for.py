import json, sys
brief, manifest, unitId = sys.argv[1:4]
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
print(head + text + tail)

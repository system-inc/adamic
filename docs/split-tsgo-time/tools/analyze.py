#!/usr/bin/env python3
# analyze.py <logs dir> <run>...: one summary per run from its .v, .build and .timeline files.
import sys, re, os, collections
logs = sys.argv[1]
rows = []
for run in sys.argv[2:]:
    p = lambda ext: os.path.join(logs, run + ext)
    v = open(p(".v")).read()
    total = float(re.search(r"--- PASS: TestSplitTSGoAgreesUnit0\d \(([\d.]+)s\)", v).group(1))
    sub = float(re.search(r"--- PASS: TestSplitTSGoAgreesUnit0\d/cache_\d \(([\d.]+)s\)", v).group(1))
    fetch = float(re.search(r"product fetch phase ([\d.]+)s", v).group(1))
    whole = float(re.search(r"whole checker run phase ([\d.]+)s", v).group(1))
    sums = collections.defaultdict(float); counts = collections.Counter(); compiles = []; link = None
    for line in open(p(".build")):
        f = line.split()
        if len(f) != 5: continue
        kind, name, key, outcome, sec = f; sec = float(sec)
        sums[(kind, outcome)] += sec; counts[(kind, outcome)] += 1
        if kind == "cCompile": compiles.append((sec, name, outcome))
        if kind == "cLink": link = sec
    t = {}
    loads = []
    for line in open(p(".timeline")):
        f = line.rstrip("\n").split(" ", 2)
        if len(f) < 3: continue
        at, tag, text = float(f[0]), f[1], f[2]
        if tag == "V" and re.match(r"=== RUN\s+\S+/cache_", text): t["sub"] = at
        if tag == "B" and text.startswith("cArchive"): t["archive"] = at
        if tag == "B" and text.startswith("cCompile"): t.setdefault("first", at); t["last"] = at
        if tag == "B" and text.startswith("cLink"): t["link"] = at
        if tag == "V" and "identical bytes" in text: t["checked"] = at
        if tag == "L" and "sub" in t and "checked" not in t: loads.append(float(text.split()[0]))
    build = t["link"] - t["sub"]; splitrun = t["checked"] - t["link"]
    box = open(p(".box")).read()
    before = re.search(r"loadavg-before (\S+)", box).group(1)
    print("== %s" % run)
    print("total %.2f  fetch %.3f  whole run %.3f  cache subtest %.2f" % (total, fetch, whole, sub))
    print("timeline (s after subtest start): runtime archive line %.2f, first cCompile line %.2f, last cCompile line %.2f, cLink line %.2f, split checked %.2f" % (
        t["archive"] - t["sub"], t["first"] - t["sub"], t["last"] - t["sub"], t["link"] - t["sub"], t["checked"] - t["sub"]))
    print("split build ~%.2f (subtest start to cLink line), split run ~%.2f (cLink line to its check logged)" % (build, splitrun))
    print("load 1m before %s, during subtest min %.2f max %.2f" % (before, min(loads), max(loads)))
    for key in sorted(sums): print("  %-9s %-6s n=%2d sum=%.3f" % (key[0], key[1], counts[key], sums[key]))
    compiles.sort(reverse=True)
    print("  slowest cCompile: " + ", ".join("%s %s %.3f" % (n, o, s) for s, n, o in compiles[:10]))
    print("  cLink %.3f" % link)
    cc = [c for c in compiles]
    hits = sum(1 for c in cc if c[2] == "hit"); miss = sum(1 for c in cc if c[2] == "miss"); off = sum(1 for c in cc if c[2] == "off")
    rows.append("| %s | %.2f | %.3f | %.3f | %.2f | %.2f over %d (%d hit, %d miss, %d off) | %.3f | %.2f | %s | %.1f-%.1f |" % (
        run, total, fetch, whole, build, sum(c[0] for c in cc), len(cc), hits, miss, off, link, splitrun, before, min(loads), max(loads)))
print()
print("| run | total s | fetch s | whole run s | split build s | cCompile sum s over N units | cLink s | split run s | load 1m before | load 1m during |")
print("|---|---|---|---|---|---|---|---|---|---|")
print("\n".join(rows))

import csv, json, sys, collections
D = "data"
sites = list(csv.DictReader(open(f"{D}/sites.csv")))
static = collections.Counter(int(r["id"]) for r in csv.DictReader(open(f"{D}/read-sites.csv")) if r["access"] != "write")
out = {}
for w in ("compiler", "services"):
    d = json.load(open(f"{D}/counts-{w}.json"))
    rows = []
    for s in sites:
        i = int(s["id"]); c = int(d["c"][i]); r = int(d["r"][i])
        rows.append({**{k: s[k] for k in ("id","file","line","target","shape","unique")}, "casts": c, "reads": r, "static_read_sites": static[i]})
    with open(f"{D}/per-site-{w}.csv", "w", newline="") as f:
        wr = csv.DictWriter(f, fieldnames=list(rows[0])); wr.writeheader(); wr.writerows(rows)
    tc = sum(r["casts"] for r in rows); tr = sum(r["reads"] for r in rows)
    byfile = collections.Counter(); 
    for r in rows: byfile[r["file"]] += r["casts"]
    top = sorted(rows, key=lambda r: -r["reads"])[:20]
    topc = sorted(rows, key=lambda r: -r["casts"])[:5]
    ordered = sorted(rows, key=lambda r: -r["reads"]); cum = 0; n80 = None
    for k, r in enumerate(ordered, 1):
        cum += r["reads"]
        if n80 is None and cum >= 0.8 * tr: n80 = k
    sub = lambda f: (sum(r["casts"] for r in rows if f(r)), sum(r["reads"] for r in rows if f(r)), sum(1 for r in rows if f(r)))
    out[w] = dict(casts=tc, reads=tr, sites_executed=sum(1 for r in rows if r["casts"]), sites_with_reads=sum(1 for r in rows if r["reads"]), top20=[(f'{r["file"]}:{r["line"]}', r["target"], r["reads"], r["casts"]) for r in top], topcasts=[(f'{r["file"]}:{r["line"]}', r["target"], r["casts"]) for r in topc], sites_for_80pct_reads=n80,
        checker_only=sub(lambda r: r["file"].endswith("/checker.ts")), unique_only=sub(lambda r: r["unique"] == "true"), top20_share=sum(r["reads"] for r in top)/tr, top5_share=sum(r["reads"] for r in ordered[:5])/tr)
json.dump(out, open(f"{D}/summary.json", "w"), indent=1)
print(json.dumps(out, indent=1))

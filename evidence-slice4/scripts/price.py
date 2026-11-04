import csv, json
CAST_NS, READ_NS = 11.0, 9.3       # measured today (bench/timings.txt); INF_* are inferred inlined-check scenarios
INF_CAST_NS, INF_READ_NS = 1.5, 1.0
node = {"compiler": 13.52, "services": 6.95}
out = {}
for w in ("compiler", "services"):
    rows = list(csv.DictReader(open(f"data/per-site-{w}.csv")))
    for r in rows:
        r["cost"] = (int(r["casts"]) * CAST_NS + int(r["reads"]) * READ_NS) / 1e9
    total = sum(r["cost"] for r in rows)
    ordered = sorted(rows, key=lambda r: -r["cost"])
    cum = 0; n80 = n50 = None
    for k, r in enumerate(ordered, 1):
        cum += r["cost"]
        if n50 is None and cum >= .5 * total: n50 = k
        if n80 is None and cum >= .8 * total: n80 = k
    tc = sum(int(r["casts"]) for r in rows); tr = sum(int(r["reads"]) for r in rows)
    ck = [r for r in rows if r["file"].endswith("/checker.ts")]
    out[w] = dict(cast_s=tc*CAST_NS/1e9, read_s=tr*READ_NS/1e9, total_s=total, share_of_node=total/node[w],
        inferred_inlined_total_s=(tc*INF_CAST_NS+tr*INF_READ_NS)/1e9, sites_for_50pct=n50, sites_for_80pct=n80,
        top5=[(f'{r["file"]}:{r["line"]}', r["target"], int(r["casts"]), int(r["reads"]), round(r["cost"],4), f'{r["cost"]/total:.1%}') for r in ordered[:5]],
        top5_share=sum(r["cost"] for r in ordered[:5])/total, top20_share=sum(r["cost"] for r in ordered[:20])/total,
        checker_only_s=sum(r["cost"] for r in ck))
    with open(f"data/price-{w}.csv", "w", newline="") as f:
        wr = csv.DictWriter(f, fieldnames=list(rows[0])); wr.writeheader(); wr.writerows(ordered)
json.dump(out, open("data/price.json", "w"), indent=1); print(json.dumps(out, indent=1))

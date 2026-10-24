import re, statistics, sys
out = "/tmp/bench/out"
def tv(path):
    text = open(path).read()
    user = float(re.search(r"User time \(seconds\): ([\d.]+)", text).group(1))
    sys_ = float(re.search(r"System time \(seconds\): ([\d.]+)", text).group(1))
    wall = re.search(r"Elapsed \(wall clock\) time \(h:mm:ss or m:ss\): (\S+)", text).group(1)
    parts = [float(p) for p in wall.split(":")]
    seconds = sum(p * 60 ** i for i, p in enumerate(reversed(parts)))
    rss = int(re.search(r"Maximum resident set size \(kbytes\): (\d+)", text).group(1))
    return seconds, user + sys_, rss
def log(path):
    units = []; runtime = {}; link = None; other = []
    for line in open(path):
        f = line.split()
        if f[0] == "cCompile" and f[1].startswith("runtime/"):
            runtime[f[-2]] = runtime.get(f[-2], 0) + 1
        elif f[0] == "cCompile":
            units.append((f[-2], float(f[-1]), f[1]))
        elif f[0] == "cLink":
            link = float(f[-1])
        else:
            other.append(f[0] + " " + f[-2])
    return units, runtime, link, other
print("| program | tree | wall s | CPU s | peak RSS MB | units | miss/off | sum unit s | largest unit s | link s | runtime lines | other |")
print("|---|---|---|---|---|---|---|---|---|---|---|---|")
for p in ["lint", "tsprinter", "json", "yaml"]:
    for t in ["base", "head"]:
        rows = [tv(f"{out}/time-{p}-{t}-{n}.txt") for n in (1, 2, 3)]
        logs = [log(f"{out}/build-{p}-{t}-{n}.log") for n in (1, 2, 3)]
        med = lambda xs: statistics.median(xs)
        units = [len(l[0]) for l in logs]
        sums = [sum(u[1] for u in l[0]) for l in logs]
        largest = [max(u[1] for u in l[0]) for l in logs]
        links = [l[2] for l in logs]
        status = sorted({"/".join(sorted({u[0] for u in l[0]})) for l in logs})
        rt = sorted({str(l[1]) for l in logs}); other = sorted({str(l[3]) for l in logs})
        big = max(logs[0][0], key=lambda u: u[1])
        print(f"| {p} | {t} | {med([r[0] for r in rows]):.2f} | {med([r[1] for r in rows]):.2f} | {med([r[2] for r in rows])/1024:.0f} | {'/'.join(map(str,sorted(set(units))))} | {status} | {med(sums):.2f} | {med(largest):.2f} | {med(links):.2f} | {rt} | {other} |")
        print(f"<!-- {p} {t} per-run wall {[round(r[0],2) for r in rows]} cpu {[round(r[1],2) for r in rows]} sums {[round(s,2) for s in sums]} largest-run1 {big[2]} {big[1]} -->")

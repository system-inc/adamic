#!/usr/bin/env python3
# Timestamps new lines in the -v log and build log, samples loadavg, and records unit .c sizes seen in adamic-units dirs.
import sys, time, os, glob
vlog, blog, out, tmp = sys.argv[1:5]
start = time.time()
pos = {vlog: 0, blog: 0}
partial = {vlog: b"", blog: b""}
seen = {}
last_load = 0
with open(out, "w") as o:
    while True:
        now = time.time()
        for path in (vlog, blog):
            try:
                with open(path, "rb") as f:
                    f.seek(pos[path]); data = f.read(); pos[path] += len(data)
            except FileNotFoundError:
                continue
            data = partial[path] + data
            lines = data.split(b"\n"); partial[path] = lines.pop()
            for line in lines:
                o.write("%.3f %s %s\n" % (now - start, "V" if path == vlog else "B", line.decode(errors="replace")))
        for c in glob.glob(os.path.join(tmp, "adamic-units-*", ".unit-*", "*")):
            name = os.path.basename(c)
            if name.endswith(".o") or name in seen: continue
            if not name.endswith(".c") and name != "units.h": continue
            try:
                size = os.path.getsize(c)
            except OSError:
                continue
            if size == 0: continue
            seen[name] = size
            o.write("%.3f U %s %d\n" % (now - start, name, size))
        if now - last_load >= 1:
            last_load = now
            o.write("%.3f L %s\n" % (now - start, open("/proc/loadavg").read().strip()))
        o.flush()
        if os.path.exists(out + ".stop"):
            break
        time.sleep(0.02)

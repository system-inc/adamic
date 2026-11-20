#!/bin/bash
# #4smyx18: bench/environment.a runs 2.35x Node (J-vpa5442). The emitted C of checkTypeRelatedTo and its
# five nested functions, the counted row, perf stat of native and Node (instructions:u, cycles, task-clock),
# then perf record of native's user cycles by function with the callers of the top two. Output checked
# against Node first. Never lands. Needs perf_event_paranoid at 2 or lower.
set -u
T=$(mktemp -d)
trap 'rm -rf "$T"' EXIT
echo "box $(hostname) $(nproc) cpus, $(uname -r), tree $(git rev-parse HEAD)"
echo "load before $(cut -d' ' -f1-3 /proc/loadavg); perf_event_paranoid $(cat /proc/sys/kernel/perf_event_paranoid)"
go build -buildvcs=false -o "$T/adamic" ./cmd/adamic || { echo "FAIL adamic build"; exit 1; }
source=bench/environment.a
"$T/adamic" build "$source" -o "$T/native" > "$T/b.log" 2>&1 || { cat "$T/b.log"; echo "FAIL native build"; exit 1; }
want=$(node --disable-warning=ExperimentalWarning oracle/node.mjs "$source")
got=$("$T/native")
[ "$got" = "$want" ] && echo "PASS native prints Node's output ($got)" || { echo "FAIL native $got, Node $want"; exit 1; }
"$T/adamic" build --count "$source" -o "$T/counted" > /dev/null 2>&1 && "$T/counted" 2>&1 >/dev/null | grep -E "counts" | head -2
echo "== the emitted C of checkTypeRelatedTo and what it calls"
"$T/adamic" c "$source" > "$T/environment.c" 2> "$T/c.log" || { head -5 "$T/c.log"; echo "NOTE adamic c failed"; }
echo "emitted C: $(wc -l < "$T/environment.c") lines"
python3 -I - "$T/environment.c" <<'PY'
import re, sys
text = open(sys.argv[1]).read()
# Every function whose name says checkTypeRelatedTo or one of its nested functions, whole, up to 400 lines.
names = ["checkTypeRelatedTo", "compare", "descend", "finish", "related", "result"]
starts = [m for m in re.finditer(r"^[^\s#/][^\n;]*\b(\w*(?:" + "|".join(names) + r")\w*)\s*\([^;{]*\)\s*\{", text, re.M)]
shown = 0
for m in starts:
    depth, i = 0, m.end() - 1
    while i < len(text):
        if text[i] == "{": depth += 1
        elif text[i] == "}":
            depth -= 1
            if depth == 0: break
        i += 1
    body = text[m.start():i + 1]
    lines = body.splitlines()
    print("\n".join(lines[:80]))
    if len(lines) > 80: print(f"... ({len(lines)} lines)")
    print()
    shown += len(lines[:80])
    if shown > 400: break
if not starts: print("no function matched; first 60 lines of the file follow"); print("\n".join(text.splitlines()[:60]))
PY
echo "== perf stat, user counters, 3 runs each"
for side in native node; do
	if [ $side = native ]; then command=("$T/native"); else command=(node --disable-warning=ExperimentalWarning oracle/node.mjs "$source"); fi
	perf stat -x, -r 3 -e instructions:u,cycles:u,task-clock:u,page-faults:u "${command[@]}" > /dev/null 2> "$T/stat.$side"
	awk -F, -v s=$side '{print s, $1, $3}' "$T/stat.$side"
done
echo "== perf record, native, user cycles, top 25 symbols"
perf record -q -e cycles:u -g -o "$T/perf.data" "$T/native" > /dev/null 2>&1
perf report -q -i "$T/perf.data" --no-children --sort symbol --stdio 2>/dev/null | grep -E '^\s+[0-9.]+%' | head -25
echo "== callers of the top two"
for symbol in $(perf report -q -i "$T/perf.data" --no-children --sort symbol --stdio 2>/dev/null | grep -E '^\s+[0-9.]+%' | head -2 | awk '{print $3}'); do
	echo "-- $symbol"
	perf report -q -i "$T/perf.data" --no-children --symbols "$symbol" -G --stdio 2>/dev/null | grep -v '^$' | head -25
done
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
echo PASS profile

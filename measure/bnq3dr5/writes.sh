#!/bin/bash
# #bnq3dr5: does bench/holdout/mandelbrot.a's 16 s in the kernel come from one write system call per byte
# it hands process.stdout.write? A copy at h = w = 1000 must print 125,000 bytes after its header; strace
# counts the write calls natively and on Node, with output checked against Node. Then full size, timed
# once each with system time, standard output to a pipe as bench gives it. Never lands.
set -u
T=$(mktemp -d)
trap 'rm -rf "$T"' EXIT
echo "box $(hostname) $(nproc) cpus, $(uname -r), tree $(git rev-parse HEAD)"
echo "load before $(cut -d' ' -f1-3 /proc/loadavg)"
command -v strace > /dev/null || { echo "FAIL no strace"; exit 1; }
go build -buildvcs=false -o "$T/adamic" ./cmd/adamic || { echo "FAIL adamic build"; exit 1; }
sed 's/^\tw = h;$/\tw = h;/; s/^const h = 16000,$/const h = 1000,/' bench/holdout/mandelbrot.a > "$T/small.a"
grep -q '^const h = 1000,$' "$T/small.a" || { echo "FAIL the size didn't change"; exit 1; }
"$T/adamic" build "$T/small.a" -o "$T/small" > "$T/b.log" 2>&1 || { cat "$T/b.log"; echo "FAIL small build"; exit 1; }
"$T/adamic" build bench/holdout/mandelbrot.a -o "$T/full" > "$T/b.log" 2>&1 || { cat "$T/b.log"; echo "FAIL full build"; exit 1; }
node --disable-warning=ExperimentalWarning oracle/node.mjs "$T/small.a" > "$T/node.out"
"$T/small" > "$T/native.out"
cmp -s "$T/node.out" "$T/native.out" && echo "PASS at 1000 native prints Node's $(wc -c < "$T/native.out") bytes" || { echo "FAIL at 1000 the outputs differ"; exit 1; }
echo "== write calls at 1000, expected 125,000 bytes after the header"
for side in native node; do
	if [ $side = native ]; then command=("$T/small"); else command=(node --disable-warning=ExperimentalWarning oracle/node.mjs "$T/small.a"); fi
	strace -f -c -e trace=write,writev,pwrite64 -o "$T/strace.$side" "${command[@]}" | cat > /dev/null
	echo "-- $side"; grep -E "calls|write|total" "$T/strace.$side"
done
echo "== full size, standard output to a pipe, wall/user/system"
/usr/bin/time -f "native %e %U %S" "$T/full" 2>&1 >/dev/null | cat > /dev/null
for side in native node; do
	if [ $side = native ]; then command=("$T/full"); else command=(node --disable-warning=ExperimentalWarning oracle/node.mjs bench/holdout/mandelbrot.a); fi
	/usr/bin/time -f "$side %e s wall, %U s user, %S s system" -o "$T/time.$side" "${command[@]}" | wc -c > "$T/bytes.$side"
	echo "$(cat "$T/time.$side"), $(cat "$T/bytes.$side") bytes"
done
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
echo PASS writes

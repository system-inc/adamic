#!/bin/bash
# #7z36qn6 lever 2, measured before anything is built: hello world's startup against C at main (lever 1 landed), and
# the same Adamic hello with the signal relay compiled out, the most a lazy relay could ever save. Never lands.
set -uo pipefail
echo "box $(hostname) $(nproc) cpus, $(uname -r), load $(cut -d' ' -f1-3 /proc/loadavg), tree $(git rev-parse HEAD)"
grep -m1 "model name" /proc/cpuinfo
T="$TMPDIR/lever2"
mkdir -p "$T"
A=internal/native_runtime/c/adamic.c
cp "$A" "$T/adamic.c.orig"
printf 'console.log("hello world");\n' > "$T/hello.a"
printf '#include <stdio.h>\nint main(void) { puts("hello world"); return 0; }\n' > "$T/hello.c"
clang -O2 -o "$T/c-hello" "$T/hello.c" || { echo "FAIL clang hello"; exit 1; }
clang -O2 -o "$T/spawnloop" measure/7z36qn6/spawnloop.c || { echo "FAIL clang spawnloop"; exit 1; }

go build -buildvcs=false -o "$T/adamic-main" ./cmd/adamic > "$T/b.log" 2>&1 || { cat "$T/b.log"; echo "FAIL adamic build"; exit 1; }
python3 -I - "$A" <<'PY' || { cp "$T/adamic.c.orig" "$A"; echo "FAIL relay seam"; exit 1; }
import sys
path = sys.argv[1]
text = open(path).read()
start = "\tstop_start();\n"
end = "\t\tstop_with(signal_number);\n\t}\n#endif\n"
if text.count(start) != 1 or text.count(end) != 1:
    sys.exit("relay seams not found once")
text = text.replace(start, "#if 0 /* lever 2 measure: no relay */\n" + start, 1).replace(end, end + "#endif\n", 1)
open(path, "w").write(text)
PY
go build -buildvcs=false -o "$T/adamic-norelay" ./cmd/adamic > "$T/b.log" 2>&1 || { cp "$T/adamic.c.orig" "$A"; cat "$T/b.log"; echo "FAIL norelay build"; exit 1; }
cp "$T/adamic.c.orig" "$A"
git diff --quiet -- "$A" && echo "PASS adamic.c restored" || { echo "FAIL adamic.c not restored"; exit 1; }
for variant in main norelay; do
	XDG_CACHE_HOME="$T/cache-$variant" ADAMIC_BUILD_CACHE_DIR="$T/build-$variant" "$T/adamic-$variant" build "$T/hello.a" -o "$T/adamic-$variant-hello" > "$T/b.log" 2>&1 || { cat "$T/b.log"; echo "FAIL hello build $variant"; exit 1; }
done
for binary in c-hello adamic-main-hello adamic-norelay-hello; do
	[ "$("$T/$binary")" = "hello world" ] || { echo "FAIL $binary prints another line"; exit 1; }
done
echo "PASS all three print hello world"

if command -v strace > /dev/null; then
	for binary in c-hello adamic-main-hello adamic-norelay-hello; do
		strace -f -c -o "$T/$binary.count" "$T/$binary" > /dev/null
		total=$(awk '/^-+ /{getline; print $(NF-1)}' "$T/$binary.count" | tail -1)
		sigactions=$(awk '$NF=="rt_sigaction"{print $4}' "$T/$binary.count")
		threads=$(awk '$NF=="clone3"||$NF=="clone"{print $4}' "$T/$binary.count")
		echo "== $binary: $total system calls, rt_sigaction ${sigactions:-0}, clone ${threads:-0}"
	done
else
	echo "GAP no strace: syscall counts owed"
fi

# A warm-up block first, then the measured blocks, each interleaved and mirrored, load stated before and after.
"$T/spawnloop" 200 pipe "$T/c-hello" "$T/adamic-main-hello" "$T/adamic-norelay-hello" > /dev/null || { echo "FAIL warm-up"; exit 1; }
for mode in pipe null; do
	for block in 1 2 3; do
		echo "== $mode block $block, load $(cut -d' ' -f1-3 /proc/loadavg)"
		"$T/spawnloop" 2000 "$mode" "$T/c-hello" "$T/adamic-main-hello" "$T/adamic-norelay-hello" || { echo "FAIL spawnloop"; exit 1; }
	done
done
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
echo VERDICT PASS

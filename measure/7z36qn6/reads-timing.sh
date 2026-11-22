#!/bin/bash
# #7z36qn6: lever (a)'s win timed beside main's read in the same blocks, and the thread's cost split into its create
# and its join at exit (a variant that registers no stop_end, measure only, never a design). Never lands.
set -uo pipefail
echo "box $(hostname) $(nproc) cpus, $(uname -r), load $(cut -d' ' -f1-3 /proc/loadavg), tree $(git rev-parse HEAD)"
grep -m1 "model name" /proc/cpuinfo
T="$TMPDIR/timing"
mkdir -p "$T"
A=internal/native_runtime/c/adamic.c
cp "$A" "$T/adamic.c.orig"
printf 'console.log("hello world");\n' > "$T/hello.a"
printf '#include <stdio.h>\nint main(void) { puts("hello world"); return 0; }\n' > "$T/hello.c"
clang -O2 -o "$T/c" "$T/hello.c" || { echo "FAIL clang hello"; exit 1; }
clang -O2 -o "$T/spawnloop" measure/7z36qn6/spawnloop.c || { echo "FAIL clang spawnloop"; exit 1; }
variants="head withread nojoin norelay"
for variant in $variants; do
	cp "$T/adamic.c.orig" "$A"
	python3 -I - "$A" "$variant" <<'PY' || { cp "$T/adamic.c.orig" "$A"; echo "FAIL seam for $variant"; exit 1; }
import sys
path, variant = sys.argv[1], sys.argv[2]
text = open(path).read()
def swap(before, after):
    global text
    if text.count(before) != 1:
        sys.exit(f"{variant}: seam not found once: {before!r}")
    text = text.replace(before, after, 1)
if variant == "withread":
    # main's stop_with, which asked each signal what it held.
    swap("static void stop_with(int signal_number) {\n\tstruct sigaction handler;\n", "static void stop_with(int signal_number) {\n\tstruct sigaction current;\n\tif (sigaction(signal_number, NULL, &current) != 0 || current.sa_handler == SIG_IGN) {\n\t\treturn;\n\t}\n\tstruct sigaction handler;\n")
elif variant == "nojoin":
    # The thread is made, but nothing wakes and joins it at exit.
    swap("atexit(stop_end) != 0", "(0 && atexit(stop_end) != 0)")
    swap("static void stop_end(void) {", "__attribute__((unused)) static void stop_end(void) {")
elif variant == "norelay":
    swap("\tstop_start();\n", "#if 0 /* measure: no relay */\n\tstop_start();\n")
    swap("\t\tstop_with(signal_number);\n\t}\n#endif\n", "\t\tstop_with(signal_number);\n\t}\n#endif\n#endif\n")
open(path, "w").write(text)
PY
	go build -buildvcs=false -o "$T/adamic-$variant" ./cmd/adamic > "$T/b.log" 2>&1 || { cp "$T/adamic.c.orig" "$A"; cat "$T/b.log"; echo "FAIL adamic build $variant"; exit 1; }
	cp "$T/adamic.c.orig" "$A"
	XDG_CACHE_HOME="$T/cache-$variant" ADAMIC_BUILD_CACHE_DIR="$T/build-$variant" "$T/adamic-$variant" build "$T/hello.a" -o "$T/$variant" > "$T/b.log" 2>&1 || { cat "$T/b.log"; echo "FAIL hello build $variant"; exit 1; }
	[ "$("$T/$variant")" = "hello world" ] || { echo "FAIL $variant prints another line"; exit 1; }
done
git diff --quiet -- "$A" && echo "PASS adamic.c restored, every variant prints hello world" || { echo "FAIL adamic.c not restored"; exit 1; }
if command -v strace > /dev/null; then
	for binary in c $variants; do
		strace -f -o "$T/$binary.trace" "$T/$binary" > /dev/null
		echo "== $binary: $(grep -c 'rt_sigaction(' "$T/$binary.trace") rt_sigaction, $(grep -c 'clone3(' "$T/$binary.trace") clone3, $(wc -l < "$T/$binary.trace") calls"
	done
fi
binaries=""
for binary in c $variants; do binaries="$binaries $T/$binary"; done
"$T/spawnloop" 200 pipe $binaries > /dev/null || { echo "FAIL warm-up"; exit 1; }
for block in 1 2 3 4 5; do
	echo "== pipe block $block, load $(cut -d' ' -f1-3 /proc/loadavg)"
	"$T/spawnloop" 2000 pipe $binaries || { echo "FAIL spawnloop"; exit 1; }
done
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
echo VERDICT PASS

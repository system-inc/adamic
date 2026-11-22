#!/bin/bash
# #7z36qn6 lever 2's split, measured before anything is built: which part of the signal relay's cost is the thread and
# which the handlers, and whether dropping stop_with's read or shrinking the thread's stack wins it back. Each variant
# is adamic.c edited for the measure only, never a head. Never lands.
set -uo pipefail
echo "box $(hostname) $(nproc) cpus, $(uname -r), load $(cut -d' ' -f1-3 /proc/loadavg), tree $(git rev-parse HEAD)"
grep -m1 "model name" /proc/cpuinfo
T="$TMPDIR/split"
mkdir -p "$T"
A=internal/native_runtime/c/adamic.c
cp "$A" "$T/adamic.c.orig"
printf 'console.log("hello world");\n' > "$T/hello.a"
printf '#include <stdio.h>\nint main(void) { puts("hello world"); return 0; }\n' > "$T/hello.c"
clang -O2 -o "$T/c" "$T/hello.c" || { echo "FAIL clang hello"; exit 1; }
clang -O2 -o "$T/spawnloop" measure/7z36qn6/spawnloop.c || { echo "FAIL clang spawnloop"; exit 1; }

variants="main norelay thread handlers noreads smallstack"
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
start = "\tstop_start();\n"
loop = "\t\tstop_with(stop_signals[index]);\n"
realtime = "\t\tstop_with(signal_number);\n"
read = "\tstruct sigaction current;\n\tif (sigaction(signal_number, NULL, &current) != 0 || current.sa_handler == SIG_IGN) {\n\t\treturn;\n\t}\n"
create = "\tint error = pthread_create(&stop_thread, NULL, stop_loop, NULL);\n"
if variant == "norelay":
    swap(start, "#if 0 /* measure: no relay */\n" + start)
    swap(realtime + "\t}\n#endif\n", realtime + "\t}\n#endif\n#endif\n")
elif variant == "thread":
    # The thread and its pipe, with no handler installed.
    swap(loop, "\t\t(void)stop_signals[index];\n")
    swap(realtime, "\t\t(void)signal_number;\n")
    swap(start, start + "\t(void)stop_with;\n")
elif variant == "handlers":
    # Every handler installed, with no thread and no pipe behind them: never signalled in the measure.
    swap(start, "\t(void)stop_start;\n")
elif variant == "noreads":
    swap(read, "")
elif variant == "smallstack":
    swap(create, "\tpthread_attr_t attributes;\n\tpthread_attr_init(&attributes);\n\tpthread_attr_setstacksize(&attributes, 64 * 1024);\n\tint error = pthread_create(&stop_thread, &attributes, stop_loop, NULL);\n\tpthread_attr_destroy(&attributes);\n")
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
		strace -f -c -o "$T/$binary.count" "$T/$binary" > /dev/null
		echo "== $binary: rt_sigaction $(awk '$NF=="rt_sigaction"{print $4}' "$T/$binary.count"), clone3 $(awk '$NF=="clone3"{print $4}' "$T/$binary.count"), mmap $(awk '$NF=="mmap"{print $4}' "$T/$binary.count"); $(grep -E ' total$' "$T/$binary.count" | tr -s ' ')"
	done
fi

binaries=""
for binary in c $variants; do binaries="$binaries $T/$binary"; done
"$T/spawnloop" 200 pipe $binaries > /dev/null || { echo "FAIL warm-up"; exit 1; }
for block in 1 2 3; do
	echo "== pipe block $block, load $(cut -d' ' -f1-3 /proc/loadavg)"
	"$T/spawnloop" 2000 pipe $binaries || { echo "FAIL spawnloop"; exit 1; }
done
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
echo VERDICT PASS

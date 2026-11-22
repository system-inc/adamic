#!/bin/bash
# #7z36qn6: what libm's dynamic load costs a hello world's start (Seats' lead, J on #7z36qn6). C hello as is and with
# libm forced into its NEEDED list, Adamic hello as built and with libm linked --as-needed (native.go's -lm, for
# the measure only, never a head), and which libm symbols hello leaves undefined. A static libm can't link into a
# dynamic executable on glibc 2.39: its pow ifunc needs _dl_x86_cpu_features (J-7y1adbd). Never lands.
set -uo pipefail
echo "box $(hostname) $(nproc) cpus, $(uname -r), load $(cut -d' ' -f1-3 /proc/loadavg), tree $(git rev-parse HEAD)"
grep -m1 "model name" /proc/cpuinfo
ldd --version | head -1
T="$TMPDIR/libm"
mkdir -p "$T"
N=internal/native/native.go
cp "$N" "$T/native.go.orig"
printf 'console.log("hello world");\n' > "$T/hello.a"
printf '#include <stdio.h>\nint main(void) { puts("hello world"); return 0; }\n' > "$T/hello.c"
clang -O2 -o "$T/c" "$T/hello.c" || { echo "FAIL clang c"; exit 1; }
clang -O2 -o "$T/c-libm" "$T/hello.c" -Wl,--no-as-needed -lm || { echo "FAIL clang c-libm"; exit 1; }
clang -O2 -o "$T/spawnloop" measure/7z36qn6/spawnloop.c || { echo "FAIL clang spawnloop"; exit 1; }
for archive in $(clang -print-file-name=libm.a) /usr/lib/x86_64-linux-gnu/libm.a; do
	[ -e "$archive" ] && echo "static libm: $archive ($(head -c 200 "$archive" | tr -d '\0' | head -1 | cut -c1-80))"
done

go build -buildvcs=false -o "$T/adamic-dynamic" ./cmd/adamic > "$T/b.log" 2>&1 || { cat "$T/b.log"; echo "FAIL adamic build"; exit 1; }
python3 -I - "$N" <<'PY' || { cp "$T/native.go.orig" "$N"; echo "FAIL seam"; exit 1; }
import sys
path = sys.argv[1]
text = open(path).read()
seam = '\targuments = append(arguments, "-lm")\n'
if text.count(seam) != 1:
    sys.exit("seam not found once")
open(path, "w").write(text.replace(seam, '\targuments = append(arguments, "-Wl,--as-needed", "-lm", "-Wl,--no-as-needed")\n', 1))
PY
go build -buildvcs=false -o "$T/adamic-static" ./cmd/adamic > "$T/b.log" 2>&1 || { cp "$T/native.go.orig" "$N"; cat "$T/b.log"; echo "FAIL as-needed build"; exit 1; }
cp "$T/native.go.orig" "$N"
git diff --quiet -- "$N" && echo "PASS native.go restored" || { echo "FAIL native.go not restored"; exit 1; }
for variant in dynamic static; do
	XDG_CACHE_HOME="$T/cache-$variant" ADAMIC_BUILD_CACHE_DIR="$T/build-$variant" "$T/adamic-$variant" build "$T/hello.a" -o "$T/adamic-$variant-hello" > "$T/b.log" 2>&1 || { cat "$T/b.log"; echo "FAIL hello build $variant "; exit 1; }
done
for binary in c c-libm adamic-dynamic-hello adamic-static-hello; do
	[ "$("$T/$binary")" = "hello world" ] || { echo "FAIL $binary prints another line"; exit 1; }
	echo "== $binary: $(stat -c %s "$T/$binary") bytes, NEEDED $(readelf -d "$T/$binary" | grep NEEDED | grep -o '\[.*\]' | tr '\n' ' ')"
	if command -v strace > /dev/null; then
		strace -f -o "$T/$binary.trace" "$T/$binary" > /dev/null
		echo "   $(wc -l < "$T/$binary.trace") calls, $(grep -c 'openat(' "$T/$binary.trace") openat, $(grep -c 'mmap(' "$T/$binary.trace") mmap"
	fi
done
libm_exports=$(nm -D --defined-only "$(readlink -f /lib/x86_64-linux-gnu/libm.so.6)" | awk '{print $3}' | sed 's/@.*//' | sort -u)
echo "== libm symbols adamic hello leaves undefined: $(nm -D --undefined-only "$T/adamic-dynamic-hello" | awk '{print $2}' | sed 's/@.*//' | sort -u | comm -12 - <(echo "$libm_exports") | tr '\n' ' ')"
binaries="$T/c $T/c-libm $T/adamic-dynamic-hello $T/adamic-static-hello"
"$T/spawnloop" 200 pipe $binaries > /dev/null || { echo "FAIL warm-up"; exit 1; }
for block in 1 2 3 4 5; do
	echo "== pipe block $block, load $(cut -d' ' -f1-3 /proc/loadavg)"
	"$T/spawnloop" 2000 pipe $binaries || { echo "FAIL spawnloop"; exit 1; }
done
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
echo VERDICT PASS

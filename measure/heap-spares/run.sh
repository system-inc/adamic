#!/usr/bin/env bash
# run.sh: heap.c's spare chunks returned past a per-thread cap, timed against keep-forever (#w7cg3be,
# a measurement, never to land). Run from the commit's worktree: bash measure/heap-spares/run.sh
#
# It lays the submodules, builds adamic five times with heap.c's ADAMIC_HEAP_SPARE_CAP line rewritten
# (unset, 0, 1, 4 and 16 MB), builds heap_churn with each adamic, checks it prints the same under every
# cap, then times it under GNU time, glibc as shipped and pinned, interleaved, best of 5 by wall. Its
# resident memory at peak, just dropped and idle it reads itself, three takes a cap and mode.
# Everything is written under $TMPDIR; the table is the log's last lines, ending with == end.
set -u -o pipefail

export TMPDIR=$HOME/scratch/system_adamic_runtime_performance/tmp
mkdir -p "$TMPDIR"
root=$(pwd)
work=$(mktemp -d "$TMPDIR/heap-spares.XXXXXX")
heap=$root/internal/native/runtime/heap.c
here=$root/measure/heap-spares
caps=(keep 0 1 4 16)
rounds=5
pinned=(MALLOC_MMAP_THRESHOLD_=131072 MALLOC_TRIM_THRESHOLD_=131072)

say() { echo "[$((SECONDS / 60))m$((SECONDS % 60))s] $*"; }
finish() {
	echo "== end"
	exit "$1"
}
fail() {
	say "FAIL $1"
	[ $# -gt 1 ] && [ -f "$2" ] && tail -n 25 "$2"
	finish 1
}

say "work $work"
for environment in "$HOME/adamic-tools/env.sh" /workspace/adamic-tools/env.sh; do
	if [ -f "$environment" ]; then
		# shellcheck disable=SC1090
		. "$environment" > /dev/null 2>&1
		say "sourced $environment"
		break
	fi
done
command -v go > /dev/null || fail "no go on PATH"
[ -x /usr/bin/time ] && /usr/bin/time -f '%e' -o "$work/time-probe" true 2> /dev/null || fail "no GNU /usr/bin/time"
say "box $(hostname) $(uname -sr), $(nproc) cpus, $(ldd --version 2> /dev/null | head -n 1)"
say "$(go version); $(command -v clang > /dev/null && clang --version | head -n 1 || echo 'clang not on PATH')"
say "page $(getconf PAGESIZE), at $(git -C "$root" rev-parse HEAD)"

git -c protocol.file.allow=always submodule update --init --recursive > "$work/submodules.log" 2>&1 || fail "submodules" "$work/submodules.log"
say "PASS submodules"

# One adamic a cap, heap.c's line rewritten and put back after each build.
cp "$heap" "$work/heap.c.original"
trap 'cp "$work/heap.c.original" "$heap"' EXIT
for cap in "${caps[@]}"; do
	cp "$work/heap.c.original" "$heap"
	if [ "$cap" != keep ]; then
		sed -i "s|^// #define ADAMIC_HEAP_SPARE_CAP ((size_t)0)\$|#define ADAMIC_HEAP_SPARE_CAP ((size_t)$((cap * 1048576)))|" "$heap"
		[ "$(grep -c '^#define ADAMIC_HEAP_SPARE_CAP' "$heap")" = 1 ] || fail "cap $cap: heap.c's line wasn't rewritten"
	fi
	go build -buildvcs=false -o "$work/adamic-$cap" ./cmd/adamic > "$work/go-build-$cap.log" 2>&1 || fail "go build at cap $cap" "$work/go-build-$cap.log"
done
cp "$work/heap.c.original" "$heap"
say "PASS adamic built at caps ${caps[*]}"

# heap_churn alone this run: the everyday programs read the same at keep and 1 MB on J-tht6x7c and
# J-wfdx4g6.
churn=$here/heap_churn.ts
mkdir -p "$work/bin" "$work/out"
for cap in "${caps[@]}"; do
	"$work/adamic-$cap" build "$churn" -o "$work/bin/heap_churn.$cap" > "$work/build-$cap.log" 2>&1 || fail "heap_churn's build at cap $cap" "$work/build-$cap.log"
done
say "PASS heap_churn built at every cap"

# Counted once, untimed: heap values against values in regions.
if "$work/adamic-keep" build "$churn" -o "$work/bin/heap_churn.counted" --count > "$work/build-counted.log" 2>&1; then
	"$work/bin/heap_churn.counted" > /dev/null 2> "$work/counted.err"
	say "count heap_churn: $(grep -o 'adamic: counts:.*' "$work/counted.err" | tail -n 1)"
else
	say "count heap_churn: the counted build failed"
	tail -n 5 "$work/build-counted.log"
fi

# Every cap prints on stdout what keep-forever prints; its resident memory goes to stderr.
for cap in "${caps[@]}"; do
	"$work/bin/heap_churn.$cap" > "$work/out/heap_churn.$cap" 2> "$work/out/heap_churn.$cap.err" || fail "heap_churn at cap $cap exited $?" "$work/out/heap_churn.$cap.err"
	cmp -s "$work/out/heap_churn.keep" "$work/out/heap_churn.$cap" || fail "heap_churn prints differently at cap $cap" "$work/out/heap_churn.$cap"
done
say "PASS outputs match across caps ($(head -c 60 "$work/out/heap_churn.keep"))"
say "first read at keep: $(cat "$work/out/heap_churn.keep.err")"
# Where /proc/self/status can't be read in the program, an outside poller reads VmRSS until the
# program writes its rss line, which it does after the spin, before it exits.
inside=1
grep -q 'rss peak [0-9]' "$work/out/heap_churn.keep.err" || inside=0
[ "$inside" = 1 ] && say "PASS heap_churn reads its own VmRSS" || say "NOTE heap_churn can't read /proc/self/status; idle is polled until its rss line, peak and dropped are -"

# Timing: rounds outermost, then modes and caps, so load drifts across every cap alike.
data=$work/timings.txt
: > "$data"
loadBefore=$(cut -d' ' -f1-3 /proc/loadavg)
say "load before $loadBefore"
for round in $(seq "$rounds"); do
	for mode in shipped pinned; do
		extra=()
		[ "$mode" = pinned ] && extra=("${pinned[@]}")
		for cap in "${caps[@]}"; do
			env ${extra[@]+"${extra[@]}"} /usr/bin/time -f '%e %U %S %M %R' -o "$work/time.txt" "$work/bin/heap_churn.$cap" > /dev/null 2>&1 || fail "heap_churn at cap $cap, $mode, exited nonzero" "$work/time.txt"
			echo "$cap $mode $round $(tail -n 1 "$work/time.txt")" >> "$data"
		done
	done
done
loadAfter=$(cut -d' ' -f1-3 /proc/loadavg)
say "load after $loadAfter"

# Resident memory, three takes a cap and mode: peak, dropped and idle in KB.
resident=$work/resident.txt
: > "$resident"
for take in 1 2 3; do
	for mode in shipped pinned; do
		extra=()
		[ "$mode" = pinned ] && extra=("${pinned[@]}")
		for cap in "${caps[@]}"; do
			err=$work/out/take.err
			: > "$err"
			env ${extra[@]+"${extra[@]}"} "$work/bin/heap_churn.$cap" > /dev/null 2> "$err" &
			pid=$!
			last=-1
			if [ "$inside" = 0 ]; then
				while ! grep -q '^rss ' "$err"; do
					sample=$(awk '$1 == "VmRSS:" { print $2 }' "/proc/$pid/status" 2> /dev/null)
					[ -n "$sample" ] || break
					last=$sample
					sleep 0.005
				done
			fi
			wait "$pid" || fail "heap_churn at cap $cap, $mode, take $take, exited nonzero" "$err"
			read -r peak dropped idle <<< "$(awk '/^rss peak/ { print $3, $5, $7 }' "$err")"
			[ "$inside" = 0 ] && idle=$last
			echo "$cap $mode $take ${peak:--1} ${dropped:--1} ${idle:--1}" >> "$resident"
		done
	done
done
say "PASS timed ${rounds} rounds and 3 resident takes"
cp "$data" "$resident" "$TMPDIR/" 2> /dev/null

# The table: per cap and mode, the best wall's run (wall, user, sys s; minor faults), and the median
# of three takes of peak, dropped and idle resident MB (- where it couldn't be read).
echo "== box $(hostname), load $loadBefore before and $loadAfter after the timing block (1, 5, 15 min)"
echo "== heap_churn: best of $rounds by wall (that run's user, sys, minor faults); resident MB, median of 3 takes"
echo "cap   glibc    |  wall   user    sys   faults |  peakMB  droppedMB  idleMB"
awk -v order="${caps[*]}" '
	function median(a, b, c) {
		if ((a <= b && b <= c) || (c <= b && b <= a)) return b
		if ((b <= a && a <= c) || (c <= a && a <= b)) return a
		return c
	}
	function megabytes(kilobytes) { return kilobytes < 0 ? "-" : sprintf("%.1f", kilobytes / 1024) }
	FILENAME == ARGV[1] {
		key = $1 ":" $2
		if (!(key in wall) || $4 < wall[key]) { wall[key] = $4; user[key] = $5; sys[key] = $6; faults[key] = $8 }
		next
	}
	{
		key = $1 ":" $2
		takes[key]++
		peak[key, takes[key]] = $4; dropped[key, takes[key]] = $5; idle[key, takes[key]] = $6
	}
	END {
		count = split(order, caps, " ")
		for (mode = 1; mode <= 2; mode++) {
			name = mode == 1 ? "shipped" : "pinned"
			for (index_ = 1; index_ <= count; index_++) {
				key = caps[index_] ":" name
				label = caps[index_] == "keep" ? "keep" : caps[index_] "M"
				printf("%-5s %-8s | %5.2f  %5.2f  %5.2f  %7d | %7s  %9s  %6s\n", label, name, wall[key], user[key], sys[key], faults[key], megabytes(median(peak[key, 1], peak[key, 2], peak[key, 3])), megabytes(median(dropped[key, 1], dropped[key, 2], dropped[key, 3])), megabytes(median(idle[key, 1], idle[key, 2], idle[key, 3])))
			}
		}
	}' "$data" "$resident"
finish 0

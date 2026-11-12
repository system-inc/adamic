#!/usr/bin/env bash
# run.sh: heap.c's spare chunks returned past a per-thread cap, timed against keep-forever (#w7cg3be,
# a measurement, never to land). Run from the commit's worktree: bash measure/heap-spares/run.sh
#
# It lays the submodules, builds adamic five times with heap.c's ADAMIC_HEAP_SPARE_CAP line rewritten
# (unset, 0, 1, 4 and 16 MB), builds each program with each adamic, checks every program prints the same
# under every cap, then times them under GNU time, glibc as shipped and pinned, interleaved, best of 5
# by wall. heap_churn's idle memory is the last VmRSS read every 5 ms while it runs, three takes a cap.
# Everything is written under $TMPDIR; the table is the log's last lines, ending with == end.
set -u -o pipefail

export TMPDIR=$HOME/scratch/system_adamic_runtime_performance/tmp
mkdir -p "$TMPDIR"
root=$(pwd)
work=$(mktemp -d "$TMPDIR/heap-spares.XXXXXX")
heap=$root/internal/native/runtime/heap.c
here=$root/measure/heap-spares
caps=(keep 0 1 4 16)
everyday=(keep 1)
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

# name|source|environment for a run|caps it's built at
rows=(
	"heap_churn|$here/heap_churn.ts||all"
	"churn_par/1|$here/heap_churn_parallel.a|ADAMIC_THREADS=1|all"
	"churn_par/4|$here/heap_churn_parallel.a|ADAMIC_THREADS=4|all"
	"word_count|$root/bench/word_count.ts||everyday"
	"tokenizer|$root/bench/tokenizer.ts||everyday"
	"sort|$root/bench/sort.ts||everyday"
)
mkdir -p "$work/bin" "$work/out"
binary() { echo "$work/bin/$(basename "$1" | sed 's/\.[a-z]*$//').$2"; }
row_caps() { if [ "$1" = all ]; then echo "${caps[@]}"; else echo "${everyday[@]}"; fi; }
live=()
for row in "${rows[@]}"; do
	IFS='|' read -r name source environment which <<< "$row"
	built=1
	for cap in $(row_caps "$which"); do
		target=$(binary "$source" "$cap")
		[ -x "$target" ] && continue
		if ! "$work/adamic-$cap" build "$source" -o "$target" > "$work/build-$(basename "$target").log" 2>&1; then
			say "SKIP $name: its build at cap $cap failed"
			tail -n 8 "$work/build-$(basename "$target").log"
			built=0
			break
		fi
	done
	[ "$built" = 1 ] && live+=("$row")
done
[ "${#live[@]}" -gt 0 ] || fail "no program built"
say "PASS built ${#live[@]} of ${#rows[@]} rows"

# Counted once, untimed: heap values against values in regions.
for source in "$here/heap_churn.ts" "$here/heap_churn_parallel.a"; do
	counted="$work/bin/$(basename "$source").counted"
	if "$work/adamic-keep" build "$source" -o "$counted" --count > "$work/build-counted.log" 2>&1; then
		ADAMIC_THREADS=4 "$counted" > /dev/null 2> "$work/counted.err"
		say "count $(basename "$source"): $(grep -o 'adamic: counts:.*' "$work/counted.err" | tail -n 1)"
	else
		say "count $(basename "$source"): the counted build failed"
		tail -n 5 "$work/build-counted.log"
	fi
done

# Every cap prints what keep-forever prints.
for row in "${live[@]}"; do
	IFS='|' read -r name source environment which <<< "$row"
	for cap in $(row_caps "$which"); do
		env $environment "$(binary "$source" "$cap")" > "$work/out/${name//\//_}.$cap" 2>&1 || fail "$name at cap $cap exited $?" "$work/out/${name//\//_}.$cap"
		cmp -s "$work/out/${name//\//_}.keep" "$work/out/${name//\//_}.$cap" || fail "$name prints differently at cap $cap" "$work/out/${name//\//_}.$cap"
	done
done
say "PASS outputs match across caps ($(head -c 60 "$work/out/heap_churn.keep"))"

# Timing: rounds outermost, then modes, programs and caps, so load drifts across every cap alike.
data=$work/timings.txt
: > "$data"
say "load before $(cut -d' ' -f1-3 /proc/loadavg)"
for round in $(seq "$rounds"); do
	for mode in shipped pinned; do
		extra=()
		[ "$mode" = pinned ] && extra=("${pinned[@]}")
		for row in "${live[@]}"; do
			IFS='|' read -r name source environment which <<< "$row"
			for cap in $(row_caps "$which"); do
				env ${extra[@]+"${extra[@]}"} $environment /usr/bin/time -f '%e %U %S %M %R' -o "$work/time.txt" "$(binary "$source" "$cap")" > /dev/null 2>&1 || fail "$name at cap $cap, $mode, exited nonzero" "$work/time.txt"
				echo "$name $cap $mode $round $(tail -n 1 "$work/time.txt")" >> "$data"
			done
		done
	done
done
say "load after $(cut -d' ' -f1-3 /proc/loadavg)"

# Idle: VmRSS every 5 ms while the program runs, keeping the last read, three takes a cap and mode.
idle=$work/idle.txt
: > "$idle"
for take in 1 2 3; do
	for mode in shipped pinned; do
		extra=()
		[ "$mode" = pinned ] && extra=("${pinned[@]}")
		for row in "${live[@]}"; do
			IFS='|' read -r name source environment which <<< "$row"
			case $name in heap_churn | churn_par/*) ;; *) continue ;; esac
			for cap in $(row_caps "$which"); do
				env ${extra[@]+"${extra[@]}"} $environment "$(binary "$source" "$cap")" > /dev/null 2>&1 &
				pid=$!
				last=0
				while :; do
					sample=$(awk '$1 == "VmRSS:" { print $2 }' "/proc/$pid/status" 2> /dev/null)
					[ -n "$sample" ] || break
					last=$sample
					sleep 0.005
				done
				wait "$pid"
				echo "$name $cap $mode $take $last" >> "$idle"
			done
		done
	done
done
say "PASS timed ${rounds} rounds and 3 idle takes"
cp "$data" "$idle" "$work/out/heap_churn.keep" "$TMPDIR/" 2> /dev/null

# The table: per program, cap and mode, the best wall's run (wall, user, sys s; minor faults; max RSS
# MB), and idle MB for each take.
echo "== table: best of $rounds by wall (that run's user, sys, minor faults, max RSS); idle is the last VmRSS, 3 takes"
echo "program     cap  | s: wall  user   sys   faults  maxMB  idleMB          | p: wall  user   sys   faults  maxMB  idleMB"
awk -v order="$(for row in "${live[@]}"; do IFS='|' read -r name source environment which <<< "$row"; for cap in $(row_caps "$which"); do printf '%s:%s ' "$name" "$cap"; done; done)" '
	FILENAME == ARGV[1] {
		key = $1 ":" $2 ":" $3
		if (!(key in wall) || $5 < wall[key]) { wall[key] = $5; user[key] = $6; sys[key] = $7; peak[key] = $8; faults[key] = $9 }
		next
	}
	{ key = $1 ":" $2 ":" $3; idle[key] = ((key in idle) ? idle[key] "/" : "") sprintf("%.1f", $5 / 1024) }
	END {
		count = split(order, keys, " ")
		for (index_ = 1; index_ <= count; index_++) {
			split(keys[index_], parts, ":")
			line = sprintf("%-11s %-4s", parts[1], parts[2] == "keep" ? "keep" : parts[2] "M")
			for (mode = 1; mode <= 2; mode++) {
				name = mode == 1 ? "shipped" : "pinned"
				key = parts[1] ":" parts[2] ":" name
				line = line sprintf(" | %s: %5.2f %5.2f %5.2f %7d %6.1f  %-15s", substr(name, 1, 1), wall[key], user[key], sys[key], faults[key], peak[key] / 1024, (key in idle) ? idle[key] : "-")
			}
			print line
		}
	}' "$data" "$idle"
finish 0

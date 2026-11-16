#!/bin/bash
# #re2xhwa's M1: proves bench -compare can say no change when nothing changed and slower when
# something did. Run from the repository root on a Linux box with Go, clang and node. In order:
# 1. go test ./bench/.
# 2. A/A: main against main on everyday, 15 samples. The sign test's interval misses with
#    probability 2 x 576/32768, about 3.5%, per program, so about one A/A run in four shows one false
#    verdict among eight; the run passes with at most one (see below).
# 3. The plant: main against main with -plant candidate, on devirtualize (the shortest everyday
#    program) and trees (about a second). Each must read slower.
# Each full log is kept under the printed directory. Never lands.
set -u
failed=0
MAIN=$(git rev-parse HEAD)
T=$(mktemp -d)
echo "box $(hostname) $(nproc) cpus, $(uname -r), main $MAIN, logs in $T"
echo "load before $(cut -d' ' -f1-3 /proc/loadavg); busy processes: $(ps -eo pcpu=,comm= | awk '$1 > 5 {printf "%s(%s) ", $2, $1}')"

# verdicts prints each program's row as "name<TAB>verdict" from a comparison's table.
verdicts() {
	awk -F'|' 'NF >= 8 && $2 !~ /program|---/ { gsub(/^ +| +$/, "", $2); gsub(/^ +| +$/, "", $6); print $2 "\t" $6 }' "$1"
}

echo "== go test ./bench/"
if go test -count=1 ./bench/ > "$T/test.log" 2>&1; then
	echo "PASS go test ./bench/"
else
	echo "FAIL go test ./bench/"
	failed=1
	tail -40 "$T/test.log"
fi

echo "== A/A: $MAIN against itself, everyday, 15 samples; every program must read no change"
go run ./bench -compare "$MAIN" "$MAIN" -suite everyday -samples 15 > "$T/aa.log" 2> "$T/aa.err"
echo "exit $?"
cat "$T/aa.log"
echo "-- its stderr, without the per-sample lines:"
grep -v '^sample ' "$T/aa.err" | tail -30
EVERYDAY=$(node -e 'console.log(require("./bench/suites.json").suites.everyday.programs.map((program) => program.name).join(" "))')
# A 95% test is wrong 5% of the time by design, so the A/A is judged on its whole run: each
# program's verdict is printed, and the run passes when at most one of the eight reads a verdict.
# Two or more false verdicts among eight at about 3.5% each happens about 3% of the time.
false_verdicts=0
for name in $EVERYDAY; do
	verdict=$(verdicts "$T/aa.log" | awk -F'\t' -v name="$name" '$1 == name { print $2 }')
	if [ "$verdict" = "no change at this confidence" ]; then
		echo "A/A $name: $verdict"
	elif [ -z "$verdict" ]; then
		echo "FAIL A/A $name: no row"
		false_verdicts=$((false_verdicts + 2))
	else
		echo "A/A $name: $verdict (a false verdict)"
		false_verdicts=$((false_verdicts + 1))
	fi
done
if [ "$false_verdicts" -le 1 ]; then
	echo "PASS A/A: $false_verdicts false verdicts among the eight, at most 1 allowed"
else
	failed=1
	echo "FAIL A/A: $false_verdicts false verdicts among the eight (a missing row counts 2), at most 1 allowed"
fi

echo "== plant: $MAIN against itself with -plant candidate, devirtualize and trees, 15 samples; each must read slower"
go run ./bench -compare "$MAIN" "$MAIN" -suite everyday -only devirtualize,trees -samples 15 -plant candidate > "$T/plant.log" 2> "$T/plant.err"
echo "exit $?"
cat "$T/plant.log"
echo "-- its stderr, without the per-sample lines:"
grep -v '^sample ' "$T/plant.err" | tail -30
for name in devirtualize trees; do
	verdict=$(verdicts "$T/plant.log" | awk -F'\t' -v name="$name" '$1 == name { print $2 }')
	if [ "$verdict" = "slower" ]; then
		echo "PASS plant $name: $verdict"
	else
		echo "FAIL plant $name: ${verdict:-no row}"
		failed=1
	fi
done
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
# A measure that fails its own proof ends red.
exit $failed

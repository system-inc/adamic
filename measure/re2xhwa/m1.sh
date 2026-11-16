#!/bin/bash
# #re2xhwa's M1: proves bench -compare can say no change when nothing changed and slower when
# something did. Run from the repository root on a Linux box with Go, clang and node. In order:
# 1. go test ./bench/.
# 2. A/A: main against main on everyday, 15 samples. The sign test's interval misses with
#    probability 2 x 576/32768, about 3.5%, per program, so about one A/A run in four shows one false
#    verdict among eight; the run passes with at most one (see below).
# 3. Each program's sensitivity from the A/A: the change it couldn't rule out, its interval's farther
#    bound, and that in milliseconds of its baseline median.
# 4. The plant, sized to the claim (Runtime on #re2xhwa): devirtualize, the shortest everyday program,
#    with a fixed 25 ms; trees, about a second, with three times the sensitivity the A/A measured for
#    it on this box. Each must read slower.
# Each full log is kept under the printed directory. Never lands.
set -u
failed=0
MAIN=$(git rev-parse HEAD)
T=$(mktemp -d)
echo "box $(hostname) $(nproc) cpus, $(uname -r), main $MAIN, logs in $T"
echo "load before $(cut -d' ' -f1-3 /proc/loadavg); busy processes: $(ps -eo pcpu=,comm= | awk '$1 > 5 {printf "%s(%s) ", $2, $1}')"

# rows prints each program's row as "name<TAB>verdict<TAB>baseline seconds<TAB>low<TAB>high" from a
# comparison's table: | program | judge | baseline | ratio | low to high | verdict | node |.
rows() {
	awk -F'|' 'NF >= 9 && $2 !~ /program|---/ {
		for (i = 2; i <= 8; i++) gsub(/^ +| +$/, "", $i)
		split($6, interval, " to ")
		baseline = $4; sub(/ s$/, "", baseline)
		print $2 "\t" $7 "\t" baseline "\t" interval[1] "\t" interval[2]
	}' "$1"
}
field() { rows "$1" | awk -F'\t' -v name="$2" -v column="$3" '$1 == name { print $column }'; }

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
	verdict=$(field "$T/aa.log" "$name" 2)
	if [ "${verdict#no change at this confidence}" != "$verdict" ]; then
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

echo "== sensitivity at 15 samples on this box, from the A/A: the change each program couldn't rule out"
for name in $EVERYDAY; do
	baseline=$(field "$T/aa.log" "$name" 3); low=$(field "$T/aa.log" "$name" 4); high=$(field "$T/aa.log" "$name" 5)
	if [ -z "$low" ]; then echo "$name: no row"; continue; fi
	awk -v n="$name" -v b="$baseline" -v l="$low" -v h="$high" 'BEGIN { w = (1 - l > h - 1) ? 1 - l : h - 1; printf "%s: %.1f%% of a %.3f s baseline, %.0f ms\n", n, 100 * w, b, 1000 * w * b }'
done
trees_ms=$(awk -v b="$(field "$T/aa.log" trees 3)" -v l="$(field "$T/aa.log" trees 4)" -v h="$(field "$T/aa.log" trees 5)" 'BEGIN { if (b == "") { print ""; exit } w = (1 - l > h - 1) ? 1 - l : h - 1; ms = int(3 * w * b * 1000 + 0.999); if (ms < 25) ms = 25; print ms }')
if [ -z "$trees_ms" ]; then
	echo "FAIL the A/A gave trees no row, so its plant can't be sized"
	failed=1
	trees_ms=100
fi

plant() { # name milliseconds
	go run ./bench -compare "$MAIN" "$MAIN" -suite everyday -only "$1" -samples 15 -plant candidate -plant-ms "$2" > "$T/plant.$1.log" 2> "$T/plant.$1.err"
	echo "exit $?"
	cat "$T/plant.$1.log"
	echo "-- its stderr, without the per-sample lines:"
	grep -v '^sample ' "$T/plant.$1.err" | tail -30
}
echo "== plant: devirtualize with 25 ms, 15 samples; must read slower"
plant devirtualize 25
echo "== plant: trees with $trees_ms ms, three times its A/A sensitivity, 15 samples; must read slower"
plant trees "$trees_ms"
for name in devirtualize trees; do
	verdict=$(field "$T/plant.$name.log" "$name" 2)
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

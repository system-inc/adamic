#!/bin/bash
# #1hp4ww2's before and after: main fd49dbfc against the cached for-in key list (the head's tree). fasta
# at 2.5M, steady.a (fasta's shape, a record whose keys never change) and churn.a (a key added before every
# loop, so the cache never helps), each with both commits' binaries, output against Node first, 5
# interleaved rounds, medians of wall, user and peak. Never lands.
set -u
BEFORE=fd49dbfc95$(git rev-parse fd49dbfc95 | cut -c11-)
AFTER=$(git rev-parse HEAD)
T=$(mktemp -d)
echo "box $(hostname) $(nproc) cpus, $(uname -r), before $BEFORE, after $AFTER"
echo "load before $(cut -d' ' -f1-3 /proc/loadavg); busy: $(ps -eo pcpu=,comm= | awk '$1 > 5 {printf "%s(%s) ", $2, $1}')"
for side in before after; do
	commit=$BEFORE; [ $side = after ] && commit=$AFTER
	git worktree add -q --detach "$T/$side-tree" "$commit" || { echo "FAIL worktree $side"; exit 1; }
	rmdir "$T/$side-tree/cohere" 2>/dev/null; ln -s "$PWD/cohere" "$T/$side-tree/cohere"
	(cd "$T/$side-tree" && go build -buildvcs=false -o "$T/adamic-$side" ./cmd/adamic) || { echo "FAIL adamic build $side"; exit 1; }
done
names=(fasta steady churn)
sources=(bench/holdout/fasta.a measure/1hp4ww2/steady.a measure/1hp4ww2/churn.a)
arguments=(2500000 "" "")
for index in 0 1 2; do
	name=${names[$index]}; source=${sources[$index]}; argument=${arguments[$index]}
	want=$(node --disable-warning=ExperimentalWarning oracle/node.mjs "$source" $argument | sha256sum | cut -c1-16)
	for side in before after; do
		"$T/adamic-$side" build "$source" -o "$T/$name-$side" > "$T/b.log" 2>&1 || { echo "FAIL build $name $side: $(head -c 300 "$T/b.log")"; exit 1; }
		got=$("$T/$name-$side" $argument | sha256sum | cut -c1-16)
		[ "$got" = "$want" ] && echo "PASS $name $side prints Node's output ($got)" || { echo "FAIL $name $side output $got, Node $want"; exit 1; }
	done
	for round in 1 2 3 4 5; do
		for side in before after; do
			/usr/bin/time -f "%e %U %M" -o "$T/t.$name.$side.$round" "$T/$name-$side" $argument > /dev/null
		done
	done
	python3 -I - "$T" $name <<'PY'
import glob, statistics, sys
root, name = sys.argv[1], sys.argv[2]
row = {}
for side in ['before', 'after']:
    times = [list(map(float, open(p).read().split()[-3:])) for p in glob.glob(f'{root}/t.{name}.{side}.*')]
    row[side] = [statistics.median(c) for c in zip(*times)]
b, a = row['before'], row['after']
print(f"{name:7} medians of 5, wall/user/peak KB: before {b[0]:.2f}/{b[1]:.2f}/{b[2]:.0f}  after {a[0]:.2f}/{a[1]:.2f}/{a[2]:.0f}  after/before user {a[1]/b[1]:.3f}")
PY
done
for side in before after; do rm "$T/$side-tree/cohere" && mkdir "$T/$side-tree/cohere" && git worktree remove "$T/$side-tree"; done
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"

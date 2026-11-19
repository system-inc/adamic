#!/bin/bash
# #2ccxtyr's before and after: main 4bd46314 against the inline adamic_array_set head 6041cc89.
# 1. fannkuchredux at n = 10 and 11, each commit's binary, 5 interleaved rounds, output against Node first;
# 2. the sanitized build time (--sanitize, ASan) of fannkuchredux and bench's regression programs at both
#    commits, best of 3, since the write is out of line under ASan and must cost no clang time there;
# 3. bench -compare over regression, 15 samples (trees, sort and the rest beside it).
# Never lands.
set -u
BEFORE=4bd46314537bc2dbd36fc48b5e69e6d144b2b714
AFTER=6041cc89e6bf06e59aeed1a4df30e29d7630d2d7
T=$(mktemp -d)
echo "box $(hostname) $(nproc) cpus, $(uname -r), tree $(git rev-parse HEAD), before $BEFORE, after $AFTER"
echo "load before $(cut -d' ' -f1-3 /proc/loadavg); busy: $(ps -eo pcpu=,comm= | awk '$1 > 5 {printf "%s(%s) ", $2, $1}')"
for side in before after; do
	commit=$BEFORE; [ $side = after ] && commit=$AFTER
	git worktree add -q --detach "$T/$side-tree" "$commit" || { echo "FAIL worktree $side"; exit 1; }
	rmdir "$T/$side-tree/cohere" 2>/dev/null; ln -s "$PWD/cohere" "$T/$side-tree/cohere"
	(cd "$T/$side-tree" && go build -buildvcs=false -o "$T/adamic-$side" ./cmd/adamic) || { echo "FAIL adamic build $side"; exit 1; }
	echo "PASS adamic built at $side $commit"
done
grep -q '^mainThread(12);$' bench/holdout/fannkuchredux.a || { echo "FAIL the port no longer calls mainThread(12)"; exit 1; }
for n in 10 11; do sed "s/^mainThread(12);\$/mainThread($n);/" bench/holdout/fannkuchredux.a > "bench/holdout/fannkuch_n$n.a"; done
echo "== 1. fannkuchredux, 5 interleaved rounds, wall and user"
for n in 10 11; do
	want=$(node --disable-warning=ExperimentalWarning oracle/node.mjs "bench/holdout/fannkuch_n$n.a")
	for side in before after; do
		"$T/adamic-$side" build "bench/holdout/fannkuch_n$n.a" -o "$T/fk$n-$side" > "$T/b.log" 2>&1 || { echo "FAIL build n=$n $side: $(head -c 300 "$T/b.log")"; exit 1; }
		[ "$("$T/fk$n-$side")" = "$want" ] && echo "PASS n=$n $side prints Node's output" || { echo "FAIL n=$n $side output"; exit 1; }
	done
	for round in 1 2 3 4 5; do
		for side in before after node; do
			if [ $side = node ]; then command=(node --disable-warning=ExperimentalWarning oracle/node.mjs "bench/holdout/fannkuch_n$n.a"); else command=("$T/fk$n-$side"); fi
			/usr/bin/time -f "%e %U" -o "$T/t.$n.$side.$round" "${command[@]}" > /dev/null
		done
	done
	python3 -I - "$T" $n <<'PY'
import glob, statistics, sys
root, n = sys.argv[1], sys.argv[2]
row = {}
for side in ['before', 'after', 'node']:
    times = [list(map(float, open(p).read().split()[-2:])) for p in glob.glob(f'{root}/t.{n}.{side}.*')]
    row[side] = [statistics.median(c) for c in zip(*times)]
print(f"n={n} medians of 5, wall/user: before {row['before'][0]:.2f}/{row['before'][1]:.2f}  after {row['after'][0]:.2f}/{row['after'][1]:.2f}  node {row['node'][0]:.2f}/{row['node'][1]:.2f}  after/before user {row['after'][1]/row['before'][1]:.3f}  after/node user {row['after'][1]/row['node'][1]:.2f}")
PY
done
echo "== 2. sanitized build time, seconds, best of 3"
for source in bench/holdout/fannkuch_n10.a bench/trees.ts bench/sort.ts bench/nbody.ts bench/tokenizer.ts bench/word_count.ts bench/devirtualize.a bench/graph.ts bench/copy.ts bench/parallel_files.a; do
	line="$source"
	for side in before after; do
		best=
		for round in 1 2 3; do
			start=$(date +%s.%N)
			"$T/adamic-$side" build "$source" --sanitize -o "$T/s-$side" > "$T/s.log" 2>&1 || { line="$line $side FAIL($(head -c 120 "$T/s.log"))"; best=x; break; }
			took=$(echo "$(date +%s.%N) - $start" | bc)
			[ -z "$best" ] || [ "$(echo "$took < $best" | bc)" = 1 ] && best=$took
		done
		line="$line $side ${best}"
	done
	echo "$line"
done
echo "== 3. bench -compare over regression"
go run ./bench -compare "$BEFORE" "$AFTER" -suite regression -samples 15 2> "$T/compare.err"
status=$?
grep -v '^sample ' "$T/compare.err" | tail -20
for side in before after; do rm "$T/$side-tree/cohere" && mkdir "$T/$side-tree/cohere" && git worktree remove "$T/$side-tree"; done
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
exit $status

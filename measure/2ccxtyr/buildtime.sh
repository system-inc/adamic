#!/bin/bash
# #2ccxtyr: what the inline adamic_array_set costs the compile. Release and --sanitize build times, at main
# 4bd46314 and at the head 6041cc89, of stage1/cohere/json/main.ts (the large unit ff1f4068 measured),
# bench/trees.ts and fannkuchredux at n = 10. Best of 3. Every timed build gets a fresh
# ADAMIC_BUILD_CACHE_DIR seeded with that side's runtime archives from one warm-up build, so it times
# compiling the program itself, never a cache hit or an archive build. Never lands.
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
	printf 'console.log("0");\n' > "$T/warm.ts"
	for mode in release sanitize; do
		flag=; [ $mode = sanitize ] && flag=--sanitize
		mkdir -p "$T/warm-$side-$mode"
		ADAMIC_BUILD_CACHE_DIR="$T/warm-$side-$mode" "$T/adamic-$side" build "$T/warm.ts" -o "$T/warm-$side-$mode.out" $flag > "$T/w.log" 2>&1 || { echo "FAIL warm-up $side $mode: $(head -c 300 "$T/w.log")"; exit 1; }
	done
	echo "PASS adamic and its runtime archives at $side $commit"
done
sed 's/^mainThread(12);$/mainThread(10);/' bench/holdout/fannkuchredux.a > bench/holdout/fannkuch_n10.a
for source in stage1/cohere/json/main.ts bench/trees.ts bench/holdout/fannkuch_n10.a; do
	for mode in release sanitize; do
		flag=; [ $mode = sanitize ] && flag=--sanitize
		line="$source $mode:"
		for side in before after; do
			best=
			for round in 1 2 3; do
				cache="$T/cache-$side-$mode-$round-$(basename "$source")"
				cp -R "$T/warm-$side-$mode" "$cache"
				start=$(date +%s.%N)
				ADAMIC_BUILD_CACHE_DIR="$cache" "$T/adamic-$side" build "$source" -o "$T/out-$side" $flag > "$T/b.log" 2>&1 || { best="FAIL($(head -c 200 "$T/b.log"))"; break; }
				took=$(echo "$(date +%s.%N) - $start" | bc)
				if [ -z "$best" ] || [ "$(echo "$took < $best" | bc)" = 1 ]; then best=$took; fi
			done
			line="$line $side $best"
		done
		echo "$line"
	done
done
for side in before after; do rm "$T/$side-tree/cohere" && mkdir "$T/$side-tree/cohere" && git worktree remove "$T/$side-tree"; done
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"

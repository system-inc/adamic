#!/bin/bash
# two-root.sh <worktree> <label>: builds one fixture sanitized under /opt/llvm-a and /opt/llvm-c-longer-name, each with a
# fresh empty build cache, then compares the binaries and counts /opt/llvm in each.
set -u
worktree=$1 label=$2
out=/tmp/two-root/$label
mkdir -p "$out"
fixture=$worktree/internal/oracle/testdata/arguments.a
for root in llvm-a llvm-c-longer-name; do
	cache=$out/cache-$root
	mkdir -p "$cache"
	echo "== $root: $(PATH=/opt/$root/bin:$PATH command -v clang), cache $cache ($(ls -A "$cache" | wc -l) entries before)"
	PATH=/opt/$root/bin:/opt/node24/bin:$PATH TMPDIR=/tmp/adamic-gate ADAMIC_BUILD_CACHE_DIR=$cache \
		"$worktree/adamic" build "$fixture" -o "$out/arguments-$root" --sanitize
	echo "build exit=$?"
done
cmp "$out/arguments-llvm-a" "$out/arguments-llvm-c-longer-name"; echo "cmp exit=$?"
for root in llvm-a llvm-c-longer-name; do
	echo "strings -a arguments-$root | grep -c /opt/llvm: $(strings -a "$out/arguments-$root" | grep -c /opt/llvm)"
	strings -a "$out/arguments-$root" | grep /opt/llvm | sort | uniq -c | sort -rn | head -5
done
sha256sum "$out"/arguments-*
"$out/arguments-llvm-a" > /dev/null; echo "run exit=$?"

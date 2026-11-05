#!/bin/bash
# usage: stress.sh <name> <worktree> <binary> <parallel>
S=$(cd "$(dirname "$0")" && pwd); name=$1; w=$2; bin=$3; par=$4
cd "$w"; d=stage1/cohere/lint/regex/testdata/shapes/plain
{ for i in $(seq 2000); do echo $d/040_runtime.a; echo $d/081_runtime.a; done
  for i in $(seq 20); do ls $d/*_runtime.a; done; } | shuf > "$S/jobs-$name.txt"
# load: 4 busy loops
for i in 1 2 3 4; do (timeout 900 sh -c 'while :; do :; done' &) ; done
xargs -P "$par" -I{} "$S/one.sh" "$w" "$bin" {} < "$S/jobs-$name.txt" > "$S/res-$name.txt"
pkill -f 'while :; do :; done'
sort "$S/res-$name.txt" | uniq -c > "$S/summary-$name.txt"

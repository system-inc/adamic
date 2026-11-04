#!/bin/bash
# 5 runs per binary; prints median seconds. visits = 1024 * 40000
for R in 0 1 4 8; do for m in view plain; do
  ts=()
  for i in 1 2 3 4 5; do s=$(date +%s.%N); out=$(./bin_${m}_$R); e=$(date +%s.%N); ts+=($(echo "$e - $s" | bc)); done
  med=$(printf '%s\n' "${ts[@]}" | sort -n | sed -n 3p)
  echo "$m R=$R median=${med}s runs=${ts[*]} out=$out"
done; done

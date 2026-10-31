#!/bin/bash
# usage: strace-run.sh <main|change> <round>; runs the prebuilt lower.test in that sha's tree under strace
v=$1; r=$2
if [ "$v" = main ]; then dir=/tmp/ev-main/internal/lower; else dir=/home/user/adamic/internal/lower; fi
source /opt/adamic-tools/env.sh
cd "$dir" || exit 9
out=/tmp/ev/11-strace-$v-$r
s=$(date +%s.%N)
strace -f -c -e trace=newfstatat,statx,stat,lstat,fstatat64 -o $out.strace.txt \
  /tmp/ev-bin/$v/lower.test -test.run '^TestCycleFinderEquality_00[0-9]$' -test.count=1 -test.v > $out.test.log 2>&1
code=$?
e=$(date +%s.%N)
echo "sha-tree=$dir binary=/tmp/ev-bin/$v/lower.test exit=$code wall=$(echo "$e-$s" | bc)s" > $out.summary.txt

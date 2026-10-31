#!/bin/bash
# sweep.sh <label> <binary> <outdir>: one run of TestProduct_HelperOracle under strace -c, then one under execve tracing.
set -u
label=$1 binary=$2 out=$3
cd /home/user/adamic/stage1/cohere/lint/helpers/oracles || exit 2
source /opt/adamic-tools/env.sh
start=$(date +%s.%N)
strace -f -c -e trace=newfstatat,lstat -o "$out/$label.strace-c.txt" "$binary" -test.run '^TestProduct_HelperOracle$' -test.count=1 -test.v > "$out/$label.test.txt" 2>&1
code=$?
end=$(date +%s.%N)
echo "label=$label exit=$code wall=$(echo "$end - $start" | bc)s" | tee "$out/$label.summary.txt"
strace -f -qq -e trace=execve -e signal=none -o "$out/$label.execve.txt" "$binary" -test.run '^TestProduct_HelperOracle$' -test.count=1 -test.v > "$out/$label.test-execve.txt" 2>&1
echo "execve-run exit=$?" >> "$out/$label.summary.txt"
echo "git execs: $(grep -c 'execve("[^"]*/git"' "$out/$label.execve.txt")" >> "$out/$label.summary.txt"
grep -o 'execve("[^"]*/git", \[[^]]*\]' "$out/$label.execve.txt" | sed 's/execve("[^"]*", //' | sort | uniq -c | sort -rn > "$out/$label.git-commands.txt"
cat "$out/$label.summary.txt"

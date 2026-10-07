#!/bin/bash
set -eu
export LD_LIBRARY_PATH=/workspace/lint-profile-tools/usr/lib/x86_64-linux-gnu
perf=/workspace/lint-profile-tools/usr/bin/perf
root=/workspace/lint-cost-baseline
cat /proc/loadavg > "$root/profile-load-before.txt"
"$perf" record -e task-clock:u -F 1999 --call-graph dwarf,16384 -o "$root/perf.data" -- bash -c 'for round in $(seq 1 20); do /workspace/lint-cost-baseline/scanner --manifest /workspace/lint-cost-baseline/compiler.txt --count; done' > "$root/profile.stdout" 2> "$root/profile.stderr"
cat /proc/loadavg > "$root/profile-load-after.txt"
"$perf" script -i "$root/perf.data" --inline -F comm,pid,time,event,ip,sym,dso > "$root/perf-script.txt" 2> "$root/perf-script.stderr"
"$perf" report -i "$root/perf.data" --stdio --no-children --percent-limit 0 --sort symbol,dso --call-graph none > "$root/perf-self.txt" 2> "$root/perf-report.stderr"

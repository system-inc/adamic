#!/bin/bash
set -eu
directory=$1
export LD_LIBRARY_PATH=/workspace/lint-profile-tools/usr/lib/x86_64-linux-gnu
/workspace/lint-profile-tools/usr/bin/perf record -F 999 --call-graph dwarf,16384 -o "$directory/perf.data" -- "$directory/profiled" --manifest "$directory/compiler.txt" --count > "$directory/perf.stdout" 2> "$directory/perf.stderr"
/workspace/lint-profile-tools/usr/bin/perf report -i "$directory/perf.data" --stdio --no-children --percent-limit 0 --sort symbol,dso --call-graph none > "$directory/perf-self.txt" 2> "$directory/perf-report.stderr"
/workspace/lint-profile-tools/usr/bin/perf report -i "$directory/perf.data" --stdio --children --percent-limit 0 --sort symbol,dso --call-graph graph,0.5 > "$directory/perf-graphs.txt" 2>> "$directory/perf-report.stderr"

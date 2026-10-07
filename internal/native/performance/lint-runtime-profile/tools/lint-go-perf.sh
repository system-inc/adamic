#!/bin/bash
set -eu
directory=$1
export LD_LIBRARY_PATH=/workspace/lint-profile-tools/usr/lib/x86_64-linux-gnu
/workspace/lint-profile-tools/usr/bin/perf record -F 999 --call-graph dwarf,16384 -o "$directory/go-perf.data" -- /bin/bash -c 'for trial in {1..10}; do "$1" --manifest "$2" --count; done' lint-go "$directory/oracle" "$directory/compiler.txt" > "$directory/go-perf.stdout" 2> "$directory/go-perf.stderr"
/workspace/lint-profile-tools/usr/bin/perf report -i "$directory/go-perf.data" --stdio --no-children --percent-limit 0 --sort symbol,dso --call-graph none > "$directory/go-perf-self.txt" 2> "$directory/go-perf-report.stderr"
/workspace/lint-profile-tools/usr/bin/perf report -i "$directory/go-perf.data" --stdio --children --percent-limit 0 --sort symbol,dso --call-graph graph,0.5 > "$directory/go-perf-graphs.txt" 2>> "$directory/go-perf-report.stderr"

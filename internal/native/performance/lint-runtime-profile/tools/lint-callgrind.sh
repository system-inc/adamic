#!/bin/bash
set -eu
directory=$1
if [[ "$directory" == *8 ]]; then
  printf '/workspace/lint-typescript/src/compiler/core.ts\tall\n' > "$directory/small.txt"
else
  printf '/workspace/lint-typescript/src/compiler/core.ts\n' > "$directory/small.txt"
fi
export VALGRIND_LIB=/workspace/lint-profile-tools/usr/libexec/valgrind
/workspace/lint-profile-tools/usr/bin/valgrind --tool=callgrind --callgrind-out-file="$directory/callgrind.out" "$directory/profiled" --manifest "$directory/small.txt" --count > "$directory/callgrind.stdout" 2> "$directory/callgrind.stderr"
/workspace/lint-profile-tools/usr/bin/callgrind_annotate --inclusive=no --threshold=99.9 "$directory/callgrind.out" > "$directory/callgrind-self.txt" 2> "$directory/callgrind-annotate.stderr"
/workspace/lint-profile-tools/usr/bin/callgrind_annotate --inclusive=yes --threshold=99.9 "$directory/callgrind.out" > "$directory/callgrind-inclusive.txt" 2>> "$directory/callgrind-annotate.stderr"
"$directory/counted" --manifest "$directory/small.txt" --count > "$directory/small-counted.stdout" 2> "$directory/small-counted.stderr"

#!/bin/sh
# probe.sh <file.a> [arguments...]: runs a probe three ways and prints each one's output and exit code:
# its source on Node (the truth), natively under ASan and UBSan with LeakSanitizer on, and through the
# JavaScript backend.
# It builds the native binary the way native.Build does under Options{Sanitize: true}.
set -u
repository=$(cd "$(dirname "$0")/../.." && pwd)
file=$1
shift
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
(cd "$repository" && go run ./cmd/adamic c "$file") > "$work/program.c" || exit 1
(cd "$repository" && go run ./cmd/adamic js "$file") > "$work/program.mjs" || exit 1
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable \
	-Wno-unused-function -ffp-contract=off -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all \
	-I "$repository/internal/native/runtime" -o "$work/program" "$work/program.c" \
	"$repository"/internal/native/runtime/*.c -lm || exit 1
echo "== node"
(cd "$repository" && node --disable-warning=ExperimentalWarning oracle/node.mjs "$file" "$@"); echo "exit=$?"
echo "== native"
ASAN_OPTIONS=detect_leaks=1 "$work/program" "$@"; echo "exit=$?"
echo "== javascript backend"
node --disable-warning=ExperimentalWarning "$repository/oracle/node.mjs" "$work/program.mjs" "$@"; echo "exit=$?"

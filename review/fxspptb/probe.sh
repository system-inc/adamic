#!/bin/sh
# probe.sh <file.a> [arguments...]: runs a probe three ways and prints each one's output and exit code:
# its source on Node (the truth), natively under ASan and UBSan with LeakSanitizer on, and through the
# JavaScript backend.
# The sanitized build copies native.Build's flags under Options{Sanitize: true} (keep them in step: a
# flag missing here once made a fixed bug look unfixed); the fourth run is adamic build's own binary.
set -u
repository=$(cd "$(dirname "$0")/../.." && pwd)
file=$1
shift
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
(cd "$repository" && go run ./cmd/adamic c "$file") > "$work/program.c" || exit 1
(cd "$repository" && go run ./cmd/adamic js "$file") > "$work/program.mjs" || exit 1
clang -std=c11 -Wall -Wextra -Werror -pedantic -Wno-unused-variable -Wno-unused-but-set-variable \
	-Wno-unused-function -Wno-unused-parameter -Wno-self-assign -ffp-contract=off -fno-optimize-sibling-calls -O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all \
	-I "$repository/internal/native/runtime" -o "$work/program" "$work/program.c" \
	"$repository"/internal/native/runtime/*.c -lm || exit 1
echo "== node"
(cd "$repository" && node --disable-warning=ExperimentalWarning oracle/node.mjs "$file" "$@"); echo "exit=$?"
echo "== native"
ASAN_OPTIONS=detect_leaks=1 "$work/program" "$@"; echo "exit=$?"
echo "== native, as adamic build makes it (-O2, no sanitizers)"
(cd "$repository" && go run ./cmd/adamic build "$file" -o "$work/built") && "$work/built" "$@"; echo "exit=$?"
echo "== javascript backend"
node --disable-warning=ExperimentalWarning "$repository/oracle/node.mjs" "$work/program.mjs" "$@"; echo "exit=$?"

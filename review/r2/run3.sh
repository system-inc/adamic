#!/bin/sh
# run3.sh <file.a> [arguments...]: build a probe three ways (its source on Node, native under ASan and
# UBSan, the JavaScript backend on Node) and print what each wrote and how each exited. Run from the
# repository's root. With PIPE set, stdout and stderr go to one pipe (2>&1 | cat), as a reader of
# both would see them.
set -u
probe=$1
shift
work=${TMPDIR:-/tmp}/run3.$$
mkdir -p "$work"
go run ./review/r2/sanitized "$probe" "$work/native" || exit 1
go run ./cmd/adamic js "$probe" > "$work/backend.mjs" || exit 1
show() {
	name=$1
	shift
	if [ -n "${PIPE:-}" ]; then
		( cd "$work" && timeout 60 env ASAN_OPTIONS=detect_leaks=1 "$@" 2>&1; echo "exit=$?" ) | cat > "$work/$name.out"
	else
		( cd "$work" && timeout 60 env ASAN_OPTIONS=detect_leaks=1 "$@" > "$work/$name.stdout" 2> "$work/$name.stderr"; echo "exit=$?" > "$work/$name.out" )
		cat "$work/$name.stdout" "$work/$name.stderr" >> "$work/$name.out"
	fi
	echo "== $name"
	cat "$work/$name.out"
}
root=$(pwd)
show node node --disable-warning=ExperimentalWarning "$root/oracle/node.mjs" "$root/$probe" "$@"
show native "$work/native" "$@"
show backend node --disable-warning=ExperimentalWarning "$root/oracle/node.mjs" "$work/backend.mjs" "$@"

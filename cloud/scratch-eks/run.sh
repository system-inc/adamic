#!/usr/bin/env bash
# run.sh: scratch probes for #eksbz4c, library objects own tags and groups, on every backend.
# Not for main.
set -u
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$root" || exit 2
work=$(mktemp -d "${TMPDIR:-/tmp}/roads.XXXXXX")
export GOTMPDIR=${GOTMPDIR:-$work/gotmp}
mkdir -p "$GOTMPDIR"
if ! go build -o "$work/adamic" ./cmd/adamic >"$work/build.log" 2>&1; then
	echo "compiler build failed"; tail -40 "$work/build.log"; exit 2
fi
printf "import { registerHooks } from 'node:module';\nregisterHooks({ resolve(specifier, context, next) { return specifier === 'adamic' ? { url: 'file://%s/oracle/adamic.mjs', shortCircuit: true } : next(specifier, context); } });\n" "$root" >"$work/hook.mjs"
show() { tail -c 1500 "$1" | sed 's/^/    /'; }
for file in cloud/scratch-eks/*.ts; do
	name=$(basename "$file" .ts)
	echo
	echo "=== $name"
	timeout 60 node --disable-warning=ExperimentalWarning oracle/node.mjs "$file" >"$work/$name.node" 2>&1; echo "--- node [exit $?]"; show "$work/$name.node"
	if "$work/adamic" js "$file" >"$work/$name.mjs" 2>"$work/$name.js.err"; then
		timeout 60 node --disable-warning=ExperimentalWarning --import "$work/hook.mjs" "$work/$name.mjs" >"$work/$name.js" 2>&1; echo "--- js [exit $?]"; show "$work/$name.js"
	else
		echo "--- js stops"; show "$work/$name.js.err"
	fi
	if "$work/adamic" build "$file" -o "$work/$name.san" --sanitize >"$work/$name.san.err" 2>&1; then
		ASAN_OPTIONS=detect_leaks=1 timeout 60 "$work/$name.san" >"$work/$name.san.out" 2>&1; echo "--- native sanitized [exit $?]"; show "$work/$name.san.out"
	else
		echo "--- native stops"; show "$work/$name.san.err"
	fi
	if "$work/adamic" build "$file" -o "$work/$name.count" --count >"$work/$name.count.err" 2>&1; then
		timeout 60 "$work/$name.count" >"$work/$name.count.out" 2>&1; echo "--- native counted [exit $?]"; show "$work/$name.count.out"
	fi
	if ADAMIC_PROGRAM_REGION=1 "$work/adamic" build "$file" -o "$work/$name.program" --sanitize >"$work/$name.program.err" 2>&1; then
		ADAMIC_PROGRAM_REGION=1 timeout 60 "$work/$name.program" >"$work/$name.program.out" 2>&1; echo "--- native Program regions [exit $?]"; show "$work/$name.program.out"
	else
		echo "--- native Program regions stops"; show "$work/$name.program.err"
	fi
	if "$work/adamic" build --target wasm32-wasi "$file" -o "$work/$name.wasm" >"$work/$name.wasm.err" 2>&1; then
		timeout 60 node --disable-warning=ExperimentalWarning oracle/wasi.mjs "$work/$name.wasm" >"$work/$name.wasm.out" 2>&1; echo "--- wasm32 [exit $?]"; show "$work/$name.wasm.out"
	else
		echo "--- wasm32 stops"; show "$work/$name.wasm.err"
	fi
done

#!/usr/bin/env bash
# run.sh: scratch probe for #hcgy751, WeakMap and WeakSet on Node, native, wasm32 and the JavaScript backend.
set -u
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$root" || exit 2
work=$(mktemp -d "${TMPDIR:-/tmp}/wk.XXXXXX")
mkdir -p "${GOTMPDIR:-$work/gotmp}" 2>/dev/null || true
export GOTMPDIR=${GOTMPDIR:-$work/gotmp}
echo "head $(git rev-parse HEAD 2>/dev/null || echo unknown), $(node --version)"
if ! go build -o "$work/adamic" ./cmd/adamic >"$work/build.log" 2>&1; then
	echo "compiler build failed"; tail -40 "$work/build.log"; exit 2
fi
printf "import { registerHooks } from 'node:module';\nregisterHooks({ resolve(specifier, context, next) { return specifier === 'adamic' ? { url: 'file://%s/oracle/adamic.mjs', shortCircuit: true } : next(specifier, context); } });\n" "$root" >"$work/hook.mjs"
for file in "${1:-cloud/scratch-wk}"/*.ts; do
	name=$(basename "$file" .ts)
	echo
	echo "=== $name"
	echo "--- node"
	timeout 60 node --disable-warning=ExperimentalWarning oracle/node.mjs "$file" 2>&1; echo "[exit $?]"
	echo "--- native build"
	"$work/adamic" build "$file" -o "$work/$name.bin" 2>&1; echo "[build exit $?]"
	if [ -x "$work/$name.bin" ]; then echo "--- native run"; timeout 60 "$work/$name.bin" 2>&1; echo "[exit $?]"; fi
	echo "--- wasm32 build"
	"$work/adamic" build --target wasm32-wasi "$file" -o "$work/$name.wasm" 2>&1; echo "[build exit $?]"
	if [ -f "$work/$name.wasm" ]; then echo "--- wasm32 run"; timeout 60 node --disable-warning=ExperimentalWarning oracle/wasi.mjs "$work/$name.wasm" 2>&1; echo "[exit $?]"; fi
	echo "--- js compile"
	if "$work/adamic" js "$file" >"$work/$name.mjs" 2>"$work/$name.js.err"; then
		echo "[compile exit 0]"; echo "--- js run"
		timeout 60 node --disable-warning=ExperimentalWarning --import "$work/hook.mjs" "$work/$name.mjs" 2>&1; echo "[exit $?]"
	else
		echo "[compile exit $?]"; cat "$work/$name.js.err"
	fi
done

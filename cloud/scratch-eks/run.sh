#!/usr/bin/env bash
# run.sh: scratch probes for #eksbz4c, compact: Node's whole answer, then each backend's first line. Not for main.
set -u
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$root" || exit 2
work=$(mktemp -d "${TMPDIR:-/tmp}/tags.XXXXXX")
export GOTMPDIR=${GOTMPDIR:-$work/gotmp}
mkdir -p "$GOTMPDIR"
if ! go build -o "$work/adamic" ./cmd/adamic >"$work/build.log" 2>&1; then
	echo "compiler build failed"; tail -40 "$work/build.log"; exit 2
fi
printf "import { registerHooks } from 'node:module';\nregisterHooks({ resolve(specifier, context, next) { return specifier === 'adamic' ? { url: 'file://%s/oracle/adamic.mjs', shortCircuit: true } : next(specifier, context); } });\n" "$root" >"$work/hook.mjs"
line() { tr '\n' '~' <"$1" | sed "s#$root/##g" | cut -c1-230; echo; }
for file in cloud/scratch-eks/*.ts; do
	name=$(basename "$file" .ts)
	echo "=== $name"
	timeout 60 node --disable-warning=ExperimentalWarning oracle/node.mjs "$file" >"$work/$name.node" 2>&1; echo "node [exit $?]"
	if [ "$name" = t1_table ]; then cat "$work/$name.node"; continue; fi
	line "$work/$name.node"
	if "$work/adamic" js "$file" >"$work/$name.mjs" 2>"$work/$name.js.err"; then
		timeout 60 node --disable-warning=ExperimentalWarning --import "$work/hook.mjs" "$work/$name.mjs" >"$work/$name.js" 2>&1; printf 'js [exit %s] ' $?; line "$work/$name.js"
	else
		printf 'js stops '; line "$work/$name.js.err"
	fi
	if "$work/adamic" build "$file" -o "$work/$name.san" --sanitize >"$work/$name.san.err" 2>&1; then
		ASAN_OPTIONS=detect_leaks=1 timeout 60 "$work/$name.san" >"$work/$name.san.out" 2>&1; printf 'native [exit %s] ' $?; line "$work/$name.san.out"
	else
		printf 'native stops '; line "$work/$name.san.err"
	fi
	if "$work/adamic" build --target wasm32-wasi "$file" -o "$work/$name.wasm" >"$work/$name.wasm.err" 2>&1; then
		timeout 60 node --disable-warning=ExperimentalWarning oracle/wasi.mjs "$work/$name.wasm" >"$work/$name.wasm.out" 2>&1; printf 'wasm32 [exit %s] ' $?; line "$work/$name.wasm.out"
	else
		printf 'wasm32 stops '; line "$work/$name.wasm.err"
	fi
done

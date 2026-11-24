#!/usr/bin/env bash
# run.sh: scratch probes for #eksbz4c's suspected miscompiles, one line a program: Node's answer, then each
# backend as same, stop, or its differing answer. Not for main.
set -u
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$root" || exit 2
work=$(mktemp -d "${TMPDIR:-/tmp}/suspects.XXXXXX")
export GOTMPDIR=${GOTMPDIR:-$work/gotmp}
mkdir -p "$GOTMPDIR"
if ! go build -o "$work/adamic" ./cmd/adamic >"$work/build.log" 2>&1; then
	echo "compiler build failed"; tail -40 "$work/build.log"; exit 2
fi
printf "import { registerHooks } from 'node:module';\nregisterHooks({ resolve(specifier, context, next) { return specifier === 'adamic' ? { url: 'file://%s/oracle/adamic.mjs', shortCircuit: true } : next(specifier, context); } });\n" "$root" >"$work/hook.mjs"
flat() { tr '\n' '~' <"$1" | sed "s#$root/##g; s#cloud/scratch-eks/##g; s#stage 0 can't lower ##; s#Adamic 0.1 refuses ##" | cut -c1-"$2"; }
differ=""
verdict() { # name backend outputfile exit
	if [ "$(cat "$3")" = "$(cat "$work/$1.node")" ] && [ "$4" = "$(cat "$work/$1.node.exit")" ]; then printf ' | %s same' "$2"; else printf ' | %s DIFF[%s] %s' "$2" "$4" "$(flat "$3" 90)"; differ="$differ $1:$2"; fi
}
for file in cloud/scratch-eks/*.ts; do
	name=$(basename "$file" .ts)
	timeout 60 node --disable-warning=ExperimentalWarning oracle/node.mjs "$file" >"$work/$name.node" 2>&1; echo $? >"$work/$name.node.exit"
	printf '%s | node[%s] %s' "$name" "$(cat "$work/$name.node.exit")" "$(flat "$work/$name.node" 70)"
	if "$work/adamic" js "$file" >"$work/$name.mjs" 2>"$work/$name.js.err"; then
		timeout 60 node --disable-warning=ExperimentalWarning --import "$work/hook.mjs" "$work/$name.mjs" >"$work/$name.js" 2>&1; verdict "$name" js "$work/$name.js" $?
		if "$work/adamic" build "$file" -o "$work/$name.san" --sanitize >"$work/$name.san.err" 2>&1; then
			ASAN_OPTIONS=detect_leaks=1 timeout 60 "$work/$name.san" >"$work/$name.san.out" 2>&1; verdict "$name" sanitized "$work/$name.san.out" $?
		else printf ' | sanitized stop %s' "$(flat "$work/$name.san.err" 80)"; fi
		if "$work/adamic" build "$file" -o "$work/$name.rel" >"$work/$name.rel.err" 2>&1; then
			timeout 60 "$work/$name.rel" >"$work/$name.rel.out" 2>&1; verdict "$name" release "$work/$name.rel.out" $?
		else printf ' | release stop'; fi
		if "$work/adamic" build --target wasm32-wasi "$file" -o "$work/$name.wasm" >"$work/$name.wasm.err" 2>&1; then
			timeout 60 node --disable-warning=ExperimentalWarning oracle/wasi.mjs "$work/$name.wasm" >"$work/$name.wasm.out" 2>&1; verdict "$name" wasm32 "$work/$name.wasm.out" $?
		else printf ' | wasm32 stop'; fi
	else
		printf ' | stops %s' "$(flat "$work/$name.js.err" 110)"
	fi
	echo
done
echo "=== lowered and differ from Node:${differ:- none}"

#!/usr/bin/env bash
# One-tree paired latent census, scratch only (never lands). Laid at a scratch commit whose tree is the head
# (diagnostics/cast-causes-1) plus this directory, which holds the base's one differing compiler
# file (internal/lower/cast_proof.go at d1a5561a). Both sides build from this tree, base first with that file swapped
# in, so the two censuses read the same adapted bytes at the same paths, which stage3/census/pair needs.
# Prints PASS or FAIL per step and both pair reports whole.
set -uo pipefail
root=$PWD
work=$TMPDIR/pair
mkdir -p "$work"
step() { local name=$1; shift; if "$@"; then echo "PASS $name"; else echo "FAIL $name (exit $?)"; exit 1; fi; }

head_file=$work/cast_proof.head.go
cp internal/lower/cast_proof.go "$head_file"
apply() { bash stage3/apply.sh "$work/adapted" > "$work/apply.log" 2>&1 || { tail -40 "$work/apply.log"; return 1; }; }
step apply apply
adapted=$work/adapted

build() {
	local side=$1 out=$work/$1
	mkdir -p "$out"
	go build -o "$out/census" ./stage3/census/tool &&
	python3 stage3/census/latent/make_overlay.py "$root" "$out/latent-overlay" > "$out/latent-overlay.log" 2>&1 &&
	gofmt -w "$out/latent-overlay/"*.go &&
	python3 stage3/meter/entry_overlay.py "$out/latent-overlay" "$out/entry-overlay" > "$out/entry-overlay.log" 2>&1 &&
	gofmt -w "$out/entry-overlay/"*.go &&
	go build -buildvcs=false -overlay="$out/entry-overlay/overlay.json" -o "$out/entry-census" ./stage3/census/latent/tool &&
	go build -buildvcs=false -overlay="$out/latent-overlay/overlay.json" -o "$out/latent-census" ./stage3/census/latent/tool &&
	sha256sum internal/lower/cast_proof.go "$out/census" "$out/entry-census" "$out/latent-census"
}
measure() {
	local side=$1 out=$work/$1 begun=$SECONDS
	"$out/census" "$adapted/src/compiler" "$out/roots.jsonl" > "$out/roots.log" 2>&1 &&
	LATENT_ASSERT_NO_OUTPUT=1 LATENT_TAIL=1 "$out/entry-census" "$adapted/src/tsc/tsc.ts" "$out/entry.jsonl" > "$out/entry.log" 2>&1 &&
	LATENT_ASSERT_NO_OUTPUT=1 LATENT_TAIL=1 "$out/latent-census" "$adapted/src/compiler" "$out/files.jsonl" > "$out/files.log" 2>&1 &&
	echo "$side measured in $((SECONDS - begun)) s"
}

cp pair-scratch/cast_proof.base.go internal/lower/cast_proof.go
step build-base build base
cp "$head_file" internal/lower/cast_proof.go
step build-head build head
cmp -s internal/lower/cast_proof.go "$head_file" || { echo "FAIL head file not restored"; exit 1; }
step measure-base measure base || { tail -20 "$work/base/"*.log; exit 1; }
step measure-head measure head || { tail -20 "$work/head/"*.log; exit 1; }

for pass in entry files; do
	step "pair-$pass" python3 stage3/census/pair/pair.py "$work/base/$pass.jsonl" "$work/head/$pass.jsonl" "$adapted" \
		--typescript stage3/api/node_modules/typescript --json "$work/$pass-delta.json" --report "$work/$pass-report.md"
	echo "===== $pass pass (entry: tsc.ts's program; files: each src/compiler file as its own program)"
	cat "$work/$pass-report.md"
done
echo "===== per-root census, lines that differ (base <, head >)"
diff <(sort "$work/base/roots.jsonl") <(sort "$work/head/roots.jsonl") && echo "roots: identical"
echo PASS

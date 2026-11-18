#!/usr/bin/env bash
# One-tree paired latent census, scratch only (never lands). Laid at a scratch commit whose tree is the head
# (diagnostics/stale-narrowing-4 ce09723b) plus this directory, which holds the base's three differing compiler
# files (internal/lower/{constraint_relation,lower,narrowed_nullish}.go at 375c551c). Both sides build from this tree,
# base first with those files swapped in, so the two censuses read the same adapted bytes at the same paths.
# Prints PASS or FAIL per step and both pair reports whole.
set -uo pipefail
root=$PWD
work=$TMPDIR/pair
# PAIR_PASSES picks what each side measures (roots, entry, files), so a 2 h proof can take it in halves.
passes=${PAIR_PASSES:-roots entry files}
wants() { case " $passes " in *" $1 "*) return 0 ;; esac; return 1; }
echo "passes: $passes"
mkdir -p "$work"
step() { local name=$1; shift; if "$@"; then echo "PASS $name"; else echo "FAIL $name (exit $?)"; exit 1; fi; }

files="constraint_relation lower narrowed_nullish"
for f in $files; do cp "internal/lower/$f.go" "$work/$f.head.go"; done
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
	sha256sum internal/lower/constraint_relation.go internal/lower/lower.go internal/lower/narrowed_nullish.go "$out/census" "$out/entry-census" "$out/latent-census"
}
measure() {
	local side=$1 out=$work/$1 begun=$SECONDS
	{ ! wants roots || "$out/census" "$adapted/src/compiler" "$out/roots.jsonl" > "$out/roots.log" 2>&1; } &&
	{ ! wants entry || LATENT_ASSERT_NO_OUTPUT=1 LATENT_TAIL=1 "$out/entry-census" "$adapted/src/tsc/tsc.ts" "$out/entry.jsonl" > "$out/entry.log" 2>&1; } &&
	{ ! wants files || LATENT_ASSERT_NO_OUTPUT=1 LATENT_TAIL=1 "$out/latent-census" "$adapted/src/compiler" "$out/files.jsonl" > "$out/files.log" 2>&1; } &&
	echo "$side measured in $((SECONDS - begun)) s"
}

for f in $files; do cp "pair-scratch/$f.base.go" "internal/lower/$f.go"; done
step build-base build base
for f in $files; do cp "$work/$f.head.go" "internal/lower/$f.go"; done
step build-head build head
for f in $files; do cmp -s "internal/lower/$f.go" "$work/$f.head.go" || { echo "FAIL head file $f not restored"; exit 1; }; done
step measure-base measure base || { tail -20 "$work/base/"*.log; exit 1; }
step measure-head measure head || { tail -20 "$work/head/"*.log; exit 1; }

for pass in entry files; do
	wants "$pass" || continue
	step "pair-$pass" python3 stage3/census/pair/pair.py "$work/base/$pass.jsonl" "$work/head/$pass.jsonl" "$adapted" \
		--typescript stage3/api/node_modules/typescript --json "$work/$pass-delta.json" --report "$work/$pass-report.md"
	echo "===== $pass pass (entry: tsc.ts's program; files: each src/compiler file as its own program)"
	cat "$work/$pass-report.md"
done
if wants roots; then
echo "===== per-root census, lines that differ (base <, head >)"
diff <(sort "$work/base/roots.jsonl") <(sort "$work/head/roots.jsonl") && echo "roots: identical"
fi
echo PASS

#!/bin/bash
# c2 (#snypn8c): what lowering admits moved, so oracle whole, the latent census and native with WASI on.
# Scratch only: never lands. Verdict lines begin "c2w: verdict". The admission delta can't run sealed (it
# builds both compilers from a cohere repository the box doesn't hold); the door runs it on submit.
set -u
# The sealed helpers are shell functions, which a child bash doesn't inherit, so they're sourced here.
source /opt/adamic-block/environment.sh
source cloud/sealed-job.sh
v() { echo "c2w: verdict $*"; }
echo "c2w: head $(git rev-parse HEAD~1)"
sealed_npm stage3/api stage3/typescript > "$TMPDIR/npm.log" 2>&1 && v "npm PASS" || { tail -5 "$TMPDIR/npm.log"; v "npm FAIL"; }
sealed_products ./internal/oracle ./internal/native > "$TMPDIR/products.log" 2>&1 && v "products PASS" || { tail -10 "$TMPDIR/products.log"; v "products FAIL"; }
for package in ./internal/oracle/ ./stage3/census/latent/...; do
	go test -count=1 -timeout 45m $package > "$TMPDIR/run.log" 2>&1
	rc=$?
	grep -E "^(--- FAIL|FAIL|ok |panic:)" "$TMPDIR/run.log" | cut -c1-240 | head -30
	v "$package exit=$rc"
done
ADAMIC_TEST_WASI=1 go test -count=1 -timeout 45m ./internal/native/ > "$TMPDIR/native.log" 2>&1
rc=$?
grep -E "^(--- FAIL|FAIL|ok |panic:)" "$TMPDIR/native.log" | cut -c1-240 | head -30
v "internal/native ADAMIC_TEST_WASI=1 exit=$rc"

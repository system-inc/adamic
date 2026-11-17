#!/bin/bash
# H1 (#tb722h2) re-cut: the moved runtime's readers, internal/lower whole (graph_records_test reads the C),
# and internal/native whole. Scratch only: this file never lands.
source /opt/adamic-block/environment.sh
source cloud/sealed-job.sh
v() { echo "h1: verdict $*"; }
go build ./... && v "build PASS" || v "build FAIL"
go test -count=1 -timeout 10m ./internal/native_runtime/... && v "source_paths guard PASS" || v "source_paths guard FAIL"
sealed_products ./internal/lower ./internal/native > "$TMPDIR/products.log" 2>&1 && v "products PASS" || { tail -20 "$TMPDIR/products.log"; v "products FAIL"; }
go test -count=1 -timeout 30m ./internal/lower/ > "$TMPDIR/lower.log" 2>&1
rc=$?
grep -E "^(--- FAIL|FAIL|ok |panic:)|no such file" "$TMPDIR/lower.log" | head -40
v "internal/lower whole exit=$rc"
go test -count=1 -timeout 30m ./internal/native/ > "$TMPDIR/native.log" 2>&1
rc=$?
grep -E "^(--- FAIL|FAIL|ok |panic:)|no such file" "$TMPDIR/native.log" | head -40
v "internal/native whole exit=$rc"

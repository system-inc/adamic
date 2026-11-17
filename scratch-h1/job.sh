#!/bin/bash
# H1 at c7204d5b (#tb722h2): the moved runtime's readers, and internal/native whole.
# Scratch only: this file never lands.
source /opt/adamic-block/environment.sh
source cloud/sealed-job.sh
v() { echo "h1: verdict $*"; }
go build ./... && v "build PASS" || v "build FAIL"
go test -count=1 -timeout 10m ./internal/native_runtime/... && v "source_paths guard PASS" || v "source_paths guard FAIL"
go test -count=1 -timeout 10m -run 'TestAsyncHostPromises' ./internal/native/ > "$TMPDIR/async.log" 2>&1 && v "async_host PASS" || { tail -40 "$TMPDIR/async.log"; v "async_host FAIL"; }
sealed_products ./internal/native > "$TMPDIR/products.log" 2>&1 && v "products PASS" || { tail -20 "$TMPDIR/products.log"; v "products FAIL"; }
go test -count=1 -timeout 34m ./internal/native/ > "$TMPDIR/native.log" 2>&1
rc=$?
grep -E "^(--- FAIL|FAIL|ok |panic:)|no such file" "$TMPDIR/native.log" | head -80
v "internal/native whole exit=$rc"

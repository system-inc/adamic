cd "$(git rev-parse --show-toplevel)"
echo "== gofmt"; gofmt -l cmd internal; echo "gofmt done"
go build -o /tmp/a ./cmd/adamic || exit 1
for p in internal/oracle/testdata/node_process_handle_view.a probe/p3.a probe/p4.a; do
  echo "===== $p"
  /tmp/a js $p > /tmp/v.mjs; echo "js exit=$?"
  echo "--- emitted js program body"; sed -n '/^let local_0_handle/,$p' /tmp/v.mjs
  grep -n "const adamicHostStdoutHandle\|const adamicCallee = " /tmp/v.mjs | cut -c1-260
  for d in pipe; do
    echo "--- node source ($d)"; node --disable-warning=ExperimentalWarning oracle/node.mjs $p | cat; echo "exit=${PIPESTATUS[0]}"
    echo "--- js backend ($d)"; node --disable-warning=ExperimentalWarning oracle/node.mjs /tmp/v.mjs | cat; echo "exit=${PIPESTATUS[0]}"
  done
  /tmp/a c $p > /tmp/v.c; echo "c exit=$?"
  echo "--- emitted c around the call"; grep -n "stdout_handle\|setBlocking\|->code(" /tmp/v.c | grep -v "^.*static adamic_string" | head -20
  n=$(grep -n "adamic_node_stdout_handle()" /tmp/v.c | head -1 | cut -d: -f1); [ -n "$n" ] && sed -n "${n},$((n+40))p" /tmp/v.c
done
echo "== vet"; go vet ./internal/... 2>&1 | tail -20; echo "vet exit=${PIPESTATUS[0]}"
echo "== javascript"; go test ./internal/javascript -count=1 2>&1 | tail -30
echo "== oracle targeted"; go test ./internal/oracle -run 'TestNodeProcess|TestReflectGet' -count=1 -v 2>&1 | grep -E "^(=== RUN|--- |ok|FAIL|PASS)|handle|Handle" | grep -v "=== RUN" | tail -80
echo "== mutant"
cp internal/javascript/library_node_process.go /tmp/lnp.go
grep -c '"handle": "adamicHostStdoutHandle(process.stdout._handle)"' internal/javascript/library_node_process.go
sed -i 's|"handle": "adamicHostStdoutHandle(process.stdout._handle)"|"handle": "process.stdout._handle"|' internal/javascript/library_node_process.go
git diff --stat
go test ./internal/oracle -run 'TestNodeProcessHandleView$' -count=1 -v 2>&1 | tail -20
cp /tmp/lnp.go internal/javascript/library_node_process.go; git diff --stat; echo "mutant restored"

cd "$(git rev-parse --show-toplevel)"
echo "== gofmt"; gofmt -l cmd internal; echo "gofmt done"
echo "== statics list (node_process)"; go test ./internal/native -run 'TestRuntimeStaticsAreListed$' -count=1 -args -runtime-statics-list 2>&1 | grep "node_process.c" 
echo "== statics"; go test ./internal/native -run 'Statics' -count=1 2>&1 | tail -20
go build -o /tmp/a ./cmd/adamic || exit 1
for d in pipe; do
  /tmp/a js internal/oracle/testdata/node_process_handle_identity.a > /tmp/i.mjs; echo "js exit=$?"
  sed -n '/^let local_0_first/,$p' /tmp/i.mjs; head -3 /tmp/i.mjs | cut -c1-200
  node --disable-warning=ExperimentalWarning oracle/node.mjs internal/oracle/testdata/node_process_handle_identity.a | cat
  node --disable-warning=ExperimentalWarning oracle/node.mjs /tmp/i.mjs | cat
  /tmp/a build internal/oracle/testdata/node_process_handle_identity.a -o /tmp/i.bin && /tmp/i.bin | cat
done
echo "== targeted"; go test ./internal/oracle -run 'TestNodeProcess|TestReflectGet|TestCountsAreRecorded' -count=1 2>&1 | tail -40
echo "== native mutant"
cp internal/native/runtime/node_process.c /tmp/np.c
grep -c 'if (handle->shape == NULL) {' internal/native/runtime/node_process.c
sed -i 's/    if (handle->shape == NULL) {/    handle = adamic_object_new(\&shape); {/' internal/native/runtime/node_process.c
git diff internal/native/runtime/node_process.c
go test ./internal/oracle -run 'TestNodeProcessHandleView' -count=1 -v 2>&1 | grep -v '^=== ' | tail -20
cp /tmp/np.c internal/native/runtime/node_process.c
echo "== javascript mutant"
cp internal/javascript/library_node_process.go /tmp/lnp.go
grep -c 'adamicStdoutHandleValue ??= ' internal/javascript/library_node_process.go
sed -i 's/adamicStdoutHandleValue ??= //' internal/javascript/library_node_process.go
git diff internal/javascript/library_node_process.go
go test ./internal/oracle -run 'TestNodeProcessHandleView' -count=1 -v 2>&1 | grep -v '^=== ' | tail -20
cp /tmp/lnp.go internal/javascript/library_node_process.go
git status --short; echo "mutants restored"

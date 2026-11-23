set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
# #cdpxjmb re-cut onto main 96a1291b: stage3/fixtures whole (the gate (B)'s red added), then bang.sh's pins and M1, M4.
echo "box: $(hostname) at $(git rev-parse HEAD)"
ok=1
sandboxed_npm stage3/api stage3/typescript > "$TMPDIR/npm.log" 2>&1 || { echo "FAIL npm"; tail -20 "$TMPDIR/npm.log"; ok=0; }
go test -count=1 ./stage3/fixtures/ > "$TMPDIR/fixtures.log" 2>&1
if [ $? -eq 0 ]; then echo "PASS stage3/fixtures whole"; else echo "FAIL stage3/fixtures whole"; grep -E "^(--- FAIL|    --- FAIL|panic:)|_test.go:[0-9]+:" "$TMPDIR/fixtures.log" | head -40; ok=0; fi
bash scratch-oc/bang.sh || ok=0
if [ $ok = 1 ]; then echo "FIXTURES PASS"; else echo "FIXTURES FAIL"; exit 1; fi

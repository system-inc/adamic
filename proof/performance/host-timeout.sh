# #q89xaef: is TestHostConfigurationFixtureReaders' red a child timing out at -test.timeout=30s? Runs it as written,
# then with the child's bound at 120s, on one box. Never merged.
set -u
. "$HOME/adamic-tools/env.sh" > /dev/null 2>&1 || { echo "VERDICT FAIL env"; exit 3; }
git -c protocol.file.allow=always submodule update --init --recursive --depth 1 > /tmp/ht-sub.log 2>&1 || { echo "VERDICT FAIL submodules"; exit 3; }
echo "main $(git log --format=%h -1 --invert-grep --grep='^proof: ' HEAD) on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
f=stage3/fixtures/host_configuration_test.go
run() { local start=$(date +%s); go test -count=1 -v -run '^TestHostConfigurationFixtureReaders$' ./stage3/fixtures/ > /tmp/ht-$1.log 2>&1; local code=$?
  echo "== $1: exit $code in $(( $(date +%s) - start )) s, timed-out children $(grep -c 'test timed out after' /tmp/ht-$1.log)"
  grep -E '^\s+--- (PASS|FAIL): TestHostConfigurationFixtureReaders/' /tmp/ht-$1.log | sed 's/^ *//' | head -12; }
if [ "${1:-alone}" = contended ]; then
  # J-avaja81's shape on Pupa: the whole package, its other tests beside the host test, on 8 cores.
  run() { local start=$(date +%s); taskset -c 0-7 go test -count=1 -v ./stage3/fixtures/... > /tmp/ht-$1.log 2>&1; local code=$?
    echo "== $1 (whole package, taskset 0-7): exit $code in $(( $(date +%s) - start )) s, timed-out children $(grep -c 'test timed out after' /tmp/ht-$1.log)"
    grep -E '^(--- (PASS|FAIL): TestHostConfigurationFixtureReaders|\s+--- FAIL)' /tmp/ht-$1.log | sed 's/^ *//' | head -6
    grep -m2 -B2 'test timed out after' /tmp/ht-$1.log | cut -c1-200; }
fi
run as-written
grep -q -- '-test.timeout=30s' $f || { echo "VERDICT FAIL no 30s bound to raise"; exit 3; }
sed -i 's/-test.timeout=30s/-test.timeout=120s/' $f
run at-120s
git checkout -q -- $f
echo "VERDICT PROBED"

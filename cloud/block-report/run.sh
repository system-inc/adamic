#!/usr/bin/env bash
# cloud/block-report/run.sh is one Adamic block's own build and tests, run as a sealed job on a rented machine with no
# network, no root, no token and no loom (#highfive, piece 6; Node's road 3 on Nov 12). It isn't the house's build of
# the tree: no loom keys are made and none are compared. Its answer is a report, read by hand beside the house's decided
# run of the same commit, and nothing in it lands, charges or clears anything.
#
#   ahra loom job 'bash cloud/block-report/run.sh' --kind block-report --task highfive --at adamic@<sha>
#
# Everything the build fetches was laid on the image ahead of time by cloud/block-report/environment.sh, read-only, and
# its environment.sh is sourced here: Go reads modules only from that cache (GOPROXY=off, -mod=readonly) and npm only
# from its cache, so a module the image lacks fails closed with its name, never by reaching out.
#
# The job starts in a worktree of the asked commit with both submodules laid. It writes its report to stdout between
# marker lines, since nothing else leaves a job yet, and keeps the raw streams beside the worktree.
set -u

environment=${BLOCK_REPORT_ENVIRONMENT:-/opt/adamic-block/environment.sh}
[ -r "$environment" ] || { echo "block-report: no environment at $environment, so nothing ran"; exit 64; }
# shellcheck disable=SC1090
. "$environment" || { echo "block-report: $environment didn't source, so nothing ran"; exit 64; }
sha=$(git rev-parse --verify 'HEAD^{commit}') || { echo "block-report: $PWD holds no commit, so nothing ran"; exit 64; }
dirty=$(git submodule status --recursive | grep -E '^[-+U]' || true)
[ -z "$dirty" ] || { echo "block-report: a submodule isn't at its gitlink: $dirty"; exit 64; }

out="$PWD.block-report"
mkdir "$out" || { echo "block-report: $out is there already or can't be made, so nothing ran"; exit 64; }
export TMPDIR="$out/tmp" GOCACHE="$out/gocache"
mkdir "$TMPDIR" "$GOCACHE" || exit 64

# The flags the house builds every product with (internal/buildcache.ReproducibleGoFlags), so a binary's bytes are a
# function of its sources and not of this checkout's path or commit.
reproducible="-trimpath -ldflags=-buildid= -buildvcs=false"
started=$(date -u +%FT%TZ)
echo "block-report: $(uname -n), $(nproc) cores, $(go version), commit $sha, go flags $reproducible $GOFLAGS, proxy $GOPROXY"
# The job is asked at a commit carrying these scripts on top of the block (#highfive: a job lays exactly its commit), so
# the report names the block, its parent, and shows that the scripts are the only difference.
block=$(git rev-parse --verify 'HEAD^' 2> /dev/null || true)
echo "block-report: the block is $block, this commit's parent; what this commit adds to it:"
[ -z "$block" ] || git diff --stat "$block" HEAD | sed 's/^/block-report:   /'

step() {
	local name=$1 begun=$SECONDS
	shift
	"$@" > "$out/$name.log" 2>&1
	local code=$?
	echo "block-report: $name exited $code in $((SECONDS - begun)) s"
	return "$code"
}
# The whole tree unless a trial names its packages (BLOCK_REPORT_PACKAGES, space-separated); the block's real run never
# sets it.
packages=${BLOCK_REPORT_PACKAGES:-./...}
# shellcheck disable=SC2086
step vet go vet $packages
vetExit=$?
# shellcheck disable=SC2086
step build go build $reproducible $packages
buildExit=$?
# One test stream for every package, uncached, in one process tree; a package that doesn't build is in the stream too.
# shellcheck disable=SC2086
step test go test -count=1 -json $reproducible $packages
testExit=$?
# The test log is the JSON stream; step wrote stdout and stderr into one file, so keep stderr's lines apart for the
# report by reading only the lines that parse.
cp "$out/test.log" "$out/test.jsonl"

echo "BEGIN block-report"
python3 -I - "$out/test.jsonl" "$sha" "$started" "$vetExit" "$buildExit" "$testExit" "$reproducible" << 'PYTHON'
import json, sys, collections
stream, sha, started, vet, build, test, flags = sys.argv[1:]
packages = collections.OrderedDict()
failing = collections.defaultdict(list)
for line in open(stream, errors='replace'):
    try:
        event = json.loads(line)
    except ValueError:
        continue
    package = event.get('Package')
    if not package:
        continue
    action = event.get('Action')
    if event.get('Test') is None and action in ('pass', 'fail', 'skip'):
        packages[package] = {'outcome': {'pass': 'ok', 'fail': 'failed', 'skip': 'skipped'}[action], 'seconds': event.get('Elapsed')}
    elif event.get('Test') and action == 'fail':
        failing[package].append(event['Test'])
counts = collections.Counter(row['outcome'] for row in packages.values())
print("This is the block's own tests run without loom, not the house's build of the tree. Nothing in it is a key, and nothing compares keys.")
print(json.dumps({'sha': sha, 'started': started, 'vetExit': int(vet), 'buildExit': int(build), 'testExit': int(test),
                  'goFlags': flags, 'packages': len(packages), 'ok': counts['ok'], 'failed': counts['failed'], 'skipped': counts['skipped']}))
for package, row in packages.items():
    print(json.dumps({'package': package, 'outcome': row['outcome'], 'seconds': row['seconds'], 'failingTests': failing.get(package, [])}))
print(f"::loom checked {len(packages)} packages at {sha[:12]}: {counts['ok']} ok, {counts['failed']} failed, {counts['skipped']} skipped")
PYTHON
echo "END block-report"
echo "block-report: raw streams kept in $out"
# The job's own answer is the tests': 0 only when vet, build and every package's tests passed. A failure here is a
# finding to read beside the house, never a verdict on the block.
[ "$vetExit" = 0 ] && [ "$buildExit" = 0 ] && [ "$testExit" = 0 ]

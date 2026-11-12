#!/usr/bin/env bash
# cloud/block-report/run.sh is one Adamic block's own build and tests, run as a sealed job on a rented machine with no
# network, no root, no token and no loom (#highfive, piece 6; Node's road 3 on Nov 12). It isn't the house's build of
# the tree: no loom keys are made and none are compared. Its answer is a report, read by hand beside the house's decided
# run of the same commit, and nothing in it lands, charges or clears anything.
#
#   ahra loom job 'bash cloud/block-report/run.sh' --kind block-report --task <id> --at adamic@<sha>
#
# Everything the build fetches was laid on the image ahead of time by cloud/block-report/environment.sh, read-only, and
# its environment.sh is sourced here: Go reads modules only from that cache and npm only from its cache, so a module the
# image lacks fails closed with its name, never by reaching out.
#
# The wrapper lays the asked commit as archives: no .git, and gitlinks one level deep, so cohere arrives and
# cohere/TypeScript is an empty directory (block 2334 on azure-block-1, Nov 12). This script asks git nothing about the
# tree, lays cohere/TypeScript itself from the machine's own mirror, and keeps every scratch file outside the tree.
set -u

mirrors=${BLOCK_REPORT_MIRRORS:-/var/lib/loom/mirrors}
environment=${BLOCK_REPORT_ENVIRONMENT:-/opt/adamic-block/environment.sh}

[ -f go.mod ] && [ -f cohere/go.mod ] || { echo "block-report: $PWD lacks adamic's go.mod or cohere's, so nothing ran"; exit 64; }
# The commit: the wrapper's word for it when it gives one (#sgn0nk9), else git's when the tree is a checkout.
sha=${LOOM_JOB_COMMIT:-$(git rev-parse --verify 'HEAD^{commit}' 2> /dev/null || echo "unnamed: the wrapper gives no commit and the tree has no .git")}
# cohere/TypeScript's pin, as the tree's own constant states it; internal/load's pin test holds it to the gitlink.
typescriptCommit=$(sed -n 's/^const TypeScriptGoCommit = "\([0-9a-f]\{40\}\)"$/\1/p' internal/corpusfiles/files.go)
[ "${#typescriptCommit}" = 40 ] || { echo "block-report: internal/corpusfiles/files.go names no TypeScriptGoCommit, so nothing ran"; exit 64; }
echo "block-report: the tree at $sha, as laid: $([ -e .git ] && echo "a checkout" || echo "an archive with no .git")"
[ -r "$environment" ] || { echo "block-report: no environment at $environment, so nothing ran"; exit 64; }
# shellcheck disable=SC1090
. "$environment" || { echo "block-report: $environment didn't source, so nothing ran"; exit 64; }

# The streams and every scratch file live outside the tree under test: tests that walk or count the tree see anything
# inside it, leading dot or not (J-bp8ryag, Nov 12: ten packages failed only because the first run kept them in the
# tree). The unit's /tmp is its own private one (PrivateTmp=yes).
out="${BLOCK_REPORT_OUT:-/tmp/block-report}"
mkdir "$out" || { echo "block-report: $out is there already or can't be made, so nothing ran"; exit 64; }
# The unit hands the script no HOME (the first small job, J-00vkbwf), so the job gives itself one there too.
export TMPDIR="$out/tmp" GOCACHE="$out/gocache" HOME="$out/home" GOPATH="$out/home/go"
mkdir "$TMPDIR" "$GOCACHE" "$HOME" "$GOPATH" || exit 64
# The mirrors are root's and the job is its own user, so git refuses them as dubious unless told otherwise.
# Configuration from the environment is protected configuration, and nothing is written to disk.
export GIT_CONFIG_COUNT=1 GIT_CONFIG_KEY_0=safe.directory GIT_CONFIG_VALUE_0='*'
echo "block-report: the job gives itself TMPDIR, GOCACHE, HOME and GOPATH under $out"

# cohere/TypeScript: go.mod replaces github.com/microsoft/TypeScript/tsc with ./cohere/TypeScript/tsc and go.work uses it,
# so an empty directory fails every go command at module load. Lay it from root's mirror, read only, or stop here.
if [ -f cohere/TypeScript/tsc/go.mod ]; then
	echo "block-report: cohere/TypeScript is laid already"
else
	mirror=
	for name in TypeScript typescript; do
		if git --git-dir="$mirrors/$name.git" cat-file -e "$typescriptCommit^{commit}" 2> /dev/null; then
			mirror="$mirrors/$name.git"
			break
		fi
	done
	[ -n "$mirror" ] || {
		echo "block-report: no mirror under $mirrors holds TypeScript $typescriptCommit (or the job can't read it): $(ls -l "$mirrors" 2>&1 | tr '\n' ' ')"
		echo "block-report: cohere/TypeScript is empty, so no go command can load the module; nothing ran"
		exit 64
	}
	[ -d cohere/TypeScript ] && [ -z "$(ls -A cohere/TypeScript)" ] || { echo "block-report: cohere/TypeScript isn't an empty directory, so nothing ran"; exit 64; }
	git --git-dir="$mirror" archive --format=tar "$typescriptCommit" | tar -x -C cohere/TypeScript || {
		echo "block-report: laying cohere/TypeScript from $mirror failed, so nothing ran"
		exit 64
	}
	[ -f cohere/TypeScript/tsc/go.mod ] || { echo "block-report: cohere/TypeScript/tsc/go.mod isn't there after the lay, so nothing ran"; exit 64; }
	echo "block-report: cohere/TypeScript laid by the job at $typescriptCommit from $mirror"
fi

# The image's TypeScript source is root's. Pinned git (internal/buildcache.PinnedGit) runs with no system or global
# configuration and an allowlisted environment, on purpose, so the safe.directory above never reaches it: the job reads
# its own copy instead (J-r6ctetr, Nov 12: internal/lower's five census tests green on it, red on root's).
if [ -n "${ADAMIC_TYPESCRIPT_SOURCE:-}" ]; then
	begun=$SECONDS
	cp -a "$ADAMIC_TYPESCRIPT_SOURCE" "$out/typescript-source" || { echo "block-report: copying $ADAMIC_TYPESCRIPT_SOURCE failed, so nothing ran"; exit 64; }
	export ADAMIC_TYPESCRIPT_SOURCE="$out/typescript-source"
	echo "block-report: the image's TypeScript source copied to $ADAMIC_TYPESCRIPT_SOURCE in $((SECONDS - begun)) s, the job's own"
fi

# node_modules for each npm project the image cached, installed offline from its cache, as a house box lays them.
for project in ${BLOCK_REPORT_NPM_PROJECTS:-}; do
	[ -f "$project/package-lock.json" ] || { echo "block-report: $project has no package-lock.json, so nothing ran"; exit 64; }
	(cd "$project" && npm ci --offline --ignore-scripts --no-audit --no-fund > "$out/npm-$(echo "$project" | tr / _).log" 2>&1) ||
		{ echo "block-report: npm ci --offline failed in $project, so nothing ran"; exit 64; }
	echo "block-report: $project/node_modules installed offline"
done

# The flags the house builds every product with (internal/buildcache.ReproducibleGoFlags), so a binary's bytes are a
# function of its sources and not of this checkout's path or commit.
reproducible="-trimpath -ldflags=-buildid= -buildvcs=false"
started=$(date -u +%FT%TZ)
echo "block-report: $(uname -n), $(nproc) cores, $(go version), node $(node --version), go flags $reproducible $GOFLAGS, proxy $GOPROXY"
echo "block-report: tests that ask git about the tree fail on an archive for the missing .git; each is a laid-tree difference, never the block's"

step() {
	local name=$1 begun=$SECONDS
	shift
	"$@" > "$out/$name.log" 2>&1
	local code=$?
	echo "block-report: $name exited $code in $((SECONDS - begun)) s"
	return "$code"
}
# The whole tree unless a rehearsal names its packages (BLOCK_REPORT_PACKAGES, space-separated); the block's real run
# never sets it.
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

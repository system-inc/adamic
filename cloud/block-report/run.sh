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
# image lacks fails closed with its name, never by reaching out. Then cloud/sandboxed-job.sh, the tree's own preamble for a
# sandboxed job, is sourced, so each gap the sandboxed road has settled is settled here the same way (#3jge4gw: the reports
# before it failed on setup gaps, not on the block).
#
# The wrapper lays the asked commit as archives: no .git, and gitlinks one level deep, so cohere arrives and
# cohere/TypeScript is an empty directory (block 2334 on azure-block-1, Nov 12). This script asks git nothing about the
# tree, lays cohere/TypeScript itself from the machine's own mirror, and keeps every scratch file outside the tree.
set -u

mirrors=${BLOCK_REPORT_MIRRORS:-/var/lib/loom/mirrors}
environment=${BLOCK_REPORT_ENVIRONMENT:-/opt/adamic-block/environment.sh}

[ -f go.mod ] && [ -f cohere/go.mod ] || { echo "block-report: $PWD lacks adamic's go.mod or cohere's, so nothing ran"; exit 64; }
[ -r "$environment" ] || { echo "block-report: no environment at $environment, so nothing ran"; exit 64; }
# A commit from before the preamble would report its own setup gaps, so it reports nothing at all (#3jge4gw). Until the
# preamble is on main, an ask may carry its copy in and name it (BLOCK_REPORT_PREAMBLE, a path inside the tree, as a probe
# lays it), which Node allowed on #zwzer5f for image 1.0.10's first night: the steps are the preamble's either way.
preamble=${BLOCK_REPORT_PREAMBLE:-cloud/sandboxed-job.sh}
[ -f "$preamble" ] || { echo "block-report: this commit has no $preamble, so nothing ran and the report counts nothing"; exit 64; }
# shellcheck disable=SC1090
. "$environment" || { echo "block-report: $environment didn't source, so nothing ran"; exit 64; }
# The preamble's steps return nonzero on a failure, and say which on stderr just above this script's own line.
# shellcheck disable=SC1091
. "$preamble" || { echo "block-report: $preamble stopped at a step that failed, in its words above, so nothing ran"; exit 64; }

# Line one is the go that runs everything, and line two the disk the streams land on, so a report read beside the
# house's starts from what ran it and where (#3jge4gw).
echo "block-report: $(go version 2>&1)"
# The streams and every scratch file live outside the tree under test: tests that walk or count the tree see anything
# inside it, leading dot or not (J-bp8ryag, Nov 12: ten packages failed only because the first run kept them in the
# tree). The data-disk directory the runner names beside the tree ($LOOM_JOB_SCRATCH, Machines' #kq1sg4k v2, where
# sandboxed-job.sh puts the build cache too) is taken first, since the unit's /tmp is a small private tmpfs that 1.0.9's
# streams filled (J-4vcc0ar); /tmp only when the runner names none.
out="${BLOCK_REPORT_OUT:-${LOOM_JOB_SCRATCH:+$LOOM_JOB_SCRATCH/block-report}}"
out="${out:-/tmp/block-report}"
mkdir "$out" || { echo "block-report: $out is there already or can't be made, so nothing ran"; exit 64; }
echo "block-report: $out is on $(df -PTk "$out" | awk 'NR == 2 {printf "%s, %s, %.0f bytes free", $1, $2, $5 * 1024}')"
# The preamble makes TMPDIR when the unit gives none, and the job's go and its tests write there first, so it moves to
# the data disk beside the streams. The preamble's build cache (ADAMIC_BUILD_CACHE_DIR) stays under the tree as it set it:
# Operations' proof holds internal/oracle green with it there (#3jge4gw). The unit hands the script no HOME (the first
# small job, J-00vkbwf), so the job gives itself one under $out too.
preambleTemporary=$TMPDIR
export TMPDIR="$out/tmp" GOCACHE="$out/gocache" HOME="$out/home" GOPATH="$out/home/go"
mkdir "$TMPDIR" "$GOCACHE" "$HOME" "$GOPATH" || exit 64
# The mirrors are root's and the job is its own user, so git refuses them as dubious unless told otherwise.
# Configuration from the environment is protected configuration, and nothing is written to disk.
export GIT_CONFIG_COUNT=1 GIT_CONFIG_KEY_0=safe.directory GIT_CONFIG_VALUE_0='*'
echo "block-report: TMPDIR is $TMPDIR, over the preamble's $preambleTemporary, and GOCACHE, HOME and GOPATH are under $out too; the build cache stays at $ADAMIC_BUILD_CACHE_DIR, the preamble's"

# The commit: the wrapper's word for it when it gives one (#sgn0nk9), else git's when the tree is a checkout.
sha=${LOOM_JOB_COMMIT:-$(git rev-parse --verify 'HEAD^{commit}' 2> /dev/null || echo "unnamed: the wrapper gives no commit and the tree has no .git")}
# cohere/TypeScript's pin, as the tree's own constant states it; internal/load's pin test holds it to the gitlink.
typescriptCommit=$(sed -n 's/^const TypeScriptGoCommit = "\([0-9a-f]\{40\}\)"$/\1/p' internal/corpusfiles/files.go)
[ "${#typescriptCommit}" = 40 ] || { echo "block-report: internal/corpusfiles/files.go names no TypeScriptGoCommit, so nothing ran"; exit 64; }
echo "block-report: the tree at $sha, as laid: $([ -e .git ] && echo "a checkout" || echo "an archive with no .git")"

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

# The image's TypeScript source is root's, and internal/buildcache.PinnedGit reads it with no configuration that could
# name it safe, so the job reads a copy of its own (J-r6ctetr, Nov 12). The preamble's sandboxed_typescript_source makes
# it, and the products below read it (#3jge4gw).
sandboxed_typescript_source || { echo "block-report: sandboxed_typescript_source failed, in its words above, so nothing ran"; exit 64; }

# node_modules for each npm project the image cached, installed offline from its cache by the preamble's sandboxed_npm, but
# only where the project's own .gitignore keeps node_modules out of the tree. Any other project's dependencies
# environment.sh lays outside the tree, and an install in cloud/markdown-width left a node_modules that markdowninline's
# corpus walk read as untracked files (#3jge4gw).
npmProjects=""
for project in ${BLOCK_REPORT_NPM_PROJECTS:-}; do
	if ! grep -qxE 'node_modules/?' "$project/.gitignore" 2> /dev/null; then
		echo "block-report: npm skips $project, whose own .gitignore doesn't list node_modules/, so an install would leave untracked files in the tree"
		continue
	fi
	[ -f "$project/package-lock.json" ] || { echo "block-report: $project has no package-lock.json, so nothing ran"; exit 64; }
	npmProjects="$npmProjects $project"
done
if [ -n "$npmProjects" ]; then
	# The projects are the image's own directory names, so they split as words here.
	# shellcheck disable=SC2086
	sandboxed_npm $npmProjects > "$out/npm.log" 2>&1 ||
		{ echo "block-report: sandboxed_npm failed, so nothing ran: $(tail -n 5 "$out/npm.log" | tr '\n' ' ')"; exit 64; }
	echo "block-report: node_modules installed offline in$npmProjects"
fi

# The gate's own four (loom's planner.GateEnvironment): products and tests both run under them, as every unit on the
# house line does, so a test the gate runs isn't skipped here (#3jge4gw).
export ADAMIC_GATE_UNCACHED=1 ADAMIC_TEST_WASI=1 ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_COHERE=1
echo "block-report: products and tests run under the gate's ADAMIC_GATE_UNCACHED=1 ADAMIC_TEST_WASI=1 ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_COHERE=1"

# The flags the house builds every product with (internal/buildcache.ReproducibleGoFlags), so a binary's bytes are a
# function of its sources and not of this checkout's path or commit.
reproducible="-trimpath -ldflags=-buildid= -buildvcs=false"
started=$(date -u +%FT%TZ)
echo "block-report: $(uname -n), $(nproc) cores, node $(node --version), go flags $reproducible $GOFLAGS, proxy $GOPROXY"
echo "block-report: tests that ask git about the tree fail on an archive for the missing .git; each is a laid-tree difference, never the block's"

# Products settle before any test reads them, since a unit test reads only products built ahead and never builds one,
# and this job's build cache starts empty. They settle in waves by loom's lay.sh rule (#3jge4gw, J-myrxd8k): each
# package's TestProduct_ tests come from go test -list, cloud/fast-gate/product-order.json says which wait on which, and
# wave k holds every test whose waits all sit in earlier waves. One python program plans it all, reading the tree's
# files as JSON and running none of them; a failing wave doesn't stop the job, and the report counts its failures.
productsBegun=$SECONDS
# shellcheck disable=SC2086
python3 -I - "$out" cloud/fast-gate/product-order.json cloud/fast-gate/compiler-dependencies.json ${BLOCK_REPORT_PACKAGES:-} << 'PYTHON'
import concurrent.futures, json, os, re, subprocess, sys
out, order_file, dependencies_file = sys.argv[1:4]
named = sys.argv[4:]
part = re.compile(r'^[A-Za-z0-9_][A-Za-z0-9_.-]*$')
product = re.compile(r'^TestProduct_[A-Za-z0-9_]*$')
problems = []

def say(line):
    print('block-report: products: ' + line, flush=True)

def path_of(name):
    # As sandboxed_products reads a package: ./<path> or the module's import path, held to plain path segments.
    for prefix in ('./', 'github.com/system-inc/adamic/'):
        if name.startswith(prefix):
            name = name[len(prefix):]
    name = name.rstrip('/')
    if not name or not all(part.match(piece) for piece in name.split('/')):
        sys.exit('block-report: products: %r is no adamic package path' % name)
    return name

def target(path):
    return '.' if path == '.' else './' + path + '/'

# Every package go list ./... gives, or a rehearsal's own (BLOCK_REPORT_PACKAGES) with each one's
# compiler-dependencies.json entries, as sandboxed_products brings them.
if named:
    listed = {}
    if os.path.exists(dependencies_file):
        listed = json.load(open(dependencies_file)).get('packages')
        if not isinstance(listed, dict):
            sys.exit('block-report: products: %s holds no packages object' % dependencies_file)
    paths = set()
    for name in named:
        path = path_of(name)
        paths.add(path)
        for dependency in listed.get(path, []):
            paths.add(path_of(dependency))
else:
    listing = subprocess.run(['go', 'list', '-e', '-f', '{{.Dir}}', './...'], stdin=subprocess.DEVNULL,
                             stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    if listing.returncode != 0:
        sys.exit('block-report: products: go list ./... exited %d: %s' % (listing.returncode, listing.stderr.strip()[-2000:]))
    paths = {os.path.relpath(directory, os.getcwd()) for directory in listing.stdout.splitlines() if directory}
paths = sorted(paths)

def list_products(path):
    run = subprocess.run(['go', 'test', '-list', '^TestProduct_', target(path)], stdin=subprocess.DEVNULL,
                         stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    return path, run

# go test -list builds each package's test binary, so a few run at once; their results are read back in path order.
names = {}
with open(os.path.join(out, 'products-list.log'), 'w') as log, \
        concurrent.futures.ThreadPoolExecutor(max_workers=min(8, os.cpu_count() or 1)) as pool:
    for path, run in pool.map(list_products, paths):
        log.write('== go test -list ^TestProduct_ %s, exit %d\n%s%s' % (target(path), run.returncode, run.stdout, run.stderr))
        if run.returncode != 0:
            problems.append('go test -list in %s exited %d' % (path, run.returncode))
            continue
        found = []
        for line in run.stdout.splitlines():
            # go test closes each package's list with its own ok or ? line.
            if not line.strip() or re.match(r'^(ok|\?)\s', line):
                continue
            if product.match(line):
                found.append(line)
            else:
                problems.append('%s lists %r, which is no TestProduct_ name' % (path, line[:200]))
        if found:
            names[path] = sorted(found)

# product-order.json's waits: only tests listed above count, and a missing file waits on nothing.
waits = {}
if os.path.exists(order_file):
    try:
        order = json.load(open(order_file))
    except ValueError as error:
        sys.exit('block-report: products: %s is no JSON: %s' % (order_file, error))
    waits = order.get('waits') if isinstance(order, dict) else None
    if not isinstance(waits, dict) or not all(isinstance(each, list) and all(isinstance(wait, str) for wait in each) for each in waits.values()):
        sys.exit('block-report: products: %s holds no waits object of lists of "<path> <test>"' % order_file)
else:
    say('%s is missing, so no product waits on another' % order_file)
tests = ['%s %s' % (path, name) for path in sorted(names) for name in names[path]]
listed_tests = set(tests)
needs = {test: {wait for wait in waits.get(test, []) if wait in listed_tests} for test in tests}

# Wave 1 is every test that waits on none, and wave k every test whose waits all sit in earlier waves. A cycle is the
# products step's failure, never broken: its tests, and every test waiting on one, never run.
waves, placed, remaining = [], set(), set(tests)
while remaining:
    wave = sorted(test for test in remaining if needs[test] <= placed)
    if not wave:
        problems.append('%d product tests wait in a cycle or on one, never run: %s' % (len(remaining), ' '.join(sorted(remaining))))
        break
    waves.append(wave)
    placed.update(wave)
    remaining.difference_update(wave)

# One plan line a wave (wave, tests, packages) for the shell, and one go test a package in each wave's own file.
with open(os.path.join(out, 'products-plan.tsv'), 'w') as plan:
    for index, wave in enumerate(waves, 1):
        by_package = {}
        for test in wave:
            path, name = test.split(' ', 1)
            by_package.setdefault(path, []).append(name)
        plan.write('%d\t%d\t%d\n' % (index, len(wave), len(by_package)))
        with open(os.path.join(out, 'products-wave-%d.tsv' % index), 'w') as runs:
            for path in sorted(by_package):
                runs.write('%s\t^(%s)$\n' % (target(path), '|'.join(by_package[path])))
with open(os.path.join(out, 'products-plan.json'), 'w') as summary:
    json.dump({'orderFile': os.path.exists(order_file), 'packages': len(paths), 'productPackages': len(names),
               'tests': len(tests), 'waves': [len(wave) for wave in waves], 'unplaced': sorted(remaining),
               'problems': problems}, summary)
for problem in problems:
    say(problem)
say('%d product tests in %d of %d packages, in %d waves' % (len(tests), len(names), len(paths), len(waves)))
PYTHON
productsPlanExit=$?
if [ "$productsPlanExit" != 0 ]; then
	echo "block-report: products didn't plan (exit $productsPlanExit), so no wave ran; the tests run on an empty build cache"
	: > "$out/products-plan.tsv"
fi
productsWaves=$(wc -l < "$out/products-plan.tsv" | tr -d ' ')
: > "$out/products.jsonl"
: > "$out/products-waves.tsv"
# Each wave runs one go test a package, packages one after another, every stream appended to products.jsonl. The house
# runs each test binary with -test.timeout=3h, as the tests below do, so a slow machine's product isn't cut at go
# test's own 10 minutes (Pupa, Nov 12).
while IFS=$'\t' read -r wave waveTests wavePackages; do
	waveExit=0
	while IFS=$'\t' read -r path pattern; do
		go test -count=1 -json -timeout 3h -run "$pattern" "$path" >> "$out/products.jsonl" 2>> "$out/products.log" < /dev/null
		code=$?
		[ "$waveExit" != 0 ] || waveExit=$code
	done < "$out/products-wave-$wave.tsv"
	echo "block-report: products wave $wave of $productsWaves: $waveTests tests in $wavePackages packages, exit $waveExit"
	printf '%s\t%s\n' "$wave" "$waveExit" >> "$out/products-waves.tsv"
done < "$out/products-plan.tsv"
echo "block-report: products took $((SECONDS - productsBegun)) s"

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
# The house runs each test binary with -test.timeout=3h; go test's own default of 10 minutes killed internal/lower whole
# on an 8-core box (Pupa, Nov 12), so a slow machine reads as a failed package.
# shellcheck disable=SC2086
step test go test -count=1 -json -timeout 3h $reproducible $packages
testExit=$?
# The streams are read where the steps wrote them, never copied: a copy doubles their bytes, and on 1.0.9's tmpfs /tmp the
# copy ran out of room and the report read 139 of 162 packages from a cut stream (J-4vcc0ar). A disk that filled anyway
# is said here and in the report, so a short stream is never read as a whole one.
outFree=$(df -Pk "$out" | awk 'NR == 2 {printf "%.0f", $4 * 1024}')
[ "${outFree:-0}" -ge 67108864 ] || echo "block-report: $out has ${outFree:-no} bytes free after the tests, so a stream may be cut short and the report below may be partial"

# The report: the tests' summary and one line a package, as before; then the products step's own summary, and each
# failing package's first failure with the first 14 lines it wrote, from either stream, so a report says why without
# its raw streams (#3jge4gw, J-mwkyqjb).
echo "BEGIN block-report"
python3 -I - "$out" "$productsPlanExit" "$sha" "$started" "$vetExit" "$buildExit" "$testExit" "$reproducible" "${outFree:-0}" << 'PYTHON'
import json, os, sys, collections
out, plan_exit, sha, started, vet, build, test, flags, out_free = sys.argv[1:]

def read(path):
    stream = {'packages': collections.OrderedDict(), 'failing': collections.defaultdict(list), 'failed': [],
              'output': collections.defaultdict(list), 'packageOutput': collections.defaultdict(list),
              'buildOutput': collections.defaultdict(list), 'failedBuild': {}}
    for line in open(path, errors='replace'):
        try:
            event = json.loads(line)
        except ValueError:
            continue
        if not isinstance(event, dict):
            continue
        action = event.get('Action')
        # A package that doesn't build says why in build-output events, named by ImportPath, and its fail event names
        # that ImportPath as FailedBuild.
        if action == 'build-output':
            stream['buildOutput'][event.get('ImportPath', '')].append(event.get('Output', ''))
            continue
        package = event.get('Package')
        if not package:
            continue
        if action == 'output':
            if event.get('Test'):
                stream['output'][(package, event['Test'])].append(event.get('Output', ''))
            else:
                stream['packageOutput'][package].append(event.get('Output', ''))
        if event.get('Test') is None and action in ('pass', 'fail', 'skip'):
            stream['packages'][package] = {'outcome': {'pass': 'ok', 'fail': 'failed', 'skip': 'skipped'}[action], 'seconds': event.get('Elapsed')}
            if action == 'fail' and package not in stream['failed']:
                stream['failed'].append(package)
                stream['failedBuild'][package] = event.get('FailedBuild', '')
        elif event.get('Test') and action == 'fail':
            stream['failing'][package].append(event['Test'])
    return stream

def first_failures(stream, name):
    # The first top-level test to fail, or the package's own output (its build's first) when no test failed.
    for package in stream['failed'] + [package for package in stream['failing'] if package not in stream['failed']]:
        tests = [test for test in stream['failing'].get(package, []) if '/' not in test]
        if tests:
            first, lines = tests[0], stream['output'][(package, tests[0])]
        else:
            first = '(package)'
            lines = stream['buildOutput'].get(stream['failedBuild'].get(package, ''), []) + stream['packageOutput'][package]
        shown = [line.rstrip()[:240] for line in lines if line.strip() and not line.startswith(('=== RUN', '=== PAUSE', '=== CONT'))]
        # Its first four lines and its last ten: a test's own log lines come first and its failure last, so a cut at the
        # front alone showed markdowninline's corpus log and never its cause (J-es4wbkq).
        kept = shown if len(shown) <= 14 else shown[:4] + ['... %d lines ...' % (len(shown) - 14)] + shown[-10:]
        print(json.dumps({'package': package, 'stream': name, 'firstFailure': first, 'lines': kept}))

tests = read(os.path.join(out, 'test.log'))
products = read(os.path.join(out, 'products.jsonl'))
plan_path = os.path.join(out, 'products-plan.json')
plan = json.load(open(plan_path)) if os.path.exists(plan_path) else {}
waves = [line.split('\t') for line in open(os.path.join(out, 'products-waves.tsv')).read().splitlines() if line]
packages = tests['packages']
counts = collections.Counter(row['outcome'] for row in packages.values())
print("This is the block's own tests run without loom, not the house's build of the tree. Nothing in it is a key, and nothing compares keys.")
print(json.dumps({'sha': sha, 'started': started, 'vetExit': int(vet), 'buildExit': int(build), 'testExit': int(test),
                  'goFlags': flags, 'packages': len(packages), 'ok': counts['ok'], 'failed': counts['failed'], 'skipped': counts['skipped'],
                  'outFreeBytes': int(out_free), 'outFull': int(out_free) < 64 * 1024 * 1024}))
# A vet or build that failed says why: its log's last twenty lines (J-es4wbkq: build exited 1 and the report said no more).
for step, code in (('vet', vet), ('build', build)):
    if int(code) != 0:
        logged = [line.rstrip()[:240] for line in open(os.path.join(out, step + '.log'), errors='replace') if line.strip()]
        print(json.dumps({'step': step, 'exit': int(code), 'tail': logged[-20:]}))
for package, row in packages.items():
    print(json.dumps({'package': package, 'outcome': row['outcome'], 'seconds': row['seconds'], 'failingTests': tests['failing'].get(package, [])}))
failed_waves = [int(wave) for wave, code in waves if code != '0']
print(json.dumps({'products': {
    'ok': int(plan_exit) == 0 and not failed_waves and not products['failed'] and not plan.get('problems') and not plan.get('unplaced'),
    'planExit': int(plan_exit), 'packages': plan.get('packages', 0), 'productPackages': plan.get('productPackages', 0),
    'tests': plan.get('tests', 0), 'waves': len(waves), 'failedWaves': failed_waves, 'failedPackages': products['failed'],
    'failedTests': {package: [test for test in failing if '/' not in test] for package, failing in products['failing'].items()},
    'unplaced': plan.get('unplaced', []), 'problems': plan.get('problems', [])}}))
first_failures(products, 'products')
first_failures(tests, 'tests')
print(f"::loom checked {len(packages)} packages at {sha[:12]}: {counts['ok']} ok, {counts['failed']} failed, {counts['skipped']} skipped")
PYTHON
echo "END block-report"
echo "block-report: raw streams kept in $out"
# The job's own answer is the tests': 0 only when vet, build and every package's tests passed, read from a whole stream.
# An out disk that filled may have cut a stream short, so a report with outFull never exits 0, and an image check can't
# pass on a short read (adamic-loom on #zwzer5f). A failure here is a finding to read beside the house, never a verdict
# on the block.
[ "$vetExit" = 0 ] && [ "$buildExit" = 0 ] && [ "$testExit" = 0 ] && [ "${outFree:-0}" -ge 67108864 ]

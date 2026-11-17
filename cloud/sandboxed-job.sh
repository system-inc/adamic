# cloud/sandboxed-job.sh: the preamble a sandboxed proof sources on a rented machine, right after the image's own
# environment, so each gap the sandboxed road shows is one small landing here rather than a line every mind relearns
# (Loom's suggestion, taken by Adamic, #t2fbrq3):
#
#   source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh && <command>
#
# It was cloud/sealed-job.sh until #sbxadamic, Loom's rename of the sealed mode to sandboxed (#sandboxed).
#
# Sourced, never run: it changes only the caller's environment, never its shell options, and returns nonzero on any
# step that fails, so the `&&` stops the job before its command. A rented machine runs the job in a laid tree with no
# .git, an empty HOME at /tmp/home and a read-only /opt/adamic-block, so:
#   - TMPDIR is made, since the job's tools write there first;
#   - the build cache (ADAMIC_BUILD_CACHE_DIR) and the TypeScript copy live outside the tree, in LOOM_JOB_SCRATCH where
#     the runner names one (the job's own directory on the tree's data disk) and in TMPDIR where it doesn't: a walker
#     that reads every file under $PWD never meets the kit's own files (internal/acheck read the cache's runtime.a as
#     programs, CLI's J-0pmqbfg; Adamic's rule on #qee4r6g), and internal/native's debug products refuse the home cache
#     (Apple's J-7me59b7 and J-8xkd55x);
#   - npm's logs go under it, since /opt/adamic-block/npmcache/_logs is read-only (WebAssembly);
#   - sandboxed_git gives the job a one-commit git of the tree as laid, for a mutant script that restores a file with
#     `git show HEAD:<file>` or reads `git status`. It costs about 27 s over adamic's tree (Loom's J-a5zgp8d), so only a
#     script that needs it calls it. HEAD has no history, no parent and no remote, so a job that reads `git log`, a
#     submodule's own repository or a remote still wants a house box. A tree the runner laid with a .git keeps it.
#   - GOMODCACHE is the node's laid cache when loom-node's cache proxy laid this very commit whole, and the image's
#     otherwise (#nodecache). The proxy's word is a lay file it writes only once every module the commit names is in
#     its cache. The node's cache is taken only when that file names this repository and commit, names the cache at
#     its one path, and lists the tree's go.sum and go.work.sum files exactly as laid here, each with the same sha256.
#     Anything else, a failed read included, keeps the image's cache, and one line on stderr says which and why.
#     GOPROXY, GOSUMDB and GOFLAGS stay the image's either way.
# Loom folds what's settled here into the image's environment.sh at its next build, so this file shrinks over time.

sandboxedJobFail() {
	echo "sandboxed-job: $*" >&2
	return 1
}

[ -n "${TMPDIR:-}" ] || export TMPDIR=/tmp/sandboxed-job
mkdir -p "$TMPDIR" || sandboxedJobFail "couldn't make TMPDIR $TMPDIR" || return 1
# Outside the tree, on the tree's disk where the runner names a place (adamic-loom on #1ndpe51: not the small private /tmp),
# else the job's TMPDIR, one rule with lay.sh's (Adamic on #qee4r6g). A scratch the runner put inside the tree is refused,
# never used: the cache would be back where a walker reads it.
sandboxedJobScratch="${LOOM_JOB_SCRATCH:-$TMPDIR}"
mkdir -p "$sandboxedJobScratch" || sandboxedJobFail "couldn't make the job's scratch $sandboxedJobScratch" || return 1
case "$(cd "$sandboxedJobScratch" && pwd -P)/" in
	"$(pwd -P)"/*) sandboxedJobFail "the job's scratch $sandboxedJobScratch is inside the tree $PWD" || return 1 ;;
esac
export ADAMIC_BUILD_CACHE_DIR="$sandboxedJobScratch/adamic-build"
mkdir -p "$ADAMIC_BUILD_CACHE_DIR" || sandboxedJobFail "couldn't make the build cache $ADAMIC_BUILD_CACHE_DIR" || return 1
export npm_config_logs_dir="$TMPDIR/npm-logs"
mkdir -p "$npm_config_logs_dir" || sandboxedJobFail "couldn't make npm's log directory $npm_config_logs_dir" || return 1
# A git identity for any commit the job makes itself (the meter commits its run; adamic-loom on #zwzer5f), since HOME is
# empty and git refuses a commit with no name. A house box's own identity is never overridden.
export GIT_AUTHOR_NAME="${GIT_AUTHOR_NAME:-sandboxed-job}" GIT_AUTHOR_EMAIL="${GIT_AUTHOR_EMAIL:-sandboxed-job@loom.invalid}"
export GIT_COMMITTER_NAME="${GIT_COMMITTER_NAME:-sandboxed-job}" GIT_COMMITTER_EMAIL="${GIT_COMMITTER_EMAIL:-sandboxed-job@loom.invalid}"

# The node's laid Go module cache, or the image's. The lay file is the proxy's contract with Loom (#nodecache): one JSON
# object, written atomically and only once the commit's modules are all laid, naming the repository, the commit, the
# cache (one exact path, never a path the file chooses, so a lay can't point a job at another directory) and every
# go.sum and go.work.sum in the commit's tree, submodules included, by path from the root with the sha256 of its bytes
# (both names, the proxy's owner on #nodecache: go checks a workspace's modules against its go.work.sum too). The tree's
# own are found by walking it from the root, skipping every .git, a submodule's included. A miss, a mismatch or
# a failed read keeps the image's GOMODCACHE: none of it ever fails the preamble. SANDBOXED_JOB_LAYS and
# SANDBOXED_JOB_LAID_GOMODCACHE are the test's stand-ins.
sandboxedJobLays="${SANDBOXED_JOB_LAYS:-/var/cache/loom-node-cache/go/lays}"
sandboxedJobLaidGoModuleCache="${SANDBOXED_JOB_LAID_GOMODCACHE:-/var/cache/loom-node-cache/go/mod}"
sandboxedJobGoModules=$(python3 -I - "$sandboxedJobLays" "$sandboxedJobLaidGoModuleCache" "${LOOM_JOB_REPOSITORY:-}" "${LOOM_JOB_COMMIT:-}" 2> /dev/null << 'PYTHON'
import hashlib, json, os, sys
lays, laid, repository, commit = sys.argv[1:5]
def image(why):
    print('image ' + why)
    sys.exit(0)
try:
    if not repository or not commit:
        image('the runner names no LOOM_JOB_REPOSITORY and LOOM_JOB_COMMIT')
    path = os.path.join(lays, repository, commit + '.json')
    if not os.path.exists(path):
        image('no lay for this commit')
    try:
        with open(path) as file:
            lay = json.load(file)
    except (OSError, ValueError) as error:
        image("the lay doesn't parse: %s" % error)
    if not isinstance(lay, dict):
        image('the lay is no object')
    if lay.get('repository') != repository:
        image('the lay names another repository')
    if lay.get('commit') != commit:
        image('the lay names another commit')
    if lay.get('gomodcache') != laid:
        image('the lay names the cache %r, not %s' % (lay.get('gomodcache'), laid))
    if not (os.path.isdir(laid) and os.access(laid, os.R_OK | os.X_OK)):
        image('the laid cache %s is no readable directory' % laid)
    sums = lay.get('goSums')
    if not isinstance(sums, list) or not all(
            isinstance(entry, dict) and isinstance(entry.get('path'), str) and isinstance(entry.get('sha256'), str)
            for entry in sums):
        image("the lay's goSums don't parse")
    named = {entry['path']: entry['sha256'].lower() for entry in sums}
    found = set()
    for directory, directories, files in os.walk('.'):
        directories[:] = [name for name in directories if name != '.git']
        for name in files:
            if name == 'go.sum':
                found.add(os.path.relpath(os.path.join(directory, name), '.'))
    for path in sorted(found ^ set(named)):
        image('%s is in the %s alone' % (path, 'tree' if path in found else 'lay'))
    for path in sorted(named):
        with open(path, 'rb') as file:
            if hashlib.sha256(file.read()).hexdigest() != named[path]:
                image('%s differs' % path)
    print('laid %s' % lay.get('modules'))
except Exception as error:
    image("the lay couldn't be checked: %s" % error)
PYTHON
) || sandboxedJobGoModules="image the lay couldn't be checked: python3 failed"
case "$sandboxedJobGoModules" in
	"laid "*)
		export GOMODCACHE="$sandboxedJobLaidGoModuleCache"
		echo "sandboxed-job: go modules: the node's laid cache for ${LOOM_JOB_COMMIT:0:12} (${sandboxedJobGoModules#laid } modules)" >&2
		;;
	*) echo "sandboxed-job: go modules: the image's cache (${sandboxedJobGoModules#image })" >&2 ;;
esac

# sandboxed_typescript_source: the image's TypeScript source becomes the job's own copy, in the job's scratch, and
# ADAMIC_TYPESCRIPT_SOURCE names it. The image's copy is root's and the job runs as its own user, so git calls it dubious
# and the corpus can't read its pin (J-hstk35n, J-6nnj4ys: internal/lower's census products). No config can name it
# safe for that reader: internal/buildcache's PinnedGit runs git with GIT_CONFIG_GLOBAL=/dev/null and no system config,
# and git ignores safe.directory given by -c or the environment. A copy the job owns is block-report's road
# (cloud/block-report/run.sh). sandboxed_products calls it; a command that reads the source some other way can too.
sandboxed_typescript_source() {
	local source="${ADAMIC_TYPESCRIPT_SOURCE:-}" copy="$sandboxedJobScratch/typescript-source" begun=$SECONDS
	[ -n "$source" ] || sandboxedJobFail "ADAMIC_TYPESCRIPT_SOURCE is unset: source the image's environment.sh first" || return 1
	[ "$source" != "$copy" ] || return 0
	[ ! -e "$copy" ] || sandboxedJobFail "$copy is already there, and isn't ADAMIC_TYPESCRIPT_SOURCE" || return 1
	cp -a "$source" "$copy" || sandboxedJobFail "copying $source to $copy failed" || return 1
	export ADAMIC_TYPESCRIPT_SOURCE="$copy"
	echo "sandboxed-job: the image's TypeScript source copied to $copy in $((SECONDS - begun)) s, the job's own"
}

# sandboxed_git: the tree as laid becomes one commit, HEAD, in a .git of the job's own, read back clean. Run from the tree's
# root. A tree that is already a repository (the runner lays sandboxed trees with a .git since Nov 16) is left as it is,
# its history never covered over, and sandboxed_git says so and returns 0, so an `&&` chain goes on to the command (Emit's
# finding, #6zvxhjq). A .git with no HEAD to read still refuses.
sandboxed_git() {
	[ -f go.work ] || [ -f go.mod ] || sandboxedJobFail "sandboxed_git runs from the tree's root" || return 1
	if [ -e .git ]; then
		local head
		head=$(git rev-parse --verify -q HEAD) || sandboxedJobFail "this tree has a .git with no HEAD to read" || return 1
		echo "sandboxed-job: the tree is already a repository at ${head:0:12}; sandboxed_git leaves its history as it is"
		return 0
	fi
	git init -q || sandboxedJobFail "git init failed" || return 1
	git add -A || sandboxedJobFail "git add failed" || return 1
	git -c user.name=sandboxed-job -c user.email=sandboxed-job@loom.invalid -c commit.gpgsign=false commit -q --no-verify \
		-m "the tree as laid" || sandboxedJobFail "git commit failed" || return 1
	[ -z "$(git status --porcelain)" ] || sandboxedJobFail "the tree doesn't read clean after its commit" || return 1
	echo "sandboxed-job: one commit $(git rev-parse --short HEAD) holds the tree as laid"
}

# sandboxed_npm [project...]: each npm project installed offline from the image's cache, as Loom's J-47wmdte proved:
# --offline, since the machine has no network, and --ignore-scripts, since an install script would fetch. With no project
# named, it takes the image's own list, BLOCK_REPORT_NPM_PROJECTS. Run from the tree's root.
sandboxed_npm() {
	local projects="${*:-${BLOCK_REPORT_NPM_PROJECTS:-}}" project count=0
	[ -n "$projects" ] || sandboxedJobFail "sandboxed_npm names no project and the image's BLOCK_REPORT_NPM_PROJECTS is empty" || return 1
	for project in $projects; do
		(cd "$project" && npm ci --offline --ignore-scripts --no-audit --no-fund) ||
			sandboxedJobFail "npm ci --offline failed in $project" || return 1
		count=$((count + 1))
	done
	echo "sandboxed-job: npm installed offline in $count projects"
}

# sandboxed_products <package>...: settles an adamic tree's products before a command reads them, as the lay kit's
# --products does on a house box (loom's contract/proof/lay.sh): a unit test reads only products built ahead and never
# builds one, and a sandboxed job's build cache starts empty. Each package (./<path> or github.com/system-inc/adamic/<path>)
# brings every package cloud/fast-gate/compiler-dependencies.json lists under its path, and one go test -run
# '^TestProduct_' runs over the union into ADAMIC_BUILD_CACHE_DIR. Nothing of the file is ever run: it's read as JSON.
# It runs with -timeout 0, so the job's own ceiling is the only clock: go's default 10 minutes killed internal/oracle's
# products before they settled on Lunatic's 8 cores (Library's J-1xphvcc, routed by Runtime on #jobs), where a 64-core
# rented machine settles them inside it.
sandboxed_products() {
	[ $# -gt 0 ] || sandboxedJobFail "sandboxed_products names no package" || return 1
	local packages
	packages=$(python3 -I - cloud/fast-gate/compiler-dependencies.json "$@" << 'PYTHON'
import json, os, re, sys
part = re.compile(r'^[A-Za-z0-9_][A-Za-z0-9_.-]*$')
def path_of(name):
    for prefix in ('./', 'github.com/system-inc/adamic/'):
        if name.startswith(prefix):
            name = name[len(prefix):]
    name = name.rstrip('/')
    if not name or not all(part.match(piece) for piece in name.split('/')):
        sys.exit('sandboxed-job: %r is no adamic package path' % name)
    return name
listed = {}
if os.path.exists(sys.argv[1]):
    listed = json.load(open(sys.argv[1])).get('packages')
    if not isinstance(listed, dict):
        sys.exit('sandboxed-job: %s holds no packages object' % sys.argv[1])
union = set()
for name in sys.argv[2:]:
    path = path_of(name)
    union.add(path)
    for dependency in listed.get(path, []):
        union.add(path_of(dependency))
print(' '.join('./' + path for path in sorted(union)))
PYTHON
	) || sandboxedJobFail "couldn't read the products to settle" || return 1
	sandboxed_typescript_source || return 1
	# The package list is held to plain paths above, so it splits as words here.
	# shellcheck disable=SC2086
	go test -count=1 -timeout 0 -run '^TestProduct_' $packages < /dev/null >&2 || sandboxedJobFail "the products didn't settle over $packages" || return 1
	echo "sandboxed-job: products settled for $(echo $packages | wc -w | tr -d ' ') packages"
}

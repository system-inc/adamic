# cloud/sealed-job.sh: the preamble a sealed proof sources on a rented machine, right after the image's own
# environment, so each gap the sealed road shows is one small landing here rather than a line every mind relearns
# (Loom's suggestion, taken by Adamic, #t2fbrq3):
#
#   source /opt/adamic-block/environment.sh && source cloud/sealed-job.sh && <command>
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
#   - sealed_git gives the job a one-commit git of the tree as laid, for a mutant script that restores a file with
#     `git show HEAD:<file>` or reads `git status`. It costs about 27 s over adamic's tree (Loom's J-a5zgp8d), so only a
#     script that needs it calls it. HEAD has no history, no parent and no remote, so a job that reads `git log`, a
#     submodule's own repository or a remote still wants a house box.
# Loom folds what's settled here into the image's environment.sh at its next build, so this file shrinks over time.

sealedJobFail() {
	echo "sealed-job: $*" >&2
	return 1
}

[ -n "${TMPDIR:-}" ] || export TMPDIR=/tmp/sealed-job
mkdir -p "$TMPDIR" || sealedJobFail "couldn't make TMPDIR $TMPDIR" || return 1
# Outside the tree, on the tree's disk where the runner names a place (adamic-loom on #1ndpe51: not the small private /tmp),
# else the job's TMPDIR, one rule with lay.sh's (Adamic on #qee4r6g). A scratch the runner put inside the tree is refused,
# never used: the cache would be back where a walker reads it.
sealedJobScratch="${LOOM_JOB_SCRATCH:-$TMPDIR}"
mkdir -p "$sealedJobScratch" || sealedJobFail "couldn't make the job's scratch $sealedJobScratch" || return 1
case "$(cd "$sealedJobScratch" && pwd -P)/" in
	"$(pwd -P)"/*) sealedJobFail "the job's scratch $sealedJobScratch is inside the tree $PWD" || return 1 ;;
esac
export ADAMIC_BUILD_CACHE_DIR="$sealedJobScratch/adamic-build"
mkdir -p "$ADAMIC_BUILD_CACHE_DIR" || sealedJobFail "couldn't make the build cache $ADAMIC_BUILD_CACHE_DIR" || return 1
export npm_config_logs_dir="$TMPDIR/npm-logs"
mkdir -p "$npm_config_logs_dir" || sealedJobFail "couldn't make npm's log directory $npm_config_logs_dir" || return 1
# A git identity for any commit the job makes itself (the meter commits its run; adamic-loom on #zwzer5f), since HOME is
# empty and git refuses a commit with no name. A house box's own identity is never overridden.
export GIT_AUTHOR_NAME="${GIT_AUTHOR_NAME:-sealed-job}" GIT_AUTHOR_EMAIL="${GIT_AUTHOR_EMAIL:-sealed-job@loom.invalid}"
export GIT_COMMITTER_NAME="${GIT_COMMITTER_NAME:-sealed-job}" GIT_COMMITTER_EMAIL="${GIT_COMMITTER_EMAIL:-sealed-job@loom.invalid}"

# sealed_typescript_source: the image's TypeScript source becomes the job's own copy, in the job's scratch, and
# ADAMIC_TYPESCRIPT_SOURCE names it. The image's copy is root's and the job runs as its own user, so git calls it dubious
# and the corpus can't read its pin (J-hstk35n, J-6nnj4ys: internal/lower's census products). No config can name it
# safe for that reader: internal/buildcache's PinnedGit runs git with GIT_CONFIG_GLOBAL=/dev/null and no system config,
# and git ignores safe.directory given by -c or the environment. A copy the job owns is block-report's road
# (cloud/block-report/run.sh). sealed_products calls it; a command that reads the source some other way can too.
sealed_typescript_source() {
	local source="${ADAMIC_TYPESCRIPT_SOURCE:-}" copy="$sealedJobScratch/typescript-source" begun=$SECONDS
	[ -n "$source" ] || sealedJobFail "ADAMIC_TYPESCRIPT_SOURCE is unset: source the image's environment.sh first" || return 1
	[ "$source" != "$copy" ] || return 0
	[ ! -e "$copy" ] || sealedJobFail "$copy is already there, and isn't ADAMIC_TYPESCRIPT_SOURCE" || return 1
	cp -a "$source" "$copy" || sealedJobFail "copying $source to $copy failed" || return 1
	export ADAMIC_TYPESCRIPT_SOURCE="$copy"
	echo "sealed-job: the image's TypeScript source copied to $copy in $((SECONDS - begun)) s, the job's own"
}

# sealed_git: the tree as laid becomes one commit, HEAD, in a .git of the job's own, read back clean. Run from the tree's
# root. It refuses a tree that already has a .git, so a house box's real history is never covered over.
sealed_git() {
	[ -f go.work ] || [ -f go.mod ] || sealedJobFail "sealed_git runs from the tree's root" || return 1
	[ ! -e .git ] || sealedJobFail "this tree already has a .git; sealed_git is for a laid tree with none" || return 1
	git init -q || sealedJobFail "git init failed" || return 1
	git add -A || sealedJobFail "git add failed" || return 1
	git -c user.name=sealed-job -c user.email=sealed-job@loom.invalid -c commit.gpgsign=false commit -q --no-verify \
		-m "the tree as laid" || sealedJobFail "git commit failed" || return 1
	[ -z "$(git status --porcelain)" ] || sealedJobFail "the tree doesn't read clean after its commit" || return 1
	echo "sealed-job: one commit $(git rev-parse --short HEAD) holds the tree as laid"
}

# sealed_npm [project...]: each npm project installed offline from the image's cache, as Loom's J-47wmdte proved:
# --offline, since the machine has no network, and --ignore-scripts, since an install script would fetch. With no project
# named, it takes the image's own list, BLOCK_REPORT_NPM_PROJECTS. Run from the tree's root.
sealed_npm() {
	local projects="${*:-${BLOCK_REPORT_NPM_PROJECTS:-}}" project count=0
	[ -n "$projects" ] || sealedJobFail "sealed_npm names no project and the image's BLOCK_REPORT_NPM_PROJECTS is empty" || return 1
	for project in $projects; do
		(cd "$project" && npm ci --offline --ignore-scripts --no-audit --no-fund) ||
			sealedJobFail "npm ci --offline failed in $project" || return 1
		count=$((count + 1))
	done
	echo "sealed-job: npm installed offline in $count projects"
}

# sealed_products <package>...: settles an adamic tree's products before a command reads them, as the lay kit's
# --products does on a house box (loom's contract/proof/lay.sh): a unit test reads only products built ahead and never
# builds one, and a sealed job's build cache starts empty. Each package (./<path> or github.com/system-inc/adamic/<path>)
# brings every package cloud/fast-gate/compiler-dependencies.json lists under its path, and one go test -run
# '^TestProduct_' runs over the union into ADAMIC_BUILD_CACHE_DIR. Nothing of the file is ever run: it's read as JSON.
# It runs with -timeout 0, so the job's own ceiling is the only clock: go's default 10 minutes killed internal/oracle's
# products before they settled on Lunatic's 8 cores (Library's J-1xphvcc, routed by Runtime on #jobs), where a 64-core
# rented machine settles them inside it.
sealed_products() {
	[ $# -gt 0 ] || sealedJobFail "sealed_products names no package" || return 1
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
        sys.exit('sealed-job: %r is no adamic package path' % name)
    return name
listed = {}
if os.path.exists(sys.argv[1]):
    listed = json.load(open(sys.argv[1])).get('packages')
    if not isinstance(listed, dict):
        sys.exit('sealed-job: %s holds no packages object' % sys.argv[1])
union = set()
for name in sys.argv[2:]:
    path = path_of(name)
    union.add(path)
    for dependency in listed.get(path, []):
        union.add(path_of(dependency))
print(' '.join('./' + path for path in sorted(union)))
PYTHON
	) || sealedJobFail "couldn't read the products to settle" || return 1
	sealed_typescript_source || return 1
	# The package list is held to plain paths above, so it splits as words here.
	# shellcheck disable=SC2086
	go test -count=1 -timeout 0 -run '^TestProduct_' $packages < /dev/null >&2 || sealedJobFail "the products didn't settle over $packages" || return 1
	echo "sealed-job: products settled for $(echo $packages | wc -w | tr -d ' ') packages"
}

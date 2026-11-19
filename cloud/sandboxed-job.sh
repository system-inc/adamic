# cloud/sandboxed-job.sh: the preamble a sandboxed proof sources on a rented machine, right after the image's own
# environment, so each gap the sandboxed road shows is one small landing here rather than a line every mind relearns
# (Loom's suggestion, taken by Adamic, #t2fbrq3):
#
#   source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh && <command>
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
#     script that needs it calls it. HEAD has no history, no parent and no remote, so a job that reads `git log` or a
#     remote still wants a house box. A tree the runner laid with a .git keeps it. A submodule's own repository, one
#     commit deep at its pin, is sandboxed_repositories'.
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
    # The cache names a commit's lay by the hex sha256 of its repository, a newline and its commit (loom-node's
    # WholeFileName, Loom's must on #nodecache), in a lays/ no job can list, so a job opens only its own commit's lay
    path = os.path.join(lays, hashlib.sha256((repository + '\n' + commit).encode()).hexdigest() + '.json')
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
            if name in ('go.sum', 'go.work.sum'):
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

# sandboxed_repositories [path...]: run from the tree's root before anything writes under a submodule, a sandboxed_npm
# of a project under cohere included, since a file added there writes a tree that isn't the pin's and refuses.
# Each submodule laid as a repository of its own holding exactly its pin, one commit deep, as the lay kit's
# --repositories lays one on a house box (loom's contract/proof/lay.sh, layRepository), for a tool that reads a
# submodule's directory as a repository: adamic's admission delta runs `git -C cohere worktree add --detach <path>
# <pin>` (cmd/adamic-admission-delta, checkoutPinnedFrom). A sandboxed job reaches no submodule's mirror, so the
# repository is built from what the job holds, the files laid at the pin and the runner's manifest of them,
# .tracked/<path>/{HEAD,commit,files} (adamic-loom's jobwrapper, writeManifest), and the pin is proved, never taken on
# the manifest's word (Operations on #52cxm6c):
#   - the pin is the gitlink at <path> in its parent's own index: the tree's .git, which the runner reads from the job's
#     commit, or the repository made here for the submodule it's inside. .tracked/<path>/HEAD must name it;
#   - .tracked/<path>/commit, `git cat-file commit`'s bytes, must hash as a commit to the pin, so its tree line is the
#     pin's tree;
#   - every file laid under <path>, with the gitlinks .tracked/<path>/files lists, must write that tree.
# A byte, a mode, a file more or fewer, or a commit or gitlink that differs breaks a hash and refuses, so a forged
# .tracked can't name another pin. The repository is made by `git init --template=`, so it holds no hook; its HEAD is
# the pin, detached, its .git/shallow names the pin when it has parents, since none is held, and git status reads it
# clean. Its .git/info/attributes unsets text, filter, ident and working-tree-encoding for every path, so each file is
# hashed as the bytes git archive laid: under cohere's `* text=auto eol=lf`, git add would turn the CRLF four of its
# blobs hold into LF, and the tree would never be the pin's. A checkout from it writes each blob's bytes as they are.
# With no path named, every submodule the tree's .gitmodules lists, and inside each, every gitlink its pin holds, depth
# first. A submodule that is already a repository at its pin is left as it is, as lay.sh leaves one, and one at another
# commit refuses. It hashes every file each submodule holds, about 77,000 over cohere's and TypeScript's pins, so only a
# script that reads a submodule's git calls it.
sandboxed_repositories() {
	[ -f go.work ] || [ -f go.mod ] || sandboxedJobFail "sandboxed_repositories runs from the tree's root" || return 1
	# laid and kept are counted by sandboxedJobRepository, which sees them as bash's locals of its caller
	local paths=("$@") key laid=0 kept=0 begun=$SECONDS path
	if [ ${#paths[@]} -eq 0 ]; then
		while IFS= read -r key; do
			paths+=("${key#* }")
		done < <(sandboxedJobGit config --no-includes --file .gitmodules --get-regexp '^submodule\..*\.path$' 2> /dev/null)
		[ ${#paths[@]} -gt 0 ] || sandboxedJobFail "sandboxed_repositories names no path and the tree's .gitmodules lists none" || return 1
	fi
	for path in "${paths[@]}"; do
		sandboxedJobRepository "${path%/}" || return 1
	done
	echo "sandboxed-job: submodules at their pins: $laid laid as repositories from their files, $kept already one, in $((SECONDS - begun)) s"
}

# sandboxedJobGit is git as sandboxed_repositories runs it, in lay.sh's pinned environment: no replace object, global
# or system config, and no GIT_DIR or kin of the caller's steering which repository it reads.
sandboxedJobGit() {
	env -u GIT_DIR -u GIT_WORK_TREE -u GIT_INDEX_FILE -u GIT_OBJECT_DIRECTORY -u GIT_COMMON_DIR \
		GIT_NO_REPLACE_OBJECTS=1 GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1 git "$@"
}

# sandboxedJobRepository <path>: the submodule at <path> under the tree's root laid as a repository at its pin, as
# sandboxed_repositories says, then each gitlink its index holds, inside it.
sandboxedJobRepository() {
	local path=$1 manifest=.tracked/$1 begun=$SECONDS fresh=no
	case "/$path/" in
		//* | */../* | */./* | *//*) sandboxedJobFail "$path is no submodule path inside the tree" || return 1 ;;
	esac
	[ -d "$path" ] && [ ! -L "$path" ] || sandboxedJobFail "$path is no directory of the tree's" || return 1
	# The pin: the parent's index's one entry at <path>, a gitlink at stage 0. Read from inside a submodule not yet laid
	# as a repository, the index is the tree's, which holds no entry under a gitlink, so its submodules come first.
	local base gitlink="" entry entries=() gitlinkEntry=$'^160000 ([0-9a-f]{40}) 0\t(.*)$'
	base=$(basename -- "$path")
	while IFS= read -r -d '' entry; do
		entries+=("$entry")
	done < <(sandboxedJobGit --literal-pathspecs -C "$(dirname -- "$path")" ls-files -s -z -- "$base" 2> /dev/null)
	if [ ${#entries[@]} -eq 1 ] && [[ ${entries[0]} =~ $gitlinkEntry ]] && [ "${BASH_REMATCH[2]}" = "$base" ]; then
		gitlink=${BASH_REMATCH[1]}
	fi
	[ -n "$gitlink" ] || sandboxedJobFail "$path is no gitlink in its parent's index, so there's no pin to prove it at" || return 1

	if [ -e "$path/.git" ]; then
		local current
		current=$(sandboxedJobGit -C "$path" rev-parse --verify -q 'HEAD^{commit}') ||
			sandboxedJobFail "$path has a .git with no HEAD to read" || return 1
		[ "$current" = "$gitlink" ] ||
			sandboxedJobFail "$path is its own repository at ${current:0:12}, not its pin ${gitlink:0:12}; left alone" || return 1
		echo "sandboxed-job: $path ${gitlink:0:12} is its own repository at the pin, left as it is" >&2
		kept=$((kept + 1))
	else
		[ -f "$manifest/HEAD" ] && [ -f "$manifest/commit" ] && [ -f "$manifest/files" ] ||
			sandboxedJobFail "$manifest doesn't hold the runner's HEAD, commit and files for $path" || return 1
		local named made tree built
		named=$(head -c 64 "$manifest/HEAD")
		[ "$named" = "$gitlink" ] ||
			sandboxedJobFail "$manifest/HEAD names ${named:0:12}, not $path's gitlink ${gitlink:0:12}" || return 1
		made=$(sandboxedJobGit hash-object -t commit --stdin < "$manifest/commit") ||
			sandboxedJobFail "$manifest/commit doesn't hash as a commit" || return 1
		[ "$made" = "$gitlink" ] ||
			sandboxedJobFail "$manifest/commit hashes to ${made:0:12}, not $path's gitlink ${gitlink:0:12}: it isn't the pin's commit" || return 1
		tree=$(sed -n '1s/^tree \([0-9a-f]\{40\}\)$/\1/p' "$manifest/commit")
		[ -n "$tree" ] || sandboxedJobFail "$manifest/commit names no tree on its first line" || return 1
		# From here a refusal leaves $path/.git with no HEAD, which a second call refuses too
		sandboxedJobGit init -q --template= --initial-branch=laying "$path" > /dev/null ||
			sandboxedJobFail "git init of $path failed" || return 1
		mkdir -p "$path/.git/info" && printf '* -text -filter -ident -working-tree-encoding\n' > "$path/.git/info/attributes" ||
			sandboxedJobFail "couldn't write $path/.git/info/attributes" || return 1
		sandboxedJobGit -C "$path" hash-object -t commit -w --stdin < "$manifest/commit" > /dev/null ||
			sandboxedJobFail "the pin's commit wouldn't be written in $path" || return 1
		# The gitlinks go in the index first, so git add takes each one's directory as a submodule's and adds none of
		# its files; then every file laid, whatever a .gitignore says (-f)
		while IFS= read -r -d '' entry; do
			case "$entry" in
				'160000 commit '*) printf '%s\0' "$entry" ;;
			esac
		done < "$manifest/files" | sandboxedJobGit -C "$path" update-index -z --index-info ||
			sandboxedJobFail "the gitlinks $manifest/files lists wouldn't go in $path's index" || return 1
		sandboxedJobGit -C "$path" add -A -f -- . || sandboxedJobFail "git add of the files laid under $path failed" || return 1
		built=$(sandboxedJobGit -C "$path" write-tree) || sandboxedJobFail "git write-tree in $path failed" || return 1
		[ "$built" = "$tree" ] ||
			sandboxedJobFail "the files laid under $path write the tree ${built:0:12}, not its pin's ${tree:0:12}: a file under it isn't the pin's" || return 1
		if sed '/^$/q' "$manifest/commit" | grep -q '^parent '; then
			printf '%s\n' "$gitlink" > "$path/.git/shallow" || sandboxedJobFail "couldn't write $path/.git/shallow" || return 1
		fi
		printf '%s\n' "$gitlink" > "$path/.git/HEAD" || sandboxedJobFail "couldn't write $path/.git/HEAD" || return 1
		[ "$(sandboxedJobGit -C "$path" rev-parse --verify -q 'HEAD^{commit}')" = "$gitlink" ] ||
			sandboxedJobFail "$path's HEAD doesn't read as its pin ${gitlink:0:12}" || return 1
		echo "sandboxed-job: $path ${gitlink:0:12} laid as a repository from its laid files, its pin proved, in $((SECONDS - begun)) s" >&2
		laid=$((laid + 1)) fresh=yes
	fi

	# Each gitlink in its index, read whole before any is laid; the paths are the index's own, relative to <path>
	local children=() child
	while IFS= read -r -d '' entry; do
		case "$entry" in
			160000\ *) [[ $entry =~ $gitlinkEntry ]] && children+=("${BASH_REMATCH[2]}") ;;
		esac
	done < <(sandboxedJobGit -C "$path" ls-files -s -z)
	for child in ${children[@]+"${children[@]}"}; do
		sandboxedJobRepository "$path/$child" || return 1
	done
	# One laid here reads clean once its own submodules are laid too, by a git status that ran whole; one left as it is
	# is its own
	[ "$fresh" = no ] && return 0
	local status
	status=$(sandboxedJobGit -C "$path" status --porcelain) && [ -z "$status" ] ||
		sandboxedJobFail "$path doesn't read clean as a repository at its pin ${gitlink:0:12}" || return 1
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
# brings every package cloud/fast-gate/compiler-dependencies.json lists under its path, and each package's product tests
# are listed with go test -list. cloud/fast-gate/product-order.json's waits put them in waves, and each wave runs a
# package at a time with -run '^(<test>|...)$' into ADAMIC_BUILD_CACHE_DIR; the first run that fails stops every run after
# it. One go test over the whole union ran its packages at once, so internal/worker's TestProduct_WorkerWasm_005 got there
# before TestProduct_WorkerWasmTransform had built what it reads.
# A wait on a test in a package outside the union brings that test and whatever it waits on in turn, though not the rest
# of its package's product tests (internal/oracle has over a hundred, and cmd/adamic's and internal/lower's wait on
# seven): stage1/cohere/estree's product tests wait on stage1/cohere/estree/oracles', which compiler-dependencies.json
# doesn't list under it. Here it parts from the kit, whose waits stop at the union. A wait on a test its package doesn't
# list is dropped, as the kit drops it, and waits that hold a cycle refuse. Nothing of either file is ever run: each is
# read as JSON, and a test's name is held to Go's identifier characters before it reaches -run.
# It runs with -timeout 0, so the job's own ceiling is the only clock: go's default 10 minutes killed internal/oracle's
# products before they settled on Lunatic's 8 cores (Library's J-1xphvcc, routed by Runtime on #jobs), where a 64-core
# rented machine settles them inside it.
# The kit runs its products under the gate's variables (ADAMIC_GATE_UNCACHED and the rest); this runs without them, as
# the job's command does, since products keyed under them might not be the ones the command's tests read.
sandboxed_products() {
	[ $# -gt 0 ] || sandboxedJobFail "sandboxed_products names no package" || return 1
	local union
	union=$(python3 -I - cloud/fast-gate/compiler-dependencies.json "$@" << 'PYTHON'
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
print(' '.join(sorted(union)))
PYTHON
	) || sandboxedJobFail "couldn't read the products to settle" || return 1
	sandboxed_typescript_source || return 1
	# The union's packages are listed first, then each package a wait reaches outside what's listed, until no wait does.
	# Every path is held to plain parts, by the reader above or the one below, so the lists split as words here.
	local unlisted="$union" listed="" package listing line tests=() plan productTest='^TestProduct_[A-Za-z0-9_]*$'
	while [ -n "$unlisted" ]; do
		for package in $unlisted; do
			listing=$(go test -timeout 0 -list '^TestProduct_' "./$package/" < /dev/null) || {
				printf '%s\n' "$listing" >&2
				sandboxedJobFail "go test -list '^TestProduct_' ./$package/ failed"
				return 1
			}
			# go's own lines (ok, no test files) are left: only a product test's name is taken.
			while IFS= read -r line; do
				case "$line" in
					TestProduct_*)
						[[ $line =~ $productTest ]] || sandboxedJobFail "./$package/ lists a product test named $(printf '%q' "$line")" || return 1
						tests+=("$package $line")
						;;
				esac
			done <<< "$listing"
			listed="$listed $package"
		done
		# The plan is either `list <path>` lines, the packages the waits reach that aren't listed yet, or one
		# `settle <waves> <tests>` line and a `wave <wave> <path> <expression>` line a package a wave, in order.
		plan=$(python3 -I - cloud/fast-gate/product-order.json "$union" "$listed" "${tests[@]}" << 'PYTHON'
import json, os, re, sys
order, union, listed, present = sys.argv[1], set(sys.argv[2].split()), set(sys.argv[3].split()), set(sys.argv[4:])
part = '[A-Za-z0-9_][A-Za-z0-9_.-]*'
shape = re.compile('^%s(/%s)* TestProduct_[A-Za-z0-9_]*$' % (part, part))
waits = {}
if os.path.exists(order):
    waits = json.load(open(order)).get('waits')
    if not isinstance(waits, dict):
        sys.exit('sandboxed-job: %s holds no waits object' % order)
def waits_of(test):
    named = waits.get(test, [])
    if not isinstance(named, list) or not all(isinstance(wait, str) and shape.match(wait) for wait in named):
        sys.exit('sandboxed-job: %s waits[%s] is no list of <path> TestProduct_<name>' % (order, json.dumps(test)))
    return named
# Every product test of the union, then each test one of them waits on, however far. A wait into a package not yet
# listed asks for its listing; a wait on a test its package doesn't list, or into a directory the tree lacks, is dropped.
chosen = set(test for test in present if test.split(' ')[0] in union)
reaching = sorted(chosen)
unlisted = set()
while reaching:
    for wait in waits_of(reaching.pop()):
        package = wait.split(' ')[0]
        if package not in listed:
            if os.path.isdir(package):
                unlisted.add(package)
        elif wait in present and wait not in chosen:
            chosen.add(wait)
            reaching.append(wait)
if unlisted:
    for package in sorted(unlisted):
        print('list ' + package)
    sys.exit(0)
needs = {test: set(wait for wait in waits_of(test) if wait in chosen) for test in chosen}
placed, left, waves = set(), sorted(chosen), []
while left:
    wave = [test for test in left if needs[test] <= placed]
    if not wave:
        sys.exit('sandboxed-job: %s waits hold a cycle, %d tests never free to run (%s)' % (order, len(left), ', '.join(left[:6]) + (', ...' if len(left) > 6 else '')))
    byPackage = {}
    for test in wave:
        package, name = test.split(' ', 1)
        byPackage.setdefault(package, []).append(name)
    waves.append(byPackage)
    placed.update(wave)
    left = [test for test in left if test not in placed]
print('settle %d %d' % (len(waves), len(chosen)))
for number, byPackage in enumerate(waves, 1):
    for package in sorted(byPackage):
        print('wave %d %s ^(%s)$' % (number, package, '|'.join(sorted(byPackage[package]))))
PYTHON
		) || sandboxedJobFail "couldn't put the products in waves" || return 1
		unlisted=$(sed -n 's/^list //p' <<< "$plan")
	done
	local kind first second third waves=0 settled=0
	while read -r kind first second third; do
		case "$kind" in
			settle) waves=$first settled=$second ;;
			wave)
				echo "sandboxed-job: products: wave $first of $waves: go test -count=1 -timeout 0 -run '$third' ./$second/" >&2
				go test -count=1 -timeout 0 -run "$third" "./$second/" < /dev/null >&2 ||
					sandboxedJobFail "the products didn't settle: wave $first of $waves failed in ./$second/, and nothing after it ran" || return 1
				;;
		esac
	done <<< "$plan"
	# shellcheck disable=SC2086
	echo "sandboxed-job: products settled for $(echo $listed | wc -w | tr -d ' ') packages in $waves waves, $settled product tests"
}

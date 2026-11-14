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
#   - the build cache lives under it (ADAMIC_BUILD_CACHE_DIR): internal/native's debug products refuse the home
#     cache (Apple's J-7me59b7 and J-8xkd55x);
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
export ADAMIC_BUILD_CACHE_DIR="$TMPDIR/buildcache"
mkdir -p "$ADAMIC_BUILD_CACHE_DIR" || sealedJobFail "couldn't make the build cache $ADAMIC_BUILD_CACHE_DIR" || return 1
export npm_config_logs_dir="$TMPDIR/npm-logs"
mkdir -p "$npm_config_logs_dir" || sealedJobFail "couldn't make npm's log directory $npm_config_logs_dir" || return 1

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

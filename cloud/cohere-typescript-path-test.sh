#!/usr/bin/env bash
# cloud/cohere-typescript-path-test.sh holds cloud/cohere-typescript-path.sh's cohere_typescript_path to its contract,
# over fixture trees it builds in a directory of its own: each case is a cohere/.gitmodules and the directories beside
# it, and the answer it must give, a path on stdout and exit 0, or one FAIL line on stderr, nothing on stdout and exit 1.
#
#   bash cloud/cohere-typescript-path-test.sh
#
# The cases: cohere's shape at its pin as of Nov 23 (TypeScript), the shape after cohere moves it
# (libraries/typescript/TypeScript), that move with the entry renamed and a decoy named TypeScript whose url names
# another repository (the entry is chosen by its url, never its name), a url with no .git, the tree's root as the
# default ., and the refusals: no .gitmodules, no entry whose url names TypeScript, two, a path with `..`, an absolute
# path, and a path that is no directory in the tree.
#
# Mutant, run here after the real helper: cohere_typescript_path printing the hard-coded TypeScript, as the scripts read
# it before. The moved-shape case must go red under it, and the test fails if it doesn't, so a suite that passes a
# hard-coded path can't pass itself.
set -u

helper="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)/cohere-typescript-path.sh"
[ -f "$helper" ] || { echo "FAIL: no helper at $helper"; exit 1; }
work=$(mktemp -d "${TMPDIR:-/tmp}/cohere-typescript-path-test.XXXXXX") || { echo "FAIL: couldn't make a scratch directory"; exit 1; }
testEnd() {
	# Only this run's own mktemp directory is ever removed whole.
	case "$work" in
		"${TMPDIR:-/tmp}"/cohere-typescript-path-test.?*) rm -rf -- "$work" ;;
	esac
}
trap testEnd EXIT

# fixture <case> <gitmodules or -> <directory under cohere>...: a tree at $work/<case> with cohere/.gitmodules holding
# the text given (none for -) and each directory named made.
fixture() {
	local root="$work/$1" gitmodules=$2 directory
	shift 2
	mkdir -p "$root/cohere"
	[ "$gitmodules" = - ] || printf '%s' "$gitmodules" > "$root/cohere/.gitmodules"
	for directory in "$@"; do
		mkdir -p "$root/cohere/$directory"
	done
}

typescriptUrl=https://github.com/system-inc/TypeScript.git
fixture today "[submodule \"TypeScript\"]
	path = TypeScript
	url = $typescriptUrl
	branch = adamic
" TypeScript
fixture moved "[submodule \"TypeScript\"]
	path = libraries/typescript/TypeScript
	url = $typescriptUrl
	branch = adamic
" libraries/typescript/TypeScript
fixture renamed "[submodule \"TypeScript\"]
	path = vendor/other
	url = https://github.com/system-inc/other.git
[submodule \"libraries/typescript/TypeScript\"]
	path = libraries/typescript/TypeScript
	url = $typescriptUrl
" vendor/other libraries/typescript/TypeScript
fixture bare "[submodule \"tsc\"]
	path = libraries/typescript/TypeScript
	url = /var/lib/loom/mirrors/typescript/
" libraries/typescript/TypeScript
fixture missing - TypeScript
fixture none "[submodule \"other\"]
	path = other
	url = https://github.com/system-inc/other.git
" other
fixture two "[submodule \"TypeScript\"]
	path = TypeScript
	url = $typescriptUrl
[submodule \"moved\"]
	path = libraries/typescript/TypeScript
	url = $typescriptUrl
" TypeScript libraries/typescript/TypeScript
fixture parent "[submodule \"TypeScript\"]
	path = ../TypeScript
	url = $typescriptUrl
" TypeScript
mkdir -p "$work/parent/TypeScript"
fixture absolute "[submodule \"TypeScript\"]
	path = $work/absolute/cohere/TypeScript
	url = $typescriptUrl
" TypeScript
fixture absent "[submodule \"TypeScript\"]
	path = libraries/typescript/TypeScript
	url = $typescriptUrl
" TypeScript

failures=0
# check <case> <expected path, or FAIL> <text the FAIL line must hold, or ->: the helper run over the case's tree,
# its root named except for the default case, which runs from inside the tree with none. One line says how it went.
check() {
	local name=$1 expected=$2 holds=$3 root="$work/$1" out err code=0
	out=$(cd "$root" && . "$helper" && cohere_typescript_path_under_test "$root" 2> "$work/$name.err") || code=$?
	err=$(cat "$work/$name.err")
	if [ "$expected" = FAIL ]; then
		if [ "$code" = 1 ] && [ -z "$out" ] && [ "$(printf '%s\n' "$err" | wc -l | tr -d ' ')" = 1 ] &&
			[[ $err == "FAIL cohere_typescript_path: "* ]] && [[ $err == *"$holds"* ]]; then
			echo "ok   $name: refused with: $err"
			return 0
		fi
	elif [ "$code" = 0 ] && [ "$out" = "$expected" ] && [ -z "$err" ]; then
		echo "ok   $name: $out"
		return 0
	fi
	echo "red  $name: wanted ${expected}${holds:+ holding '$holds'}, got exit $code, stdout '$out', stderr '$err'"
	failures=$((failures + 1))
	return 1
}

# The default case runs from inside the tree with no root named.
cases() {
	check today TypeScript ''
	check moved libraries/typescript/TypeScript ''
	check renamed libraries/typescript/TypeScript ''
	check bare libraries/typescript/TypeScript ''
	check missing FAIL "$work/missing/cohere/.gitmodules isn't there"
	check none FAIL 'has 0 submodules whose url names TypeScript, not one (it read: other https://github.com/system-inc/other.git)'
	check two FAIL 'has 2 submodules whose url names TypeScript'
	check parent FAIL "the path '../TypeScript', which is no relative path inside cohere"
	check absolute FAIL "the path '$work/absolute/cohere/TypeScript', which is no relative path inside cohere"
	check absent FAIL "$work/absent/cohere/libraries/typescript/TypeScript is no directory of the tree's"
	local out code=0
	out=$(cd "$work/moved" && . "$helper" && cohere_typescript_path_under_test 2>&1) || code=$?
	if [ "$code" = 0 ] && [ "$out" = libraries/typescript/TypeScript ]; then
		echo "ok   default root: $out"
	else
		echo "red  default root: wanted libraries/typescript/TypeScript from inside the moved tree, got exit $code, '$out'"
		failures=$((failures + 1))
	fi
}

echo "== the helper"
cohere_typescript_path_under_test() { cohere_typescript_path "$@"; }
cases
helperFailures=$failures

echo "== the mutant: cohere_typescript_path prints the hard-coded TypeScript"
failures=0
cohere_typescript_path_under_test() { printf 'TypeScript\n'; }
cases > "$work/mutant.log"
mutantFailures=$failures
cat "$work/mutant.log"
grep -q '^red  moved:' "$work/mutant.log"
mutantMovedRed=$?

[ "$helperFailures" = 0 ] || { echo "FAIL: the helper went red in $helperFailures cases"; exit 1; }
[ "$mutantMovedRed" = 0 ] || { echo "FAIL: the mutant passed the moved-shape case, so the test can't tell a hard-coded path"; exit 1; }
echo "the helper passes every case; the mutant goes red in $mutantFailures, the moved shape among them"

#!/usr/bin/env bash
# run.sh: run a directory of .a and .ts probes on Node and on both of Adamic's backends, and print one
# table saying, for each probe and each build, whether each backend agrees with Node, differs from it, or
# stopped at compile time and why (#y1p4qzt).
#
#	cloud/probe/run.sh <probe directory> [--against <git ref>] [--logs <directory>]
#	cloud/probe/run.sh --check-control
#
# It builds Adamic from the checkout it sits in (the head), and, with --against, from that ref too, in a
# worktree of its own. Each probe runs on Node through oracle/node.mjs, which strips the source's types
# and never runs anything Adamic lowered, then through the JavaScript backend with 'adamic' resolved to
# oracle/adamic.mjs, then natively under ASan and UBSan (build --sanitize). A backend agrees when its
# stdout and exit code are Node's, byte for byte. Go and clang are needed where it runs (a box or Codex),
# never on the Mac. The table goes to stdout, every probe's full output to the logs directory.
#
# --check-control runs cloud/probe/control and fails unless agrees.ts agrees everywhere and differs.a,
# Kirk's Oct 5 difference (a write past an array's end panics where Node grows the array), differs
# everywhere. That is the runner's own proof: a runner that called everything agreement would pass any
# probe.
set -uo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$root" || exit 2

probes="" against="" logs="" check=false
while [ $# -gt 0 ]; do
	case "$1" in
	--against) against="${2:?--against takes a git ref}"; shift 2 ;;
	--logs) logs="${2:?--logs takes a directory}"; shift 2 ;;
	--check-control) check=true; probes="$root/cloud/probe/control"; shift ;;
	-*) echo "run.sh: unknown flag $1" >&2; exit 2 ;;
	*) probes=$(cd "$1" && pwd) || exit 2; shift ;;
	esac
done
if [ -z "$probes" ]; then
	echo "usage: cloud/probe/run.sh <probe directory> [--against <git ref>] [--logs <directory>] | --check-control" >&2
	exit 2
fi

work=$(mktemp -d "${TMPDIR:-/tmp}/adamic-probe.XXXXXX")
logs=${logs:-$work/logs}
mkdir -p "$logs"
# The Go build wants a writable GOTMPDIR, which a box's sandbox may not give by default.
mkdir -p "${GOTMPDIR:-/tmp/adamic-gate}" 2>/dev/null || export GOTMPDIR=$work/gotmp
mkdir -p "${GOTMPDIR:-/tmp/adamic-gate}"

# build checks out the submodules if they're missing and builds the compiler at a checkout.
build() {
	local checkout=$1 binary=$2
	if [ ! -e "$checkout/cohere/TypeScript/go.mod" ]; then
		git -C "$checkout" -c protocol.file.allow=always submodule update --init --recursive --depth 1 >"$logs/submodules.log" 2>&1 || return 1
	fi
	(cd "$checkout" && go build -o "$binary" ./cmd/adamic) >"$logs/build-$(basename "$binary").log" 2>&1
}

builds=(head)
heads=("$(git rev-parse HEAD)")
if ! build "$root" "$work/head"; then
	echo "run.sh: the head didn't build; see $logs/build-head.log" >&2
	exit 2
fi
if [ -n "$against" ]; then
	# A Codex checkout may have no origin remote, so a ref that doesn't resolve is fetched from GitHub.
	commit=$(git rev-parse --verify -q "$against^{commit}")
	if [ -z "$commit" ]; then
		git fetch -q https://github.com/system-inc/adamic.git "${against#origin/}" || exit 2
		commit=$(git rev-parse FETCH_HEAD)
	fi
	git worktree add --detach -q "$work/against-tree" "$commit" || exit 2
	if ! build "$work/against-tree" "$work/against"; then
		echo "run.sh: $against didn't build; see $logs/build-against.log" >&2
		exit 2
	fi
	builds+=(against)
	heads+=("$(git -C "$work/against-tree" rev-parse HEAD)")
fi

# The JavaScript backend's output imports 'adamic', which resolves to the runtime of the checkout that
# built it.
hook() {
	local checkout=$1
	printf "import { registerHooks } from 'node:module';\nregisterHooks({ resolve(specifier, context, next) { return specifier === 'adamic' ? { url: %s, shortCircuit: true } : next(specifier, context); } });\n" \
		"'file://$checkout/oracle/adamic.mjs'"
}
hook "$root" >"$work/hook-head.mjs"
[ -n "$against" ] && hook "$work/against-tree" >"$work/hook-against.mjs"

# cell prints a table cell: its text on one line, pipes escaped, cut at 80 characters.
cell() {
	tr '\n' ' ' <"$1" | sed -e 's/|/\\|/g' -e 's/ *$//' | cut -c1-80
}

echo "Probes in ${probes#"$root"/}, on $(node --version) and $(clang --version 2>/dev/null | head -1)."
for index in "${!builds[@]}"; do
	echo "- ${builds[$index]}: ${heads[$index]}"
done
echo
echo "| probe | Node | build | JavaScript | native |"
echo "|---|---|---|---|---|"

failed=false
shopt -s nullglob
for file in "$probes"/*.a "$probes"/*.ts; do
	name=$(basename "$file")
	out="$logs/$name"
	mkdir -p "$out"
	timeout 60 node --disable-warning=ExperimentalWarning oracle/node.mjs "$file" >"$out/node.stdout" 2>"$out/node.stderr"
	node_exit=$?
	node_cell="$(cell "$out/node.stdout") (exit $node_exit)"
	for build_name in "${builds[@]}"; do
		binary="$work/$build_name"
		# JavaScript: compile, then run beside the runtime.
		if "$binary" js "$file" >"$out/$build_name.mjs" 2>"$out/$build_name.js-compile.stderr"; then
			timeout 60 node --import "$work/hook-$build_name.mjs" "$out/$build_name.mjs" >"$out/$build_name.js.stdout" 2>"$out/$build_name.js.stderr"
			js_exit=$?
			if cmp -s "$out/$build_name.js.stdout" "$out/node.stdout" && [ "$js_exit" = "$node_exit" ]; then
				js="agrees"
			else
				js="differs: $(cell "$out/$build_name.js.stdout") (exit $js_exit)"
				said=$(head -1 "$out/$build_name.js.stderr" | sed -e 's/|/\\|/g' | cut -c1-80)
				[ -n "$said" ] && js="$js, $said"
			fi
		else
			js="stops: $(head -1 "$out/$build_name.js-compile.stderr" | sed -e "s|$probes/||" -e 's/|/\\|/g' | cut -c1-120)"
		fi
		# Native, under ASan and UBSan.
		if "$binary" build "$file" -o "$out/$build_name.bin" --sanitize >"$out/$build_name.native-compile.stdout" 2>"$out/$build_name.native-compile.stderr"; then
			timeout 60 "$out/$build_name.bin" >"$out/$build_name.native.stdout" 2>"$out/$build_name.native.stderr"
			native_exit=$?
			if cmp -s "$out/$build_name.native.stdout" "$out/node.stdout" && [ "$native_exit" = "$node_exit" ]; then
				native="agrees"
			else
				native="differs: $(cell "$out/$build_name.native.stdout") (exit $native_exit)"
				# The sanitizer's report when there is one, else the first thing the binary said.
				said=$(grep -m1 -E 'ERROR: (Address|Undefined|Leak)Sanitizer|runtime error:' "$out/$build_name.native.stderr")
				[ -z "$said" ] && said=$(head -1 "$out/$build_name.native.stderr")
				said=$(printf '%s' "$said" | sed -e 's/|/\\|/g' | cut -c1-80)
				[ -n "$said" ] && native="$native, $said"
			fi
		else
			native="stops: $(head -1 "$out/$build_name.native-compile.stderr" | sed -e "s|$probes/||" -e 's/|/\\|/g' | cut -c1-120)"
		fi
		echo "| $name | $node_cell | $build_name | $js | $native |"
		if $check; then
			case "$name" in
			agrees.*) [ "$js" = agrees ] && [ "$native" = agrees ] || failed=true ;;
			differs.*) [ "${js%%:*}" = differs ] && [ "${native%%:*}" = differs ] || failed=true ;;
			esac
		fi
	done
done

echo
echo "Full output: $logs"
[ -n "$against" ] && git worktree remove --force "$work/against-tree" >/dev/null 2>&1
if $check && $failed; then
	echo "run.sh: the control didn't read as agrees.ts agreeing and differs.a differing" >&2
	exit 1
fi
exit 0

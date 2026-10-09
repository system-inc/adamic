#!/usr/bin/env bash
# cloud/gocacheprog.sh makes this machine's Go actions shareable through the shared build cache, and nothing more: the
# pinned Go (go-pin.sh), the flags that keep paths and commits out of every action (-trimpath -buildvcs=false), and
# cmd/adamic-gocacheprog built alone and named in env.sh. It never warms anything; the cache program fills Go's cache
# on demand. setup.sh runs it, and so can a machine set up before it (Oct 9: instances set up before the cache program
# landed at 17:21Z kept an env.sh without it, and re-running all of setup inside a unit warmed the whole module, filled
# a disk and broke canary 9b).
#
#   bash cloud/gocacheprog.sh <tools directory>     then: source <tools directory>/env.sh
#
# Its block in env.sh is keyed by this script, the pin and the program's source, so a second run with none of them
# changed returns at once, and a change to any of them rebuilds. Credentials stay in the caller's environment: this
# never grants cache write trust.
set -euo pipefail

[ $# = 1 ] || { echo "usage: bash cloud/gocacheprog.sh <tools directory>" >&2; exit 2; }
tools=$1
repository=$(cd "$(dirname "$0")/.." && pwd)
environment=$tools/env.sh
program=$tools/bin/adamic-gocacheprog
[ -f "$environment" ] || { echo "gocacheprog: $environment is missing; run cloud/setup.sh first" >&2; exit 2; }
# shellcheck source=cloud/go-pin.sh
source "$repository/cloud/go-pin.sh"
freeMegabytes() { df -Pm "$tools" | awk 'NR == 2 {print $4}'; }
freeBefore=$(freeMegabytes)

key=$({
	cat "$repository/cloud/go-pin.sh" "$repository/cloud/gocacheprog.sh" "$repository"/cmd/adamic-gocacheprog/*.go 2> /dev/null || true
	printf '%s\n%s\n' "$tools" "$repository"
} | sha256sum | cut -c1-64)
if grep -qx "# gocacheprog $key on" "$environment" && [ -x "$program" ]; then
	echo "gocacheprog: shared cache ready (unchanged, $goPin)"
	exit 0
fi

state=on reason=""
case $(uname -m) in
	x86_64) goArchitecture=amd64 ;;
	aarch64 | arm64) goArchitecture=arm64 ;;
	*) goArchitecture="" ;;
esac
if [ "${ADAMIC_GOCACHE_OFF:-0}" = 1 ]; then
	state=off reason="ADAMIC_GOCACHE_OFF=1"
elif [ ! -d "$repository/cmd/adamic-gocacheprog" ]; then
	state=off reason="cmd/adamic-gocacheprog absent"
elif [ -z "$goArchitecture" ] || ! installPinnedGo "$tools" "$goArchitecture" > /dev/null 2>&1; then
	# A split toolchain must never read or write the shared cache.
	state=off reason="go is $(GOTOOLCHAIN=local "$tools/go/bin/go" env GOVERSION 2> /dev/null || echo absent), the pin is $goPin"
elif ! (cd "$repository" && env -u GOCACHEPROG GOTOOLCHAIN=local "$tools/go/bin/go" build -o "$program.new" ./cmd/adamic-gocacheprog) > "$tools/gocacheprog.log" 2>&1; then
	state=off reason="adamic-gocacheprog build failed; see $tools/gocacheprog.log"
else
	mv "$program.new" "$program"
fi

# Everything between the markers is this script's, and so are the lines an older setup.sh wrote for the same job; the
# rest of env.sh stays as setup wrote it.
staging=$(mktemp "$environment.XXXXXX")
awk '
	/^# gocacheprog [0-9a-f]+ (on|off)$/ { skipping = 1; next }
	skipping { if ($0 == "# gocacheprog end") skipping = 0; next }
	/^unset GOCACHEPROG$/ { next }
	/^if \[ "\$\{ADAMIC_GOCACHE_OFF:-0\}" != 1 \]/ { legacy = 1; next }
	legacy { if ($0 == "fi") legacy = 0; next }
	/^# A product.s bytes are a function of its key/ || /^# and the checkout path, inputs no key sees/ { next }
	/^case " \$GOFLAGS " in/ { next }
	{ print }
' "$environment" > "$staging"
{
	echo "# gocacheprog $key $state"
	[ -n "$reason" ] && echo "# shared cache off: $reason"
	# shellcheck disable=SC2016
	printf 'case ":$PATH:" in *:%q:*) ;; *) export PATH=%q:"$PATH" ;; esac\n' "$tools/go/bin" "$tools/go/bin"
	# A product's bytes are a function of its key (@system_adamic, Oct 9): without these, every binary stamps the
	# commit and the checkout path, inputs no key sees, and no two trees share an action.
	# shellcheck disable=SC2016
	echo 'case " $GOFLAGS " in *" -buildvcs=false -trimpath "*) ;; *) export GOFLAGS="${GOFLAGS:+$GOFLAGS }-buildvcs=false -trimpath" ;; esac'
	echo 'unset GOCACHEPROG'
	if [ "$state" = on ]; then
		printf 'if [ "${ADAMIC_GOCACHE_OFF:-0}" != 1 ] && [ -d %q ] && [ -x %q ]; then\n\texport GOCACHEPROG=%q\nfi\n' \
			"$repository/cmd/adamic-gocacheprog" "$program" "$program"
	fi
	echo "# gocacheprog end"
} >> "$staging"
mv "$staging" "$environment"

if [ "$state" = on ]; then
	echo "gocacheprog: shared cache ready ($goPin, free $freeBefore to $(freeMegabytes) MB)"
else
	echo "gocacheprog: shared cache off ($reason)"
fi

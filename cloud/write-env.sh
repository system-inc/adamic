#!/bin/bash
# Writes setup.sh's env.sh, the one file every shell sources (the agent's shell in Codex is a different session from
# setup's). Arguments: tools gate markdownDependencies repository gocacheprog(on|off) wasiDirectory(or empty).
#
# setup.sh owns its own lines and rewrites them every run, but an env.sh also holds lines another step wrote: a gate
# box's 20 gate-input exports (ADAMIC_TYPESCRIPT_SOURCE, the cycle ledger, the css, library and prettier paths,
# ADAMIC_CLANG_TSGO_ARCHIVE). A full rewrite dropped them on Server at 22:15Z Oct 9 and killed its canaries under
# set -u (#nwsyj7z), so every ADAMIC_* export setup.sh doesn't write itself is kept, after setup's own lines.
set -euo pipefail
tools=$1 gate=$2 markdownDependencies=$3 repository=$4 gocacheprog=$5 wasiDirectory=$6
environment="$tools/env.sh"

kept=""
if [ -f "$environment" ]; then
	kept=$(grep -E '^export ADAMIC_[A-Z0-9_]+=' "$environment" | grep -vE '^export ADAMIC_MARKDOWNWIDTH_DEPS=' || true)
fi

next="$environment.writing"
cat > "$next" << ENV
export PATH="$tools/bin:$([ -x "$tools/go/bin/go" ] && echo "$tools/go/bin:")\$PATH"
export GOTOOLCHAIN=auto
export GOPROXY="https://proxy.golang.org|direct"
export TMPDIR=$gate
export ADAMIC_MARKDOWNWIDTH_DEPS="$markdownDependencies"
# A product's bytes are a function of its key (@system_adamic, Oct 9): without these, every binary stamps the commit
# and the checkout path, inputs no key sees. Added to whatever GOFLAGS holds, once, however often this is sourced.
case " \${GOFLAGS:-} " in *" -buildvcs=false -trimpath "*) ;; *) export GOFLAGS="\${GOFLAGS:+\$GOFLAGS }-buildvcs=false -trimpath" ;; esac
ENV
# Persist only a successful bootstrap; recheck opt-out and branch presence in later shells.
# Credentials stay in the caller's environment: cloud setup never grants cache write trust.
printf 'unset GOCACHEPROG\n' >> "$next"
if [ "$gocacheprog" = on ]; then
	printf 'if [ "${ADAMIC_GOCACHE_OFF:-0}" != 1 ] && [ -d %q ] && [ -x %q ]; then\n\texport GOCACHEPROG=%q\nfi\n' \
		"$repository/cmd/adamic-gocacheprog" "$tools/bin/adamic-gocacheprog" "$tools/bin/adamic-gocacheprog" >> "$next"
fi
if [ -n "$wasiDirectory" ]; then
	printf 'export WASI_SYSROOT=%q\n' "$wasiDirectory/share/wasi-sysroot" >> "$next"
fi
if [ -n "$kept" ]; then
	printf '%s\n' "$kept" >> "$next"
fi
mv "$next" "$environment"

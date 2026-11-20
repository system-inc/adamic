#!/bin/bash
# #m52sper on Golem: the Benchmarks Game's Go and Rust entries against Adamic native and Node, from the
# toolchains and offline sets Machines laid at /srv/loom-tools, each named by its sha256. The first
# argument is -samples. Builds rust-deps' rlibs offline from the vendored crates, with pcre2-sys linked
# against the laid libpcre2 (pkg-config with the laid prefix as its sysroot), points Go at the laid
# module cache through a file proxy, then runs `go run ./bench -game`. Never lands.
set -u
samples=${1:-1}
tools=/srv/loom-tools
lock=43c9e2341c8abef912574aedb33b75ea2f5349489e4a5d71e5a63938edf38f12
sum=0c6479bbb1d4a1d11622310552db0b5cc5176978d8d6c936134f76f7b7c67c06
vendor=$tools/cargo-vendor/$lock
modules=$tools/gomod-bench/$sum
T=$(mktemp -d)
export PATH=$tools/pkgconf/bin:$PATH
trap 'rm -rf "$T"' EXIT
echo "box $(hostname) $(nproc) cpus, $(uname -r), tree $(git rev-parse HEAD)"
echo "perf_event_paranoid $(cat /proc/sys/kernel/perf_event_paranoid)"
[ "$(sha256sum < bench/game/rust-deps/Cargo.lock | cut -c1-64)" = $lock ] && [ "$(sha256sum < bench/game/go.sum | cut -c1-64)" = $sum ] && echo "PASS the tree's Cargo.lock and go.sum are the laid sets'" || { echo "FAIL the tree's lockfiles aren't the laid sets'"; exit 1; }
[ -d "$vendor" ] && [ -d "$modules/cache/download" ] && echo "PASS the offline sets are laid" || { echo "FAIL missing $vendor or $modules/cache/download"; exit 1; }
"$tools/go1.23.1/bin/go" version && "$tools/rust-1.84.1/bin/rustc" --version && echo "pkg-config $(pkg-config --version), libpcre2-8 by it: $(PKG_CONFIG_LIBDIR=$tools/pcre/usr/lib/x86_64-linux-gnu/pkgconfig PKG_CONFIG_SYSROOT_DIR=$tools/pcre pkg-config --cflags --libs libpcre libpcre2-8)" || { echo "FAIL toolchains"; exit 1; }
(
	cd bench/game/rust-deps &&
	PATH=$tools/rust-1.84.1/bin:$PATH CARGO_HOME=$T/cargo CARGO_TARGET_DIR=$T/target \
	PKG_CONFIG_LIBDIR=$tools/pcre/usr/lib/x86_64-linux-gnu/pkgconfig PKG_CONFIG_SYSROOT_DIR=$tools/pcre \
	cargo build --release --locked --offline --target x86_64-unknown-linux-gnu \
		--config 'source.crates-io.replace-with="laid"' --config "source.laid.directory=\"$vendor\""
) > "$T/crates.log" 2>&1 && echo "PASS rust-deps built offline: $(ls "$T"/target/x86_64-unknown-linux-gnu/release/deps/*.rlib | wc -l) rlibs for the target, build scripts' crates kept apart" || { tail -40 "$T/crates.log"; echo "FAIL rust-deps"; exit 1; }
crates=$T/target/x86_64-unknown-linux-gnu/release/deps
echo "libc rlibs the entries link from: $(ls "$crates"/liblibc-*.rlib | wc -l)"
echo "pcre2-sys linked: $(cat "$T"/target/x86_64-unknown-linux-gnu/release/build/pcre2-sys-*/output 2>/dev/null | grep -E 'rustc-link-(lib|search)' | tr '\n' ' ')"
export GOPROXY="file://$modules/cache/download,https://proxy.golang.org" GOSUMDB=off
echo "go-pcre alone, every line: GOMODCACHE=$(go env GOMODCACHE), HOME=$HOME"
(cd bench/game && GOWORK=off GOTOOLCHAIN=local GOFLAGS= GOAMD64=v2 CGO_CFLAGS="-I$tools/pcre/usr/include" CGO_LDFLAGS="-L$tools/pcre/usr/lib/x86_64-linux-gnu" PKG_CONFIG_LIBDIR=$tools/pcre/usr/lib/x86_64-linux-gnu/pkgconfig PKG_CONFIG_SYSROOT_DIR=$tools/pcre "$tools/go1.23.1/bin/go" build -o "$T/go5" ./regexredux/regexredux.go-5.go) > "$T/go5.log" 2>&1 && echo "PASS regexredux go-5 builds alone" || { cat "$T/go5.log"; echo "NOTE regexredux go-5 doesn't build alone, the lines above"; }
(go vet ./bench/ && go test -count=1 ./bench/ ./internal/compilercensus/) > "$T/tests.log" 2>&1 && echo "PASS vet, ./bench/ and ./internal/compilercensus/" || { tail -30 "$T/tests.log"; echo "FAIL tests"; exit 1; }
echo "load before $(cut -d' ' -f1-3 /proc/loadavg); busy: $(ps -eo pcpu=,comm= | awk '$1 > 5 {printf "%s(%s) ", $2, $1}')"
go run ./bench -game -go "$tools/go1.23.1/bin/go" -rustc "$tools/rust-1.84.1/bin/rustc" -crates "$crates" -pcre "$tools/pcre" -samples "$samples" -timeout 10m
code=$?
echo "load after $(cut -d' ' -f1-3 /proc/loadavg)"
[ $code = 0 ] && echo "PASS game, $samples samples" || { echo "FAIL game exit $code"; exit 1; }

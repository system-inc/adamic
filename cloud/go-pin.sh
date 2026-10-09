# cloud/go-pin.sh names the one Go every gate machine builds with, sourced by setup.sh and gocacheprog.sh.
# Go hashes its own version into every action ID, so two toolchains share no cache entry, and a verdict's toolchain
# would depend on the day a machine was set up (Oct 9: setup installed go.dev's newest, go1.27.2, on every Codex
# instance, while the gate boxes ran go1.27.1). The pin is the pool's go1.27.2, so the boxes move rather than the pool's
# 190 instances and their warm caches. Move the pin on purpose, every machine at once.
goPin=go1.27.2

# pinnedGo <go>: true when that binary is the pinned toolchain. GOTOOLCHAIN=local asks the binary itself, never a
# toolchain a go.mod would switch to.
pinnedGo() { [ "$(cd / && GOTOOLCHAIN=local "$1" env GOVERSION 2> /dev/null)" = "$goPin" ]; }

# installPinnedGo <tools> <goArchitecture>: puts the pinned Go in <tools>/go, replacing another version there only once
# the pinned one is unpacked and answers, so a failed download never leaves a machine without a Go.
installPinnedGo() {
	local staging
	pinnedGo "$1/go/bin/go" && return 0
	staging=$(mktemp -d "$1/go-pin.XXXXXX")
	if curl -fsSL "https://dl.google.com/go/$goPin.linux-$2.tar.gz" | tar --no-same-owner -xz -C "$staging" && pinnedGo "$staging/go/bin/go"; then
		rm -rf "$1/go" && mv "$staging/go" "$1/go" && rmdir "$staging"
	else
		rm -rf "$staging"
		return 1
	fi
}

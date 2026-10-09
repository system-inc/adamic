package release

import "strings"

// Target is one platform a release ships a binary for.
type Target struct {
	// GoOperatingSystem and GoArchitecture are what `go build` is asked for, as GOOS and GOARCH.
	GoOperatingSystem string
	GoArchitecture    string
}

// Targets is every platform a release ships.
//
// The list is explicit rather than derived from `go tool dist list`, because shipping is a promise:
// a platform in this list has a published package and a machine that installs it gets a working
// binary. Deriving the list would silently grow it to platforms nobody has ever run, and a package
// that exists but was never exercised is worse than an honest "no binary for this platform" — it
// converts a clear failure at install time into an obscure one at runtime.
//
// Windows on arm64 is included because Go cross-compiles it for free from the same matrix, and a
// developer on that hardware otherwise gets an error naming a platform we could trivially have
// built. Linux arm64 is here for CI runners and containers, which are increasingly arm.
var Targets = []Target{
	{GoOperatingSystem: "darwin", GoArchitecture: "arm64"},
	{GoOperatingSystem: "darwin", GoArchitecture: "amd64"},
	{GoOperatingSystem: "linux", GoArchitecture: "arm64"},
	{GoOperatingSystem: "linux", GoArchitecture: "amd64"},
	{GoOperatingSystem: "windows", GoArchitecture: "arm64"},
	{GoOperatingSystem: "windows", GoArchitecture: "amd64"},
}

// PackageName is the npm package this target publishes as.
func (target Target) PackageName() string {
	return PlatformPackageName(target.GoOperatingSystem, target.GoArchitecture)
}

// DirectoryName is the on-disk directory for this target's package, like "cohere-darwin-arm64".
//
// It is the package name without the scope, so that a staged release directory reads the same way
// `node_modules/@system-inc/` does, the dispatcher's "cohere" beside each platform's.
func (target Target) DirectoryName() string {
	return strings.TrimPrefix(target.PackageName(), PackageScope+"/")
}

// BinaryFileName is the executable's name inside this target's package.
func (target Target) BinaryFileName() string {
	return BinaryFileName(target.GoOperatingSystem)
}

// String names the target the way Go does, for build output and error messages.
func (target Target) String() string {
	return target.GoOperatingSystem + "/" + target.GoArchitecture
}

// IsMacOS reports whether this target needs signing and notarization.
//
// Gatekeeper blocks an unsigned binary downloaded from the network, and an npm tarball counts. It
// does not block one built locally, which is exactly why this is easy to miss: it works on the
// machine that built it and fails on every machine that installs it.
func (target Target) IsMacOS() bool {
	return target.GoOperatingSystem == "darwin"
}

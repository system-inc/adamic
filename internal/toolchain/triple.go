package toolchain

import (
	"fmt"
	"runtime"
	"strings"
)

// triples are the native targets, one row for each platform a native build is for: a build's goos and goarch, and the
// triple clang is handed for them (--target), Debian's spelling on linux and Apple's on macOS (#5xsd2ha). A platform
// with no row has no native build. Here, in the package both native and cgo's compilers read, so the table has one home.
var triples = [...]struct{ goos, goarch, triple string }{
	{"linux", "amd64", "x86_64-linux-gnu"},
	{"linux", "arm64", "aarch64-linux-gnu"},
	{"darwin", "arm64", "arm64-apple-macosx"},
	{"darwin", "amd64", "x86_64-apple-macosx"},
}

// Triple is the target a native build for goos and goarch names to clang, or an error naming the pair when the table
// holds no row for it.
func Triple(goos, goarch string) (string, error) {
	for _, row := range triples {
		if row.goos == goos && row.goarch == goarch {
			return row.triple, nil
		}
	}
	return "", fmt.Errorf("no native target for %s/%s: a native build is for linux/amd64, linux/arm64, darwin/arm64 or darwin/amd64", goos, goarch)
}

// TriplePlatform is the goos and goarch whose row holds triple, and whether one does.
func TriplePlatform(triple string) (goos string, goarch string, found bool) {
	for _, row := range triples {
		if row.triple == triple {
			return row.goos, row.goarch, true
		}
	}
	return "", "", false
}

// cCompiler and cPlusPlusCompiler are the C and C++ compilers Env names by name, clang's, found as Path finds them.
const (
	cCompiler         = "clang"
	cPlusPlusCompiler = "clang++"
)

// Compilers are the C and C++ compilers Env names when the caller names neither: clang and clang++ by name, each with
// the triple of environment's platform, so a `GOARCH=arm64` build's cgo compiles for aarch64-linux-gnu where it would
// have compiled for the box (#5xsd2ha). The platform is read as go reads it from environment, a list of settings whose
// last setting of a name wins, `GOOS` and `GOARCH` unset or empty being this process's own. A platform with no row (js
// and wasip1, whose builds run no cgo) names clang and clang++ alone.
func Compilers(environment []string) (string, string) {
	goos, goarch := "", ""
	for _, setting := range environment {
		switch name, value, _ := strings.Cut(setting, "="); name {
		case "GOOS":
			goos = value
		case "GOARCH":
			goarch = value
		}
	}
	if goos == "" {
		goos = runtime.GOOS
	}
	if goarch == "" {
		goarch = runtime.GOARCH
	}
	triple, err := Triple(goos, goarch)
	if err != nil {
		return cCompiler, cPlusPlusCompiler
	}
	return cCompiler + " --target=" + triple, cPlusPlusCompiler + " --target=" + triple
}

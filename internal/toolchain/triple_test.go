package toolchain

import (
	"runtime"
	"strings"
	"testing"
)

// The triple table holds four rows and nothing else (#5xsd2ha): each platform's triple, Debian's spelling on linux and
// Apple's on macOS, each found again from its triple; any other pair refused by its name. Mutants: a row's triple
// misspelled; a fifth row.
func TestTheTripleTable(t *testing.T) {
	t.Parallel()
	want := map[[2]string]string{
		{"linux", "amd64"}:  "x86_64-linux-gnu",
		{"linux", "arm64"}:  "aarch64-linux-gnu",
		{"darwin", "arm64"}: "arm64-apple-macosx",
		{"darwin", "amd64"}: "x86_64-apple-macosx",
	}
	for platform, triple := range want {
		if got, err := Triple(platform[0], platform[1]); err != nil || got != triple {
			t.Errorf("Triple(%s, %s) is %q, %v; want %s", platform[0], platform[1], got, err, triple)
		}
		if goos, goarch, found := TriplePlatform(triple); !found || goos != platform[0] || goarch != platform[1] {
			t.Errorf("TriplePlatform(%s) is %s/%s, %t; want %s/%s", triple, goos, goarch, found, platform[0], platform[1])
		}
	}
	if len(triples) != len(want) {
		t.Errorf("the table holds %d rows, want the %d above", len(triples), len(want))
	}
	for _, platform := range [][2]string{{"windows", "amd64"}, {"linux", "riscv64"}, {"darwin", "386"}, {"", ""}} {
		_, err := Triple(platform[0], platform[1])
		if err == nil || !strings.Contains(err.Error(), "no native target for "+platform[0]+"/"+platform[1]+":") {
			t.Errorf("Triple(%q, %q): %v, want a refusal naming the pair", platform[0], platform[1], err)
		}
	}
	if _, _, found := TriplePlatform("riscv64-linux-gnu"); found {
		t.Error("TriplePlatform found a triple the table doesn't hold")
	}
}

// cgo's compilers name the build's triple, read as go reads `GOOS` and `GOARCH`: the last setting wins, so a request's
// own GOARCH=arm64 gives `CC=clang --target=aarch64-linux-gnu` over a caller's amd64, an unset or empty one is this
// process's platform, and a platform with no row (js/wasm, which runs no cgo) names clang alone. Mutants: Compilers
// reading the first setting of a name; Env handing Compilers the caller's settings without extra's.
// Not parallel: t.Setenv.
func TestCgosCompilersNameTheBuildsTriple(t *testing.T) {
	t.Setenv("CC", "")
	t.Setenv("CXX", "")
	t.Setenv("GOOS", "linux")
	t.Setenv("GOARCH", "amd64")
	if environment := values(Env("GOARCH=arm64")); environment["CC"] != "clang --target=aarch64-linux-gnu" || environment["CXX"] != "clang++ --target=aarch64-linux-gnu" {
		t.Errorf("GOARCH=arm64 over a caller's amd64: CC=%q, CXX=%q", environment["CC"], environment["CXX"])
	}
	if environment := values(Env()); environment["CC"] != "clang --target=x86_64-linux-gnu" {
		t.Errorf("the caller's linux/amd64: CC=%q", environment["CC"])
	}
	host, err := Triple(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Skipf("this host has no native target: %v", err)
	}
	if c, cPlusPlus := Compilers([]string{"GOOS=", "GOARCH="}); c != "clang --target="+host || cPlusPlus != "clang++ --target="+host {
		t.Errorf("an empty platform names %q and %q, want this host's %s", c, cPlusPlus, host)
	}
	if c, cPlusPlus := Compilers([]string{"GOOS=js", "GOARCH=wasm"}); c != cCompiler || cPlusPlus != cPlusPlusCompiler {
		t.Errorf("js/wasm names %q and %q, want clang and clang++ alone", c, cPlusPlus)
	}
}

// A build's platform is the last setting of `GOOS` and of `GOARCH`, as go reads a list of settings, and this process's
// own where either is unset or empty, even after an earlier setting named another. Mutants: Platform reading the first
// setting of a name; an empty setting kept rather than read as this process's.
func TestPlatformIsTheLastSettingOrThisProcesss(t *testing.T) {
	t.Parallel()
	for _, row := range []struct {
		environment  []string
		goos, goarch string
	}{
		{nil, runtime.GOOS, runtime.GOARCH},
		{[]string{"GOOS=linux", "GOARCH=arm64"}, "linux", "arm64"},
		{[]string{"GOARCH=amd64", "PATH=/bin", "GOARCH=arm64"}, runtime.GOOS, "arm64"},
		{[]string{"GOOS=darwin", "GOOS=", "GOARCH=arm64", "GOARCH="}, runtime.GOOS, runtime.GOARCH},
		{[]string{"GOOS=", "GOARCH="}, runtime.GOOS, runtime.GOARCH},
	} {
		if goos, goarch := Platform(row.environment); goos != row.goos || goarch != row.goarch {
			t.Errorf("Platform(%q) is %s/%s, want %s/%s", row.environment, goos, goarch, row.goos, row.goarch)
		}
	}
}

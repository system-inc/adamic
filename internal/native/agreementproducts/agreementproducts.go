// Package agreementproducts is internal/lower's native agreement binaries as lean buildcache products: each row that
// runs a program natively fetches the binary its C builds, and TestProduct_LowerAgreementNative builds every one ahead.
//
// Built as plain products, each traced window held the whole process's reads, so the 14 builds read each other's, a
// ring Loom refuses to record (#kfjqsnh, #egjbcre). So each is a LeanProductOf build (BuildAgreementNative), keyed by
// its own window and what it's handed: the C, in a file the test wrote, and the runtime archive its features need,
// found first as runtimeproducts' own product. It builds through native.BuildLean, which reads nothing of the process,
// and the code that build reaches is the files it's keyed by (NativeFiles). The recipe is here, not in lower's tests,
// since buildcache builds lean only for a package it has reviewed (its leanCallers).
package agreementproducts

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/native"
)

// NativeFiles are internal/native's files defining everything native.BuildLean reaches, which an agreement native's
// lean recipe keys as files: runtimeproducts' TestTheLeanBuildReachesOnlyItsDeclaredFiles holds BuildLean to them, and
// the rest of internal/native, the compiler, is no input of the binary. The runtime is keyed as the archive it's handed.
var NativeFiles = []string{"internal/native/jobserver.go", "internal/native/lean_build.go", "internal/native/library.go", "internal/native/native.go", "internal/native/regexp_compiler_library.go", "internal/native/target.go"}

// Product is the directory holding the native agreement binary, native, that the C at source builds, linked with the
// runtime archive in archive (a runtimeproducts product's directory). split is ADAMIC_NATIVE_SPLIT as the caller read
// it: native.BuildLean builds one translation unit, and a split build reads the unit cache, so split on is refused.
func Product(t testing.TB, source string, archive string, split string) string {
	t.Helper()
	if split == "1" {
		t.Fatal("ADAMIC_NATIVE_SPLIT=1: an agreement native is built lean, in one translation unit (native.BuildLean)")
	}
	code, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := archiveDigest(archive)
	if err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{
		// v3: a lean build of what it's handed, keyed by its own window (#egjbcre).
		Name:      "lower agreement native v3",
		Files:     append(slices.Clone(NativeFiles), "internal/native/agreementproducts"),
		Flags:     append(native.Flags(options()), fmt.Sprintf("C=%x", sha256.Sum256(code)), "archive="+digest, "ADAMIC_NATIVE_SPLIT="+split),
		Toolchain: []string{runtime.GOOS, runtime.GOARCH, buildcache.Tool("clang", "--version")},
	}
	return buildcache.LeanProductOf(t, buildcache.Handed{Inputs: inputs, Paths: []string{source, archive}}, BuildAgreementNative)
}

// BuildAgreementNative builds what it's handed, the C and the archive's directory, into directory as native, under the
// process's compile jobserver, which decides when clang runs and never what it writes. It's a top-level function of
// what it's handed alone (buildcache.LeanProductOf).
func BuildAgreementNative(handed buildcache.Handed, directory string) error {
	if len(handed.Paths) != 2 {
		return fmt.Errorf("agreementproducts: handed %q, want the C and the archive's directory", handed.Paths)
	}
	jobserver, err := ccjobserver.Process()
	if err != nil {
		return err
	}
	return native.BuildLean(handed.Paths[0], handed.Paths[1], filepath.Join(directory, "native"), options(), jobserver)
}

// options are what every agreement native is compiled with: native, unsanitized, as a user's build.
func options() native.Options {
	return native.Options{}
}

// archiveDigest is the sha256 of every file the archive's directory holds, each by its name and the sha256 of what it
// holds: so a build outside a trace, keyed by Inputs alone, keys apart on another runtime.
func archiveDigest(directory string) (string, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}
		content, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		if err != nil {
			return "", err
		}
		fmt.Fprintf(hash, "%s %x\n", entry.Name(), sha256.Sum256(content))
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

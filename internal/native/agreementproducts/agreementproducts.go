// Package agreementproducts is internal/lower's native agreement binaries as lean buildcache products: each row that
// runs a program natively fetches the binary its C builds, and TestProduct_LowerAgreementNative builds every one ahead.
//
// Built as plain products, each traced window held the whole process's reads, so the 14 builds read each other's, a
// ring Loom refuses to record (#kfjqsnh, #egjbcre). So each is a LeanProductOf build (BuildAgreementNative), keyed by
// its own window and what it's handed: the C, a file of the lower agreement C (C, a product of its own, built ahead by
// TestProduct_LowerAgreementC), and the runtime archive its features need, found first as runtimeproducts' own product.
// It builds through native.BuildLean, which reads nothing of the process, and the code that build reaches is the files
// it's keyed by (NativeFiles). The recipe is here, not in lower's tests, since buildcache builds lean only for a package
// it has reviewed (its leanCallers), and the C's build is here too, since it runs stage 0 rather than lowering in
// process: a package that imported internal/lower couldn't be imported by lower's own tests.
package agreementproducts

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
)

// NativeFiles are internal/native's files defining everything native.BuildLean reaches, which an agreement native's
// lean recipe keys as files: runtimeproducts' TestTheLeanBuildReachesOnlyItsDeclaredFiles holds BuildLean to them, and
// the rest of internal/native, the compiler, is no input of the binary. The runtime is keyed as the archive it's handed.
var NativeFiles = []string{"internal/native/jobserver.go", "internal/native/lean_build.go", "internal/native/library.go", "internal/native/native.go", "internal/native/regexp_compiler_library.go", "internal/native/target.go"}

// Product is the directory holding the native agreement binary, native, that the C at source, a file of the lower
// agreement C (FindC), builds, linked with the runtime archive in archive (a runtimeproducts product's directory).
// split is ADAMIC_NATIVE_SPLIT as the caller read it: native.BuildLean builds one translation unit, and a split build
// reads the unit cache, so split on is refused.
//
// clang is found here, outside the build's window, as Build finds it (toolchain.Path), keyed by its report
// (buildcache.Tool, which names what it is and never where), and handed to the build by its absolute path, the third of
// handed.Paths: a handed path never keys, so the key is the same on every machine running the same clang, and the
// window values what it reads there by content. A Flags entry would key the path's spelling, per box.
func Product(t testing.TB, source string, archive string, split string) string {
	t.Helper()
	if split == "1" {
		t.Fatal("ADAMIC_NATIVE_SPLIT=1: an agreement native is built lean, in one translation unit (native.BuildLean)")
	}
	compiler, err := toolchain.Path("clang")
	if err != nil {
		t.Fatal(err)
	}
	// A C its test wrote, outside every product, is a read under the temporary directory, which a traced window can only
	// refuse as temp not made by this run (#egjbcre): refused here first, on every machine, naming where it came from.
	if !strings.HasPrefix(buildcache.Relative(source), "<build cache>/") {
		t.Fatalf("agreementproducts: the C at %s isn't a product's: hand a native the lower agreement C's <name>/main.c (FindC)", source)
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
		Toolchain: []string{runtime.GOOS, runtime.GOARCH, buildcache.Tool(compiler, "--version")},
	}
	return buildcache.LeanProductOf(t, buildcache.Handed{Inputs: inputs, Paths: []string{source, archive, compiler}}, BuildAgreementNative)
}

// BuildAgreementNative builds what it's handed, the C, the archive's directory and clang, into directory as native.
// It's a top-level function of what it's handed alone (buildcache.LeanProductOf), but for one ambient handle: the
// process's compile jobserver (ccjobserver.Process), the only thing it takes that it isn't handed, since a named build
// can't close over its caller's. The jobserver decides when clang runs, never what it writes: BuildLean's binary is the
// same bytes under none, one token and eight (internal/lower's TestNativeAgreementLeanBuildIsBuildsBytes).
func BuildAgreementNative(handed buildcache.Handed, directory string) error {
	if len(handed.Paths) != 3 {
		return fmt.Errorf("agreementproducts: handed %q, want the C, the archive's directory and clang", handed.Paths)
	}
	jobserver, err := ccjobserver.Process()
	if err != nil {
		return err
	}
	return native.BuildLean(handed.Paths[2], handed.Paths[0], handed.Paths[1], filepath.Join(directory, "native"), options(), jobserver)
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

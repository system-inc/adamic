package countproducts

import (
	"crypto/sha256"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native/agreementproducts"
	"github.com/system-inc/adamic/internal/native_runtime"
	"github.com/system-inc/adamic/internal/toolchain"
)

// planted is a shard's keying of fixed values, so a test moves one of them alone, and keys it without clang.
func planted() keying {
	return keying{
		rows:     fmt.Sprintf("%x", sha256.Sum256([]byte("rows\n"))),
		harness:  Harness,
		runtime:  "runtime",
		flags:    []string{"--target=x86_64-unknown-linux-gnu", "-DADAMIC_COUNT=1"},
		compiler: "clang version 21.1.8",
		platform: []string{"linux", "amd64"},
	}
}

// keyOf is the key of the shard keyed, as an untraced build keys it (buildcache.Key of its inputs), never built.
func keyOf(t *testing.T, keyed keying) string {
	t.Helper()
	key, err := buildcache.Key(repositoryRoot(t), keyed.inputs())
	if err != nil {
		t.Fatal(err)
	}
	return key
}

// A shard's key names each thing a count reads besides its code: the rows file, the harness, the runtime's sources, the
// counted build's flags and the target they name, clang's report and the build's platform. Each moved alone moves it.
// Mutant: any one of them left out of inputs (the rows file's digest dropped from its Flags).
func TestAShardsKeyNamesEachInput(t *testing.T) {
	t.Parallel()
	base := keyOf(t, planted())
	for _, input := range []struct {
		name  string
		plant func(keyed *keying)
	}{
		{"the rows file", func(keyed *keying) { keyed.rows = fmt.Sprintf("%x", sha256.Sum256([]byte("rows!\n"))) }},
		{"the harness", func(keyed *keying) { keyed.harness += "!" }},
		{"the runtime's sources", func(keyed *keying) { keyed.runtime += "!" }},
		{"the counted flags", func(keyed *keying) { keyed.flags = append(slices.Clone(keyed.flags), "-O3") }},
		{"the target", func(keyed *keying) { keyed.flags = []string{"--target=aarch64-unknown-linux-gnu", "-DADAMIC_COUNT=1"} }},
		{"clang's report", func(keyed *keying) { keyed.compiler += "!" }},
		{"the platform", func(keyed *keying) { keyed.platform = []string{"darwin", "arm64"} }},
	} {
		keyed := planted()
		input.plant(&keyed)
		if key := keyOf(t, keyed); key == base {
			t.Errorf("%s moved, and the shard's key stayed %s", input.name, base[:12])
		}
	}
}

// What keyingOf keys is what a count reads on this machine: the rows file's own bytes, Harness, the runtime's sources,
// the counted build's flags, clang's report under their target, and the platform native's builds read. Mutant: keyingOf
// keying the uncounted build's flags.
func TestAShardsKeyingIsWhatItsCountsRead(t *testing.T) {
	t.Parallel()
	compiler, err := toolchain.Path(native_runtime.CompilerName(options()))
	if err != nil {
		t.Fatal(err)
	}
	rows := []byte("rows\n")
	keyed, err := keyingOf(rows, compiler)
	if err != nil {
		t.Fatal(err)
	}
	sources, err := native_runtime.RuntimeSourceDigest()
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := native_runtime.ResolveTarget(native_runtime.Options{Count: true})
	if err != nil {
		t.Fatal(err)
	}
	goos, goarch := toolchain.Platform(os.Environ())
	want := keying{
		rows:     fmt.Sprintf("%x", sha256.Sum256(rows)),
		harness:  Harness,
		runtime:  sources,
		flags:    native_runtime.Flags(native_runtime.Options{Count: true}),
		compiler: buildcache.Tool(compiler, "--target="+resolved.Target, "--version"),
		platform: []string{goos, goarch},
	}
	if keyed.rows != want.rows || keyed.harness != want.harness || keyed.runtime != want.runtime || !slices.Equal(keyed.flags, want.flags) ||
		keyed.compiler != want.compiler || !slices.Equal(keyed.platform, want.platform) {
		t.Errorf("a shard keyed %+v, want %+v", keyed, want)
	}
}

// Each shard is keyed by its own rows: a byte added to one row's C moves that row's shard's key and no other of the 16,
// and a change to the runtime's sources, which every counted archive is built from, moves all 16 though no row's C
// moved. Mutants: Rows dropping a row's C; a shard keyed by every shard's rows; the runtime's digest left out of inputs.
func TestAOneByteOfOneRowsCMovesOnlyItsShard(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	const shards = 16
	const changed = 3
	table := make([][]Row, shards)
	for shard := range shards {
		table[shard] = []Row{{Key: fmt.Sprintf("internal/oracle/testdata/shard_%02d.a", shard), C: "int main(void) { return 0; }\n", Run: Plain}}
	}
	keysOf := func(runtime string, table [][]Row) []string {
		t.Helper()
		var keys []string
		for _, rows := range table {
			encoded, _, err := Rows(rows, root)
			if err != nil {
				t.Fatal(err)
			}
			keyed := planted()
			keyed.rows, keyed.runtime = fmt.Sprintf("%x", sha256.Sum256(encoded)), runtime
			keys = append(keys, keyOf(t, keyed))
		}
		return keys
	}
	before := keysOf("runtime", table)
	edited := slices.Clone(table)
	edited[changed] = []Row{table[changed][0]}
	edited[changed][0].C += " "
	after := keysOf("runtime", edited)
	for shard := range shards {
		if moved := before[shard] != after[shard]; moved != (shard == changed) {
			t.Errorf("a byte added to shard %d's C: shard %d's key moved %v", changed, shard, moved)
		}
	}
	for shard, key := range keysOf("runtime!", table) {
		if key == before[shard] {
			t.Errorf("the runtime's sources moved, and shard %d's key stayed %s", shard, key[:12])
		}
	}
}

// A count keys as files the native files native.BuildLean reaches, the list runtimeproducts' census holds BuildLean to
// (agreementproducts.NativeFiles), spelled here so a count's recipe needn't import agreementproducts. Mutant: a file
// added to agreementproducts.NativeFiles and not to nativeFiles.
func TestTheNativeFilesAreAgreementProductsOwn(t *testing.T) {
	t.Parallel()
	if !slices.Equal(nativeFiles(), agreementproducts.NativeFiles) {
		t.Errorf("a count keys the native files %q, and BuildLean reaches %q", nativeFiles(), agreementproducts.NativeFiles)
	}
}

// A shard's key reads this package's own files and the native files BuildLean reaches as files, so an untraced settle
// moves the key on a harness edit as a traced one does, and every non-test file of this package is among them.
// Mutants: either list dropped from inputs' Files, and a harness edit reuses old counts; a new file of the package left
// out of declaredFiles.
func TestAShardsKeyHoldsItsOwnFiles(t *testing.T) {
	t.Parallel()
	files := keying{}.inputs().Files
	for _, want := range append(nativeFiles(), declaredFiles()...) {
		if !slices.Contains(files, want) {
			t.Errorf("a shard's key doesn't read %s as a file", want)
		}
	}
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		if path := "internal/oracle/countproducts/" + name; !slices.Contains(declaredFiles(), path) {
			t.Errorf("%s is this package's and isn't among its declared files", path)
		}
	}
}

package typeaware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/corpusfiles"
)

// Repository cases use 32 stable hash buckets per mode, with room for growth.
// Compiler ranges are fixed by corpusfiles.TypeScriptCommit, not repository HEAD.
// Both assignments have separate plain and sanitized leaves.
const testVolumeProfileCorporaShards = 2 * (32 + 77)

// Cancel the process group so Go and stage0 cannot leave compiler children
// running after a deadline. Setpgid and negative-PID Kill work on Linux and macOS.
func volumeProfileCorporaCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	return command
}

func volumeProfileCorporaNativeInputs(ctx context.Context, h *harness, sanitize bool) buildcache.Inputs {
	h.t.Helper()
	name := "typeaware volume"
	if sanitize {
		name += " asan"
	}
	// Use the compiler source recipe rather than executable metadata: workers
	// have different Git revisions and Go archives embed different scratch paths.
	archiveName, archiveFlags := "typeaware checker archive", ""
	toolContext, cancelTool := context.WithCancel(ctx)
	defer cancelTool()
	cc, err := volumeProfileCorporaCommand(toolContext, "go", "env", "CC").Output()
	if err != nil {
		h.t.Fatal(err)
	}
	if sanitize {
		archiveName += " asan"
		archiveFlags = "CC=clang CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all"
		cc = []byte("clang")
	}
	inputs := buildcache.Inputs{
		Name:      name,
		Flags:     []string{"build", "stage1/cohere/typeaware/volume_suite.ts", "--tsgo", "go build -buildvcs=false ./cmd/adamic", "archive=" + archiveName, "go build -buildvcs=false -buildmode=c-archive ./bridge/tsgo/archive", archiveFlags, fmt.Sprintf("sanitize=%t", sanitize), "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"), "ADAMIC_NATIVE_JOBS=" + os.Getenv("ADAMIC_NATIVE_JOBS"), "ADAMIC_GATE_UNCACHED=" + os.Getenv("ADAMIC_GATE_UNCACHED")},
		Files:     []string{"bridge/tsgo", "cohere/TypeScript/tsc/internal", "cohere/TypeScript/tsc/go.mod", "cohere/TypeScript/tsc/go.sum", "cohere/TypeScript-shim", "go.mod", "cohere/go.mod", "cohere/go.sum"},
		Toolchain: []string{buildcache.Tool("clang", "--version"), buildcache.Tool(strings.Fields(string(cc))[0], "--version"), buildcache.Tool("go", "version"), buildcache.Tool("go", "env", "-json", "GOOS", "GOARCH", "GOAMD64", "GOARM64", "CGO_ENABLED", "CC", "CXX", "CGO_CFLAGS", "CGO_CPPFLAGS", "CGO_CXXFLAGS", "CGO_LDFLAGS", "GOFLAGS", "GOEXPERIMENT")},
	}
	listContext, cancelList := context.WithCancel(ctx)
	defer cancelList()
	list := volumeProfileCorporaCommand(listContext, "go", "list", "-deps", "-json", "./cmd/adamic")
	list.Dir = h.repository
	data, err := list.Output()
	if err != nil {
		h.t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	files := map[string]bool{}
	for {
		var pkg struct {
			Dir, ImportPath                                                  string
			Standard                                                         bool
			GoFiles, CgoFiles, CFiles, HFiles, SFiles, SysoFiles, EmbedFiles []string
			Module                                                           *struct{ GoMod string }
		}
		err := decoder.Decode(&pkg)
		if err == io.EOF {
			break
		}
		if err != nil {
			h.t.Fatal(err)
		}
		if pkg.Standard {
			continue
		}
		var paths []string
		for _, names := range [][]string{pkg.GoFiles, pkg.CgoFiles, pkg.CFiles, pkg.HFiles, pkg.SFiles, pkg.SysoFiles, pkg.EmbedFiles} {
			for _, name := range names {
				paths = append(paths, filepath.Join(pkg.Dir, name))
			}
		}
		if pkg.Module != nil {
			paths = append(paths, pkg.Module.GoMod)
		}
		for _, path := range paths {
			relative, err := filepath.Rel(h.repository, path)
			if err != nil {
				h.t.Fatal(err)
			}
			if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
				data, err := os.ReadFile(path)
				if err != nil {
					h.t.Fatal(err)
				}
				inputs.Flags = append(inputs.Flags, fmt.Sprintf("dependency %s %s %x", pkg.ImportPath, filepath.Base(path), sha256.Sum256(data)))
			} else {
				files[filepath.ToSlash(relative)] = true
			}
		}
	}
	var compilerFiles []string
	for path := range files {
		compilerFiles = append(compilerFiles, path)
	}
	slices.Sort(compilerFiles)
	inputs.Files = append(inputs.Files, compilerFiles...)
	if _, err := os.Stat(filepath.Join(h.repository, "go.sum")); err == nil {
		inputs.Files = append(inputs.Files, "go.sum")
	} else if os.IsNotExist(err) {
		inputs.Flags = append(inputs.Flags, "go.sum absent")
	} else {
		h.t.Fatal(err)
	}
	for _, directory := range []string{"stage1/cohere/typeaware", "stage1/cohere/lint", "stage1/typescript"} {
		err := filepath.WalkDir(filepath.Join(h.repository, directory), func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !entry.IsDir() && strings.HasSuffix(path, ".ts") {
				relative, err := filepath.Rel(h.repository, path)
				if err != nil {
					return err
				}
				inputs.Files = append(inputs.Files, filepath.ToSlash(relative))
			}
			return nil
		})
		if err != nil {
			h.t.Fatal(err)
		}
	}
	return inputs
}

// Emission is a separate cached product shared by both sanitizer modes.
// Compile products read that immutable C; they never run stage zero again.
func volumeProfileCorporaNative(ctx context.Context, h *harness, stage0, archive string, inputs buildcache.Inputs, sanitize bool) string {
	lowered := volumeProfileControlsLowered(h, stage0)
	source, err := os.ReadFile(lowered)
	if err != nil {
		h.t.Fatal(err)
	}
	inputs.Flags = append(slices.Clone(inputs.Flags), "cached emission compile v2", fmt.Sprintf("source sha256=%x", sha256.Sum256(source)))
	directory := buildcache.Product(h.t, inputs, func(directory string) error {
		if err := typeAwareNativeBuild(string(source), filepath.Join(directory, "volume"), archive, sanitize); err != nil {
			return fmt.Errorf("%s: %w", inputs.Name, err)
		}
		return nil
	})
	return filepath.Join(directory, "volume")
}

type volumeProfileCorporaShard struct {
	corpus, config string
	paths          []string
	sanitize       bool
	enabled        bool
}

func volumeProfileCorporaEnumeration(t *testing.T) (string, []volumeProfileCorporaShard) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	var shards []volumeProfileCorporaShard
	for _, corpus := range []struct {
		name, manifest, config string
		ranges                 int
	}{
		{"repository", os.Getenv("ADAMIC_VOLUME_REPOSITORY_MANIFEST"), filepath.Join(repository, "tsconfig.json"), 32},
		{"compiler", os.Getenv("ADAMIC_VOLUME_COMPILER_MANIFEST"), filepath.Join(os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), "src/compiler/tsconfig.json"), 77},
	} {
		var paths []string
		if corpus.manifest != "" {
			data, err := os.ReadFile(corpus.manifest)
			if err != nil {
				t.Fatal(err)
			}
			// Match both original consumers: only empty lines are omitted; retain
			// manifest order, relative paths, whitespace and duplicate occurrences.
			for _, path := range strings.Split(string(data), "\n") {
				if path != "" {
					paths = append(paths, path)
				}
			}
		}
		resolve := func(path string) string {
			if !filepath.IsAbs(path) {
				path = filepath.Join(filepath.Dir(corpus.config), path)
			}
			absolute, err := filepath.Abs(path)
			if err != nil {
				t.Fatal(err)
			}
			return absolute
		}
		if corpus.manifest != "" {
			if len(paths) == 0 {
				t.Fatalf("%s corpus manifest is empty", corpus.name)
			}
			if corpus.name == "compiler" {
				expected := corpusfiles.Upstream(t, os.Getenv("ADAMIC_TYPESCRIPT_SOURCE"), corpusfiles.TypeScriptCommit, []string{"src/compiler"}, []string{"*.ts"})
				actual := make([]string, len(paths))
				for i, path := range paths {
					actual[i] = resolve(path)
				}
				slices.Sort(actual)
				if !slices.Equal(actual, expected) {
					t.Fatalf("compiler manifest has %d files; must match all %d files at pin %s", len(actual), len(expected), corpusfiles.TypeScriptCommit)
				}
			}
		}
		parts := make([][2][]string, corpus.ranges)
		for mode := 0; mode < 2; mode++ {
			if corpus.name == "repository" {
				for _, path := range paths {
					relative, err := filepath.Rel(repository, resolve(path))
					if err != nil {
						t.Fatal(err)
					}
					if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
						t.Fatalf("repository corpus file outside repository: %s", path)
					}
					key := filepath.ToSlash(relative) + "\x00" + []string{"plain", "asan"}[mode]
					digest := sha256.Sum256([]byte(key))
					bucket := int(binary.LittleEndian.Uint64(digest[:8]) % uint64(corpus.ranges))
					parts[bucket][mode] = append(parts[bucket][mode], path)
				}
			} else {
				for i := range parts {
					parts[i][mode] = paths[len(paths)*i/corpus.ranges : len(paths)*(i+1)/corpus.ranges]
				}
			}
		}
		for _, part := range parts {
			for mode, paths := range part {
				shards = append(shards, volumeProfileCorporaShard{corpus.name, corpus.config, paths, mode == 1, corpus.manifest != ""})
			}
		}
		for _, sanitize := range []bool{false, true} {
			want, got := map[string]int{}, map[string]int{}
			for _, path := range paths {
				want[path]++
			}
			for _, s := range shards {
				if s.corpus == corpus.name && s.sanitize == sanitize {
					for _, path := range s.paths {
						got[path]++
					}
				}
			}
			if !maps.Equal(want, got) {
				t.Fatalf("%s sanitize=%t shard union differs from live manifest", corpus.name, sanitize)
			}
			t.Logf("%s sanitize=%t union: %d manifest cases, each in exactly one of %d shards", corpus.name, sanitize, len(paths), corpus.ranges)
		}
	}
	if len(shards) != testVolumeProfileCorporaShards {
		t.Fatalf("enumerated %d shards; want %d", len(shards), testVolumeProfileCorporaShards)
	}
	return repository, shards
}

type volumeProfileCorporaProductPaths struct {
	Binary string
	Asan   string
	Oracle string
}

func volumeProfileCorporaProductInputs(ctx context.Context, h *harness) (buildcache.Inputs, buildcache.Inputs, buildcache.Inputs) {
	plain := volumeProfileCorporaNativeInputs(ctx, h, false)
	sanitized := volumeProfileCorporaNativeInputs(ctx, h, true)
	// The independent Go oracle also depends on cohere's production lint rules.
	// Include their complete pinned sources, module files and the overlay source.
	bundle := plain
	bundle.Name = "typeaware volume corpus products"
	bundle.Files = append(slices.Clone(plain.Files), "cohere/internal", "cohere/policy", "cohere/rule_runner", "cohere/static_single_assignment", "cohere/mutation_aliasing", "stage1/cohere/typeaware/testdata/oracle_volume.go")
	bundle.Flags = append(slices.Clone(plain.Flags), "corpus products format v2 cached emission", "Go builds -buildvcs=false", "asan CC=clang CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all")
	return plain, sanitized, bundle
}

func volumeProfileCorporaReadProducts(t *testing.T, directory string) volumeProfileCorporaProductPaths {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(directory, "products.json"))
	if err != nil {
		t.Fatal(err)
	}
	var products volumeProfileCorporaProductPaths
	if err := json.Unmarshal(data, &products); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{products.Binary, products.Asan, products.Oracle} {
		if path == "" {
			t.Fatal("incomplete corpus setup products")
		}
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("corpus setup product unavailable: %v", err)
		}
	}
	return products
}

func volumeProfileCorporaGoProduct(t *testing.T, ctx context.Context, repository, output string) string {
	h := &harness{t: t, repository: repository}
	_, _, inputs := volumeProfileCorporaProductInputs(ctx, h)
	deadline := ctx
	var stage0, archive, asanArchive, oracle string
	jobs := []struct {
		name, output string
		product      *string
		flags        []string
		command      func(string) *exec.Cmd
	}{
		{"typeaware stage0", "adamic", &stage0, []string{"go build -buildvcs=false ./cmd/adamic"}, func(directory string) *exec.Cmd {
			return volumeProfileCorporaCommand(deadline, "go", "build", "-buildvcs=false", "-o", filepath.Join(directory, "adamic"), "./cmd/adamic")
		}},
		{"typeaware checker archive", "checker.a", &archive, []string{"go build -buildvcs=false -buildmode=c-archive ./bridge/tsgo/archive"}, func(directory string) *exec.Cmd {
			return volumeProfileCorporaCommand(deadline, "go", "build", "-buildvcs=false", "-buildmode=c-archive", "-o", filepath.Join(directory, "checker.a"), "./bridge/tsgo/archive")
		}},
		{"typeaware checker archive asan", "checker-asan.a", &asanArchive, []string{"go build -buildvcs=false -buildmode=c-archive ./bridge/tsgo/archive", "CC=clang", "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all"}, func(directory string) *exec.Cmd {
			command := volumeProfileCorporaCommand(deadline, "go", "build", "-buildvcs=false", "-buildmode=c-archive", "-o", filepath.Join(directory, "checker-asan.a"), "./bridge/tsgo/archive")
			command.Env = append(os.Environ(), "CC=clang", "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all")
			return command
		}},
		{"typeaware volume oracle", "volume-oracle", &oracle, []string{"go build -buildvcs=false -overlay oracle_volume.go cohere/adamic_volume-oracle.go"}, func(directory string) *exec.Cmd {
			virtual := filepath.Join(repository, "cohere/adamic_volume-oracle.go")
			data, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(repository, "stage1/cohere/typeaware/testdata/oracle_volume.go")}})
			if err != nil {
				t.Fatal(err)
			}
			overlay := filepath.Join(directory, "volume-oracle-overlay.json")
			if err := os.WriteFile(overlay, data, 0600); err != nil {
				t.Fatal(err)
			}
			command := volumeProfileCorporaCommand(deadline, "go", "build", "-buildvcs=false", "-overlay", overlay, "-o", filepath.Join(directory, "volume-oracle"), virtual)
			command.Dir = filepath.Join(repository, "cohere")
			return command
		}},
	}
	for _, job := range jobs {
		if job.output != output {
			continue
		}
		productInputs := inputs
		productInputs.Name = job.name
		productInputs.Flags = append(slices.Clone(inputs.Flags), job.flags...)
		directory := buildcache.Product(t, productInputs, func(directory string) error {
			command := job.command(directory)
			if command.Dir == "" {
				command.Dir = repository
			}
			if data, err := command.CombinedOutput(); err != nil {
				return fmt.Errorf("%s: %w\n%s", job.name, err, data)
			}
			return nil
		})
		return filepath.Join(directory, job.output)
	}
	t.Fatalf("unknown corpus Go product %q", output)
	return ""
}

// Shared products are prepared once per process by any selected shard or product
// declaration, before the shard's own deadline. Loom bounds the whole unit.
var volumeProfileCorporaShared struct {
	once     sync.Once
	products volumeProfileCorporaProductPaths
}

func volumeProfileCorporaPrepare(t *testing.T) volumeProfileCorporaProductPaths {
	t.Helper()
	volumeProfileCorporaShared.once.Do(func() {
		volumeProfileCorporaShared.products = volumeProfileCorporaBuildProducts(t)
	})
	if volumeProfileCorporaShared.products.Oracle == "" {
		t.Fatal("shared corpus preparation did not complete")
	}
	return volumeProfileCorporaShared.products
}

func volumeProfileCorporaBuildProducts(t *testing.T) volumeProfileCorporaProductPaths {
	deadline := context.Background()
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: t.TempDir()}
	plain, sanitized, inputs := volumeProfileCorporaProductInputs(deadline, h)
	// Every build belongs to setup. Each Go product has a stable source recipe;
	// -buildvcs=false avoids worker-specific Git metadata in cached executables.
	var stage0, archive, asanArchive, oracle string
	var builds sync.WaitGroup
	for _, job := range []struct {
		output  string
		product *string
	}{
		{"adamic", &stage0}, {"checker.a", &archive}, {"checker-asan.a", &asanArchive}, {"volume-oracle", &oracle},
	} {
		builds.Add(1)
		go func() {
			defer builds.Done()
			*job.product = volumeProfileCorporaGoProduct(t, deadline, repository, job.output)
		}()
	}
	builds.Wait()
	if t.Failed() {
		t.Fatal("corpus Go product preparation failed")
	}
	var products volumeProfileCorporaProductPaths
	products.Oracle = oracle
	// Prepare the two native products concurrently, without a setup deadline.
	builds.Add(2)
	go func() {
		defer builds.Done()
		products.Binary = volumeProfileCorporaNative(deadline, h, stage0, archive, plain, false)
	}()
	go func() {
		defer builds.Done()
		products.Asan = volumeProfileCorporaNative(deadline, h, stage0, asanArchive, sanitized, true)
	}()
	builds.Wait()
	if t.Failed() {
		t.Fatal("corpus native product preparation failed")
	}
	directory := buildcache.Product(t, inputs, func(directory string) error {
		data, err := json.Marshal(products)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "products.json"), data, 0600)
	})
	return volumeProfileCorporaReadProducts(t, directory)
}

// Individual shards can initialize the same products without another test.
func volumeProfileCorporaCachedProducts(t *testing.T, repository string) volumeProfileCorporaProductPaths {
	return volumeProfileCorporaPrepare(t)
}

func TestVolumeProfileCorporaUnion(t *testing.T) {
	t.Parallel()
	volumeProfileCorporaEnumeration(t)
}

func volumeProfileCorporaRun(t *testing.T, index int) {
	t.Helper()
	repository, shards := volumeProfileCorporaEnumeration(t)
	if index < 0 || index >= len(shards) {
		t.Fatalf("invalid corpus shard %d", index)
	}
	s := shards[index]
	if !s.enabled {
		t.Skip("set ADAMIC_VOLUME_" + strings.ToUpper(s.corpus) + "_MANIFEST")
	}
	if len(s.paths) == 0 {
		t.Skip("empty shard")
	}
	products := volumeProfileCorporaCachedProducts(t, repository)
	h := &harness{t: t, repository: repository, directory: t.TempDir()}
	manifest := h.write("corpus.manifest", strings.Join(s.paths, "\n")+"\n")
	executable, name := products.Binary, s.corpus
	if s.sanitize {
		executable, name = products.Asan, name+"-asan"
	}
	t.Logf("%s: %d files, first=%s last=%s", name, len(s.paths), s.paths[0], s.paths[len(s.paths)-1])
	started := time.Now()
	deadline, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	run := func(label, binary string) result {
		r := h.run(label, volumeProfileCorporaCommand(deadline, binary, s.config, manifest))
		if deadline.Err() != nil {
			t.Fatalf("cooked: shard exceeded 90s hard deadline (60s budget)")
		}
		if r.err != nil {
			t.Fatalf("%s: %v\n%s\n%s", label, r.err, r.stdout, r.stderr)
		}
		return r
	}
	want := run(name+"-go", products.Oracle)
	got := run(name+"-native", executable)
	if len(got.stderr) != 0 {
		t.Fatalf("sanitizer stderr: %s", got.stderr)
	}
	if !bytes.Equal(got.stdout, want.stdout) {
		i := firstDifference(got.stdout, want.stdout)
		t.Fatalf("%s mismatch byte %d: native %q Go %q", name, i, got.stdout[max(0, i-50):min(len(got.stdout), i+250)], want.stdout[max(0, i-50):min(len(want.stdout), i+250)])
	}
	t.Logf("%s: %d identical finding bytes; %s; cooked=false", name, len(want.stdout), summary(want.stdout))
	t.Logf("comparison sides: oracle %.3fs, native including byte check %.3fs", want.elapsed.Seconds(), time.Since(started).Seconds()-want.elapsed.Seconds())
}

// These wrappers are deliberately top-level: the gate discovers them with
// go test -list, without a children table or any non-test changes.
func TestVolumeProfileCorpora_000(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 0) }
func TestVolumeProfileCorpora_001(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 1) }
func TestVolumeProfileCorpora_002(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 2) }
func TestVolumeProfileCorpora_003(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 3) }
func TestVolumeProfileCorpora_004(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 4) }
func TestVolumeProfileCorpora_005(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 5) }
func TestVolumeProfileCorpora_006(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 6) }
func TestVolumeProfileCorpora_007(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 7) }
func TestVolumeProfileCorpora_008(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 8) }
func TestVolumeProfileCorpora_009(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 9) }
func TestVolumeProfileCorpora_010(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 10) }
func TestVolumeProfileCorpora_011(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 11) }
func TestVolumeProfileCorpora_012(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 12) }
func TestVolumeProfileCorpora_013(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 13) }
func TestVolumeProfileCorpora_014(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 14) }
func TestVolumeProfileCorpora_015(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 15) }
func TestVolumeProfileCorpora_016(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 16) }
func TestVolumeProfileCorpora_017(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 17) }
func TestVolumeProfileCorpora_018(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 18) }
func TestVolumeProfileCorpora_019(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 19) }
func TestVolumeProfileCorpora_020(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 20) }
func TestVolumeProfileCorpora_021(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 21) }
func TestVolumeProfileCorpora_022(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 22) }
func TestVolumeProfileCorpora_023(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 23) }
func TestVolumeProfileCorpora_024(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 24) }
func TestVolumeProfileCorpora_025(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 25) }
func TestVolumeProfileCorpora_026(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 26) }
func TestVolumeProfileCorpora_027(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 27) }
func TestVolumeProfileCorpora_028(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 28) }
func TestVolumeProfileCorpora_029(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 29) }
func TestVolumeProfileCorpora_030(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 30) }
func TestVolumeProfileCorpora_031(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 31) }
func TestVolumeProfileCorpora_032(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 32) }
func TestVolumeProfileCorpora_033(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 33) }
func TestVolumeProfileCorpora_034(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 34) }
func TestVolumeProfileCorpora_035(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 35) }
func TestVolumeProfileCorpora_036(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 36) }
func TestVolumeProfileCorpora_037(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 37) }
func TestVolumeProfileCorpora_038(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 38) }
func TestVolumeProfileCorpora_039(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 39) }
func TestVolumeProfileCorpora_040(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 40) }
func TestVolumeProfileCorpora_041(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 41) }
func TestVolumeProfileCorpora_042(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 42) }
func TestVolumeProfileCorpora_043(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 43) }
func TestVolumeProfileCorpora_044(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 44) }
func TestVolumeProfileCorpora_045(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 45) }
func TestVolumeProfileCorpora_046(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 46) }
func TestVolumeProfileCorpora_047(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 47) }
func TestVolumeProfileCorpora_048(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 48) }
func TestVolumeProfileCorpora_049(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 49) }
func TestVolumeProfileCorpora_050(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 50) }
func TestVolumeProfileCorpora_051(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 51) }
func TestVolumeProfileCorpora_052(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 52) }
func TestVolumeProfileCorpora_053(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 53) }
func TestVolumeProfileCorpora_054(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 54) }
func TestVolumeProfileCorpora_055(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 55) }
func TestVolumeProfileCorpora_056(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 56) }
func TestVolumeProfileCorpora_057(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 57) }
func TestVolumeProfileCorpora_058(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 58) }
func TestVolumeProfileCorpora_059(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 59) }
func TestVolumeProfileCorpora_060(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 60) }
func TestVolumeProfileCorpora_061(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 61) }
func TestVolumeProfileCorpora_062(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 62) }
func TestVolumeProfileCorpora_063(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 63) }
func TestVolumeProfileCorpora_064(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 64) }
func TestVolumeProfileCorpora_065(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 65) }
func TestVolumeProfileCorpora_066(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 66) }
func TestVolumeProfileCorpora_067(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 67) }
func TestVolumeProfileCorpora_068(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 68) }
func TestVolumeProfileCorpora_069(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 69) }
func TestVolumeProfileCorpora_070(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 70) }
func TestVolumeProfileCorpora_071(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 71) }
func TestVolumeProfileCorpora_072(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 72) }
func TestVolumeProfileCorpora_073(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 73) }
func TestVolumeProfileCorpora_074(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 74) }
func TestVolumeProfileCorpora_075(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 75) }
func TestVolumeProfileCorpora_076(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 76) }
func TestVolumeProfileCorpora_077(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 77) }
func TestVolumeProfileCorpora_078(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 78) }
func TestVolumeProfileCorpora_079(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 79) }
func TestVolumeProfileCorpora_080(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 80) }
func TestVolumeProfileCorpora_081(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 81) }
func TestVolumeProfileCorpora_082(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 82) }
func TestVolumeProfileCorpora_083(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 83) }
func TestVolumeProfileCorpora_084(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 84) }
func TestVolumeProfileCorpora_085(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 85) }
func TestVolumeProfileCorpora_086(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 86) }
func TestVolumeProfileCorpora_087(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 87) }
func TestVolumeProfileCorpora_088(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 88) }
func TestVolumeProfileCorpora_089(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 89) }
func TestVolumeProfileCorpora_090(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 90) }
func TestVolumeProfileCorpora_091(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 91) }
func TestVolumeProfileCorpora_092(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 92) }
func TestVolumeProfileCorpora_093(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 93) }
func TestVolumeProfileCorpora_094(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 94) }
func TestVolumeProfileCorpora_095(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 95) }
func TestVolumeProfileCorpora_096(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 96) }
func TestVolumeProfileCorpora_097(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 97) }
func TestVolumeProfileCorpora_098(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 98) }
func TestVolumeProfileCorpora_099(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 99) }
func TestVolumeProfileCorpora_100(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 100) }
func TestVolumeProfileCorpora_101(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 101) }
func TestVolumeProfileCorpora_102(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 102) }
func TestVolumeProfileCorpora_103(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 103) }
func TestVolumeProfileCorpora_104(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 104) }
func TestVolumeProfileCorpora_105(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 105) }
func TestVolumeProfileCorpora_106(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 106) }
func TestVolumeProfileCorpora_107(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 107) }
func TestVolumeProfileCorpora_108(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 108) }
func TestVolumeProfileCorpora_109(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 109) }
func TestVolumeProfileCorpora_110(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 110) }
func TestVolumeProfileCorpora_111(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 111) }
func TestVolumeProfileCorpora_112(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 112) }
func TestVolumeProfileCorpora_113(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 113) }
func TestVolumeProfileCorpora_114(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 114) }
func TestVolumeProfileCorpora_115(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 115) }
func TestVolumeProfileCorpora_116(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 116) }
func TestVolumeProfileCorpora_117(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 117) }
func TestVolumeProfileCorpora_118(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 118) }
func TestVolumeProfileCorpora_119(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 119) }
func TestVolumeProfileCorpora_120(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 120) }
func TestVolumeProfileCorpora_121(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 121) }
func TestVolumeProfileCorpora_122(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 122) }
func TestVolumeProfileCorpora_123(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 123) }
func TestVolumeProfileCorpora_124(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 124) }
func TestVolumeProfileCorpora_125(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 125) }
func TestVolumeProfileCorpora_126(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 126) }
func TestVolumeProfileCorpora_127(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 127) }
func TestVolumeProfileCorpora_128(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 128) }
func TestVolumeProfileCorpora_129(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 129) }
func TestVolumeProfileCorpora_130(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 130) }
func TestVolumeProfileCorpora_131(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 131) }
func TestVolumeProfileCorpora_132(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 132) }
func TestVolumeProfileCorpora_133(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 133) }
func TestVolumeProfileCorpora_134(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 134) }
func TestVolumeProfileCorpora_135(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 135) }
func TestVolumeProfileCorpora_136(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 136) }
func TestVolumeProfileCorpora_137(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 137) }
func TestVolumeProfileCorpora_138(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 138) }
func TestVolumeProfileCorpora_139(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 139) }
func TestVolumeProfileCorpora_140(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 140) }
func TestVolumeProfileCorpora_141(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 141) }
func TestVolumeProfileCorpora_142(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 142) }
func TestVolumeProfileCorpora_143(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 143) }
func TestVolumeProfileCorpora_144(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 144) }
func TestVolumeProfileCorpora_145(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 145) }
func TestVolumeProfileCorpora_146(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 146) }
func TestVolumeProfileCorpora_147(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 147) }
func TestVolumeProfileCorpora_148(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 148) }
func TestVolumeProfileCorpora_149(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 149) }
func TestVolumeProfileCorpora_150(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 150) }
func TestVolumeProfileCorpora_151(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 151) }
func TestVolumeProfileCorpora_152(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 152) }
func TestVolumeProfileCorpora_153(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 153) }
func TestVolumeProfileCorpora_154(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 154) }
func TestVolumeProfileCorpora_155(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 155) }
func TestVolumeProfileCorpora_156(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 156) }
func TestVolumeProfileCorpora_157(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 157) }
func TestVolumeProfileCorpora_158(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 158) }
func TestVolumeProfileCorpora_159(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 159) }
func TestVolumeProfileCorpora_160(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 160) }
func TestVolumeProfileCorpora_161(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 161) }
func TestVolumeProfileCorpora_162(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 162) }
func TestVolumeProfileCorpora_163(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 163) }
func TestVolumeProfileCorpora_164(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 164) }
func TestVolumeProfileCorpora_165(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 165) }
func TestVolumeProfileCorpora_166(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 166) }
func TestVolumeProfileCorpora_167(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 167) }
func TestVolumeProfileCorpora_168(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 168) }
func TestVolumeProfileCorpora_169(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 169) }
func TestVolumeProfileCorpora_170(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 170) }
func TestVolumeProfileCorpora_171(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 171) }
func TestVolumeProfileCorpora_172(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 172) }
func TestVolumeProfileCorpora_173(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 173) }
func TestVolumeProfileCorpora_174(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 174) }
func TestVolumeProfileCorpora_175(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 175) }
func TestVolumeProfileCorpora_176(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 176) }
func TestVolumeProfileCorpora_177(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 177) }
func TestVolumeProfileCorpora_178(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 178) }
func TestVolumeProfileCorpora_179(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 179) }
func TestVolumeProfileCorpora_180(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 180) }
func TestVolumeProfileCorpora_181(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 181) }
func TestVolumeProfileCorpora_182(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 182) }
func TestVolumeProfileCorpora_183(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 183) }
func TestVolumeProfileCorpora_184(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 184) }
func TestVolumeProfileCorpora_185(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 185) }
func TestVolumeProfileCorpora_186(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 186) }
func TestVolumeProfileCorpora_187(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 187) }
func TestVolumeProfileCorpora_188(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 188) }
func TestVolumeProfileCorpora_189(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 189) }
func TestVolumeProfileCorpora_190(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 190) }
func TestVolumeProfileCorpora_191(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 191) }
func TestVolumeProfileCorpora_192(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 192) }
func TestVolumeProfileCorpora_193(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 193) }
func TestVolumeProfileCorpora_194(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 194) }
func TestVolumeProfileCorpora_195(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 195) }
func TestVolumeProfileCorpora_196(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 196) }
func TestVolumeProfileCorpora_197(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 197) }
func TestVolumeProfileCorpora_198(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 198) }
func TestVolumeProfileCorpora_199(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 199) }
func TestVolumeProfileCorpora_200(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 200) }
func TestVolumeProfileCorpora_201(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 201) }
func TestVolumeProfileCorpora_202(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 202) }
func TestVolumeProfileCorpora_203(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 203) }
func TestVolumeProfileCorpora_204(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 204) }
func TestVolumeProfileCorpora_205(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 205) }
func TestVolumeProfileCorpora_206(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 206) }
func TestVolumeProfileCorpora_207(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 207) }
func TestVolumeProfileCorpora_208(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 208) }
func TestVolumeProfileCorpora_209(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 209) }
func TestVolumeProfileCorpora_210(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 210) }
func TestVolumeProfileCorpora_211(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 211) }
func TestVolumeProfileCorpora_212(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 212) }
func TestVolumeProfileCorpora_213(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 213) }
func TestVolumeProfileCorpora_214(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 214) }
func TestVolumeProfileCorpora_215(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 215) }
func TestVolumeProfileCorpora_216(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 216) }
func TestVolumeProfileCorpora_217(t *testing.T) { t.Parallel(); volumeProfileCorporaRun(t, 217) }

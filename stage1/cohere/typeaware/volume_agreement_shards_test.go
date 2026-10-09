package typeaware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

const volumeAgreementShards = 32

func volumeAgreementChanges() []struct{ name, path, from, to string } {
	return []struct{ name, path, from, to string }{
		{"assignable-types", "bridge/tsgo/checker/facts.go", "out.yes(checker.Checker_isTypeAssignableTo(c, selected[0], selected[1]))", "out.yes(checker.Checker_isTypeAssignableTo(c, selected[1], selected[0]))"},
		{"widened-shape", "bridge/tsgo/checker/facts.go", "t = checker.Checker_getWidenedType(c, t)", "// Mutant keeps the fresh type."},
		{"enum-types", "bridge/tsgo/checker/facts.go", "base = c.GetTypeAtLocation(symbol.ValueDeclaration.Parent)", "base = part"},
		{"scope-locals", "bridge/tsgo/checker/scopes.go", "out.text(name)", "out.text(name + \"wrong\")"},
		{"call-returns", "bridge/tsgo/checker/facts.go", "roots = append(roots, g.add(c.GetReturnTypeOfSignature(signature)))", "_ = signature; roots = append(roots, g.add(checker.Checker_numberType(c)))"},
		{"property-shape", "bridge/tsgo/checker/facts.go", "g.add(c.GetTypeOfSymbolAtLocation(property, node))", "g.add(checker.Checker_numberType(c))"},
		{"contextual-shape", "bridge/tsgo/checker/facts.go", "t := checker.Checker_getContextualType(c, node, checker.ContextFlagsNone)", "t := c.GetTypeAtLocation(node)"},
		{"symbol-origin", "bridge/tsgo/checker/facts.go", "file = f.FileName()", "file = source.FileName()"},
		{"type-origin", "bridge/tsgo/checker/metadata.go", "out.text(symbol.Name)", "out.text(symbol.Name + \"wrong\")"},
		{"property-info", "bridge/tsgo/checker/metadata.go", "out.text(strings.TrimPrefix(declaration.Kind.String(), \"Kind\"))", "out.text(\"PropertySignature\")"},
		{"call-count", "bridge/tsgo/checker/facts.go", "out.number(uint64(len(c.GetSignaturesOfType(subject, checker.SignatureKindCall))))", "out.number(0)"},
		{"call-parameters", "bridge/tsgo/checker/facts.go", "g.add(checker.Checker_getApparentType(c, c.GetTypeOfSymbolAtLocation(params[0], node)))", "g.add(checker.Checker_numberType(c))"},
		{"apparent-shape", "bridge/tsgo/checker/facts.go", "g.add(checker.Checker_getApparentType(c, subject))", "g.add(checker.Checker_numberType(c))"},
		{"base-shapes", "bridge/tsgo/checker/facts.go", "roots = append(roots, g.add(base))", "_ = base"},
	}
}

const volumeAgreementReleasedSource = `import { programArguments, tsgoProgram, tsgoInspect, tsgoRelease } from 'adamic';
const args=programArguments(); const path=args[1] ?? ''; const program=tsgoProgram(args[0] ?? '',[path]);tsgoRelease(program);
console.log(tsgoInspect(program,path,0,1,'Identifier','call-returns'));
`

func volumeAgreementKeys() []string {
	keys := []string{"controls", "controls-asan", "released-handle"}
	for _, c := range volumeAgreementChanges() {
		keys = append(keys, c.name)
	}
	return keys
}
func volumeAgreementOwner(key string) int {
	sum := sha256.Sum256([]byte(key))
	return int(sum[0]) % volumeAgreementShards
}
func volumeAgreementIDs(shard int) []string {
	var keys []string
	for _, key := range volumeAgreementKeys() {
		if volumeAgreementOwner(key) == shard {
			keys = append(keys, key)
		}
	}
	return keys
}
func volumeAgreementSurvived(got, want result) error {
	if got.err != nil || len(got.stderr) != 0 {
		return fmt.Errorf("mutant execution: %v %s", got.err, got.stderr)
	}
	if bytes.Equal(got.stdout, want.stdout) {
		return fmt.Errorf("checker question mutant survived")
	}
	return nil
}
func TestVolumeAgreementAndMutantsUnion(t *testing.T) {
	t.Parallel()
	seen := map[string]int{}
	caught := 0
	for shard := 0; shard < volumeAgreementShards; shard++ {
		for _, key := range volumeAgreementIDs(shard) {
			seen[key]++
			if key == "property-shape" && volumeAgreementSurvived(result{stdout: []byte("planted")}, result{stdout: []byte("planted")}) != nil {
				caught++
				t.Logf("planted survivor caught exactly by TestVolumeAgreementAndMutants_%03d", shard)
			}
		}
	}
	for _, key := range volumeAgreementKeys() {
		if seen[key] != 1 {
			t.Fatalf("case %s occurs %d times", key, seen[key])
		}
	}
	if len(seen) != len(volumeAgreementKeys()) || caught != 1 {
		t.Fatalf("union=%d planted catches=%d", len(seen), caught)
	}
	t.Logf("counted union: %d cases over %d shards", len(seen), volumeAgreementShards)
}

// Product declarations and selected shards use this same recipe. Preparation
// has no test-side deadline; the build phase owns cache misses.
func volumeAgreementInputs(name string, sanitize bool) buildcache.Inputs {
	inputs := volumeGuardInputs(context.Background(), "agreement-"+name, sanitize)
	for _, key := range []string{"ADAMIC_NATIVE_SPLIT", "ADAMIC_NATIVE_JOBS", "CPATH", "C_INCLUDE_PATH", "LIBRARY_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET", "GOOS", "GOARCH", "CGO_ENABLED", "GOTOOLCHAIN"} {
		inputs.Flags = append(inputs.Flags, key+"="+os.Getenv(key))
	}
	return inputs
}
func volumeAgreementSource(h *volumeGuardHarness, name string) string {
	entry := filepath.Join(h.repository, "stage1/cohere/typeaware/volume_suite.ts")
	if name == "released-source" {
		entry = h.write("released-inspect.ts", volumeAgreementReleasedSource)
	}
	inputs := volumeAgreementInputs(name, false)
	inputs.Flags = append(inputs.Flags, name)
	dir := buildcache.Product(h.t, inputs, func(dir string) error {
		loaded, err := load.Load([]string{entry})
		if err != nil {
			return err
		}
		loaded.EnableTSGo()
		program, err := lower.Lower(context.Background(), loaded)
		if err != nil {
			return err
		}
		source, err := native.TSGoC(program)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "source.c"), []byte(source), 0444)
	})
	return filepath.Join(dir, "source.c")
}
func volumeAgreementNative(h *volumeGuardHarness, name, archive, source string, sanitize bool) string {
	inputs := volumeAgreementInputs(name, sanitize)
	inputs.Flags = append(inputs.Flags, "BuildSplitTSGo Jobs=4")
	for _, path := range []string{archive, source} {
		sum, err := typeSymbolFileHash(path)
		if err != nil {
			h.t.Fatal(err)
		}
		inputs.Flags = append(inputs.Flags, sum)
	}
	dir := buildcache.Product(h.t, inputs, func(dir string) error {
		data, err := os.ReadFile(source)
		if err != nil {
			return err
		}
		return native.BuildSplitTSGo(string(data), filepath.Join(dir, "native"), archive, native.Options{Sanitize: sanitize, Split: true, Jobs: 4})
	})
	return filepath.Join(dir, "native")
}
func volumeAgreementProduct(h *volumeGuardHarness, name string) string {
	switch name {
	case "oracle", "archive-checker", "archive-checker-asan":
		return volumeGuardProduct(h, name)
	case "volume-source", "released-source":
		return volumeAgreementSource(h, name)
	case "volume", "volume-asan":
		archiveName := "archive-checker"
		if name == "volume-asan" {
			archiveName += "-asan"
		}
		return volumeAgreementNative(h, name, volumeAgreementFetch(h.t, archiveName), volumeAgreementFetch(h.t, "volume-source"), name == "volume-asan")
	}
	archiveName := strings.TrimSuffix(name, "-native")
	var overlay string
	if archiveName == "released-registry" {
		overlay = h.overlay(archiveName, "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant keeps released program live.")
	} else if archiveName != "released-handle" {
		for _, c := range volumeAgreementChanges() {
			if c.name == archiveName {
				overlay = h.overlay(c.name, c.path, c.from, c.to)
			}
		}
		if overlay == "" {
			h.t.Fatalf("unknown product %s", name)
		}
	}
	archive := ""
	if overlay == "" {
		archive = volumeAgreementFetch(h.t, "archive-checker")
	} else if strings.HasSuffix(name, "-native") {
		archive = volumeAgreementFetch(h.t, archiveName)
	} else {
		inputs := volumeAgreementInputs("archive-"+archiveName, false)
		inputs.Flags = append(inputs.Flags, archiveName)
		dir := buildcache.Product(h.t, inputs, func(dir string) error {
			command, cancel := volumeGuardCommand(context.Background(), "go", "build", "-buildmode=c-archive", "-overlay", overlay, "-o", filepath.Join(dir, "checker.a"), "./bridge/tsgo/archive")
			defer cancel()
			command.Dir = h.repository
			output, err := volumeGuardOutput(command)
			if err != nil {
				return fmt.Errorf("archive %s: %w %s", archiveName, err, output)
			}
			return nil
		})
		archive = filepath.Join(dir, "checker.a")
	}
	if !strings.HasSuffix(name, "-native") {
		return archive
	}
	sourceName := "volume-source"
	if archiveName == "released-handle" || archiveName == "released-registry" {
		sourceName = "released-source"
	}
	return volumeAgreementNative(h, name, archive, volumeAgreementFetch(h.t, sourceName), false)
}

type volumeAgreementSlot struct {
	once sync.Once
	path string
}

var volumeAgreementSlots sync.Map

func volumeAgreementFetch(t *testing.T, name string) string {
	value, _ := volumeAgreementSlots.LoadOrStore(name, &volumeAgreementSlot{})
	slot := value.(*volumeAgreementSlot)
	slot.once.Do(func() { slot.path = volumeAgreementProduct(volumeGuardPreparationHarness(t), name) })
	if slot.path == "" {
		t.Fatalf("preparation failed: %s", name)
	}
	return slot.path
}
func volumeAgreementManifest(h *volumeGuardHarness) string {
	var paths []string
	for i, source := range typeSymbolCases() {
		paths = append(paths, h.write(fmt.Sprintf("control-%03d.ts", i), source))
	}
	return h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
}
func runVolumeAgreementShard(t *testing.T, shard int) {
	keys := volumeAgreementIDs(shard)
	if len(keys) == 0 {
		t.Log("own work=0s (empty hash bucket)")
		return
	}
	products := map[string]string{}
	for _, key := range keys {
		needs := []string{"oracle"}
		switch key {
		case "controls":
			needs = append(needs, "volume")
		case "controls-asan":
			needs = append(needs, "volume-asan")
		case "released-handle":
			needs = []string{"released-handle-native", "released-registry-native"}
		default:
			needs = append(needs, key+"-native")
		}
		for _, name := range needs {
			if products[name] == "" {
				products[name] = volumeAgreementFetch(t, name)
			}
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	started := time.Now()
	defer func() {
		t.Logf("own work=%.6fs", time.Since(started).Seconds())
		if ctx.Err() != nil {
			t.Error("shard own work exceeded 60s")
		}
	}()
	h := volumeGuardPreparationHarness(t)
	h.ctx = ctx
	config := filepath.Join(h.repository, "stage1/cohere/typeaware/testdata/tsconfig.json")
	manifest := volumeAgreementManifest(h)
	sort.Strings(keys)
	for _, key := range keys {
		switch key {
		case "controls", "controls-asan":
			name := "volume"
			if key == "controls-asan" {
				name = "volume-asan"
			}
			truth := volumeGuardCompare(h, key, products["oracle"], products[name], config, manifest)
			if key == "controls" {
				for _, rule := range volumeRules {
					if !bytes.Contains(truth.stdout, []byte("@typescript-eslint/"+strings.ReplaceAll(rule, "_", "-"))) {
						t.Fatalf("rule %s lacks a positive control", rule)
					}
				}
			}
		case "released-handle":
			probe := h.write("probe.ts", "x;\n")
			got := volumeGuardRun(h, "released-handle", exec.Command(products["released-handle-native"], config, probe))
			if exit, ok := got.err.(*exec.ExitError); !ok || exit.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
				t.Fatalf("released program escaped: %v %s", got.err, got.stderr)
			}
			volumeGuardMust(h, "released-registry", exec.Command(products["released-registry-native"], config, probe))
		default:
			truth := volumeGuardMust(h, key+"-truth", exec.Command(products["oracle"], config, manifest))
			got := volumeGuardMust(h, key+"-run", exec.Command(products[key+"-native"], config, manifest))
			if err := volumeAgreementSurvived(got, truth); err != nil {
				t.Fatalf("%s: %v", key, err)
			}
			t.Logf("%s caught at byte %d", key, firstDifference(got.stdout, truth.stdout))
		}
	}
}

func TestVolumeAgreementAndMutants_000(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 0) }
func TestVolumeAgreementAndMutants_001(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 1) }
func TestVolumeAgreementAndMutants_002(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 2) }
func TestVolumeAgreementAndMutants_003(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 3) }
func TestVolumeAgreementAndMutants_004(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 4) }
func TestVolumeAgreementAndMutants_005(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 5) }
func TestVolumeAgreementAndMutants_006(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 6) }
func TestVolumeAgreementAndMutants_007(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 7) }
func TestVolumeAgreementAndMutants_008(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 8) }
func TestVolumeAgreementAndMutants_009(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 9) }
func TestVolumeAgreementAndMutants_010(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 10) }
func TestVolumeAgreementAndMutants_011(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 11) }
func TestVolumeAgreementAndMutants_012(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 12) }
func TestVolumeAgreementAndMutants_013(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 13) }
func TestVolumeAgreementAndMutants_014(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 14) }
func TestVolumeAgreementAndMutants_015(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 15) }
func TestVolumeAgreementAndMutants_016(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 16) }
func TestVolumeAgreementAndMutants_017(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 17) }
func TestVolumeAgreementAndMutants_018(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 18) }
func TestVolumeAgreementAndMutants_019(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 19) }
func TestVolumeAgreementAndMutants_020(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 20) }
func TestVolumeAgreementAndMutants_021(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 21) }
func TestVolumeAgreementAndMutants_022(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 22) }
func TestVolumeAgreementAndMutants_023(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 23) }
func TestVolumeAgreementAndMutants_024(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 24) }
func TestVolumeAgreementAndMutants_025(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 25) }
func TestVolumeAgreementAndMutants_026(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 26) }
func TestVolumeAgreementAndMutants_027(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 27) }
func TestVolumeAgreementAndMutants_028(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 28) }
func TestVolumeAgreementAndMutants_029(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 29) }
func TestVolumeAgreementAndMutants_030(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 30) }
func TestVolumeAgreementAndMutants_031(t *testing.T) { t.Parallel(); runVolumeAgreementShard(t, 31) }

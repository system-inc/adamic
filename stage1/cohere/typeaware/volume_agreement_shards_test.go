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
		{
			"assignable-types",
			"bridge/tsgo/checker/facts.go",
			"out.yes(checker.Checker_isTypeAssignableTo(c, selected[0], selected[1]))",
			"out.yes(checker.Checker_isTypeAssignableTo(c, selected[1], selected[0]))",
		},
		{
			"widened-shape",
			"bridge/tsgo/checker/facts.go",
			"t = checker.Checker_getWidenedType(c, t)",
			"// Mutant keeps the fresh type.",
		},
		{
			"enum-types",
			"bridge/tsgo/checker/facts.go",
			"base = c.GetTypeAtLocation(symbol.ValueDeclaration.Parent)",
			"base = part",
		},
		{"scope-locals", "bridge/tsgo/checker/scopes.go", "out.text(name)", "out.text(name + \"wrong\")"},
		{
			"call-returns",
			"bridge/tsgo/checker/facts.go",
			"roots = append(roots, g.add(c.GetReturnTypeOfSignature(signature)))",
			"_ = signature; roots = append(roots, g.add(checker.Checker_numberType(c)))",
		},
		{
			"property-shape",
			"bridge/tsgo/checker/facts.go",
			"g.add(c.GetTypeOfSymbolAtLocation(property, node))",
			"g.add(checker.Checker_numberType(c))",
		},
		{
			"contextual-shape",
			"bridge/tsgo/checker/facts.go",
			"t := checker.Checker_getContextualType(c, node, checker.ContextFlagsNone)",
			"t := c.GetTypeAtLocation(node)",
		},
		{"symbol-origin", "bridge/tsgo/checker/facts.go", "file = f.FileName()", "file = source.FileName()"},
		{
			"type-origin",
			"bridge/tsgo/checker/metadata.go",
			"out.text(symbol.Name)",
			"out.text(symbol.Name + \"wrong\")",
		},
		{
			"property-info",
			"bridge/tsgo/checker/metadata.go",
			"out.text(strings.TrimPrefix(declaration.Kind.String(), \"Kind\"))",
			"out.text(\"PropertySignature\")",
		},
		{
			"call-count",
			"bridge/tsgo/checker/facts.go",
			"out.number(uint64(len(c.GetSignaturesOfType(subject, checker.SignatureKindCall))))",
			"out.number(0)",
		},
		{
			"call-parameters",
			"bridge/tsgo/checker/facts.go",
			"g.add(checker.Checker_getApparentType(c, c.GetTypeOfSymbolAtLocation(params[0], node)))",
			"g.add(checker.Checker_numberType(c))",
		},
		{
			"apparent-shape",
			"bridge/tsgo/checker/facts.go",
			"g.add(checker.Checker_getApparentType(c, subject))",
			"g.add(checker.Checker_numberType(c))",
		},
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
func TestVolumeAgreementAndMutantsUnion(test *testing.T) {
	test.Parallel()
	seen := map[string]int{}
	caught := 0
	for shard := 0; shard < volumeAgreementShards; shard++ {
		for _, key := range volumeAgreementIDs(shard) {
			seen[key]++
			if key == "property-shape" &&
				volumeAgreementSurvived(result{stdout: []byte("planted")}, result{stdout: []byte("planted")}) != nil {
				caught++
				test.Logf("planted survivor caught exactly by TestVolumeAgreementAndMutants_%03d", shard)
			}
		}
	}
	for _, key := range volumeAgreementKeys() {
		if seen[key] != 1 {
			test.Fatalf("case %s occurs %d times", key, seen[key])
		}
	}
	if len(seen) != len(volumeAgreementKeys()) || caught != 1 {
		test.Fatalf("union=%d planted catches=%d", len(seen), caught)
	}
	test.Logf("counted union: %d cases over %d shards", len(seen), volumeAgreementShards)
}

// Product declarations and selected shards use this same recipe. Preparation
// has no test-side deadline; the build phase owns cache misses.
func volumeAgreementInputs(name string, sanitize bool) buildcache.Inputs {
	inputs := volumeGuardInputs(context.Background(), "agreement-"+name, sanitize)
	for _, key := range []string{
		"ADAMIC_NATIVE_SPLIT",
		"CPATH",
		"C_INCLUDE_PATH",
		"LIBRARY_PATH",
		"SDKROOT",
		"MACOSX_DEPLOYMENT_TARGET",
		"GOOS",
		"GOARCH",
		"CGO_ENABLED",
	} {
		inputs.Flags = append(inputs.Flags, key+"="+os.Getenv(key))
	}
	return inputs
}

// The oracle retains every original program root while its walk visits only
// the shard manifest. Its unchanged production rules still decide the findings.
// The adapter is a Mutation of oracle_volume.go, so the oracle is a GoBuildIn
// product as the package's other oracles are (typeAwareOverlayOracle).
func volumeAgreementOracle(h *volumeGuardHarness) string {
	from := "program := compiler.NewProgram(compiler.ProgramOptions{Config: config.WithFileNames(roots), Host: host, SingleThreaded: core.TSTrue})"
	to := `programRoots := roots
 if len(args)>3 {
  all,err:=os.ReadFile(args[3]); if err!=nil { panic(err) }
  programRoots=nil
  for _,path:=range strings.Split(string(all),"\n") { if path!="" { programRoots=append(programRoots,tspath.RootedFilePathFromAbsolute(filepath.ToSlash(path))) } }
  for _,path:=range config.FileNames() { if strings.HasSuffix(path.AsString(),".d.ts") && !slices.Contains(programRoots,path) { programRoots=append(programRoots,path) } }
 }
 program := compiler.NewProgram(compiler.ProgramOptions{Config: config.WithFileNames(programRoots), Host: host, SingleThreaded: core.TSTrue})`
	return typeAwareOverlayOracle(
		h.harness,
		"oracle",
		"adamic_volume_shard_oracle.go",
		"oracle_volume.go",
		buildcache.Mutation{File: "stage1/cohere/typeaware/testdata/oracle_volume.go", Before: from, After: to},
	)
}
func volumeAgreementSource(h *volumeGuardHarness, name string) string {
	entry := filepath.Join(h.repository, "stage1/cohere/typeaware/volume_suite.ts")
	if name == "volume-source" {
		data, readError := os.ReadFile(entry)
		if readError != nil {
			h.t.Fatal(readError)
		}
		text := string(data)
		from := "const program = tsgoProgram(config, paths);"
		to := `const rootsManifest = args[3] ? readTextFile(args[3]) : manifest;
if(rootsManifest.kind === 'Error') { panic(rootsManifest.message); }
const program = tsgoProgram(config, rootsManifest.text.split('\n').filter((path) => path !== ''));`
		if strings.Count(text, from) != 1 {
			h.t.Fatal("nonunique volume program-root adapter")
		}
		text = strings.Replace(text, from, to, 1)
		for _, line := range strings.Split(text, "\n") {
			if !strings.Contains(line, "from './") && !strings.Contains(line, "from '../") {
				continue
			}
			pieces := strings.Split(line, "from '")
			path := strings.Split(pieces[1], "'")[0]
			absolute := filepath.ToSlash(filepath.Join(filepath.Dir(entry), path))
			text = strings.Replace(text, "from '"+path+"'", "from '"+absolute+"'", 1)
		}
		entry = h.write("volume-suite-with-program-roots.ts", text)
	}

	if name == "released-source" {
		entry = h.write("released-inspect.ts", volumeAgreementReleasedSource)
	}
	inputs := volumeAgreementInputs(name, false)
	inputs.Flags = append(inputs.Flags, name)
	dir := buildcache.Product(h.t, inputs, func(dir string) error {
		loaded, loadError := load.Load([]string{entry})
		if loadError != nil {
			return loadError
		}
		loaded.EnableTSGo()
		program, lowerError := lower.Lower(context.Background(), loaded)
		if lowerError != nil {
			return lowerError
		}
		source, tsGoCError := native.TSGoC(program)
		if tsGoCError != nil {
			return tsGoCError
		}
		return os.WriteFile(filepath.Join(dir, "source.c"), []byte(source), 0444)
	})
	return filepath.Join(dir, "source.c")
}
func volumeAgreementNative(h *volumeGuardHarness, name, archive, source string, sanitize bool) string {
	inputs := volumeAgreementInputs(name, sanitize)
	inputs.Flags = append(inputs.Flags, "BuildSplitTSGo Jobs=4")
	for _, path := range []string{archive, source} {
		sum, typeSymbolFileHashError := typeSymbolFileHash(path)
		if typeSymbolFileHashError != nil {
			h.t.Fatal(typeSymbolFileHashError)
		}
		inputs.Flags = append(inputs.Flags, sum)
	}
	dir := buildcache.Product(h.t, inputs, func(dir string) error {
		data, readError := os.ReadFile(source)
		if readError != nil {
			return readError
		}
		return native.BuildSplitTSGo(
			string(data),
			filepath.Join(dir, "native"),
			archive,
			native.Options{Sanitize: sanitize, Split: true, Jobs: 4},
		)
	})
	return filepath.Join(dir, "native")
}
func volumeAgreementProduct(h *volumeGuardHarness, name string) string {
	switch name {
	case "oracle":
		return volumeAgreementOracle(h)
	case "archive-checker", "archive-checker-asan":
		return volumeGuardProduct(h, name)
	case "volume-source", "released-source":
		return volumeAgreementSource(h, name)
	case "volume", "volume-asan":
		archiveName := "archive-checker"
		if name == "volume-asan" {
			archiveName += "-asan"
		}
		return volumeAgreementNative(
			h,
			name,
			volumeAgreementFetch(h.t, archiveName),
			volumeAgreementFetch(h.t, "volume-source"),
			name == "volume-asan",
		)
	}
	archiveName := strings.TrimSuffix(name, "-native")
	var mutation *buildcache.Mutation
	if archiveName == "released-registry" {
		mutation = &buildcache.Mutation{
			File:   checkerArchive,
			Before: "delete(programs.live, uint64(handle))",
			After:  "// Mutant keeps released program live.",
		}
	} else if archiveName != "released-handle" {
		for _, c := range volumeAgreementChanges() {
			if c.name == archiveName {
				mutation = &buildcache.Mutation{File: c.path, Before: c.from, After: c.to}
			}
		}
		if mutation == nil {
			h.t.Fatalf("unknown product %s", name)
		}
	}
	archive := ""
	if mutation == nil {
		archive = volumeAgreementFetch(h.t, "archive-checker")
	} else if strings.HasSuffix(name, "-native") {
		archive = volumeAgreementFetch(h.t, archiveName)
	} else {
		archive = plantedCheckerArchive(h.t, "agreement-"+archiveName+"-checker.a", *mutation)
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

func volumeAgreementFetch(test *testing.T, name string) string {
	value, _ := volumeAgreementSlots.LoadOrStore(name, &volumeAgreementSlot{})
	slot := value.(*volumeAgreementSlot)
	slot.once.Do(func() { slot.path = volumeAgreementProduct(volumeGuardPreparationHarness(test), name) })
	if slot.path == "" {
		test.Fatalf("preparation failed: %s", name)
	}
	return slot.path
}
func volumeAgreementManifest(h *volumeGuardHarness) string {
	var paths []string
	cases := typeSymbolCases()
	for i, source := range cases {
		name := fmt.Sprintf("control-%03d.ts", i)
		if i == len(cases)-2 {
			name = "native-globals.d.ts"
		}
		if i == len(cases)-1 {
			name = "native-console.ts"
		}
		paths = append(paths, h.write(name, source))
	}
	return h.write("controls.manifest", strings.Join(paths, "\n")+"\n")
}
func runVolumeAgreementShard(test *testing.T, shard int) {
	keys := volumeAgreementIDs(shard)
	if len(keys) == 0 {
		test.Log("own work=0s (empty hash bucket)")
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
				products[name] = volumeAgreementFetch(test, name)
			}
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	started := time.Now()
	defer func() {
		test.Logf("own work=%.6fs", time.Since(started).Seconds())
		if ctx.Err() != nil {
			test.Error("shard own work exceeded 60s")
		}
	}()
	h := volumeGuardPreparationHarness(test)
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
						test.Fatalf("rule %s lacks a positive control", rule)
					}
				}
			}
		case "released-handle":
			probe := h.write("probe.ts", "x;\n")
			got := volumeGuardRun(h, "released-handle", exec.Command(products["released-handle-native"], config, probe))
			if exit, ok := got.err.(*exec.ExitError); !ok ||
				exit.ExitCode() != 70 ||
				string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
				test.Fatalf("released program escaped: %v %s", got.err, got.stderr)
			}
			volumeGuardMust(h, "released-registry", exec.Command(products["released-registry-native"], config, probe))
		default:
			truth := volumeGuardMust(h, key+"-truth", exec.Command(products["oracle"], config, manifest))
			got := volumeGuardMust(h, key+"-run", exec.Command(products[key+"-native"], config, manifest))
			if volumeAgreementSurvivedError := volumeAgreementSurvived(
				got,
				truth,
			); volumeAgreementSurvivedError != nil {
				test.Fatalf("%s: %v", key, volumeAgreementSurvivedError)
			}
			test.Logf("%s caught at byte %d", key, firstDifference(got.stdout, truth.stdout))
		}
	}
}

func TestVolumeAgreementAndMutants_000(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 0)
}
func TestVolumeAgreementAndMutants_001(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 1)
}
func TestVolumeAgreementAndMutants_002(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 2)
}
func TestVolumeAgreementAndMutants_003(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 3)
}
func TestVolumeAgreementAndMutants_004(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 4)
}
func TestVolumeAgreementAndMutants_005(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 5)
}
func TestVolumeAgreementAndMutants_006(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 6)
}
func TestVolumeAgreementAndMutants_007(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 7)
}
func TestVolumeAgreementAndMutants_008(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 8)
}
func TestVolumeAgreementAndMutants_009(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 9)
}
func TestVolumeAgreementAndMutants_010(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 10)
}
func TestVolumeAgreementAndMutants_011(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 11)
}
func TestVolumeAgreementAndMutants_012(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 12)
}
func TestVolumeAgreementAndMutants_013(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 13)
}
func TestVolumeAgreementAndMutants_014(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 14)
}
func TestVolumeAgreementAndMutants_015(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 15)
}
func TestVolumeAgreementAndMutants_016(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 16)
}
func TestVolumeAgreementAndMutants_017(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 17)
}
func TestVolumeAgreementAndMutants_018(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 18)
}
func TestVolumeAgreementAndMutants_019(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 19)
}
func TestVolumeAgreementAndMutants_020(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 20)
}
func TestVolumeAgreementAndMutants_021(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 21)
}
func TestVolumeAgreementAndMutants_022(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 22)
}
func TestVolumeAgreementAndMutants_023(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 23)
}
func TestVolumeAgreementAndMutants_024(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 24)
}
func TestVolumeAgreementAndMutants_025(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 25)
}
func TestVolumeAgreementAndMutants_026(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 26)
}
func TestVolumeAgreementAndMutants_027(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 27)
}
func TestVolumeAgreementAndMutants_028(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 28)
}
func TestVolumeAgreementAndMutants_029(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 29)
}
func TestVolumeAgreementAndMutants_030(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 30)
}
func TestVolumeAgreementAndMutants_031(test *testing.T) {
	test.Parallel()
	runVolumeAgreementShard(test, 31)
}

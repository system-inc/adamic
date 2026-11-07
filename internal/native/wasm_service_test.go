package native

import (
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

// The service oracle's claim is 100,000 generated requests through one instance, every response equal to Node's,
// linear memory flat and, counted, nothing live after each request. Its legs build nothing: each has the host generate
// the requests and Node's responses, and runs the modules or the sanitized native command, products the tree builds
// ahead (TestProduct_WASIServiceModules and TestProduct_WASIServiceNative). The legs are the plain and counted wasm
// runs, the sanitized native command's agreement in eight parts, and one test per mutant, one of which shrinks the
// request count and must be refused.
const serviceRequests = 100000

// The service's files, relative to the repository.
const (
	serviceSource  = "internal/native/wasm/service/service.a"
	serviceCommand = "internal/native/wasm/service/command.a"
	serviceHost    = "internal/native/wasm/service/host.mjs"
	serviceTests   = "internal/native/wasm_service_test.go"
)

// serviceParts are the request ranges, [start, end), the sanitized native command answers, one test each. A sanitized
// run of all 100,000 at once took about 57 s idle, past the gate's 60 s per-test budget on a loaded box, and a quarter
// of them about 14 s idle, past its 30 s aim there, so every request is still checked under ASan, but in eight
// instances of 12,500: the trade is that no sanitized native instance holds all 100,000, which the counted wasm leg's
// single-instance live 0 still proves.
var serviceParts = [][2]int{
	{0, 12500},
	{12500, 25000},
	{25000, 37500},
	{37500, 50000},
	{50000, 62500},
	{62500, 75000},
	{75000, 87500},
	{87500, 100000},
}

// A serviceMutation is a source mutant: a copy of service.a with before changed to after at its one location and
// appended added at its end, which the host, holding the original source as its oracle, must catch by caught.
type serviceMutation struct {
	name, before, after, appended, caught string
}

// serviceMutants are each built into a counted module of the modules product; retained-object appends the declaration
// its change needs.
var serviceMutants = []serviceMutation{
	{"response-byte", `service":"orders`, `service":"ordert`, "", "wasm semantic pin"},
	{
		"retained-object",
		"const parser = new Parser(request);",
		"const parser = new Parser(request); retained.push(parser);",
		"\nconst retained: Parser[] = [];\n",
		"retained a live allocation",
	},
	{"unicode-escape", "String.fromCharCode(code)", "String.fromCharCode(code + 1)", "", "wasm semantic pin"},
}

// tileRequests refuses ranges that don't cover [0, total) exactly once, in order: a gap leaves requests unchecked and
// an overlap checks some twice while the claim counts them once.
func tileRequests(ranges [][2]int, total int) error {
	next := 0
	for index, span := range ranges {
		if span[0] != next {
			return fmt.Errorf("range %d starts at %d, want %d: a gap or an overlap", index, span[0], next)
		}
		if span[1] <= span[0] {
			return fmt.Errorf("range %d, [%d, %d), is empty", index, span[0], span[1])
		}
		next = span[1]
	}
	if next != total {
		return fmt.Errorf("ranges end at %d, want %d", next, total)
	}
	return nil
}

func serviceRepository(test *testing.T) string {
	test.Helper()
	repository, absolutePathError := filepath.Abs("../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return repository
}

// serviceOptedIn skips a leg unless the service oracle is opted in, and returns the repository.
func serviceOptedIn(test *testing.T) string {
	test.Helper()
	if os.Getenv("ADAMIC_TEST_WASI") != "1" {
		test.Skip("set ADAMIC_TEST_WASI=1, source cloud/setup.sh --wasi-sdk's environment, and use Node 24")
	}
	if os.Getenv("WASI_SYSROOT") == "" {
		test.Fatal("WASI_SYSROOT is required once the service oracle is opted in")
	}
	return serviceRepository(test)
}

// runServiceHost runs host over module with the original service source as its oracle, and returns its output. Module
// "-" generates the requests and Node's responses only, writing those in span, [start, end), into directory; a module
// serves every request, so its span must be all of them, and it writes nothing.
func runServiceHost(repository, host, module, directory, counted, regions string, span [2]int) ([]byte, error) {
	timeoutContext, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	command := programCommand(
		timeoutContext,
		repository,
		"node",
		"--disable-warning=ExperimentalWarning",
		host,
		module,
		filepath.Join(repository, serviceSource),
		filepath.Join(directory, "requests.jsonl"),
		filepath.Join(directory, "responses.jsonl"),
		counted,
		regions,
		strconv.Itoa(span[0]),
		strconv.Itoa(span[1]),
	)
	return command.CombinedOutput()
}

// serviceCompiledInputs keys a product the stage 0 compiler makes from the service's files, named name and built by
// compiler: the stage 0 compiler by its name key (its key computed without building it, as buildcache.Adamic asks for
// it), and the service's files. The compiler's own files are declared too, as a lowered product's are: a traced key
// reads the compiler as the product its build ran, but this machine's untraced key reads only declared files, and
// without them it would keep a product a changed compiler made.
func serviceCompiledInputs(test *testing.T, repository, name, compiler string) buildcache.Inputs {
	test.Helper()
	adamic, goBuildInInputsError := buildcache.GoBuildInInputs("", "adamic", "./cmd/adamic", nil, nil)
	if goBuildInInputsError != nil {
		test.Fatal(goBuildInInputsError)
	}
	return buildcache.Inputs{
		Name:      name,
		Files:     append([]string{serviceSource, serviceTests}, adamic.Files...),
		Flags:     []string{"compiler " + buildcache.NameKey(repository, adamic)},
		Toolchain: []string{buildcache.Tool(compiler, "--version")},
	}
}

// serviceModules is the directory of the modules the wasm legs and the mutants run, a product built ahead by
// TestProduct_WASIServiceModules: service.wasm, service-counted.wasm, a counted <name>.wasm for each of serviceMutants,
// and regions, "1" when the C the compiler generates for service.a ends statement regions and "0" otherwise. Each leg
// used to build its own, and each mutant a whole counted build.
func serviceModules(test *testing.T) string {
	test.Helper()
	if os.Getenv("WASI_SYSROOT") == "" {
		test.Fatal("WASI_SYSROOT is required to build the service modules")
	}
	repository := serviceRepository(test)
	inputs := serviceCompiledInputs(test, repository, "wasi-service-modules", wasiCompiler())
	for _, mutant := range serviceMutants {
		inputs.Flags = append(inputs.Flags, "mutant "+
			mutant.name+
			" "+
			strconv.Quote(mutant.before)+
			" -> "+
			strconv.Quote(mutant.after)+
			" appending "+
			strconv.Quote(mutant.appended))
	}
	return buildcache.Product(test, inputs, func(directory string) error {
		compiler := buildcache.Adamic(test)
		source := filepath.Join(repository, serviceSource)
		contents, readError := os.ReadFile(source)
		if readError != nil {
			return readError
		}
		wasiCommand(
			test,
			repository,
			compiler,
			"build",
			"--target",
			"wasm32-wasi",
			source,
			"-o",
			filepath.Join(directory, "service.wasm"),
		)
		wasiCommand(
			test,
			repository,
			compiler,
			"build",
			"--target",
			"wasm32-wasi",
			source,
			"-o",
			filepath.Join(directory, "service-counted.wasm"),
			"--count",
		)
		// Each mutant's source is written outside the product, which holds modules alone.
		scratch, temporaryDirectoryError := os.MkdirTemp("", "adamic-service-")
		if temporaryDirectoryError != nil {
			return temporaryDirectoryError
		}
		defer os.RemoveAll(scratch)
		for _, mutant := range serviceMutants {
			if strings.Count(string(contents), mutant.before) != 1 {
				return fmt.Errorf("mutant %s must change exactly one source location", mutant.name)
			}
			mutantSource := filepath.Join(scratch, mutant.name+".a")
			if writeError := os.WriteFile(
				mutantSource,
				[]byte(strings.Replace(string(contents), mutant.before, mutant.after, 1)+mutant.appended),
				0o644,
			); writeError != nil {
				return writeError
			}
			wasiCommand(
				test,
				repository,
				compiler,
				"build",
				"--target",
				"wasm32-wasi",
				mutantSource,
				"-o",
				filepath.Join(directory, mutant.name+".wasm"),
				"--count",
			)
		}
		regions := "0"
		if strings.Contains(string(wasiCommand(test, repository, compiler, "c", source)), "adamic_region_end(") {
			regions = "1"
		}
		return os.WriteFile(filepath.Join(directory, "regions"), []byte(regions), 0o644)
	})
}

// serviceRegions is the modules product's answer to whether the generated C ends statement regions, which a counted
// run then requires to advance.
func serviceRegions(test *testing.T, modules string) string {
	test.Helper()
	regions, readError := os.ReadFile(filepath.Join(modules, "regions"))
	if readError != nil {
		test.Fatal(readError)
	}
	if string(regions) != "0" && string(regions) != "1" {
		test.Fatalf("the service modules answer regions %q, want 0 or 1", regions)
	}
	return string(regions)
}

// serviceNative is the sanitized native command every agreement part runs, a product built ahead by
// TestProduct_WASIServiceNative: command.a, which imports the very same handler, built once from the tree with
// --sanitize. Each quarter used to build the identical command from a copy in a temporary directory, which a build
// could name, so no two builds were sure to match.
func serviceNative(test *testing.T) string {
	test.Helper()
	repository := serviceRepository(test)
	inputs := serviceCompiledInputs(test, repository, "wasi-service-native", compilerName(Options{}))
	inputs.Files = append(inputs.Files, serviceCommand)
	inputs.Flags = append(inputs.Flags, "--sanitize", "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	directory := buildcache.Product(test, inputs, func(directory string) error {
		wasiCommand(
			test,
			repository,
			buildcache.Adamic(test),
			"build",
			filepath.Join(repository, serviceCommand),
			"-o",
			filepath.Join(directory, "service-native"),
			"--sanitize",
		)
		return nil
	})
	return filepath.Join(directory, "service-native")
}

// generatedLines is the lines of a file the host generated, which holds only one part's span, so a part never holds
// the whole request file, about 400 MB.
func generatedLines(path string) ([]string, error) {
	contents, readError := os.ReadFile(path)
	if readError != nil {
		return nil, readError
	}
	if len(contents) == 0 {
		return nil, nil
	}
	return strings.Split(strings.TrimSuffix(string(contents), "\n"), "\n"), nil
}

// serviceReportLine is the host's closing line, its report.
func serviceReportLine(output []byte) (map[string]any, error) {
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var report map[string]any
	if decodeError := json.Unmarshal([]byte(lines[len(lines)-1]), &report); decodeError != nil {
		return nil, fmt.Errorf("service report: %v\n%s", decodeError, output)
	}
	return report, nil
}

// serviceReport reads the host's closing line; a run that served fewer than all the requests is no evidence.
func serviceReport(output []byte, counted bool) error {
	report, serviceReportLineError := serviceReportLine(output)
	if serviceReportLineError != nil {
		return serviceReportLineError
	}
	if requests, _ := report["requests"].(float64); requests != serviceRequests {
		return fmt.Errorf("service served %v requests, want %d\n%s", report["requests"], serviceRequests, output)
	}
	if counted && report["live"] != float64(0) {
		return fmt.Errorf("service left %v values live, want 0\n%s", report["live"], output)
	}
	return nil
}

// wasmLeg has the host serve all 100,000 requests through one instance of the plain or the counted module.
func wasmLeg(test *testing.T, counted string) {
	test.Helper()
	repository := serviceOptedIn(test)
	modules := serviceModules(test)
	module := "service.wasm"
	if counted == "1" {
		module = "service-counted.wasm"
	}
	output, runServiceHostError := runServiceHost(
		repository,
		filepath.Join(repository, serviceHost),
		filepath.Join(modules, module),
		test.TempDir(),
		counted,
		serviceRegions(test, modules),
		[2]int{0, serviceRequests},
	)
	if runServiceHostError != nil {
		test.Fatalf("service counted=%s: %v\n%s", counted, runServiceHostError, output)
	}
	if serviceReportError := serviceReport(output, counted == "1"); serviceReportError != nil {
		test.Fatal(serviceReportError)
	}
	test.Logf("service counted=%s: %s", counted, output)
}

// Not parallel: the service oracle measures request rates, and a neighbor's load would skew them.
func TestWASIService(test *testing.T) {
	wasmLeg(test, "0")
}

// Not parallel: the service oracle measures request rates, and a neighbor's load would skew them.
func TestWASIServiceCounted(test *testing.T) {
	wasmLeg(test, "1")
}

// nativePart has the host step through all the requests and write one part's, and Node's responses to them, then has
// the product's sanitized native command answer that request file, each line held to Node's response line for it.
func nativePart(test *testing.T, part int) {
	test.Helper()
	if tileRequestsError := tileRequests(serviceParts, serviceRequests); tileRequestsError != nil {
		test.Fatal(tileRequestsError)
	}
	repository := serviceOptedIn(test)
	directory := test.TempDir()
	span := serviceParts[part]
	output, runServiceHostError := runServiceHost(
		repository,
		filepath.Join(repository, serviceHost),
		"-",
		directory,
		"0",
		"0",
		span,
	)
	if runServiceHostError != nil {
		test.Fatalf("generate: %v\n%s", runServiceHostError, output)
	}
	if serviceReportError := serviceReport(output, false); serviceReportError != nil {
		test.Fatal(serviceReportError)
	}
	report, serviceReportLineError := serviceReportLine(output)
	if serviceReportLineError != nil {
		test.Fatal(serviceReportLineError)
	}
	generated, _ := report["range"].([]any)
	if len(generated) != 2 ||
		generated[0] != float64(span[0]) ||
		generated[1] != float64(span[1]) ||
		report["lines"] != float64(span[1]-span[0]) {
		test.Fatalf(
			"host generated range %v in %v lines, want [%d %d] in %d\n%s",
			report["range"],
			report["lines"],
			span[0],
			span[1],
			span[1]-span[0],
			output,
		)
	}
	partRequests := filepath.Join(directory, "requests.jsonl")
	requestLines, generatedLinesError := generatedLines(partRequests)
	if generatedLinesError != nil {
		test.Fatal(generatedLinesError)
	}
	responseLines, generatedLinesError := generatedLines(filepath.Join(directory, "responses.jsonl"))
	if generatedLinesError != nil {
		test.Fatal(generatedLinesError)
	}
	if len(requestLines) != span[1]-span[0] || len(responseLines) != span[1]-span[0] {
		test.Fatalf(
			"generated %d requests and %d responses for [%d, %d), want %d of each",
			len(requestLines),
			len(responseLines),
			span[0],
			span[1],
			span[1]-span[0],
		)
	}
	binary := serviceNative(test)
	timeoutContext, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	native, outputError := exec.CommandContext(timeoutContext, binary, partRequests).Output()
	if outputError != nil {
		if failed, ok := outputError.(*exec.ExitError); ok {
			test.Fatalf("native service: %v\n%s", outputError, failed.Stderr)
		}
		test.Fatalf("native service: %v", outputError)
	}
	actualLines := strings.Split(strings.TrimSuffix(string(native), "\n"), "\n")
	for index := span[0]; index < span[1]; index++ {
		if index-span[0] >= len(actualLines) {
			test.Fatalf("native command ended before response %d", index)
		}
		if actualLines[index-span[0]] != responseLines[index-span[0]] {
			test.Fatalf(
				"native response %d: got %q, want %q",
				index,
				actualLines[index-span[0]],
				responseLines[index-span[0]],
			)
		}
	}
	if len(actualLines) != span[1]-span[0] {
		test.Fatalf("native command emitted %d lines, want %d", len(actualLines), span[1]-span[0])
	}
	test.Logf("native sanitized command: responses %d to %d agree with Node", span[0], span[1]-1)
}

func TestWASIServiceNativeAgrees_0(test *testing.T) {
	test.Parallel()
	nativePart(test, 0)
}

func TestWASIServiceNativeAgrees_1(test *testing.T) {
	test.Parallel()
	nativePart(test, 1)
}

func TestWASIServiceNativeAgrees_2(test *testing.T) {
	test.Parallel()
	nativePart(test, 2)
}

func TestWASIServiceNativeAgrees_3(test *testing.T) {
	test.Parallel()
	nativePart(test, 3)
}

func TestWASIServiceNativeAgrees_4(test *testing.T) {
	test.Parallel()
	nativePart(test, 4)
}

func TestWASIServiceNativeAgrees_5(test *testing.T) {
	test.Parallel()
	nativePart(test, 5)
}

func TestWASIServiceNativeAgrees_6(test *testing.T) {
	test.Parallel()
	nativePart(test, 6)
}

func TestWASIServiceNativeAgrees_7(test *testing.T) {
	test.Parallel()
	nativePart(test, 7)
}

// serviceAgreementParts are the parts this file's TestWASIServiceNativeAgrees_<n> tests run, each test's n, which must
// be the part its nativePart call names.
func serviceAgreementParts(test *testing.T) []int {
	test.Helper()
	files := token.NewFileSet()
	parsed, parseFileError := parser.ParseFile(files, "wasm_service_test.go", nil, parser.SkipObjectResolution)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	var parts []int
	for _, declaration := range parsed.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		suffix, found := strings.CutPrefix(function.Name.Name, "TestWASIServiceNativeAgrees_")
		if !found {
			continue
		}
		part, parseError := strconv.Atoi(suffix)
		if parseError != nil {
			test.Fatalf("%s names no part", function.Name.Name)
		}
		runs := false
		ast.Inspect(function.Body, func(node ast.Node) bool {
			if call, ok := node.(*ast.CallExpr); ok && len(call.Args) == 2 {
				name, named := call.Fun.(*ast.Ident)
				literal, literally := call.Args[1].(*ast.BasicLit)
				runs = runs || named && name.Name == "nativePart" && literally && literal.Value == suffix
			}
			return true
		})
		if !runs {
			test.Fatalf("%s doesn't run nativePart(t, %s)", function.Name.Name, suffix)
		}
		parts = append(parts, part)
	}
	slices.Sort(parts)
	return parts
}

// The tests above are one per part, so a part without its test would go unrun, and a test without its part would fail
// at its index; and parts with a gap or an overlap must be refused, not passed.
func TestWASIServicePartsTile(test *testing.T) {
	test.Parallel()
	parts := serviceAgreementParts(test)
	if len(parts) != len(serviceParts) {
		test.Fatalf("%d parts, and %d TestWASIServiceNativeAgrees_ tests run them", len(serviceParts), len(parts))
	}
	for index, part := range parts {
		if part != index {
			test.Fatalf(
				"TestWASIServiceNativeAgrees_ tests run parts %v, want 0 to %d once each",
				parts,
				len(serviceParts)-1,
			)
		}
	}
	if tileRequestsError := tileRequests(serviceParts, serviceRequests); tileRequestsError != nil {
		test.Fatal(tileRequestsError)
	}
	changed := func(change func(ranges [][2]int)) [][2]int {
		ranges := slices.Clone(serviceParts)
		change(ranges)
		return ranges
	}
	for _, mutant := range []struct {
		name   string
		ranges [][2]int
	}{
		{"gap", changed(func(ranges [][2]int) { ranges[1][0]++ })},
		{"overlap", changed(func(ranges [][2]int) { ranges[1][0]-- })},
		{"short", changed(func(ranges [][2]int) { ranges[len(ranges)-1][1]-- })},
		{"empty", changed(func(ranges [][2]int) { ranges[1][1], ranges[2][0] = ranges[1][0], ranges[1][0] })},
	} {
		if tileRequests(mutant.ranges, serviceRequests) == nil {
			test.Errorf("%s parts passed the tiling check: %v", mutant.name, mutant.ranges)
		}
	}
}

// serviceMutant has the host run the named mutant's counted module, holding the original source as its oracle, and
// requires the mutant caught by its named assertion.
func serviceMutant(test *testing.T, name string) {
	test.Helper()
	repository := serviceOptedIn(test)
	index := slices.IndexFunc(serviceMutants, func(mutant serviceMutation) bool { return mutant.name == name })
	if index < 0 {
		test.Fatalf("no service mutant %s", name)
	}
	modules := serviceModules(test)
	output, runServiceHostError := runServiceHost(
		repository,
		filepath.Join(repository, serviceHost),
		filepath.Join(modules, name+".wasm"),
		test.TempDir(),
		"1",
		serviceRegions(test, modules),
		[2]int{0, serviceRequests},
	)
	if runServiceHostError == nil || !strings.Contains(string(output), serviceMutants[index].caught) {
		test.Fatalf("mutant not caught by %q: %v\n%s", serviceMutants[index].caught, runServiceHostError, output)
	}
	test.Logf("mutant compiled and was caught by %q\n%s", serviceMutants[index].caught, output)
}

func TestWASIServiceMutantResponseByte(test *testing.T) {
	test.Parallel()
	serviceMutant(test, "response-byte")
}

func TestWASIServiceMutantRetainedObject(test *testing.T) {
	test.Parallel()
	serviceMutant(test, "retained-object")
}

func TestWASIServiceMutantUnicodeEscape(test *testing.T) {
	test.Parallel()
	serviceMutant(test, "unicode-escape")
}

// A leg that served fewer requests than the claim must be refused, not passed: the host told to generate a thousand
// stands in for any split that quietly shrinks the count.
func TestWASIServiceMutantShrunkCount(test *testing.T) {
	test.Parallel()
	repository := serviceOptedIn(test)
	original, readError := os.ReadFile(filepath.Join(repository, serviceHost))
	if readError != nil {
		test.Fatal(readError)
	}
	if strings.Count(string(original), "const requestCount = 100000;") != 1 {
		test.Fatal("shrunk-count mutant must change the host's one request count")
	}
	shrunk := strings.Replace(string(original), "const requestCount = 100000;", "const requestCount = 1000;", 1)
	directory := test.TempDir()
	host := filepath.Join(directory, "host-shrunk.mjs")
	writeWASIFile(test, host, shrunk)
	output, runServiceHostError := runServiceHost(repository, host, "-", directory, "0", "0", [2]int{0, 1000})
	if runServiceHostError != nil {
		test.Fatalf("shrunk host: %v\n%s", runServiceHostError, output)
	}
	if serviceReport(output, false) == nil {
		test.Fatalf("a run of fewer than %d requests passed the report check\n%s", serviceRequests, output)
	}
	test.Log("shrunk-count mutant refused by the report check")
}

// TestProduct_WASIServiceModules builds the service's modules, the mutants' and the regions answer the wasm legs and
// the mutants read, where they run, so the tree build builds them ahead.
func TestProduct_WASIServiceModules(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_TEST_WASI") != "1" {
		test.Skip("set ADAMIC_TEST_WASI=1")
	}
	serviceModules(test)
}

// TestProduct_WASIServiceNative builds the sanitized native command the agreement parts run, where they run, so the
// tree build builds it ahead.
func TestProduct_WASIServiceNative(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_TEST_WASI") != "1" {
		test.Skip("set ADAMIC_TEST_WASI=1")
	}
	serviceNative(test)
}

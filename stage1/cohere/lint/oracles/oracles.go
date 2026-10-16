// Package oracles owns lint reference products without linking the compiler.
package oracles

import (
	bytes "bytes"
	context "context"
	sha256 "crypto/sha256"
	json "encoding/json"
	fmt "fmt"
	buildcache "github.com/system-inc/adamic/internal/buildcache"
	childguard "github.com/system-inc/adamic/internal/childguard"
	testguard "github.com/system-inc/adamic/internal/testguard"
	toolchain "github.com/system-inc/adamic/internal/toolchain"
	registry "github.com/system-inc/adamic/stage1/cohere/lint/registry"
	fs "io/fs"
	os "os"
	filepath "path/filepath"
	regexp "regexp"
	runtime "runtime"
	sort "sort"
	strings "strings"
	sync "sync"
	testing "testing"
	time "time"
)

var packageDirectory string

var sharedDirectory = os.TempDir()

var environmentOnce sync.Once

var oracleProducer bool

func ensureEnvironment() {
	environmentOnce.Do(func() {
		var err error
		packageDirectory, err = filepath.Abs(".")
		if err != nil {
			panic(err)
		}
	})
}

func oracleProduct(t *testing.T, inputs buildcache.Inputs, build func(string) error) string {
	return oracleProductDirectory(t, inputs, build)
}

// A nil build resolves a product the caller already prepared, without building it again.
// Both producers and artifact rows use this one owner identity.
func oracleProductDirectory(t *testing.T, inputs buildcache.Inputs, build func(string) error) string {
	t.Helper()
	inputs.Name += " oracle-owner-v1"
	if build == nil {
		// Get deliberately rebuilds every request in uncached mode. Absolute also resolves
		// directories already held by this process, including uncached products.
		if os.Getenv("ADAMIC_BUILD_CACHE") == "off" {
			root, err := filepath.Abs(repository)
			if err != nil {
				t.Fatal(err)
			}
			return buildcache.Absolute("<build cache>/" + buildcache.NameKey(root, inputs))
		}
		directory, err := buildcache.Get(inputs, func(string) error {
			return fmt.Errorf("prepare oracle product before checking its artifacts: %s", inputs.Name)
		})
		if err != nil {
			t.Fatal(err)
		}
		return directory
	}
	if oracleProducer {
		return buildcache.Product(t, inputs, build)
	}
	return buildcache.BuiltElsewhere(t, inputs.Name, inputs, "stage1/cohere/lint/oracles owners", func(string) error {
		return fmt.Errorf("settle stage1/cohere/lint/oracles owners first: %s", inputs.Name)
	})
}

func oracleFiles(t *testing.T) []string {
	files := []string{"cohere", "go.mod", "go.work", "stage1/cohere/lint/registry", "stage1/cohere/lint/testdata", "stage1/cohere/lint/oracles"}
	for _, f := range portFiles(t) {
		files = append(files, "stage1/cohere/lint/"+filepath.ToSlash(f))
	}
	return files
}

func guardedRun(commandName string, directory string, environment []string, args ...string) ([]byte, error) {
	command := toolchain.CommandContextIn(context.Background(), directory, commandName, args...)
	if strings.ContainsRune(commandName, filepath.Separator) {
		command.Env = append(os.Environ(), environment...)
	} else {
		command.Env = append(command.Env, environment...)
	}
	if directory != "" {
		command.Env = append(command.Env, "PWD="+directory)
	}
	output, err := os.CreateTemp(sharedDirectory, "oracle-stdout-")
	if err != nil {
		return nil, err
	}
	defer os.Remove(output.Name())
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err = testguard.PrepareCPU(command, testguard.Budget); err != nil {
		return nil, err
	}
	err = childguard.Run(command, childguard.Options{FirstOutput: testguard.Ceiling, Stall: testguard.Ceiling, Ceiling: testguard.Ceiling})
	if err != nil || len(commandDiagnostics(commandName, stderr.Bytes())) != 0 {
		return nil, fmt.Errorf("%s %v: %v\n%s", commandName, args, err, &stderr)
	}
	return os.ReadFile(output.Name())
}

func rulesAgreeCaptureRun(directory string, environment []string, args ...string) ([]byte, error) {
	return guardedRun("go", directory, environment, args...)
}

func testShardsAgreeRun(directory string, environment []string, name string, args ...string) ([]byte, error) {
	command := toolchain.CommandContextIn(context.Background(), directory, name, args...)
	if strings.ContainsRune(name, filepath.Separator) {
		command.Env = append(os.Environ(), environment...)
		if directory != "" {
			command.Env = append(command.Env, "PWD="+directory)
		}
	} else {
		command.Env = append(command.Env, environment...)
	}
	output, err := os.CreateTemp(sharedDirectory, "oracle-stdout-")
	if err != nil {
		return nil, err
	}
	defer os.Remove(output.Name())
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := childguard.Run(command, childguard.Options{Unbounded: true}); err != nil || len(commandDiagnostics(name, stderr.Bytes())) != 0 {
		return nil, fmt.Errorf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	return os.ReadFile(output.Name())
}

func execute(t *testing.T, directory, name string, args ...string) execution {
	t.Helper()
	start := time.Now()
	output, err := guardedRun(name, directory, nil, args...)
	if err != nil {
		t.Fatal(err)
	}
	return execution{output, time.Since(start)}
}

func lintGoOracleProduct(t *testing.T) string {
	t.Helper()
	lintOracleOnce.Do(func() {
		prepareRegistry(t, packageDirectory)
		path, err := goOracleOverlay(packageDirectory, t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		const output = "oracle-owner-v1"
		arguments := []string{"-tags=lintoracle", "-overlay=" + path}
		if oracleProducer {
			lintOraclePath = buildcache.GoBuildIn(t, "cohere", output, "./"+goOraclePackage, arguments)
			return
		}
		inputs, err := buildcache.GoBuildInInputs("cohere", output, "./"+goOraclePackage, arguments, nil)
		if err != nil {
			t.Fatal(err)
		}
		directory := buildcache.BuiltElsewhere(t, "lint Go oracle", inputs, "stage1/cohere/lint/oracles TestProduct_LintGoOracle", func(string) error { return fmt.Errorf("settle the lint Go oracle owner first") })
		lintOraclePath = filepath.Join(directory, output)
	})
	if lintOraclePath == "" {
		t.Fatal("lint Go oracle preparation failed")
	}
	return lintOraclePath
}

func completeSuggestionOracleProduct(ctx context.Context, t *testing.T, directory string) string {
	t.Helper()
	return oracleProduct(t, buildcache.Inputs{
		Name: "complete-suggestion-go-oracle", Files: oracleFiles(t),
		Flags:     []string{packageDirectory, "serialization-only-no-debugger-v1", "go build " + strings.Join(buildcache.ReproducibleGoFlags(), " ")},
		Toolchain: []string{runtime.Version(), buildcache.Tool("go", "version"), lintGoSettings()},
	}, func(out string) error { return completeSuggestionBuildOracle(ctx, directory, out) })
}

func rulesAgreeCapture(t *testing.T) string {
	t.Helper()
	rulesAgreeCaptureOnce.Do(func() {
		directory := packageDirectory
		captureInputs := buildcache.Inputs{
			Name:      "lint-rules-agree-capture",
			Files:     oracleFiles(t),
			Flags:     []string{"asserted-cases-overlay", "-count=1", "-timeout=0"},
			Toolchain: []string{runtime.Version(), lintGoSettings()},
		}
		rulesAgreeCaptureDirectory = oracleProduct(t, captureInputs, func(output string) error {
			captured, err := rulesAgreeCaptureUpstream(t, directory, output)
			if err != nil {
				return err
			}
			for index, row := range captured {
				fields := strings.SplitN(row, "\t", 2)
				relative, err := filepath.Rel(output, fields[0])
				if err != nil {
					return err
				}
				captured[index] = relative + "\t" + fields[1]
			}
			data, err := json.Marshal(captured)
			if err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(output, "rows.json"), data, 0644)
		})

	})
	if rulesAgreeCaptureDirectory == "" {
		t.Fatal("capture preparation failed")
	}
	return rulesAgreeCaptureDirectory
}

func testShardsAgreeUpstream(t *testing.T) []string {
	d := oracleProduct(t, testShardsAgreeInputs(t, "capture"), func(d string) error {
		rows, err := testShardsAgreeCaptureUpstream(packageDirectory, d)
		if err != nil {
			return err
		}
		// Each row names a case in this product, by its key, wherever the reader's cache holds it (#tqrqx60).
		return os.WriteFile(filepath.Join(d, "rows"), []byte(buildcache.Relative(strings.Join(rows, "\n"))), 0644)
	})
	data, err := os.ReadFile(filepath.Join(d, "rows"))
	if err != nil {
		t.Fatal(err)
	}
	rows := strings.Split(buildcache.Absolute(string(data)), "\n")
	for _, row := range rows {
		if !strings.HasPrefix(row, d+"/case-") {
			t.Fatalf("capture path: %q", row)
		}
	}
	return rows
}

func compilationOracleInputs(t *testing.T, binary, source []byte) buildcache.Inputs {
	t.Helper()
	witness := strings.SplitN(compilationCase(t), "\t", 2)[0]
	return buildcache.Inputs{Name: "lint-profile-oracle-outputs", Files: []string{witness, "stage1/cohere/lint/oracles"}, Flags: []string{"mode=no-var", fmt.Sprintf("oracle-sha256=%x", sha256.Sum256(binary)), fmt.Sprintf("case-sha256=%x", sha256.Sum256(source))}, Toolchain: []string{runtime.GOOS, runtime.GOARCH}}
}

func compilationOracleOutputs(t *testing.T, oracle, manifestPath string) ([]byte, []byte) {
	t.Helper()
	binary, err := os.ReadFile(oracle)
	if err != nil {
		t.Fatal(err)
	}
	text, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	sourcePath := strings.SplitN(string(text), "\t", 2)[0]
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	inputs := compilationOracleInputs(t, binary, source)
	product := oracleProduct(t, inputs, func(dir string) error {
		for _, mode := range []string{"wire", "count"} {
			args := []string{"--manifest", manifestPath}
			if mode == "count" {
				args = append(args, "--count")
			}
			data := execute(t, "", oracle, args...).output
			data = bytes.ReplaceAll(data, []byte(sourcePath), []byte("@profile-case@"))
			if err := os.WriteFile(filepath.Join(dir, mode), data, 0644); err != nil {
				return err
			}
		}
		return nil
	})
	read := func(name string) []byte {
		data, err := os.ReadFile(filepath.Join(product, name))
		if err != nil {
			t.Fatal(err)
		}
		return bytes.ReplaceAll(data, []byte("@profile-case@"), []byte(sourcePath))
	}
	return read("wire"), read("count")
}

func yepestaPrepare(t *testing.T) string {
	t.Helper()
	directory := oracleProduct(t, yepestaInputs(), func(out string) error {
		oracle := lintGoOracleProduct(t)
		rows := jsxTextnodesRows(t, oracle)
		witnesses := recoveryRows(t, oracle, ownedWitnessRows(t, ".", jsxTextnodesDescriptor(t)))
		bundle := yepestaBundle{Oracle: filepath.Join(out, "oracle")}
		data, err := os.ReadFile(oracle)
		if err != nil {
			return err
		}
		if err := os.WriteFile(bundle.Oracle, data, 0755); err != nil {
			return err
		}
		persist := func(rows []string, prefix string) ([]string, error) {
			var result []string
			for i, row := range rows {
				fields := strings.Split(row, "\t")
				data, err := os.ReadFile(fields[0])
				if err != nil {
					return nil, err
				}
				name := filepath.Join(prefix, fmt.Sprintf("%03d", i), filepath.Base(fields[0]))
				if err := os.MkdirAll(filepath.Dir(filepath.Join(out, name)), 0755); err != nil {
					return nil, err
				}
				if err := os.WriteFile(filepath.Join(out, name), data, 0644); err != nil {
					return nil, err
				}
				fields[0] = name
				result = append(result, strings.Join(fields, "\t"))
			}
			return result, nil
		}
		bundle.Rows, err = persist(rows, "rows")
		if err != nil {
			return err
		}
		bundle.Witnesses, err = persist(witnesses, "witnesses")
		if err != nil {
			return err
		}
		bundle.Oracle = "oracle"
		data, err = json.Marshal(bundle)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(out, "bundle.json"), data, 0644)
	})
	return directory
}

func testShardsAgreeInputs(t *testing.T, name string) buildcache.Inputs {
	if name != "capture" {
		t.Fatal("oracle owner only captures upstream")
	}
	return buildcache.Inputs{Name: "shards-agree-capture", Files: oracleFiles(t), Flags: []string{"asserted-cases-overlay"}, Toolchain: []string{runtime.Version(), buildcache.Tool("go", "version"), lintGoSettings()}}
}

func yepestaInputs() buildcache.Inputs {
	return buildcache.Inputs{Name: "yepesta-setup-v2", Files: []string{"go.mod", "go.work", "stage1/cohere/lint/rules", "stage1/cohere/lint/oracles"}, Toolchain: []string{buildcache.Tool("go", "version"), lintGoSettings()}}
}

func jsxGoOracle(t *testing.T, name, root, side, virtualName string) string {
	t.Helper()
	inputs := jsxProductInputs{Name: name, Files: jsxInputFiles(t, side, filepath.Join(repository, "cohere")), Flags: []string{"go build " + strings.Join(buildcache.ReproducibleGoFlags(), " "), "-overlay", virtualName}, Toolchain: []string{runtime.Version(), lintGoSettings()}}
	directory := oracleProduct(t, buildcache.Inputs{Name: inputs.Name, Files: inputs.Files, Flags: inputs.Flags, Toolchain: inputs.Toolchain}, func(dir string) error {
		virtual := filepath.Join(root, virtualName)
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side}})
		if err != nil {
			return err
		}
		overlayPath := filepath.Join(dir, "overlay.json")
		if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
			return err
		}
		// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
		defer os.Remove(overlayPath)
		// toolchain's go, under its environment, in root with PWD set to it: the go and the settings the key reads.
		command := toolchain.CommandContextIn(t.Context(), root, "go", append(append([]string{"build"}, buildcache.ReproducibleGoFlags()...), "-overlay="+overlayPath, "-o", filepath.Join(dir, "oracle"), virtual)...)
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		if err := childguard.Run(command, childguard.Options{Unbounded: true}); err != nil {
			return fmt.Errorf("build %s: %w\n%s\n%s", name, err, &stdout, &stderr)
		}
		if diagnostics := commandDiagnostics("go", stderr.Bytes()); len(diagnostics) != 0 {
			return fmt.Errorf("build %s: %s", name, diagnostics)
		}
		return nil
	})
	return filepath.Join(directory, "oracle")
}

func completeSuggestionBuildOracle(ctx context.Context, sourceRoot, directory string) error {
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		return err
	}
	side, err := filepath.Abs(filepath.Join(packageDirectory, "testdata/oracle.go"))
	if err != nil {
		return err
	}
	descriptors, err := registry.Generate(sourceRoot)
	if err != nil {
		return err
	}
	replacements := map[string]string{}
	var virtualFiles []string
	var failure error
	add := func(name, source string) {
		virtual := filepath.Join(root, "adamic_lint_"+name+".go")
		absolute, err := filepath.Abs(source)
		if err != nil {
			failure = err
			return
		}
		replacements[virtual] = absolute
		virtualFiles = append(virtualFiles, virtual)
	}
	add("oracle", side)
	add("registry", filepath.Join(sourceRoot, ".generated/registry.go"))
	for _, d := range descriptors {
		add(strings.ReplaceAll(d.Slug, "-", "_"), filepath.Join(sourceRoot, "rules", d.Slug, "oracle.go"))
	}
	if failure != nil {
		return failure
	}
	overlay, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		return err
	}
	path := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		return err
	}
	// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
	defer os.Remove(path)
	binary := filepath.Join(directory, "oracle")
	args := append(append(append([]string{"build"}, buildcache.ReproducibleGoFlags()...), "-overlay="+path, "-o", binary), virtualFiles...)
	command := toolchain.CommandContextIn(ctx, root, "go", args...)
	var output, stderr bytes.Buffer
	command.Stdout, command.Stderr = &output, &stderr
	err = childguard.Run(command, childguard.Options{Unbounded: true})
	if err != nil || len(commandDiagnostics("go", stderr.Bytes())) != 0 {
		return fmt.Errorf("Go oracle build: %v\n%s\n%s", err, &output, &stderr)
	}
	return nil
}

func suggestionAlongsideGoOracle(t *testing.T, sourceRoot, directory string) (string, error) {
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		return "", err
	}
	side, err := filepath.Abs(filepath.Join(packageDirectory, "testdata/oracle.go"))
	if err != nil {
		return "", err
	}
	descriptors, err := registry.Generate(sourceRoot)
	if err != nil {
		return "", err
	}
	replacements := map[string]string{}
	var virtualFiles []string
	var failure error
	add := func(name, source string) {
		virtual := filepath.Join(root, "adamic_lint_"+name+".go")
		absolute, err := filepath.Abs(source)
		if err != nil {
			failure = err
			return
		}
		replacements[virtual] = absolute
		virtualFiles = append(virtualFiles, virtual)
	}
	add("oracle", side)
	add("registry", filepath.Join(sourceRoot, ".generated/registry.go"))
	for _, d := range descriptors {
		add(strings.ReplaceAll(d.Slug, "-", "_"), filepath.Join(sourceRoot, "rules", d.Slug, "oracle.go"))
	}
	if failure != nil {
		return "", failure
	}
	overlay, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		return "", err
	}
	path := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		return "", err
	}
	// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
	defer os.Remove(path)
	binary := filepath.Join(directory, "oracle")
	args := append(append(append([]string{"build"}, buildcache.ReproducibleGoFlags()...), "-overlay="+path, "-o", binary), virtualFiles...)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	command := toolchain.CommandContextIn(ctx, root, "go", args...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := childguard.Run(command, childguard.Options{Unbounded: true}); err != nil || len(commandDiagnostics("go", stderr.Bytes())) != 0 {
		return "", fmt.Errorf("go build: %v\n%s", err, &stderr)
	}
	return binary, nil
}

func suggestionAlongsideFixture(t *testing.T) (string, string) {
	directory, err := os.MkdirTemp(sharedDirectory, "suggestion-alongside-")
	if err != nil {
		t.Fatal(err)
	}
	copyPort(t, directory, "", "")
	// This pinned case selects only no-debugger. Exclude unrelated rules from
	// the generated dispatch so setup does not lower the entire live corpus.
	rules, err := os.ReadDir(filepath.Join(directory, "rules"))
	if err != nil {
		t.Fatal(err)
	}
	for _, rule := range rules {
		if rule.IsDir() && rule.Name() != "no-debugger" {
			if err := os.RemoveAll(filepath.Join(directory, "rules", rule.Name())); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, name := range []string{"rule.a", "oracle.go"} {
		data, err := os.ReadFile(filepath.Join("testdata/serialization", name))
		if err != nil {
			t.Fatal(err)
		}
		from, to := "debuggerMessage, '', '', ''", "debuggerMessage, 'fix', ';', ''"
		if name == "oracle.go" {
			from, to = "d.Fixes = nil", "d.Fixes[0].Text = \";\""
		}
		if bytes.Count(data, []byte(from)) != 1 {
			t.Fatal("automatic fix anchor changed")
		}
		if err := os.WriteFile(filepath.Join(directory, "rules/no-debugger", name), bytes.Replace(data, []byte(from), []byte(to), 1), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Remove(filepath.Join(directory, "rules/no-debugger/rule.ts")); err != nil {
		t.Fatal(err)
	}
	prepareRegistry(t, directory)
	source := filepath.Join(directory, "mixed.ts")
	if err := os.WriteFile(source, []byte("/*😀*/debugger;\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Keep the manifest alive after the setup test returns.
	path := filepath.Join(directory, "manifest.txt")
	if err := os.WriteFile(path, []byte(source+"\tno-debugger\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return directory, path
}

func suggestionAlongsideOracleProduct(t *testing.T, directory string) string {
	inputs := buildcache.Inputs{Name: "suggestion-alongside-go-oracle-v1", Files: oracleFiles(t), Flags: []string{"go build " + strings.Join(buildcache.ReproducibleGoFlags(), " "), "overlay=serialization-semicolon", "GOFLAGS=" + os.Getenv("GOFLAGS"), "CGO_ENABLED=" + os.Getenv("CGO_ENABLED")}, Toolchain: []string{buildcache.Tool("go", "version"), lintGoSettings(), runtime.GOOS, runtime.GOARCH}}
	out := oracleProduct(t, inputs, func(out string) error { _, err := suggestionAlongsideGoOracle(t, directory, out); return err })
	return filepath.Join(out, "oracle")
}

func compilerAgreementCheckerAllAnswer(t *testing.T, owned []compilerAgreementCase) []byte {
	t.Helper()
	oracle := lintGoOracleProduct(t)
	binary, err := os.ReadFile(oracle)
	if err != nil {
		t.Fatal(err)
	}
	files := make([][]byte, len(owned))
	var rows []string
	for i, c := range owned {
		if c.key != compilerAgreementCheckerKey || c.rule != "all" {
			t.Fatalf("the checker.ts combined-rule owner holds %s/%s", c.key, c.rule)
		}
		// The source bytes are key material (their SHA is in Inputs.Flags).
		// The child reads only the private snapshot of those exact bytes below.
		buildcache.Keying(func() { files[i], err = os.ReadFile(c.path) })
		if err != nil {
			t.Fatal(err)
		}
		rows = append(rows, c.path+"\t"+c.rule)
	}
	if len(rows) != 1 {
		t.Fatalf("the checker.ts combined-rule owner holds %d rows, want one", len(rows))
	}
	directory := oracleProduct(t, compilerAgreementCheckerAllInputs(binary, owned, files), func(directory string) error {
		scratch, err := os.MkdirTemp("", "compiler-agreement-checker-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(scratch)
		path := filepath.Join(scratch, "manifest.txt")
		source := filepath.Join(scratch, "checker.ts")
		if err := os.WriteFile(source, files[0], 0644); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(source+"\tall\n"), 0644); err != nil {
			return err
		}
		output, err := testShardsAgreeRun("", nil, oracle, "--manifest", path)
		if err != nil {
			return err
		}
		if bytes.Contains(output, []byte(compilerAgreementCheckerPath)) {
			return fmt.Errorf("the Go answer for checker.ts already holds %q, so its path can't stand in for it", compilerAgreementCheckerPath)
		}
		return os.WriteFile(filepath.Join(directory, "answer"), bytes.ReplaceAll(output, []byte(source), []byte(compilerAgreementCheckerPath)), 0644)
	})
	answer, err := os.ReadFile(filepath.Join(directory, "answer"))
	if err != nil {
		t.Fatal(err)
	}
	return bytes.ReplaceAll(answer, []byte(compilerAgreementCheckerPath), []byte(owned[0].path))
}

func compilerAgreementCheckerAllInputs(oracle []byte, owned []compilerAgreementCase, files [][]byte) buildcache.Inputs {
	flags := []string{fmt.Sprintf("Go oracle binary sha256 %x", sha256.Sum256(oracle)), "--manifest"}
	for i, c := range owned {
		flags = append(flags, fmt.Sprintf("row %s sha256 %x rule %s", c.key, sha256.Sum256(files[i]), c.rule))
	}
	return buildcache.Inputs{Name: "compiler agreement checker.ts combined-rule Go answer", Files: []string{"stage1/cohere/lint/oracles"}, Flags: flags}
}

// The pinned TypeScript-go fixture is byte-for-byte checker.ts from the
// agreement corpus. Its bytes key the same answer; the live-oracle test checks
// both that equality and the normalized answer against the supplied checkout.
func prepareCheckerAnswer(t *testing.T) {
	path := filepath.Join(repository, "cohere/TypeScript/tsc/testdata/fixtures/compiler/checker.ts")
	compilerAgreementCheckerAllAnswer(t, []compilerAgreementCase{{path: path, key: compilerAgreementCheckerKey, rule: "all"}})
}

func prepareProfileOracleOutputs(t *testing.T) {
	compilationCase(t)
	directory := t.TempDir()
	copyPort(t, directory, "", "")
	prepareRegistry(t, directory)
	path := manifest(t, []string{ownedWitnesses(t, directory, "no-var")[0] + "\tno-var"})
	compilationOracleOutputs(t, lintGoOracleProduct(t), path)
}

const compilerAgreementCheckerKey = "compiler/src/compiler/checker.ts"

type compilerAgreementCase struct {
	path, key, rule string
	// group is a one-based checker side-4 part; zero is an ordinary corpus row.
	group int
}

const compilerAgreementCheckerPath = "<compiler agreement checker.ts row>"

func completeSuggestionFixture(t *testing.T, mutant bool) string {
	t.Helper()
	directory, err := os.MkdirTemp(sharedDirectory, "complete-suggestion-")
	if err != nil {
		t.Fatal(err)
	}
	copyPort(t, directory, "", "")
	for _, name := range []string{"rule.a", "oracle.go"} {
		data, err := os.ReadFile(filepath.Join("testdata/serialization", name))
		if err != nil {
			t.Fatal(err)
		}
		if mutant && name == "rule.a" {
			from := []byte("start + 1, start + 2, ''")
			if bytes.Count(data, from) != 1 {
				t.Fatal("suggestion mutant anchor changed")
			}
			data = bytes.Replace(data, from, []byte("start + 1, start + 3, ''"), 1)
		}
		if err = os.WriteFile(filepath.Join(directory, "rules/no-debugger", name), data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err = os.Remove(filepath.Join(directory, "rules/no-debugger/rule.ts")); err != nil {
		t.Fatal(err)
	}
	completeSuggestionOnlyRule(t, directory)
	return directory
}

func completeSuggestionOnlyRule(t *testing.T, directory string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(directory, "rules"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "no-debugger" {
			if err := os.RemoveAll(filepath.Join(directory, "rules", entry.Name())); err != nil {
				t.Fatal(err)
			}
		}
	}
	prepareRegistry(t, directory)
}

var rulesAgreeCaptureOnce sync.Once

var rulesAgreeCaptureDirectory string

func rulesAgreeCaptureUpstream(t *testing.T, sourceRoot, directory string) ([]string, error) {
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		return nil, err
	}
	harness := filepath.Join(root, "internal/lint/testing/rule_testing.go")
	data, err := os.ReadFile(harness)
	if err != nil {
		return nil, err
	}
	original := "return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}"
	replacement := "result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result"
	if strings.Count(string(data), original) != 1 {
		return nil, fmt.Errorf("capture overlay anchor changed")
	}
	side := filepath.Join(directory, "rule_testing.go")
	if err := os.WriteFile(side, []byte(strings.Replace(string(data), original, replacement, 1)), 0644); err != nil {
		return nil, err
	}
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{harness: side}})
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		return nil, err
	}
	// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
	defer os.Remove(overlayPath)
	capture := filepath.Join(directory, "capture")
	descriptors, err := registry.Generate(sourceRoot)
	if err != nil {
		return nil, err
	}
	discovered := map[string]bool{}
	packages := map[string][]string{}
	for _, d := range descriptors {
		discovered[d.Name] = true
		packages[d.UpstreamPackage] = append(packages[d.UpstreamPackage], d.UpstreamTest)
	}
	var names []string
	for name := range packages {
		names = append(names, name)
	}
	sort.Strings(names)
	// Independent upstream packages capture into separate destinations. Bound
	// concurrent packages to the instance's four CPUs; capture is shared setup.
	var workers sync.WaitGroup
	slots := make(chan struct{}, 4)
	failures := make([]error, len(names))
	for index, name := range names {
		workers.Add(1)
		go func() {
			defer workers.Done()
			slots <- struct{}{}
			defer func() { <-slots }()
			destination := filepath.Join(capture, name)
			environment := []string{"COHERE_DOCS_CAPTURE=" + destination}
			started := time.Now()
			_, failures[index] = rulesAgreeCaptureRun(root, environment, "test", "-overlay="+overlayPath, "./internal/lint/rules/"+name, "-run", "^("+strings.Join(packages[name], "|")+")", "-count=1", "-timeout=0")
			t.Logf("capture package %s: %.3fs", name, time.Since(started).Seconds())
		}()
	}
	workers.Wait()
	for _, failure := range failures {
		if failure != nil {
			return nil, failure
		}
	}

	var files []string
	err = filepath.WalkDir(capture, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".jsonl") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	type record struct {
		Rule, File, Source, Outcome, FixedSource string
		Options                                  json.RawMessage
	}
	unique := map[string]record{}
	// The raw capture is the upstream tests' own record, a file per process named by its pid, its lines in the
	// order parallel tests wrote them: never a product's bytes, which must be the same build to build (landable-15's
	// tree, Oct 11: lint-rules-agree-capture held capture-<pid>.jsonl). The rows and cases made from it are sorted.
	defer os.RemoveAll(capture)
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		for _, line := range bytes.Split(data, []byte("\n")) {
			if len(line) == 0 {
				continue
			}
			var row record
			if err := json.Unmarshal(line, &row); err != nil {
				return nil, err
			}
			if !discovered[row.Rule] {
				continue
			}
			key := fmt.Sprintf("%s\t%s\t%+v\t%s", row.Rule, row.File, row.Options, row.Source)
			unique[key] = row
		}
	}
	var keys []string
	for key := range unique {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var rows []string
	for i, key := range keys {
		row := unique[key]
		// The case keeps its file name's directories, not only its base name: a rule that judges a
		// path (a utils folder, a page directory) reads them, and Go's capture recorded them.
		name := filepath.Clean(strings.TrimLeft(strings.ReplaceAll(row.File, "\\", "/"), "/"))
		if name == "." || name == "" || strings.HasPrefix(name, "..") {
			name = filepath.Base(name)
		}
		if name == "." || name == "" || name == ".." {
			name = "source.ts"
		}
		caseDirectory := filepath.Join(directory, fmt.Sprintf("case-%03d", i))
		path := filepath.Join(caseDirectory, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, []byte(row.Source), 0644); err != nil {
			return nil, err
		}
		var legacy struct {
			Mode, Null      string
			AllowEmptyCatch bool
		}
		if len(row.Options) > 0 && row.Options[0] == '{' {
			if err := json.Unmarshal(row.Options, &legacy); err != nil {
				return nil, err
			}
		}
		mode := ""
		if row.Rule == "@typescript-eslint/method-signature-style" {
			switch row.Source {
			case "type T = { m: => void };":
				mode = "recovery"
			case "interface I", "interface I { m(a: string): void;", "interface I { m<(a: string): void; }", "interface I { m<T(a: T): T; }":
				mode = "unsupported-recovery"
			}
		}
		if row.Rule == "no-div-regex" && (row.Source == "var a = /;" || row.Source == "var a = /" || row.Source == "var a = [/];" || row.Source == "if (/) {}" || row.Source == "var a = /=") {
			mode = "recovery"
		}
		rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s\t%t\t%s\t%s", path, row.Rule, legacy.Mode, legacy.Null, legacy.AllowEmptyCatch, string(row.Options), mode))
	}
	if len(rows) < 150 {
		return nil, fmt.Errorf("capture unexpectedly small: %d cases", len(rows))
	}
	return rows, nil
}

type jsxProductInputs struct {
	Name      string
	Files     []string
	Flags     []string
	Toolchain []string
}

func jsxInputFiles(t *testing.T, roots ...string) []string {
	t.Helper()
	files := map[string]bool{}
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if entry.Name() == ".git" || entry.Name() == "node_modules" || lintRuleSubdirectory(path) {
					return filepath.SkipDir
				}
				return nil
			}
			switch filepath.Ext(path) {
			case ".go", ".ts", ".a", ".c", ".h", ".mod", ".sum", ".work", ".json":
				if strings.HasSuffix(path, "_test.go") {
					return nil
				}
				absolute, err := filepath.Abs(path)
				if err != nil {
					return err
				}
				files[absolute] = true
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	result := make([]string, 0, len(files))
	for path := range files {
		result = append(result, path)
	}
	sort.Strings(result)
	return result
}

func jsxParserOracle(t *testing.T) string {
	t.Helper()
	root, _ := filepath.Abs(filepath.Join(repository, "cohere/TypeScript/tsc"))
	side, _ := filepath.Abs(filepath.Join(repository, "stage1/typescript/parser/testdata/oracle.go"))
	return jsxGoOracle(t, "jsx-parser", root, side, "adamic_jsx_oracle.go")
}

func jsxMembershipOracle(t *testing.T) string {
	t.Helper()
	root, _ := filepath.Abs(filepath.Join(repository, "cohere"))
	side, _ := filepath.Abs("testdata/jsx_inventory.go")
	return jsxGoOracle(t, "jsx-membership", root, side, "adamic_jsx_inventory.go")
}

const repository = "../../.."

const compilerCommit = "050880ce59e30b356b686bd3144efe24f875ebc8"

func prepareRegistry(t *testing.T, directory string) []registry.Descriptor {
	t.Helper()
	descriptors, err := registry.Generate(directory)
	if err != nil {
		t.Fatal(err)
	}
	return descriptors
}

func portFiles(t *testing.T) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (path == "gaps" || strings.HasSuffix(path, "testdata") || lintRuleSubdirectory(path)) {
			return filepath.SkipDir
		}
		if !entry.IsDir() && (strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".a") || strings.HasSuffix(path, "rule.json") || strings.HasSuffix(path, "mutant.json") || strings.HasSuffix(path, "oracle.go")) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func lintRuleSubdirectory(path string) bool {
	slashed := filepath.ToSlash(filepath.Clean(path))
	if index := strings.LastIndex(slashed, "stage1/cohere/lint/rules/"); index >= 0 {
		slashed = slashed[index+len("stage1/cohere/lint/"):]
	}
	parts := strings.Split(slashed, "/")
	return len(parts) >= 3 && parts[0] == "rules"
}

type execution struct {
	output   []byte
	duration time.Duration
}

var moduleDownload = regexp.MustCompile(`(?m)^go: downloading \S+ \S+\n`)

func commandDiagnostics(name string, stderr []byte) []byte {
	if name == "go" {
		return moduleDownload.ReplaceAll(stderr, nil)
	}
	return stderr
}

func manifest(t *testing.T, rows []string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "manifest.txt")
	if err := os.WriteFile(path, []byte(strings.Join(rows, "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func recoveryRows(t *testing.T, oracle string, rows []string) []string {
	t.Helper()
	if len(rows) == 0 {
		return rows
	}
	answer := execute(t, "", oracle, "--manifest", manifest(t, rows), "--diagnostics")
	flags := strings.Fields(string(answer.output))
	if len(flags) != len(rows) {
		t.Fatalf("diagnostics answered %d rows of %d", len(flags), len(rows))
	}
	result := make([]string, len(rows))
	for index, row := range rows {
		result[index] = row
		if flags[index] != "1" {
			continue
		}
		fields := strings.Split(row, "\t")
		for len(fields) < 7 {
			fields = append(fields, "")
		}
		if fields[6] == "" {
			fields[6] = "recovery"
		}
		result[index] = strings.Join(fields, "\t")
	}
	return result
}

func generated(t *testing.T) []string {
	sources := []string{
		"debugger; if(x) debugger; while(x) debugger; label: debugger; function f(){ debugger; } function noop(){} switch(x){case 1: debugger;}",
		"if(x){} else { /* intent */ } try{}catch(e){}finally{}; switch('/*'){}; switch(x){/* intent */}; class C{static{}}",
		"a == b; 1 == 2; 1 == 1n; (typeof a) != 'number'; (null) === null; /a/ == null; `a` != 'b';",
		"var a; for(var x=0;x<3;x++){}; for(var y in z){}; for(var q of z){}; declare var b; declare global{ var c; namespace N{var d;} } namespace M{var e;}",
		"switch(x){case f( a /* c */ ):break;case f(a):break;case a?.b:break;case a.b:break;case (a):break;case a:break;case /[/*]/:break;case /[/*]/:break;case -1:break;case +1:break;case 1.0:break;case 1:break;}",
		"const é = '😀';\r\n é == null;\n debugger;\r\n if(é){} // outside\n",
		"outer: while(x){continue outer; break outer; await f(); async function nested(){await g();}} with(x){y();} new C(); void new D(); const holes=[,,x]; ([,a]=b); function* missing(){function* inner(){yield 1;} return 1;} function* good(){yield 1;} f(); var late;",
		"async function f(){for(let i=await init();await test();await update()){await body();} for await(const x of xs){await ignored();} for(const x of xs){await body();}} function g(){'use strict'; var top; f(); var late;} class C{static{f();var late;}}",
		"const placeholder='${x}${y}'; const empty='${}'; const regex=/=foo/g; const escaped=/\\=foo/; a|0; a|(0); 0|a; a>>>=b; ~a; for(a,b,c; a,b; a++,b++); for(let r=/;/; a,b; a++,b++); const arrow = a => (b,c); const deliberate = a => ((b,c));",
		"const nested=(a?true:false)?false:true; const bool=a == b ? false : true; const same=f()?true:true; const coerced=(a + b)?true:false; const assertion=(a as any)?false:true; const fallback=a?a:b??c; const nonnull=a!?true:false;",
		"// TODO: fix this xxx\nconst url='http://todo'; const fake=/[/*]TODO/; const template=`// TODO ${f() /* FIXME */} /* XXX */`; /* eslint no-warning-comments: error todo */\n// TODOé and fixme\n/** TODO decorated */ const arr=[1, // FIXME trailing comma\n]; import data from 'data' with {/* TODO attributes */}; type Empty=import('data', {with:{/* TODO import type */}});",
		"\ufeff\ufeff // TODO two marks\nconst x=1;",
		"// ſomething\n// K\n// ſa\n// Σ σ ς\n// TODO éééééééééééééééééééé é\n// TODO \\ quoted \"text\"\n// ABC TODO\n// K ſ TODO\n// -/ TODO\n",
		"async function f(){for await([, x] of values){await work();} for([,y] of values){} for([,z] in values){}}",
	}
	var rows []string
	for i, source := range sources {
		path := filepath.Join(t.TempDir(), fmt.Sprintf("generated-%d.ts", i))
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		for _, options := range []string{"all\tAlways\tAlways\tfalse", "all\tSmart\t\ttrue", "all\tAlways\tNever\tfalse", "all\tAlways\tIgnore\ttrue", "all\tAlways\tAlways\tfalse\t{\"allowLoop\":true,\"allowSwitch\":true,\"int32Hint\":true,\"allow\":[\"~\"],\"allowInParentheses\":false,\"defaultAssignment\":false,\"terms\":[\"todo\",\"fixme\",\"xxx\"],\"location\":\"anywhere\",\"decoration\":[\"*\"]}", "all\tAlways\tAlways\tfalse\t{\"Require\":\"always\",\"terms\":[]}", "no-warning-comments\t\t\tfalse\t{\"terms\":[\"s\",\"k\",\"σ\",\"todo\"],\"location\":\"anywhere\"}", "no-warning-comments\t\t\tfalse\t{\"decoration\":[\"a\",\"-\",\"z\"]}", "no-warning-comments\t\t\tfalse\t{\"decoration\":[\"z\",\"-\",\"a\"]}", "no-warning-comments\t\t\tfalse\t{\"decoration\":[\"-\",\"/\"]}"} {
			rows = append(rows, path+"\t"+options)
		}
	}
	// One source exercising the ten rules that came from the frequency-ranked slice (VOLUME.md), by default
	// and under the options each takes.
	path := filepath.Join(t.TempDir(), "options.ts")
	if err := os.WriteFile(path, []byte("interface Bare { n: number }; interface WrongType { value: number }; type Alias = string; const Choice={Yes:'Yes'} as const; function guard(x: unknown): x is string { return true; } console.log('one'); console['warn']('two'); let count=0; count++; for(let i=0;i<3;i++){count++;} interface CallableType { method(value: string): number; readonly property: (value: string) => number; } enum Direction { Left=1, Right=Left|2, Other=compute() } let boxed: Number; class Thing implements Boolean {} if(!flag){doOne();}else{doTwo();} const chosen=!flag?left:right; const assigning=()=>left=right; const wrapped=()=>(left=right); function returning(){return left=right;}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return append(rows, path,
		path+"\tno-plusplus\t\t\t\t{\"AllowForLoopAfterthoughts\":true}",
		path+"\t@typescript-eslint/method-signature-style\t\t\t\t{\"Style\":\"method\"}",
		path+"\t@typescript-eslint/prefer-literal-enum-member\t\t\t\t{\"AllowBitwiseExpressions\":true}",
		path+"\tno-return-assign\t\t\t\t\"always\"",
	)
}

func copyPort(t *testing.T, directory, from, to string, targets ...string) string {
	return copyPortImporting(t, directory, rewritePortImports, from, to, targets...)
}

func copyPortImporting(t *testing.T, directory string, imports func(t *testing.T, file, source string) string, from, to string, targets ...string) string {
	if len(targets) == 0 && from != "" {
		targets = []string{"lint.ts"}
	}
	changed := 0
	prepareRegistry(t, ".")
	for _, file := range portFiles(t) {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		source := string(data)
		selected := len(targets) == 0 || filepath.Clean(file) == filepath.Clean(targets[0])
		if from != "" && selected && !strings.HasSuffix(file, "mutant.json") && strings.Contains(source, from) {
			if strings.Count(source, from) != 1 {
				t.Fatalf("mutant anchor repeated in %s", file)
			}
			source = strings.Replace(source, from, to, 1)
			changed++
		}
		if strings.HasSuffix(file, ".ts") || strings.HasSuffix(file, ".a") {
			source = imports(t, file, source)
		}
		destination := filepath.Join(directory, file)
		if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(destination, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// Validation requires each rule's owned witnesses too.
	for _, d := range prepareRegistry(t, ".") {
		paths, err := registry.Witnesses(filepath.Join("rules", d.Slug))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range paths {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(directory, path)
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(target, data, 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	if from != "" && changed != 1 {
		t.Fatalf("mutant anchor count: %d for %q", changed, from)
	}
	return directory
}

const nativeCanaryRule = "react/jsx-no-comment-textnodes"

var portImport = regexp.MustCompile(`(?m)(^import\s+[^;]*?\s+from\s+)(['"])([^'"]+)(['"])`)

func rewritePortImports(t *testing.T, file, source string) string {
	t.Helper()
	root, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	return portImport.ReplaceAllStringFunc(source, func(declaration string) string {
		parts := portImport.FindStringSubmatch(declaration)
		if !strings.HasPrefix(parts[3], ".") {
			return declaration
		}
		absolute := filepath.Clean(filepath.Join(root, filepath.Dir(file), parts[3]))
		relative, err := filepath.Rel(root, absolute)
		if err != nil {
			t.Fatal(err)
		}
		if relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return declaration
		}
		return parts[1] + parts[2] + filepath.ToSlash(absolute) + parts[4]
	})
}

var lintOracleOnce sync.Once

var lintOraclePath string

func compilationCase(t *testing.T) string {
	t.Helper()
	paths, err := registry.Witnesses("rules/no-var")
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("profile witness corpus is empty")
	}
	return filepath.ToSlash(filepath.Join("stage1/cohere/lint", paths[0])) + "\tno-var"
}

func ownedWitnesses(t *testing.T, directory, slug string) []string {
	t.Helper()
	paths, err := registry.Witnesses(filepath.Join(directory, "rules", slug))
	if err != nil {
		t.Fatal(err)
	}
	var sources []string
	for index, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(t.TempDir(), fmt.Sprintf("%s-%d%s", slug, index, filepath.Ext(strings.TrimSuffix(path, ".txt"))))
		// A witness in a directory under testdata is linted at that relative path, for a rule that judges it.
		relative, err := filepath.Rel(filepath.Join(directory, "rules", slug, "testdata"), path)
		if err != nil {
			t.Fatal(err)
		}
		if filepath.Dir(relative) != "." {
			target = filepath.Join(t.TempDir(), strings.TrimSuffix(relative, ".txt"))
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(target, data, 0644); err != nil {
			t.Fatal(err)
		}
		sources = append(sources, target)
	}
	if len(sources) == 0 {
		t.Fatalf("%s has no witnesses", slug)
	}
	return sources
}

func ownedWitnessRows(t *testing.T, directory string, d registry.Descriptor) []string {
	t.Helper()
	paths, err := registry.Witnesses(filepath.Join(directory, "rules", d.Slug))
	if err != nil {
		t.Fatal(err)
	}
	sources := ownedWitnesses(t, directory, d.Slug)
	var rows []string
	for index, path := range paths {
		row := sources[index] + "\t" + d.Name
		sidecar := strings.TrimSuffix(strings.TrimSuffix(path, ".txt"), filepath.Ext(strings.TrimSuffix(path, ".txt"))) +
			".options.json"
		data, err := os.ReadFile(sidecar)
		if err == nil {
			var options any
			if err := json.Unmarshal(data, &options); err != nil {
				t.Fatalf("%s: %v", sidecar, err)
			}
			compact, err := json.Marshal(options)
			if err != nil {
				t.Fatal(err)
			}
			row += "\t\t\tfalse\t" + string(compact)
		} else if !os.IsNotExist(err) {
			t.Fatal(err)
		}
		rows = append(rows, row)
	}
	return rows
}

func testShardsAgreeCaptureUpstream(sourceRoot, directory string) ([]string, error) {
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		return nil, err
	}
	harness := filepath.Join(root, "internal/lint/testing/rule_testing.go")
	data, err := os.ReadFile(harness)
	if err != nil {
		return nil, err
	}
	original := "return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}"
	replacement := "result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result"
	if strings.Count(string(data), original) != 1 {
		return nil, fmt.Errorf("capture overlay anchor changed")
	}
	side := filepath.Join(directory, "rule_testing.go")
	if err := os.WriteFile(side, []byte(strings.Replace(string(data), original, replacement, 1)), 0644); err != nil {
		return nil, err
	}
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{harness: side}})
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		return nil, err
	}
	// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
	defer os.Remove(overlayPath)
	capture := filepath.Join(directory, "capture")
	environment := []string{"COHERE_DOCS_CAPTURE=" + capture}
	descriptors, err := registry.Generate(sourceRoot)
	if err != nil {
		return nil, err
	}
	discovered := map[string]bool{}
	packages := map[string][]string{}
	for _, d := range descriptors {
		discovered[d.Name] = true
		packages[d.UpstreamPackage] = append(packages[d.UpstreamPackage], d.UpstreamTest)
	}
	var names []string
	for name := range packages {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if _, err := testShardsAgreeRun(root, environment, "go", "test", "-overlay="+overlayPath, "./internal/lint/rules/"+name, "-run", "^("+strings.Join(packages[name], "|")+")", "-count=1", "-timeout=0"); err != nil {
			return nil, err
		}
	}
	files, err := filepath.Glob(filepath.Join(capture, "*.jsonl"))
	if err != nil {
		return nil, err
	}
	type record struct {
		Rule, File, Source, Outcome, FixedSource string
		Options                                  json.RawMessage
	}
	unique := map[string]record{}
	// The raw capture is the upstream tests' own record, a file per process named by its pid, its lines in the
	// order parallel tests wrote them: never a product's bytes, which must be the same build to build (landable-15's
	// tree, Oct 11: lint-rules-agree-capture held capture-<pid>.jsonl). The rows and cases made from it are sorted.
	defer os.RemoveAll(capture)
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		for _, line := range bytes.Split(data, []byte("\n")) {
			if len(line) == 0 {
				continue
			}
			var row record
			if err := json.Unmarshal(line, &row); err != nil {
				return nil, err
			}
			if !discovered[row.Rule] {
				continue
			}
			key := fmt.Sprintf("%s\t%s\t%+v\t%s", row.Rule, row.File, row.Options, row.Source)
			unique[key] = row
		}
	}
	var keys []string
	for key := range unique {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var rows []string
	for i, key := range keys {
		row := unique[key]
		// The case keeps its file name's directories, not only its base name: a rule that judges a
		// path (a utils folder, a page directory) reads them, and Go's capture recorded them.
		name := filepath.Clean(strings.TrimLeft(strings.ReplaceAll(row.File, "\\", "/"), "/"))
		if name == "." || name == "" || strings.HasPrefix(name, "..") {
			name = filepath.Base(name)
		}
		if name == "." || name == "" || name == ".." {
			name = "source.ts"
		}
		caseDirectory := filepath.Join(directory, fmt.Sprintf("case-%03d", i))
		path := filepath.Join(caseDirectory, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, []byte(row.Source), 0644); err != nil {
			return nil, err
		}
		var legacy struct {
			Mode, Null      string
			AllowEmptyCatch bool
		}
		if len(row.Options) > 0 && row.Options[0] == '{' {
			if err := json.Unmarshal(row.Options, &legacy); err != nil {
				return nil, err
			}
		}
		mode := ""
		if row.Rule == "@typescript-eslint/method-signature-style" {
			switch row.Source {
			case "type T = { m: => void };":
				mode = "recovery"
			case "interface I", "interface I { m(a: string): void;", "interface I { m<(a: string): void; }", "interface I { m<T(a: T): T; }":
				mode = "unsupported-recovery"
			}
		}
		if row.Rule == "no-div-regex" && (row.Source == "var a = /;" || row.Source == "var a = /" || row.Source == "var a = [/];" || row.Source == "if (/) {}" || row.Source == "var a = /=") {
			mode = "recovery"
		}
		rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s\t%t\t%s\t%s", path, row.Rule, legacy.Mode, legacy.Null, legacy.AllowEmptyCatch, string(row.Options), mode))
	}
	if len(rows) < 150 {
		return nil, fmt.Errorf("capture unexpectedly small: %d cases", len(rows))
	}
	return rows, nil
}

const goOraclePackage = "adamic_lint_oracle"

func goOracleOverlay(sourceRoot, directory string) (string, error) {
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		return "", err
	}
	side, err := filepath.Abs(filepath.Join(packageDirectory, "testdata/oracle.go"))
	if err != nil {
		return "", err
	}
	descriptors, err := registry.Generate(sourceRoot)
	if err != nil {
		return "", err
	}
	replacements := map[string]string{}
	var failure error
	add := func(name, source string) {
		virtual := filepath.Join(root, goOraclePackage, "adamic_lint_"+name+".go")
		absolute, err := filepath.Abs(source)
		if err != nil {
			failure = err
			return
		}
		replacements[virtual] = absolute
	}
	add("oracle", side)
	add("registry", filepath.Join(sourceRoot, ".generated/registry.go"))
	for _, d := range descriptors {
		add(strings.ReplaceAll(d.Slug, "-", "_"), filepath.Join(sourceRoot, "rules", d.Slug, "oracle.go"))
	}
	if failure != nil {
		return "", failure
	}
	overlay, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		return "", err
	}
	path := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		return "", err
	}
	return path, nil
}

func lintGoSettings() string {
	return buildcache.Tool("go", append([]string{"env"}, toolchain.GoSettings...)...)
}

func jsxTextnodesDescriptor(t *testing.T) registry.Descriptor {
	t.Helper()
	for _, d := range prepareRegistry(t, ".") {
		if d.Name == nativeCanaryRule {
			return d
		}
	}
	t.Fatal("missing JSX textnode canary")
	return registry.Descriptor{}
}

func jsxTextnodesRows(t *testing.T, oracle string) []string {
	t.Helper()
	d := jsxTextnodesDescriptor(t)
	var rows []string
	for _, row := range generated(t) {
		selected := "all"
		fields := strings.Split(row, "\t")
		if len(fields) > 1 && fields[1] != "" {
			selected = fields[1]
		}
		if selected == "all" || selected == d.Name {
			rows = append(rows, row)
		}
	}
	return append(rows, recoveryRows(t, oracle, ownedWitnessRows(t, ".", d))...)
}

type yepestaBundle struct {
	Oracle          string
	Rows, Witnesses []string
}

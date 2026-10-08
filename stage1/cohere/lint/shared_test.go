package lint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/testguard"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

// What every test in one run can share, made once (#axg2xys): the unmodified port's sanitized and release
// builds and its emitted JavaScript, the Go oracle over this package's rules, and the upstream capture.
// Each used to be rebuilt by every test that asked, which is most of the package's time and grows with
// every rule. A mutant's copy of the port is never shared: only a directory that is this package itself is.
//
// What a shared helper returns is read-only. A test that changes an artifact copies it first.
//
// sharedDirectory lives for the run; TestMain makes and removes it, so nothing a test's TempDir owns is
// ever handed to another test.
var sharedDirectory string

type sharedValue struct {
	once sync.Once
	path string
	rows []string
	// configs is each typed upstream case's program, by its file's path (captureUpstream).
	configs map[string]typedProgram
	report  string
	err     error
}

var sharedValues sync.Map

// ruleScope is the rule directories ADAMIC_LINT_RULES names, comma-separated, or nil when it's unset and the
// whole package runs (#60hxabf). A batch that only adds or changes rule directories gates on those rules: their
// upstream agreement and mutants, the generated rows that run every rule together, every rule's witnesses, the
// registry and the harness's own tests. The tests that run every rule over a whole corpus skip by name, and the
// full package stays the gate to main. The seat sets the variable only when the batch's diff touches nothing
// outside stage1/cohere/lint/rules/<slug>/.
var ruleScope = func() map[string]bool {
	named := os.Getenv("ADAMIC_LINT_RULES")
	if named == "" {
		return nil
	}
	scope := map[string]bool{}
	for _, slug := range strings.Split(named, ",") {
		if slug = strings.TrimSpace(slug); slug != "" {
			scope[slug] = true
		}
	}
	return scope
}()

// inRuleScope is whether a rule runs in this run's scope: every rule when the run isn't scoped.
func inRuleScope(descriptor registry.Descriptor) bool {
	return ruleScope == nil || ruleScope[descriptor.Slug]
}

// validateRuleScope refuses a scope naming a rule directory the registry doesn't hold, so a typo can't scope
// a run down to nothing.
func validateRuleScope(descriptors []registry.Descriptor) error {
	known := map[string]bool{}
	for _, descriptor := range descriptors {
		known[descriptor.Slug] = true
	}
	for slug := range ruleScope {
		if !known[slug] {
			return fmt.Errorf("ADAMIC_LINT_RULES names %s, which is no rule directory", slug)
		}
	}
	return nil
}

// skipWhenRuleScoped skips a test that runs every rule over a whole corpus, saying so, when the run is scoped.
func skipWhenRuleScoped(t *testing.T) {
	t.Helper()
	if ruleScope != nil {
		var slugs []string
		for slug := range ruleScope {
			slugs = append(slugs, slug)
		}
		sort.Strings(slugs)
		t.Skipf("scoped to %s by ADAMIC_LINT_RULES: this test runs every rule over a whole corpus, and the full package gates the merge to main", strings.Join(slugs, ", "))
	}
}

// shared is the run's one value for key. Its maker never calls t.Fatal: a goroutine that exits inside a
// sync.Once marks it done with nothing in it, and every later test would read an empty answer.
func shared(key string, make func(value *sharedValue)) *sharedValue {
	stored, _ := sharedValues.LoadOrStore(key, &sharedValue{})
	value := stored.(*sharedValue)
	value.once.Do(func() { make(value) })
	return value
}

var packageDirectory = func() string {
	directory, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}
	return directory
}()

// isPackage is whether directory is this package, the unmodified port, rather than a test's copy of it.
func isPackage(directory string) bool {
	absolute, err := filepath.Abs(directory)
	return err == nil && absolute == packageDirectory
}

// sharedPath is a fresh path under the run's shared directory.
func sharedPath(name string) (string, error) {
	directory, err := os.MkdirTemp(sharedDirectory, "")
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, name), nil
}

// run is execute without a test: output goes to a file, never a pipe, and anything on standard error is
// a failure, as execute requires.
func run(directory string, environment []string, name string, args ...string) ([]byte, error) {
	command := exec.Command(name, args...)
	command.Dir = directory
	// An explicit environment loses the PWD os/exec sets from Dir, and the go command trusts PWD over the
	// real working directory, so a stale one resolves the module through the wrong path.
	command.Env = append(os.Environ(), environment...)
	if directory != "" {
		command.Env = append(command.Env, "PWD="+directory)
	}
	output, err := os.CreateTemp(sharedDirectory, "stdout-")
	if err != nil {
		return nil, err
	}
	defer os.Remove(output.Name())
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := testguard.Run(command, testguard.Budget, testguard.Ceiling); err != nil || stderr.Len() != 0 {
		return nil, fmt.Errorf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	return os.ReadFile(output.Name())
}

// goOracleIn builds the Go oracle over sourceRoot's rules into directory, through an overlay inside cohere
// so the upstream rules stay unmodified.
func goOracleIn(sourceRoot, directory string) (string, error) {
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
	binary := filepath.Join(directory, "oracle")
	args := append([]string{"build", "-overlay=" + path, "-o", binary}, virtualFiles...)
	if _, err := run(root, nil, "go", args...); err != nil {
		return "", err
	}
	return binary, nil
}

// captureUpstream runs cohere's own tests for sourceRoot's rules under a capture overlay and writes each
// unique asserted case as a file under directory, returning one manifest row per case. Every Run is
// captured, including tests that assert repair fields directly; the overlay changes no rule. The capture's
// destination is passed in the subprocess's environment, so no test's process environment changes.
//
// A typed case also returns its upstream program's compiler options, by the case file's path. Upstream's
// typed tests build their own tsconfig, a JavaScript script program with allowJs for one, and the replay
// used to rebuild it as strict alone, so a case needing any other option never reached the comparison
// (#0jkpds7). The typed harness's overlay writes the tsconfig text each typed run was built from beside
// the capture, keyed the way the capture keys its cases.
func captureUpstream(sourceRoot, directory string) ([]string, map[string]typedProgram, error) {
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		return nil, nil, err
	}
	harness := filepath.Join(root, "internal/lint/testing/rule_testing.go")
	data, err := os.ReadFile(harness)
	if err != nil {
		return nil, nil, err
	}
	original := "return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}"
	replacement := "result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result"
	if strings.Count(string(data), original) != 1 {
		return nil, nil, fmt.Errorf("capture overlay anchor changed")
	}
	side := filepath.Join(directory, "rule_testing.go")
	if err := os.WriteFile(side, []byte(strings.Replace(string(data), original, replacement, 1)), 0644); err != nil {
		return nil, nil, err
	}
	typedHarness := filepath.Join(root, "internal/lint/testing/program.go")
	typedData, err := os.ReadFile(typedHarness)
	if err != nil {
		return nil, nil, fmt.Errorf("%v", err)
	}
	typedOriginal := "return Result{\n\t\tDiagnostics: diagnostics,\n\t\tSourceFile:  sourceFile,\n\t\tcapture:     newCapturedRun(subject, subjectFileName, len(files)-1, options),\n\t}"
	if strings.Count(string(typedData), typedOriginal) != 1 {
		return nil, nil, fmt.Errorf("%v", "typed capture overlay anchor changed")
	}
	typedReplacement := "result := Result{\n\t\tDiagnostics: diagnostics,\n\t\tSourceFile: sourceFile,\n\t\tcapture: newCapturedRun(subject, subjectFileName, len(files)-1, options),\n\t}\n\tRecordAssertedCase(t,result)\n" + typedConfigRecord + "\treturn result"
	typedImports := "import (\n\t\"context\"\n"
	if strings.Count(string(typedData), typedImports) != 1 {
		return nil, nil, fmt.Errorf("%v", "typed capture overlay import anchor changed")
	}
	typedSource := strings.Replace(string(typedData), typedOriginal, typedReplacement, 1)
	typedSource = strings.Replace(typedSource, typedImports, "import (\n\t\"context\"\n\t\"encoding/json\"\n", 1)
	typedSide := filepath.Join(directory, "program.go")
	if err := os.WriteFile(typedSide, []byte(typedSource), 0644); err != nil {
		return nil, nil, fmt.Errorf("%v", err)
	}
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{harness: side, typedHarness: typedSide}})
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		return nil, nil, err
	}
	capture := filepath.Join(directory, "capture")
	environment := []string{"COHERE_DOCS_CAPTURE=" + capture}
	descriptors, err := registry.Generate(sourceRoot)
	if err != nil {
		return nil, nil, err
	}
	discovered := map[string]bool{}
	typedRules := map[string]bool{}
	packages := map[string][]string{}
	for _, d := range descriptors {
		discovered[d.Name] = true
		typedRules[d.Name] = d.Typed
		packages[d.UpstreamPackage] = append(packages[d.UpstreamPackage], d.UpstreamTest)
	}
	var names []string
	for name := range packages {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if _, err := run(root, environment, "go", "test", "-overlay="+overlayPath, "./internal/lint/rules/"+name, "-run", "^("+strings.Join(packages[name], "|")+")", "-count=1", "-timeout=0"); err != nil {
			return nil, nil, err
		}
	}
	files, err := filepath.Glob(filepath.Join(capture, "*.jsonl"))
	if err != nil {
		return nil, nil, err
	}
	type record struct {
		Rule, File, Source, Outcome, FixedSource string
		Options                                  json.RawMessage
	}
	unique := map[string]record{}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, nil, err
		}
		for _, line := range bytes.Split(data, []byte("\n")) {
			if len(line) == 0 {
				continue
			}
			var row record
			if err := json.Unmarshal(line, &row); err != nil {
				return nil, nil, err
			}
			if !discovered[row.Rule] {
				continue
			}
			key := fmt.Sprintf("%s\t%s\t%+v\t%s", row.Rule, row.File, row.Options, row.Source)
			unique[key] = row
		}
	}
	typedConfigs, err := readTypedConfigs(filepath.Join(capture, typedConfigFile))
	if err != nil {
		return nil, nil, err
	}
	var keys []string
	for key := range unique {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var rows []string
	configs := map[string]typedProgram{}
	// A typed rule's case built under several compiler option sets is one case per set: one source under
	// strict and not under it is two different programs.
	caseNumber := 0
	for _, key := range keys {
		row := unique[key]
		// Only a typed rule's case replays in a program, so only its programs make separate cases.
		var variants []typedProgram
		if typedRules[row.Rule] {
			variants = typedConfigs[key]
		}
		if len(variants) == 0 {
			variants = []typedProgram{{}}
		}
		for _, variant := range variants {
			i := caseNumber
			caseNumber++
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
				return nil, nil, err
			}
			if err := os.WriteFile(path, []byte(row.Source), 0644); err != nil {
				return nil, nil, err
			}
			var legacy struct {
				Mode, Null      string
				AllowEmptyCatch bool
			}
			if len(row.Options) > 0 && row.Options[0] == '{' {
				if err := json.Unmarshal(row.Options, &legacy); err != nil {
					return nil, nil, err
				}
			}
			mode := ""
			if row.Rule == "@typescript-eslint/method-signature-style" {
				switch row.Source {
				case "type T = { m: => void };":
					mode = "recovery"
				case "interface I", "interface I { m(a: string): void;", "interface I { m<(a: string): void; }", "interface I { m<T(a: T): T; }":
					mode = "recovery"
				}
			}
			if row.Rule == "no-div-regex" && (row.Source == "var a = /;" || row.Source == "var a = /" || row.Source == "var a = [/];" || row.Source == "if (/) {}" || row.Source == "var a = /=") {
				mode = "recovery"
			}
			rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s\t%t\t%s\t%s", path, row.Rule, legacy.Mode, legacy.Null, legacy.AllowEmptyCatch, string(row.Options), mode))
			if variant.CompilerOptions != "" {
				configs[path] = variant
			}
		}
	}
	if len(rows) < 150 {
		return nil, nil, fmt.Errorf("capture unexpectedly small: %d cases", len(rows))
	}
	// Captured fixtures are replayable inputs, including parser diagnostics that
	// cohere's own rule tests deliberately run on recovered trees. Carry that
	// boundary in the manifest rather than requiring every consumer to rediscover it.
	oracleDirectory := filepath.Join(directory, "recovery-oracle")
	if err := os.MkdirAll(oracleDirectory, 0755); err != nil {
		return nil, nil, err
	}
	oracle, err := goOracleIn(sourceRoot, oracleDirectory)
	if err != nil {
		return nil, nil, err
	}
	manifest := filepath.Join(directory, "recovery-inputs.txt")
	if err := os.WriteFile(manifest, []byte(strings.Join(rows, "\n")+"\n"), 0644); err != nil {
		return nil, nil, err
	}
	flags, err := run("", nil, oracle, "--manifest", manifest, "--diagnostics")
	if err != nil {
		return nil, nil, err
	}
	classified, err := classifyRecoveryRows(rows, strings.Fields(string(flags)))
	return classified, configs, err
}

// typedConfigFile is where the typed harness's overlay writes each typed run's tsconfig, one JSON line per
// run, beside the capture's own records and outside its *.jsonl glob.
const typedConfigFile = "typed-configs.txt"

// typedConfigRecord is the code the overlay puts after a typed run's result: one line holding the case's key,
// the capture's own rule, file, options and source, the tsconfig the run's program was built from, and how
// many findings the rule reported there, which the replay must reproduce under the options it writes.
// Options are encoded as the capture encodes them, absent when the run had none, so the keys match. A run
// with a setup hook built from the tsconfig on disk, the hook's own or the default; a cached run built in
// memory from defaultTsConfig and wrote none, so a missing file means that one.
const typedConfigRecord = `	if captureDirectory := os.Getenv("COHERE_DOCS_CAPTURE"); captureDirectory != "" {
		configText, err := os.ReadFile(filepath.Join(directory, "tsconfig.json"))
		if err != nil {
			configText = []byte(defaultTsConfig)
		}
		var encodedOptions json.RawMessage
		if options != nil {
			encodedOptions, _ = json.Marshal(options)
		}
		line, _ := json.Marshal(struct {
			Rule, File, Source, TsConfig string
			Findings, OtherFiles         int
			Options                      json.RawMessage ` + "`json:\",omitempty\"`" + `
		}{subject.Name, subjectFileName, sourceFile.Text(), string(configText), len(diagnostics), len(files) - 1, encodedOptions})
		if file, err := os.OpenFile(filepath.Join(captureDirectory, "` + typedConfigFile + `"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
			file.Write(append(line, '\n'))
			file.Close()
		}
	}
`

// typedProgram is the program upstream's test built one typed case in: its compiler options, and how many
// findings the rule reported there. Findings is -1 when that program held other fixture files: the replay lints
// the case's file alone, so it can't reproduce that program's count.
type typedProgram struct {
	CompilerOptions string
	Findings        int
}

// readTypedConfigs is the distinct programs each typed case was built in, by the capture's case key, sorted by
// their compiler options so the case numbering is stable. A run whose tsconfig can't be read or parsed fails
// here, and so do two single-file runs of one case under one program that reported different counts. A case
// also run with other fixture files keeps its options, and its count is unknown (-1).
func readTypedConfigs(path string) (map[string][]typedProgram, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return map[string][]typedProgram{}, nil
	}
	if err != nil {
		return nil, err
	}
	distinct := map[string]map[string]int{}
	for _, line := range bytes.Split(data, []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		var record struct {
			Rule, File, Source, TsConfig string
			Findings, OtherFiles         int
			Options                      json.RawMessage
		}
		if err := json.Unmarshal(line, &record); err != nil {
			return nil, fmt.Errorf("typed config record: %v", err)
		}
		var tsconfig struct {
			CompilerOptions json.RawMessage `json:"compilerOptions"`
		}
		if err := json.Unmarshal([]byte(record.TsConfig), &tsconfig); err != nil || len(tsconfig.CompilerOptions) == 0 {
			return nil, fmt.Errorf("%s %s: typed run's tsconfig has no readable compilerOptions (%v): %q", record.Rule, record.File, err, record.TsConfig)
		}
		key := fmt.Sprintf("%s\t%s\t%+v\t%s", record.Rule, record.File, record.Options, record.Source)
		options := string(tsconfig.CompilerOptions)
		if distinct[key] == nil {
			distinct[key] = map[string]int{}
		}
		findings := record.Findings
		if record.OtherFiles > 0 {
			findings = -1
		}
		previous, seen := distinct[key][options]
		switch {
		case !seen:
			distinct[key][options] = findings
		case previous == -1 || findings == -1:
			distinct[key][options] = -1
		case previous != findings:
			return nil, fmt.Errorf("%s %s: one case under one single-file program reported %d findings and %d", record.Rule, record.File, previous, findings)
		}
	}
	programs := map[string][]typedProgram{}
	for key, sets := range distinct {
		for options, findings := range sets {
			programs[key] = append(programs[key], typedProgram{options, findings})
		}
		sort.Slice(programs[key], func(i, j int) bool { return programs[key][i].CompilerOptions < programs[key][j].CompilerOptions })
	}
	return programs, nil
}

// A nonempty mode belongs to its caller. In particular, unsupported-recovery
// stays visible as a port limitation rather than becoming a successful port case.
func classifyRecoveryRows(rows, flags []string) ([]string, error) {
	if len(flags) != len(rows) {
		return nil, fmt.Errorf("diagnostics answered %d rows of %d", len(flags), len(rows))
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
	return result, nil
}

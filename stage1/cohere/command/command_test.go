package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func absolute(t *testing.T, path string) string {
	t.Helper()
	result, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
func write(t *testing.T, path, text string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte(text), mode); err != nil {
		t.Fatal(err)
	}
}
func shellQuote(text string) string { return "'" + strings.ReplaceAll(text, "'", "'\\''") + "'" }
func wrapper(t *testing.T, source string, backend bool) string {
	t.Helper()
	if backend {
		program := lowered(t, source)
		path := filepath.Join(t.TempDir(), "backend.mjs")
		write(t, path, javascript.JavaScript(program), 0644)
		source = path
	}
	runner := absolute(t, filepath.Join(repository, "oracle/node.mjs"))
	path := filepath.Join(t.TempDir(), "node-probe")
	write(t, path, "#!/bin/sh\nexec node --disable-warning=ExperimentalWarning "+shellQuote(runner)+" "+shellQuote(source)+" \"$@\"\n", 0755)
	return path
}
func originalTests(t *testing.T, source string, nativeProbe string) run {
	t.Helper()
	root := absolute(t, filepath.Join(repository, "cohere"))
	replace := map[string]string{}
	for _, entry := range []struct {
		file  string
		names []string
	}{{"ownership.go", []string{"resolveOwnership"}}, {"projects.go", []string{"childArguments", "discoveryApplies"}}, {"discovery.go", []string{"discoveryRoot"}}, {"root.go", []string{"locateProject"}}, {"scope.go", []string{"namedPathsScope", "narrowTo", "narrowToEnumeration"}}} {
		original := filepath.Join(root, "command/cohere", entry.file)
		content, err := os.ReadFile(original)
		if err != nil {
			t.Fatal(err)
		}
		text := string(content)
		for _, name := range entry.names {
			prefix := "func " + name + "("
			changed := "func adamicGo" + strings.ToUpper(name[:1]) + name[1:] + "("
			if name == "narrowTo" || name == "narrowToEnumeration" {
				prefix = "func (s formatScope) " + name + "("
				changed = "func (s formatScope) adamicGo" + strings.ToUpper(name[:1]) + name[1:] + "("
			}
			if strings.Count(text, prefix) != 1 {
				t.Fatalf("unknown Go boundary %s", prefix)
			}
			text = strings.Replace(text, prefix, changed, 1)
		}
		path := filepath.Join(t.TempDir(), entry.file)
		write(t, path, text, 0644)
		replace[original] = path
	}
	replace[filepath.Join(root, "command/cohere/adamic_boundaries_test.go")] = absolute(t, "testdata/boundaries.go.txt")
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	raw, err := json.Marshal(map[string]any{"Replace": replace})
	if err != nil {
		t.Fatal(err)
	}
	write(t, overlay, string(raw), 0644)
	filter := "^(TestDiscoveryRootIsTheRepositoryTheCallerStandsIn|TestDiscoveryAppliesOnlyToARunThatNamesNoProject|TestAdamicScopeTiming|TestAdamicCommandFilesystemEdges|TestASolutionRootIsNotRunAndWhatItReferencesIs|TestNestedProjectsAreFoundUnderAnyRoot|TestAFileIncludedTwiceBelongsToTheNearerTsconfig|TestASharedFileNeitherHoldsGoesByPathThenOrder|TestOneProjectIsNotRead|TestChildArgumentsCarryTheLintConfigAbsolute|TestLocateProject.*|TestDotFromASubdirectoryIsThatSubdirectory|TestNamedPathsScope.*|TestTheScopeStatesBothCounts|TestNarrowingAWholeTreeScopeChangesNothing|TestTheEmptyScopeNamingSurvivesNarrowing|TestAWholeTreeScopeBecomesTheEnumeration|TestDeclinedExtensionsAreNamedRatherThanOmitted|TestTheEnumerationDescriptionNamesTheRootAndTheCounts|TestIgnoreLayersAndNestedRepositoriesAreReported|TestANarrowScopeReportsHowMuchIsFormattable|TestAnEmptyScopeSurvivesEnumerationNarrowing|TestNarrowToUsesTheWholeProgramNotTheLintScope|TestNarrowToAppendsToTheScopesOwnWording)$"
	command := bounded(t, "go", "test", "-count=1", "-v", "-overlay="+overlay, "-run="+filter, "./command/cohere")
	command.Dir = root
	command.Env = append(os.Environ(), "ADAMIC_NATIVE_PROBE="+nativeProbe, "ADAMIC_NODE_PROBE="+wrapper(t, source, false), "ADAMIC_BACKEND_PROBE="+wrapper(t, source, true))
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err = command.Run()
	if command.ProcessState == nil {
		t.Fatal(err)
	}
	return run{stdout.Bytes(), stderr.Bytes(), command.ProcessState.ExitCode()}
}
func TestOriginalCommandBoundaries(t *testing.T) {
	source := absolute(t, "probe.ts")
	program := lowered(t, source)
	binary := filepath.Join(t.TempDir(), "probe")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := originalTests(t, source, binary)
	if got.exitCode != 0 {
		t.Fatalf("original tests: exit %d\n%s\n%s", got.exitCode, got.stdout, got.stderr)
	}
	t.Logf("%s", got.stdout)
}

// Binding only the host exit status leaves stdout/stderr and every command decision in TypeScript.
// Assert the emitter's main epilogue instead of silently patching an unrelated return.
func cliC(t *testing.T, program *ir.Program) string {
	t.Helper()
	index := -1
	for at, local := range program.Locals {
		if local.Global && local.Name == "commandExitCode" {
			if index >= 0 {
				t.Fatal("duplicate exit status")
			}
			index = at
		}
	}
	if index < 0 {
		t.Fatal("missing exit status")
	}
	c := native.C(program)
	epilogue := "\treturn 0;\n}\n"
	if !strings.HasSuffix(c, epilogue) {
		t.Fatal("native main epilogue changed")
	}
	return strings.TrimSuffix(c, epilogue) + fmt.Sprintf("\treturn (int)adamic_global_%d_commandExitCode;\n}\n", index)
}
func buildCLI(t *testing.T, program *ir.Program, binary string) {
	t.Helper()
	if err := native.Build(cliC(t, program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
}

func TestFlagErrorsAndHelpMatchGoCLI(t *testing.T) {
	root := absolute(t, filepath.Join(repository, "cohere"))
	goBinary := filepath.Join(t.TempDir(), "cohere")
	build := bounded(t, "go", "build", "-o", goBinary, "./command/cohere")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("Go build %v: %s", err, output)
	}
	source := absolute(t, "main.ts")
	program := lowered(t, source)
	binary := filepath.Join(t.TempDir(), "port")
	if err := native.Build(cliC(t, program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	js := javascript.JavaScript(program)
	status := ""
	for index, local := range program.Locals {
		if local.Global && local.Name == "commandExitCode" {
			status = javascript.Name(program, index)
		}
	}
	if status == "" {
		t.Fatal("missing JavaScript exit status")
	}
	backend := filepath.Join(t.TempDir(), "cli.mjs")
	write(t, backend, js+"\nexport const commandExitCode = "+status+";\n", 0644)
	adapter := absolute(t, "testdata/cli.mjs")
	runner := absolute(t, filepath.Join(repository, "oracle/node.mjs"))
	// The argument transported ahead of the user's arguments is argv[0], used only by Go's help.
	cases := [][]string{{"rename", "--help"}, {"rename", "--unknown"}, {"rename", "--write=bad"}, {"rename"}, {"rename", "target"}, {"rename", "a", "b", "c"}, {"rename", "--write", "--dry-run", "target", "newName"}, {"rename", "--tsconfig"}, {"--help"}, {"-h"}, {"--help=ignored"}, {"--unknown"}, {"---types"}, {"-=bad"}, {"--tsconfig"}, {"--fix-passes"}, {"--types=wrong"}, {"--fix-passes=bad"}, {"--fix-passes=9223372036854775808"}, {"--fix-passes=-9223372036854775809"}, {"--fix-passes=08"}, {"--fix-passes=1__0"}, {"--verbose", "--json"}, {"--fix-passes=0x7fffffffffffffff", "--help"}, {"--fix-passes=-0x8000000000000000", "--help"}, {"--fix-passes=0b1_0", "--help"}, {"--fix-passes=0o_17", "--help"}, {"--fix-passes=0_17", "--help"}, {"--fix-passes=+10", "--help"}, {"--fix-passes=-0", "--help"}, {"--types=FALSE", "--help"}, {"--types=TrUe"}, {"--types=false", "--types=true", "--help"}, {"--tsconfig=", "--help"}, {"--directory", "-h", "--help"}, {"--no-cache=false", "--help"}}
	for _, args := range cases {
		want := execute(t, nil, goBinary, args...)
		nativeResult := execute(t, nil, binary, append([]string{goBinary}, args...)...)
		nodeArguments := []string{"--disable-warning=ExperimentalWarning", adapter, source, runner, goBinary}
		backendArguments := []string{"--disable-warning=ExperimentalWarning", adapter, backend, runner, goBinary}
		for _, side := range []struct {
			name   string
			result run
		}{{"native", nativeResult}, {"Node", execute(t, nil, "node", append(nodeArguments, args...)...)}, {"backend", execute(t, nil, "node", append(backendArguments, args...)...)}} {
			got := side.result
			if got.exitCode != want.exitCode || !bytes.Equal(got.stdout, want.stdout) || !bytes.Equal(got.stderr, want.stderr) {
				t.Fatalf("%q\ngot exit %d stdout %q stderr %s\nGo exit %d stdout %q stderr %s", args, got.exitCode, got.stdout, got.stderr, want.exitCode, want.stdout, want.stderr)
			}
		}
	}
	t.Logf("%d actual CLI cases: stdout, stderr, exit status; ASan, UBSan and LeakSanitizer", len(cases))
}

func TestParsedArgumentsMatchGoFlag(t *testing.T) {
	oracle := filepath.Join(t.TempDir(), "flags")
	build := bounded(t, "go", "build", "-o", oracle, "testdata/definitions.go")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("flag oracle %v: %s", err, output)
	}
	source := absolute(t, "probe.ts")
	program := lowered(t, source)
	binary := filepath.Join(t.TempDir(), "probe")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	node, backend := wrapper(t, source, false), wrapper(t, source, true)
	cases := [][]string{nil, {"--"}, {"-"}, {"file.ts", "--types"}, {"--types", "false", "--help"}, {"--types=false", "--types=true", "one", "two"}, {"--tsconfig", "-h", "--"}, {"--help", "remaining"}, {"--unknown", "remaining"}, {"--fix-passes=0_0", "--types"}, {"--fix-passes=0", "--types"}, {"--fix-passes=0123"}}
	for _, value := range []string{"", "true", "TRUE", "True", "t", "T", "1", "false", "FALSE", "False", "f", "F", "0", "TrUe", "yes", "😀", "\n", "\x01", "\u00a0"} {
		cases = append(cases, []string{"--types=" + value, "rest"})
	}
	for _, value := range []string{"0", "00", "0_0", "0_", "0__0", "01_2", "+0", "-0", "+", "-", "_1", "1_", "1__0", "1_0", "0x_1", "0b_1", "0o_1", "0x", "0b", "0o", "0x__1", "0b2", "08", "9223372036854775807", "9223372036854775808", "-9223372036854775808", "-9223372036854775809", "0x7fffffffffffffff", "0x8000000000000000", "-0x8000000000000000", "18446744073709551615", "18446744073709551616", "9999999999999999999999999bad", "9999999999999999999999999_", "_9999999999999999999999999", "0x__ffffffffffffffffffffffff", "1__999999999999999999999999", "9223372036854775808bad", "9223372036854775808_", "18446744073709551615_", "18446744073709551616_", "-18446744073709551616bad"} {
		cases = append(cases, []string{"--fix-passes=" + value, "rest"})
	}
	for _, arguments := range cases {
		input, err := json.Marshal(map[string]any{"kind": "flags", "program": "cohere", "arguments": arguments})
		if err != nil {
			t.Fatal(err)
		}
		want := execute(t, nil, oracle, absolute(t, filepath.Join(repository, "cohere/command/cohere/main.go")), string(input))
		if want.exitCode != 0 {
			t.Fatalf("Go flag oracle: %s", want.stderr)
		}
		for _, side := range []string{binary, node, backend} {
			got := execute(t, nil, side, string(input))
			if got.exitCode != 0 || len(got.stderr) > 0 || !bytes.Equal(got.stdout, want.stdout) {
				t.Fatalf("%s %q\ngot exit %d stderr %s stdout %s\nGo %s", side, arguments, got.exitCode, got.stderr, got.stdout, want.stdout)
			}
		}
	}
	// Go's default usage omits "of" when argv[0] is empty.
	input := `{"kind":"flags","program":"","arguments":["--help"]}`
	want := execute(t, nil, oracle, absolute(t, filepath.Join(repository, "cohere/command/cohere/main.go")), input)
	if want.exitCode != 0 || len(want.stderr) > 0 {
		t.Fatal("empty-name Go oracle failed")
	}
	for _, side := range []string{binary, node, backend} {
		got := execute(t, nil, side, input)
		if got.exitCode != 0 || len(got.stderr) > 0 || !bytes.Equal(got.stdout, want.stdout) {
			t.Fatalf("empty argv[0] differs: %s %s", got.stdout, want.stdout)
		}
	}
	t.Logf("%d parsed argument cases agree across Go flag, native, Node and backend", len(cases)+1)
}

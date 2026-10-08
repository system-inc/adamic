package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Use the scout sources unchanged, including the compiler's inactive profiler
// callback. Mutants are source changes, never edits to runtime's C primitives.
func TestNodeStartupHostShapes(t *testing.T) {
	for _, one := range []struct{ file, before, after string }{
		{"15_getExecutingFilePath.a", `__filename.endsWith("sys.js")`, `__filename.endsWith("sys.cjs")`},
		{"17_write.a", "process.stdout.write(s);", `process.stdout.write(s + "?");`},
		{"18_exit_0.a", "nodeSystem.exit(0);", "nodeSystem.exit(1);"},
		{"19_exit_1.a", "nodeSystem.exit(1);", "nodeSystem.exit(2);"},
		{"20_exit_2.a", "nodeSystem.exit(2);", "nodeSystem.exit(0);"},
		{"23_newLine.a", "newLine: _os.EOL", `newLine: "\r\n"`},
	} {
		t.Run(one.file, func(t *testing.T) {
			fixture := "stage3/fixtures/host/" + one.file
			absolute, err := filepath.Abs(filepath.Join(repository, fixture))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := lowered(t, absolute); err != nil {
				var missing *lower.NotYet
				if !errors.As(err, &missing) {
					t.Fatal(err)
				}
				reason := "literal method call with an unrepresented result"
				if strings.Contains(one.file, "exit_") {
					reason = "value of type undefined"
				}
				if !strings.Contains(missing.What, reason) {
					t.Fatalf("unexpected compiler prerequisite: %v", missing)
				}
				truth := onNode(t, absolute)
				data, readErr := os.ReadFile(absolute)
				if readErr != nil {
					t.Fatal(readErr)
				}
				if strings.Count(string(data), one.before) != 1 {
					t.Fatal("source mutant anchor changed")
				}
				mutant := filepath.Join(t.TempDir(), "mutant.a")
				if err := os.WriteFile(mutant, []byte(strings.Replace(string(data), one.before, one.after, 1)), 0600); err != nil {
					t.Fatal(err)
				}
				want := "stdout differs"
				if strings.Contains(one.file, "exit_") {
					want = "exit codes differ"
				}
				bad := onNode(t, mutant)
				if disagreement(truth, bad) != want || len(bad.stderr) != 0 {
					t.Fatal("Node source mutant survived")
				}
				t.Skipf("Node source mutant caught; compiler prerequisite: %v", missing)
			}
			path, binary, script := sanitized(t, fixture)
			truth := onNode(t, path)
			for _, got := range []run{execute(t, binary), onNode(t, script)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Fatal(difference)
				}
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(data), one.before) != 1 {
				t.Fatal("source mutant anchor changed")
			}
			mutated := filepath.Join(t.TempDir(), "mutant.a")
			if err := os.WriteFile(mutated, []byte(strings.Replace(string(data), one.before, one.after, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			_, mutantBinary, mutantScript := nodeStartupSanitizedMutant(t, mutated)
			want := "stdout differs"
			if strings.Contains(one.file, "exit_") {
				want = "exit codes differ"
			}
			results := []run{onNode(t, mutated), execute(t, mutantBinary), onNode(t, mutantScript)}
			for _, bad := range results {
				if difference := disagreement(truth, bad); difference != want || len(bad.stderr) != 0 {
					t.Fatalf("source mutant must fail only Node comparison: %s; stderr %q", difference, bad.stderr)
				}
			}
			t.Log("original agrees; source mutant compiles and runs, caught only by Node comparison")
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				t.Run("WASI", func(t *testing.T) {
					if difference := disagreement(truth, onWASI(t, native.C(program))); difference != "" {
						t.Fatal(difference)
					}
					mutantProgram, err := lowered(t, mutated)
					if err != nil {
						t.Fatal(err)
					}
					bad := onWASI(t, native.C(mutantProgram))
					if disagreement(truth, bad) != want || len(bad.stderr) != 0 {
						t.Fatal("WASI source mutant survived")
					}
				})
			}
		})
	}
}

func TestNodeStartupBindings(t *testing.T) {
	for _, one := range []struct {
		name, before, after string
		identity            bool
	}{
		{"clocks", "next >= first", "next < first", false},
		{"memory", "usage.rss > 0", "usage.rss < 0", false},
		{"identity", "path.dirname(__filename) === __dirname", "path.dirname(__filename) !== __dirname", true},
	} {
		t.Run(one.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/node_startup_"+one.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			directory := t.TempDir()
			source := filepath.Join(directory, "source.a")
			if err := os.WriteFile(source, data, 0600); err != nil {
				t.Fatal(err)
			}
			build := func(path string) (string, string, string) {
				p, err := lowered(t, path)
				if err != nil {
					t.Fatal(err)
				}
				c := native.C(p)
				binary := filepath.Join(t.TempDir(), "program")
				if err := native.Build(c, binary, native.Options{Sanitize: true}); err != nil {
					t.Fatal(err)
				}
				script := binary + ".mjs"
				if err := os.WriteFile(script, []byte(javascript.JavaScript(p)), 0600); err != nil {
					t.Fatal(err)
				}
				if one.identity {
					entries, err := os.ReadDir(filepath.Dir(binary))
					if err != nil {
						t.Fatal(err)
					}
					for _, entry := range entries {
						if strings.HasPrefix(entry.Name(), "lib.") {
							content, err := os.ReadFile(filepath.Join(filepath.Dir(binary), entry.Name()))
							if err != nil {
								t.Fatal(err)
							}
							if err := os.WriteFile(filepath.Join(directory, entry.Name()), content, 0600); err != nil {
								t.Fatal(err)
							}
						}
					}
				}
				return binary, script, c
			}
			truthRun := func(path string) run {
				if !one.identity {
					return onNode(t, path)
				}
				return execute(t, "node", "--disable-warning=ExperimentalWarning", "-e", `const fs=require('node:fs'),m=require('node:module'),p=require('node:path'); const file=process.argv[1]; new Function('require','__filename','__dirname',m.stripTypeScriptTypes(fs.readFileSync(file,'utf8'),{mode:'transform'}))(m.createRequire(file),file,p.dirname(file));`, path)
			}
			binary, script, c := build(source)
			truth := truthRun(source)
			for _, got := range []run{execute(t, binary), onNode(t, script)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Fatalf("%s: Node %q %q; backend %q %q", difference, truth.stdout, truth.stderr, got.stdout, got.stderr)
				}
			}
			if strings.Count(string(data), one.before) != 1 {
				t.Fatal("mutant anchor changed")
			}
			mutated := filepath.Join(directory, "mutant.a")
			if err := os.WriteFile(mutated, []byte(strings.Replace(string(data), one.before, one.after, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			badBinary, badScript, badC := build(mutated)
			for _, bad := range []run{truthRun(mutated), execute(t, badBinary), onNode(t, badScript)} {
				if disagreement(truth, bad) != "stdout differs" || bad.exitCode != 0 || len(bad.stderr) != 0 {
					t.Fatalf("mutant did not fail only Node comparison: %q %q", bad.stdout, bad.stderr)
				}
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				t.Run("WASI", func(t *testing.T) {
					if difference := disagreement(truth, onWASI(t, c)); difference != "" {
						t.Fatal(difference)
					}
					if bad := onWASI(t, badC); disagreement(truth, bad) != "stdout differs" || len(bad.stderr) != 0 {
						t.Fatal("WASI mutant survived")
					}
				})
			}
		})
	}
}

func TestNodeStartupExitBindings(t *testing.T) {
	for _, code := range []string{"0", "1", "2"} {
		t.Run(code, func(t *testing.T) {
			directory := t.TempDir()
			source := filepath.Join(directory, "exit.a")
			text := `import process from 'node:process'; process.exitCode=2; console.log(String(process.exitCode)); process.stdout.write('before'); try { process.exit(` + code + `); } finally { process.stdout.write('finally'); }`
			if err := os.WriteFile(source, []byte(text), 0600); err != nil {
				t.Fatal(err)
			}
			_, binary, script := nodeStartupSanitizedMutant(t, source)
			truth := onNode(t, source)
			for _, got := range []run{execute(t, binary), onNode(t, script)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Fatalf("%s: %q %q", difference, got.stdout, got.stderr)
				}
			}
			mutated := filepath.Join(directory, "mutant.a")
			next := map[string]string{"0": "1", "1": "2", "2": "0"}[code]
			if err := os.WriteFile(mutated, []byte(strings.Replace(text, "process.exit("+code+")", "process.exit("+next+")", 1)), 0600); err != nil {
				t.Fatal(err)
			}
			_, badBinary, badScript := nodeStartupSanitizedMutant(t, mutated)
			for _, bad := range []run{onNode(t, mutated), execute(t, badBinary), onNode(t, badScript)} {
				if disagreement(truth, bad) != "exit codes differ" || len(bad.stderr) != 0 || string(bad.stdout) != string(truth.stdout) {
					t.Fatal("exit mutant must fail only Node status comparison")
				}
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				p, err := lowered(t, source)
				if err != nil {
					t.Fatal(err)
				}
				if difference := disagreement(truth, onWASI(t, native.C(p))); difference != "" {
					t.Fatal(difference)
				}
				bad, err := lowered(t, mutated)
				if err != nil {
					t.Fatal(err)
				}
				if got := onWASI(t, native.C(bad)); disagreement(truth, got) != "exit codes differ" || len(got.stderr) != 0 || string(got.stdout) != string(truth.stdout) {
					t.Fatal("WASI exit mutant survived")
				}
			}
		})
	}
}

func TestNodeStartupRemainingScoutShapes(t *testing.T) {
	for _, one := range []struct{ file, reason, before, after string }{
		{"14_getCurrentDirectory.a", "non-null assertion", "callback = undefined!;", "// callback stays live"},
		{"16_getEnvironmentVariable.a", "BinaryExpression with a string and a string", `return process.env[name] || "";`, `return "";`},
	} {
		t.Run(one.file, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/host", one.file))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, path)
			if err == nil || !strings.Contains(err.Error(), one.reason) {
				t.Fatalf("want recorded compiler prerequisite %s, got %v", one.reason, err)
			}
			t.Logf("blocked before emission: %v", err)
			truth := onNode(t, path)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Count(string(data), one.before) != 1 {
				t.Fatal("source mutant anchor changed")
			}
			mutant := filepath.Join(t.TempDir(), "mutant.a")
			if err := os.WriteFile(mutant, []byte(strings.Replace(string(data), one.before, one.after, 1)), 0600); err != nil {
				t.Fatal(err)
			}
			bad := onNode(t, mutant)
			if disagreement(truth, bad) != "stdout differs" || bad.exitCode != 0 || len(bad.stderr) != 0 {
				t.Fatal("Node source mutant survived")
			}
			t.Log("Node source mutant caught; no native or JavaScript-backend success claimed")
		})
	}
}

// Not parallel: supply the process environment inherited by all three runners.
func TestNodeStartupEnvironmentReads(t *testing.T) {
	t.Setenv("ADAMIC_GATE_UNCACHED", "1")
	t.Setenv("ADAMIC_HOST29_ABSENT", "")
	if err := os.Unsetenv("ADAMIC_HOST29_ABSENT"); err != nil {
		t.Fatal(err)
	}
	path, binary, script := sanitized(t, "internal/oracle/testdata/node_startup_environment.a")
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"", "0", "héllo 🌍"} {
		t.Setenv("ADAMIC_HOST29_ENV", text)
		truth := onNode(t, path)
		results := []run{execute(t, binary), onNode(t, script)}
		if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
			results = append(results, onWASI(t, native.C(program)))
		}
		for _, got := range results {
			if difference := disagreement(truth, got); difference != "" {
				t.Fatal(difference)
			}
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	anchor := `return process.env[name] ?? "";`
	if strings.Count(string(data), anchor) != 1 {
		t.Fatal("source mutant anchor changed")
	}
	mutant := filepath.Join(t.TempDir(), "environment-mutant.a")
	if err := os.WriteFile(mutant, []byte(strings.Replace(string(data), anchor, `return "";`, 1)), 0600); err != nil {
		t.Fatal(err)
	}
	_, mutantBinary, mutantScript := nodeStartupSanitizedMutant(t, mutant)
	truth := onNode(t, path)
	results := []run{onNode(t, mutant), execute(t, mutantBinary), onNode(t, mutantScript)}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		mutantProgram, err := lowered(t, mutant)
		if err != nil {
			t.Fatal(err)
		}
		results = append(results, onWASI(t, native.C(mutantProgram)))
	}
	for _, bad := range results {
		if disagreement(truth, bad) != "stdout differs" || bad.exitCode != 0 || len(bad.stderr) != 0 {
			t.Fatal("environment mutant must fail only Node stdout comparison")
		}
	}
}

func nodeStartupSanitizedMutant(t *testing.T, path string) (string, string, string) {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	relative, err := filepath.Rel(root, path)
	if err != nil {
		t.Fatal(err)
	}
	return sanitized(t, relative)
}

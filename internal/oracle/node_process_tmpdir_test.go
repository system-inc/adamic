package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNodeProcessTmpdir(test *testing.T) {
	test.Parallel()
	path, binary, script := sanitized(test, "internal/oracle/testdata/node_process_tmpdir.a")
	for _, environment := range [][]string{
		{"TMPDIR=", "TMP=", "TEMP="},
		{"TMPDIR=/", "TMP=second", "TEMP=third"},
		{"TMPDIR=//", "TMP=second", "TEMP=third"},
		{"TMPDIR=/tmp///", "TMP=second", "TEMP=third"},
		{"TMPDIR=relative/", "TMP=second", "TEMP=third"},
		{"TMPDIR=héllo 🌍/", "TMP=second", "TEMP=third"},
		{"TMPDIR=", "TMP=second/", "TEMP=third/"},
		{"TMPDIR=", "TMP=", "TEMP=third/"},
		{"TMPDIR=bad\xff/", "TMP=", "TEMP="},
	} {
		runner := filepath.Join(repository, "oracle/node.mjs")
		truth := executeWith(test, environment, "node", "--disable-warning=ExperimentalWarning", runner, path)
		for _, got := range []run{
			executeWith(test, environment, binary),
			executeWith(test, environment, "node", "--disable-warning=ExperimentalWarning", runner, script),
		} {
			if difference := disagreement(truth, got); difference != "" {
				test.Fatalf("%q: %s, Node %q, got %q %q", environment, difference, truth.stdout, got.stdout, got.stderr)
			}
		}
		test.Logf("%q -> %q", environment, truth.stdout)
	}
	path, mutant := nodeProcessMutant(
		test,
		"node_process_tmpdir.a",
		"if (length > 1 && value->bytes[length - 1] == '/') { length--; }",
		"if (length > 1 && value->bytes[length - 1] == '/') { length -= 2; }",
	)
	environment := []string{"TMPDIR=relative/", "TMP=", "TEMP="}
	truth := executeWith(
		test,
		environment,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle/node.mjs"),
		path,
	)
	bad := executeWith(test, environment, mutant)
	if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
		test.Fatal("tmpdir mutant not caught only by Node bytes")
	}
	test.Log("clean tmpdir mutant caught by Node stdout")
}

func TestNodeProcessOSNamespace(test *testing.T) {
	test.Parallel()
	path, binary, script := sanitized(test, "internal/oracle/testdata/node_process_os_namespace.a")
	truth := onNode(test, path)
	for _, got := range []run{execute(test, binary), onNode(test, script)} {
		if difference := disagreement(truth, got); difference != "" {
			test.Fatalf("%s: %q %q", difference, got.stdout, got.stderr)
		}
	}
	_, mutant := nodeProcessMutant(
		test,
		"node_process_os_namespace.a",
		"return adamic_host_eol();",
		"static adamic_string newline = ADAMIC_STRING(\"\\r\\n\"); return &newline;",
	)
	bad := execute(test, mutant)
	if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
		test.Fatal("namespace EOL mutant not caught only by Node bytes")
	}
	data, readError := os.ReadFile(script)
	if readError != nil {
		test.Fatal(readError)
	}
	if !strings.Contains(string(data), `"\n"`) {
		test.Fatal("JavaScript EOL mutant anchor changed")
	}
	javascriptMutant := filepath.Join(test.TempDir(), "namespace-mutant.mjs")
	if writeError := os.WriteFile(
		javascriptMutant,
		[]byte(strings.ReplaceAll(string(data), `"\n"`, `"\r\n"`)),
		0600,
	); writeError != nil {
		test.Fatal(writeError)
	}
	bad = onNode(test, javascriptMutant)
	if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
		test.Fatal("JavaScript namespace EOL mutant not caught only by Node bytes")
	}
	test.Log(
		"System.newLine namespace body agrees on both backends; native and JavaScript EOL mutants caught only by Node stdout comparison",
	)
}

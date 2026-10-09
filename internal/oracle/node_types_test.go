package oracle

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha512"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, name := range []string{"node_types_outside.a", "node_types_project.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path: "internal/oracle/testdata/" + name, lowers: true})
	}
}

// A separate test process supplies the real foreign compiler working directory
// without changing the directory beneath parallel oracle tests.
func TestEmbeddedNodeTypesPortable(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"outside", "project"} {
		t.Run(name, func(t *testing.T) {
			directory := t.TempDir()
			source, err := os.ReadFile(filepath.Join(root, "internal/oracle/testdata/node_types_"+name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(directory, "main.a"), source, 0600); err != nil {
				t.Fatal(err)
			}
			if name == "project" {
				installDifferentNodeTypes(t, root, directory)
			}
			command := bounded(t, os.Args[0], "-test.run=^TestEmbeddedNodeTypesPortableWorker$", "-test.v")
			command.Dir = directory
			command.Env = append(os.Environ(), "ADAMIC_NODE_PORTABLE_WORKER=1", "ADAMIC_NODE_PORTABLE_RUNNER="+filepath.Join(root, "oracle/node.mjs"))
			output, err := combinedChildOutput(command)
			if err != nil {
				t.Fatalf("foreign-directory compiler: %v\n%s", err, output)
			}
		})
	}
}

func installDifferentNodeTypes(t *testing.T, root, directory string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "internal/load/testdata/node_types/node-24.0.0.tgz"))
	if err != nil {
		t.Fatal(err)
	}
	digest := sha512.Sum512(data)
	if "sha512-"+base64.StdEncoding.EncodeToString(digest[:]) != "sha512-yZQa2zm87aRVcqDyH5+4Hv9KYgSdgwX1rFnGvpbzMaC7YAljmhBET93TPiTd3ObwTL+gSpIzPKg5BqVxdCvxKg==" {
		t.Fatal("different-version fixture package integrity mismatch")
	}
	compressed, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer compressed.Close()
	archive := tar.NewReader(compressed)
	for {
		entry, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if entry.Typeflag != tar.TypeReg {
			continue
		}
		_, name, ok := strings.Cut(entry.Name, "/")
		if !ok || strings.Contains(name, "..") {
			t.Fatal("invalid fixture package entry")
		}
		target := filepath.Join(directory, "node_modules/@types/node", filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			t.Fatal(err)
		}
		content, err := io.ReadAll(archive)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, content, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestEmbeddedNodeTypesPortableWorker(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_NODE_PORTABLE_WORKER") != "1" {
		t.Skip("executed by the foreign-directory portability test")
	}
	checked, err := load.Load([]string{"main.a"})
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range checked.CompilerProgram().GetSourceFiles() {
		if strings.Contains(file.FileName().AsString(), "/node_modules/@types/node/") && !load.IsNodeLibrary(file) {
			t.Fatalf("caller Node package was loaded: %s", file.FileName())
		}
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	how := inputRun{directory: directory}
	truth := executeInput(t, how, nil, "node", "--disable-warning=ExperimentalWarning", os.Getenv("ADAMIC_NODE_PORTABLE_RUNNER"), filepath.Join(directory, "main.a"))
	if truth.exitCode != 0 {
		t.Fatalf("Node: %q", truth.stderr)
	}
	binary := filepath.Join(t.TempDir(), "native")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	environment := []string{"ASAN_OPTIONS=detect_leaks=0"}
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=1"}
	}
	actual := executeInput(t, how, environment, binary)
	if difference := disagreement(truth, actual); difference != "" {
		t.Fatalf("native: %s", difference)
	}
	script := filepath.Join(t.TempDir(), "backend.mjs")
	if err := os.WriteFile(script, []byte(javascript.JavaScript(program)), 0600); err != nil {
		t.Fatal(err)
	}
	actual = executeInput(t, how, nil, "node", "--disable-warning=ExperimentalWarning", os.Getenv("ADAMIC_NODE_PORTABLE_RUNNER"), script)
	if difference := disagreement(truth, actual); difference != "" {
		t.Fatalf("JavaScript: %s; exit=%d stderr=%q", difference, actual.exitCode, actual.stderr)
	}
}

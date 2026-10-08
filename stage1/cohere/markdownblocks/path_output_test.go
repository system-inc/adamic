package markdownblocks

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// AstPath observations exceed V8's single-string limit. Keep the complete bytes
// on disk, including native and leak-run output, rather than retaining copies.
type pathOutput struct {
	run
	file string
}

func pathExecute(t *testing.T, environment []string, name string, arguments ...string) pathOutput {
	t.Helper()
	output, err := os.CreateTemp(t.TempDir(), "stdout-")
	if err != nil {
		t.Fatal(err)
	}
	command := bounded(t, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	var stderr bytes.Buffer
	command.Stdout, command.Stderr = output, &stderr
	err = command.Run()
	closeErr := output.Close()
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		t.Fatalf("running %s: %v", name, err)
	}
	return pathOutput{run: run{stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}, file: output.Name()}
}

func pathNode(t *testing.T, source string, arguments ...string) pathOutput {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	return pathExecute(t, nil, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, source}, arguments...)...)
}

func pathBackend(t *testing.T, program *ir.Program, arguments ...string) pathOutput {
	t.Helper()
	source := filepath.Join(t.TempDir(), "program.mjs")
	write(t, source, []byte(javascript.JavaScript(program)))
	return pathNode(t, source, arguments...)
}

func pathLeaks(t *testing.T, program *ir.Program, sanitized string, arguments ...string) {
	t.Helper()
	switch runtime.GOOS {
	case "linux":
		report := pathExecute(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, sanitized, arguments...)
		if report.exitCode != 0 {
			t.Fatalf("exit %d\n%s", report.exitCode, report.stderr)
		}
		removePathOutput(t, report.file)
	case "darwin":
		binary := filepath.Join(t.TempDir(), "port")
		if err := native.Build(native.C(program), binary, native.Options{}); err != nil {
			t.Fatal(err)
		}
		report := pathExecute(t, nil, "leaks", append([]string{"--atExit", "--", binary}, arguments...)...)
		if report.exitCode != 0 {
			data, err := os.ReadFile(report.file)
			if err != nil {
				t.Fatal(err)
			}
			t.Fatal(string(data))
		}
		removePathOutput(t, report.file)
	default:
		t.Fatalf("no leak check for %s", runtime.GOOS)
	}
}

// Return the absolute differing byte and its context index without constructing
// an output-sized string. Length differences, including final newlines, count.
func pathDifference(t *testing.T, got, want string) (int64, int, bool) {
	t.Helper()
	a, err := os.Open(got)
	if err != nil {
		t.Fatal(err)
	}
	defer a.Close()
	b, err := os.Open(want)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	ab, bb := make([]byte, 64*1024), make([]byte, 64*1024)
	var offset int64
	context := 0
	for {
		an, ae := io.ReadFull(a, ab)
		bn, be := io.ReadFull(b, bb)
		for _, e := range []error{ae, be} {
			if e != nil && e != io.EOF && e != io.ErrUnexpectedEOF {
				t.Fatal(e)
			}
		}
		for i := 0; i < min(an, bn); i++ {
			if ab[i] != bb[i] {
				return offset + int64(i), context, false
			}
			if bb[i] == '\n' {
				context++
			}
		}
		if an != bn {
			return offset + int64(min(an, bn)), context, false
		}
		offset += int64(an)
		if ae != nil || be != nil {
			return offset, context, true
		}
	}
}

func pathEqual(t *testing.T, name, got, want string, inputs []auditInput) {
	t.Helper()
	offset, context, equal := pathDifference(t, got, want)
	if equal {
		return
	}
	label := "end of contexts"
	if context < len(inputs) {
		label = inputs[context].Name
	}
	t.Fatalf("%s %s first byte difference at %d (lengths %d/%d)", name, label, offset, pathOutputSize(t, got), pathOutputSize(t, want))
}

func pathOutputSize(t *testing.T, file string) int64 {
	t.Helper()
	info, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	return info.Size()
}

func removePathOutput(t *testing.T, file string) {
	t.Helper()
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
}

func copyPathOutput(t *testing.T, from, to string) {
	t.Helper()
	input, err := os.Open(from)
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	output, err := os.Create(to)
	if err != nil {
		t.Fatal(err)
	}
	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		t.Fatal(copyErr)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
}

func TestAstPathOutputComparison(t *testing.T) {
	prefix := bytes.Repeat([]byte("x"), 65535)
	want := append(append(append([]byte{}, prefix...), '\n'), 'y', '\n')
	dir := t.TempDir()
	expected := filepath.Join(dir, "want")
	write(t, expected, want)
	for _, check := range []struct {
		name    string
		data    []byte
		offset  int64
		context int
		equal   bool
	}{
		{"identical across buffer boundary", want, int64(len(want)), 2, true},
		{"changed first byte", append([]byte("z"), want[1:]...), 0, 0, false},
		{"changed second context", append(append([]byte{}, want[:65536]...), 'z', '\n'), 65536, 1, false},
		{"missing final newline", want[:len(want)-1], int64(len(want) - 1), 1, false},
		{"extra newline", append(append([]byte{}, want...), '\n'), int64(len(want)), 2, false},
		{"empty output", nil, 0, 0, false},
	} {
		t.Run(check.name, func(t *testing.T) {
			got := filepath.Join(dir, fmt.Sprintf("got-%d", len(check.data)))
			write(t, got, check.data)
			offset, context, equal := pathDifference(t, got, expected)
			if offset != check.offset || context != check.context || equal != check.equal {
				t.Fatalf("difference = %d, context %d, equal %v", offset, context, equal)
			}
		})
	}
}

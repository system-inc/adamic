package oracle

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Each changes the real raw writer in a private runtime archive. Compiling is
// required to succeed; only the corresponding fixture comparison kills it.
func stdoutRuntimeMutant(test *testing.T, fixture, before, after, catcher string, closed bool) {
	test.Helper()
	path, _ := filepath.Abs("testdata/host_stdout/" + fixture + ".a")
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	source := native.C(program)
	directory := test.TempDir()
	entries, readDirectoryError := os.ReadDir("../native/runtime")
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, readError := os.ReadFile(filepath.Join("../native/runtime", entry.Name()))
		if readError != nil {
			test.Fatal(readError)
		}
		if entry.Name() == "adamic.c" {
			if strings.Count(string(data), before) != 1 {
				test.Fatal("mutant must match exactly once")
			}
			data = []byte(strings.Replace(string(data), before, after, 1))
		}
		if writeError := os.WriteFile(filepath.Join(directory, entry.Name()), data, 0600); writeError != nil {
			test.Fatal(writeError)
		}
	}
	options := native.Options{Sanitize: true}
	archive, runtimeLibraryForSourceError := native.RuntimeLibraryForSource(directory, source, options)
	if runtimeLibraryForSourceError != nil {
		test.Fatal(runtimeLibraryForSourceError)
	}
	binary := filepath.Join(test.TempDir(), "mutant")
	if buildWithRuntimeError := native.BuildWithRuntime(
		source,
		binary,
		options,
		archive,
	); buildWithRuntimeError != nil {
		test.Fatal(buildWithRuntimeError)
	}
	if closed {
		reader, writer, pipeError := os.Pipe()
		if pipeError != nil {
			test.Fatal(pipeError)
		}
		reader.Close()
		stderr, createError := os.Create(filepath.Join(test.TempDir(), "stderr"))
		if createError != nil {
			test.Fatal(createError)
		}
		code := withStreams(test, writer, stderr, binary)
		writer.Close()
		stderr.Close()
		said, readError := os.ReadFile(stderr.Name())
		if readError != nil {
			test.Fatal(readError)
		}
		if code != 0 || len(said) != 0 {
			test.Fatalf("mutant must finish cleanly with wrong exit0, got %d %q", code, said)
		}
		test.Logf("%s goes red: closed stdout exited %d instead of70", catcher, code)
	} else {
		truth := onNode(test, path)
		got := execute(test, binary)
		if truth.exitCode != 0 || got.exitCode != 0 || len(got.stderr) != 0 {
			test.Fatalf("must finish sanitizer-clean, got %+v", got)
		}
		if bytes.Equal(truth.stdout, got.stdout) {
			test.Fatal("mutant survived " + catcher)
		}
		test.Logf("%s goes red: %s", catcher, disagreement(truth, got))
	}
}

func TestStdoutWriteMutantBypassesOrdering(test *testing.T) {
	test.Parallel()
	stdoutRuntimeMutant(test, "order",
		"bool adamic_write_raw(enum adamic_stream stream, const adamic_string *text) {\n\tpthread_mutex_lock(&output_lock);\n\toutput_start();\n\tflush();\n\twrite_text(stream, text->bytes, text->length);\n\tflush();\n\tbool ok = !broken[stream];\n\tpthread_mutex_unlock(&output_lock);\n\treturn ok;\n}",
		"bool adamic_write_raw(enum adamic_stream stream, const adamic_string *text) {\n\treturn write_all(stream, text->bytes, text->length) == 0;\n}",
		"TestStdoutWriteOrderingMatchesNode", false)
}
func TestStdoutWriteMutantPassesLoneSurrogates(test *testing.T) {
	test.Parallel()
	stdoutRuntimeMutant(
		test,
		"text",
		"write_text(stream, text->bytes, text->length);",
		"put(stream, text->bytes, text->length);",
		"TestStdoutWriteTextMatchesNode",
		false,
	)
}
func TestStdoutWriteMutantOmitsExitHook(test *testing.T) {
	test.Parallel()
	stdoutRuntimeMutant(test, "closed",
		"bool adamic_write_raw(enum adamic_stream stream, const adamic_string *text) {\n\tpthread_mutex_lock(&output_lock);\n\toutput_start();",
		"bool adamic_write_raw(enum adamic_stream stream, const adamic_string *text) {\n\tpthread_mutex_lock(&output_lock);",
		"TestStdoutWriteClosedPipeMatchesNode", true)
}

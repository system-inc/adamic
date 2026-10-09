package dispatch

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// The formatter's identity: what the format record is keyed by, so it survives a cohere commit that
// leaves the formatter alone.
//
// The record holds the hashes of texts the formatter returned unchanged, and that claim is only as old
// as the printers that made it. Keyed by the whole binary, every cohere commit threw it away, lint-only
// ones included, and the next `s c` reformatted every file. Keyed by this, it is thrown away exactly
// when something that can change what the formatter prints has changed:
//
//   - every Go source the native printers compile from, transitively, which includes the typescript-go
//     parser they parse with, so a compiler pin moving counts;
//   - the pass loop in command/cohere/format.go, which decides what "formatted" means;
//   - the toolchain.
//
// Each file is named by its import path and its own name rather than by where it sits on disk, so a
// committed snapshot, whose compiler is linked in from the extraction cache, and a working tree name
// the same file the same way.

// formatterPackage is the package whose transitive sources decide what the formatter prints.
const formatterPackage = "./internal/format/native"

// FormatterPassLoop is the one file outside it that decides what the formatter's output means.
const FormatterPassLoop = "command/cohere/format.go"

// FormatterIdentity hashes the formatter's inputs in moduleDirectory. environment is added to the `go
// list` that finds them, which a snapshot needs to point GOWORK at its own workspace.
//
// A module with no formatter package answers empty, and the binary is built unstamped: its format record
// is then keyed by the binary itself, which is always sound and only slower. The package moving is
// caught by TestFormatterInputsCoverWhatCanChangeTheOutput, which needs the real module's list to be
// whole.
func FormatterIdentity(moduleDirectory string, goVersion string, environment []string) (string, error) {
	if _, err := os.Stat(filepath.Join(moduleDirectory, filepath.FromSlash(formatterPackage))); errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	inputs, err := FormatterInputs(moduleDirectory, environment)
	if err != nil {
		return "", err
	}
	return hashFormatterInputs(goVersion, inputs), nil
}

// FormatterInput is one file the formatter is built from: its name by import path, and its content's
// SHA-256.
type FormatterInput struct {
	Name string
	Sum  string
}

// FormatterInputs lists every file the formatter is built from.
func FormatterInputs(moduleDirectory string, environment []string) ([]FormatterInput, error) {
	const template = `{{if not .Standard}}{{$package := .ImportPath}}{{$directory := .Dir}}` +
		`{{range .GoFiles}}{{$package}}{{"\t"}}{{$directory}}{{"\t"}}{{.}}{{"\n"}}{{end}}` +
		`{{range .EmbedFiles}}{{$package}}{{"\t"}}{{$directory}}{{"\t"}}{{.}}{{"\n"}}{{end}}` +
		`{{end}}`
	command := exec.Command("go", "list", "-deps", "-f", template, formatterPackage)
	command.Dir = moduleDirectory
	command.Env = append(os.Environ(), environment...)
	output, err := command.Output()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return nil, fmt.Errorf("listing the formatter's inputs: %w\n%s", err, exitError.Stderr)
		}
		return nil, fmt.Errorf("listing the formatter's inputs: %w", err)
	}

	inputs := []FormatterInput{}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) != 3 {
			continue
		}
		sum, err := fileSum(filepath.Join(fields[1], fields[2]))
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, FormatterInput{Name: fields[0] + "/" + filepath.ToSlash(fields[2]), Sum: sum})
	}
	// The printers alone are dozens of files, so a short list is a `go list` that answered about
	// something else, and an identity hashed from it would match a formatter it does not describe.
	if len(inputs) < 10 {
		return nil, fmt.Errorf("go list found %d formatter inputs in %s, which cannot be right", len(inputs), moduleDirectory)
	}

	sum, err := fileSum(filepath.Join(moduleDirectory, FormatterPassLoop))
	if err != nil {
		return nil, err
	}
	return append(inputs, FormatterInput{Name: FormatterPassLoop, Sum: sum}), nil
}

// hashFormatterInputs is the identity of a set of inputs under a toolchain, whatever order they came in.
func hashFormatterInputs(goVersion string, inputs []FormatterInput) string {
	lines := make([]string, 0, len(inputs))
	for _, input := range inputs {
		lines = append(lines, input.Name+"\x00"+input.Sum)
	}
	sort.Strings(lines)
	digest := sha256.New()
	fmt.Fprintf(digest, "formatter\x00%s\x00", goVersion)
	for _, line := range lines {
		fmt.Fprintf(digest, "%s\x00", line)
	}
	return hex.EncodeToString(digest.Sum(nil))[:hashLength]
}

func fileSum(path string) (string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading formatter input %s: %w", path, err)
	}
	sum := sha256.Sum256(contents)
	return hex.EncodeToString(sum[:]), nil
}

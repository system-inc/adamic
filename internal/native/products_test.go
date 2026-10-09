package native

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	buildcache "github.com/system-inc/adamic/internal/native/testbuildcache"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

var testExecutable struct {
	sync.Once
	digest string
	err    error
}

type testBuildCache struct{ inputs buildcache.Inputs }

func newTestBuildCache(repository string, directories ...string) (*testBuildCache, error) {
	testExecutable.Do(func() {
		path, err := os.Executable()
		if err != nil {
			testExecutable.err = err
			return
		}
		file, err := os.Open(path)
		if err != nil {
			testExecutable.err = err
			return
		}
		defer file.Close()
		h := sha256.New()
		_, testExecutable.err = io.Copy(h, file)
		testExecutable.digest = fmt.Sprintf("%x", h.Sum(nil))
	})
	if testExecutable.err != nil {
		return nil, testExecutable.err
	}
	in := buildcache.Inputs{Name: "native tests", Files: append([]string{"go.mod", "cmd/adamic"}, directories...), Flags: []string{"test=" + testExecutable.digest}, Toolchain: []string{buildcache.Tool("go", "version"), buildcache.Tool("clang", "--version"), buildcache.Tool("node", "--version")}}
	for _, name := range []string{"GOFLAGS", "CGO_CFLAGS", "CGO_LDFLAGS", "CC", "GOTOOLCHAIN"} {
		in.Flags = append(in.Flags, name+"="+os.Getenv(name))
	}
	return &testBuildCache{in}, nil
}
func (c *testBuildCache) Tree(label string, inputs [][]byte, build func(string) error) (string, error) {
	in := c.inputs
	in.Name = label
	in.Flags = append([]string(nil), c.inputs.Flags...)
	for index, input := range inputs {
		in.Flags = append(in.Flags, fmt.Sprintf("input[%d]=%x", index, sha256.Sum256(input)))
	}
	return buildcache.Get(in, build)
}

// Command caches compiler output only. It refuses commands without a single -o.
// Explicit input files and Go overlays contribute their bytes, not temporary paths.
func (c *testBuildCache) Command(command *exec.Cmd, scratch string, extraInputs ...[]byte) ([]byte, error) {
	outputIndex := -1
	for i, arg := range command.Args {
		if arg == "-o" && i+1 < len(command.Args) {
			if outputIndex != -1 {
				return nil, fmt.Errorf("multiple outputs")
			}
			outputIndex = i + 1
		}
	}
	if outputIndex < 0 {
		return nil, fmt.Errorf("build command has no output")
	}
	output := command.Args[outputIndex]
	inputs := append([][]byte(nil), extraInputs...)
	for i, arg := range command.Args {
		if i == outputIndex {
			continue
		}
		normalized := strings.ReplaceAll(arg, scratch, "$SCRATCH")
		inputs = append(inputs, []byte(normalized))
		if data, err := os.ReadFile(arg); err == nil {
			if strings.HasSuffix(arg, ".json") {
				var overlay struct{ Replace map[string]string }
				if json.Unmarshal(data, &overlay) == nil && len(overlay.Replace) > 0 {

					data = nil
					originals := make([]string, 0, len(overlay.Replace))
					for original := range overlay.Replace {
						originals = append(originals, original)
					}
					sort.Strings(originals)
					for _, original := range originals {
						contents, err := os.ReadFile(overlay.Replace[original])
						if err != nil {
							return nil, err
						}
						inputs = append(inputs, []byte(original), contents)
					}
				}
			}
			inputs = append(inputs, data)
		}
	}
	// Command.Env only differs for the sanitized archive; do not persist environment values.
	for _, variable := range command.Env {
		for _, name := range []string{"CC=", "CGO_CFLAGS=", "CGO_LDFLAGS="} {
			if strings.HasPrefix(variable, name) {
				inputs = append(inputs, []byte(variable))
			}
		}
	}
	name := filepath.Base(output)
	entry, err := c.Tree(name, inputs, func(directory string) error {
		command.Args[outputIndex] = filepath.Join(directory, name)
		defer func() { command.Args[outputIndex] = output }()
		data, err := command.CombinedOutput()
		os.WriteFile(filepath.Join(directory, "build.log"), data, 0644)
		if err != nil {
			return fmt.Errorf("%s: %w\n%s", command.Args, err, data)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(entry, name))
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(output, data, 0755); err != nil {
		return nil, err
	}
	return os.ReadFile(filepath.Join(entry, "build.log"))
}

package typeaware

import (
	"flag"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

type volumeMutation struct{ name, path, from, to string }

func volumeMutations() []volumeMutation {
	return []volumeMutation{
		{"assignable-types", "bridge/tsgo/checker/facts.go", "out.yes(checker.Checker_isTypeAssignableTo(c, selected[0], selected[1]))", "out.yes(checker.Checker_isTypeAssignableTo(c, selected[1], selected[0]))"},
		{"widened-shape", "bridge/tsgo/checker/facts.go", "t = checker.Checker_getWidenedType(c, t)", "// Mutant keeps the fresh type."},
		{"enum-types", "bridge/tsgo/checker/facts.go", "base = c.GetTypeAtLocation(symbol.ValueDeclaration.Parent)", "base = part"},
		{"type-symbol", "bridge/tsgo/checker/facts.go", "name = symbol.Name", "name = symbol.Name + \"wrong\""},
		{"scope-locals", "bridge/tsgo/checker/scopes.go", "out.text(name)", "out.text(name + \"wrong\")"},
		{"call-returns", "bridge/tsgo/checker/facts.go", "roots = append(roots, g.add(c.GetReturnTypeOfSignature(signature)))", "_ = signature; roots = append(roots, g.add(checker.Checker_numberType(c)))"},
		{"property-shape", "bridge/tsgo/checker/facts.go", "g.add(c.GetTypeOfSymbolAtLocation(property, node))", "g.add(checker.Checker_numberType(c))"},
		{"contextual-shape", "bridge/tsgo/checker/facts.go", "t := checker.Checker_getContextualType(c, node, checker.ContextFlagsNone)", "t := c.GetTypeAtLocation(node)"},
		{"symbol-origin", "bridge/tsgo/checker/facts.go", "file = f.FileName()", "file = source.FileName()"},
		{"type-origin", "bridge/tsgo/checker/metadata.go", "out.text(symbol.Name)", "out.text(symbol.Name + \"wrong\")"},
		{"property-info", "bridge/tsgo/checker/metadata.go", "out.text(strings.TrimPrefix(declaration.Kind.String(), \"Kind\"))", "out.text(\"PropertySignature\")"},
		{"call-count", "bridge/tsgo/checker/facts.go", "out.number(uint64(len(c.GetSignaturesOfType(subject, checker.SignatureKindCall))))", "out.number(0)"},
		{"call-parameters", "bridge/tsgo/checker/facts.go", "g.add(checker.Checker_getApparentType(c, c.GetTypeOfSymbolAtLocation(params[0], node)))", "g.add(checker.Checker_numberType(c))"},
		{"apparent-shape", "bridge/tsgo/checker/facts.go", "g.add(checker.Checker_getApparentType(c, subject))", "g.add(checker.Checker_numberType(c))"},
		{"base-shapes", "bridge/tsgo/checker/facts.go", "roots = append(roots, g.add(base))", "_ = base"},
	}
}

const volumeReleaseSource = `import { programArguments, tsgoProgram, tsgoInspect, tsgoRelease } from 'adamic';
const args=programArguments(); const path=args[1] ?? ''; const program=tsgoProgram(args[0] ?? '',[path]);tsgoRelease(program);
console.log(tsgoInspect(program,path,0,1,'Identifier','call-returns'));
`

type volumeInput struct{ path, hash string }

var volumeInputs map[string]volumeInput

func volumePreparedInput(t *testing.T, name string) string {
	t.Helper()
	input, ok := volumeInputs[name]
	if !ok {
		t.Fatalf("missing prepared build input %s", name)
	}
	actual, err := volumeFileHash(input.path)
	if err != nil {
		t.Fatal(err)
	}
	if actual != input.hash {
		t.Fatalf("build input %s changed: got %s, want %s", name, actual, input.hash)
	}
	return input.path
}

// Product preparation precedes m.Run: these are build inputs, not test units.
// CPU accounting for the complete go test command still includes all preparation.
// Listing and unrelated focused tests never prepare the volume products.
func TestMain(m *testing.M) {
	flag.Parse()
	pattern := strings.Split(flag.Lookup("test.run").Value.String(), "/")[0]
	selected, err := regexp.MatchString(pattern, "TestVolumeAgreementAndMutants")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var directory string
	if selected && flag.Lookup("test.list").Value.String() == "" {
		directory, err = os.MkdirTemp("", "volume-build-inputs-")
		if err == nil {
			err = volumePrepareInputs(directory)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "volume build inputs: %v\n", err)
			os.RemoveAll(directory)
			os.Exit(1)
		}
	}
	code := m.Run()
	if directory != "" {
		os.RemoveAll(directory)
	}
	os.Exit(code)
}

// GoInputs/Get is the TestMain counterpart of GoBuild (which needs testing.TB).
func volumeGoInput(repository, output, pkg string, args []string, environment ...string) (string, error) {
	inputs, err := buildcache.GoInputs(output, pkg, args, environment)
	if err != nil {
		return "", err
	}
	directory, err := buildcache.Get(inputs, func(directory string) error {
		command := exec.Command("go", append(append(append([]string{"build"}, args...), "-o", filepath.Join(directory, output)), pkg)...)
		command.Env = append(os.Environ(), environment...)
		started := time.Now()
		err := volumeCommand(repository, command)
		fmt.Printf("volume-build %s cold=%.6fs\n", output, time.Since(started).Seconds())
		return err
	})
	return filepath.Join(directory, output), err
}

func volumePrivateInput(repository, directory, output string, inputs buildcache.Inputs, build func(string) error) (string, error) {
	key, err := buildcache.Key(repository, inputs)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(directory, 0700); err != nil {
		return "", err
	}
	if err := build(directory); err != nil {
		return "", err
	}
	path := filepath.Join(directory, output)
	hash, err := volumeFileHash(path)
	if err != nil {
		return "", err
	}
	if err := os.Chmod(path, 0444); err != nil {
		return "", err
	}
	if output == "native" {
		if err := os.Chmod(path, 0555); err != nil {
			return "", err
		}
	}
	fmt.Printf("volume-private-input %s key=%s product=%s\n", inputs.Name, key, hash)
	return path, nil
}

func volumePrepareInputs(directory string) error {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		return err
	}
	stage0, err := volumeGoInput(repository, "adamic", "./cmd/adamic", nil)
	if err != nil {
		return err
	}
	normal, err := volumeGoInput(repository, "tsgo.a", "./bridge/tsgo/archive", []string{"-buildmode=c-archive"})
	if err != nil {
		return err
	}
	entry := filepath.Join(repository, "stage1/cohere/typeaware/volume_suite.ts")
	for _, item := range []struct {
		name, archive string
		sanitize      bool
	}{{"volume", normal, false}, {"volume-asan", "", true}} {
		archive := item.archive
		if item.sanitize {
			archive, err = volumeGoInput(repository, "tsgo-asan.a", "./bridge/tsgo/archive", []string{"-buildmode=c-archive"}, "CC=clang", "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all")
			if err != nil {
				return err
			}
		}
		inputs, build, err := volumeNativeSpec(repository, stage0, item.name, entry, archive, item.sanitize)
		if err != nil {
			return err
		}
		if _, err := buildcache.Get(inputs, build); err != nil {
			return err
		}
	}

	workspace, err := volumeOracleWorkspace(repository)
	if err != nil {
		return err
	}
	if _, err := volumeGoInput(repository, "volume-oracle", "./stage1/cohere/typeaware/testdata/volume-oracle-go", nil, "GOWORK="+workspace); err != nil {
		return err
	}

	released := filepath.Join(directory, "released-inspect.ts")
	if err := os.WriteFile(released, []byte(volumeReleaseSource), 0600); err != nil {
		return err
	}
	inputs, build, err := volumeNativeSpec(repository, stage0, "released-inspect", released, normal, false)
	if err != nil {
		return err
	}
	if _, err := buildcache.Get(inputs, build); err != nil {
		return err
	}
	volumeInputs = make(map[string]volumeInput)
	mutations := append(volumeMutations(), volumeMutation{"released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant keeps released program live."})
	for _, change := range mutations {
		// Omit unselected private products, while the test still checks the full union.
		index := 1
		for i, candidate := range mutations {
			if candidate.name == change.name {
				index = i + 1
				break
			}
		}
		selected, err := volumeSelected(os.Getenv("ADAMIC_TEST_SHARD"), index)
		if err != nil {
			return err
		}
		if !selected {
			continue
		}
		original := filepath.Join(repository, change.path)
		data, err := os.ReadFile(original)
		if err != nil {
			return err
		}
		if strings.Count(string(data), change.from) != 1 {
			return fmt.Errorf("nonunique mutant %s", change.name)
		}
		side := filepath.Join(directory, change.name+".go")
		if err := os.WriteFile(side, []byte(strings.Replace(string(data), change.from, change.to, 1)), 0600); err != nil {
			return err
		}
		overlay := filepath.Join(directory, change.name+".json")
		encoded := fmt.Sprintf("{\"Replace\":{%q:%q}}", original, side)
		if err := os.WriteFile(overlay, []byte(encoded), 0600); err != nil {
			return err
		}
		archiveName := change.name + "-checker"
		archiveInputs, err := buildcache.GoInputs(archiveName, "./bridge/tsgo/archive", []string{"-buildmode=c-archive"}, nil)
		if err != nil {
			return err
		}
		replacementHash, err := volumeFileHash(side)
		if err != nil {
			return err
		}
		archiveInputs.Flags = append(archiveInputs.Flags, change.path, change.from, change.to, "-overlay="+overlay, "replacement-path="+side, "replacement-sha256="+replacementHash)
		archive, err := volumePrivateInput(repository, filepath.Join(directory, archiveName), "checker.a", archiveInputs, func(directory string) error {
			started := time.Now()
			err := volumeCommand(repository, exec.Command("go", "build", "-buildmode=c-archive", "-overlay", overlay, "-o", filepath.Join(directory, "checker.a"), "./bridge/tsgo/archive"))
			fmt.Printf("volume-build %s private cold=%.6fs\n", archiveName, time.Since(started).Seconds())
			return err
		})
		if err != nil {
			return err
		}
		nativeName := change.name + "-native"
		source := entry
		if change.name == "released-registry" {
			source = released
		}
		nativeInputs, nativeBuild, err := volumeNativeSpec(repository, stage0, nativeName, source, archive, false)
		if err != nil {
			return err
		}
		native, err := volumePrivateInput(repository, filepath.Join(directory, nativeName), "native", nativeInputs, nativeBuild)
		if err != nil {
			return err
		}
		for name, path := range map[string]string{archiveName: archive, nativeName: native} {
			hash, err := volumeFileHash(path)
			if err != nil {
				return err
			}
			volumeInputs[name] = volumeInput{path, hash}
		}
	}
	return nil
}

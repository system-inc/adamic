package generate

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Framework restricts exports to public headers below Headers. Dependencies
// that should also be bound must be listed explicitly.
type Framework struct{ Name, Headers string }
type Configuration struct {
	Frameworks []Framework
	Platform   string // macos, ios, tvos, watchos, visionos
	Umbrella   string // check-source include; not copied into output as an absolute path
	ReadSource func(string) ([]byte, error)
}
type File struct {
	Path    string
	Content []byte
}
type Output struct {
	Files []File
	Check []byte
	// Leaves is the cycle finder's leaf table, with its derivation (leaves.go), and Holds what every
	// other class may hold, by the same derivation.
	Leaves []byte
	Holds  []byte
}

// Generate is the loader-facing library entry. The caller owns clang's stream
// and checks its process exit status before publishing Output.
func Generate(input io.Reader, configuration Configuration) (Output, error) {
	if configuration.ReadSource == nil {
		configuration.ReadSource = os.ReadFile
	}
	if configuration.Platform == "" {
		configuration.Platform = "macos"
	}
	switch configuration.Platform {
	case "macos", "ios", "tvos", "watchos", "visionos":
	default:
		return Output{}, fmt.Errorf("unsupported platform %q", configuration.Platform)
	}
	if len(configuration.Frameworks) == 0 {
		return Output{}, fmt.Errorf("no frameworks")
	}
	configuration.Frameworks = append([]Framework{}, configuration.Frameworks...)
	for i, f := range configuration.Frameworks {
		if !safeName(f.Name) || f.Headers == "" {
			return Output{}, fmt.Errorf("invalid framework %q", f.Name)
		}
		absolute, err := filepath.Abs(f.Headers)
		if err != nil {
			return Output{}, err
		}
		configuration.Frameworks[i].Headers = filepath.Clean(absolute)
	}
	g := &generator{configuration: configuration, types: map[string]*definition{}, sources: map[string][]byte{}, modules: map[string]*module{}, bound: map[string]bool{}}
	err := readAST(input, func(n *node) error {
		if !n.Location.Valid {
			return nil
		}
		file, err := filepath.Abs(n.Location.File)
		if err != nil {
			return err
		}
		for _, f := range configuration.Frameworks {
			relative, err := filepath.Rel(f.Headers, file)
			home := ""
			for moved, framework := range movedHeaders {
				if strings.HasSuffix(filepath.ToSlash(file), "/"+moved) {
					home = framework
				}
			}
			if home == f.Name || home == "" && err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && strings.HasSuffix(strings.ToLower(file), ".h") {
				n.Framework = f.Name
				g.nodes = append(g.nodes, n)
				break
			}
		}
		return nil
	})
	if err != nil {
		return Output{}, fmt.Errorf("clang JSON: %w", err)
	}
	return g.build()
}

// movedHeaders are headers Apple moved below the framework their declarations belong to: CGRect,
// CGPoint and CGSize are defined in CoreFoundation's CFCGTypes.h, while Swift and every caller
// find them in CoreGraphics, so they're generated there when CoreGraphics is.
//
// NSObject is declared by the Objective-C runtime's objc/NSObject.h; the rename table already names
// it FoundationObject, so it's generated with Foundation, and any object (id) binds to it.
var movedHeaders = map[string]string{"CoreFoundation.framework/Headers/CFCGTypes.h": "CoreGraphics", "usr/include/objc/NSObject.h": "Foundation"}

func safeName(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if !(c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_') {
			return false
		}
	}
	return true
}

// RunClang invokes clang with a pipe, so the SDK AST is never written or read as
// one byte slice. Failed clang invocations never produce a successful binding.
func RunClang(ctx context.Context, clang string, arguments []string, configuration Configuration) (Output, error) {
	command := exec.CommandContext(ctx, clang, arguments...)
	stream, err := command.StdoutPipe()
	if err != nil {
		return Output{}, err
	}
	var diagnostics strings.Builder
	command.Stderr = &diagnostics
	if err := command.Start(); err != nil {
		return Output{}, err
	}
	output, generateError := Generate(stream, configuration)
	if generateError != nil {
		_ = command.Process.Kill()
	}
	waitError := command.Wait()
	if generateError != nil {
		if waitError != nil {
			return Output{}, fmt.Errorf("generation: %w\nclang: %v\n%s", generateError, waitError, diagnostics.String())
		}
		return Output{}, generateError
	}
	if waitError != nil {
		return Output{}, fmt.Errorf("clang: %w\n%s", waitError, diagnostics.String())
	}
	return output, nil
}

// LeavesFile is where Write puts the leaf table, beside the bindings, for the cycle finder.
const LeavesFile = "leaves.txt"

// HoldsFile is where Write puts what every class that isn't a leaf may hold.
const HoldsFile = "holds.txt"

// Write writes paths in canonical order. Generation, validation and clang's
// successful exit have completed before this function can be called.
func (o Output) Write(directory string) error {
	files := append([]File{}, o.Files...)
	files = append(files, File{Path: "bindings-check.m", Content: o.Check}, File{Path: LeavesFile, Content: o.Leaves}, File{Path: HoldsFile, Content: o.Holds})
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	for _, file := range files {
		clean := filepath.Clean(filepath.FromSlash(file.Path))
		if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return fmt.Errorf("invalid output path %q", file.Path)
		}
		name := filepath.Join(directory, clean)
		if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(name, file.Content, 0644); err != nil {
			return err
		}
	}
	return nil
}

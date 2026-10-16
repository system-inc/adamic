// Package goplan runs a go product's build by a plan made before the build's window opens (buildcache's goProduct): the
// go command resolved, its arguments, directory and environment, the overlay's map, each outside header's bytes and
// each mutated file's text. Run, inside the window, writes only what the plan holds and runs go. The package holds no
// init and no package variable, so a lean build's census of Run (buildcache's
// TestAGoProductsBuildReadsNoStateButItsInitializers) reads its whole reach here, and a variable added later is a
// finding (#earlycut lever 3, Node's ruling (B) on b20037f75b).
package goplan

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
)

// A Command is the go a plan runs: Program, in Directory with Environment, as Arguments (no -overlay among them), -o
// Output inside the product, then Packages. Verb and Package are the request's, for the build's error.
type Command struct {
	Program, Directory, Output, Verb, Package string
	Arguments, Packages, Environment          []string
}

// A Header is one the build lays into a package: Target, the path in the package it replaces, and Content, its bytes.
type Header struct {
	Target  string
	Content []byte
}

// A Mutation is one mutated file: File, its path in the repository, and Text, every mutation of it applied. Original is
// the file's absolute path, and Pointers the overlay's entries whose replacement is it: Run points each at the copy, or
// replaces Original by the copy when none is.
type Mutation struct {
	File, Text, Original string
	Pointers             []string
}

// A Plan is one go product's build, made whole by New, or by Failed when it couldn't be made.
type Plan struct {
	// failed is why the plan couldn't be made, which fails the build, never the lookup of a product already built. A
	// string, since an error the build is handed would reach every error type of the module (leancensus).
	failed  string
	command Command
	// replace is the request's own overlay's map, as it declares it.
	replace map[string]string
	headers []Header
	mutated []Mutation
}

// New is the plan that runs command over the request's overlay map replace, with headers laid into their packages and
// mutated files laid over what they mutate.
func New(command Command, replace map[string]string, headers []Header, mutated []Mutation) Plan {
	return Plan{command: command, replace: replace, headers: headers, mutated: mutated}
}

// Failed is the plan that couldn't be made, for why: its build fails with it.
func Failed(why string) Plan {
	return Plan{failed: why}
}

// Run builds the plan's product into directory. Each mutated file's copy, each header and the overlay itself go into a
// temporary directory of its own, since the overlay names this checkout's paths, and it leaves with the build
// (#tqrqx60); then go runs, its output the product's only file.
func (plan Plan) Run(directory string) error {
	if plan.failed != "" {
		return errors.New(plan.failed)
	}
	arguments := slices.Clone(plan.command.Arguments)
	if len(plan.replace) > 0 || len(plan.headers) > 0 || len(plan.mutated) > 0 {
		scratch, err := os.MkdirTemp("", "buildcache-overlay-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(scratch)
		replace := map[string]string{}
		maps.Copy(replace, plan.replace)
		for _, mutation := range plan.mutated {
			copied := filepath.Join(scratch, "mutations", filepath.FromSlash(mutation.File))
			if err := os.MkdirAll(filepath.Dir(copied), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(copied, []byte(mutation.Text), 0o644); err != nil {
				return err
			}
			for _, pointer := range mutation.Pointers {
				replace[pointer] = copied
			}
			if len(mutation.Pointers) == 0 {
				replace[mutation.Original] = copied
			}
		}
		for index, header := range plan.headers {
			path := filepath.Join(scratch, fmt.Sprintf("%d.h", index))
			if err := os.WriteFile(path, header.Content, 0o644); err != nil {
				return err
			}
			replace[header.Target] = path
		}
		encoded, err := json.Marshal(map[string]any{"Replace": replace})
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(scratch, "overlay.json"), encoded, 0o644); err != nil {
			return err
		}
		arguments = append(arguments, "-overlay="+filepath.Join(scratch, "overlay.json"))
	}
	command := exec.Command(plan.command.Program, append(append(arguments, "-o", filepath.Join(directory, plan.command.Output)), plan.command.Packages...)...)
	command.Dir, command.Env = plan.command.Directory, plan.command.Environment
	if combined, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("go %s %s: %v\n%s", plan.command.Verb, plan.command.Package, err, combined)
	}
	return nil
}

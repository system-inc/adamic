package wave28next

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func (h *harness) continueChecks(stage0, entry, archive, oracle, binary, config, manifest string, truth []byte) {
	owned := filepath.Dir(entry)
	imports := regexp.MustCompile(`from '([^']+)'`)
	absolute := func(source string) string {
		return imports.ReplaceAllStringFunc(source, func(match string) string {
			parts := imports.FindStringSubmatch(match)
			if !strings.HasPrefix(parts[1], ".") {
				return match
			}
			return "from '" + filepath.Clean(filepath.Join(owned, parts[1])) + "'"
		})
	}
	for _, m := range []struct{ name, file, from, to string }{
		{"exit", "no_process_exit_after_output.a", "if(walk.exits.includes(exit))", "if(true)"},
		{"timeout", "no_uncleared_race_timeout.a", "identity===symbol.id && !this.target(at)", "identity===symbol.id && this.target(at)"},
		{"blocking", "require_blocking_standard_streams.a", "this.ordered=program.canBlock.has(this.tree.rules.path);", "this.ordered=true;"},
	} {
		data, err := os.ReadFile(filepath.Join(owned, m.file))
		if err != nil {
			h.t.Fatal(err)
		}
		if strings.Count(string(data), m.from) != 1 {
			h.t.Fatalf("nonunique %s", m.name)
		}
		mutant := h.write("mutant-"+m.name+".a", absolute(strings.Replace(string(data), m.from, m.to, 1)))
		runner, err := os.ReadFile(entry)
		if err != nil {
			h.t.Fatal(err)
		}
		rewritten := strings.Replace(absolute(string(runner)), filepath.Join(owned, m.file), mutant, 1)
		executable := h.build(stage0, "mutant-"+m.name, h.write("runner-"+m.name+".a", rewritten), archive, false)
		got := h.must("mutant-"+m.name+"-run", exec.Command(executable, config, manifest))
		if len(got.stderr) != 0 || bytes.Equal(got.stdout, truth) {
			h.t.Fatalf("mutant %s not killed solely by comparison: %s", m.name, got.stderr)
		}
		h.t.Logf("mutant %s exited zero, empty stderr, comparison caught byte %d", m.name, firstDifference(got.stdout, truth))
	}
	asanArchive := h.archive("checker-asan", "", true)
	asan := h.build(stage0, "next-asan", entry, asanArchive, true)
	h.compare("controls-asan", oracle, asan, config, manifest)
	for _, corpus := range []struct{ name, config, manifest string }{
		{"repository", filepath.Join(h.repository, "tsconfig.json"), "/workspace/wave-28-artifacts/repository.manifest"},
		{"compiler", "/workspace/wave-28-corpus/src/compiler/tsconfig.json", "/workspace/wave-28-artifacts/compiler.manifest"},
	} {
		if _, err := os.Stat(corpus.manifest); err != nil {
			h.t.Logf("external frozen corpus %s not supplied: %v", corpus.name, err)
			continue
		}
		h.compare(corpus.name, oracle, binary, corpus.config, corpus.manifest)
		h.compare(corpus.name+"-asan", oracle, asan, corpus.config, corpus.manifest)
		for round := 0; round < 3; round++ {
			want := h.must(fmt.Sprintf("%s-timing-go-%d", corpus.name, round), exec.Command(oracle, corpus.config, corpus.manifest))
			got := h.must(fmt.Sprintf("%s-timing-native-%d", corpus.name, round), exec.Command(binary, corpus.config, corpus.manifest))
			if !bytes.Equal(want.stdout, got.stdout) || len(got.stderr) != 0 {
				h.t.Fatal("timing comparison failure")
			}
			h.t.Logf("%s round %d native %s Go %s", corpus.name, round, got.elapsed, want.elapsed)
		}
	}
}

func (h *harness) releasedChecks() {
	t := h.t
	directory := h.directory
	stage0 := filepath.Join(directory, "adamic")
	archive := filepath.Join(directory, "checker.a")
	config := filepath.Join(directory, "dom-tsconfig.json")
	released := h.write("released.a", `import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);
console.log(tsgoInspect(program,file,0,1,'Identifier','stream-symbol'));
`)
	probe := h.write("probe.a", "x;\n")
	for _, sanitize := range []bool{false, true} {
		name := "released"
		lib := archive
		if sanitize {
			name += "-asan"
			lib = filepath.Join(directory, "checker-asan.a")
		}
		executable := h.build(stage0, name, released, lib, sanitize)
		got := h.run(name+"-run", exec.Command(executable, config, probe))
		if code, ok := got.err.(*exec.ExitError); !ok || code.ExitCode() != 70 || string(got.stderr) != "adamic: panic: invalid or released checker handle\n" {
			t.Fatalf("released handle escaped: %v %s", got.err, got.stderr)
		}
		t.Logf("%s: panic 70, no sanitizer findings", name)
	}
	overlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant retains the released handle.")
	mutantArchive := h.archive("released-registry", overlay, false)
	mutant := h.build(stage0, "released-mutant", released, mutantArchive, false)
	h.must("released-mutant-run", exec.Command(mutant, config, probe))
	t.Log("released registry mutant exited zero; required panic assertion catches it")
}

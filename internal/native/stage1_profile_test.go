package native

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// Not parallel: compiler command auditing and diagnostic capture use process-wide state.
func TestStage1ProfileStalenessAndDeterminism(t *testing.T) {
	directory := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", filepath.Join(directory, "cache"))
	source := "// A\nint main(void) { return 0; }\n"
	generate := filepath.Join(directory, "generate")
	if err := Build(source, generate, Options{Release: true, ProfileGenerate: true}); err != nil {
		t.Fatal(err)
	}
	raw := filepath.Join(directory, "training.profraw")
	command := exec.Command(generate)
	command.Env = append(os.Environ(), "LLVM_PROFILE_FILE="+raw)
	if out, err := command.CombinedOutput(); err != nil || len(out) != 0 {
		t.Fatalf("training: %v %s", err, out)
	}
	compiler, err := exec.LookPath("clang")
	if err != nil {
		t.Fatal(err)
	}
	profdata, err := profileDataTool()
	if err != nil {
		t.Fatal(err)
	}
	text := filepath.Join(directory, "profile.txt")
	if out, err := exec.Command(profdata, "merge", "--text", raw, "-o", text).CombinedOutput(); err != nil {
		t.Fatalf("merge: %v %s", err, out)
	}
	data, err := os.ReadFile(text)
	if err != nil {
		t.Fatal(err)
	}
	record, err := NewStage1ProfileManifest(source, data, "test-only-corpus")
	if err != nil {
		t.Fatal(err)
	}
	writeManifest := func(r Stage1ProfileManifest) {
		t.Helper()
		data, err := json.Marshal(r)
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(directory, "manifest.json"), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	audit := filepath.Join(directory, "commands.jsonl")
	wrapperDir := filepath.Join(directory, "wrapper")
	if err = os.Mkdir(wrapperDir, 0o755); err != nil {
		t.Fatal(err)
	}
	encode := func(s string) string { b, _ := json.Marshal(s); return string(b) }
	wrapper := "#!/usr/bin/env python3\nimport os,sys,json\nwith open(" + encode(audit) + ",'a') as f: f.write(json.dumps(sys.argv[1:])+'\\n')\nos.execv(" + encode(compiler) + ", [" + encode(compiler) + ",*sys.argv[1:]])\n"
	if err = os.WriteFile(filepath.Join(wrapperDir, "clang"), []byte(wrapper), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", wrapperDir+string(os.PathListSeparator)+filepath.Dir(profdata)+string(os.PathListSeparator)+os.Getenv("PATH"))
	// The audit wrapper is itself a different compiler executable; bind it explicitly.
	record, err = NewStage1ProfileManifest(source, data, "test-only-corpus")
	if err != nil {
		t.Fatal(err)
	}
	writeManifest(record)
	runtimeFiles, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	expectedRuntime := map[string]bool{}
	for _, file := range runtimeFiles {
		if strings.HasSuffix(file.name, ".c") {
			expectedRuntime[file.name] = true
		}
	}
	options := Options{Release: true, Profile: text}
	if _, err := RuntimeLibrary("", options); err == nil {
		t.Fatal("unverified profile reached the runtime builder")
	}
	checkCommands := func(profile bool, stable bool) {
		t.Helper()
		data, err := os.ReadFile(audit)
		if err != nil {
			t.Fatal(err)
		}
		compiles, links := 0, 0
		seen := map[string]bool{}
		for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
			var args []string
			if err = json.Unmarshal([]byte(line), &args); err != nil {
				t.Fatal(err)
			}
			if slices.Contains(args, "--version") {
				continue
			}
			if slices.Contains(args, "-c") {
				compiles++
			} else {
				links++
			}
			if stable {
				units := 0
				for _, arg := range args {
					if !strings.HasSuffix(arg, ".c") {
						continue
					}
					units++
					if slices.Contains(args, "-c") {
						if !expectedRuntime[arg] || seen[arg] {
							t.Fatalf("runtime profile source identity changed or duplicated: %q", args)
						}
						seen[arg] = true
					} else if arg != record.Units[0].Name {
						t.Fatalf("emitted profile source identity changed: got %q want %q", arg, record.Units[0].Name)
					}
				}
				if units != 1 {
					t.Fatalf("profile invocation did not consume exactly one recorded C unit: %q", args)
				}
			}
			present := false
			for _, arg := range args {
				if strings.HasPrefix(arg, "-fprofile-use=") {
					present = true
				}
			}
			if present != profile {
				t.Fatalf("stale/profile flag reached wrong clang invocation: profile=%v argv=%q", profile, args)
			}
			for _, flag := range []string{"-ffp-contract=off", "-fno-optimize-sibling-calls", "-flto=thin"} {
				if !slices.Contains(args, flag) {
					t.Fatalf("missing semantic/release flag %s: %q", flag, args)
				}
			}
		}
		if links != 1 || (stable && compiles != 48) {
			t.Fatalf("incomplete actual command audit: compile=%d link=%d", compiles, links)
		}
	}
	reset := func() {
		t.Helper()
		if err = os.WriteFile(audit, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Audit training identities too; the same profile contract applies on both sides.
	reset()
	if err = Build(source, generate, Options{Release: true, ProfileGenerate: true}); err != nil {
		t.Fatal(err)
	}
	checkCommands(false, true)
	first, second := filepath.Join(directory, "first"), filepath.Join(directory, "second")
	reset()
	if err = Build(source, first, options); err != nil {
		t.Fatal(err)
	}
	checkCommands(true, true)
	t.Setenv("XDG_CACHE_HOME", filepath.Join(directory, "independent-cold-cache"))
	if err = Build(source, second, options); err != nil {
		t.Fatal(err)
	}
	a, err := os.ReadFile(first)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(second)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("same source and profile produced different binary bytes")
	}
	b = append([]byte{}, b...)
	b[len(b)/2] ^= 1
	if bytes.Equal(a, b) {
		t.Fatal("one-byte binary determinism mutant survived")
	}
	t.Logf("two profile builds are byte-identical: %s; all 48 runtime compiles and link retain both semantic flags", profileHash(a))
	capture := func(buildSource string) string {
		t.Helper()
		reset()
		log, err := os.Create(filepath.Join(directory, "fallback.stderr"))
		if err != nil {
			t.Fatal(err)
		}
		previous := os.Stderr
		os.Stderr = log
		buildErr := Build(buildSource, filepath.Join(directory, "fallback"), options)
		os.Stderr = previous
		log.Close()
		if buildErr != nil {
			t.Fatal(buildErr)
		}
		body, err := os.ReadFile(log.Name())
		if err != nil {
			t.Fatal(err)
		}
		checkCommands(false, false)
		if strings.Count(string(body), "\n") != 1 || !strings.Contains(string(body), "using plain ThinLTO") {
			t.Fatalf("missing one-line fallback diagnostic: %q", body)
		}
		return string(body)
	}
	changed := strings.Replace(source, "// A", "// B", 1)
	t.Log("one changed emitted-C byte:", capture(changed))
	// Each mismatch must produce an actual plain-ThinLTO build: no profile flag
	// on any runtime compile or the main compilation/link, plus one diagnostic.
	mutations := []struct {
		name   string
		change func(*Stage1ProfileManifest)
	}{
		{"manifest version", func(r *Stage1ProfileManifest) { r.Version = 1 }},
		{"runtime snapshot", func(r *Stage1ProfileManifest) { r.Runtime = "changed-runtime-byte" }},
		{"emitted unit name", func(r *Stage1ProfileManifest) { r.Units[0].Name = "other.c" }},
		{"emitted unit byte", func(r *Stage1ProfileManifest) { r.Units[0].SHA256 = "changed-unit-byte" }},
		{"added emitted unit", func(r *Stage1ProfileManifest) {
			r.Units = append(r.Units, Stage1ProfileUnit{Name: "extra.c", SHA256: profileHash([]byte("int extra;"))})
		}},
		{"text profile", func(r *Stage1ProfileManifest) { r.Profile = "changed-profile-byte" }},
		{"compiler executable", func(r *Stage1ProfileManifest) { r.CompilerSHA256 = "changed-compiler-byte" }},
		{"compiler version", func(r *Stage1ProfileManifest) { r.Compiler += "changed" }},
		{"compile flags", func(r *Stage1ProfileManifest) { r.Flags[0] = "-std=c99" }},
		{"link flags", func(r *Stage1ProfileManifest) { r.LinkFlags[0] = "-std=c99" }},
		{"training compile flags", func(r *Stage1ProfileManifest) { r.TrainingFlags[0] = "-std=c99" }},
		{"training link flags", func(r *Stage1ProfileManifest) { r.TrainingLinkFlags[0] = "-std=c99" }},
		{"target", func(r *Stage1ProfileManifest) { r.Target = "different-target" }},
	}
	for _, mutation := range mutations {
		changed := record
		changed.Units = slices.Clone(record.Units)
		changed.Flags = slices.Clone(record.Flags)
		changed.LinkFlags = slices.Clone(record.LinkFlags)
		changed.TrainingFlags = slices.Clone(record.TrainingFlags)
		changed.TrainingLinkFlags = slices.Clone(record.TrainingLinkFlags)
		mutation.change(&changed)
		writeManifest(changed)
		t.Log(mutation.name, "mismatch:", capture(source))
	}
	writeManifest(record)
	originalOptions := options
	options.cpu = "native"
	t.Log("actual compile/link flags changed:", capture(source))
	options = originalOptions
	for _, opts := range []Options{{Profile: text}, {Release: true, Count: true, Profile: text}, {Release: true, Sanitize: true, Profile: text}, {Release: true, Target: "wasm32-wasi", Profile: text}} {
		for _, flags := range [][]string{Flags(opts), LinkFlags(opts)} {
			for _, flag := range flags {
				if strings.HasPrefix(flag, "-fprofile") {
					t.Fatalf("profile escaped into nonshipping flags: %q", flags)
				}
			}
		}
	}
}

func TestRuntimeFingerprintCoversEveryFile(t *testing.T) {
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		t.Fatal(err)
	}
	original := runtimeFingerprint(files)
	for index, file := range files {
		if len(file.contents) == 0 {
			t.Fatalf("empty runtime input: %s", file.name)
		}
		changed := slices.Clone(files)
		changed[index].contents = slices.Clone(file.contents)
		changed[index].contents[len(file.contents)/2] ^= 1
		if runtimeFingerprint(changed) == original {
			t.Fatalf("changed runtime byte did not invalidate %s", file.name)
		}
		changed = slices.Clone(files)
		changed[index].name += ".changed"
		if runtimeFingerprint(changed) == original {
			t.Fatalf("changed runtime name did not invalidate %s", file.name)
		}
	}
	if runtimeFingerprint(files[:len(files)-1]) == original {
		t.Fatal("removed runtime file did not invalidate snapshot")
	}
	t.Logf("a real changed byte and name invalidate each of %d runtime C/header files", len(files))
}

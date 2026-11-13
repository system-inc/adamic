package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const platformCompileScript = "#!/bin/sh\necho 'platform unavailable' >&2\nexit 5\n"
const platformBuildScript = `#!/bin/sh
case "$1" in
admission-lower) exit 0;;
js) echo "console.log('JS observed');";;
build) echo 'native platform unavailable' >&2; exit 5;;
esac
`

func platformWrite(t *testing.T, dir, path, text string) {
	t.Helper()
	name := filepath.Join(dir, path)
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
}
func platformCommit(t *testing.T, dir string) string {
	t.Helper()
	if _, err := git(dir, "add", "."); err != nil {
		t.Fatal(err)
	}
	if _, err := git(dir, "commit", "-qm", "platform fixture"); err != nil {
		t.Fatal(err)
	}
	sha, err := git(dir, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	return sha
}
func platformRun(t *testing.T, dir, base, head, baseCompiler, headCompiler, corpusDir string) (report, observation) {
	t.Helper()
	listing, err := git(dir, "ls-tree", "-r", "--name-only", head, "--", corpusDir)
	if err != nil {
		t.Fatal(err)
	}
	c := corpusEntry(listing)
	for index := range c.Programs {
		blob, err := git(dir, "rev-parse", head+":"+c.Programs[index].Path)
		if err != nil {
			t.Fatal(err)
		}
		c.Programs[index].Blob = blob
	}
	data, err := json.Marshal(manifest{SHA: head, Corpora: []corpus{c}})
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(manifestPath, data, 0600); err != nil {
		t.Fatal(err)
	}
	o := invokeCommand(t, dir, "--base", base, "--head", head, "--base-binary", baseCompiler, "--head-binary", headCompiler, "--manifest", manifestPath, "--json")
	var r report
	if err := json.Unmarshal([]byte(o.Stdout), &r); err != nil {
		t.Fatalf("%v: %s %s", err, o.Stdout, o.Stderr)
	}
	return r, o
}
func requireDarwin(t *testing.T, r report, o observation, path string) {
	t.Helper()
	if o.Exit != 0 || r.Verdict != "pass" || r.Void || r.DarwinCount != 1 || len(r.DarwinPaths) != 1 || r.DarwinPaths[0] != path || r.Admitted != 0 || r.NodeAgreements != 0 || len(r.Programs) != 1 {
		t.Fatalf("darwin row judged: exit %d %+v %s", o.Exit, r, o.Stderr)
	}
	p := r.Programs[0]
	if p.Class != "darwin" || p.Agree != nil || p.VoidReason != "darwin: unjudged here; judged by the darwin witnesses" {
		t.Fatalf("row is not wholly unjudged: %+v", p)
	}
	if !strings.Contains(o.Stderr, "darwin: 1 (unjudged here; judged by the darwin witnesses)\n") {
		t.Fatal("missing counted class", o.Stderr)
	}
}

// Each subtest fails with its parser case removed. Comments, strings and dirty
// working files cannot supply the first key. Local module cycles stay bounded.
func TestDarwinImportForms(t *testing.T) {
	t.Parallel()
	for _, form := range []struct{ name, source string }{
		{"import", "import {View} from 'apple/appkit/view';"},
		{"import_type", "import type {View} from 'apple/appkit/view';"},
		{"export_from", "export {View} from 'apple/appkit/view';"},
		{"side_effect", "import 'apple/appkit/view';"},
		{"dynamic", "void import('apple/appkit/view');"},
		{"import_type_expression", "type View = import('apple/appkit/view').View;"},
		{"import_equals", "import View = require('apple/appkit/view');"},
		{"syntax_after_import", "import 'apple/appkit/view'; const broken = ;"},
		{"local_module", "import './bridge.js';"},
	} {
		t.Run(form.name, func(t *testing.T) {
			dir, base, refusing, headCompiler := commandFixture(t)
			platformWrite(t, dir, "corpus/input.a", form.source+"\nconsole.log('42');\n")
			if form.name == "local_module" {
				platformWrite(t, dir, "corpus/bridge.ts", "export * from './nested';\n")
				platformWrite(t, dir, "corpus/nested/index.ts", "import '../bridge'; import type {View} from 'apple/appkit/view';\n")
			}
			head := platformCommit(t, dir)
			// The command must read head, not this edit in the invoking tree.
			platformWrite(t, dir, "corpus/input.a", "console.log('dirty');\n")
			writeExecutable(t, headCompiler, platformCompileScript)
			r, o := platformRun(t, dir, base, head, refusing, headCompiler, "corpus")
			requireDarwin(t, r, o, "corpus/input.a")
			if r.Programs[0].Head.Exit != 5 || r.Programs[0].Node != nil || r.Programs[0].Sampled {
				t.Fatal("compile-stop observations changed", r.Programs[0])
			}
		})
	}
}

func TestPlatformStopWithoutApple(t *testing.T) {
	t.Parallel()
	for _, phase := range []struct{ name, script string }{{"compile", platformCompileScript}, {"native", platformBuildScript}} {
		t.Run(phase.name, func(t *testing.T) {
			dir, sha, base, head := commandFixture(t)
			writeExecutable(t, head, phase.script)
			r, o := platformRun(t, dir, sha, sha, base, head, "corpus")
			if o.Exit != 1 || r.Verdict != "fail" || r.Void || r.DarwinCount != 0 || r.Programs[0].Class != "platform-stop-without-apple" {
				t.Fatalf("unkeyed platform stop hidden: %+v %s", r, o.Stderr)
			}
		})
	}
}

func TestDarwinRejectsImportLookalikes(t *testing.T) {
	t.Parallel()
	for _, source := range []string{"// import 'apple/appkit/view';\nconsole.log('42');", "const broken = ; // import 'apple/view';", "const text = `import('apple/appkit/view')`;", "void import('apple/' + 'view');", "object.import('apple/view');", "import 'pineapple/view';"} {
		dir, base, refusing, headCompiler := commandFixture(t)
		platformWrite(t, dir, "corpus/input.a", source)
		platformWrite(t, dir, "corpus/unused.ts", "import 'apple/view';")
		head := platformCommit(t, dir)
		writeExecutable(t, headCompiler, platformCompileScript)
		r, o := platformRun(t, dir, base, head, refusing, headCompiler, "corpus")
		if o.Exit != 1 || r.Programs[0].Class != "platform-stop-without-apple" {
			t.Fatalf("non-import gave first key: %q %+v", source, r)
		}
	}
}

func TestMovedIntoDarwin(t *testing.T) {
	t.Parallel()
	dir, _, baseCompiler, headCompiler := commandFixture(t)
	platformWrite(t, dir, "corpus/input.a", "import './bridge'; console.log('42');")
	platformWrite(t, dir, "corpus/bridge.ts", "import 'apple/view';")
	base := platformCommit(t, dir)
	platformWrite(t, dir, "corpus/input.a", "import './bridge'; console.log('42 changed');")
	head := platformCommit(t, dir)
	// A head-only or working-tree read of the base dependency would miss the key.
	platformWrite(t, dir, "corpus/bridge.ts", "export {};")
	writeExecutable(t, baseCompiler, "#!/bin/sh\nexit 0\n")
	writeExecutable(t, headCompiler, platformCompileScript)
	r, o := platformRun(t, dir, base, head, baseCompiler, headCompiler, "corpus")
	if o.Exit != 1 || r.Verdict != "fail" || r.Void || r.DarwinCount != 0 || r.Programs[0].Class != "moved-into-darwin" {
		t.Fatalf("existing program moved into exemption: %+v %s", r, o.Stderr)
	}
}

func TestNewFileEntersDarwin(t *testing.T) {
	t.Parallel()
	dir, base, refusing, headCompiler := commandFixture(t)
	platformWrite(t, dir, "programs/new.a", "import 'apple/view';")
	head := platformCommit(t, dir)
	writeExecutable(t, headCompiler, platformCompileScript)
	r, o := platformRun(t, dir, base, head, refusing, headCompiler, "programs")
	requireDarwin(t, r, o, "programs/new.a")
}

// views.a's road: lowering succeeds; native supplies the typed stop. Node and
// JavaScript deliberately differ, and stay recorded without becoming a verdict.
func TestDarwinNativeStopKeepsObservations(t *testing.T) {
	t.Parallel()
	dir, base, refusing, headCompiler := commandFixture(t)
	platformWrite(t, dir, "corpus/input.a", "import 'apple/swiftui/native'; console.log('42');")
	head := platformCommit(t, dir)
	writeExecutable(t, headCompiler, platformBuildScript)
	r, o := platformRun(t, dir, base, head, refusing, headCompiler, "corpus")
	requireDarwin(t, r, o, "corpus/input.a")
	p := r.Programs[0]
	if p.Head.Exit != 0 || p.Node == nil || p.Node.Stdout != "42\n" || p.JavaScriptCompile == nil || p.JavaScript == nil || p.JavaScript.Stdout != "changed\n" || p.NativeCompile == nil || p.NativeCompile.Exit != 5 || p.NativeCompile.Stderr != "native platform unavailable\n" || p.Native == nil || !p.Sampled {
		t.Fatalf("lost native-stop observations: %+v", p)
	}
}

func TestLeavingDarwinIsJudged(t *testing.T) {
	t.Parallel()
	dir, _, baseCompiler, headCompiler := commandFixture(t)
	platformWrite(t, dir, "corpus/input.a", "import 'apple/view'; console.log('42');")
	base := platformCommit(t, dir)
	platformWrite(t, dir, "corpus/input.a", "console.log('42');")
	head := platformCommit(t, dir)
	writeExecutable(t, baseCompiler, platformCompileScript)
	// commandFixture's head prints 41 natively, so leaving darwin must fail.
	r, o := platformRun(t, dir, base, head, baseCompiler, headCompiler, "corpus")
	if o.Exit != 1 || r.DarwinCount != 0 || r.Programs[0].Class != "newly-accepted" || r.Programs[0].Agree == nil || *r.Programs[0].Agree {
		t.Fatalf("leaving darwin escaped judgment: %+v %s", r, o.Stderr)
	}
}

func corpusEntry(listing string) corpus {
	c := corpus{Name: "fixtures", Programs: []program{}}
	for _, path := range strings.Split(listing, "\n") {
		if strings.HasSuffix(path, ".a") {
			c.Programs = append(c.Programs, program{Path: path})
		}
	}
	return c
}

func TestMovedIntoDarwinNative(t *testing.T) {
	t.Parallel()
	dir, _, baseCompiler, headCompiler := commandFixture(t)
	platformWrite(t, dir, "corpus/input.a", "import 'apple/view'; console.log('42');")
	base := platformCommit(t, dir)
	writeExecutable(t, baseCompiler, "#!/bin/sh\nexit 0\n")
	writeExecutable(t, headCompiler, platformBuildScript)
	r, o := platformRun(t, dir, base, base, baseCompiler, headCompiler, "corpus")
	if o.Exit != 1 || r.Programs[0].Class != "moved-into-darwin" || r.Programs[0].NativeCompile == nil || r.Programs[0].NativeCompile.Exit != 5 {
		t.Fatalf("native stop escaped selection: %+v %s", r, o.Stderr)
	}
}

func TestDarwinWholeRowIgnoresBaseError(t *testing.T) {
	t.Parallel()
	dir, base, baseCompiler, headCompiler := commandFixture(t)
	platformWrite(t, dir, "corpus/input.a", "import 'apple/view'; console.log('42');")
	head := platformCommit(t, dir)
	writeExecutable(t, baseCompiler, "#!/bin/sh\necho 'unrecognized compiler failure' >&2\nexit 1\n")
	writeExecutable(t, headCompiler, platformBuildScript)
	r, o := platformRun(t, dir, base, head, baseCompiler, headCompiler, "corpus")
	requireDarwin(t, r, o, "corpus/input.a")
	if r.Programs[0].Base.Stderr != "unrecognized compiler failure\n" {
		t.Fatal("base observation erased")
	}
}

func TestDarwinDoesNotHideSiblingFailure(t *testing.T) {
	t.Parallel()
	dir, base, baseCompiler, headCompiler := commandFixture(t)
	platformWrite(t, dir, "corpus/input.a", "import 'apple/view'; console.log('42');")
	platformWrite(t, dir, "corpus/bad.a", "console.log('42');")
	head := platformCommit(t, dir)
	writeExecutable(t, headCompiler, "#!/bin/sh\ncase \"$2\" in *bad.a) echo 'unrecognized compiler failure' >&2; exit 1;; esac\nexit 5\n")
	r, o := platformRun(t, dir, base, head, baseCompiler, headCompiler, "corpus")
	if o.Exit != 1 || r.Verdict != "fail" || r.DarwinCount != 1 || r.Void {
		t.Fatalf("darwin hid a sibling: %+v %s", r, o.Stderr)
	}
}

func TestAppleImportWithoutPlatformStopIsJudged(t *testing.T) {
	t.Parallel()
	dir, base, baseCompiler, headCompiler := commandFixture(t)
	platformWrite(t, dir, "corpus/input.a", "import 'apple/view'; console.log('42');")
	head := platformCommit(t, dir)
	r, o := platformRun(t, dir, base, head, baseCompiler, headCompiler, "corpus")
	if o.Exit != 1 || r.DarwinCount != 0 || r.Programs[0].Class != "newly-accepted" || r.Programs[0].Agree == nil || *r.Programs[0].Agree {
		t.Fatalf("import alone hid a mismatch: %+v %s", r, o.Stderr)
	}
}

func TestDarwinDoesNotJudgeOldStopAtItsPath(t *testing.T) {
	t.Parallel()
	dir, _, baseCompiler, headCompiler := commandFixture(t)
	source, _, list := runtimeNotYetFixture(t)
	platformWrite(t, dir, "corpus/input.a", string(source))
	writeRuntimeNotYetList(t, filepath.Join(dir, "cloud/admission-corpus/runtime-notyet-stops.json"), list)
	base := platformCommit(t, dir)
	platformWrite(t, dir, "corpus/input.a", "import 'apple/view'; console.log('42');")
	head := platformCommit(t, dir)
	writeExecutable(t, headCompiler, platformCompileScript)
	r, o := platformRun(t, dir, base, head, baseCompiler, headCompiler, "corpus")
	requireDarwin(t, r, o, "corpus/input.a")
	if len(r.StaleRuntimeStops) != 0 {
		t.Fatal("judged an old stop at an unjudged row", r.StaleRuntimeStops)
	}
}

func TestDarwinRelativeResolution(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, specifier string
		files           map[string]string
		apple           bool
	}{
		{"extensionless_a", "./bridge", map[string]string{"corpus/bridge.a": "import 'apple/view';"}, false},
		{"explicit_a", "./bridge.a", map[string]string{"corpus/bridge.a": "import 'apple/view';"}, true},
		{"real_ts_twin", "./bridge.a", map[string]string{"corpus/bridge.a": "import 'apple/view';", "corpus/bridge.a.ts": "export {};"}, false},
		{"directory_index_a", "./bridge", map[string]string{"corpus/bridge/index.a": "import 'apple/view';"}, false},
		{"jsx_substitution", "./bridge.jsx", map[string]string{"corpus/bridge.tsx": "import 'apple/view';"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir, base, baseCompiler, headCompiler := commandFixture(t)
			platformWrite(t, dir, "corpus/input.a", "import '"+tc.specifier+"'; console.log('42');")
			for path, source := range tc.files {
				platformWrite(t, dir, path, source)
			}
			head := platformCommit(t, dir)
			writeExecutable(t, headCompiler, "#!/bin/sh\ncase \"$2\" in corpus/input.a) exit 5;; *) echo 'error TS2307: fixture module' >&2; exit 1;; esac\n")
			r, o := platformRun(t, dir, base, head, baseCompiler, headCompiler, "corpus")
			found := false
			for _, p := range r.Programs {
				if p.Path == "corpus/input.a" {
					found = true
					want := "platform-stop-without-apple"
					if tc.apple {
						want = "darwin"
					}
					if p.Class != want || p.AppleImport != tc.apple {
						t.Fatalf("relative source key: %+v, want %s", p, want)
					}
				}
			}
			if !found || (tc.apple && o.Exit != 0) || (!tc.apple && o.Exit != 1) {
				t.Fatalf("relative run: %+v %s", r, o.Stderr)
			}
		})
	}
}

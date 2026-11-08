package oracle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// installed-layout's error is Node's answer to require("source-map-support") where tsc 6.0.3 is
// installed, which ships no dependencies: its name, code and the first line of its message. The
// data file's rows say native's output in them, TestRuledBackendDivergencesSayTheMeasuredErrors holds
// the rows to them, and TestInstalledLayoutIsNodes measures them on Node.
const (
	inspectorUnavailableName    = "Error"
	inspectorUnavailableCode    = "ERR_INSPECTOR_NOT_AVAILABLE"
	inspectorUnavailableMessage = "Inspector is not available"
	inspectorUnavailableString  = "Error [ERR_INSPECTOR_NOT_AVAILABLE]: Inspector is not available"

	installedLayoutName    = "Error"
	installedLayoutCode    = "MODULE_NOT_FOUND"
	installedLayoutMessage = "Cannot find module 'source-map-support'"
)

// noJavaScriptPlugins is the message native's computed require throws for a path, in Diagnostics'
// words (#redxes2).
func noJavaScriptPlugins(path string) string {
	return "Cannot load JavaScript module '" + path + "': a native Adamic program can't run JavaScript; run it on Node to load modules at run time"
}

// requireInESModule is what Node gives the source of a fixture that calls require: it runs every
// program as an ES module (oracle/node.mjs), where require isn't defined. A catch sees the first
// line, and Node adds the rest to the message as an uncaught one leaves the module.
const requireInESModule = "ReferenceError: require is not defined"

// backendDivergencesPath is the ruled rows' one home, a data file, since the admission delta
// (cmd/adamic-admission-delta) reads the same rows (#tqm95v6). Only Adamic adds a row.
const backendDivergencesPath = "internal/oracle/testdata/backend_divergences.json"

// Only these exact fixtures have ruled differences. Each side remains pinned:
// membership never excuses an arbitrary exit, diagnostic or output change. A row whose node is set
// pins the source on Node too, for fixtures where native and the JavaScript backend agree with each
// other and differ from the source; name is the ruled row's, which may hold more than one fixture. A
// row whose format is lower.ModuleFormatCommonJSBundle is built only as one bundled image, so its
// fixture is never among the oracle's, which lower every program as an ES module: the installed-layout
// harness runs it (TestInstalledLayoutPathDivergences), with each leg's lib directory made <lib>.
// The rows are read once, as the package loads: a list that can't be read panics there, failing every
// test in the package rather than comparing against no rows.
var ruledBackendDivergences = func() []ruledBackendDivergence {
	rows, err := readRuledBackendDivergences(filepath.Join(repository, backendDivergencesPath))
	if err != nil {
		panic(err)
	}
	return rows
}()

type ruledBackendDivergence struct {
	fixture    string
	name       string
	ruling     string
	reason     string
	format     lower.ModuleFormat
	node       *run
	native     run
	javascript run
}

// backendDivergenceOutcome is one side of a row as the data file says it, its literal bytes.
type backendDivergenceOutcome struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

func (outcome backendDivergenceOutcome) run() run {
	return run{stdout: []byte(outcome.Stdout), stderr: []byte(outcome.Stderr), exitCode: outcome.ExitCode}
}

// readRuledBackendDivergences reads the rows in the file's order. A row with no fixture, no ruling or no
// reason is refused, as is a format no build declares and a field the rows don't have.
func readRuledBackendDivergences(path string) ([]ruledBackendDivergence, error) {
	var list struct {
		Comment     string `json:"_comment"`
		Divergences []struct {
			Fixture    string                    `json:"fixture"`
			Name       string                    `json:"name"`
			Ruling     string                    `json:"ruling"`
			Reason     string                    `json:"reason"`
			Format     lower.ModuleFormat        `json:"format"`
			Node       *backendDivergenceOutcome `json:"node"`
			Native     backendDivergenceOutcome  `json:"native"`
			JavaScript backendDivergenceOutcome  `json:"javascript"`
		} `json:"divergences"`
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&list); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if !strings.Contains(list.Comment, "Adding a row is a ruling") || !strings.Contains(list.Comment, "@system_adamic") {
		return nil, fmt.Errorf("%s needs its ruling header", path)
	}
	var rows []ruledBackendDivergence
	for index, row := range list.Divergences {
		if row.Fixture == "" || strings.TrimSpace(row.Ruling) == "" || row.Reason == "" {
			return nil, fmt.Errorf("%s: row %d needs its fixture, its ruling and its reason", path, index+1)
		}
		if row.Format != lower.ModuleFormatEcmaScript && row.Format != lower.ModuleFormatCommonJSBundle {
			return nil, fmt.Errorf("%s: row %d declares format %q, which no build declares", path, index+1, row.Format)
		}
		divergence := ruledBackendDivergence{fixture: row.Fixture, name: row.Name, ruling: row.Ruling, reason: row.Reason, format: row.Format, native: row.Native.run(), javascript: row.JavaScript.run()}
		if row.Node != nil {
			node := row.Node.run()
			divergence.node = &node
		}
		rows = append(rows, divergence)
	}
	return rows, nil
}

// The two divergences' reasons, each shared by its fixtures' rows, which the data file must say.
const (
	installedLayoutReason     = "require(\"source-map-support\") throws Node's MODULE_NOT_FOUND where tsc 6.0.3 is installed, which ships no dependencies; native gives its name, code and the first line of its message. The Require stack tail and requireStack are the requiring module's own path, which has no reproducible native answer yet, and tsc's catch ignores the error, so tsc observes nothing. The source on Node, an ES module, has no require."
	noJavaScriptPluginsReason = "A computed require evaluates its path and throws an Error naming it. Stock Node loads the plugin; native can't run JavaScript. The source on Node, an ES module, has no require."
)

// backendDisagreement applies a named ruling or the ordinary byte-exact comparison.
// Every unlisted inserted check must agree on both sides.
func backendDisagreement(fixture string, javascript, native run) string {
	for _, entry := range ruledBackendDivergences {
		if entry.fixture != fixture {
			continue
		}
		if difference := disagreement(entry.native, native); difference != "" {
			return fmt.Sprintf("ruled native outcome: %s (%s)", difference, entry.ruling)
		}
		if difference := disagreement(entry.javascript, javascript); difference != "" {
			return fmt.Sprintf("ruled JavaScript outcome: %s (%s)", difference, entry.ruling)
		}
		return ""
	}
	return disagreement(javascript, native)
}

// sourceDisagreement holds a backend's run (native, or the JavaScript backend's when javascript is
// set) to the source on Node: byte for byte, or for a ruled fixture that pins Node's side, each to
// its own pinned outcome, so neither can change unseen.
func sourceDisagreement(fixture string, source, actual run, javascript bool) string {
	for _, entry := range ruledBackendDivergences {
		if entry.fixture != fixture || entry.node == nil {
			continue
		}
		if difference := disagreement(*entry.node, source); difference != "" {
			return fmt.Sprintf("ruled Node outcome: %s (%s)", difference, entry.ruling)
		}
		pinned, side := entry.native, "native"
		if javascript {
			pinned, side = entry.javascript, "JavaScript"
		}
		if difference := disagreement(pinned, actual); difference != "" {
			return fmt.Sprintf("ruled %s outcome: %s (%s)", side, difference, entry.ruling)
		}
		return ""
	}
	return disagreement(source, actual)
}

// ruledNativeOutcome is a ruled fixture's native outcome, the witness for its wasm32 build, which is native too.
func ruledNativeOutcome(fixture string) (run, bool) {
	for _, entry := range ruledBackendDivergences {
		if entry.fixture == fixture {
			return entry.native, true
		}
	}
	return run{}, false
}

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/d96d304_try_stack.a",
		"internal/oracle/testdata/backend_stop_catches.a",
		"internal/oracle/testdata/node_inspector_unavailable.a",
		"internal/oracle/testdata/node_inspector_unavailable_uncaught.a",
		"internal/oracle/testdata/node_commonjs_require_missing.a",
		"internal/oracle/testdata/node_commonjs_require_missing_uncaught.a",
		"internal/oracle/testdata/node_commonjs_require_computed.a",
		"internal/oracle/testdata/node_commonjs_require_computed_uncaught.a",
		"internal/oracle/testdata/node_commonjs_require_missing_cast.a",
		"internal/oracle/testdata/node_commonjs_require_computed_return.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, path == "internal/oracle/testdata/d96d304_try_stack.a"})
	}
}

// The data file holds expanded literals, so this holds them to the constants Node was measured against
// (TestInstalledLayoutIsNodes, the inspector measurement): each family's rows give its reason where the
// family shares one, both backends' outcomes say one of its measured errors, and Node's says require
// isn't defined in its ES module. A remeasured error can't leave a row saying the old one.
func TestRuledBackendDivergencesSayTheMeasuredErrors(t *testing.T) {
	t.Parallel()
	families := map[string]struct {
		reason string
		says   []string
	}{
		"installed-layout": {installedLayoutReason, []string{
			installedLayoutName + " " + installedLayoutCode + " " + installedLayoutMessage + "\n",
			"adamic: panic: " + installedLayoutName + ": " + installedLayoutMessage + "\n",
		}},
		"no-javascript-plugins":  {noJavaScriptPluginsReason, []string{noJavaScriptPlugins("./plugin-1.js"), noJavaScriptPlugins("./plugin.js")}},
		"node-without-inspector": {"", []string{inspectorUnavailableString + "\n"}},
	}
	rows := map[string]int{}
	for _, row := range ruledBackendDivergences {
		family, named := families[row.name]
		if !named {
			continue
		}
		rows[row.name]++
		if family.reason != "" && row.reason != family.reason {
			t.Errorf("%s: the reason isn't %s's", row.fixture, row.name)
		}
		for side, outcome := range map[string]run{"native": row.native, "JavaScript": row.javascript} {
			said := false
			for _, measured := range family.says {
				said = said || strings.Contains(string(outcome.stdout)+string(outcome.stderr), measured)
			}
			if !said {
				t.Errorf("%s: the %s outcome says none of %s's measured errors %q", row.fixture, side, row.name, family.says)
			}
		}
		if row.node == nil || !strings.Contains(string(row.node.stdout)+string(row.node.stderr), strings.TrimPrefix(requireInESModule, "ReferenceError: ")) {
			t.Errorf("%s: Node's outcome doesn't say %q", row.fixture, requireInESModule)
		}
	}
	for name := range families {
		if rows[name] == 0 {
			t.Errorf("no row is %s", name)
		}
	}
}

// The loader refuses a row whose ruling is empty or only spaces, and reads one that names its ruling.
// Mutant: readRuledBackendDivergences without its ruling check reads both.
func TestRuledBackendDivergencesRefuseAnEmptyRuling(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "divergences.json")
	read := func(ruling string) error {
		t.Helper()
		row := fmt.Sprintf(`{"fixture": "internal/oracle/testdata/d96d304_try_stack.a", "ruling": %q, "reason": "a reason", "native": {"stdout": "", "stderr": "", "exit_code": 70}, "javascript": {"stdout": "", "stderr": "", "exit_code": 70}}`, ruling)
		if err := os.WriteFile(path, []byte(`{"_comment": "Adding a row is a ruling routed to @system_adamic", "divergences": [`+row+`]}`), 0600); err != nil {
			t.Fatal(err)
		}
		_, err := readRuledBackendDivergences(path)
		return err
	}
	if err := read("@system_adamic, task #z1vjxxd, chain step 24"); err != nil {
		t.Fatal("a ruled row refused:", err)
	}
	for _, ruling := range []string{"", "   "} {
		if read(ruling) == nil {
			t.Errorf("a row whose ruling is %q read", ruling)
		}
	}
}

// The loader reads a row declared one bundled image and refuses a format no build declares, which would otherwise
// leave its fixture among neither the oracle's nor the installed-layout harness's. Mutant: readRuledBackendDivergences
// without its format check reads "CommonJS".
func TestRuledBackendDivergencesRefuseAnUnknownFormat(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "divergences.json")
	read := func(format string) ([]ruledBackendDivergence, error) {
		t.Helper()
		row := fmt.Sprintf(`{"fixture": "internal/oracle/testdata/installed_layout/filename.ts", "ruling": "@system_adamic, task #redxes2", "reason": "a reason", "format": %q, "native": {"stdout": "", "stderr": "", "exit_code": 0}, "javascript": {"stdout": "", "stderr": "", "exit_code": 0}}`, format)
		if err := os.WriteFile(path, []byte(`{"_comment": "Adding a row is a ruling routed to @system_adamic", "divergences": [`+row+`]}`), 0600); err != nil {
			t.Fatal(err)
		}
		return readRuledBackendDivergences(path)
	}
	rows, err := read(string(lower.ModuleFormatCommonJSBundle))
	if err != nil || len(rows) != 1 || rows[0].format != lower.ModuleFormatCommonJSBundle {
		t.Fatalf("a bundled row read as %+v, %v", rows, err)
	}
	if _, err := read("CommonJS"); err == nil {
		t.Error("a row declaring format CommonJS read")
	}
}

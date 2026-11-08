package oracle

import "fmt"

// installed-layout's error is Node's answer to require("source-map-support") where tsc 6.0.3 is
// installed, which ships no dependencies: its name, code and the first line of its message. The
// rows below pin native's output from them, and TestInstalledLayoutIsNodes measures them on Node.
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

// Only these exact fixtures have ruled differences. Each side remains pinned:
// membership never excuses an arbitrary exit, diagnostic or output change. A row whose node is set
// pins the source on Node too, for fixtures where native and the JavaScript backend agree with each
// other and differ from the source; name is the ruled row's, which may hold more than one fixture.
var ruledBackendDivergences = []struct {
	fixture    string
	name       string
	ruling     string
	reason     string
	node       *run
	native     run
	javascript run
}{
	{
		fixture:    "internal/oracle/testdata/d96d304_try_stack.a",
		ruling:     "@system_adamic, task #z1vjxxd, chain step 24",
		reason:     "Native stack guards name the exhausted function; JavaScript preserves V8's unnamed overflow diagnostic. Both stops are terminal.",
		native:     run{stderr: []byte("adamic: panic: RangeError: Maximum call stack size exceeded in depth\n"), exitCode: 70},
		javascript: run{stderr: []byte("adamic: panic: RangeError: Maximum call stack size exceeded\n"), exitCode: 70},
	},

	{fixture: "internal/oracle/testdata/node_inspector_unavailable.a", name: "node-without-inspector",
		ruling:     "@system_adamic and @system_adamic_platforms, task #cn4yapb, Nov 5",
		reason:     "Native and generated JavaScript model Node built --without-inspector; the source is an ES module with no require.",
		node:       &run{stdout: []byte("ReferenceError\nnone\nrequire is not defined\nReferenceError: require is not defined\nafter\n")},
		native:     run{stdout: []byte(inspectorUnavailableName + "\n" + inspectorUnavailableCode + "\n" + inspectorUnavailableMessage + "\n" + inspectorUnavailableString + "\nafter\n")},
		javascript: run{stdout: []byte(inspectorUnavailableName + "\n" + inspectorUnavailableCode + "\n" + inspectorUnavailableMessage + "\n" + inspectorUnavailableString + "\nafter\n")}},
	{fixture: "internal/oracle/testdata/node_inspector_unavailable_uncaught.a", name: "node-without-inspector",
		ruling:     "@system_adamic and @system_adamic_platforms, task #cn4yapb, Nov 5",
		reason:     "The no-inspector host raises on load; source Node has no CommonJS require in its ES module.",
		node:       &run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: " + requireInESModule + " in ES module scope, you can use import instead\n"), exitCode: 70},
		native:     run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: " + inspectorUnavailableString + "\n"), exitCode: 70},
		javascript: run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: " + inspectorUnavailableString + "\n"), exitCode: 70}},
	{
		fixture:    "internal/oracle/testdata/node_commonjs_require_missing.a",
		name:       "installed-layout",
		ruling:     "@system_adamic, task #redxes2, Nov 3",
		reason:     installedLayoutReason,
		node:       &run{stdout: []byte("ReferenceError none require is not defined\nReferenceError none require is not defined\nafter\n")},
		native:     run{stdout: []byte(installedLayoutName + " " + installedLayoutCode + " " + installedLayoutMessage + "\n" + installedLayoutName + " " + installedLayoutCode + " " + installedLayoutMessage + "\nafter\n")},
		javascript: run{stdout: []byte(installedLayoutName + " " + installedLayoutCode + " " + installedLayoutMessage + "\n" + installedLayoutName + " " + installedLayoutCode + " " + installedLayoutMessage + "\nafter\n")},
	},
	{
		fixture:    "internal/oracle/testdata/node_commonjs_require_missing_uncaught.a",
		name:       "installed-layout",
		ruling:     "@system_adamic, task #redxes2, Nov 3",
		reason:     installedLayoutReason,
		node:       &run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: " + requireInESModule + " in ES module scope, you can use import instead\n"), exitCode: 70},
		native:     run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: " + installedLayoutName + ": " + installedLayoutMessage + "\n"), exitCode: 70},
		javascript: run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: " + installedLayoutName + ": " + installedLayoutMessage + "\n"), exitCode: 70},
	},
	{
		fixture:    "internal/oracle/testdata/node_commonjs_require_computed.a",
		name:       "no-javascript-plugins",
		ruling:     "@system_adamic, task #redxes2, Nov 3",
		reason:     noJavaScriptPluginsReason,
		node:       &run{stdout: []byte("ReferenceError none require is not defined\npaths built: 0\n")},
		native:     run{stdout: []byte("Error none " + noJavaScriptPlugins("./plugin-1.js") + "\npaths built: 1\n")},
		javascript: run{stdout: []byte("Error none " + noJavaScriptPlugins("./plugin-1.js") + "\npaths built: 1\n")},
	},
	{
		fixture:    "internal/oracle/testdata/node_commonjs_require_computed_uncaught.a",
		name:       "no-javascript-plugins",
		ruling:     "@system_adamic, task #redxes2, Nov 3",
		reason:     noJavaScriptPluginsReason,
		node:       &run{stdout: []byte("before\ncaught: require is not defined\n"), stderr: []byte("adamic: panic: " + requireInESModule + " in ES module scope, you can use import instead\n"), exitCode: 70},
		native:     run{stdout: []byte("before\nresolving first\ncaught: " + noJavaScriptPlugins("./first.js") + "\nresolving plugin\n"), stderr: []byte("adamic: panic: Error: " + noJavaScriptPlugins("./plugin.js") + "\n"), exitCode: 70},
		javascript: run{stdout: []byte("before\nresolving first\ncaught: " + noJavaScriptPlugins("./first.js") + "\nresolving plugin\n"), stderr: []byte("adamic: panic: Error: " + noJavaScriptPlugins("./plugin.js") + "\n"), exitCode: 70},
	},
	// tsc's own forms, sys.ts:1597's cast and 1619's return, which the host audit's Modelled row lets
	// past the cast and any-relation rules (#3sh9398).
	{
		fixture:    "internal/oracle/testdata/node_commonjs_require_missing_cast.a",
		name:       "installed-layout",
		ruling:     "@system_adamic, task #redxes2, Nov 3",
		reason:     installedLayoutReason,
		node:       &run{stdout: []byte("ReferenceError none require is not defined\nafter\n")},
		native:     run{stdout: []byte(installedLayoutName + " " + installedLayoutCode + " " + installedLayoutMessage + "\nafter\n")},
		javascript: run{stdout: []byte(installedLayoutName + " " + installedLayoutCode + " " + installedLayoutMessage + "\nafter\n")},
	},
	{
		fixture:    "internal/oracle/testdata/node_commonjs_require_computed_return.a",
		name:       "no-javascript-plugins",
		ruling:     "@system_adamic, task #redxes2, Nov 3",
		reason:     noJavaScriptPluginsReason,
		node:       &run{stdout: []byte("failed: require is not defined\nafter\n")},
		native:     run{stdout: []byte("failed: " + noJavaScriptPlugins("./plugin.js") + "\nafter\n")},
		javascript: run{stdout: []byte("failed: " + noJavaScriptPlugins("./plugin.js") + "\nafter\n")},
	},
}

// The two rows' reasons, one per row, shared by its fixtures.
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

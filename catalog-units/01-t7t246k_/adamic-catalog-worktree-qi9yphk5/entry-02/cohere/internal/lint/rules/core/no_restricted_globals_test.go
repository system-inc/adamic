package core

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// noRestrictedGlobalsFile is where the fixtures pretend to live.
const noRestrictedGlobalsFile = "/repository/source/NoRestrictedGlobals.ts"

// noRestrictedGlobalsExpectation is one finding upstream asserts.
type noRestrictedGlobalsExpectation struct {
	id            string
	name          string
	customMessage string
}

// noRestrictedGlobalsCase is one upstream case.
//
// `optionsJson` is raw config text rather than a built struct, so every case is routed through the
// rule's own decoder. That is the only thing that puts the two wire shapes under test: the same
// case appears in the corpus four times over, as a bare string, as `{name}`, as
// `{globals:[string]}` and as `{globals:[{name}]}`, and a decoder handling only one of them would
// pass a quarter of the suite.
//
// `environmentGlobals` records what upstream put in `languageOptions.globals` for the case. It is
// recorded rather than acted on; see the harness note on the runner.
type noRestrictedGlobalsCase struct {
	source             string
	optionsJson        string
	environmentGlobals *string
	// ecmaVersion is what upstream configured, or 0 for unset. Recorded rather than applied; see
	// the runner for the one case whose verdict it decides.
	ecmaVersion int
	findings    []noRestrictedGlobalsExpectation
}

func noRestrictedGlobalsStringPointer(value string) *string { return &value }

// decodeNoRestrictedGlobalsOptionsForTest routes a case's options through the shipped decoder.
func decodeNoRestrictedGlobalsOptionsForTest(t *testing.T, optionsJson string) any {
	t.Helper()
	if optionsJson == "" {
		return nil
	}
	decoded, err := DecodeNoRestrictedGlobalsOptions(json.RawMessage(optionsJson))
	if err != nil {
		t.Fatalf("decoding options %q: %v", optionsJson, err)
	}
	return decoded
}

// Block 0, upstream's JavaScript RuleTester.
var noRestrictedGlobalsCleanCasesBlock0 = []noRestrictedGlobalsCase{
	{source: "foo", optionsJson: "", environmentGlobals: nil, ecmaVersion: 0},
	{source: "foo", optionsJson: "[\"bar\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "var foo = 1;", optionsJson: "[\"foo\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "event", optionsJson: "[\"bar\"]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0},
	{source: "import foo from 'bar';", optionsJson: "[\"foo\"]", environmentGlobals: nil, ecmaVersion: 6},
	{source: "function foo() {}", optionsJson: "[\"foo\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "function fn() { var foo; }", optionsJson: "[\"foo\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "foo.bar", optionsJson: "[\"bar\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "foo", optionsJson: "[{\"name\":\"bar\",\"message\":\"Use baz instead.\"}]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "foo", optionsJson: "[{\"globals\":[\"bar\"]}]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "const foo = 1", optionsJson: "[{\"globals\":[\"foo\"]}]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "event", optionsJson: "[{\"globals\":[\"bar\"]}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0},
	{source: "import foo from 'bar';", optionsJson: "[{\"globals\":[\"foo\"]}]", environmentGlobals: nil, ecmaVersion: 6},
	{source: "function foo() {}", optionsJson: "[{\"globals\":[\"foo\"]}]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "function fn() { let foo; }", optionsJson: "[{\"globals\":[\"foo\"]}]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "foo.bar", optionsJson: "[{\"globals\":[\"bar\"]}]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "foo", optionsJson: "[{\"globals\":[{\"name\":\"bar\",\"message\":\"Use baz instead.\"}]}]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "window.foo()", optionsJson: "[{\"globals\":[\"foo\"]}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0},
	{source: "self.foo()", optionsJson: "[{\"globals\":[\"foo\"]}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0},
	{source: "globalThis.foo()", optionsJson: "[{\"globals\":[\"foo\"]}]", environmentGlobals: nil, ecmaVersion: 2020},
	{source: "myGlobal.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"globalObjects\":[\"myGlobal\"]}]", environmentGlobals: noRestrictedGlobalsStringPointer("myGlobal"), ecmaVersion: 0},
	{source: "window.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true}]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "self.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true}]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "globalThis.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true}]", environmentGlobals: nil, ecmaVersion: 6},
	{source: "myGlobal.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true,\"globalObjects\":[\"myGlobal\"]}]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "otherGlobal.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true,\"globalObjects\":[\"myGlobal\"]}]", environmentGlobals: noRestrictedGlobalsStringPointer("otherGlobal"), ecmaVersion: 0},
	{source: "foo.window.bar()", optionsJson: "[{\"globals\":[\"bar\"],\"checkGlobalObject\":true}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0},
	{source: "foo.self.bar()", optionsJson: "[{\"globals\":[\"bar\"],\"checkGlobalObject\":true}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0},
	{source: "foo.globalThis.bar()", optionsJson: "[{\"globals\":[\"bar\"],\"checkGlobalObject\":true}]", environmentGlobals: nil, ecmaVersion: 2020},
	{source: "foo.myGlobal.bar()", optionsJson: "[{\"globals\":[\"bar\"],\"checkGlobalObject\":true,\"globalObjects\":[\"myGlobal\"]}]", environmentGlobals: noRestrictedGlobalsStringPointer("myGlobal"), ecmaVersion: 0},
	{source: "let window; window.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0},
	{source: "let self; self.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0},
	{source: "let globalThis; globalThis.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true}]", environmentGlobals: nil, ecmaVersion: 2020},
	{source: "let myGlobal; myGlobal.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true,\"globalObjects\":[\"myGlobal\"]}]", environmentGlobals: noRestrictedGlobalsStringPointer("myGlobal"), ecmaVersion: 0},
}

var noRestrictedGlobalsFiringCasesBlock0 = []noRestrictedGlobalsCase{
	{source: "foo", optionsJson: "[\"foo\"]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "function fn() { foo; }", optionsJson: "[\"foo\"]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "function fn() { foo; }", optionsJson: "[\"foo\"]", environmentGlobals: noRestrictedGlobalsStringPointer("foo"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "event", optionsJson: "[\"foo\",\"event\"]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "event", customMessage: ""}}},
	{source: "foo", optionsJson: "[\"foo\"]", environmentGlobals: noRestrictedGlobalsStringPointer("foo"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "foo()", optionsJson: "[\"foo\"]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "foo.bar()", optionsJson: "[\"foo\"]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "foo", optionsJson: "[{\"name\":\"foo\"}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "function fn() { foo; }", optionsJson: "[{\"name\":\"foo\"}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "function fn() { foo; }", optionsJson: "[{\"name\":\"foo\"}]", environmentGlobals: noRestrictedGlobalsStringPointer("foo"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "event", optionsJson: "[\"foo\",{\"name\":\"event\"}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "event", customMessage: ""}}},
	{source: "foo", optionsJson: "[{\"name\":\"foo\"}]", environmentGlobals: noRestrictedGlobalsStringPointer("foo"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "foo()", optionsJson: "[{\"name\":\"foo\"}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "foo.bar()", optionsJson: "[{\"name\":\"foo\"}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "foo", optionsJson: "[{\"name\":\"foo\",\"message\":\"Use bar instead.\"}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "customMessage", name: "foo", customMessage: "Use bar instead."}}},
	{source: "function fn() { foo; }", optionsJson: "[{\"name\":\"foo\",\"message\":\"Use bar instead.\"}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "customMessage", name: "foo", customMessage: "Use bar instead."}}},
	{source: "function fn() { foo; }", optionsJson: "[{\"name\":\"foo\",\"message\":\"Use bar instead.\"}]", environmentGlobals: noRestrictedGlobalsStringPointer("foo"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "customMessage", name: "foo", customMessage: "Use bar instead."}}},
	{source: "event", optionsJson: "[\"foo\",{\"name\":\"event\",\"message\":\"Use local event parameter.\"}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "customMessage", name: "event", customMessage: "Use local event parameter."}}},
	{source: "foo", optionsJson: "[{\"name\":\"foo\",\"message\":\"Use bar instead.\"}]", environmentGlobals: noRestrictedGlobalsStringPointer("foo"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "customMessage", name: "foo", customMessage: "Use bar instead."}}},
	{source: "foo()", optionsJson: "[{\"name\":\"foo\",\"message\":\"Use bar instead.\"}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "customMessage", name: "foo", customMessage: "Use bar instead."}}},
	{source: "foo.bar()", optionsJson: "[{\"name\":\"foo\",\"message\":\"Use bar instead.\"}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "customMessage", name: "foo", customMessage: "Use bar instead."}}},
	{source: "var foo = obj => hasOwnProperty(obj, 'name');", optionsJson: "[\"hasOwnProperty\"]", environmentGlobals: nil, ecmaVersion: 6, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "hasOwnProperty", customMessage: ""}}},
	{source: "foo", optionsJson: "[{\"globals\":[\"foo\"]}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "function fn() { foo; }", optionsJson: "[{\"globals\":[\"foo\"]}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "function fn() { foo; }", optionsJson: "[{\"globals\":[\"foo\"]}]", environmentGlobals: noRestrictedGlobalsStringPointer("foo"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "event", optionsJson: "[{\"globals\":[\"foo\",\"event\"]}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "event", customMessage: ""}}},
	{source: "foo", optionsJson: "[{\"globals\":[\"foo\"]}]", environmentGlobals: noRestrictedGlobalsStringPointer("foo"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "foo()", optionsJson: "[{\"globals\":[\"foo\"]}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "foo.bar()", optionsJson: "[{\"globals\":[\"foo\"]}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "foo", optionsJson: "[{\"globals\":[{\"name\":\"foo\"}]}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "function fn() { foo; }", optionsJson: "[{\"globals\":[{\"name\":\"foo\"}]}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "function fn() { foo; }", optionsJson: "[{\"globals\":[{\"name\":\"foo\"}]}]", environmentGlobals: noRestrictedGlobalsStringPointer("foo"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "event", optionsJson: "[{\"globals\":[\"foo\",{\"name\":\"event\"}]}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "event", customMessage: ""}}},
	{source: "foo", optionsJson: "[{\"globals\":[{\"name\":\"foo\"}]}]", environmentGlobals: noRestrictedGlobalsStringPointer("foo"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "foo()", optionsJson: "[{\"globals\":[{\"name\":\"foo\"}]}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "foo.bar()", optionsJson: "[{\"globals\":[{\"name\":\"foo\"}]}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "foo", optionsJson: "[{\"globals\":[{\"name\":\"foo\",\"message\":\"Use bar instead.\"}]}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "customMessage", name: "foo", customMessage: "Use bar instead."}}},
	{source: "function fn() { foo; }", optionsJson: "[{\"globals\":[{\"name\":\"foo\",\"message\":\"Use bar instead.\"}]}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "customMessage", name: "foo", customMessage: "Use bar instead."}}},
	{source: "function fn() { foo; }", optionsJson: "[{\"globals\":[{\"name\":\"foo\",\"message\":\"Use bar instead.\"}]}]", environmentGlobals: noRestrictedGlobalsStringPointer("foo"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "customMessage", name: "foo", customMessage: "Use bar instead."}}},
	{source: "event", optionsJson: "[{\"globals\":[\"foo\",{\"name\":\"event\",\"message\":\"Use local event parameter.\"}]}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "customMessage", name: "event", customMessage: "Use local event parameter."}}},
	{source: "foo", optionsJson: "[{\"globals\":[{\"name\":\"foo\",\"message\":\"Use bar instead.\"}]}]", environmentGlobals: noRestrictedGlobalsStringPointer("foo"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "customMessage", name: "foo", customMessage: "Use bar instead."}}},
	{source: "foo()", optionsJson: "[{\"globals\":[{\"name\":\"foo\",\"message\":\"Use bar instead.\"}]}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "customMessage", name: "foo", customMessage: "Use bar instead."}}},
	{source: "foo.bar()", optionsJson: "[{\"globals\":[{\"name\":\"foo\",\"message\":\"Use bar instead.\"}]}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "customMessage", name: "foo", customMessage: "Use bar instead."}}},
	{source: "var foo = obj => hasOwnProperty(obj, 'name');", optionsJson: "[{\"globals\":[\"hasOwnProperty\"]}]", environmentGlobals: nil, ecmaVersion: 6, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "hasOwnProperty", customMessage: ""}}},
	{source: "window.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "self.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "window.window.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "self.self.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "globalThis.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true}]", environmentGlobals: nil, ecmaVersion: 2020, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "globalThis.globalThis.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true}]", environmentGlobals: nil, ecmaVersion: 2020, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "myGlobal.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true,\"globalObjects\":[\"myGlobal\"]}]", environmentGlobals: noRestrictedGlobalsStringPointer("myGlobal"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "myGlobal.myGlobal.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true,\"globalObjects\":[\"myGlobal\"]}]", environmentGlobals: noRestrictedGlobalsStringPointer("myGlobal"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "window[\"foo\"]", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "self[\"foo\"]", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "globalThis[\"foo\"]", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true}]", environmentGlobals: nil, ecmaVersion: 2020, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "myGlobal[\"foo\"]", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true,\"globalObjects\":[\"myGlobal\"]}]", environmentGlobals: noRestrictedGlobalsStringPointer("myGlobal"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "window?.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "self?.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "window.foo(); myGlobal.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true,\"globalObjects\":[\"myGlobal\"]}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event,myGlobal"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}, {id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "myGlobal.foo(); myOtherGlobal.bar()", optionsJson: "[{\"globals\":[\"foo\",\"bar\"],\"checkGlobalObject\":true,\"globalObjects\":[\"myGlobal\",\"myOtherGlobal\"]}]", environmentGlobals: noRestrictedGlobalsStringPointer("myGlobal,myOtherGlobal"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}, {id: "defaultMessage", name: "bar", customMessage: ""}}},
	{source: "foo(); window.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}, {id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "foo(); self.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}, {id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "foo(); myGlobal.foo()", optionsJson: "[{\"globals\":[\"foo\"],\"checkGlobalObject\":true,\"globalObjects\":[\"myGlobal\"]}]", environmentGlobals: noRestrictedGlobalsStringPointer("myGlobal"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}, {id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "function onClick(event) { console.log(event); console.log(window.event); }", optionsJson: "[{\"globals\":[\"event\"],\"checkGlobalObject\":true}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "event", customMessage: ""}}},
	{source: "function onClick(event) { console.log(event); console.log(self.event); }", optionsJson: "[{\"globals\":[\"event\"],\"checkGlobalObject\":true}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "event", customMessage: ""}}},
	{source: "function onClick(event) { console.log(event); console.log(globalThis.event); }", optionsJson: "[{\"globals\":[\"event\"],\"checkGlobalObject\":true}]", environmentGlobals: nil, ecmaVersion: 2020, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "event", customMessage: ""}}},
	{source: "function onClick(event) { console.log(event); console.log(myGlobal.event); }", optionsJson: "[{\"globals\":[\"event\"],\"checkGlobalObject\":true,\"globalObjects\":[\"myGlobal\"]}]", environmentGlobals: noRestrictedGlobalsStringPointer("myGlobal"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "event", customMessage: ""}}},
}

// Block 1, upstream's TypeScript RuleTester.
var noRestrictedGlobalsCleanCasesBlock1 = []noRestrictedGlobalsCase{
	{source: "foo", optionsJson: "", environmentGlobals: nil, ecmaVersion: 0},
	{source: "foo", optionsJson: "[\"bar\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "const foo: number = 1;", optionsJson: "[\"foo\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "event", optionsJson: "[\"bar\"]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0},
	{source: "import foo from 'bar';", optionsJson: "[\"foo\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "function foo(): void {}", optionsJson: "[\"foo\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "function fn(): void { let foo; }", optionsJson: "[\"foo\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "foo.bar", optionsJson: "[\"bar\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "foo", optionsJson: "[{\"name\":\"bar\",\"message\":\"Use baz instead.\"}]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "\n\t\t\texport default class Test {\n\t\t\t\tprivate status: string;\n\t\t\t\tgetStatus() {\n\t\t\t\t\treturn this.status;\n\t\t\t\t}\n\t\t\t}", optionsJson: "[\"status\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "type Handler = (event: string) => any", optionsJson: "[\"event\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "let value: bigint", optionsJson: "[\"bigint\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "let value: boolean", optionsJson: "[\"boolean\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "let value: never", optionsJson: "[\"never\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "let value: null", optionsJson: "[\"null\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "let value: number", optionsJson: "[\"number\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "let value: object", optionsJson: "[\"object\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "let value: string", optionsJson: "[\"string\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "let value: symbol", optionsJson: "[\"symbol\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "let value: undefined", optionsJson: "[\"undefined\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "let value: unknown", optionsJson: "[\"unknown\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "let value: void", optionsJson: "[\"void\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "let value: []", optionsJson: "[\"[]\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "let value: {}", optionsJson: "[\"{}\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "let value: Test", optionsJson: "[\"Test\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "let value: Test[]", optionsJson: "[\"Test\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "let value: [Test]", optionsJson: "[\"Test\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "let b: { c: Test }", optionsJson: "[\"Test\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "function foo(param: Test) {}", optionsJson: "[\"Test\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "1 as Test", optionsJson: "[\"Test\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "class Derived implements Test {}", optionsJson: "[\"Test\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "class Derived implements Test1, Test2 {}", optionsJson: "[\"Test1\",\"Test2\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "interface Derived extends Test {}", optionsJson: "[\"Test\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "type Intersection = Test & {}", optionsJson: "[\"Test\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "type Union = Test | {}", optionsJson: "[\"Test\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "let value: NS.Test", optionsJson: "[\"NS\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "let value: NS.Test", optionsJson: "[\"Test\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "let value: NS.Test", optionsJson: "[\"NS.Test\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "let value: typeof Test", optionsJson: "[\"Test\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "let value: Type<Test>", optionsJson: "[\"Type\",\"Test\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "type Intersection = Test<any>", optionsJson: "[\"Test\",\"any\"]", environmentGlobals: nil, ecmaVersion: 0},
	{source: "type Intersection = Test<A, B>", optionsJson: "[\"Test\",\"A\",\"B\"]", environmentGlobals: nil, ecmaVersion: 0},
}

var noRestrictedGlobalsFiringCasesBlock1 = []noRestrictedGlobalsCase{
	{source: "foo", optionsJson: "[\"foo\"]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "function fn(): void { foo; }", optionsJson: "[\"foo\"]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "function fn(): void { foo; }", optionsJson: "[\"foo\"]", environmentGlobals: noRestrictedGlobalsStringPointer("foo"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "event", optionsJson: "[\"foo\",\"event\"]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "event", customMessage: ""}}},
	{source: "foo", optionsJson: "[\"foo\"]", environmentGlobals: noRestrictedGlobalsStringPointer("foo"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "foo()", optionsJson: "[\"foo\"]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "foo.bar()", optionsJson: "[\"foo\"]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "foo", optionsJson: "[{\"name\":\"foo\"}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "function fn(): void { foo; }", optionsJson: "[{\"name\":\"foo\"}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "function fn(): void { foo; }", optionsJson: "[{\"name\":\"foo\"}]", environmentGlobals: noRestrictedGlobalsStringPointer("foo"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "event", optionsJson: "[\"foo\",{\"name\":\"event\"}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "event", customMessage: ""}}},
	{source: "foo", optionsJson: "[{\"name\":\"foo\"}]", environmentGlobals: noRestrictedGlobalsStringPointer("foo"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "foo()", optionsJson: "[{\"name\":\"foo\"}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "foo.bar()", optionsJson: "[{\"name\":\"foo\"}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "foo", customMessage: ""}}},
	{source: "foo", optionsJson: "[{\"name\":\"foo\",\"message\":\"Use bar instead.\"}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "customMessage", name: "foo", customMessage: "Use bar instead."}}},
	{source: "function fn(): void { foo; }", optionsJson: "[{\"name\":\"foo\",\"message\":\"Use bar instead.\"}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "customMessage", name: "foo", customMessage: "Use bar instead."}}},
	{source: "function fn(): void { foo; }", optionsJson: "[{\"name\":\"foo\",\"message\":\"Use bar instead.\"}]", environmentGlobals: noRestrictedGlobalsStringPointer("foo"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "customMessage", name: "foo", customMessage: "Use bar instead."}}},
	{source: "event", optionsJson: "[\"foo\",{\"name\":\"event\",\"message\":\"Use local event parameter.\"}]", environmentGlobals: noRestrictedGlobalsStringPointer("window,self,event"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "customMessage", name: "event", customMessage: "Use local event parameter."}}},
	{source: "foo", optionsJson: "[{\"name\":\"foo\",\"message\":\"Use bar instead.\"}]", environmentGlobals: noRestrictedGlobalsStringPointer("foo"), ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "customMessage", name: "foo", customMessage: "Use bar instead."}}},
	{source: "foo()", optionsJson: "[{\"name\":\"foo\",\"message\":\"Use bar instead.\"}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "customMessage", name: "foo", customMessage: "Use bar instead."}}},
	{source: "foo.bar()", optionsJson: "[{\"name\":\"foo\",\"message\":\"Use bar instead.\"}]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "customMessage", name: "foo", customMessage: "Use bar instead."}}},
	{source: "const foo = obj => hasOwnProperty(obj, 'name');", optionsJson: "[\"hasOwnProperty\"]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "hasOwnProperty", customMessage: ""}}},
	{source: "const x: Promise<any> = Promise.resolve();", optionsJson: "[\"Promise\"]", environmentGlobals: nil, ecmaVersion: 0, findings: []noRestrictedGlobalsExpectation{{id: "defaultMessage", name: "Promise", customMessage: ""}}},
}

// runNoRestrictedGlobalsCase drives one case and asserts every finding's id and rendered message.
//
// # How upstream's `languageOptions.globals` is expressed here
//
// Upstream declares environment globals through a config knob. We have no such knob: our program is
// built from a tsconfig, and the fixture harness pins `lib: ["ES2022"]`, so `window`, `self` and
// `event` resolve to no symbol at all. Rather than weaken the rule to make those cases green, each
// case carries what upstream configured and the runner writes an equivalent AMBIENT DECLARATION
// FILE beside the fixture. That is how a real tree declares these names, so the rule takes the same
// path it takes on a repository rather than one arranged for the test.
//
// The distinction is load bearing rather than cosmetic, and the corpus is what proves it. Under
// `checkGlobalObject`, `window.foo()` REPORTS when the browser set is declared and is CLEAN when it
// is not, while `globalThis.foo()` reports either way. Measured, our checker draws that same
// three-way line by itself: `globalThis` resolves to a symbol carrying zero declarations, `window`
// resolves to nothing without a declaration and to a declaration-file symbol with one. So the
// ambient file reproduces the environment sensitivity upstream's knob was expressing, rather than
// papering over it.
//
// Only four globals are ever named explicitly across the corpus, and a case marked `browser` needs
// the handful the cases actually touch.
func runNoRestrictedGlobalsCase(t *testing.T, testCase noRestrictedGlobalsCase) {
	t.Helper()

	// A case whose verdict is decided ABOVE the rule, by the language version rather than by the
	// rule's own judgment. `globalThis.foo()` appears in upstream's valid AND invalid lists with
	// byte-identical options, separated only by `ecmaVersion`: the name did not exist before
	// ES2020, so eslint's scope does not know it under ecmaVersion 6 and the input is clean there.
	//
	// Our program is built at ES2022 and the checker knows `globalThis` unconditionally, so the
	// pre-ES2020 half cannot be expressed. Skipped with the reason stated rather than greened by
	// weakening the rule, which would also have silenced the ES2020 half that upstream reports.
	if testCase.ecmaVersion != 0 && testCase.ecmaVersion < 2020 &&
		strings.Contains(testCase.source, "globalThis") {
		t.Skipf("upstream runs this at ecmaVersion %d, where globalThis is not a known name; "+
			"our program is ES2022 and the checker knows it unconditionally", testCase.ecmaVersion)
	}

	files := map[string]string{"Probe.ts": testCase.source}
	if declarations := noRestrictedGlobalsAmbientDeclarations(testCase.environmentGlobals); declarations != "" {
		files["EnvironmentGlobals.d.ts"] = declarations
	}

	result := rule_testing.RunTypedFilesWithOptions(t, NoRestrictedGlobals, files, "Probe.ts",
		decodeNoRestrictedGlobalsOptionsForTest(t, testCase.optionsJson))

	wantIds := make([]string, len(testCase.findings))
	for i, finding := range testCase.findings {
		wantIds[i] = finding.id
	}
	rule_testing.ExpectFindings(t, result, wantIds...)

	for i, finding := range testCase.findings {
		// Upstream's two templates: `Unexpected use of '{{name}}'.` and
		// `Unexpected use of '{{name}}'. {{customMessage}}`. Asserted as an exact prefix rather
		// than with a substring test, because the custom message is interpolated and a
		// `strings.Contains` cannot see a doubled or misplaced interpolation.
		want := "Unexpected use of '" + finding.name + "'."
		if finding.customMessage != "" {
			want += " " + finding.customMessage
		}
		if got := result.Diagnostics[i].Message.Description; !strings.HasPrefix(got, want) {
			t.Errorf("finding %d: message should begin %q, got %q", i, want, got)
		}
	}
}

// noRestrictedGlobalsAmbientDeclarations renders one case's environment as a declaration file.
//
// The field holds the names upstream declared, comma separated. Upstream's browser set is 1164
// names of which these cases only ever read `window`, `self` and `event`, so the extractor records
// those three rather than the whole list, plus any name added on top of the set. That last part is
// load bearing: one case configures the browser set AND `myGlobal`, and recording it as a bare
// "browser" flag dropped the extra and made the case report once where upstream reports twice.
//
// `globalThis` is deliberately never emitted. The checker knows it without a declaration, which is
// why upstream's `globalThis.foo()` reports with no globals configured while its `window.foo()`
// twin is clean.
func noRestrictedGlobalsAmbientDeclarations(environmentGlobals *string) string {
	if environmentGlobals == nil || *environmentGlobals == "" {
		return ""
	}
	var builder strings.Builder
	for _, name := range strings.Split(*environmentGlobals, ",") {
		if name == "" {
			continue
		}
		builder.WriteString("declare var " + name + ": any;\n")
	}
	return builder.String()
}

// TestNoRestrictedGlobalsFires runs both of upstream's invalid blocks.
func TestNoRestrictedGlobalsFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range noRestrictedGlobalsFiringCasesBlock0 {
		t.Run("javascript/"+testCase.source, func(t *testing.T) {
			t.Parallel()
			runNoRestrictedGlobalsCase(t, testCase)
		})
	}
	for _, testCase := range noRestrictedGlobalsFiringCasesBlock1 {
		t.Run("typescript/"+testCase.source, func(t *testing.T) {
			t.Parallel()
			runNoRestrictedGlobalsCase(t, testCase)
		})
	}
}

// TestNoRestrictedGlobalsStaysSilent runs both of upstream's valid blocks.
//
// The TypeScript block is the more valuable of the two here: 32 of its 42 cases are type positions,
// which is the exclusion a port is most likely to get wrong, and several of them differ only in
// how deeply the type is nested.
func TestNoRestrictedGlobalsStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range noRestrictedGlobalsCleanCasesBlock0 {
		t.Run("javascript/"+testCase.source, func(t *testing.T) {
			t.Parallel()
			runNoRestrictedGlobalsCase(t, testCase)
		})
	}
	for _, testCase := range noRestrictedGlobalsCleanCasesBlock1 {
		t.Run("typescript/"+testCase.source, func(t *testing.T) {
			t.Parallel()
			runNoRestrictedGlobalsCase(t, testCase)
		})
	}
}

// TestNoRestrictedGlobalsSpan asserts where each finding points.
//
// Upstream reports two different nodes depending on which half of the rule fired, and every
// message-id fixture here stays green over the wrong choice. A mutation reporting the whole member
// access instead of the property survived all 165 corpus cases, which is the fixture set asserting
// which rule fired and never where.
//
//	a bare reference          the identifier itself
//	window.foo                the PROPERTY, `foo`, not `window.foo` and not `window`
//	window["foo"]             the string literal, again the property half
func TestNoRestrictedGlobalsSpan(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name        string
		source      string
		optionsJson string
		globals     string
		// wantText is the source the finding's own range covers.
		wantText string
	}{
		{
			name:        "a bare reference points at the identifier",
			source:      "foo;",
			optionsJson: `["foo"]`,
			wantText:    "foo",
		},
		{
			name:        "a dotted global access points at the property",
			source:      "window.foo();",
			optionsJson: `[{"globals":["foo"],"checkGlobalObject":true}]`,
			globals:     "window",
			wantText:    "foo",
		},
		{
			name:        "a subscripted global access points at the string literal",
			source:      `window["foo"];`,
			optionsJson: `[{"globals":["foo"],"checkGlobalObject":true}]`,
			globals:     "window",
			wantText:    `"foo"`,
		},
		{
			name:        "a doubled global object still points at the final property",
			source:      "window.window.foo();",
			optionsJson: `[{"globals":["foo"],"checkGlobalObject":true}]`,
			globals:     "window",
			wantText:    "foo",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			files := map[string]string{"Probe.ts": testCase.source}
			if testCase.globals != "" {
				files["EnvironmentGlobals.d.ts"] = "declare var " + testCase.globals + ": any;\n"
			}
			result := rule_testing.RunTypedFilesWithOptions(t, NoRestrictedGlobals, files, "Probe.ts",
				decodeNoRestrictedGlobalsOptionsForTest(t, testCase.optionsJson))
			rule_testing.ExpectFindings(t, result, "defaultMessage")

			// `RunTypedFiles` writes each fixture as TrimSpace(contents)+"\n", so the literal above
			// is one byte offset from the file on disk unless it is transformed the same way.
			// Transformed rather than sliced from the literal, which is the off-by-one this brief
			// warns about and which reads exactly like a rule defect.
			onDisk := strings.TrimSpace(testCase.source) + "\n"
			reported := onDisk[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.wantText {
				t.Errorf("finding should point at %q, got %q", testCase.wantText, reported)
			}
		})
	}
}

// TestNoRestrictedGlobalsIsInertWithoutConfiguration pins the unconfigured path.
//
// This rule reports nothing until somebody names a global, which is upstream's design: `create`
// returns an empty visitor when the list is empty. A rule configured as a bare severity is handed
// nil options rather than a struct, so this asserts the path no case in the corpus takes and the
// one the live config will take if the rule is enabled without a list.
//
// The early return itself is an OPTIMISATION rather than a discrimination, and the mutation
// neutralising it survives for that reason: with an empty map the identifier listener looks every
// name up, finds nothing restricted, and reports nothing anyway. No input can distinguish the two,
// because distinguishing them would need a name that is restricted while the map is empty. Kept
// because registering no listener at all is measurably cheaper on a tree of this size than one
// map lookup per identifier in every file.
func TestNoRestrictedGlobalsIsInertWithoutConfiguration(t *testing.T) {
	t.Parallel()

	for _, source := range []string{"foo;", "window.foo();", "event;"} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoRestrictedGlobals,
				noRestrictedGlobalsFile, source, nil))
		})
	}
}

// TestNoRestrictedGlobalsRequiresTheTypedHarness asserts the rule declines a nil checker.
//
// A typed rule handed the plain harness goes silent rather than crashing, which makes every
// StaysSilent case pass vacuously and every Fires case look like a rule defect. Asserted so a later
// revert to `Run` fails loudly here rather than quietly everywhere.
func TestNoRestrictedGlobalsRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	decoded := decodeNoRestrictedGlobalsOptionsForTest(t, `["foo"]`)

	rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, NoRestrictedGlobals,
		noRestrictedGlobalsFile, "foo;", decoded))

	// The control: the same input through the typed harness does report, so the silence above is
	// the nil-checker guard rather than a rule that cannot fire.
	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, NoRestrictedGlobals,
		noRestrictedGlobalsFile, "foo;", decoded), "defaultMessage")
}

// TestDecodeNoRestrictedGlobalsOptions pins the two list shapes the config layer delivers and the
// ones it must refuse. Upstream's variadic spelling, `["error", "event", "fdescribe"]`, used to
// reach this decoder as the bare string "event" and fail, while `["error", ["event", "fdescribe"]]`
// was the workaround; now the list arrives whole and every element is a global.
func TestDecodeNoRestrictedGlobalsOptions(t *testing.T) {
	t.Parallel()

	decoded, err := DecodeNoRestrictedGlobalsOptions([]byte(
		`["event", {"name": "fdescribe", "message": "Do not commit fdescribe."}]`))
	if err != nil {
		t.Fatalf("upstream's variadic spelling was refused: %v", err)
	}
	settings := decoded.(NoRestrictedGlobalsSettings)
	if len(settings.GlobalsOrder) != 2 || settings.Globals["fdescribe"] != "Do not commit fdescribe." {
		t.Errorf("both elements must be read, the second with its message: %+v", settings)
	}

	decoded, err = DecodeNoRestrictedGlobalsOptions([]byte(
		`[{"globals": ["event"], "checkGlobalObject": true, "globalObjects": ["myGlobal"]}]`))
	if err != nil {
		t.Fatalf("the object form was refused: %v", err)
	}
	settings = decoded.(NoRestrictedGlobalsSettings)
	if !settings.CheckGlobalObject || len(settings.GlobalObjects) != 1 || len(settings.GlobalsOrder) != 1 {
		t.Errorf("the object form's flags must be read: %+v", settings)
	}

	for _, raw := range []string{
		// The bare string the config layer used to deliver.
		`"event"`,
		// The nested workaround: an element that is neither a string nor a {name} object.
		`[["event", "fdescribe"]]`,
		// The object form takes no second element.
		`[{"globals": ["event"]}, "fdescribe"]`,
		// Misspelled keys, which upstream's additionalProperties:false refuses.
		`[{"globals": ["event"], "checkGlobalObjects": true}]`,
		`[{"name": "event", "mesage": "x"}]`,
	} {
		if decoded, err := DecodeNoRestrictedGlobalsOptions([]byte(raw)); err == nil {
			t.Errorf("%s decoded to %+v; it must be refused", raw, decoded)
		}
	}
}

// TestNoRestrictedGlobalsJudgesOnlyValueReferences pins the names that spell a restricted global
// without reading it (#g5b8q7e), found when the confusing-browser-globals list was run over ahra.
//
// An intrinsic JSX tag is named by HTML, an attribute by the component, and a destructuring or
// object-literal key by the object's type. None reads the global. They reported because the key's
// symbol, when the object's type lives in a declaration file, is that file's property: `open:` in
// Collapsible.tsx:43 destructures Radix's props, declared in its `.d.ts`, so the shadow check saw a
// name not declared in source. `length`, declared in the standard library, reproduces it in one line.
//
// The firing rows are the controls: a component tag and a shorthand both read a binding, and
// WisdomGateItems.ts:304 before #war4qsy read a bare `status`, which is `window.status` and rendered
// empty every time.
func TestNoRestrictedGlobalsJudgesOnlyValueReferences(t *testing.T) {
	t.Parallel()

	const file = "/repository/source/NoRestrictedGlobals.tsx"
	const confusing = `["status", "name", "open", "stop", "length", "event", "Option"]`
	silent := []struct{ name, source string }{
		{"an intrinsic tag, the svg gradient stop", "export const Gradient = () => <linearGradient><stop offset=\"0\" /><stop offset=\"1\"></stop></linearGradient>;\n"},
		{"an attribute name", "declare const Dialog: (properties: { open: boolean }) => null;\nexport const Shown = () => <Dialog open={true} />;\n"},
		{"a destructuring key over a standard-library type", "declare const text: string;\nconst { length: size } = text;\nexport { size };\n"},
		{"an object-literal key typed by a standard-library type", "export const descriptor: PropertyDescriptor & { length?: number } = { length: 1 };\n"},
		{"an object-literal method typed by a standard-library type", "export const lengthy: { length(): number } = { length() { return 1; } };\n"},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, NoRestrictedGlobals, file,
				testCase.source, decodeNoRestrictedGlobalsOptionsForTest(t, confusing)))
		})
	}

	firing := []struct {
		name, source string
		ids          []string
	}{
		{"WisdomGateItems.ts:304 before #war4qsy, a bare status", "declare const intent: { task: { routedTo: string | null } };\nexport const text = `${status.toLowerCase()}${intent.task.routedTo ? ` · @${intent.task.routedTo}` : ''}`;\n", []string{"defaultMessage"}},
		{"a component tag reads its binding", "export const Shown = () => <Option />;\n", []string{"defaultMessage"}},
		{"a shorthand reads the global", "export const values = { status };\n", []string{"defaultMessage"}},
		{"a destructuring default reads the global", "declare const source: { label?: string };\nconst { label = name } = source;\nexport { label };\n", []string{"defaultMessage"}},
	}
	for _, testCase := range firing {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, NoRestrictedGlobals, file,
				testCase.source, decodeNoRestrictedGlobalsOptionsForTest(t, confusing)), testCase.ids...)
		})
	}
}

// TestNoRestrictedGlobalsSkipsAnotherModulesExportName pins the specifier half (#wajqfd1).
//
// `export { print as printGraphQlNode } from 'graphql'` in Base's GraphQlLibrary.ts:53 reported once
// `print` joined the confusing-browser-globals list. The `print` there names graphql's export, not a
// binding in this file, and it resolves to graphql's declaration file, so the shadow check saw a
// name not declared in source. An import specifier's imported name is the same shape.
//
// Every row was measured on ESLint 10 with @typescript-eslint/parser restricting `print`: the
// export-from and import rows are clean, and the firing rows report at the `print` they name. A
// local `export { print }` with no `from` exports whatever `print` resolves to here, which is the
// global, so it is a read and reports.
func TestNoRestrictedGlobalsSkipsAnotherModulesExportName(t *testing.T) {
	t.Parallel()

	const graphql = "export declare function print(): string;\nexport declare const other: number;\n"
	run := func(t *testing.T, source string) rule_testing.Result {
		t.Helper()
		return rule_testing.RunTypedFilesWithOptions(t, NoRestrictedGlobals, map[string]string{
			"graphql.d.ts": graphql,
			"a.ts":         source,
		}, "a.ts", decodeNoRestrictedGlobalsOptionsForTest(t, `["print"]`))
	}

	silent := []struct{ name, source string }{
		{"GraphQlLibrary.ts:53, an aliased export-from", "export { print as printGraphQlNode } from './graphql';\n"},
		{"a plain export-from", "export { print } from './graphql';\n"},
		{"an export-from renamed to the restricted name", "export { other as print } from './graphql';\n"},
		{"an aliased import", "import { print as printNode } from './graphql';\nexport const printed = printNode();\n"},
		{"a local export renamed to the restricted name", "const other = 1;\nexport { other as print };\n"},
		{"a global exported under the restricted name", "export { undeclaredThing as print };\n"},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, run(t, testCase.source))
		})
	}

	firing := []struct{ name, source string }{
		{"a bare call", "print();\nexport {};\n"},
		{"a local export of the global", "export { print };\n"},
		{"a local export of the global, aliased", "export { print as printPage };\n"},
		{"a local export of the global under its own name, reported once", "export { print as print };\n"},
	}
	for _, testCase := range firing {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, run(t, testCase.source), "defaultMessage")
		})
	}
}

// TestNoRestrictedGlobalsSkipsDeclarationNamesInADeclarationFile pins a name being introduced.
//
// A declaration's own name is never a read, and in a source file the shadow check declines it,
// since the name's symbol is declared right there. In a declaration file that check sees only a
// declaration file and declines nothing, so Base's photon_rs_bg.d.ts:41 reported the parameter
// `top` of a declared function once `top` was restricted. ESLint 10 under @typescript-eslint/parser
// is clean on each row below, restricting `top`.
//
// The firing row is the control: a value read in a declaration file is still a read.
func TestNoRestrictedGlobalsSkipsDeclarationNamesInADeclarationFile(t *testing.T) {
	t.Parallel()

	run := func(t *testing.T, source string) rule_testing.Result {
		t.Helper()
		return rule_testing.RunTypedFilesWithOptions(t, NoRestrictedGlobals, map[string]string{
			"photon.d.ts": source,
		}, "photon.d.ts", decodeNoRestrictedGlobalsOptionsForTest(t, `["top"]`))
	}

	silent := []struct{ name, source string }{
		{"photon_rs_bg.d.ts:41, a parameter of a declared function", "export function crop(left: number, top: number): void;\n"},
		{"a declared constant", "export declare const top: number;\n"},
		{"an interface member", "export interface Box { top: number }\n"},
		{"a type query is a type position", "export declare const edge: typeof top;\n"},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, run(t, testCase.source))
		})
	}

	rule_testing.ExpectFindings(t, run(t, "export declare const edge: number;\nexport default top;\n"), "defaultMessage")
}

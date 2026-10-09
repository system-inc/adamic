package core

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// yodaFile is where the fixtures pretend to live.
const yodaFile = "/repository/source/Yoda.ts"

// yodaExpectation is one finding upstream asserts, with the two values it interpolates.
type yodaExpectation struct {
	expectedSide string
	operator     string
}

// yodaCase is one upstream case.
//
// `optionsJson` is raw config text rather than a built struct, so every case is routed through the
// rule's own decoder. That matters more here than usual: upstream's option is a two element list
// whose second element carries both flags, and a fixture built from a struct would leave the
// decoder's reading of it untested.
type yodaCase struct {
	source      string
	optionsJson string
	fixedSource *string
	findings    []yodaExpectation
}

func yodaStringPointer(value string) *string { return &value }

// decodeYodaOptionsForTest routes a case's options through the shipped decoder.
func decodeYodaOptionsForTest(t *testing.T, optionsJson string) any {
	t.Helper()
	if optionsJson == "" {
		return nil
	}
	decoded, err := DecodeYodaOptions(json.RawMessage(optionsJson))
	if err != nil {
		t.Fatalf("decoding options %q: %v", optionsJson, err)
	}
	return decoded
}

var yodaCleanCases = []yodaCase{
	{source: "if (value === \"red\") {}", optionsJson: "[\"never\"]"},
	{source: "if (value === value) {}", optionsJson: "[\"never\"]"},
	{source: "if (value != 5) {}", optionsJson: "[\"never\"]"},
	{source: "if (5 & foo) {}", optionsJson: "[\"never\"]"},
	{source: "if (5 === 4) {}", optionsJson: "[\"never\"]"},
	{source: "if (value === `red`) {}", optionsJson: "[\"never\"]"},
	{source: "if (`red` === `red`) {}", optionsJson: "[\"never\"]"},
	{source: "if (`${foo}` === `red`) {}", optionsJson: "[\"never\"]"},
	{source: "if (`${\"\"}` === `red`) {}", optionsJson: "[\"never\"]"},
	{source: "if (`${\"red\"}` === foo) {}", optionsJson: "[\"never\"]"},
	{source: "if (b > `a` && b > `a`) {}", optionsJson: "[\"never\"]"},
	{source: "if (`b` > `a` && \"b\" > \"a\") {}", optionsJson: "[\"never\"]"},
	{source: "if (\"blue\" === value) {}", optionsJson: "[\"always\"]"},
	{source: "if (value === value) {}", optionsJson: "[\"always\"]"},
	{source: "if (4 != value) {}", optionsJson: "[\"always\"]"},
	{source: "if (foo & 4) {}", optionsJson: "[\"always\"]"},
	{source: "if (5 === 4) {}", optionsJson: "[\"always\"]"},
	{source: "if (`red` === value) {}", optionsJson: "[\"always\"]"},
	{source: "if (`red` === `red`) {}", optionsJson: "[\"always\"]"},
	{source: "if (`red` === `${foo}`) {}", optionsJson: "[\"always\"]"},
	{source: "if (`red` === `${\"\"}`) {}", optionsJson: "[\"always\"]"},
	{source: "if (foo === `${\"red\"}`) {}", optionsJson: "[\"always\"]"},
	{source: "if (`a` > b && `a` > b) {}", optionsJson: "[\"always\"]"},
	{source: "if (`b` > `a` && \"b\" > \"a\") {}", optionsJson: "[\"always\"]"},
	{source: "if (\"a\" < x && x < MAX ) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (1 < x && x < MAX ) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if ('a' < x && x < MAX ) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (x < `x` || `x` <= x) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (0 < x && x <= 1) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (0 <= x && x < 1) {}", optionsJson: "[\"always\",{\"exceptRange\":true}]"},
	{source: "if ('blue' < x.y && x.y < 'green') {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (0 < x[``] && x[``] < 100) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (0 < x[''] && x[``] < 100) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (a < 4 || (b[c[0]].d['e'] < 0 || 1 <= b[c[0]].d['e'])) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (0 <= x['y'] && x['y'] <= 100) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (a < 0 && (0 < b && b < 1)) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if ((0 < a && a < 1) && b < 0) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (-1 < x && x < 0) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (0 <= this.prop && this.prop <= 1) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (0 <= index && index < list.length) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (ZERO <= index && index < 100) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (value <= MIN || 10 < value) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (value <= 0 || MAX < value) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (0 <= a.b && a[\"b\"] <= 100) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (0 <= a.b && a[`b`] <= 100) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (-1n < x && x <= 1n) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (-1n <= x && x < 1n) {}", optionsJson: "[\"always\",{\"exceptRange\":true}]"},
	{source: "if (x < `1` || `1` < x) {}", optionsJson: "[\"always\",{\"exceptRange\":true}]"},
	{source: "if (1 <= a['/(?<zero>0)/'] && a[/(?<zero>0)/] <= 100) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (x <= `bar` || `foo` < x) {}", optionsJson: "[\"always\",{\"exceptRange\":true}]"},
	{source: "if ('a' < x && x < MAX ) {}", optionsJson: "[\"always\",{\"exceptRange\":true}]"},
	{source: "if ('a' < x && x < MAX ) {}", optionsJson: "[\"always\"]"},
	{source: "if (MIN < x && x < 'a' ) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (MIN < x && x < 'a' ) {}", optionsJson: "[\"never\"]"},
	{source: "if (`blue` < x.y && x.y < `green`) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (0 <= x[`y`] && x[`y`] <= 100) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (0 <= x[`y`] && x[\"y\"] <= 100) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if ('a' <= x && x < 'b') {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (x < -1n || 1n <= x) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (x < -1n || 1n <= x) {}", optionsJson: "[\"always\",{\"exceptRange\":true}]"},
	{source: "if (1 < a && a <= 2) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (x < -1 || 1 < x) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (x <= 'bar' || 'foo' < x) {}", optionsJson: "[\"always\",{\"exceptRange\":true}]"},
	{source: "if (x < 0 || 1 <= x) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if('a' <= x && x < MAX) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (0 <= obj?.a && obj?.a < 1) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]"},
	{source: "if (0 < x && x <= 1) {}", optionsJson: "[\"never\",{\"onlyEquality\":true}]"},
	{source: "if (x !== 'foo' && 'foo' !== x) {}", optionsJson: "[\"never\",{\"onlyEquality\":true}]"},
	{source: "if (x < 2 && x !== -3) {}", optionsJson: "[\"always\",{\"onlyEquality\":true}]"},
	{source: "if (x !== `foo` && `foo` !== x) {}", optionsJson: "[\"never\",{\"onlyEquality\":true}]"},
	{source: "if (x < `2` && x !== `-3`) {}", optionsJson: "[\"always\",{\"onlyEquality\":true}]"},
}

var yodaFiringCases = []yodaCase{
	{source: "if (x <= 'foo' || 'bar' < x) {}", optionsJson: "[\"always\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("if ('foo' >= x || 'bar' < x) {}"), findings: []yodaExpectation{{expectedSide: "left", operator: "<="}}},
	{source: "if (\"red\" == value) {}", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("if (value == \"red\") {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "=="}}},
	{source: "if (true === value) {}", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("if (value === true) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "==="}}},
	{source: "if (5 != value) {}", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("if (value != 5) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "!="}}},
	{source: "if (5n != value) {}", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("if (value != 5n) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "!="}}},
	{source: "if (null !== value) {}", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("if (value !== null) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "!=="}}},
	{source: "if (\"red\" <= value) {}", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("if (value >= \"red\") {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "if (`red` <= value) {}", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("if (value >= `red`) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "if (`red` <= `${foo}`) {}", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("if (`${foo}` >= `red`) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "if (`red` <= `${\"red\"}`) {}", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("if (`${\"red\"}` >= `red`) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "if (true >= value) {}", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("if (value <= true) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: ">="}}},
	{source: "var foo = (5 < value) ? true : false", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("var foo = (value > 5) ? true : false"), findings: []yodaExpectation{{expectedSide: "right", operator: "<"}}},
	{source: "function foo() { return (null > value); }", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("function foo() { return (value < null); }"), findings: []yodaExpectation{{expectedSide: "right", operator: ">"}}},
	{source: "if (-1 < str.indexOf(substr)) {}", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("if (str.indexOf(substr) > -1) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<"}}},
	{source: "if (value == \"red\") {}", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("if (\"red\" == value) {}"), findings: []yodaExpectation{{expectedSide: "left", operator: "=="}}},
	{source: "if (value == `red`) {}", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("if (`red` == value) {}"), findings: []yodaExpectation{{expectedSide: "left", operator: "=="}}},
	{source: "if (value === true) {}", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("if (true === value) {}"), findings: []yodaExpectation{{expectedSide: "left", operator: "==="}}},
	{source: "if (value === 5n) {}", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("if (5n === value) {}"), findings: []yodaExpectation{{expectedSide: "left", operator: "==="}}},
	{source: "if (`${\"red\"}` <= `red`) {}", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("if (`red` >= `${\"red\"}`) {}"), findings: []yodaExpectation{{expectedSide: "left", operator: "<="}}},
	{source: "if (a < 0 && 0 <= b && b < 1) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("if (a < 0 && b >= 0 && b < 1) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "if (0 <= a && a < 1 && b < 1) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("if (a >= 0 && a < 1 && b < 1) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "if (1 < a && a < 0) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("if (a > 1 && a < 0) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<"}}},
	{source: "0 < a && a < 1", optionsJson: "[\"never\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("a > 0 && a < 1"), findings: []yodaExpectation{{expectedSide: "right", operator: "<"}}},
	{source: "var a = b < 0 || 1 <= b;", optionsJson: "[\"never\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("var a = b < 0 || b >= 1;"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "if (0 <= x && x < -1) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("if (x >= 0 && x < -1) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "var a = (b < 0 && 0 <= b);", optionsJson: "[\"always\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("var a = (0 > b && 0 <= b);"), findings: []yodaExpectation{{expectedSide: "left", operator: "<"}}},
	{source: "var a = (b < `0` && `0` <= b);", optionsJson: "[\"always\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("var a = (`0` > b && `0` <= b);"), findings: []yodaExpectation{{expectedSide: "left", operator: "<"}}},
	{source: "if (`green` < x.y && x.y < `blue`) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("if (x.y > `green` && x.y < `blue`) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<"}}},
	{source: "if (0 <= a[b] && a['b'] < 1) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("if (a[b] >= 0 && a['b'] < 1) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "if (0 <= a[b] && a[`b`] < 1) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("if (a[b] >= 0 && a[`b`] < 1) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "if (`0` <= a[b] && a[`b`] < `1`) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("if (a[b] >= `0` && a[`b`] < `1`) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "if (0 <= a[b] && a.b < 1) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("if (a[b] >= 0 && a.b < 1) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "if (0 <= a[''] && a.b < 1) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("if (a[''] >= 0 && a.b < 1) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "if (0 <= a[''] && a[' '] < 1) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("if (a[''] >= 0 && a[' '] < 1) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "if (0 <= a[''] && a[null] < 1) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("if (a[''] >= 0 && a[null] < 1) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "if (0 <= a[``] && a[null] < 1) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("if (a[``] >= 0 && a[null] < 1) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "if (0 <= a[''] && a[b] < 1) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("if (a[''] >= 0 && a[b] < 1) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "if (0 <= a[''] && a[b()] < 1) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("if (a[''] >= 0 && a[b()] < 1) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "if (0 <= a[``] && a[b()] < 1) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("if (a[``] >= 0 && a[b()] < 1) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "if (0 <= a[b()] && a[b()] < 1) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("if (a[b()] >= 0 && a[b()] < 1) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "if (0 <= a.null && a[/(?<zero>0)/] <= 1) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("if (a.null >= 0 && a[/(?<zero>0)/] <= 1) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "if (3 == a) {}", optionsJson: "[\"never\",{\"onlyEquality\":true}]", fixedSource: yodaStringPointer("if (a == 3) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "=="}}},
	{source: "foo(3 === a);", optionsJson: "[\"never\",{\"onlyEquality\":true}]", fixedSource: yodaStringPointer("foo(a === 3);"), findings: []yodaExpectation{{expectedSide: "right", operator: "==="}}},
	{source: "foo(a === 3);", optionsJson: "[\"always\",{\"onlyEquality\":true}]", fixedSource: yodaStringPointer("foo(3 === a);"), findings: []yodaExpectation{{expectedSide: "left", operator: "==="}}},
	{source: "foo(a === `3`);", optionsJson: "[\"always\",{\"onlyEquality\":true}]", fixedSource: yodaStringPointer("foo(`3` === a);"), findings: []yodaExpectation{{expectedSide: "left", operator: "==="}}},
	{source: "if (0 <= x && x < 1) {}", optionsJson: "", fixedSource: yodaStringPointer("if (x >= 0 && x < 1) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "if ( /* a */ 0 /* b */ < /* c */ foo /* d */ ) {}", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("if ( /* a */ foo /* b */ > /* c */ 0 /* d */ ) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<"}}},
	{source: "if ( /* a */ foo /* b */ > /* c */ 0 /* d */ ) {}", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("if ( /* a */ 0 /* b */ < /* c */ foo /* d */ ) {}"), findings: []yodaExpectation{{expectedSide: "left", operator: ">"}}},
	{source: "if (foo()===1) {}", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("if (1===foo()) {}"), findings: []yodaExpectation{{expectedSide: "left", operator: "==="}}},
	{source: "if (foo()     === 1) {}", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("if (1     === foo()) {}"), findings: []yodaExpectation{{expectedSide: "left", operator: "==="}}},
	{source: "while (0 === (a));", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("while ((a) === 0);"), findings: []yodaExpectation{{expectedSide: "right", operator: "==="}}},
	{source: "while (0 === (a = b));", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("while ((a = b) === 0);"), findings: []yodaExpectation{{expectedSide: "right", operator: "==="}}},
	{source: "while ((a) === 0);", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("while (0 === (a));"), findings: []yodaExpectation{{expectedSide: "left", operator: "==="}}},
	{source: "while ((a = b) === 0);", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("while (0 === (a = b));"), findings: []yodaExpectation{{expectedSide: "left", operator: "==="}}},
	{source: "if (((((((((((foo)))))))))) === ((((((5)))))));", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("if (((((((5)))))) === ((((((((((foo)))))))))));"), findings: []yodaExpectation{{expectedSide: "left", operator: "==="}}},
	{source: "function *foo() { yield(1) < a }", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("function *foo() { yield a > (1) }"), findings: []yodaExpectation{{expectedSide: "right", operator: "<"}}},
	{source: "function *foo() { yield((1)) < a }", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("function *foo() { yield a > ((1)) }"), findings: []yodaExpectation{{expectedSide: "right", operator: "<"}}},
	{source: "function *foo() { yield 1 < a }", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("function *foo() { yield a > 1 }"), findings: []yodaExpectation{{expectedSide: "right", operator: "<"}}},
	{source: "function *foo() { yield/**/1 < a }", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("function *foo() { yield/**/a > 1 }"), findings: []yodaExpectation{{expectedSide: "right", operator: "<"}}},
	{source: "function *foo() { yield(1) < ++a }", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("function *foo() { yield++a > (1) }"), findings: []yodaExpectation{{expectedSide: "right", operator: "<"}}},
	{source: "function *foo() { yield(1) < (a) }", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("function *foo() { yield(a) > (1) }"), findings: []yodaExpectation{{expectedSide: "right", operator: "<"}}},
	{source: "x=1 < a", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("x=a > 1"), findings: []yodaExpectation{{expectedSide: "right", operator: "<"}}},
	{source: "function *foo() { yield++a < 1 }", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("function *foo() { yield 1 > ++a }"), findings: []yodaExpectation{{expectedSide: "left", operator: "<"}}},
	{source: "function *foo() { yield(a) < 1 }", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("function *foo() { yield 1 > (a) }"), findings: []yodaExpectation{{expectedSide: "left", operator: "<"}}},
	{source: "function *foo() { yield a < 1 }", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("function *foo() { yield 1 > a }"), findings: []yodaExpectation{{expectedSide: "left", operator: "<"}}},
	{source: "function *foo() { yield/**/a < 1 }", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("function *foo() { yield/**/1 > a }"), findings: []yodaExpectation{{expectedSide: "left", operator: "<"}}},
	{source: "function *foo() { yield++a < (1) }", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("function *foo() { yield(1) > ++a }"), findings: []yodaExpectation{{expectedSide: "left", operator: "<"}}},
	{source: "x=a < 1", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("x=1 > a"), findings: []yodaExpectation{{expectedSide: "left", operator: "<"}}},
	{source: "0 < f()in obj", optionsJson: "", fixedSource: yodaStringPointer("f() > 0 in obj"), findings: []yodaExpectation{{expectedSide: "right", operator: "<"}}},
	{source: "1 > x++instanceof foo", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("x++ < 1 instanceof foo"), findings: []yodaExpectation{{expectedSide: "right", operator: ">"}}},
	{source: "x < ('foo')in bar", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("('foo') > x in bar"), findings: []yodaExpectation{{expectedSide: "left", operator: "<"}}},
	{source: "false <= ((x))in foo", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("((x)) >= false in foo"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "x >= (1)instanceof foo", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("(1) <= x instanceof foo"), findings: []yodaExpectation{{expectedSide: "left", operator: ">="}}},
	{source: "false <= ((x)) in foo", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("((x)) >= false in foo"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "x >= 1 instanceof foo", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("1 <= x instanceof foo"), findings: []yodaExpectation{{expectedSide: "left", operator: ">="}}},
	{source: "x >= 1/**/instanceof foo", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("1 <= x/**/instanceof foo"), findings: []yodaExpectation{{expectedSide: "left", operator: ">="}}},
	{source: "(x >= 1)instanceof foo", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("(1 <= x)instanceof foo"), findings: []yodaExpectation{{expectedSide: "left", operator: ">="}}},
	{source: "(x) >= (1)instanceof foo", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("(1) <= (x)instanceof foo"), findings: []yodaExpectation{{expectedSide: "left", operator: ">="}}},
	{source: "1 > x===foo", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("x < 1===foo"), findings: []yodaExpectation{{expectedSide: "right", operator: ">"}}},
	{source: "1 > x", optionsJson: "[\"never\"]", fixedSource: yodaStringPointer("x < 1"), findings: []yodaExpectation{{expectedSide: "right", operator: ">"}}},
	{source: "if (`green` < x.y && x.y < `blue`) {}", optionsJson: "[\"always\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("if (`green` < x.y && `blue` > x.y) {}"), findings: []yodaExpectation{{expectedSide: "left", operator: "<"}}},
	{source: "if('a' <= x && x < 'b') {}", optionsJson: "[\"always\"]", fixedSource: yodaStringPointer("if('a' <= x && 'b' > x) {}"), findings: []yodaExpectation{{expectedSide: "left", operator: "<"}}},
	{source: "if ('b' <= x && x < 'a') {}", optionsJson: "[\"never\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("if (x >= 'b' && x < 'a') {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "if('a' <= x && x < 1) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("if(x >= 'a' && x < 1) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<="}}},
	{source: "if (0 < a && b < max) {}", optionsJson: "[\"never\",{\"exceptRange\":true}]", fixedSource: yodaStringPointer("if (a > 0 && b < max) {}"), findings: []yodaExpectation{{expectedSide: "right", operator: "<"}}},
}

// TestYodaFires runs upstream's 85 invalid cases.
//
// Every one carries an `output`, so this rule's corpus is 85 fix vectors as well as 85 judgments.
// Each case asserts three things: that the finding appears, that its rendered message names the
// right side and the right operator, and that the repair writes exactly what upstream's output
// says. The last is the one that matters most for a fixer, because a repair anchored correctly can
// still write the wrong bytes and no message assertion can see it.
func TestYodaFires(t *testing.T) {
	t.Parallel()

	for _, testCase := range yodaFiringCases {
		t.Run(testCase.source, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, Yoda, yodaFile, testCase.source,
				decodeYodaOptionsForTest(t, testCase.optionsJson))

			wantIds := make([]string, len(testCase.findings))
			for i := range wantIds {
				wantIds[i] = "expected"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			for i, finding := range testCase.findings {
				// Upstream's template is
				// `Expected literal to be on the {{expectedSide}} side of {{operator}}.`
				// Asserted as an exact prefix rather than with a substring test, because both
				// values are interpolated and a `strings.Contains` cannot see a swapped pair.
				want := "Expected literal to be on the " + finding.expectedSide +
					" side of " + finding.operator + "."
				if got := result.Diagnostics[i].Message.Description; !strings.HasPrefix(got, want) {
					t.Errorf("finding %d: message should begin %q, got %q", i, want, got)
				}
			}

			if testCase.fixedSource == nil {
				for i, diagnostic := range result.Diagnostics {
					if len(diagnostic.Fixes) != 0 {
						t.Errorf("finding %d proposes a fix; upstream declines this case", i)
					}
				}
				return
			}
			rule_testing.ExpectFixedSource(t, result, *testCase.fixedSource)
		})
	}
}

// TestYodaStaysSilent runs upstream's 71 valid cases.
//
// Forty two of them exercise `exceptRange`, which is the bulk of this rule's judgment and the part
// most likely to be wrong: the range test compares two operands structurally, orders the two
// bounds, and requires the whole thing to be parenthesised, and a port getting any of the three
// wrong passes the other cases.
func TestYodaStaysSilent(t *testing.T) {
	t.Parallel()

	for _, testCase := range yodaCleanCases {
		t.Run(testCase.source, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, Yoda, yodaFile,
				testCase.source, decodeYodaOptionsForTest(t, testCase.optionsJson)))
		})
	}
}

// yodaTypeScriptCases are shapes upstream's corpus cannot contain, because its corpus is
// JavaScript and the gap between it and our tree is exactly TypeScript syntax.
//
// This is the set that matters for a FIXER. Two rules in this tree silently destroyed type
// information while passing every one of upstream's fixture cases, both because they CONSTRUCTED a
// replacement from node properties. This fixer reorders verbatim source slices instead, so the
// claim is that nothing inside an operand can be lost, and these cases are what test the claim
// rather than assert it.
//
// Every expected output below was measured by driving the installed rule with the typescript-eslint
// parser, not derived from reading the fixer.
var yodaTypeScriptCases = []struct {
	name        string
	source      string
	optionsJson string
	fixedSource string
	reason      string
}{
	{
		name:        "a type assertion survives the flip",
		source:      "if ('red' === (color as string)) {}",
		fixedSource: "if ((color as string) === 'red') {}",
		reason:      "The `as` and its type are inside the operand's own text, so they move with it.",
	},
	{
		name:        "a non-null assertion survives",
		source:      "if ('red' === color!) {}",
		fixedSource: "if (color! === 'red') {}",
		reason:      "The `!` sits at the end of the operand rather than outside it.",
	},
	{
		name:        "optional chaining survives",
		source:      "if ('red' === color?.name) {}",
		fixedSource: "if (color?.name === 'red') {}",
		reason:      "The whole chain is one operand.",
	},
	{
		name:        "a satisfies expression survives, and the operator flips",
		source:      "if (0 <= (x satisfies number)) {}",
		fixedSource: "if ((x satisfies number) >= 0) {}",
		reason: "Both halves at once: the operand keeps its `satisfies` and `<=` becomes `>=`, " +
			"which is what makes the repair mean the same thing.",
	},
	{
		name:        "an angle-bracket assertion survives",
		source:      "if ('red' === <string>color) {}",
		fixedSource: "if (<string>color === 'red') {}",
		reason: "The older assertion spelling. Worth its own case because it puts a `<` at the " +
			"START of the operand being moved to the left of a comparison.",
	},
	{
		name:        "explicit type arguments on a call survive",
		source:      "if (1 === foo<number>()) {}",
		fixedSource: "if (foo<number>() === 1) {}",
		reason:      "Type arguments are inside the call's own text.",
	},
	{
		name:        "a generic type annotation on the enclosing statement is untouched",
		source:      "const check: Array<string> = []; if ('red' === check[0]) {}",
		fixedSource: "const check: Array<string> = []; if (check[0] === 'red') {}",
		reason: "The replaced span is the comparison alone, so nothing outside it can be reached. " +
			"This is the shape that a fixer computing its range from a neighbour would damage.",
	},
	{
		name:        "a comment before the operator stays with the operator",
		source:      "if ('red' /* mid */ === color) {}",
		fixedSource: "if (color /* mid */ === 'red') {}",
		reason: "Not what a reader would guess. The slices are cut at the operator TOKEN, so the " +
			"text between the left operand and the operator stays in place while the operands " +
			"swap around it. Measured against the installed rule, which does the same.",
	},
	{
		name:        "a comment after the operator also stays put",
		source:      "if ('red' ===/*a*/ color) {}",
		fixedSource: "if (color ===/*a*/ 'red') {}",
		reason:      "The other side of the same cut.",
	},
	{
		name:        "a trailing comment inside the right operand moves with it",
		source:      "if ('red' === color /* trailing */) {}",
		fixedSource: "if (color === 'red' /* trailing */) {}",
		reason: "This one is inside the comparison's own span but outside either operand's token " +
			"range, so it rides along at the end. Measured.",
	},
}

// TestYodaTypeScriptOperandsSurviveTheFix covers what upstream's corpus structurally cannot.
func TestYodaTypeScriptOperandsSurviveTheFix(t *testing.T) {
	t.Parallel()

	for _, testCase := range yodaTypeScriptCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, Yoda, yodaFile, testCase.source,
				decodeYodaOptionsForTest(t, testCase.optionsJson))
			rule_testing.ExpectFindings(t, result, "expected")
			rule_testing.ExpectFixedSource(t, result, testCase.fixedSource)
		})
	}
}

// TestYodaRangeTestRequiresLessThanOperators pins the range-test operator restriction.
//
// A range test must use `<` or `<=` on BOTH sides. Upstream's `isRangeTestOperator` names those two
// and no others, so a `>`-shaped range is not exempt even when it expresses the same interval.
// Every case below reports, and the exempt control beside them is what shows the option is
// otherwise working.
//
// Added because a mutation widening that predicate to accept every operator survived all 156
// corpus cases: upstream writes no `>` range anywhere, so nothing imported can tell a working
// restriction from an absent one. Each verdict was measured against the installed rule.
func TestYodaRangeTestRequiresLessThanOperators(t *testing.T) {
	t.Parallel()

	options := decodeYodaOptionsForTest(t, `["never",{"exceptRange":true}]`)

	// The control: a genuine range test, exempt.
	rule_testing.ExpectClean(t,
		rule_testing.RunWithOptions(t, Yoda, yodaFile, "if (0 <= x && x < 1) {}", options))

	for _, source := range []string{
		"if (1 > x && x > 0) {}",
		"if (x > 0 && 1 > x) {}",
		"if (0 >= x && x >= 1) {}",
		"if (1 >= x || x >= 0) {}",
		"if (0 < x && x > 1) {}",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t,
				rule_testing.RunWithOptions(t, Yoda, yodaFile, source, options), "expected")
		})
	}
}

// TestDecodeYodaOptions pins the list shapes the config layer delivers and the ones it must refuse.
// The second element is where `exceptRange` and `onlyEquality` live, and it used to be dropped by
// the config layer, so the row reading it is the one that matters.
func TestDecodeYodaOptions(t *testing.T) {
	t.Parallel()

	decoded, err := DecodeYodaOptions([]byte(`["always", {"exceptRange": true, "onlyEquality": true}]`))
	if err != nil {
		t.Fatalf("upstream's spelling was refused: %v", err)
	}
	if settings := decoded.(YodaSettings); !settings.Always || !settings.ExceptRange || !settings.OnlyEquality {
		t.Errorf("the second element was not read: %+v", settings)
	}

	for _, raw := range []string{
		// A bare string, which the config layer never delivers to a list rule.
		`"always"`,
		// The one-object workaround, and the nested one.
		`{"when": "always", "exceptRange": true}`,
		`[["always", {"exceptRange": true}]]`,
		// A key upstream does not declare, a third element, a mode outside the two.
		`["always", {"exceptRanges": true}]`,
		`["always", {"exceptRange": true}, "never"]`,
		`["sometimes"]`,
	} {
		if decoded, err := DecodeYodaOptions([]byte(raw)); err == nil {
			t.Errorf("%s decoded to %+v; it must be refused", raw, decoded)
		}
	}
}

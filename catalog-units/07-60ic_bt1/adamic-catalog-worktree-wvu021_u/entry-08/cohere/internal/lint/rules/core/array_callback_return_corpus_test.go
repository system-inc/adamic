// Code generated from upstream's corpus by /tmp/tjpst83/gen_acr_fixtures.js. Do not hand-edit;
// see arrayCallbackReturnCorpusProvenance in array_callback_return_test.go.

package core

// arrayCallbackReturnCase is one imported upstream case: its source, the options it was tested
// under, and the message ids upstream reports on it in order.
//
// suggestions holds, per finding, the suggestion ids upstream offers and the source each one
// produces when applied. A finding offering none carries an empty slice, which is a decision to
// reproduce rather than an absence.
type arrayCallbackReturnCase struct {
	source      string
	options     any
	wantIds     []string
	suggestions [][]arrayCallbackReturnSuggestion
}

// arrayCallbackReturnSuggestion is one offered repair and the source applying it yields.
type arrayCallbackReturnSuggestion struct {
	id     string
	output string
}

var arrayCallbackReturnCleanCases = []arrayCallbackReturnCase{
	{
		source:  `foo.every(function(){}())`,
		options: nil,
	},
	{
		source:  `foo.every(function(){ return function() { return true; }; }())`,
		options: nil,
	},
	{
		source:  `foo.every(function(){ return function() { return; }; })`,
		options: nil,
	},
	{
		source:  `foo.forEach(bar || function(x) { var a=0; })`,
		options: nil,
	},
	{
		source:  `foo.forEach(bar || function(x) { return a; })`,
		options: nil,
	},
	{
		source:  `foo.forEach(function() {return function() { var a = 0;}}())`,
		options: nil,
	},
	{
		source:  `foo.forEach(function(x) { var a=0; })`,
		options: nil,
	},
	{
		source:  `foo.forEach(function(x) { return a;})`,
		options: nil,
	},
	{
		source:  `foo.forEach(function(x) { return; })`,
		options: nil,
	},
	{
		source:  `foo.forEach(function(x) { if (a === b) { return;} var a=0; })`,
		options: nil,
	},
	{
		source:  `foo.forEach(function(x) { if (a === b) { return x;} var a=0; })`,
		options: nil,
	},
	{
		source:  `foo.bar().forEach(function(x) { return; })`,
		options: nil,
	},
	{
		source:  `["foo","bar","baz"].forEach(function(x) { return x; })`,
		options: nil,
	},
	{
		source:  `foo.forEach(x => { var a=0; })`,
		options: nil,
	},
	{
		source:  `foo.forEach(x => { if (a === b) { return;} var a=0; })`,
		options: nil,
	},
	{
		source:  `foo.forEach(x => x)`,
		options: nil,
	},
	{
		source:  `foo.forEach(val => y += val)`,
		options: nil,
	},
	{
		source:  `foo.map(async function(){})`,
		options: nil,
	},
	{
		source:  `foo.map(async () => {})`,
		options: nil,
	},
	{
		source:  `foo.map(function* () {})`,
		options: nil,
	},
	{
		source:  `Array.from(x, function() { return true; })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: false},
	},
	{
		source:  `Int32Array.from(x, function() { return true; })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: false},
	},
	{
		source:  `foo.every(function() { return true; })`,
		options: nil,
	},
	{
		source:  `foo.filter(function() { return true; })`,
		options: nil,
	},
	{
		source:  `foo.find(function() { return true; })`,
		options: nil,
	},
	{
		source:  `foo.findIndex(function() { return true; })`,
		options: nil,
	},
	{
		source:  `foo.findLast(function() { return true; })`,
		options: nil,
	},
	{
		source:  `foo.findLastIndex(function() { return true; })`,
		options: nil,
	},
	{
		source:  `foo.flatMap(function() { return true; })`,
		options: nil,
	},
	{
		source:  `foo.forEach(function() { return; })`,
		options: nil,
	},
	{
		source:  `foo.map(function() { return true; })`,
		options: nil,
	},
	{
		source:  `foo.reduce(function() { return true; })`,
		options: nil,
	},
	{
		source:  `foo.reduceRight(function() { return true; })`,
		options: nil,
	},
	{
		source:  `foo.some(function() { return true; })`,
		options: nil,
	},
	{
		source:  `foo.sort(function() { return 0; })`,
		options: nil,
	},
	{
		source:  `foo.toSorted(function() { return 0; })`,
		options: nil,
	},
	{
		source:  `foo.every(() => { return true; })`,
		options: nil,
	},
	{
		source:  `foo.every(function() { if (a) return true; else return false; })`,
		options: nil,
	},
	{
		source:  `foo.every(function() { switch (a) { case 0: bar(); default: return true; } })`,
		options: nil,
	},
	{
		source:  `foo.every(function() { try { bar(); return true; } catch (err) { return false; } })`,
		options: nil,
	},
	{
		source:  `foo.every(function() { try { bar(); } finally { return true; } })`,
		options: nil,
	},
	{
		source:  `Array.from(x, function() { return; })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
	},
	{
		source:  `Int32Array.from(x, function() { return; })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
	},
	{
		source:  `foo.every(function() { return; })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
	},
	{
		source:  `foo.filter(function() { return; })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
	},
	{
		source:  `foo.find(function() { return; })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
	},
	{
		source:  `foo.findIndex(function() { return; })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
	},
	{
		source:  `foo.findLast(function() { return; })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
	},
	{
		source:  `foo.findLastIndex(function() { return; })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
	},
	{
		source:  `foo.flatMap(function() { return; })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
	},
	{
		source:  `foo.forEach(function() { return; })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
	},
	{
		source:  `foo.map(function() { return; })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
	},
	{
		source:  `foo.reduce(function() { return; })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
	},
	{
		source:  `foo.reduceRight(function() { return; })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
	},
	{
		source:  `foo.some(function() { return; })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
	},
	{
		source:  `foo.sort(function() { return; })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
	},
	{
		source:  `foo.toSorted(function() { return; })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
	},
	{
		source:  `foo.every(() => { return; })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
	},
	{
		source:  `foo.every(function() { if (a) return; else return a; })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
	},
	{
		source:  `foo.every(function() { switch (a) { case 0: bar(); default: return; } })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
	},
	{
		source:  `foo.every(function() { try { bar(); return; } catch (err) { return; } })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
	},
	{
		source:  `foo.every(function() { try { bar(); } finally { return; } })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
	},
	{
		source:  `foo.forEach(function(x) { return; })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
	},
	{
		source:  `foo.forEach(function(x) { var a=0; })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
	},
	{
		source:  `foo.forEach(function(x) { if (a === b) { return;} var a=0; })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
	},
	{
		source:  `foo.forEach(function() {return function() { if (a == b) { return; }}}())`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
	},
	{
		source:  `foo.forEach(x => { var a=0; })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
	},
	{
		source:  `foo.forEach(x => { if (a === b) { return;} var a=0; })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
	},
	{
		source:  `foo.forEach(x => { x })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
	},
	{
		source:  `foo.forEach(bar || function(x) { return; })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
	},
	{
		source:  `Array.from(x, function() { return true; })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
	},
	{
		source:  `Int32Array.from(x, function() { return true; })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
	},
	{
		source:  `foo.every(() => { return true; })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
	},
	{
		source:  `foo.every(function() { if (a) return 1; else return a; })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
	},
	{
		source:  `foo.every(function() { switch (a) { case 0: return bar(); default: return a; } })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
	},
	{
		source:  `foo.every(function() { try { bar(); return 1; } catch (err) { return err; } })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
	},
	{
		source:  `foo.every(function() { try { bar(); } finally { return 1; } })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
	},
	{
		source:  `foo.every(function() { return; })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true, CheckForEach: true},
	},
	{
		source:  `foo.forEach((x) => void x)`,
		options: ArrayCallbackReturnOptions{CheckForEach: true, AllowVoid: true},
	},
	{
		source:  `foo.forEach((x) => void bar(x))`,
		options: ArrayCallbackReturnOptions{CheckForEach: true, AllowVoid: true},
	},
	{
		source:  `foo.forEach(function (x) { return void bar(x); })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true, AllowVoid: true},
	},
	{
		source:  `foo.forEach((x) => { return void bar(x); })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true, AllowVoid: true},
	},
	{
		source:  `foo.forEach((x) => { if (a === b) { return void a; } bar(x) })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true, AllowVoid: true},
	},
	{
		source:  `Arrow.from(x, function() {})`,
		options: nil,
	},
	{
		source:  `foo.abc(function() {})`,
		options: nil,
	},
	{
		source:  `every(function() {})`,
		options: nil,
	},
	{
		source:  `foo[every](function() {})`,
		options: nil,
	},
	{
		source:  `var every = function() {}`,
		options: nil,
	},
	{
		source:  "foo[`${every}`](function() {})",
		options: nil,
	},
	{
		source:  `foo.every(() => true)`,
		options: nil,
	},
	{
		source:  `Array.fromAsync(x, function() { return true; })`,
		options: nil,
	},
	{
		source:  `Array.fromAsync(x, async function() { return true; })`,
		options: nil,
	},
	{
		source:  `Array.fromAsync(x, function() { return; })`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
	},
	{
		source:  `Array.fromAsync(x, async () => true)`,
		options: nil,
	},
	{
		source:  `Array.fromAsync(x, function * () {})`,
		options: nil,
	},
	{
		source:  `Float64Array.fromAsync(x, function() {})`,
		options: nil,
	},
	{
		source:  `Array.fromAsync(function() {})`,
		options: nil,
	},
}

var arrayCallbackReturnReportingCases = []arrayCallbackReturnCase{
	{
		source:  `Array.from(x, function() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `Array.from(x, function foo() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `Int32Array.from(x, function() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `Int32Array.from(x, function foo() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.every(function() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.every(function foo() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.filter(function() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.filter(function foo() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.find(function() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.find(function foo() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.findLast(function() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.findLast(function foo() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.findIndex(function() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.findIndex(function foo() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.findLastIndex(function() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.findLastIndex(function foo() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.flatMap(function() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.flatMap(function foo() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.map(function() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.map(function foo() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.reduce(function() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.reduce(function foo() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.reduceRight(function() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.reduceRight(function foo() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.some(function() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.some(function foo() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.sort(function() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.sort(function foo() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.toSorted(function() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.toSorted(function foo() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.bar.baz.every(function() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.bar.baz.every(function foo() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo["every"](function() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo["every"](function foo() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  "foo[`every`](function() {})",
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  "foo[`every`](function foo() {})",
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.every(() => {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.every(function() { if (a) return true; })`,
		options: nil,
		wantIds: []string{"expectedAtEnd"},
	},
	{
		source:  `foo.every(function cb() { if (a) return true; })`,
		options: nil,
		wantIds: []string{"expectedAtEnd"},
	},
	{
		source:  `foo.every(function() { switch (a) { case 0: break; default: return true; } })`,
		options: nil,
		wantIds: []string{"expectedAtEnd"},
	},
	{
		source:  `foo.every(function foo() { switch (a) { case 0: break; default: return true; } })`,
		options: nil,
		wantIds: []string{"expectedAtEnd"},
	},
	{
		source:  `foo.every(function() { try { bar(); } catch (err) { return true; } })`,
		options: nil,
		wantIds: []string{"expectedAtEnd"},
	},
	{
		source:  `foo.every(function foo() { try { bar(); } catch (err) { return true; } })`,
		options: nil,
		wantIds: []string{"expectedAtEnd"},
	},
	{
		source:  `foo.every(function() { return; })`,
		options: nil,
		wantIds: []string{"expectedReturnValue"},
	},
	{
		source:  `foo.every(function foo() { return; })`,
		options: nil,
		wantIds: []string{"expectedReturnValue"},
	},
	{
		source:  `foo.every(function() { if (a) return; })`,
		options: nil,
		wantIds: []string{"expectedAtEnd", "expectedReturnValue"},
	},
	{
		source:  `foo.every(function foo() { if (a) return; })`,
		options: nil,
		wantIds: []string{"expectedAtEnd", "expectedReturnValue"},
	},
	{
		source:  `foo.every(function() { if (a) return; else return; })`,
		options: nil,
		wantIds: []string{"expectedReturnValue", "expectedReturnValue"},
	},
	{
		source:  `foo.every(function foo() { if (a) return; else return; })`,
		options: nil,
		wantIds: []string{"expectedReturnValue", "expectedReturnValue"},
	},
	{
		source:  `foo.every(cb || function() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.every(cb || function foo() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.every(a ? function() {} : function() {})`,
		options: nil,
		wantIds: []string{"expectedInside", "expectedInside"},
	},
	{
		source:  `foo.every(a ? function foo() {} : function bar() {})`,
		options: nil,
		wantIds: []string{"expectedInside", "expectedInside"},
	},
	{
		source:  `foo.every(function(){ return function() {}; }())`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.every(function(){ return function foo() {}; }())`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.every(() => {})`,
		options: ArrayCallbackReturnOptions{AllowImplicit: false},
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.every(() => {})`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `Array.from(x, function() {})`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.every(function() {})`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.filter(function foo() {})`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.find(function foo() {})`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.map(function() {})`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.reduce(function() {})`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.reduceRight(function() {})`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.bar.baz.every(function foo() {})`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.every(cb || function() {})`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `["foo","bar"].sort(function foo() {})`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `["foo","bar"].toSorted(function foo() {})`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.forEach(x => x)`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true, CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
		suggestions: [][]arrayCallbackReturnSuggestion{
			{
				{id: "wrapBraces", output: `foo.forEach(x => {x})`},
			},
		},
	},
	{
		source:  `foo.forEach(function(x) { if (a == b) {return x;}})`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true, CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
	},
	{
		source:  `foo.forEach(function bar(x) { return x;})`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true, CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
	},
	{
		source:  `foo.forEach(x => x)`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
		suggestions: [][]arrayCallbackReturnSuggestion{
			{
				{id: "wrapBraces", output: `foo.forEach(x => {x})`},
			},
		},
	},
	{
		source:  `foo.forEach(x => (x))`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
		suggestions: [][]arrayCallbackReturnSuggestion{
			{
				{id: "wrapBraces", output: `foo.forEach(x => {(x)})`},
			},
		},
	},
	{
		source:  `foo.forEach(val => y += val)`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
		suggestions: [][]arrayCallbackReturnSuggestion{
			{
				{id: "wrapBraces", output: `foo.forEach(val => {y += val})`},
			},
		},
	},
	{
		source:  `["foo","bar"].forEach(x => ++x)`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
		suggestions: [][]arrayCallbackReturnSuggestion{
			{
				{id: "wrapBraces", output: `["foo","bar"].forEach(x => {++x})`},
			},
		},
	},
	{
		source:  `foo.bar().forEach(x => x === y)`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
		suggestions: [][]arrayCallbackReturnSuggestion{
			{
				{id: "wrapBraces", output: `foo.bar().forEach(x => {x === y})`},
			},
		},
	},
	{
		source:  `foo.forEach(function() {return function() { if (a == b) { return a; }}}())`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
	},
	{
		source:  `foo.forEach(function(x) { if (a == b) {return x;}})`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
	},
	{
		source:  `foo.forEach(function(x) { if (a == b) {return undefined;}})`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
	},
	{
		source:  `foo.forEach(function bar(x) { return x;})`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
	},
	{
		source:  `foo.bar().forEach(function bar(x) { return x;})`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
	},
	{
		source:  `["foo","bar"].forEach(function bar(x) { return x;})`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
	},
	{
		source:  `foo.forEach((x) => { return x;})`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
	},
	{
		source:  `Array.from(x, function() {})`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.every(function() {})`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.filter(function foo() {})`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.filter(function foo() { return; })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedReturnValue"},
	},
	{
		source:  `foo.every(cb || function() {})`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.forEach((x) => void x)`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
		suggestions: [][]arrayCallbackReturnSuggestion{
			{
				{id: "wrapBraces", output: `foo.forEach((x) => {void x})`},
			},
		},
	},
	{
		source:  `foo.forEach((x) => void bar(x))`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
		suggestions: [][]arrayCallbackReturnSuggestion{
			{
				{id: "wrapBraces", output: `foo.forEach((x) => {void bar(x)})`},
			},
		},
	},
	{
		source:  `foo.forEach((x) => { return void bar(x); })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
	},
	{
		source:  `foo.forEach((x) => { if (a === b) { return void a; } bar(x) })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
	},
	{
		source:  `foo.forEach(x => x)`,
		options: ArrayCallbackReturnOptions{CheckForEach: true, AllowVoid: true},
		wantIds: []string{"expectedNoReturnValue"},
		suggestions: [][]arrayCallbackReturnSuggestion{
			{
				{id: "wrapBraces", output: `foo.forEach(x => {x})`},
				{id: "prependVoid", output: `foo.forEach(x => void x)`},
			},
		},
	},
	{
		source:  `foo.forEach(x => !x)`,
		options: ArrayCallbackReturnOptions{CheckForEach: true, AllowVoid: true},
		wantIds: []string{"expectedNoReturnValue"},
		suggestions: [][]arrayCallbackReturnSuggestion{
			{
				{id: "wrapBraces", output: `foo.forEach(x => {!x})`},
				{id: "prependVoid", output: `foo.forEach(x => void !x)`},
			},
		},
	},
	{
		source:  `foo.forEach(x => (x))`,
		options: ArrayCallbackReturnOptions{CheckForEach: true, AllowVoid: true},
		wantIds: []string{"expectedNoReturnValue"},
		suggestions: [][]arrayCallbackReturnSuggestion{
			{
				{id: "wrapBraces", output: `foo.forEach(x => {(x)})`},
				{id: "prependVoid", output: `foo.forEach(x => void (x))`},
			},
		},
	},
	{
		source:  `foo.forEach((x) => { return x; })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true, AllowVoid: true},
		wantIds: []string{"expectedNoReturnValue"},
		suggestions: [][]arrayCallbackReturnSuggestion{
			{
				{id: "prependVoid", output: `foo.forEach((x) => { return void x; })`},
			},
		},
	},
	{
		source:  `foo.forEach((x) => { return !x; })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true, AllowVoid: true},
		wantIds: []string{"expectedNoReturnValue"},
		suggestions: [][]arrayCallbackReturnSuggestion{
			{
				{id: "prependVoid", output: `foo.forEach((x) => { return void !x; })`},
			},
		},
	},
	{
		source:  `foo.forEach((x) => { return(x); })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true, AllowVoid: true},
		wantIds: []string{"expectedNoReturnValue"},
		suggestions: [][]arrayCallbackReturnSuggestion{
			{
				{id: "prependVoid", output: `foo.forEach((x) => { return void (x); })`},
			},
		},
	},
	{
		source:  `foo.forEach((x) => { return (x + 1); })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true, AllowVoid: true},
		wantIds: []string{"expectedNoReturnValue"},
		suggestions: [][]arrayCallbackReturnSuggestion{
			{
				{id: "prependVoid", output: `foo.forEach((x) => { return void (x + 1); })`},
			},
		},
	},
	{
		source:  `foo.forEach((x) => { if (a === b) { return x; } })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true, AllowVoid: true},
		wantIds: []string{"expectedNoReturnValue"},
		suggestions: [][]arrayCallbackReturnSuggestion{
			{
				{id: "prependVoid", output: `foo.forEach((x) => { if (a === b) { return void x; } })`},
			},
		},
	},
	{
		source:  `foo.forEach((x) => { if (a === b) { return !x; } })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true, AllowVoid: true},
		wantIds: []string{"expectedNoReturnValue"},
		suggestions: [][]arrayCallbackReturnSuggestion{
			{
				{id: "prependVoid", output: `foo.forEach((x) => { if (a === b) { return void !x; } })`},
			},
		},
	},
	{
		source:  `foo.forEach((x) => { if (a === b) { return (x + a); } })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true, AllowVoid: true},
		wantIds: []string{"expectedNoReturnValue"},
		suggestions: [][]arrayCallbackReturnSuggestion{
			{
				{id: "prependVoid", output: `foo.forEach((x) => { if (a === b) { return void (x + a); } })`},
			},
		},
	},
	{
		source:  `foo.filter(bar => { baz(); } )`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source: `foo.filter(
() => {} )`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.filter(bar || ((baz) => {}) )`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.filter(bar => { return; })`,
		options: nil,
		wantIds: []string{"expectedReturnValue"},
	},
	{
		source:  `Array.from(foo, bar => { bar })`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.forEach(bar => bar)`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
		suggestions: [][]arrayCallbackReturnSuggestion{
			{
				{id: "wrapBraces", output: `foo.forEach(bar => {bar})`},
			},
		},
	},
	{
		source:  `foo.forEach((function () { return (bar) => bar; })())`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
		suggestions: [][]arrayCallbackReturnSuggestion{
			{
				{id: "wrapBraces", output: `foo.forEach((function () { return (bar) => {bar}; })())`},
			},
		},
	},
	{
		source: `foo.forEach((() => {
 return bar => bar; })())`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
		suggestions: [][]arrayCallbackReturnSuggestion{
			{
				{id: "wrapBraces", output: `foo.forEach((() => {
 return bar => {bar}; })())`},
			},
		},
	},
	{
		source:  `foo.forEach((bar) => { if (bar) { return; } else { return bar ; } })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
	},
	{
		source:  `foo.filter(function(){})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.filter(function (){})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source: `foo.filter(function
(){})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.filter(function bar(){})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo.filter(function bar  (){})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source: `foo.filter(function
 bar() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `Array.from(foo, function bar(){})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `Array.from(foo, bar ? function (){} : baz)`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source: `foo.filter(function bar() { return 
 })`,
		options: nil,
		wantIds: []string{"expectedReturnValue"},
	},
	{
		source: `foo.forEach(function () { 
if (baz) return bar
else return
 })`,
		options: ArrayCallbackReturnOptions{CheckForEach: true},
		wantIds: []string{"expectedNoReturnValue"},
	},
	{
		source: `Array.fromAsync(x,
async	function \u0066oo // bar
   () {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo?.filter(() => { console.log('hello') })`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `(foo?.filter)(() => { console.log('hello') })`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `Array?.from([], () => { console.log('hello') })`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `(Array?.from)([], () => { console.log('hello') })`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `foo?.filter((function() { return () => { console.log('hello') } })?.())`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `Array.fromAsync(x, function() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `Array.fromAsync(x, async function() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `Array.fromAsync(x, function() {})`,
		options: ArrayCallbackReturnOptions{AllowImplicit: true},
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `Array.fromAsync(x, () => {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `Array.fromAsync(x, async () => {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
	{
		source:  `Array.fromAsync(x, function foo() {})`,
		options: nil,
		wantIds: []string{"expectedInside"},
	},
}

package typescript

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// preferForOfFile names the fixture file.
//
// The rule reads no path and gates on no extension. It ends in .ts rather than .tsx because several
// upstream cases use an angle-bracket type assertion, which is not parseable in a .tsx file at all.
const preferForOfFile = "/repository/source/Loops.ts"

// preferForOfCaseName numbers a row so a failure names which one.
func preferForOfCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// asTheTypedPreferForOfHarnessWroteIt transforms a fixture the way RunTyped transforms its input.
//
// rule_testing/program.go writes each fixture as strings.TrimSpace(contents)+"\n", so the file on
// disk is offset from the string in the Go literal. Every case in this corpus carries a leading
// newline and trailing indentation, so slicing the literal to check a span reports a result one byte
// short at each end, which reads exactly like an off-by-one in the rule.
func asTheTypedPreferForOfHarnessWroteIt(text string) string {
	return strings.TrimSpace(text) + "\n"
}

// TestPreferForOfStaysSilentOnUpstreamPassCases is the imported clean corpus, verbatim.
//
// All fifty-two of upstream's passing inputs, extracted from the clone's test file by parsing it
// with the TypeScript compiler rather than by reading it, then byte verified. Every one was replayed
// through the installed 8.x build, which reported nothing on all fifty-two.
//
// This list is most of the rule. A for loop that looks like an iteration and is not is the entire
// false-positive surface, and upstream has thought about far more of them than a port would invent:
// a second declarator, a non-zero start, a decrementing update, a call rather than a property named
// length, an optional chain in the test, an index used arithmetically, and thirteen separate shapes
// where the indexed element is being ASSIGNED rather than read, each of which makes a for-of
// rewrite wrong rather than merely different.
func TestPreferForOfStaysSilentOnUpstreamPassCases(t *testing.T) {
	t.Parallel()

	cases := []string{
		"\nfor (let i = 0; i < arr1.length; i++) {\n  const x = arr1[i] === arr2[i];\n}\n    ",
		"\nfor (let i = 0; i < arr.length; i++) {\n  arr[i] = 0;\n}\n    ",
		"\nfor (var c = 0; c < arr.length; c++) {\n  doMath(c);\n}\n    ",
		"\nfor (var d = 0; d < arr.length; d++) doMath(d);\n    ",
		"\nfor (var e = 0; e < arr.length; e++) {\n  if (e > 5) {\n    doMath(e);\n  }\n  console.log(arr[e]);\n}\n    ",
		"\nfor (var f = 0; f <= 40; f++) {\n  doMath(f);\n}\n    ",
		"\nfor (var g = 0; g <= 40; g++) doMath(g);\n    ",
		"\nfor (var h = 0, len = arr.length; h < len; h++) {}\n    ",
		"\nfor (var i = 0, len = arr.length; i < len; i++) arr[i];\n    ",
		"\nvar m = 0;\nfor (;;) {\n  if (m > 3) break;\n  console.log(m);\n  m++;\n}\n    ",
		"\nvar n = 0;\nfor (; n < 9; n++) {\n  console.log(n);\n}\n    ",
		"\nvar o = 0;\nfor (; o < arr.length; o++) {\n  console.log(arr[o]);\n}\n    ",
		"\nfor (; x < arr.length; x++) {}\n    ",
		"\nfor (let x = 0; ; x++) {}\n    ",
		"\nfor (let x = 0; x < arr.length;) {}\n    ",
		"\nfor (let x = 0; NOTX < arr.length; x++) {}\n    ",
		"\nfor (let x = 0; x < arr.length; NOTX++) {}\n    ",
		"\nfor (let NOTX = 0; x < arr.length; x++) {}\n    ",
		"\nfor (let x = 0; x < arr.length; x--) {}\n    ",
		"\nfor (let x = 0; x <= arr.length; x++) {}\n    ",
		"\nfor (let x = 1; x < arr.length; x++) {}\n    ",
		"\nfor (let x = 0; x < arr.length(); x++) {}\n    ",
		"\nfor (let x = 0; x < arr.length; x += 11) {}\n    ",
		"\nfor (let x = arr.length; x > 1; x -= 1) {}\n    ",
		"\nfor (let x = 0; x < arr.length; x *= 2) {}\n    ",
		"\nfor (let x = 0; x < arr.length; x = x + 11) {}\n    ",
		"\nfor (let x = 0; x < arr.length; x++) {\n  x++;\n}\n    ",
		"\nfor (let x = 0; true; x++) {}\n    ",
		"\nfor (var q in obj) {\n  if (obj.hasOwnProperty(q)) {\n    console.log(q);\n  }\n}\n    ",
		"\nfor (var r of arr) {\n  console.log(r);\n}\n    ",
		"\nfor (let x = 0; x < arr.length; x++) {\n  let y = arr[x + 1];\n}\n    ",
		"\nfor (let i = 0; i < arr.length; i++) {\n  delete arr[i];\n}\n    ",
		"\nfor (let i = 0; i < arr.length; i++) {\n  arr[i]++;\n}\n    ",
		"\nfor (let i = 0; i < arr.length; i++) {\n  arr[i]!++;\n}\n    ",
		"\nfor (let i = 0; i < arr.length; i++) {\n  arr[i]!!!++;\n}\n    ",
		"\nfor (let i = 0; i < arr.length; i++) {\n  (arr[i] as number)++;\n}\n    ",
		"\nfor (let i = 0; i < arr.length; i++) {\n  (<number>arr[i])++;\n}\n    ",
		"\nfor (let i = 0; i < arr.length; i++) {\n  (arr[i] as unknown as number)++;\n}\n    ",
		"\nfor (let i = 0; i < arr.length; i++) {\n  (arr[i] satisfies number)++;\n}\n    ",
		"\nfor (let i = 0; i < arr.length; i++) {\n  (arr[i]! satisfies number)++;\n}\n    ",
		"\nfor (let i = 0; i < arr.length; i++) {\n  [arr[i]] = [1];\n}\n    ",
		"\nfor (let i = 0; i < arr.length; i++) {\n  [...arr[i]] = [1];\n}\n    ",
		"\nfor (let i = 0; i < arr.length; i++) {\n  [...arr[i]!] = [1];\n}\n    ",
		"\nfor (let i = 0; i < arr.length; i++) {\n  ({ foo: arr[i] }) = { foo: 0 };\n}\n    ",
		"\nfor (let i = 0; i < arr1?.length; i++) {\n  const x = arr1[i] === arr2[i];\n}\n    ",
		"\nfor (let i = 0; i < arr?.length; i++) {\n  arr[i] = 0;\n}\n    ",
		"\nfor (var c = 0; c < arr?.length; c++) {\n  doMath(c);\n}\n    ",
		"\nfor (var d = 0; d < arr?.length; d++) doMath(d);\n    ",
		"\nfor (var c = 0; c < arr.length; c++) {\n  doMath?.(c);\n}\n    ",
		"\nfor (var d = 0; d < arr.length; d++) doMath?.(d);\n    ",
		"\nfor (let i = 0; i < test.length; ++i) {\n  this[i];\n}\n    ",
		"\nfor (let i = 0; i < this.length; ++i) {\n  yield this[i];\n}\n    ",
	}
	for index, sourceText := range cases {
		t.Run(preferForOfCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, PreferForOf,
				preferForOfFile, sourceText))
		})
	}
}

// TestPreferForOfFiresOnUpstreamFailCases is the imported failing corpus, verbatim, with the span of
// every finding asserted.
//
// The span is the WHOLE for statement, which is unusual enough to be worth pinning: it runs from the
// `for` keyword to the closing brace of the body, so a port anchoring on the initializer or on the
// index name satisfies every message id here while pointing at a fraction of what it means.
//
// One row reports twice, on a loop nested inside a loop where the inner index SHADOWS the outer. It
// is the case that separates resolving an identifier to its declaration from matching it by name.
func TestPreferForOfFiresOnUpstreamFailCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantIds    []string
		wantSpans  []string
	}{
		{
			sourceText: "\nfor (var a = 0; a < obj.arr.length; a++) {\n  console.log(obj.arr[a]);\n}\n      ",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (var a = 0; a < obj.arr.length; a++) {\n  console.log(obj.arr[a]);\n}"},
		},
		{
			sourceText: "\nfor (var b = 0; b < arr.length; b++) console.log(arr[b]);\n      ",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (var b = 0; b < arr.length; b++) console.log(arr[b]);"},
		},
		{
			sourceText: "\nfor (let a = 0; a < arr.length; a++) {\n  console.log(arr[a]);\n}\n      ",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let a = 0; a < arr.length; a++) {\n  console.log(arr[a]);\n}"},
		},
		{
			sourceText: "\nfor (var b = 0; b < arr.length; b++) console?.log(arr[b]);\n      ",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (var b = 0; b < arr.length; b++) console?.log(arr[b]);"},
		},
		{
			sourceText: "\nfor (let a = 0; a < arr.length; a++) {\n  console?.log(arr[a]);\n}\n      ",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let a = 0; a < arr.length; a++) {\n  console?.log(arr[a]);\n}"},
		},
		{
			sourceText: "\nfor (let a = 0; a < arr.length; ++a) {\n  arr[a].whatever();\n}\n      ",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let a = 0; a < arr.length; ++a) {\n  arr[a].whatever();\n}"},
		},
		{
			sourceText: "\nfor (let x = 0; x < arr.length; x++) {}\n      ",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let x = 0; x < arr.length; x++) {}"},
		},
		{
			sourceText: "\nfor (let x = 0; x < arr.length; x += 1) {}\n      ",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let x = 0; x < arr.length; x += 1) {}"},
		},
		{
			sourceText: "\nfor (let x = 0; x < arr.length; x = x + 1) {}\n      ",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let x = 0; x < arr.length; x = x + 1) {}"},
		},
		{
			sourceText: "\nfor (let x = 0; x < arr.length; x = 1 + x) {}\n      ",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let x = 0; x < arr.length; x = 1 + x) {}"},
		},
		{
			sourceText: "\nfor (let shadow = 0; shadow < arr.length; shadow++) {\n  for (let shadow = 0; shadow < arr.length; shadow++) {}\n}\n      ",
			wantIds:    []string{"preferForOf", "preferForOf"},
			wantSpans:  []string{"for (let shadow = 0; shadow < arr.length; shadow++) {\n  for (let shadow = 0; shadow < arr.length; shadow++) {}\n}", "for (let shadow = 0; shadow < arr.length; shadow++) {}"},
		},
		{
			sourceText: "\nfor (let i = 0; i < arr.length; i++) {\n  obj[arr[i]] = 1;\n}\n      ",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  obj[arr[i]] = 1;\n}"},
		},
		{
			sourceText: "\nfor (let i = 0; i < arr.length; i++) {\n  delete obj[arr[i]];\n}\n      ",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  delete obj[arr[i]];\n}"},
		},
		{
			sourceText: "\nfor (let i = 0; i < arr.length; i++) {\n  [obj[arr[i]]] = [1];\n}\n      ",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  [obj[arr[i]]] = [1];\n}"},
		},
		{
			sourceText: "\nfor (let i = 0; i < arr.length; i++) {\n  [...obj[arr[i]]] = [1];\n}\n      ",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  [...obj[arr[i]]] = [1];\n}"},
		},
		{
			sourceText: "\nfor (let i = 0; i < arr.length; i++) {\n  ({ foo: obj[arr[i]] } = { foo: 1 });\n}\n      ",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  ({ foo: obj[arr[i]] } = { foo: 1 });\n}"},
		},
		{
			sourceText: "\nfor (let i = 0; i < this.item.length; ++i) {\n  this.item[i];\n}\n      ",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < this.item.length; ++i) {\n  this.item[i];\n}"},
		},
		{
			sourceText: "\nfor (let i = 0; i < this.array.length; ++i) {\n  yield this.array[i];\n}\n      ",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < this.array.length; ++i) {\n  yield this.array[i];\n}"},
		},
	}
	for index, testCase := range cases {
		t.Run(preferForOfCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PreferForOf, preferForOfFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d", len(testCase.wantSpans), len(result.Diagnostics))
			}
			written := asTheTypedPreferForOfHarnessWroteIt(testCase.sourceText)
			for findingIndex, wantSpan := range testCase.wantSpans {
				reported := result.Diagnostics[findingIndex]
				gotSpan := written[reported.Range.Pos():reported.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d points at %q, wanted %q", findingIndex, gotSpan, wantSpan)
				}
			}
		})
	}
}

// TestPreferForOfOnShapesUpstreamsCorpusCannotSeparate covers three guards that upstream tests only
// in combination with something else that already declines.
//
// Upstream has fifty-two passing cases and they are unusually thorough, so it is worth being precise
// about what is missing. Nothing here is a shape upstream forgot. Each is a shape upstream writes
// while ALSO tripping a second test, so removing the guard under examination leaves every one of the
// seventy imported cases green. Its four optional-chain rows each additionally use two receivers,
// assign to the element, or use the index directly; its two-declarator rows compare against the
// second declarator rather than against a length; and it writes no const index at all.
//
// Found by mutation rather than by reading: three separate guards survived the whole imported
// corpus. Every verdict below was then measured against the installed 8.x build, with a reporting
// control in the same run.
func TestPreferForOfOnShapesUpstreamsCorpusCannotSeparate(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		why        string
		sourceText string
		wantIds    []string
		wantSpans  []string
	}{
		{
			name:       "optional-length-with-clean-body",
			why:        "an optional access in the test with a body that would otherwise report. Upstream writes four optional-chain cases and every one of them ALSO fails another test, so its own corpus cannot tell the optional-chain rejection from a rule that never had it, and a mutation removing that rejection survives all seventy imported rows",
			sourceText: "for (let i = 0; i < arr?.length; i++) {\n  console.log(arr[i]);\n}\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "optional-length-var",
			why:        "the same through a var index, since the optional rejection sits above the declaration-kind test and a port could have put it below",
			sourceText: "for (var i = 0; i < arr?.length; i++) {\n  console.log(arr[i]);\n}\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "const-index",
			why:        "a const index, which cannot be incremented at runtime and which upstream's corpus never writes; without the const test this reports and upstream does not",
			sourceText: "for (const i = 0; i < arr.length; i++) {\n  console.log(arr[i]);\n}\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "two-declarators-length-test",
			why:        "two declarators with a real length test. Upstream's two-declarator cases all compare against the SECOND declarator rather than against a length, so the test declines them first and nothing imported can see the single-declarator requirement",
			sourceText: "for (var h = 0, len = 1; h < arr.length; h++) {\n  console.log(arr[h]);\n}\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "two-declarators-both-used",
			why:        "a second declarator that is never mentioned again, which is the weakest form of the row above and still silent",
			sourceText: "for (var h = 0, k = 0; h < arr.length; h++) {\n  console.log(arr[h]);\n}\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "property-named-size-not-length",
			why:        "the property in the test must be named length. Upstream's nearest case tests against a length CALL rather than against a differently named property, so a port accepting any property name survives all seventy imported rows; a for-of over an array bounded by anything else is not the same loop",
			sourceText: "for (let i = 0; i < arr.size; i++) {\n  console.log(arr[i]);\n}\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "property-named-count",
			why:        "a second spelling of the row above, so a port special-casing one name is not satisfied by it",
			sourceText: "for (let i = 0; i < arr.count; i++) {\n  console.log(arr[i]);\n}\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "control-p6",
			why:        "the control: the same loop bounded by length, which reports",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  console.log(arr[i]);\n}\n",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  console.log(arr[i]);\n}"},
		},
		{
			name:       "index-as-receiver",
			why:        "the index is the RECEIVER of an access rather than its subscript, which upstream declines and its corpus never writes. Upstream expresses this as property identity because a computed and a dotted access are one node type there; ours are separate kinds, so the equivalent test is which side of the element access the identifier sits on, and without it this reports",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  i[arr];\n}\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "index-both-sides",
			why:        "the same identifier used correctly once and incorrectly once, which a port checking only the first reference would pass",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  arr[i];\n  i[arr];\n}\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "control-p12",
			why:        "the control: the same body without the receiver use, which reports",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  arr[i];\n}\n",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  arr[i];\n}"},
		},
		{
			name:       "control-pfo",
			why:        "the control: the plainest reporting shape, so a rule that had stopped reporting entirely would still be visible here",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  console.log(arr[i]);\n}\n",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  console.log(arr[i]);\n}"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PreferForOf, preferForOfFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d, on the case that covers: %s",
					len(testCase.wantSpans), len(result.Diagnostics), testCase.why)
			}
			written := asTheTypedPreferForOfHarnessWroteIt(testCase.sourceText)
			for findingIndex, wantSpan := range testCase.wantSpans {
				reported := result.Diagnostics[findingIndex]
				gotSpan := written[reported.Range.Pos():reported.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d points at %q, wanted %q (%s)",
						findingIndex, gotSpan, wantSpan, testCase.why)
				}
			}
		})
	}
}

// TestPreferForOfDoesNotMistakeAReadForAWrite covers the false positives our flatter assignment
// spelling makes available and upstream's does not.
//
// The rule reports a loop only when every use of the index is a plain READ of one array element, so
// every one of these nine rows must report. They exist because upstream reads distinct node types
// where our parser uses one kind plus an operator: AssignmentExpression against every other binary
// operator, and UpdateExpression against every other prefix operator. Upstream literally cannot
// confuse `arr[i] < 1` with `arr[i] = 1`, so its corpus never separates them, and four mutations
// weakening those tests survived all seventy imported rows.
//
// A port failing any of these does not over-report, it goes SILENT on ordinary reading loops, which
// is the failure a rule about readability can afford least and which no amount of clean-case
// coverage would reveal.
//
// Measured against the installed 8.x build, all ten including the control reporting.
func TestPreferForOfDoesNotMistakeAReadForAWrite(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		why        string
		sourceText string
		wantIds    []string
		wantSpans  []string
	}{
		{
			name:       "index-read-in-comparison",
			why:        "the indexed element on the left of a COMPARISON rather than an assignment. Our parser files every binary operator under one kind, so the assignee test has to ask whether the operator assigns; upstream reads a distinct AssignmentExpression node and cannot make this mistake. Nothing in its corpus puts an indexed read on the left of a non-assignment",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  if (arr[i] < 1) {\n    doThing();\n  }\n}\n",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  if (arr[i] < 1) {\n    doThing();\n  }\n}"},
		},
		{
			name:       "index-read-in-addition",
			why:        "the same through an arithmetic operator rather than a relational one",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  const x = arr[i] + 1;\n}\n",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  const x = arr[i] + 1;\n}"},
		},
		{
			name:       "index-read-on-assignment-right",
			why:        "the indexed element on the RIGHT of an assignment, which is a read; without the left-side test this reads as a write and the loop goes silent",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  other = arr[i];\n}\n",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  other = arr[i];\n}"},
		},
		{
			name:       "index-read-on-compound-right",
			why:        "the same through a compound assignment, where the left side is a different variable entirely",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  other += arr[i];\n}\n",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  other += arr[i];\n}"},
		},
		{
			name:       "index-read-in-plain-object",
			why:        "an indexed read inside an object literal that is NOT being destructured into. Upstream's corpus has the destructuring form and not this one, so the test requiring the object itself be an assignee is unexercised there and a port without it goes silent on every object literal holding an indexed read",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  const o = { foo: arr[i] };\n}\n",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  const o = { foo: arr[i] };\n}"},
		},
		{
			name:       "index-read-in-nested-object",
			why:        "the same object literal passed as an argument, which is the shape this appears in most often in real code",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  send({ foo: arr[i] });\n}\n",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  send({ foo: arr[i] });\n}"},
		},
		{
			name:       "index-read-under-logical-not",
			why:        "a prefix operator that is not an update. Upstream reads a distinct UpdateExpression; ours puts logical not, unary minus and the increment operators on one kind, so the operator has to be checked or every negated indexed read reads as a write",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  if (!arr[i]) {\n    doThing();\n  }\n}\n",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  if (!arr[i]) {\n    doThing();\n  }\n}"},
		},
		{
			name:       "index-read-under-unary-minus",
			why:        "the arithmetic negation form of the row above",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  const x = -arr[i];\n}\n",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  const x = -arr[i];\n}"},
		},
		{
			name:       "index-read-under-typeof",
			why:        "the typeof form, which is a third prefix operator on the same node kind here",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  const x = typeof arr[i];\n}\n",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  const x = typeof arr[i];\n}"},
		},
		{
			name:       "index-as-computed-property-key",
			why:        "the indexed element is the KEY of a computed property rather than its value, which is a read; upstream's corpus writes the value position only, so the test requiring the node be the initializer is unexercised and a port without it goes silent here",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  ({ [arr[i]]: 1 });\n}\n",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  ({ [arr[i]]: 1 });\n}"},
		},
		{
			name:       "index-as-computed-key-assigned",
			why:        "the same key position inside an object that IS being destructured into, which is the row that separates the initializer test from the assignee recursion below it",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  ({ [arr[i]]: x } = y);\n}\n",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  ({ [arr[i]]: x } = y);\n}"},
		},
		{
			name:       "control-p25",
			why:        "the control, and it is upstream's own destructuring case with the outer parentheses it is normally written without, so this row also pins that the parenthesis arm passes the question through in the object direction",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  ({ foo: arr[i] } = { foo: 0 });\n}\n",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  ({ foo: arr[i] } = { foo: 0 });\n}"},
		},
		{
			name:       "control-p21",
			why:        "the control: a plain indexed read, so a rule that had stopped reporting would still be visible",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  console.log(arr[i]);\n}\n",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  console.log(arr[i]);\n}"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PreferForOf, preferForOfFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d, on the case that covers: %s",
					len(testCase.wantSpans), len(result.Diagnostics), testCase.why)
			}
			written := asTheTypedPreferForOfHarnessWroteIt(testCase.sourceText)
			for findingIndex, wantSpan := range testCase.wantSpans {
				reported := result.Diagnostics[findingIndex]
				gotSpan := written[reported.Range.Pos():reported.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d points at %q, wanted %q (%s)",
						findingIndex, gotSpan, wantSpan, testCase.why)
				}
			}
		})
	}
}

// TestPreferForOfRendersUpstreamsMessageText asserts what a reader is told.
//
// rule.Message is {Id, Description} with no interpolation, so there is nothing to render and nothing
// a format string could get wrong. What there is to get wrong is the text, and every other fixture
// here goes through ExpectFindings, which compares ids and count and nothing else; a mutation
// rewriting the description survived the whole suite until this existed. The wanted string is typed
// as a literal rather than read from the rule's own constant, because a comparison against the
// constant moves with any mutation of it.
func TestPreferForOfRendersUpstreamsMessageText(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, PreferForOf, preferForOfFile,
		"for (let i = 0; i < arr.length; i++) {\n  console.log(arr[i]);\n}\n")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
	}
	reported := result.Diagnostics[0]
	if reported.Message.Id != "preferForOf" {
		t.Errorf("reported id %q, wanted %q", reported.Message.Id, "preferForOf")
	}
	const wantMessage = "Expected a `for-of` loop instead of a `for` loop with this simple iteration."
	if reported.Message.Description != wantMessage {
		t.Errorf("reported message %q, wanted %q", reported.Message.Description, wantMessage)
	}
}

// TestPreferForOfRequiresTheTypedHarness pins the checker guard.
//
// The listener starts with a nil check, and the shim answers nil from GetSymbolAtLocation on a nil
// checker rather than panicking, so a rule missing that guard does not crash. It goes silent, or
// worse: with no symbol to resolve, every body reference fails to match the declarator, the loop
// looks like one whose index is never used, and the rule reports every counted loop in the tree.
//
// This asserts the untyped harness produces nothing on an input the typed one reports, so a later
// revert to rule_testing.Run fails loudly rather than quietly.
func TestPreferForOfRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	const sourceText = "for (let i = 0; i < arr.length; i++) {\n  console.log(arr[i]);\n}\n"
	rule_testing.ExpectClean(t, rule_testing.Run(t, PreferForOf, preferForOfFile, sourceText))
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, PreferForOf,
		preferForOfFile, sourceText), "preferForOf")
}

// TestPreferForOfReproducesUpstreamsObjectPatternAsymmetry pins a behavior that looks like a defect
// upstream and is reproduced rather than corrected.
//
// `isAssignee` requires the enclosing object be an `ObjectExpression`, and upstream's parser gives a
// REAL destructuring target the distinct type `ObjectPattern`. So the arm fires on the spelling
// where the parentheses sit inside the assignment, which is not a destructuring at all, and does
// NOT fire on the spelling that actually writes through the indexed element. Measured: with the
// parentheses around the whole assignment the loop reports, and with them around the object alone it
// is silent, which is the opposite of what the two shapes do at runtime.
//
// Our parser has no ObjectPattern, so both spellings arrive as an object literal and the difference
// has to be read off the position instead. Upstream's corpus contains only one of the two, so
// nothing imported can see this; it was found by measuring the other spelling and it changed the
// port.
func TestPreferForOfReproducesUpstreamsObjectPatternAsymmetry(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		why        string
		sourceText string
		wantIds    []string
		wantSpans  []string
	}{
		{
			name:       "object-literal-compared-not-assigned",
			why:        "an object literal on the left of a COMPARISON, which is not a destructuring target and where the indexed element is therefore a read",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  ({ foo: arr[i] }) === x;\n}\n",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  ({ foo: arr[i] }) === x;\n}"},
		},
		{
			name:       "object-literal-on-right-of-assignment",
			why:        "an object literal on the RIGHT of an assignment, which is also a read; upstream's parser types it as an ObjectExpression either way and the position is what decides",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  x = { foo: arr[i] };\n}\n",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  x = { foo: arr[i] };\n}"},
		},
		{
			name:       "object-literal-compound-assigned",
			why:        "a compound assignment, which is silent upstream and which is what pins the destructuring test to the plain equals token: only a plain assignment makes an object literal a pattern, so a test accepting any operator declines this and the loop stops reporting",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  ({ foo: arr[i] }) += x;\n}\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "bare-object-compound-assigned",
			why:        "a BARE object literal, no parentheses of its own, on the left of a COMPOUND assignment. This is the only shape that reaches the destructuring test with a non-equals operator, and it is what pins that test to the plain equals token: a compound assignment does not make an object literal a pattern, so upstream still types it ObjectExpression and reports, while a test accepting any assignment operator calls it a pattern and goes silent",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  ({ foo: arr[i] } += x);\n}\n",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  ({ foo: arr[i] } += x);\n}"},
		},
		{
			name:       "bare-object-plain-assigned",
			why:        "the plain-equals half, which upstream also reports because there the object IS a pattern and its isAssignee arm therefore declines. Both spellings reporting for opposite reasons is the whole shape of upstream's asymmetry",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  ({ foo: arr[i] } = { foo: 0 });\n}\n",
			wantIds:    []string{"preferForOf"},
			wantSpans:  []string{"for (let i = 0; i < arr.length; i++) {\n  ({ foo: arr[i] } = { foo: 0 });\n}"},
		},
		{
			name:       "control-p30",
			why:        "upstream's own parenthesized-object shape, silent, which is the row the three above are read against",
			sourceText: "for (let i = 0; i < arr.length; i++) {\n  ({ foo: arr[i] }) = { foo: 0 };\n}\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PreferForOf, preferForOfFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d, on the case that covers: %s",
					len(testCase.wantSpans), len(result.Diagnostics), testCase.why)
			}
			written := asTheTypedPreferForOfHarnessWroteIt(testCase.sourceText)
			for findingIndex, wantSpan := range testCase.wantSpans {
				reported := result.Diagnostics[findingIndex]
				gotSpan := written[reported.Range.Pos():reported.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d points at %q, wanted %q (%s)",
						findingIndex, gotSpan, wantSpan, testCase.why)
				}
			}
		})
	}
}

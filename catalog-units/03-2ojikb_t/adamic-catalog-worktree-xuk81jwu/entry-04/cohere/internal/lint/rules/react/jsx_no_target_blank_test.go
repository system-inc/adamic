package react

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// jsxNoTargetBlankFile is where the fixtures pretend to live.
//
// A `.tsx` name because every case is JSX and needs a parser that reads it. It is NOT load-bearing:
// this rule has no file gate, and `TestJsxNoTargetBlankHasNoFileGate` pins that by running one
// reporting source under three JSX-capable extensions.
const jsxNoTargetBlankFile = "/repository/source/JsxNoTargetBlank.tsx"

// The corpus is `eslint-plugin-react`'s own, extracted rather than retyped.
//
// `/tmp/lint-sources/eslint-plugin-react/tests/lib/rules/jsx-no-target-blank.js` holds 63 valid and
// 50 invalid cases. Every string below was pulled out of that file by loading it with a stubbed
// `RuleTester` and serializing the captured object, then verified byte against byte by extracting
// the same file a second time through an independent path that writes raw bytes rather than JSON.
// All 113 codes were byte-identical across the two extractions.
//
// All 113 were then run against the installed build, version 7.37.5, through the ESLint Linter API,
// and its verdict agreed with the corpus on every one: 63 clean and 50 reporting with the stated
// counts and message ids. So the two authorities are one here.
//
// # Five cases carry `linkComponents` settings, and all five go clean without them
//
// `linkComponentsUtil` starts from `['a']` and concatenates `settings.linkComponents`. cohere
// reads no React settings, so `<Link>` is not a link component here by any route.
// Each of the five was re-run against the installed build with its settings REMOVED and every one
// went clean, which is the answer this port has to produce. They are recorded as clean fixtures
// with the note at the line.
//
// # The options column is RAW JSON, on purpose
//
// Two of the five options have a default that is not Go's zero value: `enforceDynamicLinks` is a
// string enum whose absent value means `always`, and `links` defaults to true. Handing
// `RunWithOptions` a struct built by hand would leave the decoder, the enum inversion, and the
// `links` default entirely untested, which are the three lines with no upstream counterpart and
// therefore the three most likely to be wrong. Every case below routes through
// `DecodeJsxNoTargetBlankOptions`, the same function the config calls.
//
// An empty string in that column means the rule was configured as a bare severity, which is what
// hands a real rule nil options.

// TestJsxNoTargetBlankFires asserts ids, count, AND the repaired source where upstream ships one.
//
// 26 of the 48 reporting cases carry an `output`, which is the exact text upstream's fixer writes.
// A finding with a wrong repair reads as a correct finding to any id assertion, so the fix column
// is checked with `ExpectFixedSource`, which compares the whole rewritten file. The 22 with no
// output are cases upstream reports and declines to fix, and their empty column asserts that
// decline rather than skipping the question.
func TestJsxNoTargetBlankFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
		wantIds    []string
		wantFixed  string
	}{
		{"upstream invalid-0", "<a target=\"_blank\" href=\"https://example.com/1\"></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target=\"_blank\" href=\"https://example.com/1\" rel=\"noreferrer\"></a>"},
		{"upstream invalid-1", "<a target=\"_blank\" rel=\"\" href=\"https://example.com/2\"></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target=\"_blank\" rel=\"noreferrer\" href=\"https://example.com/2\"></a>"},
		{"upstream invalid-2", "<a target=\"_blank\" rel={0} href=\"https://example.com/3\"></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target=\"_blank\" rel=\"noreferrer\" href=\"https://example.com/3\"></a>"},
		{"upstream invalid-3", "<a target=\"_blank\" rel={1} href=\"https://example.com/3\"></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target=\"_blank\" rel=\"noreferrer\" href=\"https://example.com/3\"></a>"},
		{"upstream invalid-4", "<a target=\"_blank\" rel={false} href=\"https://example.com/4\"></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target=\"_blank\" rel=\"noreferrer\" href=\"https://example.com/4\"></a>"},
		{"upstream invalid-5", "<a target=\"_blank\" rel={null} href=\"https://example.com/5\"></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target=\"_blank\" rel=\"noreferrer\" href=\"https://example.com/5\"></a>"},
		{"upstream invalid-6", "<a target=\"_blank\" rel=\"noopenernoreferrer\" href=\"https://example.com/6\"></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target=\"_blank\" rel=\"noopener noreferrer\" href=\"https://example.com/6\"></a>"},
		{"upstream invalid-7", "<a target=\"_blank\" rel=\"no referrer\" href=\"https://example.com/7\"></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target=\"_blank\" rel=\"no referrer noreferrer\" href=\"https://example.com/7\"></a>"},
		{"upstream invalid-8", "<a target=\"_BLANK\" href=\"https://example.com/8\"></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target=\"_BLANK\" href=\"https://example.com/8\" rel=\"noreferrer\"></a>"},
		{"upstream invalid-9", "<a target=\"_blank\" href=\"//example.com/9\"></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target=\"_blank\" href=\"//example.com/9\" rel=\"noreferrer\"></a>"},
		{"upstream invalid-10", "<a target=\"_blank\" href=\"//example.com/10\" rel={true}></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target=\"_blank\" href=\"//example.com/10\" rel=\"noreferrer\"></a>"},
		{"upstream invalid-11", "<a target=\"_blank\" href=\"//example.com/11\" rel={3}></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target=\"_blank\" href=\"//example.com/11\" rel=\"noreferrer\"></a>"},
		{"upstream invalid-12", "<a target=\"_blank\" href=\"//example.com/12\" rel={null}></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target=\"_blank\" href=\"//example.com/12\" rel=\"noreferrer\"></a>"},
		{"upstream invalid-13", "<a target=\"_blank\" href=\"//example.com/13\" rel={getRel()}></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, ""},
		{"upstream invalid-14", "<a target=\"_blank\" href=\"//example.com/14\" rel={\"noopenernoreferrer\"}></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target=\"_blank\" href=\"//example.com/14\" rel={\"noopener noreferrer\"}></a>"},
		{"upstream invalid-15", "<a target={\"_blank\"} href={\"//example.com/15\"} rel={\"noopenernoreferrer\"}></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target={\"_blank\"} href={\"//example.com/15\"} rel={\"noopener noreferrer\"}></a>"},
		{"upstream invalid-16", "<a target={\"_blank\"} href={\"//example.com/16\"} rel={\"noopenernoreferrernoreferrernoreferrernoreferrernoreferrer\"}></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target={\"_blank\"} href={\"//example.com/16\"} rel={\"noopener noreferrer\"}></a>"},
		{"upstream invalid-17", "<a target=\"_blank\" href=\"//example.com/17\" rel></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target=\"_blank\" href=\"//example.com/17\" rel=\"noreferrer\"></a>"},
		{"upstream invalid-18", "<a target=\"_blank\" href={ dynamicLink }></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target=\"_blank\" href={ dynamicLink } rel=\"noreferrer\"></a>"},
		{"upstream invalid-19", "<a target={'_blank'} href=\"//example.com/18\"></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target={'_blank'} href=\"//example.com/18\" rel=\"noreferrer\"></a>"},
		{"upstream invalid-20", "<a target={\"_blank\"} href=\"//example.com/19\"></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target={\"_blank\"} href=\"//example.com/19\" rel=\"noreferrer\"></a>"},
		{"upstream invalid-21", "<a href=\"https://example.com/20\" target=\"_blank\" rel></a>", "{\"allowReferrer\": true}", []string{"noTargetBlankWithoutNoopener"}, "<a href=\"https://example.com/20\" target=\"_blank\" rel=\"noopener\"></a>"},
		{"upstream invalid-22", "<a href=\"https://example.com/20\" target=\"_blank\"></a>", "{\"allowReferrer\": true}", []string{"noTargetBlankWithoutNoopener"}, "<a href=\"https://example.com/20\" target=\"_blank\" rel=\"noopener\"></a>"},
		{"upstream invalid-23", "<a target=\"_blank\" href={ dynamicLink }></a>", "{\"enforceDynamicLinks\": \"always\"}", []string{"noTargetBlankWithoutNoreferrer"}, "<a target=\"_blank\" href={ dynamicLink } rel=\"noreferrer\"></a>"},
		{"upstream invalid-24", "<a {...someObject}></a>", "{\"enforceDynamicLinks\": \"always\", \"warnOnSpreadAttributes\": true}", []string{"noTargetBlankWithoutNoreferrer"}, ""},
		{"upstream invalid-25", "<a {...someObject} target=\"_blank\"></a>", "{\"enforceDynamicLinks\": \"always\", \"warnOnSpreadAttributes\": true}", []string{"noTargetBlankWithoutNoreferrer"}, ""},
		{"upstream invalid-26", "<a href=\"foobar\" {...someObject} target=\"_blank\"></a>", "{\"enforceDynamicLinks\": \"always\", \"warnOnSpreadAttributes\": true}", []string{"noTargetBlankWithoutNoreferrer"}, ""},
		{"upstream invalid-27", "<a href=\"foobar\" target=\"_blank\" rel=\"noreferrer\" {...someObject}></a>", "{\"enforceDynamicLinks\": \"always\", \"warnOnSpreadAttributes\": true}", []string{"noTargetBlankWithoutNoreferrer"}, ""},
		{"upstream invalid-28", "<a href=\"foobar\" target=\"_blank\" {...someObject}></a>", "{\"enforceDynamicLinks\": \"always\", \"warnOnSpreadAttributes\": true}", []string{"noTargetBlankWithoutNoreferrer"}, ""},
		{"upstream invalid-31", "<a href=\"some-link\" {...otherProps} target=\"some-non-blank-target\"></a>", "{\"warnOnSpreadAttributes\": true}", []string{"noTargetBlankWithoutNoreferrer"}, ""},
		{"upstream invalid-32", "<a href=\"some-link\" target=\"some-non-blank-target\" {...otherProps}></a>", "{\"warnOnSpreadAttributes\": true}", []string{"noTargetBlankWithoutNoreferrer"}, ""},
		{"upstream invalid-33", "<a target=\"_blank\" href=\"//example.com\" rel></a>", "{\"links\": true}", []string{"noTargetBlankWithoutNoreferrer"}, "<a target=\"_blank\" href=\"//example.com\" rel=\"noreferrer\"></a>"},
		{"upstream invalid-34", "<a target=\"_blank\" href=\"//example.com\" rel></a>", "{\"links\": true, \"forms\": true}", []string{"noTargetBlankWithoutNoreferrer"}, "<a target=\"_blank\" href=\"//example.com\" rel=\"noreferrer\"></a>"},
		{"upstream invalid-35", "<a target=\"_blank\" href=\"//example.com\" rel></a>", "{\"links\": true, \"forms\": false}", []string{"noTargetBlankWithoutNoreferrer"}, "<a target=\"_blank\" href=\"//example.com\" rel=\"noreferrer\"></a>"},
		{"upstream invalid-36", "<form method=\"POST\" action=\"https://example.com\" target=\"_blank\"></form>", "{\"forms\": true}", []string{"noTargetBlankWithoutNoreferrer"}, ""},
		{"upstream invalid-37", "<form method=\"POST\" action=\"https://example.com\" rel=\"\" target=\"_blank\"></form>", "{\"forms\": true}", []string{"noTargetBlankWithoutNoreferrer"}, ""},
		{"upstream invalid-38", "<form method=\"POST\" action=\"https://example.com\" rel=\"noopenernoreferrer\" target=\"_blank\"></form>", "{\"forms\": true}", []string{"noTargetBlankWithoutNoreferrer"}, ""},
		{"upstream invalid-39", "<form method=\"POST\" action=\"https://example.com\" rel=\"noopenernoreferrer\" target=\"_blank\"></form>", "{\"forms\": true, \"links\": false}", []string{"noTargetBlankWithoutNoreferrer"}, ""},
		{"upstream invalid-40", "<a href={href} target=\"_blank\" rel={isExternal ? \"undefined\" : \"undefined\"} />", "", []string{"noTargetBlankWithoutNoreferrer"}, ""},
		{"upstream invalid-41", "<a href={href} target=\"_blank\" rel={isExternal ? \"noopener\" : undefined} />", "", []string{"noTargetBlankWithoutNoreferrer"}, ""},
		{"upstream invalid-42", "<a href={href} target=\"_blank\" rel={isExternal ? \"undefined\" : \"noopener\"} />", "", []string{"noTargetBlankWithoutNoreferrer"}, ""},
		{"upstream invalid-43", "<a href={href} target={isExternal ? \"_blank\" : undefined} rel={isExternal ? undefined : \"noopener noreferrer\"} />", "", []string{"noTargetBlankWithoutNoreferrer"}, ""},
		{"upstream invalid-44", "<a href={href} target=\"_blank\" rel={isExternal ? 3 : \"noopener noreferrer\"} />", "", []string{"noTargetBlankWithoutNoreferrer"}, ""},
		{"upstream invalid-45", "<a href={href} target=\"_blank\" rel={isExternal ? \"noopener noreferrer\" : \"3\"} />", "", []string{"noTargetBlankWithoutNoreferrer"}, ""},
		{"upstream invalid-46", "<a href={href} target=\"_blank\" rel={isExternal ? \"noopener\" : \"2\"} />", "{\"allowReferrer\": true}", []string{"noTargetBlankWithoutNoopener"}, ""},
		{"upstream invalid-47", "<form action={action} target=\"_blank\" />", "{\"allowReferrer\": true, \"forms\": true}", []string{"noTargetBlankWithoutNoopener"}, ""},
		{"upstream invalid-48", "<form action={action} target=\"_blank\" />", "{\"forms\": true}", []string{"noTargetBlankWithoutNoreferrer"}, ""},
		{"upstream invalid-49", "<form action={action} {...spread} />", "{\"forms\": true, \"warnOnSpreadAttributes\": true}", []string{"noTargetBlankWithoutNoreferrer"}, ""},

		// --- cases upstream does not cover, each measured against the installed build 7.37.5 ---

		// The `_blank` comparison is case-INSENSITIVE; upstream lowercases before comparing.
		{"t1", "<a target='_BLANK' href='http://x.com'></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target='_BLANK' href='http://x.com' rel=\"noreferrer\"></a>"},

		// A `mailto:` scheme matches `/^(?:\w+:|\/\/)/` and counts as external. Reads as over-broad
		// and is upstream's judgment; `tel:` below is the same mechanism.
		{"t2", "<a target='_blank' href='mailto:a@b.c'></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target='_blank' href='mailto:a@b.c' rel=\"noreferrer\"></a>"},

		// A protocol-relative href, the second half of the same pattern.
		{"t3", "<a target='_blank' href='//example.com'></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target='_blank' href='//example.com' rel=\"noreferrer\"></a>"},

		// The LAST target decides. `findLastIndex`, not `indexOf`, and a first-match search would call
		// this clean.
		{"t4", "<a target='_self' target='_blank' href='http://x.com'></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target='_self' target='_blank' href='http://x.com' rel=\"noreferrer\"></a>"},

		// The judgment reads the LAST rel and the fixer repairs the FIRST. Upstream uses
		// `findLastIndex` for one and `find` for the other, so this reports on the trailing `rel="x"`
		// and rewrites the leading `rel="noreferrer"`. Reproduced rather than unified; the fix column
		// is what pins which one moved.
		{"t6", "<a href='http://x.com' target='_blank' rel='noreferrer' rel='x'></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a href='http://x.com' target='_blank' rel=\"noreferrer\" rel='x'></a>"},

		// A template literal with a substitution: only `quasis[0]` is read, which is empty here, so the
		// trailing `noreferrer` is invisible and this reports. Its substitution-free twin is clean below.
		{"t8", "<a href='http://x.com' target='_blank' rel={`${x} noreferrer`}></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, ""},

		// The conditional pairing, wrong branch. Same test as the target, and the branch that pairs with
		// `_blank` is the insecure one.
		{"t12", "<a href='http://x.com' target={c ? '_blank' : '_self'} rel={c ? 'x' : 'noreferrer'}></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, ""},

		// The conditional pairing declined: the tests differ, so BOTH rel branches must be secure.
		{"t13", "<a href='http://x.com' target={c ? '_blank' : '_self'} rel={d ? 'noreferrer' : 'x'}></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, ""},

		// `rel={undefined}` is an IDENTIFIER in JavaScript, not a literal, so it misses upstream's
		// non-string-literal branch and gets no fix. Upstream's own comment names undefined in that
		// branch and the code cannot reach it. The three literals below do get the fix.
		{"t14", "<a href='http://x.com' target='_blank' rel={undefined}></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, ""},
		{"t15", "<a href='http://x.com' target='_blank' rel={null}></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a href='http://x.com' target='_blank' rel=\"noreferrer\"></a>"},
		{"t16", "<a href='http://x.com' target='_blank' rel={5}></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a href='http://x.com' target='_blank' rel=\"noreferrer\"></a>"},
		{"t17", "<a href='http://x.com' target='_blank' rel={true}></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a href='http://x.com' target='_blank' rel=\"noreferrer\"></a>"},

		// A bare `rel` with no value: the fix inserts `="noreferrer"` after the attribute.
		{"t18", "<a href='http://x.com' target='_blank' rel></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a href='http://x.com' target='_blank' rel=\"noreferrer\"></a>"},

		// `links: false` changes nothing. `configuration.links` is set by the rule and read nowhere in
		// its body; both spellings and no options at all report identically. Upstream defect,
		// reproduced, and this pair is what would fail if a later reader wired it up.
		{"t19", "<a target='_blank' href='http://x.com'></a>", "{\"links\": false}", []string{"noTargetBlankWithoutNoreferrer"}, "<a target='_blank' href='http://x.com' rel=\"noreferrer\"></a>"},
		{"t20", "<a target='_blank' href='http://x.com'></a>", "{\"links\": true}", []string{"noTargetBlankWithoutNoreferrer"}, "<a target='_blank' href='http://x.com' rel=\"noreferrer\"></a>"},

		// A form under `{forms: true, allowReferrer: true}` with `rel="noopener"`. It REPORTS, because
		// the form arm calls `hasSecureRel(node)` with one argument so `allowReferrer` arrives
		// undefined inside it, while the message id is chosen outside the call and still switches. So
		// the finding says noopener would do and noopener did not do. Upstream inconsistency, reproduced.
		{"t21", "<form target='_blank' action='http://x.com' rel='noopener'></form>", "{\"forms\": true, \"allowReferrer\": true}", []string{"noTargetBlankWithoutNoopener"}, ""},

		// The form arm reports with NO fix, where the link arm ships one. The empty fix column is the
		// assertion.
		{"t22", "<form target='_blank' action='http://x.com'></form>", "{\"forms\": true}", []string{"noTargetBlankWithoutNoreferrer"}, ""},

		// Under `warnOnSpreadAttributes` a `rel` written before a spread stops counting as secure, and
		// the fix still repairs it. Its flagless twin is clean below.
		{"t23", "<a rel='noreferrer' {...p} target='_blank' href='http://x.com'></a>", "{\"warnOnSpreadAttributes\": true}", []string{"noTargetBlankWithoutNoreferrer"}, "<a rel=\"noreferrer\" {...p} target='_blank' href='http://x.com'></a>"},

		// A spread and no target at all: `targetIndex` is -1, which is less than the spread index, so
		// the element is dangerous and the fix is declined.
		{"t25", "<a {...p} href='http://x.com'></a>", "{\"warnOnSpreadAttributes\": true}", []string{"noTargetBlankWithoutNoreferrer"}, ""},

		// A target before a spread: reportable under the flag, and the fixer declines because the spread
		// may be supplying the real target.
		{"t26", "<a target='_blank' {...p} href='http://x.com'></a>", "{\"warnOnSpreadAttributes\": true}", []string{"noTargetBlankWithoutNoreferrer"}, ""},

		// A self-closing element. Our parser gives it a distinct kind with no JsxOpeningElement inside,
		// so a rule listening on only one kind would be silent here.
		{"t29", "<a target='_blank' href='http://x.com' />", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target='_blank' href='http://x.com' rel=\"noreferrer\" />"},
		{"t33", "<a target='_blank' href='tel:123'></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target='_blank' href='tel:123' rel=\"noreferrer\"></a>"},

		// An empty string in a conditional branch fails upstream's truthiness guard, so only the
		// `_blank` branch decides and this reports.
		{"t36", "<a target={c ? '' : '_blank'} href='http://x.com'></a>", "", []string{"noTargetBlankWithoutNoreferrer"}, "<a target={c ? '' : '_blank'} href='http://x.com' rel=\"noreferrer\"></a>"},

		// Under `allowReferrer` the INSERTED value is `noopener`, not `noreferrer`. Only the two insert
		// arms read `relValue`; all three replace arms hardcode `noreferrer`.
		{"t38", "<a target='_blank' href='http://x.com'></a>", "{\"allowReferrer\": true}", []string{"noTargetBlankWithoutNoopener"}, "<a target='_blank' href='http://x.com' rel=\"noopener\"></a>"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			options := decodeTargetBlankOptionsForTest(t, testCase.rawOptions)
			result := rule_testing.RunWithOptions(t, JsxNoTargetBlank, jsxNoTargetBlankFile,
				testCase.sourceText, options)
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if testCase.wantFixed != "" {
				rule_testing.ExpectFixedSource(t, result, testCase.wantFixed)
				return
			}
			// An empty fix column is an assertion, not a skip.
			//
			// It says upstream reports this input and deliberately offers no repair, which is a
			// decision this port has to reproduce: a fixer that rewrote a shape upstream declines
			// to touch would be applied unattended. `ExpectFixedSource` refuses a no-fix result
			// rather than agreeing with any expectation, so the decline has to be checked here.
			//
			// Written after a mutation removing BOTH of the fixer's decline paths survived the
			// whole suite: every case covering them asserted the message id and never what was
			// offered, which is the failure mode the empty column looked like it was closing.
			for index, diagnostic := range result.Diagnostics {
				if len(diagnostic.Fixes) != 0 {
					t.Errorf("finding %d offered %d fixes; upstream declines to fix this shape",
						index, len(diagnostic.Fixes))
				}
			}
		})
	}
}

func TestJsxNoTargetBlankStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
	}{
		{"upstream valid-0", "<a href=\"foobar\"></a>", ""},
		{"upstream valid-1", "<a randomTag></a>", ""},
		{"upstream valid-2", "<a target />", ""},
		{"upstream valid-3", "<a href=\"foobar\" target=\"_blank\" rel=\"noopener noreferrer\"></a>", ""},
		{"upstream valid-4", "<a href=\"foobar\" target=\"_blank\" rel=\"noreferrer\"></a>", ""},
		{"upstream valid-5", "<a href=\"foobar\" target=\"_blank\" rel={\"noopener noreferrer\"}></a>", ""},
		{"upstream valid-6", "<a href=\"foobar\" target=\"_blank\" rel={\"noreferrer\"}></a>", ""},
		{"upstream valid-7", "<a href={\"foobar\"} target={\"_blank\"} rel={\"noopener noreferrer\"}></a>", ""},
		{"upstream valid-8", "<a href={\"foobar\"} target={\"_blank\"} rel={\"noreferrer\"}></a>", ""},
		{"upstream valid-9", "<a href={'foobar'} target={'_blank'} rel={'noopener noreferrer'}></a>", ""},
		{"upstream valid-10", "<a href={'foobar'} target={'_blank'} rel={'noreferrer'}></a>", ""},
		{"upstream valid-11", "<a href={`foobar`} target={`_blank`} rel={`noopener noreferrer`}></a>", ""},
		{"upstream valid-12", "<a href={`foobar`} target={`_blank`} rel={`noreferrer`}></a>", ""},
		{"upstream valid-13", "<a target=\"_blank\" {...spreadProps} rel=\"noopener noreferrer\"></a>", ""},
		{"upstream valid-14", "<a target=\"_blank\" {...spreadProps} rel=\"noreferrer\"></a>", ""},
		{"upstream valid-15", "<a {...spreadProps} target=\"_blank\" rel=\"noopener noreferrer\" href=\"https://example.com\">s</a>", ""},
		{"upstream valid-16", "<a {...spreadProps} target=\"_blank\" rel=\"noreferrer\" href=\"https://example.com\">s</a>", ""},
		{"upstream valid-17", "<a target=\"_blank\" rel=\"noopener noreferrer\" {...spreadProps}></a>", ""},
		{"upstream valid-18", "<a target=\"_blank\" rel=\"noreferrer\" {...spreadProps}></a>", ""},
		{"upstream valid-19", "<p target=\"_blank\"></p>", ""},
		{"upstream valid-20", "<a href=\"foobar\" target=\"_BLANK\" rel=\"NOOPENER noreferrer\"></a>", ""},
		{"upstream valid-21", "<a href=\"foobar\" target=\"_BLANK\" rel=\"NOREFERRER\"></a>", ""},
		{"upstream valid-22", "<a target=\"_blank\" rel={relValue}></a>", ""},
		{"upstream valid-23", "<a target={targetValue} rel=\"noopener noreferrer\"></a>", ""},
		{"upstream valid-24", "<a target={targetValue} rel=\"noreferrer\"></a>", ""},
		{"upstream valid-25", "<a target={targetValue} rel={\"noopener noreferrer\"}></a>", ""},
		{"upstream valid-26", "<a target={targetValue} rel={\"noreferrer\"}></a>", ""},
		{"upstream valid-27", "<a target={targetValue} href=\"relative/path\"></a>", ""},
		{"upstream valid-28", "<a target={targetValue} href=\"/absolute/path\"></a>", ""},
		{"upstream valid-29", "<a target={'targetValue'} href=\"/absolute/path\"></a>", ""},
		{"upstream valid-30", "<a target={\"targetValue\"} href=\"/absolute/path\"></a>", ""},
		{"upstream valid-31", "<a target={null} href=\"//example.com\"></a>", ""},
		{"upstream valid-32", "<a {...someObject} href=\"/absolute/path\"></a>", "{\"enforceDynamicLinks\": \"always\", \"warnOnSpreadAttributes\": true}"},
		{"upstream valid-33", "<a {...someObject} rel=\"noreferrer\"></a>", "{\"enforceDynamicLinks\": \"always\", \"warnOnSpreadAttributes\": true}"},
		{"upstream valid-34", "<a {...someObject} rel=\"noreferrer\" target=\"_blank\"></a>", "{\"enforceDynamicLinks\": \"always\", \"warnOnSpreadAttributes\": true}"},
		{"upstream valid-35", "<a {...someObject} href=\"foobar\" target=\"_blank\"></a>", "{\"enforceDynamicLinks\": \"always\", \"warnOnSpreadAttributes\": true}"},
		{"upstream valid-36", "<a target=\"_blank\" href={ dynamicLink }></a>", "{\"enforceDynamicLinks\": \"never\"}"},
		{"upstream valid-37", "<a target={\"_blank\"} href={ dynamicLink }></a>", "{\"enforceDynamicLinks\": \"never\"}"},
		{"upstream valid-38", "<a target={'_blank'} href={ dynamicLink }></a>", "{\"enforceDynamicLinks\": \"never\"}"},
		{"upstream valid-39", "<Link target=\"_blank\" href={ dynamicLink }></Link>", "{\"enforceDynamicLinks\": \"never\"}"}, // upstream sets linkComponents; cohere reads no React settings, so this is the unconfigured answer
		{"upstream valid-40", "<Link target=\"_blank\" to={ dynamicLink }></Link>", "{\"enforceDynamicLinks\": \"never\"}"},   // upstream sets linkComponents; cohere reads no React settings, so this is the unconfigured answer
		{"upstream valid-41", "<Link target=\"_blank\" to={ dynamicLink }></Link>", "{\"enforceDynamicLinks\": \"never\"}"},   // upstream sets linkComponents; cohere reads no React settings, so this is the unconfigured answer
		{"upstream valid-42", "<a href=\"foobar\" target=\"_blank\" rel=\"noopener\"></a>", "{\"allowReferrer\": true}"},
		{"upstream valid-43", "<a href=\"foobar\" target=\"_blank\" rel=\"noreferrer\"></a>", "{\"allowReferrer\": true}"},
		{"upstream valid-44", "<a target={3} />", ""},
		{"upstream valid-45", "<a href=\"some-link\" {...otherProps} target=\"some-non-blank-target\"></a>", ""},
		{"upstream valid-46", "<a href=\"some-link\" target=\"some-non-blank-target\" {...otherProps}></a>", ""},
		{"upstream valid-47", "<a target=\"_blank\" href=\"/absolute/path\"></a>", "{\"forms\": false}"},
		{"upstream valid-48", "<a target=\"_blank\" href=\"/absolute/path\"></a>", "{\"forms\": false, \"links\": true}"},
		{"upstream valid-49", "<form action=\"https://example.com\" target=\"_blank\"></form>", ""},
		{"upstream valid-50", "<form action=\"https://example.com\" target=\"_blank\" rel=\"noopener noreferrer\"></form>", "{\"forms\": true}"},
		{"upstream valid-51", "<form action=\"https://example.com\" target=\"_blank\" rel=\"noopener noreferrer\"></form>", "{\"forms\": true, \"links\": false}"},
		{"upstream valid-52", "<a href target=\"_blank\"/>", ""},
		{"upstream valid-53", "<a href={href} target={isExternal ? \"_blank\" : undefined} rel=\"noopener noreferrer\" />", ""},
		{"upstream valid-54", "<a href={href} target={isExternal ? undefined : \"_blank\"} rel={isExternal ? \"noreferrer\" : \"noopener noreferrer\"} />", ""},
		{"upstream valid-55", "<a href={href} target={isExternal ? undefined : \"_blank\"} rel={isExternal ? \"noreferrer noopener\" : \"noreferrer\"} />", ""},
		{"upstream valid-56", "<a href={href} target=\"_blank\" rel={isExternal ? \"noreferrer\" : \"noopener\"} />", "{\"allowReferrer\": true}"},
		{"upstream valid-57", "<a href={href} target={isExternal ? \"_blank\" : undefined} rel={isExternal ? \"noreferrer\" : undefined} />", ""},
		{"upstream valid-58", "<a href={href} target={isSelf ? \"_self\" : \"_blank\"} rel={isSelf ? undefined : \"noreferrer\"} />", ""},
		{"upstream valid-59", "<a href={href} target={isSelf ? \"_self\" : \"\"} rel={isSelf ? undefined : \"\"} />", ""},
		{"upstream valid-60", "<a href={href} target={isExternal ? \"_blank\" : undefined} rel={isExternal ? \"noopener noreferrer\" : undefined} />", ""},
		{"upstream valid-61", "<form action={action} />", "{\"forms\": true}"},
		{"upstream valid-62", "<form action={action} {...spread} />", "{\"forms\": true}"},
		{"upstream invalid-29", "<Link target=\"_blank\" href={ dynamicLink }></Link>", "{\"enforceDynamicLinks\": \"always\"}"}, // upstream sets linkComponents; cohere reads no React settings, so this is the unconfigured answer
		{"upstream invalid-30", "<Link target=\"_blank\" to={ dynamicLink }></Link>", "{\"enforceDynamicLinks\": \"always\"}"},   // upstream sets linkComponents; cohere reads no React settings, so this is the unconfigured answer

		// --- cases upstream does not cover, each measured against the installed build 7.37.5 ---

		// The LAST href decides, and it is relative.
		{"t5", "<a target='_blank' href='http://x.com' href='/rel'></a>", ""},

		// The last rel is secure, so no finding at all.
		{"t7", "<a href='http://x.com' target='_blank' rel='x' rel='noreferrer'></a>", ""},

		// A template literal with no substitution: `quasis[0]` is the whole text.
		{"t9", "<a href='http://x.com' target='_blank' rel={`noreferrer`}></a>", ""},

		// The conditional pairing, matching branch. Same test as the target, and the branch paired with
		// `_blank` is secure.
		{"t10", "<a href='http://x.com' target={c ? '_blank' : '_self'} rel={c ? 'noreferrer' : 'x'}></a>", ""},

		// The same pairing with the branches swapped on both attributes.
		{"t11", "<a href='http://x.com' target={c ? '_self' : '_blank'} rel={c ? 'x' : 'noreferrer'}></a>", ""},

		// Without `warnOnSpreadAttributes` a rel before a spread is trusted. The twin of the reporting
		// case above; the pair is what the flag moves.
		{"t24", "<a rel='noreferrer' {...p} target='_blank' href='http://x.com'></a>", ""},

		// A member-expression tag is a component reference, not the intrinsic `a` element.
		{"t27", "<a.b target='_blank' href='http://x.com'></a.b>", ""},

		// A capitalized tag is a component.
		{"t28", "<A target='_blank' href='http://x.com'></A>", ""},

		// `enforceDynamicLinks: "never"` drops the dynamic-href arm.
		{"t30", "<a target='_blank' href={x}></a>", "{\"enforceDynamicLinks\": \"never\"}"},

		// A leading colon with no word character before it does not match `\w+:`.
		{"t32", "<a target='_blank' href=':nocolonword'></a>", ""},

		// The rel comparison is lowercased too.
		{"t34", "<a target='_blank' href='http://x.com' rel='NOREFERRER'></a>", ""},

		// An empty target string fails the truthiness guard and is not `_blank` anyway.
		{"t35", "<a target={''} href='http://x.com'></a>", ""},

		// Under `allowReferrer`, `rel="noopener"` alone satisfies.
		{"t37", "<a target='_blank' href='http://x.com' rel='noopener'></a>", "{\"allowReferrer\": true}"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			options := decodeTargetBlankOptionsForTest(t, testCase.rawOptions)
			result := rule_testing.RunWithOptions(t, JsxNoTargetBlank, jsxNoTargetBlankFile,
				testCase.sourceText, options)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// decodeTargetBlankOptionsForTest routes a fixture's raw JSON through the rule's own decoder.
//
// This is the whole point of storing raw JSON in the tables: the enum inversion and the `links`
// default live in that function and nowhere else, so a fixture that built the struct directly would
// be testing the rule against options no config can produce.
func decodeTargetBlankOptionsForTest(t *testing.T, raw string) any {
	t.Helper()
	options, err := DecodeJsxNoTargetBlankOptions([]byte(raw))
	if err != nil {
		t.Fatalf("decoding options %q: %v", raw, err)
	}
	return options
}

// TestJsxNoTargetBlankSpans pins that the finding lands on the OPENING element.
//
// Upstream reports `node`, which is the `JSXOpeningElement` its listener receives, so the underline
// stops at the closing angle bracket and never covers the children or the closing tag. Our tree has
// the same node for the paired form and a separate kind for the self-closing one, and reporting the
// whole `JsxElement` instead would be a plausible wrong answer that every id assertion above stays
// green over.
//
// This matters twice for this rule, because 42 of its fixtures carry a repair: a finding anchored
// on the wrong node while carrying a fix means the edit lands somewhere the reader was never shown.
func TestJsxNoTargetBlankSpans(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		rawOptions string
		want       string
	}{
		{"paired element", "<a target='_blank' href='http://x.com'>text</a>", "",
			"<a target='_blank' href='http://x.com'>"},
		{"self-closing element", "<a target='_blank' href='http://x.com' />", "",
			"<a target='_blank' href='http://x.com' />"},
		{"form element", "<form target='_blank' action='http://x.com'>x</form>", "{\"forms\":true}",
			"<form target='_blank' action='http://x.com'>"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			options := decodeTargetBlankOptionsForTest(t, testCase.rawOptions)
			result := rule_testing.RunWithOptions(t, JsxNoTargetBlank, jsxNoTargetBlankFile,
				testCase.sourceText, options)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want 1 finding, got %d", len(result.Diagnostics))
			}
			// `rule_testing.RunWithOptions` does not trim, so the source on disk is this literal and
			// slicing it directly is sound. The typed harness would need the trim applied first.
			reported := testCase.sourceText[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.want {
				t.Errorf("span: got %q, want %q", reported, testCase.want)
			}
		})
	}
}

// TestDecodeJsxNoTargetBlankOptions pins the two defaults that are not Go's zero value.
//
// Both lines here have no upstream counterpart to copy from: `enforceDynamicLinks` is a string
// enum inverted into a bool, and `links` defaults to true. A fixture that built the options struct
// by hand would exercise neither, which is why every table above routes through this function.
func TestDecodeJsxNoTargetBlankOptions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
		want JsxNoTargetBlankOptions
	}{
		// The nil path, which is what a rule configured as a bare `"error"` is handed.
		{"empty input keeps upstream defaults", "", JsxNoTargetBlankOptions{Links: true}},
		{"empty object keeps upstream defaults", "{}", JsxNoTargetBlankOptions{Links: true}},

		// `enforceDynamicLinks` absent means `always`, so the inverted flag stays false.
		{"always is the absent value", "{\"enforceDynamicLinks\":\"always\"}",
			JsxNoTargetBlankOptions{Links: true}},
		{"never inverts the flag", "{\"enforceDynamicLinks\":\"never\"}",
			JsxNoTargetBlankOptions{Links: true, EnforceDynamicLinksNever: true}},

		// `links` defaults to TRUE, so the zero struct is wrong for it and an explicit false has to
		// survive the decode even though nothing reads it.
		{"links false survives", "{\"links\":false}", JsxNoTargetBlankOptions{Links: false}},

		{"every flag at once",
			"{\"allowReferrer\":true,\"warnOnSpreadAttributes\":true,\"forms\":true,\"enforceDynamicLinks\":\"never\",\"links\":false}",
			JsxNoTargetBlankOptions{AllowReferrer: true, WarnOnSpreadAttributes: true, Forms: true,
				EnforceDynamicLinksNever: true, Links: false}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeJsxNoTargetBlankOptions([]byte(testCase.raw))
			if err != nil {
				t.Fatalf("decoding %q: %v", testCase.raw, err)
			}
			if decoded != any(testCase.want) {
				t.Errorf("got %+v, want %+v", decoded, testCase.want)
			}
		})
	}

	if _, err := DecodeJsxNoTargetBlankOptions([]byte("not json")); err == nil {
		t.Error("malformed input should error rather than silently producing defaults")
	}
}

// TestJsxNoTargetBlankHandlesNilOptions pins the path that bypasses the decoder entirely.
//
// A rule can reach `Run` with nil, and the comma-ok assertion then yields the zero struct whose
// `Links` is false rather than upstream's true. Nothing reads `Links` today, so this cannot change
// a verdict; the test exists because the rule restores the default explicitly and a later reader
// deleting that restoration should fail here rather than ship a silent change.
func TestJsxNoTargetBlankHandlesNilOptions(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunWithOptions(t, JsxNoTargetBlank, jsxNoTargetBlankFile,
		"<a target=\"_blank\" href=\"http://x.com\"></a>", nil)
	rule_testing.ExpectFindings(t, result, "noTargetBlankWithoutNoreferrer")

	// And the dynamic-href arm, which is the one the absent enum turns ON. A nil-options rule that
	// read the enum as "never" would be silent here.
	dynamic := rule_testing.RunWithOptions(t, JsxNoTargetBlank, jsxNoTargetBlankFile,
		"<a target=\"_blank\" href={dynamicLink}></a>", nil)
	rule_testing.ExpectFindings(t, dynamic, "noTargetBlankWithoutNoreferrer")
}

// TestJsxNoTargetBlankHasNoFileGate pins that the extension decides nothing.
//
// Upstream registers no filename predicate. Only the three JSX-capable extensions are exercised,
// because a `.ts` file cannot parse `<a ... />` as JSX at all and its silence would say nothing
// about the rule.
func TestJsxNoTargetBlankHasNoFileGate(t *testing.T) {
	t.Parallel()

	const source = "<a target=\"_blank\" href=\"http://x.com\"></a>"
	for _, fileName := range []string{
		"/repository/source/Probe.tsx",
		"/repository/source/Probe.jsx",
		"/repository/source/Probe.js",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, JsxNoTargetBlank, fileName, source, nil)
			rule_testing.ExpectFindings(t, result, "noTargetBlankWithoutNoreferrer")
		})
	}
}

package typescript

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const banTsCommentFile = "/repository/source/Thing.ts"

// decodeBanTsCommentOptions runs a configuration through the real decoder rather than building
// the options struct directly.
//
// The decoder holds five defaults, a polymorphic value for four of the keys, and a regex
// compile, so a fixture constructing BanTsCommentOptions by hand would leave the whole config
// half of the port unexercised. An empty configuration is the bare `"error"` case, which the
// config layer turns into nil options, and it reaches the rule the same way here.
func decodeBanTsCommentOptions(t *testing.T, configuration string) any {
	t.Helper()

	if configuration == "" {
		// A rule configured as a bare `"error"` is handed nil, not an empty struct. Passing nil
		// here is what puts the rule's own fallback under test rather than the decoder's.
		return nil
	}
	options, err := DecodeBanTsCommentOptions([]byte(configuration))
	if err != nil {
		t.Fatalf("could not decode %s: %v", configuration, err)
	}
	return options
}

// TestBanTsCommentUpstreamCorpus replays typescript-eslint 8.71.0's own rows and the edge rows
// (ban_ts_comment_corpus_data_test.go says how they were recorded). Each row asserts the message id,
// the exact text the finding covers, and what its suggestion writes, and that no finding carries an
// autofix, because upstream offers none.
func TestBanTsCommentUpstreamCorpus(t *testing.T) {
	t.Parallel()

	valid, invalid := 0, 0
	for _, row := range banTsCommentCorpus {
		if row.edge != "" {
			continue
		}
		if len(row.findings) == 0 {
			valid++
		} else {
			invalid++
		}
	}
	if valid != banTsCommentCorpusUpstreamValid || invalid != banTsCommentCorpusUpstreamInvalid {
		t.Fatalf("the corpus holds %d valid and %d invalid upstream rows, and upstream has %d and %d",
			valid, invalid, banTsCommentCorpusUpstreamValid, banTsCommentCorpusUpstreamInvalid)
	}

	for _, row := range banTsCommentCorpus {
		name := fmt.Sprintf("upstream-%d", row.index)
		if row.edge != "" {
			name = row.edge
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, BanTsComment, banTsCommentFile, row.source,
				decodeBanTsCommentOptions(t, row.options))
			if len(row.findings) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}

			diagnostics := result.Diagnostics
			sort.SliceStable(diagnostics, func(first, second int) bool {
				return diagnostics[first].Range.Pos() < diagnostics[second].Range.Pos()
			})
			reported := []banTsCommentCorpusFinding{}
			for _, diagnostic := range diagnostics {
				if len(diagnostic.Fixes) > 0 {
					t.Errorf("the rule proposes an autofix typescript-eslint 8.71.0 does not: %+v", diagnostic.Fixes)
				}
				suggestion := ""
				if len(diagnostic.Suggestions) > 0 {
					for _, fix := range diagnostic.Suggestions[0].Fixes {
						suggestion = row.source[:fix.Range.Pos()] + fix.Text + row.source[fix.Range.End():]
					}
				}
				reported = append(reported, banTsCommentCorpusFinding{
					id:         diagnostic.Message.Id,
					text:       row.source[diagnostic.Range.Pos():diagnostic.Range.End()],
					suggestion: suggestion,
				})
			}
			if fmt.Sprint(reported) != fmt.Sprint(row.findings) {
				t.Fatalf("the rule reports %q, and typescript-eslint 8.71.0 reports %q", reported, row.findings)
			}
			// Through the harness assertion as well, so the docs capture records the case.
			ids := []string{}
			for _, finding := range row.findings {
				ids = append(ids, finding.id)
			}
			rule_testing.ExpectFindings(t, result, ids...)
		})
	}
}

// banTsCommentCaseName names a subtest by its index in a table.
func banTsCommentCaseName(index int) string {
	return "case-" + strconv.Itoa(index)
}

// TestBanTsCommentDefaultsBindWithNoConfiguration is the fixture that bypasses the decoder.
//
// A rule configured as a bare `"error"` is handed nil options, and a bare type assertion on nil
// yields four zero-valued settings that match no arm, which registers the rule on every file and
// reports nothing while every decoder-routed fixture stays green. This passes nil directly, so it
// is the rule's own fallback under test rather than DecodeBanTsCommentOptions.
//
// Each row is one of typescript-eslint's defaultOptions. The `@ts-check` row is the one a reader is
// most likely to guess backwards: `@ts-check` turns checking ON, so it defaults to allowed.
func TestBanTsCommentDefaultsBindWithNoConfiguration(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText string
		wantIds    []string
		reason     string
	}{
		{"// @ts-ignore\n", []string{"tsIgnoreInsteadOfExpectError"}, "ts-ignore defaults to banned"},
		{"// @ts-nocheck\n", []string{"tsDirectiveComment"}, "ts-nocheck defaults to banned"},
		{"// @ts-check\n", nil, "ts-check defaults to ALLOWED"},
		{"// @ts-expect-error\n", []string{"tsDirectiveCommentRequiresDescription"},
			"ts-expect-error defaults to allow-with-description"},
		{"// @ts-expect-error ab\n", []string{"tsDirectiveCommentRequiresDescription"},
			"two characters is under the default minimum of three"},
		{"// @ts-expect-error abc\n", nil, "three characters reaches the default minimum"},
	}
	for index, testCase := range cases {
		t.Run(banTsCommentCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, BanTsComment, banTsCommentFile,
				testCase.sourceText, nil)
			// A silent row asserts clean, so the docs capture records it as clean rather than as a
			// firing case with no findings.
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
		})
	}
}

// TestDecodeBanTsCommentOptionsReadsEveryShape exercises the decoder directly.
//
// Four of the five keys accept a boolean, the string `allow-with-description`, or an object holding
// a pattern, which no struct tag can express. An absent key keeps the default; an unrecognized value
// and an unparseable pattern are refused, below. An object whose pattern is absent, null or empty
// holds no pattern, because upstream tests `option.descriptionFormat` for truth.
func TestDecodeBanTsCommentOptionsReadsEveryShape(t *testing.T) {
	t.Parallel()

	defaults := DefaultBanTsCommentOptions()

	cases := []struct {
		configuration string
		check         func(t *testing.T, options BanTsCommentOptions)
	}{
		{"{\"ts-ignore\":false}", func(t *testing.T, options BanTsCommentOptions) {
			if options.TsIgnore.Kind != BanTsCommentBoolean || options.TsIgnore.Banned {
				t.Fatalf("expected an allowing boolean, got %+v", options.TsIgnore)
			}
		}},
		{"{\"ts-check\":true}", func(t *testing.T, options BanTsCommentOptions) {
			if options.TsCheck.Kind != BanTsCommentBoolean || !options.TsCheck.Banned {
				t.Fatalf("expected a banning boolean, got %+v", options.TsCheck)
			}
		}},
		{"{\"ts-expect-error\":\"allow-with-description\"}",
			func(t *testing.T, options BanTsCommentOptions) {
				if options.TsExpectError.Kind != BanTsCommentRequireDescription {
					t.Fatalf("expected require-description, got %+v", options.TsExpectError)
				}
			}},
		{"{\"ts-nocheck\":{\"descriptionFormat\":\"^: TS\\\\d+$\"}}",
			func(t *testing.T, options BanTsCommentOptions) {
				if options.TsNoCheck.Kind != BanTsCommentDescriptionFormat {
					t.Fatalf("expected description-format, got %+v", options.TsNoCheck)
				}
				if options.TsNoCheck.DescriptionFormat == nil ||
					options.TsNoCheck.DescriptionFormat.Source() != "^: TS\\d+$" {
					t.Fatalf("pattern did not survive the decode: %+v", options.TsNoCheck)
				}
			}},
		{"{\"minimumDescriptionLength\":21}", func(t *testing.T, options BanTsCommentOptions) {
			if options.MinimumDescriptionLength != 21 {
				t.Fatalf("expected 21, got %d", options.MinimumDescriptionLength)
			}
		}},
		{"{}", func(t *testing.T, options BanTsCommentOptions) {
			if options != defaults {
				t.Fatalf("an empty object should leave every default, got %+v", options)
			}
		}},
		{"{\"ts-ignore\":{\"descriptionFormat\":\"\"}}", func(t *testing.T, options BanTsCommentOptions) {
			if options.TsIgnore.Kind != BanTsCommentDescriptionFormat || options.TsIgnore.DescriptionFormat != nil {
				t.Fatalf("an empty pattern should leave no pattern, got %+v", options.TsIgnore)
			}
		}},
		{"{\"ts-ignore\":{\"descriptionFormat\":null}}", func(t *testing.T, options BanTsCommentOptions) {
			if options.TsIgnore.Kind != BanTsCommentDescriptionFormat || options.TsIgnore.DescriptionFormat != nil {
				t.Fatalf("a null pattern should leave no pattern, got %+v", options.TsIgnore)
			}
		}},
		// A pattern only JavaScript compiles: Go's regexp has no lookahead, and ESLint accepts it.
		{"{\"ts-ignore\":{\"descriptionFormat\":\"^(?=: TS)\"}}",
			func(t *testing.T, options BanTsCommentOptions) {
				if options.TsIgnore.DescriptionFormat == nil || options.TsIgnore.DescriptionFormat.Source() != "^(?=: TS)" {
					t.Fatalf("a lookahead pattern did not survive the decode: %+v", options.TsIgnore)
				}
			}},
	}
	for index, testCase := range cases {
		t.Run(banTsCommentCaseName(index), func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeBanTsCommentOptions([]byte(testCase.configuration))
			if err != nil {
				t.Fatalf("could not decode %s: %v", testCase.configuration, err)
			}
			options, isOptions := decoded.(BanTsCommentOptions)
			if !isOptions {
				t.Fatalf("decoder returned %T rather than BanTsCommentOptions", decoded)
			}
			testCase.check(t, options)
		})
	}
}

// TestDecodeBanTsCommentOptionsRefusesWhatUpstreamRefuses pins the refusals that replaced the decoder's
// fallback to the default (ruled on #pd2chkx): each value here fails typescript-eslint's schema or its
// `new RegExp`, so ESLint refuses the config, and so does cohere, naming the key.
func TestDecodeBanTsCommentOptionsRefusesWhatUpstreamRefuses(t *testing.T) {
	t.Parallel()

	cases := []struct {
		configuration string
		want          string
	}{
		{`{"ts-ignore":"nonsense"}`, `ts-ignore: expected true, false, "allow-with-description"`},
		{`{"ts-check":"Allow-with-description"}`, `ts-check: expected true, false, "allow-with-description"`},
		{`{"ts-nocheck":3}`, `ts-nocheck: expected true, false, "allow-with-description"`},
		{`{"ts-expect-error":[]}`, `ts-expect-error: expected true, false, "allow-with-description"`},
		{`{"ts-ignore":{"descriptionFormat":"^(unclosed"}}`, `ts-ignore: descriptionFormat "^(unclosed" is not a pattern JavaScript can compile`},
	}
	for index, testCase := range cases {
		t.Run(banTsCommentCaseName(index), func(t *testing.T) {
			t.Parallel()
			_, err := DecodeBanTsCommentOptions([]byte(testCase.configuration))
			if err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("decoding %s: error %v, want one containing %q", testCase.configuration, err, testCase.want)
			}
		})
	}
}

// TestBanTsCommentDeclinesJavaScriptFiles pins the gate kept as a stand-in for config scope.
//
// Upstream's rule has no file-type gate; ESLint configs commonly scope it to TypeScript files. cohere's
// overrides can too, but readiness resolves cohere:adamic once for `index.ts`, so the scope cannot
// move into the set yet. The rule's doc comment has the TanStack measurement. #6aa3wrx moves the
// scope and deletes this test with the gate.
func TestBanTsCommentDeclinesJavaScriptFiles(t *testing.T) {
	t.Parallel()

	for _, fileName := range []string{
		"/repository/source/Thing.js",
		"/repository/source/Thing.jsx",
		"/repository/source/Thing.mjs",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunWithOptions(t, BanTsComment, fileName, "// @ts-ignore\n", nil))
		})
	}
	for _, fileName := range []string{
		"/repository/source/Thing.ts",
		"/repository/source/Thing.tsx",
		"/repository/source/Thing.mts",
		"/repository/source/Thing.cts",
	} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, BanTsComment, fileName, "// @ts-ignore\n", nil)
			rule_testing.ExpectFindings(t, result, "tsIgnoreInsteadOfExpectError")
		})
	}
}

// TestBanTsCommentMessageTextNamesTheDirective asserts the rendered text, exactly.
//
// Three of the four messages interpolate, and ExpectFindings sees only the id, so a format string
// that dropped the directive name or doubled a character would leave every fixture above green.
//
// Two things about how this is written. The comparison is EQUALITY on the whole description rather
// than a Contains on a fragment, because a predicate weaker than the property it guards is not a
// guard: a doubled character produces a string that still contains the needle. And every expected
// string is a literal typed here rather than a call to the rule's own message constructor, because
// comparing a finding against the constant it was reported with is equality between two values that
// move together under mutation and therefore proves nothing.
func TestBanTsCommentMessageTextNamesTheDirective(t *testing.T) {
	t.Parallel()

	cases := []struct {
		configuration string
		sourceText    string
		wantId        string
		wantText      string
	}{
		{"", "// @ts-nocheck\n", "tsDirectiveComment",
			"This file silences the TypeScript compiler with `@ts-nocheck`. A directive comment " +
				"switches off checking for a line or a whole file, so the error it hides stays in " +
				"the code and stops being reported to anyone. Fix the underlying type error " +
				"instead, or narrow the suppression to `@ts-expect-error` with a description " +
				"saying why it is necessary."},
		{"{\"ts-check\":true}", "// @ts-check\n", "tsDirectiveComment",
			"This file silences the TypeScript compiler with `@ts-check`. A directive comment " +
				"switches off checking for a line or a whole file, so the error it hides stays in " +
				"the code and stops being reported to anyone. Fix the underlying type error " +
				"instead, or narrow the suppression to `@ts-expect-error` with a description " +
				"saying why it is necessary."},
		{"", "// @ts-ignore\n", "tsIgnoreInsteadOfExpectError",
			"This suppression uses `@ts-ignore`, which keeps silencing the line even after the " +
				"underlying error is gone, so a stale suppression is invisible forever. " +
				"`@ts-expect-error` reports when the line it guards has stopped failing, which is " +
				"what makes the suppression removable."},
		{"", "// @ts-expect-error\n", "tsDirectiveCommentRequiresDescription",
			"This `@ts-expect-error` carries no explanation, so nobody reading it later can tell " +
				"whether the suppression is still needed or what it was hiding. Write at least 3 " +
				"characters after the directive saying which error it suppresses and why the " +
				"error cannot be fixed."},
		{"{\"ts-nocheck\":\"allow-with-description\",\"minimumDescriptionLength\":21}",
			"// @ts-nocheck ab\n", "tsDirectiveCommentRequiresDescription",
			"This `@ts-nocheck` carries no explanation, so nobody reading it later can tell " +
				"whether the suppression is still needed or what it was hiding. Write at least 21 " +
				"characters after the directive saying which error it suppresses and why the " +
				"error cannot be fixed."},
		{"{\"ts-ignore\":{\"descriptionFormat\":\"^: TS[0-9]+$\"}}", "// @ts-ignore: nope\n",
			"tsDirectiveCommentDescriptionNotMatchPattern",
			"The explanation after this `@ts-ignore` does not match the shape this project " +
				"requires, `^: TS[0-9]+$`. The format exists so a suppression can be traced back " +
				"to the error it hides, which a free-form note cannot be."},
	}
	for index, testCase := range cases {
		t.Run(banTsCommentCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunWithOptions(t, BanTsComment, banTsCommentFile,
				testCase.sourceText, decodeBanTsCommentOptions(t, testCase.configuration))
			if len(result.Diagnostics) != 1 {
				t.Fatalf("expected one finding, got %d: %v", len(result.Diagnostics),
					result.MessageIds())
			}
			if result.Diagnostics[0].Message.Id != testCase.wantId {
				t.Fatalf("expected id %q, got %q", testCase.wantId,
					result.Diagnostics[0].Message.Id)
			}
			if result.Diagnostics[0].Message.Description != testCase.wantText {
				t.Fatalf("message text was\n  %q\nexpected\n  %q",
					result.Diagnostics[0].Message.Description, testCase.wantText)
			}
		})
	}
}

// TestBanTsCommentSurvivesAnUnterminatedBlockComment pins the guard in banTsCommentValue.
//
// Those guards are crash protection rather than behavioral filters, which is why the sweep reports
// them as survivors under any assertion about findings: no ExpectFindings fixture can see a panic.
// Ask what the line PREVENTS rather than what it decides, and this is what it prevents.
//
// The reachable input is error recovery. A well-formed block comment is at least four bytes
// (`/**/`), so a naive read says the four-byte guard is dead. Probed against the parser instead: an
// unterminated `/*` at the end of a file comes back as a block comment of length TWO, and `/*/` and
// `/**` come back as length three. Subtracting two from each end of a two-byte comment produces a
// range whose start is past its end, and slicing the source with it panics. Measured by
// neutralizing both guards, which turns every row below into a stack trace.
//
// A file ending in an unterminated comment is not exotic. It is what a half-written file on disk
// looks like while somebody is typing, and the linter runs on those.
func TestBanTsCommentSurvivesAnUnterminatedBlockComment(t *testing.T) {
	t.Parallel()

	for _, sourceText := range []string{"/*", "/*/", "/**", "/*@ts-ignore", "//", "/*@ts-ignore\n"} {
		t.Run(sourceText, func(t *testing.T) {
			t.Parallel()
			// The assertion is that this returns at all. A finding would also be acceptable for
			// some of these; a panic is not, and a panic is what the guards prevent.
			_ = rule_testing.RunWithOptions(t, BanTsComment, banTsCommentFile, sourceText, nil)
		})
	}
}

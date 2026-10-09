package nexus

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const shoutingFile = "/repository/source/Thing.ts"

func TestConsistencyNoShoutingFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"shouted word in a line comment", "// NEVER cache this\nexport const Value = 1;\n"},
		{"shouted word in a block comment", "/* this is REALLY important */\nexport const Value = 1;\n"},
		{"banner line", "// DO NOT EDIT\nexport const Value = 1;\n"},
		{"shouted two-letter word", "// this IS the one\nexport const Value = 1;\n"},
		{"shouted word in jsdoc", "/** ALWAYS returns a copy */\nexport const Value = 1;\n"},
		// An unterminated backtick must not mask the rest of the comment. If it did, opening a
		// backtick would be a way to turn the rule off for everything after it.
		{"unclosed backtick does not mask the rest", "// `code and then NEVER do this\nexport const Value = 1;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoShouting, shoutingFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "shoutingInComment")
		})
	}
}

func TestConsistencyNoShoutingStaysSilent(t *testing.T) {
	t.Parallel()

	// These are the discriminating cases. A rule that only detects capitals would fire on every
	// one of them, and a rule that fires on ordinary technical prose is a rule people disable.
	cases := []struct {
		name       string
		sourceText string
	}{
		{"ordinary prose", "// parse the body before validating it\nexport const Value = 1;\n"},
		{"allowlisted acronyms", "// the JSON body arrives over HTTP from the API\nexport const Value = 1;\n"},
		{"consonant cluster acronym", "// send it over RPC to the DB\nexport const Value = 1;\n"},
		{"backticked identifier", "// the `NOT_FOUND` branch returns early\nexport const Value = 1;\n"},
		{"backticked shouted word", "// the `NEVER` constant is a sentinel\nexport const Value = 1;\n"},
		{"identifier with underscore", "// read DATABASE_URL from the environment\nexport const Value = 1;\n"},
		{"identifier with a digit", "// the H3 heading and the C1 column\nexport const Value = 1;\n"},
		{"currency code", "// amounts in KRW have no minor unit\nexport const Value = 1;\n"},
		{"double quoted literal", "// \"ORDER STATUS\" is the column label\nexport const Value = 1;\n"},
		{"single quoted token", "// pass 'US' as the region\nexport const Value = 1;\n"},
		{"fenced block", "/*\n * ```\n * NEVER DO THIS\n * ```\n */\nexport const Value = 1;\n"},
		{"command line", "/*\n * git commit --amend NEVER\n */\nexport const Value = 1;\n"},
		{"two-letter abbreviation", "// the IP address and the OS version\nexport const Value = 1;\n"},
		{"notation shorthand", "// TODO: handle the empty case\nexport const Value = 1;\n"},
		{"no comments at all", "export const Value = 1;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoShouting, shoutingFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The 2026-10-01 rule review hand-checked every finding on ahra. Five were acronyms the vowel test
// misread as raised voice, and those lines are here verbatim as must-stay-silent cases. The
// true positives from the same review sit beside them, verbatim, so the safe list growing can never
// quietly swallow a real shout or a code constant that still owes its backticks.
func TestConsistencyNoShoutingKnowsTheAcronymsTheReviewFound(t *testing.T) {
	t.Parallel()

	silent := []struct {
		name       string
		sourceText string
	}{
		{"TUI in a jsdoc block (CodexConversationSource.ts)",
			"/**\n * A freshly started Codex TUI writes session metadata before the human has typed\n */\nexport const Value = 1;\n"},
		{"TUI in a line comment (CreateMindRuntimeTables.ts)",
			"// adapter projects its native TUI into the same two columns.\nexport const Value = 1;\n"},
		{"ARP (UnifiProtectChimeCommandLineInterface.ts)",
			"/*\n * kept dialing .176 on port 7442 for six months (seen on the wire 2026-10-01 as ARP who-has every\n */\nexport const Value = 1;\n"},
		{"CDATA (PensieveDailies.ts)",
			"// XML character references are literal inside CDATA. Unwrap the one admitted\nexport const Value = 1;\n"},
		{"CAF (SpeechRenderer.ts)",
			"// Writes `text` spoken by `voice` (an identifier or a voice name) to a CAF file at `outputPath`\nexport const Value = 1;\n"},
		// Not from the ahra review: the health-records interoperability standard, flagged in
		// www-phi-health's ClinicalRecords.tsx.
		{"FHIR (ClinicalRecords.tsx)",
			"// FHIR interpretation code: \"N\" = Normal, \"A\"/\"AA\"/\"H\"/\"HH\"/\"L\"/\"LL\" = abnormal.\nexport const Value = 1;\n"},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoShouting, shoutingFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}

	fires := []struct {
		name       string
		sourceText string
		token      string
	}{
		// An errno name is a literal the reader matches against what the system emits, so it is code
		// and owes backticks. A dictionary test would read it as an acronym and let it through.
		{"errno name (AhraOsPositionFifo.ts)",
			"/*\n * wake and a delivery to a reader-less pipe already fails ENXIO and shrugs. Idempotent,\n */\nexport const Value = 1;\n",
			"ENXIO"},
		{"emphasis (AhraOsPositionFifo.ts)",
			"/*\n * safe while watchers are armed BECAUSE the pipe is only the fast-path poke: the durable\n */\nexport const Value = 1;\n",
			"BECAUSE"},
		{"emphasis (NextEventLoopTurn.ts)",
			"/**\n * This YIELDS. It makes no claim about what finished, because it cannot know\n */\nexport const Value = 1;\n",
			"YIELDS"},
		{"banner and a sqlite keyword (FinancePositionCommandLineInterface.ts)",
			"/*\n * OUTPUT MUST BE VALID INPUT. The name is printed in FULL, never fit() to the column width: a\n */\nexport const Value = 1;\n",
			"OUTPUT"},
		{"emphasis (AhraOsTriggers.ts)",
			"/*\n * DO NOT rename this symbol to `positionsMissingPortrait` with the rest of the\n */\nexport const Value = 1;\n",
			"NOT"},
		{"emphasis (FinanceHoldingsCommandLineInterface.ts)",
			"// straight back in. A miss is a LOUD failure with near misses named by ID, never a silent\nexport const Value = 1;\n",
			"LOUD"},
	}
	for _, testCase := range fires {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoShouting, shoutingFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "shoutingInComment")
			if description := result.Diagnostics[0].Message.Description; !strings.Contains(description, `"`+testCase.token+`"`) {
				t.Fatalf("expected the message to name %q, got: %s", testCase.token, description)
			}
		})
	}
}

// Kirk's ruling on the review: all caps belongs in backticks, single quotes or double quotes unless
// it is on the safe list. Backticks, double quotes and a single-quoted token were already masked; a
// single-quoted phrase was not, because an apostrophe is the same character as an opening quote and
// a naive pair-match swallows the rest of a sentence. These pin the narrow version: the opening quote
// follows no letter or digit, the closing one is followed by none, and nothing inside is lowercase.
func TestConsistencyNoShoutingMasksASingleQuotedCapitalPhrase(t *testing.T) {
	t.Parallel()

	silent := []struct {
		name       string
		sourceText string
	}{
		{"quoted sql phrase", "// a 'RENAME COLUMN' statement is not idempotent\nexport const Value = 1;\n"},
		{"quoted phrase at the start of the comment", "// 'ON CONFLICT DO NOTHING' keeps the first row\nexport const Value = 1;\n"},
		{"quoted phrase before punctuation", "// the pragma reads 'SYNCHRONOUS FULL', then commits\nexport const Value = 1;\n"},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoShouting, shoutingFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}

	fires := []struct {
		name       string
		sourceText string
		token      string
	}{
		// A plural possessive is an apostrophe after a letter. It must not open a quote that the
		// next apostrophe closes, or the shout between them disappears.
		{"possessive apostrophe opens nothing", "// the users' data is NEVER cached by the dogs' walker\nexport const Value = 1;\n", "NEVER"},
		{"contraction opens nothing", "// don't EVER cache what isn't yours\nexport const Value = 1;\n", "EVER"},
		// The apostrophe after BOSS follows a letter, so it opens nothing, and the closing apostrophe
		// after ORDERS must not pair with it to hide ORDERS.
		{"opening quote glued to a letter", "// the BOSS'S ORDERS' wording\nexport const Value = 1;\n", "ORDERS"},
		{"closing quote glued to a letter", "// a 'RENAME COLUMN'd table is NEVER safe\nexport const Value = 1;\n", "RENAME"},
		// The apostrophe hazard itself: a leading apostrophe after a space and a plural possessive
		// later in the sentence look like a pair at both boundaries. Lowercase inside is what refuses it.
		{"apostrophes that bracket prose", "// the '90s were LOUD and the users' data was quiet\nexport const Value = 1;\n", "LOUD"},
		{"lowercase inside the quotes", "// the 'tis season, NEVER the 'quiet' one\nexport const Value = 1;\n", "NEVER"},
		{"shout outside a quoted phrase", "// 'NOT NULL' is ALWAYS enforced\nexport const Value = 1;\n", "ALWAYS"},
	}
	for _, testCase := range fires {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoShouting, shoutingFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "shoutingInComment")
			if description := result.Diagnostics[0].Message.Description; !strings.Contains(description, `"`+testCase.token+`"`) {
				t.Fatalf("expected the message to name %q, got: %s", testCase.token, description)
			}
		})
	}
}

// An indented code block is code, like a fence (#vy055vj): a sandbox command's TMPDIR and $HOME are
// not a raised voice. The blank line before it is what makes it a block; an indented line straight
// after prose continues the paragraph and is still read. A block comment's own indentation is set
// aside first, so only a block indented past the gutter counts. Mirrors nexus's Shouting.test.ts.
func TestConsistencyNoShoutingMasksAnIndentedCodeBlock(t *testing.T) {
	t.Parallel()

	silent := []struct {
		name       string
		sourceText string
	}{
		{"indented block in a block comment", "/**\n * Run it sandboxed:\n *\n *     TMPDIR=$HOME/tmp sandbox-exec -f profile.sb node run.mjs\n */\nexport const Value = 1;\n"},
		{"tab-indented block", "/*\nRun it:\n\n\tHOME=/x TMPDIR=/y node run.mjs\n*/\nexport const Value = 1;\n"},
		{"fenced block, the control", "/**\n * Run it:\n * ```\n * TMPDIR=$HOME/tmp node run.mjs\n * ```\n */\nexport const Value = 1;\n"},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoShouting, shoutingFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}

	fires := []struct {
		name       string
		sourceText string
		token      string
	}{
		{"indented line straight after prose", "/*\nThis is prose\n    and you must NEVER do it\n*/\nexport const Value = 1;\n", "NEVER"},
		{"prose after an indented block", "/**\n * Run:\n *\n *     HOME=/x run\n *\n * Then ALWAYS check.\n */\nexport const Value = 1;\n", "ALWAYS"},
		{"a deeply indented comment line is prose", "export function run() {\n    /**\n     * You must NEVER do this.\n     */\n    return 1;\n}\n", "NEVER"},
	}
	for _, testCase := range fires {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoShouting, shoutingFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "shoutingInComment")
			if description := result.Diagnostics[0].Message.Description; !strings.Contains(description, `"`+testCase.token+`"`) {
				t.Fatalf("expected the message to name %q, got: %s", testCase.token, description)
			}
		})
	}
}

// Kirk's ruling: a quoted all-caps phrase is a literal, and a block comment that reflows a quote
// onto the next line has not unquoted it. The first silent row is ahra's Shouting.ts verbatim, where
// the quote opens on one line and closes behind the next line's gutter. The reporting rows pin the
// edges: capitals after the closing quote, a stray inch mark with nothing on the next line to close
// it, and a quote that only closes two lines later (one wrap is the limit).
func TestConsistencyNoShoutingMasksADoubleQuoteThatWraps(t *testing.T) {
	t.Parallel()

	silent := []struct {
		name       string
		sourceText string
	}{
		{"Shouting.ts, a quote wrapped behind a gutter",
			"/*\n * Kept apart from the token pattern above so prose stays safe: the run must start and end on a\n * capital, digit or underscore and hold only capitals and token punctuation, so \"the caller's job\n * and you must NOT\" still cannot pair its apostrophe with a later one. Mirrors cohere's shouting.go.\n */\nexport const Value = 1;\n"},
		{"a wrapped quote after a one-line quote on the same line",
			"/*\n * \"ok\" then \"a phrase that\n * says NEVER\" here\n */\nexport const Value = 1;\n"},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoShouting, shoutingFile, testCase.sourceText)
			rule_testing.ExpectClean(t, result)
		})
	}

	fires := []struct {
		name       string
		sourceText string
		token      string
	}{
		{"capitals after the closing quote on the next line",
			"/*\n * the \"quiet\n * phrase\" is NEVER loud\n */\nexport const Value = 1;\n", "NEVER"},
		{"a stray inch mark opens nothing",
			"/*\n * a 5\" board\n * is NEVER warped\n */\nexport const Value = 1;\n", "NEVER"},
		{"a quote that closes two lines later",
			"/*\n * the \"first\n * line is NEVER\n * closed\" here\n */\nexport const Value = 1;\n", "NEVER"},
	}
	for _, testCase := range fires {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoShouting, shoutingFile, testCase.sourceText)
			rule_testing.ExpectFindings(t, result, "shoutingInComment")
			if description := result.Diagnostics[0].Message.Description; !strings.Contains(description, `"`+testCase.token+`"`) {
				t.Fatalf("expected the message to name %q, got: %s", testCase.token, description)
			}
		})
	}
}

func TestConsistencyNoShoutingRespectsTheAllowOption(t *testing.T) {
	t.Parallel()

	sourceText := "// the WIDGET subsystem owns this\nexport const Value = 1;\n"

	withoutOption := rule_testing.Run(t, ConsistencyNoShouting, shoutingFile, sourceText)
	rule_testing.ExpectFindings(t, withoutOption, "shoutingInComment")

	withOption := rule_testing.RunWithOptions(t, ConsistencyNoShouting, shoutingFile, sourceText,
		ConsistencyNoShoutingOptions{Allow: []string{"WIDGET"}})
	rule_testing.ExpectClean(t, withOption)
}

// The message has to name what it saw. A finding that says only "this comment shouts" makes the
// reader rescan a paragraph to find the word that tripped it.
func TestConsistencyNoShoutingNamesTheTokens(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, ConsistencyNoShouting, shoutingFile,
		"// NEVER cache this, it is REALLY bad\nexport const Value = 1;\n")
	rule_testing.ExpectFindings(t, result, "shoutingInComment")

	description := result.Diagnostics[0].Message.Description
	for _, want := range []string{`"NEVER"`, `"REALLY"`} {
		if !strings.Contains(description, want) {
			t.Fatalf("expected the message to name %s, got: %s", want, description)
		}
	}
}

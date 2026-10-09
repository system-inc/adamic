package core

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// The verdicts, spans and suggestions ESLint gives are in preserve_caught_error_corpus_test.go, read
// off the installed ESLint. The tests here pin what that table cannot: the one place the rule departs
// from ESLint, the decoder, the checker's two jobs, and the messages.

// TestPreserveCaughtErrorFiresAndStaysQuiet is the fixture pair in its smallest form.
func TestPreserveCaughtErrorFiresAndStaysQuiet(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, PreserveCaughtError, "input.ts",
		`try { a(); } catch (err) { throw new Error("m"); }`), "missingCause")
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, PreserveCaughtError, "input.ts",
		`try { a(); } catch (err) { throw new Error("m", { cause: err.message }); }`), "incorrectCause")
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, PreserveCaughtError, "input.ts",
		`try { a(); } catch ({ message }) { throw new Error(message); }`), "partiallyLostError")
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, PreserveCaughtError, "input.ts",
		`try { a(); } catch (err) { { const err = 1; throw new Error("m", { cause: err }); } }`), "caughtErrorShadowed")
	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, PreserveCaughtError, "input.ts",
		`try { a(); } catch { throw new Error("m"); }`, PreserveCaughtErrorOptions{RequireCatchParameter: true}),
		"missingCatchErrorParam")

	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, PreserveCaughtError, "input.ts",
		`try { a(); } catch (err) { throw new Error("m", { cause: err }); }`))
	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, PreserveCaughtError, "input.ts",
		`try { a(); } catch (err) { throw new Error("m", { "cause": err }); }`))
}

// TestPreserveCaughtErrorWritesIntoTheArgumentListTheParserFound pins the one suggestion that departs
// from ESLint.
//
// ESLint inserts after the first opening parenthesis following the callee, which for
// `new Error<() => void>()` is the one inside the type arguments: measured on the installed 10.8.1,
// it writes `new Error<("", { cause: err }) => void>()`, which no longer parses as the same code. The
// rule inserts at the argument list's own position instead, which also steps past a parenthesis in
// a block or line comment.
func TestPreserveCaughtErrorWritesIntoTheArgumentListTheParserFound(t *testing.T) {
	t.Parallel()

	vectors := []struct {
		before string
		after  string
	}{
		{
			before: `try { a(); } catch (err) { throw new Error<() => void>(); }`,
			after:  `try { a(); } catch (err) { throw new Error<() => void>("", { cause: err }); }`,
		},
		{
			before: `try { a(); } catch (err) { throw new Error/* ( */(); }`,
			after:  `try { a(); } catch (err) { throw new Error/* ( */("", { cause: err }); }`,
		},
		{
			before: "try { a(); } catch (err) {\n  throw new Error // (\n  ();\n}",
			after:  "try { a(); } catch (err) {\n  throw new Error // (\n  (\"\", { cause: err });\n}",
		},
	}
	for index, vector := range vectors {
		result := rule_testing.RunTyped(t, PreserveCaughtError, "input.ts", vector.before)
		if len(result.Diagnostics) != 1 || len(result.Diagnostics[0].Suggestions) != 1 {
			t.Errorf("vector %d: want one finding with one suggestion, got %d findings", index, len(result.Diagnostics))
			continue
		}
		if applied := applySuggestion(t, result.SourceFile.Text(), result.Diagnostics[0].Suggestions[0]); applied != vector.after+"\n" {
			t.Errorf("vector %d: the suggestion writes\n  %q\nwant\n  %q", index, applied, vector.after+"\n")
		}
	}
}

// TestPreserveCaughtErrorRequiresTheGlobalBinding pins the checker's first job.
//
// A local declaration of a built-in's name shadows the global, and its constructor takes whatever it
// takes, so the throw is not this rule's. Naming the class in errorClassNames makes it the rule's
// again, at the position given, which is what a project that wraps Error under its own name wants.
func TestPreserveCaughtErrorRequiresTheGlobalBinding(t *testing.T) {
	t.Parallel()

	shadowed := []string{
		"class Error {}\ntry { a(); } catch (err) { throw new Error(\"m\"); }",
		"let Error = X;\ntry { a(); } catch (err) { throw new Error(\"m\"); }",
		"function TypeError() {}\ntry { a(); } catch (err) { throw new TypeError(\"m\"); }",
	}
	for _, source := range shadowed {
		rule_testing.ExpectClean(t, rule_testing.RunTyped(t, PreserveCaughtError, "input.ts", source))
	}

	named := PreserveCaughtErrorOptions{ErrorClassNames: []PreserveCaughtErrorClassName{{Name: "Error", ArgumentPosition: 2}}}
	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, PreserveCaughtError, "input.ts", shadowed[0], named),
		"missingCause")
}

// TestPreserveCaughtErrorComparesTheCauseBySymbol pins the checker's second job, in both directions,
// in one-line form so a mutant comparing names has something small to fail against.
func TestPreserveCaughtErrorComparesTheCauseBySymbol(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, PreserveCaughtError, "input.ts",
		`try { a(); } catch (err) { if (w) { throw new Error("m", { cause: err }); } }`))
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, PreserveCaughtError, "input.ts",
		`try { a(); } catch (err) { if (w) { const err = other; throw new Error("m", { cause: err }); } }`),
		"caughtErrorShadowed")

	// A shorthand names the variable `cause`, read through the shorthand to its value's symbol.
	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, PreserveCaughtError, "input.ts",
		`try { a(); } catch (cause) { throw new Error("m", { cause }); }`))
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, PreserveCaughtError, "input.ts",
		`try { a(); } catch (cause) { { const cause = 1; throw new Error("m", { cause }); } }`),
		"caughtErrorShadowed")
}

// TestPreserveCaughtErrorDecodesErrorClassNames routes both spellings through the registered decoder,
// since that is what turns a configuration into options.
func TestPreserveCaughtErrorDecodesErrorClassNames(t *testing.T) {
	t.Parallel()

	decode := rule.DecodeOptionsInto[PreserveCaughtErrorOptions]()

	decoded, err := decode([]byte(`{"errorClassNames": ["AppError", {"name": "WideError", "argumentPosition": 3}]}`))
	if err != nil {
		t.Fatalf("decoding both spellings failed: %v", err)
	}
	options := decoded.(PreserveCaughtErrorOptions)
	want := []PreserveCaughtErrorClassName{{Name: "AppError", ArgumentPosition: 2}, {Name: "WideError", ArgumentPosition: 3}}
	if len(options.ErrorClassNames) != 2 || options.ErrorClassNames[0] != want[0] || options.ErrorClassNames[1] != want[1] {
		t.Errorf("decoded %#v, want %#v", options.ErrorClassNames, want)
	}

	// The same name twice takes the later position, as ESLint's Map does. Position 3 reads the third
	// argument, so the object in second place is not the options and the throw has none.
	twice, err := decode([]byte(`{"errorClassNames": [{"name": "AppError", "argumentPosition": 2}, {"name": "AppError", "argumentPosition": 3}]}`))
	if err != nil {
		t.Fatalf("decoding a repeated name failed: %v", err)
	}
	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, PreserveCaughtError, "input.ts",
		`try { a(); } catch (err) { throw new AppError("m", { cause: err }); }`, twice), "missingCause")

	refused := []string{
		`{"errorClassNames": [{"name": "AppError", "argumentPosition": 0}]}`,
		`{"errorClassNames": [{"name": "AppError"}]}`,
		`{"errorClassNames": [{"argumentPosition": 2}]}`,
		`{"errorClassNames": [{"name": "AppError", "argumentPosition": 2, "extra": true}]}`,
		`{"errorClassNames": [{"name": "AppError", "argumentPosition": 1.5}]}`,
		`{"errorClassNames": [7]}`,
	}
	for _, configuration := range refused {
		if _, err := decode([]byte(configuration)); err == nil {
			t.Errorf("ESLint's schema refuses %s, and the decoder accepted it", configuration)
		}
	}
}

// TestPreserveCaughtErrorRequireCatchParameterDefaultsOff pins the default through the decoder and
// through nil, which is what a rule configured as bare "error" is handed.
func TestPreserveCaughtErrorRequireCatchParameterDefaultsOff(t *testing.T) {
	t.Parallel()

	bareCatch := `try { a(); } catch { throw new Error("m"); }`
	decode := rule.DecodeOptionsInto[PreserveCaughtErrorOptions]()

	rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, PreserveCaughtError, "input.ts", bareCatch, nil))

	explicitFalse, err := decode([]byte(`{"requireCatchParameter": false}`))
	if err != nil {
		t.Fatalf("decoding an explicit false failed: %v", err)
	}
	rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, PreserveCaughtError, "input.ts", bareCatch, explicitFalse))

	explicitTrue, err := decode([]byte(`{"requireCatchParameter": true}`))
	if err != nil {
		t.Fatalf("decoding an explicit true failed: %v", err)
	}
	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(t, PreserveCaughtError, "input.ts", bareCatch, explicitTrue),
		"missingCatchErrorParam")

	// A parameterless catch that throws nothing new is silent even with the option on: the finding
	// is about a throw that cannot carry the caught error, not about the clause.
	rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(t, PreserveCaughtError, "input.ts",
		`try { a(); } catch { log(); }`, explicitTrue))
}

// TestPreserveCaughtErrorRequiresTheTypedHarness fails loudly if somebody reverts the checker
// declaration, which would otherwise buy a vacuous green.
//
// The nil-checker guard covers every message. A built-in's global test already declines without a
// checker, but a class named in errorClassNames is recognized by name alone, so without the guard its
// throw would reach the missing-parameter and destructuring findings and then the cause comparison,
// which asks the checker. Uniform silence is the answer: a file whose program failed to build should
// read as a rule that could not run, not as one with a few odd complaints.
func TestPreserveCaughtErrorRequiresTheTypedHarness(t *testing.T) {
	t.Parallel()

	if !PreserveCaughtError.NeedsTypeChecker {
		t.Fatalf("this rule resolves symbols and must declare NeedsTypeChecker")
	}
	options := PreserveCaughtErrorOptions{
		RequireCatchParameter: true,
		ErrorClassNames:       []PreserveCaughtErrorClassName{{Name: "AppError", ArgumentPosition: 2}},
	}
	sources := []string{
		`try { a(); } catch (err) { throw new Error("m"); }`,
		`try { a(); } catch { throw new Error("m"); }`,
		`try { a(); } catch ({ message }) { throw new Error(message); }`,
		// A class named in errorClassNames needs no checker to be recognized, so this is the input
		// that reaches the cause comparison, which does.
		`try { a(); } catch (err) { throw new AppError("m", { cause: other }); }`,
	}
	for _, source := range sources {
		if result := rule_testing.RunWithOptions(t, PreserveCaughtError, "input.ts", source, options); len(result.Diagnostics) != 0 {
			t.Errorf("the untyped harness hands the rule a nil checker, and it must decline:\n%s", source)
		}
		if result := rule_testing.RunTypedWithOptions(t, PreserveCaughtError, "input.ts", source, options); len(result.Diagnostics) != 1 {
			t.Errorf("the typed harness must reach one finding, got %d:\n%s", len(result.Diagnostics), source)
		}
	}
}

// TestPreserveCaughtErrorMessagesSayDifferentThings guards the messages against each other, with the
// ids typed here as literals: a mutation rewriting a constant moves both sides of a comparison against
// the constant, and the assertion stays green.
func TestPreserveCaughtErrorMessagesSayDifferentThings(t *testing.T) {
	t.Parallel()

	messages := map[string]rule.Message{
		"missingCause":           messageMissingCause,
		"incorrectCause":         messageIncorrectCause,
		"missingCatchErrorParam": messageMissingCatchErrorParam,
		"partiallyLostError":     messagePartiallyLostError,
		"caughtErrorShadowed":    messageCaughtErrorShadowed,
		"includeCause":           messageIncludeCause,
	}
	descriptions := map[string]string{}
	for id, message := range messages {
		if message.Id != id {
			t.Errorf("the message filed under %s has id %q", id, message.Id)
		}
		if other, taken := descriptions[message.Description]; taken {
			t.Errorf("%s and %s share a description", id, other)
		}
		descriptions[message.Description] = id
		if !strings.Contains(message.Description, "`cause`") {
			t.Errorf("%s must name the `cause` option, since that is what every finding is about", id)
		}
	}
}

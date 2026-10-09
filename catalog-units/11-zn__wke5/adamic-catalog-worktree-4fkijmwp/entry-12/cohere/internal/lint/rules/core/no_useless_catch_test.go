package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

func TestNoUselessCatchReportsBareRethrows(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
		wantId string
	}{
		{"plain rethrow", `try { foo(); } catch (err) { throw err; }`, "unnecessaryCatch"},
		{"rethrow with a finally", `try { foo(); } catch (err) { throw err; } finally { cleanUp(); }`, "unnecessaryCatchClause"},
		{"comment before the rethrow", `try { foo(); } catch (err) { /* comment */ throw err; }`, "unnecessaryCatch"},
		{"typed catch binding", `try { foo(); } catch (err: unknown) { throw err; }`, "unnecessaryCatch"},
		{"parenthesized rethrow", `try { foo(); } catch (err) { throw (err); }`, "unnecessaryCatch"},
		{"non-null assertion is erased", `try { foo(); } catch (err) { throw err!; }`, "unnecessaryCatch"},
		{"as expression is erased", `try { foo(); } catch (err) { throw err as Error; }`, "unnecessaryCatch"},
		{"statements after the rethrow are unreachable", `try { foo(); } catch (err) { throw err; unreachable(); }`, "unnecessaryCatch"},
		{"nested inside another try", `try { try { foo(); } catch (err) { throw err; } } catch (outer) { log(outer); }`, "unnecessaryCatch"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoUselessCatch, "file.ts", testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.wantId)
		})
	}
}

func TestNoUselessCatchAcceptsClausesThatDoWork(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
	}{
		{"logs the error", `try { foo(); } catch (err) { console.error(err); }`},
		{"logs with a finally", `try { foo(); } catch (err) { console.error(err); } finally { bar(); }`},
		{"work before the rethrow", `try { foo(); } catch (err) { doSomethingBeforeRethrow(); throw err; }`},
		{"throws a property of the error", `try { foo(); } catch (err) { throw err.message; }`},
		{"throws a new error", `try { foo(); } catch (err) { throw new Error("whoops"); }`},
		{"throws a different binding", `try { foo(); } catch (err) { throw bar; }`},
		{"empty catch body", `try { foo(); } catch (err) { }`},
		{"object destructured binding", `try { foo(); } catch ({ err }) { throw err; }`},
		{"array destructured binding", `try { foo(); } catch ([ err ]) { throw err; }`},
		{"optional catch binding", `try { throw new Error("foo"); } catch { throw new Error("foo"); }`},
		{"try with only a finally", `try { foo(); } finally { cleanUp(); }`},
		{"rethrow wrapped in a call", `try { foo(); } catch (err) { throw wrap(err); }`},
		{"throw inside a nested block, not first", `try { foo(); } catch (err) { if (a) { throw err; } log(err); }`},
		{"a throw elsewhere in the function", `function f() { throw error; }`},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NoUselessCatch, "file.ts", testCase.source))
		})
	}
}

// Without a finally the whole try is dead, so the range must cover the try statement; with one, only
// the catch clause is, so the range must start at `catch`. Reporting the wrong node would name a
// repair that deletes a finally the author needs.
func TestNoUselessCatchReportsTheNodeMatchingTheRepair(t *testing.T) {
	t.Parallel()

	const withoutFinally = `try { foo(); } catch (err) { throw err; }`
	result := rule_testing.Run(t, NoUselessCatch, "file.ts", withoutFinally)
	rule_testing.ExpectFindings(t, result, "unnecessaryCatch")
	if got := withoutFinally[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]; got != withoutFinally {
		t.Fatalf("reported range = %q, want the whole try statement", got)
	}

	const withFinally = `try { foo(); } catch (err) { throw err; } finally { cleanUp(); }`
	result = rule_testing.Run(t, NoUselessCatch, "file.ts", withFinally)
	rule_testing.ExpectFindings(t, result, "unnecessaryCatchClause")
	const wantClause = `catch (err) { throw err; }`
	if got := withFinally[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]; got != wantClause {
		t.Fatalf("reported range = %q, want %q", got, wantClause)
	}
}

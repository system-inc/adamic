package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestRegExpUncaughtDiagnosticsRefuse(t *testing.T) {
	for _, source := range []string{
		`new RegExp('(');`,
		`function invalid(): RegExp { return new RegExp('('); } invalid();`,
		`const invalid = (): RegExp => new RegExp('('); invalid();`,
	} {
		_, err := lowerSource(t, source)
		var refusal *NotYet
		if !errors.As(err, &refusal) || !strings.Contains(err.Error(), "unhandled RegExp SyntaxError") {
			t.Fatalf("unhandled parser diagnostic must refuse: %s: %v", source, err)
		}
	}
}

func TestRegExpHandledDiagnosticsAccepted(t *testing.T) {
	for _, source := range []string{
		`try { new RegExp('('); } catch (error) { console.log(error instanceof SyntaxError ? "syntax" : "other"); }`,
		`function invalid(): RegExp { return new RegExp('('); } try { invalid(); } catch (error) { console.log(error instanceof SyntaxError ? "syntax" : "other"); }`,
		`const invalid = (): RegExp => new RegExp('('); try { invalid(); } catch (error) { console.log(error instanceof SyntaxError ? "syntax" : "other"); }`,
	} {
		if _, err := lowerSource(t, source); err != nil {
			t.Fatalf("handled parser diagnostic must compile: %s: %v", source, err)
		}
	}
}

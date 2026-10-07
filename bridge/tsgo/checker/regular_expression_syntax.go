package checker

import (
	"fmt"
	"strings"

	esregexp "github.com/system-inc/adamic/bridge/tsgo/regular_expression_syntax"
)

// Raw syntax validation. No rule predicate, message, span or edit is returned.
func regularExpressionSyntax(out *fields, question string) (string, error) {
	parts := strings.SplitN(question, "\n", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("regular-expression-syntax requires a pattern")
	}
	_, err := esregexp.Compile(parts[1], "")
	out.yes(err == nil)
	return out.String(), nil
}

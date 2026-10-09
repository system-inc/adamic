package printing

import "errors"

// Syntax marks err as the source failing to parse, keeping its message.
//
// The format phase treats a file that does not parse as a skip and every other failure as a broken
// formatter, so it has to tell the two apart. It used to tell by matching the goja engine's messages,
// which belong to upstream and differ by parser: JSON's "unexpected" was read as a broken formatter.
// A parser that wraps its failures here is recognized whatever its wording.
func Syntax(err error) error {
	if err == nil {
		return nil
	}
	return syntaxError{err}
}

// IsSyntax reports whether err, or anything it wraps, was marked by Syntax.
func IsSyntax(err error) bool {
	var marked syntaxError
	return errors.As(err, &marked)
}

type syntaxError struct{ error }

func (marked syntaxError) Unwrap() error { return marked.error }

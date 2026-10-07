package regexp

import "errors"

type AdamicSlot02Batch8Result struct {
	Kind, Set, Rune, Width int
	Negated                bool
	Error, Cause           string
}

func adamicSlot02Batch8Result(d decodedEscape, err error) AdamicSlot02Batch8Result {
	r := AdamicSlot02Batch8Result{Kind: int(d.kind), Set: int(d.set), Rune: int(d.r), Width: d.width, Negated: d.negated}
	if err != nil {
		r.Error = err.Error()
		if errors.Is(err, ErrUnsupportedSyntax) {
			r.Cause = ErrUnsupportedSyntax.Error()
		} else {
			panic("unexpected regexp error cause")
		}
		if errors.Unwrap(err) != ErrUnsupportedSyntax {
			panic("unexpected regexp unwrap")
		}
	}
	return r
}
func AdamicSlot02Batch8Identity(r int, size int, u, cl bool, groups int, named bool) AdamicSlot02Batch8Result {
	d, e := identityEscape(rune(r), size, escapeContext{unicode: u, inClass: cl, groups: groups, named: named})
	return adamicSlot02Batch8Result(d, e)
}
func AdamicSlot02Batch8Literal(r int) string { return literalRune(rune(r)) }

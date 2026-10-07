package tailwind

import "github.com/microsoft/TypeScript/tsc/shim/core"

func AdamicTrim(start, end, leading, trailing int) (int, int) {
	r := trimDelimiters(core.NewTextRange(start, end), leading, trailing)
	return r.Pos(), r.End()
}

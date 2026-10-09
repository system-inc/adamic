// Package locale re-exports typescript-go's diagnostic locale type.
//
// Hand-written rather than generated, for the same reason the parser and format shims are: typescript/generate_shims
// covers the packages a headless linter needs, and until a diagnostic had to be rendered this was not
// among them.
//
// One type and nothing else, because one type is all that is needed. A compiler diagnostic carries a
// message template and its arguments rather than finished text, and `Diagnostic.Localize` is what
// turns those into a sentence. It takes a locale, so printing a type error at all requires naming
// this type once.
package locale

import "github.com/microsoft/TypeScript/tsc/internal/locale"

// Locale selects the language a diagnostic renders in.
//
// The zero value selects the compiler's built-in English text, which is what every other line this
// tool prints already uses.
type Locale = locale.Locale

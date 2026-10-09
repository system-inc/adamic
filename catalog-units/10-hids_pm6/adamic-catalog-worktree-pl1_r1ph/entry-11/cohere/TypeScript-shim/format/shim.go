// Package format re-exports typescript-go's formatter.
//
// Hand-written rather than generated, for the same reason the parser shim is: typescript/generate_shims covers the
// packages a headless linter needs, and formatting is not among them.
//
// Note this is internal/format, the editor's formatter, and not internal/printer, which is the
// emitter. The emitter produces JavaScript from an AST and does not preserve comments or the
// author's line breaks. The formatter returns text changes against the original source, which is
// both what formatting means and the same shape our fixes already take.
package format

import (
	"context"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/format"
	"github.com/microsoft/TypeScript/tsc/internal/ls/lsutil"
)

// FormatCodeSettings is what the formatter is configured with.
//
// It lives in `ls/lsutil` rather than in `format`: the settings are the language server's, and the
// formatter is one of several consumers. It also became a value rather than a pointer in the move to
// `microsoft/TypeScript`, so callers hold a copy instead of sharing one.
type FormatCodeSettings = lsutil.FormatCodeSettings

// SemicolonPreference is what the formatter does about semicolons it did not write.
//
// It became a named string type in the move to `microsoft/TypeScript`; it was a bare string before,
// so an assignment of "ignore" that used to compile now needs the constant.
type SemicolonPreference = lsutil.SemicolonPreference

const (
	SemicolonPreferenceIgnore = lsutil.SemicolonPreferenceIgnore
	SemicolonPreferenceInsert = lsutil.SemicolonPreferenceInsert
	SemicolonPreferenceRemove = lsutil.SemicolonPreferenceRemove
)

// WithFormatCodeSettings puts settings on a context, which is how the formatter reads them.
func WithFormatCodeSettings(ctx context.Context, settings FormatCodeSettings, newLine string) context.Context {
	return format.WithFormatCodeSettings(ctx, settings, newLine)
}

// GetDefaultFormatCodeSettings returns TypeScript's own defaults.
//
// It no longer takes a newline: the setting moved inside the struct, so a caller that wants a
// specific line ending sets NewLineCharacter on the result rather than passing it here.
func GetDefaultFormatCodeSettings() FormatCodeSettings {
	return lsutil.GetDefaultFormatCodeSettings()
}

// FormatDocument returns the text changes that would format a whole file.
func FormatDocument(ctx context.Context, sourceFile *ast.SourceFile) []core.TextChange {
	return format.FormatDocument(ctx, sourceFile)
}

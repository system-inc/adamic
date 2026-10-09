// Package formatoptions is what a file formats with: the options, Prettier's names for them, and how a
// directory resolves them from its repository's config.
//
// It carries no formatter. The native printers and the goja oracle both take these options, and
// keeping them here is what lets the cohere binary link the printers without linking goja.
package formatoptions

// Options are the Prettier settings a file formats with.
//
// The last five are the options a config in our repositories names beyond ahra's five. Four of them are
// Prettier 3's defaults, so setting them changes nothing, and they are carried anyway so a resolved
// config is passed whole rather than filtered by someone's belief about which keys matter.
// BracketSameLine is the one that is not a default, and leaving it out is what made api-phi-health's JSX
// measure against the wrong formatter.
type Options struct {
	TabWidth    int
	UseTabs     bool
	Semi        bool
	SingleQuote bool
	PrintWidth  int

	TrailingComma   string
	BracketSpacing  bool
	BracketSameLine bool
	ArrowParens     string
	EndOfLine       string
}

// Default mirrors the prettier block in the ahra package.json, over Prettier's own defaults.
//
// A caller formatting a repository it can locate should use Resolve instead. This exists for tests and
// for callers with no directory to resolve from.
func Default() Options {
	options := PrettierDefaults()
	options.TabWidth = 4
	options.UseTabs = false
	options.Semi = true
	options.SingleQuote = true
	options.PrintWidth = 120
	return options
}

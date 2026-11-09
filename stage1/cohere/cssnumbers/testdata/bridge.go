package css

import "github.com/system-inc/cohere/internal/format/format_options"

// Bridge calls the actual Go printer, not a second implementation.
func AdamicAdjustStrings(value string, single bool) string {
	options := format_options.PrettierDefaults()
	options.SingleQuote = single
	return adjustNumbers(adjustStrings(value, &printerOptions{Settings: printSettings{Options: options, parser: "css"}}))
}

package css

import "github.com/system-inc/cohere/internal/format/formatoptions"

// Bridge calls the actual Go printer, not a second implementation.
func AdamicAdjustStrings(value string, single bool) string {
	options := formatoptions.PrettierDefaults()
	options.SingleQuote = single
	return adjustStrings(value, &printerOptions{Settings: printSettings{Options: options, parser: "css"}})
}

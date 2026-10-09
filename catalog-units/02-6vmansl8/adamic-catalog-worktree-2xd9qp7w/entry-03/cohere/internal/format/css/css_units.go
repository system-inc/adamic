package css

// src/language-css/print/css-units.evaluate.js: css-units-list 2.1.0, keyed by its lowercased spelling,
// mapped to its canonical one.

import "strings"

// cssUnitsList is css-units-list's default export, in its order.
var cssUnitsList = []string{
	"em", "rem", "ex", "rex", "cap", "rcap", "ch", "rch", "ic", "ric", "lh", "rlh",
	"vw", "svw", "lvw", "dvw", "vh", "svh", "lvh", "dvh", "vi", "svi", "lvi", "dvi",
	"vb", "svb", "lvb", "dvb", "vmin", "svmin", "lvmin", "dvmin", "vmax", "svmax", "lvmax", "dvmax",
	"cm", "mm", "Q", "in", "pt", "pc", "px",
	"deg", "grad", "rad", "turn",
	"s", "ms",
	"Hz", "kHz",
	"dpi", "dpcm", "dppx", "x",
	"cqw", "cqh", "cqi", "cqb", "cqmin", "cqmax",
	"fr",
}

// cssUnits is CSS_UNITS: new Map(cssUnits.map((unit) => [unit.toLowerCase(), unit])).
var cssUnits = func() map[string]string {
	units := make(map[string]string, len(cssUnitsList))
	for _, unit := range cssUnitsList {
		units[strings.ToLower(unit)] = unit
	}
	return units
}()

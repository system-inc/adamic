// isColor: the color predicate, ported from Tailwind 4.3.3 `src/utils/is-color.ts`.
//
// Three ways a value is a color, and they are checked in this order because the cheap ones come
// first: it starts with `#`, it starts with a known color function, or it is a named color.
//
// What is *not* here is any validation. `#zzzz` is a color, `rgb(` with nothing after it is a color,
// and `color-mix(anything)` is a color. The engine is deciding which utility a value selects, not
// whether the CSS is valid, and a browser that rejects the declaration is the layer that catches
// the rest. Tightening this would change which utility gets picked.
package tailwind

import "strings"

// colorFunctions ports `IS_COLOR_FN = /^(rgba?|hsla?|hwb|color|(ok)?(lab|lch)|light-dark|color-mix|--alpha)\(/i`.
//
// The regexp is case-insensitive and matched as a prefix. Note that `color(` is in the alternation
// and `color-mix(` is too, but the alternation is unordered in effect because each is followed by
// the required `\(`: `color-mix(` cannot match the `color` branch, since the paren does not
// immediately follow. Listed here as separate literals so that property is explicit rather than
// emergent.
var colorFunctions = []string{
	"rgb(", "rgba(", "hsl(", "hsla(", "hwb(",
	"color(", "lab(", "lch(", "oklab(", "oklch(",
	"light-dark(", "color-mix(", "--alpha(",
}

// IsColor reports whether value is a color the way Tailwind reads colors.
//
// Callers should reach for InferDataType with a type list rather than this directly, unless they
// genuinely mean "is this a color" without the `var(` short-circuit. `IsColor("var(--color-red)")`
// is false, and that is correct for this predicate and wrong as a reading of the value: the engine
// declines to type `var()` at all, one level up.
func IsColor(value string) bool {
	if strings.HasPrefix(value, "#") {
		return true
	}
	for _, function := range colorFunctions {
		if len(value) >= len(function) && strings.EqualFold(value[:len(function)], function) {
			return true
		}
	}
	return IsNamedColor(value)
}

// IsNamedColor reports whether value is a CSS named color, keyword, or system color.
//
// Lowercased before lookup, matching `NAMED_COLORS.has(value.toLowerCase())`. JavaScript's
// `toLowerCase` is Unicode-aware and Go's strings.ToLower is too; they agree on every ASCII input,
// and every member of the set is ASCII, so a non-ASCII input fails the lookup under both.
func IsNamedColor(value string) bool {
	return namedColors[strings.ToLower(value)]
}

// namedColors is the upstream set: CSS Level 1, Level 2/3, the `transparent` and `currentcolor`
// keywords, and the system colors. Generated from the source rather than retyped.
var namedColors = map[string]bool{
	"black": true, "silver": true, "gray": true, "white": true,
	"maroon": true, "red": true, "purple": true, "fuchsia": true,
	"green": true, "lime": true, "olive": true, "yellow": true,
	"navy": true, "blue": true, "teal": true, "aqua": true,
	"aliceblue": true, "antiquewhite": true, "aquamarine": true, "azure": true,
	"beige": true, "bisque": true, "blanchedalmond": true, "blueviolet": true,
	"brown": true, "burlywood": true, "cadetblue": true, "chartreuse": true,
	"chocolate": true, "coral": true, "cornflowerblue": true, "cornsilk": true,
	"crimson": true, "cyan": true, "darkblue": true, "darkcyan": true,
	"darkgoldenrod": true, "darkgray": true, "darkgreen": true, "darkgrey": true,
	"darkkhaki": true, "darkmagenta": true, "darkolivegreen": true, "darkorange": true,
	"darkorchid": true, "darkred": true, "darksalmon": true, "darkseagreen": true,
	"darkslateblue": true, "darkslategray": true, "darkslategrey": true, "darkturquoise": true,
	"darkviolet": true, "deeppink": true, "deepskyblue": true, "dimgray": true,
	"dimgrey": true, "dodgerblue": true, "firebrick": true, "floralwhite": true,
	"forestgreen": true, "gainsboro": true, "ghostwhite": true, "gold": true,
	"goldenrod": true, "greenyellow": true, "grey": true, "honeydew": true,
	"hotpink": true, "indianred": true, "indigo": true, "ivory": true,
	"khaki": true, "lavender": true, "lavenderblush": true, "lawngreen": true,
	"lemonchiffon": true, "lightblue": true, "lightcoral": true, "lightcyan": true,
	"lightgoldenrodyellow": true, "lightgray": true, "lightgreen": true, "lightgrey": true,
	"lightpink": true, "lightsalmon": true, "lightseagreen": true, "lightskyblue": true,
	"lightslategray": true, "lightslategrey": true, "lightsteelblue": true, "lightyellow": true,
	"limegreen": true, "linen": true, "magenta": true, "mediumaquamarine": true,
	"mediumblue": true, "mediumorchid": true, "mediumpurple": true, "mediumseagreen": true,
	"mediumslateblue": true, "mediumspringgreen": true, "mediumturquoise": true, "mediumvioletred": true,
	"midnightblue": true, "mintcream": true, "mistyrose": true, "moccasin": true,
	"navajowhite": true, "oldlace": true, "olivedrab": true, "orange": true,
	"orangered": true, "orchid": true, "palegoldenrod": true, "palegreen": true,
	"paleturquoise": true, "palevioletred": true, "papayawhip": true, "peachpuff": true,
	"peru": true, "pink": true, "plum": true, "powderblue": true,
	"rebeccapurple": true, "rosybrown": true, "royalblue": true, "saddlebrown": true,
	"salmon": true, "sandybrown": true, "seagreen": true, "seashell": true,
	"sienna": true, "skyblue": true, "slateblue": true, "slategray": true,
	"slategrey": true, "snow": true, "springgreen": true, "steelblue": true,
	"tan": true, "thistle": true, "tomato": true, "turquoise": true,
	"violet": true, "wheat": true, "whitesmoke": true, "yellowgreen": true,
	"transparent": true, "currentcolor": true, "canvas": true, "canvastext": true,
	"linktext": true, "visitedtext": true, "activetext": true, "buttonface": true,
	"buttontext": true, "buttonborder": true, "field": true, "fieldtext": true,
	"highlight": true, "highlighttext": true, "selecteditem": true, "selecteditemtext": true,
	"mark": true, "marktext": true, "graytext": true, "accentcolor": true,
	"accentcolortext": true,
}

// The framework's multi-declaration functional utilities.
//
// Waves 2b and 2c of #271785y: every framework root whose reading was measured rather than derived
// from the shape of its handler. The filter and backdrop-filter families, the skews, `ease`, the
// mask gradients, the grid templates, `transition`, `divide-x` and `divide-y`, and the rest.
//
// Not all of these emit several declarations. `aspect` and `col-span` emit one, and they are here
// rather than in frameworkutility.go because their bare-value handlers rewrite the value rather than
// pass it through, so the table that carries a property name has nothing to say about them.
//
// # Why these carry a reading and waves 1 and 2a carry a property
//
// The earlier waves emit one declaration, so naming its property is enough and PropertySort computes
// the rest. These do not, and the source cannot be read to find out how many. Twenty-one of them
// match one pattern in `utilities.ts` (a properties wrapper, a `--tw-*` declaration, a final
// declaration), and extracting that pattern predicts they all read alike. Measured against the
// engine, they do not:
//
//	grayscale     [334,339]#2      one --tw- variable plus filter
//	skew-3        [69,70,71]#3     writes both --tw-skew-x and --tw-skew-y
//	skew-x-3      [69,71]#2        writes one
//	ease-linear   [354]#2          two declarations sharing one property position
//
// `ease` is the case that settles the shape. Two declarations at a single order index means a table
// keyed on property names would have said `#1`, and nothing in the handler's text says otherwise. So
// the reading is measured per root rather than derived, and this table carries `{order, count}`.
//
// # The wrapper contributes nothing, and that is measured rather than assumed
//
// Every one of these calls a properties wrapper first: `filterProperties()`, `transformProperties()`,
// `backdropFilterProperties()`. Each returns `atRoot([...])` holding a dozen `@property` blocks, and
// PropertySort descends into rules and at-rules but not into at-root, so the wrapper adds zero to the
// reading. `grayscale` reading `[334,339]#2` rather than `#15` is the confirmation: thirteen
// `@property` declarations are structurally invisible.
//
// Counted at the tag while porting the handle bodies: there are 21 such wrappers, not the ten an
// earlier scoping named, and every one returns `atRoot` and contains no `decl(` call. None of them
// needed a representation in frameworkmultihandlers.go.
//
// # Repository-invariant, and checked rather than claimed
//
// Generated against ahra and against www-connected-app, which resolve different themes. Zero roots
// differ in their reading set, 256 answered on each. A root whose reading depended on a repository's
// tokens would show up as a differing set, and none does, which is what makes this checked in rather
// than derived live like the namespace half of the descriptor table.
package tailwind

// FrameworkMultiDeclarationUtility is one root whose reading was measured rather than derived.
//
// The value axes (theme keys, bare-value predicate, default) are here for the same reason they are in
// FrameworkFunctionalUtility: they decide whether a class resolves at all.
//
// Reading no longer decides what a candidate reads. The ported handle bodies in
// frameworkmultihandlers.go emit declarations and `readingOf` runs PropertySort over them, the same
// two steps the engine takes. Reading stays because it is the independent measurement the acceptance
// test compares those emitters against: corrupting one row fails that test while leaving the
// differential at full agreement, which is what makes the two sides genuinely separate claims.
type FrameworkMultiDeclarationUtility struct {
	// Reading is what this root reads for any value that resolves, as measured against the engine.
	//
	// Read by the acceptance test rather than by the resolution path. See the type's doc comment.
	Reading Reading
	// LiteralReadings are the root-defined keywords the handler answers directly.
	//
	// Most read exactly as their root does and are carried so the resolution path knows the keyword
	// is answered without a theme lookup. Four read differently: `ease-initial` reads `[]#1` where
	// every other `ease-*` reads `[354]#2`, and `divide-none`, `transition-none` and `translate-none`
	// each emit a shorter list than their root's ordinary path. Those four have an entry in
	// frameworkMultiLiteralEmitters; the rest deliberately do not.
	LiteralReadings map[string]Reading

	ThemeKeys        []string
	SupportsNegative bool
	// SupportsFractions lets `basis-1/2` resolve without `1/2` being a theme key.
	SupportsFractions bool
	// RefusesEmptyValue marks a root that compiles to nothing when written with no value.
	//
	// Measured per root rather than derived from the other fields: `blur` has no default value
	// and reads bare through a theme lookup, `gap` has no default and reads nothing. See ReadingFor.
	RefusesEmptyValue bool
	// RefusesArbitraryValue marks a root that takes only named values.
	//
	// `max-w-screen` is the one: it reads `max-w-screen-sm` through the `--breakpoint` namespace and
	// compiles `max-w-screen-[3px]` to nothing, because the registration accepts a theme key rather
	// than a value. Without this the table invents a reading for every arbitrary form.
	RefusesArbitraryValue bool
	// AcceptsColorKeywords marks a root that answers `current`, `inherit` and `transparent`.
	AcceptsColorKeywords bool
	DefaultValue         string
	DefaultValuePresent  bool
	BareValue            BareValueKind
	BareValueSuffix      string
}

// Description converts a row into the shape ResolveFunctionalUtilityValue takes.
func (utility FrameworkMultiDeclarationUtility) Description() *FunctionalUtilityDescription {
	description := &FunctionalUtilityDescription{
		ThemeKeys:           utility.ThemeKeys,
		SupportsNegative:    utility.SupportsNegative,
		SupportsFractions:   utility.SupportsFractions,
		DefaultValue:        utility.DefaultValue,
		DefaultValuePresent: utility.DefaultValuePresent,
	}
	// Shared with FrameworkFunctionalUtility rather than duplicated, so the two tables cannot drift
	// about what a BareValueKind means. An inline copy here is what left `aspect` with a nil handler
	// after BareValueFraction was added: the kind existed, the table row named it, and this function
	// had never heard of it.
	description.HandleBareValue = bareValueHandler(utility.BareValue, utility.BareValueSuffix)
	return description
}

// Reading returns what a candidate reads against this root, and whether it reads anything.
//
// The literal check runs before resolution, matching upstream: a root-defined keyword is answered by
// the handler directly and never reaches the theme. Checking after would let a repository declaring
// `--ease-initial` shadow the literal, which the engine does not allow.
func (utility FrameworkMultiDeclarationUtility) ReadingFor(root string, candidate *ParsedCandidate, theme *Theme, negative bool) (Reading, bool) {
	if candidate == nil {
		return Reading{}, false
	}
	if candidate.Value != nil && candidate.Value.Kind == ParsedValueKindNamed {
		if _, found := utility.LiteralReadings[candidate.Value.Value]; found {
			return utility.readingOf(root, candidate.Value.Value, ResolvedUtilityValue{})
		}
		// The universal colour keywords, which are in no theme namespace and infer as nothing.
		//
		// `accent-current` compiles and `current` is absent from `--color`: they are Tailwind's own
		// built-ins, resolved by the colour path before the theme is consulted at all. A colour root
		// therefore answers them at its ordinary reading, and without this the theme lookup fails and
		// the table declines eighteen classes the engine reads. Same bucket `bareReading` carries in
		// descriptor.go, for the same reason.
		if utility.AcceptsColorKeywords && colorKeywords[candidate.Value.Value] {
			return utility.readingOf(root, "", ResolvedUtilityValue{})
		}
	}
	// A valueless candidate is a per-root fact, not a consequence of the other fields.
	//
	// `blur` and `transition` read bare, through a theme lookup and a default respectively. `gap`,
	// `basis`, `size`, `leading`, `indent`, `space-x` and the border spacings compile to no CSS at
	// all: upstream's `spacingUtility` passes `defaultValue: null`, which is not the same as omitting
	// it, and the `--spacing` namespace answers a null candidate value, so the theme lookup succeeds
	// where the engine refuses.
	//
	// A first attempt keyed this on `DefaultValuePresent` and broke `blur`, which has no default and
	// answers anyway. So the flag is measured per root against the engine rather than derived from
	// the shape, which is the same lesson wave 2b's readings taught.
	if candidate.Value == nil && utility.RefusesEmptyValue {
		return Reading{}, false
	}
	if utility.RefusesArbitraryValue && candidate.Value != nil && candidate.Value.Kind == ParsedValueKindArbitrary {
		return Reading{}, false
	}
	resolved, produced := ResolveFunctionalUtilityValue(candidate, utility.Description(), theme, negative)
	if !produced {
		return Reading{}, false
	}
	return utility.readingOf(root, "", resolved)
}

// readingOf runs the root's ported handle body and sorts what it emits.
//
// The one place `Reading` is no longer read on the ordinary path. It stays on the row because it is
// the independent measurement frameworkmultihandlers_test.go compares the emitters against, and two
// statements of one fact are worth carrying only while something forces them to agree.
//
// An unported root reads nothing rather than falling back to the stored `Reading`, for the reason
// the single-declaration Emit gives: a fallback makes a missing emitter indistinguishable from a
// working one everywhere except the acceptance test.
func (utility FrameworkMultiDeclarationUtility) readingOf(root, literal string, resolved ResolvedUtilityValue) (Reading, bool) {
	nodes := utility.Emit(root, literal, resolved)
	if len(nodes) == 0 {
		return Reading{}, false
	}
	sorted := PropertySort(nodes)
	return Reading{Order: sorted.Order, Count: sorted.Count}, true
}

// FrameworkMultiDeclarationUtilities is the table.
var FrameworkMultiDeclarationUtilities = map[string]FrameworkMultiDeclarationUtility{
	"accent":              {Reading: Reading{Order: []int{310}, Count: 1}, ThemeKeys: []string{"--accent-color", "--color"}, RefusesEmptyValue: true, AcceptsColorKeywords: true},
	"aspect":              {Reading: Reading{Order: []int{41}, Count: 1}, ThemeKeys: []string{"--aspect"}, BareValue: BareValueFraction, LiteralReadings: map[string]Reading{"auto": Reading{Order: []int{41}, Count: 1}, "square": Reading{Order: []int{41}, Count: 1}, "video": Reading{Order: []int{41}, Count: 1}}},
	"auto-cols":           {Reading: Reading{Order: []int{117}, Count: 1}, ThemeKeys: []string{"--grid-auto-columns"}, BareValue: BareValueSpacingMultiplier, LiteralReadings: map[string]Reading{"auto": Reading{Order: []int{117}, Count: 1}, "fr": Reading{Order: []int{117}, Count: 1}, "max": Reading{Order: []int{117}, Count: 1}, "min": Reading{Order: []int{117}, Count: 1}}},
	"auto-rows":           {Reading: Reading{Order: []int{119}, Count: 1}, ThemeKeys: []string{"--grid-auto-rows"}, BareValue: BareValueSpacingMultiplier, LiteralReadings: map[string]Reading{"auto": Reading{Order: []int{119}, Count: 1}, "fr": Reading{Order: []int{119}, Count: 1}, "max": Reading{Order: []int{119}, Count: 1}, "min": Reading{Order: []int{119}, Count: 1}}},
	"backdrop-blur":       {Reading: Reading{Order: []int{340, 349}, Count: 3}, ThemeKeys: []string{"--backdrop-blur", "--blur"}, LiteralReadings: map[string]Reading{"none": Reading{Order: []int{340, 349}, Count: 3}}},
	"backdrop-brightness": {Reading: Reading{Order: []int{341, 349}, Count: 3}, ThemeKeys: []string{"--backdrop-brightness", "--brightness"}, BareValue: BareValuePositiveInteger, BareValueSuffix: "%"},
	"backdrop-contrast":   {Reading: Reading{Order: []int{342, 349}, Count: 3}, ThemeKeys: []string{"--backdrop-contrast", "--contrast"}, BareValue: BareValuePositiveInteger, BareValueSuffix: "%"},
	"backdrop-grayscale":  {Reading: Reading{Order: []int{343, 349}, Count: 3}, ThemeKeys: []string{"--backdrop-grayscale", "--grayscale"}, DefaultValue: "100%", DefaultValuePresent: true, BareValue: BareValuePositiveInteger, BareValueSuffix: "%"},
	"backdrop-hue-rotate": {Reading: Reading{Order: []int{344, 349}, Count: 3}, ThemeKeys: []string{"--backdrop-hue-rotate", "--hue-rotate"}, SupportsNegative: true, BareValue: BareValuePositiveInteger, BareValueSuffix: "deg"},
	"backdrop-invert":     {Reading: Reading{Order: []int{345, 349}, Count: 3}, ThemeKeys: []string{"--backdrop-invert", "--invert"}, DefaultValue: "100%", DefaultValuePresent: true, BareValue: BareValuePositiveInteger, BareValueSuffix: "%"},
	"backdrop-opacity":    {Reading: Reading{Order: []int{346, 349}, Count: 3}, ThemeKeys: []string{"--backdrop-opacity", "--opacity"}, BareValue: BareValueOpacity, BareValueSuffix: "%"},
	"backdrop-saturate":   {Reading: Reading{Order: []int{347, 349}, Count: 3}, ThemeKeys: []string{"--backdrop-saturate", "--saturate"}, BareValue: BareValuePositiveInteger, BareValueSuffix: "%"},
	"backdrop-sepia":      {Reading: Reading{Order: []int{348, 349}, Count: 3}, ThemeKeys: []string{"--backdrop-sepia", "--sepia"}, DefaultValue: "100%", DefaultValuePresent: true, BareValue: BareValuePositiveInteger, BareValueSuffix: "%"},
	"basis":               {Reading: Reading{Order: []int{51}, Count: 1}, ThemeKeys: []string{"--flex-basis", "--spacing", "--container"}, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"blur":                {Reading: Reading{Order: []int{330, 339}, Count: 2}, ThemeKeys: []string{"--blur"}, LiteralReadings: map[string]Reading{"none": Reading{Order: []int{330, 339}, Count: 2}}},
	"border-spacing":      {Reading: Reading{Order: []int{55}, Count: 3}, ThemeKeys: []string{"--border-spacing", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"border-spacing-x":    {Reading: Reading{Order: []int{55}, Count: 2}, ThemeKeys: []string{"--border-spacing", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"border-spacing-y":    {Reading: Reading{Order: []int{55}, Count: 2}, ThemeKeys: []string{"--border-spacing", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"bottom":              {Reading: Reading{Order: []int{13}, Count: 1}, ThemeKeys: []string{"--inset", "--spacing"}, SupportsNegative: true, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"brightness":          {Reading: Reading{Order: []int{331, 339}, Count: 2}, ThemeKeys: []string{"--brightness"}, BareValue: BareValuePositiveInteger, BareValueSuffix: "%"},
	"caret":               {Reading: Reading{Order: []int{309}, Count: 1}, ThemeKeys: []string{"--caret-color", "--color"}, RefusesEmptyValue: true, AcceptsColorKeywords: true},
	"col-span":            {Reading: Reading{Order: []int{18}, Count: 1}, BareValue: BareValuePositiveInteger, LiteralReadings: map[string]Reading{"full": Reading{Order: []int{18}, Count: 1}}},
	"content":             {Reading: Reading{Order: []int{357}, Count: 2}, ThemeKeys: []string{"--content"}},
	"contrast":            {Reading: Reading{Order: []int{332, 339}, Count: 2}, ThemeKeys: []string{"--contrast"}, BareValue: BareValuePositiveInteger, BareValueSuffix: "%"},
	"divide":              {Reading: Reading{Order: []int{139}, Count: 2}, ThemeKeys: []string{"--divide-color", "--border-color", "--color"}, LiteralReadings: map[string]Reading{"none": Reading{Order: []int{138}, Count: 3}}, RefusesEmptyValue: true, AcceptsColorKeywords: true},
	"divide-x":            {Reading: Reading{Order: []int{135}, Count: 5}, DefaultValue: "1px", DefaultValuePresent: true, ThemeKeys: []string{"--divide-width", "--border-width"}, BareValue: BareValuePositiveInteger, BareValueSuffix: "px"},
	"divide-y":            {Reading: Reading{Order: []int{136}, Count: 6}, DefaultValue: "1px", DefaultValuePresent: true, ThemeKeys: []string{"--divide-width", "--border-width"}, BareValue: BareValuePositiveInteger, BareValueSuffix: "px"},
	"ease":                {Reading: Reading{Order: []int{354}, Count: 2}, ThemeKeys: []string{"--ease"}, LiteralReadings: map[string]Reading{"linear": Reading{Order: []int{354}, Count: 2}, "initial": Reading{Order: nil, Count: 1}}},
	"end":                 {Reading: Reading{Order: []int{8}, Count: 1}, ThemeKeys: []string{"--inset", "--spacing"}, SupportsNegative: true, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"flex-grow":           {Reading: Reading{Order: []int{50}, Count: 1}, DefaultValue: "1", DefaultValuePresent: true, BareValue: BareValuePositiveInteger},
	"flex-shrink":         {Reading: Reading{Order: []int{49}, Count: 1}, DefaultValue: "1", DefaultValuePresent: true, BareValue: BareValuePositiveInteger},
	"font-stretch":        {Reading: Reading{Order: []int{300}, Count: 1}, BareValue: BareValueFontStretchPercentage, LiteralReadings: map[string]Reading{"100%": Reading{Order: []int{300}, Count: 1}, "105%": Reading{Order: []int{300}, Count: 1}, "110%": Reading{Order: []int{300}, Count: 1}, "125%": Reading{Order: []int{300}, Count: 1}, "150%": Reading{Order: []int{300}, Count: 1}, "200%": Reading{Order: []int{300}, Count: 1}, "50%": Reading{Order: []int{300}, Count: 1}, "75%": Reading{Order: []int{300}, Count: 1}, "90%": Reading{Order: []int{300}, Count: 1}, "95%": Reading{Order: []int{300}, Count: 1}}},
	"gap":                 {Reading: Reading{Order: []int{130}, Count: 1}, ThemeKeys: []string{"--gap", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"gap-x":               {Reading: Reading{Order: []int{131}, Count: 1}, ThemeKeys: []string{"--gap", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"gap-y":               {Reading: Reading{Order: []int{132}, Count: 1}, ThemeKeys: []string{"--gap", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"grayscale":           {Reading: Reading{Order: []int{334, 339}, Count: 2}, ThemeKeys: []string{"--grayscale"}, DefaultValue: "100%", DefaultValuePresent: true, BareValue: BareValuePositiveInteger, BareValueSuffix: "%"},
	"grid-cols":           {Reading: Reading{Order: []int{120}, Count: 1}, ThemeKeys: []string{"--grid-template-columns"}, BareValue: BareValueGridRepeat, LiteralReadings: map[string]Reading{"none": Reading{Order: []int{120}, Count: 1}, "subgrid": Reading{Order: []int{120}, Count: 1}}},
	"grid-rows":           {Reading: Reading{Order: []int{121}, Count: 1}, ThemeKeys: []string{"--grid-template-rows"}, BareValue: BareValueGridRepeat, LiteralReadings: map[string]Reading{"none": Reading{Order: []int{121}, Count: 1}, "subgrid": Reading{Order: []int{121}, Count: 1}}},
	"h":                   {Reading: Reading{Order: []int{42}, Count: 1}, ThemeKeys: []string{"--height", "--spacing"}, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"hue-rotate":          {Reading: Reading{Order: []int{335, 339}, Count: 2}, ThemeKeys: []string{"--hue-rotate"}, SupportsNegative: true, BareValue: BareValuePositiveInteger, BareValueSuffix: "deg"},
	"indent":              {Reading: Reading{Order: []int{282}, Count: 1}, ThemeKeys: []string{"--text-indent", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"inset":               {Reading: Reading{Order: []int{4}, Count: 1}, ThemeKeys: []string{"--inset", "--spacing"}, SupportsNegative: true, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"inset-be":            {Reading: Reading{Order: []int{10}, Count: 1}, ThemeKeys: []string{"--inset", "--spacing"}, SupportsNegative: true, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"inset-bs":            {Reading: Reading{Order: []int{9}, Count: 1}, ThemeKeys: []string{"--inset", "--spacing"}, SupportsNegative: true, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"inset-e":             {Reading: Reading{Order: []int{8}, Count: 1}, ThemeKeys: []string{"--inset", "--spacing"}, SupportsNegative: true, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"inset-s":             {Reading: Reading{Order: []int{7}, Count: 1}, ThemeKeys: []string{"--inset", "--spacing"}, SupportsNegative: true, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"inset-x":             {Reading: Reading{Order: []int{5}, Count: 1}, ThemeKeys: []string{"--inset", "--spacing"}, SupportsNegative: true, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"inset-y":             {Reading: Reading{Order: []int{6}, Count: 1}, ThemeKeys: []string{"--inset", "--spacing"}, SupportsNegative: true, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"invert":              {Reading: Reading{Order: []int{336, 339}, Count: 2}, ThemeKeys: []string{"--invert"}, DefaultValue: "100%", DefaultValuePresent: true, BareValue: BareValuePositiveInteger, BareValueSuffix: "%"},
	"leading":             {Reading: Reading{Order: []int{287}, Count: 2}, ThemeKeys: []string{"--leading", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true, LiteralReadings: map[string]Reading{"none": Reading{Order: []int{287}, Count: 2}}},
	"left":                {Reading: Reading{Order: []int{14}, Count: 1}, ThemeKeys: []string{"--inset", "--spacing"}, SupportsNegative: true, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"line-clamp":          {Reading: Reading{Order: []int{39, 143}, Count: 4}, ThemeKeys: []string{"--line-clamp"}, BareValue: BareValuePositiveInteger, LiteralReadings: map[string]Reading{"none": Reading{Order: []int{39, 143}, Count: 4}}},
	"m":                   {Reading: Reading{Order: []int{27}, Count: 1}, ThemeKeys: []string{"--margin", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"mask-conic":          {BareValue: BareValuePositiveInteger, BareValueSuffix: "deg", Reading: Reading{Order: []int{209, 244, 245, 257}, Count: 4}, SupportsNegative: true},
	"mask-linear":         {BareValue: BareValuePositiveInteger, BareValueSuffix: "deg", Reading: Reading{Order: []int{209, 230, 231, 257}, Count: 4}, SupportsNegative: true},
	"mask-radial":         {Reading: Reading{Order: []int{209, 236, 238, 257}, Count: 4}},
	"max-block":           {Reading: Reading{Order: nil, Count: 1}, ThemeKeys: []string{"--spacing"}, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"max-h":               {Reading: Reading{Order: []int{43}, Count: 1}, ThemeKeys: []string{"--max-height", "--height", "--spacing"}, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"max-inline":          {Reading: Reading{Order: nil, Count: 1}, ThemeKeys: []string{"--spacing", "--container"}, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"max-w":               {Reading: Reading{Order: []int{46}, Count: 1}, ThemeKeys: []string{"--max-width", "--spacing", "--container"}, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"max-w-screen":        {Reading: Reading{Order: []int{46}, Count: 1}, ThemeKeys: []string{"--breakpoint"}, RefusesEmptyValue: true, RefusesArbitraryValue: true, LiteralReadings: map[string]Reading{"sm": Reading{Order: []int{46}, Count: 1}}},
	"mb":                  {Reading: Reading{Order: []int{36}, Count: 1}, ThemeKeys: []string{"--margin", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"mbe":                 {Reading: Reading{Order: []int{33}, Count: 1}, ThemeKeys: []string{"--margin", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"mbs":                 {Reading: Reading{Order: []int{32}, Count: 1}, ThemeKeys: []string{"--margin", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"me":                  {Reading: Reading{Order: []int{31}, Count: 1}, ThemeKeys: []string{"--margin", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"min-block":           {Reading: Reading{Order: nil, Count: 1}, ThemeKeys: []string{"--spacing"}, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"min-h":               {Reading: Reading{Order: []int{44}, Count: 1}, ThemeKeys: []string{"--min-height", "--height", "--spacing"}, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"min-inline":          {Reading: Reading{Order: nil, Count: 1}, ThemeKeys: []string{"--spacing", "--container"}, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"min-w":               {Reading: Reading{Order: []int{47}, Count: 1}, ThemeKeys: []string{"--min-width", "--spacing", "--container"}, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"ml":                  {Reading: Reading{Order: []int{37}, Count: 1}, ThemeKeys: []string{"--margin", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"mr":                  {Reading: Reading{Order: []int{35}, Count: 1}, ThemeKeys: []string{"--margin", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"ms":                  {Reading: Reading{Order: []int{30}, Count: 1}, ThemeKeys: []string{"--margin", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"mt":                  {Reading: Reading{Order: []int{34}, Count: 1}, ThemeKeys: []string{"--margin", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"mx":                  {Reading: Reading{Order: []int{28}, Count: 1}, ThemeKeys: []string{"--margin", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"my":                  {Reading: Reading{Order: []int{29}, Count: 1}, ThemeKeys: []string{"--margin", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"p":                   {Reading: Reading{Order: []int{270}, Count: 1}, ThemeKeys: []string{"--padding", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"pb":                  {Reading: Reading{Order: []int{279}, Count: 1}, ThemeKeys: []string{"--padding", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"pbe":                 {Reading: Reading{Order: []int{276}, Count: 1}, ThemeKeys: []string{"--padding", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"pbs":                 {Reading: Reading{Order: []int{275}, Count: 1}, ThemeKeys: []string{"--padding", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"pe":                  {Reading: Reading{Order: []int{274}, Count: 1}, ThemeKeys: []string{"--padding", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"pl":                  {Reading: Reading{Order: []int{280}, Count: 1}, ThemeKeys: []string{"--padding", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"placeholder":         {Reading: Reading{Order: []int{308}, Count: 2}, ThemeKeys: []string{"--placeholder-color", "--color"}, RefusesEmptyValue: true, AcceptsColorKeywords: true},
	"pr":                  {Reading: Reading{Order: []int{278}, Count: 1}, ThemeKeys: []string{"--padding", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"ps":                  {Reading: Reading{Order: []int{273}, Count: 1}, ThemeKeys: []string{"--padding", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"pt":                  {Reading: Reading{Order: []int{277}, Count: 1}, ThemeKeys: []string{"--padding", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"px":                  {Reading: Reading{Order: []int{271}, Count: 1}, ThemeKeys: []string{"--padding", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"py":                  {Reading: Reading{Order: []int{272}, Count: 1}, ThemeKeys: []string{"--padding", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"right":               {Reading: Reading{Order: []int{12}, Count: 1}, ThemeKeys: []string{"--inset", "--spacing"}, SupportsNegative: true, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"rotate-x":            {Reading: Reading{Order: []int{66, 71}, Count: 2}, ThemeKeys: []string{"--rotate"}, SupportsNegative: true, BareValue: BareValuePositiveInteger, BareValueSuffix: "deg", RefusesEmptyValue: true},
	"rotate-y":            {Reading: Reading{Order: []int{67, 71}, Count: 2}, ThemeKeys: []string{"--rotate"}, SupportsNegative: true, BareValue: BareValuePositiveInteger, BareValueSuffix: "deg", RefusesEmptyValue: true},
	"rotate-z":            {Reading: Reading{Order: []int{68, 71}, Count: 2}, ThemeKeys: []string{"--rotate"}, SupportsNegative: true, BareValue: BareValuePositiveInteger, BareValueSuffix: "deg", RefusesEmptyValue: true},
	"rounded":             {Reading: Reading{Order: []int{150}, Count: 1}, ThemeKeys: []string{"--radius"}, LiteralReadings: map[string]Reading{"full": Reading{Order: []int{150}, Count: 1}, "none": Reading{Order: []int{150}, Count: 1}}},
	"rounded-b":           {Reading: Reading{Order: []int{163, 164}, Count: 2}, ThemeKeys: []string{"--radius"}, LiteralReadings: map[string]Reading{"full": Reading{Order: []int{163, 164}, Count: 2}, "none": Reading{Order: []int{163, 164}, Count: 2}}},
	"rounded-bl":          {Reading: Reading{Order: []int{164}, Count: 1}, ThemeKeys: []string{"--radius"}, LiteralReadings: map[string]Reading{"full": Reading{Order: []int{164}, Count: 1}, "none": Reading{Order: []int{164}, Count: 1}}},
	"rounded-br":          {Reading: Reading{Order: []int{163}, Count: 1}, ThemeKeys: []string{"--radius"}, LiteralReadings: map[string]Reading{"full": Reading{Order: []int{163}, Count: 1}, "none": Reading{Order: []int{163}, Count: 1}}},
	"rounded-e":           {Reading: Reading{Order: []int{158, 159}, Count: 2}, ThemeKeys: []string{"--radius"}, LiteralReadings: map[string]Reading{"full": Reading{Order: []int{158, 159}, Count: 2}, "none": Reading{Order: []int{158, 159}, Count: 2}}},
	"rounded-ee":          {Reading: Reading{Order: []int{159}, Count: 1}, ThemeKeys: []string{"--radius"}, LiteralReadings: map[string]Reading{"full": Reading{Order: []int{159}, Count: 1}, "none": Reading{Order: []int{159}, Count: 1}}},
	"rounded-es":          {Reading: Reading{Order: []int{160}, Count: 1}, ThemeKeys: []string{"--radius"}, LiteralReadings: map[string]Reading{"full": Reading{Order: []int{160}, Count: 1}, "none": Reading{Order: []int{160}, Count: 1}}},
	"rounded-l":           {Reading: Reading{Order: []int{161, 164}, Count: 2}, ThemeKeys: []string{"--radius"}, LiteralReadings: map[string]Reading{"full": Reading{Order: []int{161, 164}, Count: 2}, "none": Reading{Order: []int{161, 164}, Count: 2}}},
	"rounded-r":           {Reading: Reading{Order: []int{162, 163}, Count: 2}, ThemeKeys: []string{"--radius"}, LiteralReadings: map[string]Reading{"full": Reading{Order: []int{162, 163}, Count: 2}, "none": Reading{Order: []int{162, 163}, Count: 2}}},
	"rounded-s":           {Reading: Reading{Order: []int{157, 160}, Count: 2}, ThemeKeys: []string{"--radius"}, LiteralReadings: map[string]Reading{"full": Reading{Order: []int{157, 160}, Count: 2}, "none": Reading{Order: []int{157, 160}, Count: 2}}},
	"rounded-se":          {Reading: Reading{Order: []int{158}, Count: 1}, ThemeKeys: []string{"--radius"}, LiteralReadings: map[string]Reading{"full": Reading{Order: []int{158}, Count: 1}, "none": Reading{Order: []int{158}, Count: 1}}},
	"rounded-ss":          {Reading: Reading{Order: []int{157}, Count: 1}, ThemeKeys: []string{"--radius"}, LiteralReadings: map[string]Reading{"full": Reading{Order: []int{157}, Count: 1}, "none": Reading{Order: []int{157}, Count: 1}}},
	"rounded-t":           {Reading: Reading{Order: []int{161, 162}, Count: 2}, ThemeKeys: []string{"--radius"}, LiteralReadings: map[string]Reading{"full": Reading{Order: []int{161, 162}, Count: 2}, "none": Reading{Order: []int{161, 162}, Count: 2}}},
	"rounded-tl":          {Reading: Reading{Order: []int{161}, Count: 1}, ThemeKeys: []string{"--radius"}, LiteralReadings: map[string]Reading{"full": Reading{Order: []int{161}, Count: 1}, "none": Reading{Order: []int{161}, Count: 1}}},
	"rounded-tr":          {Reading: Reading{Order: []int{162}, Count: 1}, ThemeKeys: []string{"--radius"}, LiteralReadings: map[string]Reading{"full": Reading{Order: []int{162}, Count: 1}, "none": Reading{Order: []int{162}, Count: 1}}},
	"row-span":            {Reading: Reading{Order: []int{21}, Count: 1}, BareValue: BareValuePositiveInteger, LiteralReadings: map[string]Reading{"full": Reading{Order: []int{21}, Count: 1}}},
	"saturate":            {Reading: Reading{Order: []int{337, 339}, Count: 2}, ThemeKeys: []string{"--saturate"}, BareValue: BareValuePositiveInteger, BareValueSuffix: "%"},
	"scale-x":             {Reading: Reading{Order: []int{61, 62}, Count: 2}, ThemeKeys: []string{"--scale"}, SupportsNegative: true, BareValue: BareValuePositiveInteger, BareValueSuffix: "%", RefusesEmptyValue: true},
	"scale-y":             {Reading: Reading{Order: []int{61, 63}, Count: 2}, ThemeKeys: []string{"--scale"}, SupportsNegative: true, BareValue: BareValuePositiveInteger, BareValueSuffix: "%", RefusesEmptyValue: true},
	"scale-z":             {Reading: Reading{Order: []int{61, 64}, Count: 2}, ThemeKeys: []string{"--scale"}, SupportsNegative: true, BareValue: BareValuePositiveInteger, BareValueSuffix: "%", RefusesEmptyValue: true},
	"scroll-m":            {Reading: Reading{Order: []int{84}, Count: 1}, ThemeKeys: []string{"--scroll-margin", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"scroll-mb":           {Reading: Reading{Order: []int{93}, Count: 1}, ThemeKeys: []string{"--scroll-margin", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"scroll-mbe":          {Reading: Reading{Order: []int{90}, Count: 1}, ThemeKeys: []string{"--scroll-margin", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"scroll-mbs":          {Reading: Reading{Order: []int{89}, Count: 1}, ThemeKeys: []string{"--scroll-margin", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"scroll-me":           {Reading: Reading{Order: []int{88}, Count: 1}, ThemeKeys: []string{"--scroll-margin", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"scroll-ml":           {Reading: Reading{Order: []int{94}, Count: 1}, ThemeKeys: []string{"--scroll-margin", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"scroll-mr":           {Reading: Reading{Order: []int{92}, Count: 1}, ThemeKeys: []string{"--scroll-margin", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"scroll-ms":           {Reading: Reading{Order: []int{87}, Count: 1}, ThemeKeys: []string{"--scroll-margin", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"scroll-mt":           {Reading: Reading{Order: []int{91}, Count: 1}, ThemeKeys: []string{"--scroll-margin", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"scroll-mx":           {Reading: Reading{Order: []int{85}, Count: 1}, ThemeKeys: []string{"--scroll-margin", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"scroll-my":           {Reading: Reading{Order: []int{86}, Count: 1}, ThemeKeys: []string{"--scroll-margin", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"scroll-p":            {Reading: Reading{Order: []int{95}, Count: 1}, ThemeKeys: []string{"--scroll-padding", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"scroll-pb":           {Reading: Reading{Order: []int{104}, Count: 1}, ThemeKeys: []string{"--scroll-padding", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"scroll-pbe":          {Reading: Reading{Order: []int{101}, Count: 1}, ThemeKeys: []string{"--scroll-padding", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"scroll-pbs":          {Reading: Reading{Order: []int{100}, Count: 1}, ThemeKeys: []string{"--scroll-padding", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"scroll-pe":           {Reading: Reading{Order: []int{99}, Count: 1}, ThemeKeys: []string{"--scroll-padding", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"scroll-pl":           {Reading: Reading{Order: []int{105}, Count: 1}, ThemeKeys: []string{"--scroll-padding", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"scroll-pr":           {Reading: Reading{Order: []int{103}, Count: 1}, ThemeKeys: []string{"--scroll-padding", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"scroll-ps":           {Reading: Reading{Order: []int{98}, Count: 1}, ThemeKeys: []string{"--scroll-padding", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"scroll-pt":           {Reading: Reading{Order: []int{102}, Count: 1}, ThemeKeys: []string{"--scroll-padding", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"scroll-px":           {Reading: Reading{Order: []int{96}, Count: 1}, ThemeKeys: []string{"--scroll-padding", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"scroll-py":           {Reading: Reading{Order: []int{97}, Count: 1}, ThemeKeys: []string{"--scroll-padding", "--spacing"}, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"scrollbar-thumb":     {Reading: Reading{Order: []int{107}, Count: 2}, ThemeKeys: []string{"--color"}, RefusesEmptyValue: true, AcceptsColorKeywords: true},
	"scrollbar-track":     {Reading: Reading{Order: []int{107}, Count: 2}, ThemeKeys: []string{"--color"}, RefusesEmptyValue: true, AcceptsColorKeywords: true},
	"sepia":               {Reading: Reading{Order: []int{338, 339}, Count: 2}, ThemeKeys: []string{"--sepia"}, DefaultValue: "100%", DefaultValuePresent: true, BareValue: BareValuePositiveInteger, BareValueSuffix: "%"},
	"size":                {Reading: Reading{Order: []int{42, 45}, Count: 3}, ThemeKeys: []string{"--size", "--spacing"}, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"skew":                {Reading: Reading{Order: []int{69, 70, 71}, Count: 3}, ThemeKeys: []string{"--skew"}, SupportsNegative: true, BareValue: BareValuePositiveInteger, BareValueSuffix: "deg"},
	"skew-x":              {Reading: Reading{Order: []int{69, 71}, Count: 2}, ThemeKeys: []string{"--skew"}, SupportsNegative: true, BareValue: BareValuePositiveInteger, BareValueSuffix: "deg"},
	"skew-y":              {Reading: Reading{Order: []int{70, 71}, Count: 2}, ThemeKeys: []string{"--skew"}, SupportsNegative: true, BareValue: BareValuePositiveInteger, BareValueSuffix: "deg"},
	"space-x":             {Reading: Reading{Order: []int{132}, Count: 4}, ThemeKeys: []string{"--space", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"space-y":             {Reading: Reading{Order: []int{131}, Count: 4}, ThemeKeys: []string{"--space", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"start":               {Reading: Reading{Order: []int{7}, Count: 1}, ThemeKeys: []string{"--inset", "--spacing"}, SupportsNegative: true, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"top":                 {Reading: Reading{Order: []int{11}, Count: 1}, ThemeKeys: []string{"--inset", "--spacing"}, SupportsNegative: true, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"tracking":            {Reading: Reading{Order: []int{289}, Count: 2}, ThemeKeys: []string{"--tracking"}, SupportsNegative: true, LiteralReadings: map[string]Reading{"normal": Reading{Order: []int{289}, Count: 2}, "tight": Reading{Order: []int{289}, Count: 2}, "tighter": Reading{Order: []int{289}, Count: 2}, "wide": Reading{Order: []int{289}, Count: 2}, "wider": Reading{Order: []int{289}, Count: 2}, "widest": Reading{Order: []int{289}, Count: 2}}},
	"transition":          {Reading: Reading{Order: []int{350, 353, 354}, Count: 3}, ThemeKeys: []string{"--transition-property"}, DefaultValue: "color, background-color, border-color, outline-color, text-decoration-color, fill, stroke, --tw-gradient-from, --tw-gradient-via, --tw-gradient-to, opacity, box-shadow, transform, translate, scale, rotate, filter, -webkit-backdrop-filter, backdrop-filter, display, content-visibility, overlay, pointer-events", DefaultValuePresent: true, LiteralReadings: map[string]Reading{"all": Reading{Order: []int{350, 353, 354}, Count: 3}, "colors": Reading{Order: []int{350, 353, 354}, Count: 3}, "none": Reading{Order: []int{350}, Count: 1}, "opacity": Reading{Order: []int{350, 353, 354}, Count: 3}, "shadow": Reading{Order: []int{350, 353, 354}, Count: 3}, "transform": Reading{Order: []int{350, 353, 354}, Count: 3}}},
	"translate":           {Reading: Reading{Order: []int{57, 58, 59}, Count: 3}, ThemeKeys: []string{"--translate", "--spacing"}, SupportsNegative: true, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, LiteralReadings: map[string]Reading{"none": Reading{Order: []int{57}, Count: 1}}, RefusesEmptyValue: true},
	"translate-x":         {Reading: Reading{Order: []int{57, 58}, Count: 2}, ThemeKeys: []string{"--translate", "--spacing"}, SupportsNegative: true, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"translate-y":         {Reading: Reading{Order: []int{57, 59}, Count: 2}, ThemeKeys: []string{"--translate", "--spacing"}, SupportsNegative: true, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"translate-z":         {Reading: Reading{Order: []int{57, 60}, Count: 2}, ThemeKeys: []string{"--translate", "--spacing"}, SupportsNegative: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
	"underline-offset":    {Reading: Reading{Order: []int{306}, Count: 1}, ThemeKeys: []string{"--text-underline-offset"}, SupportsNegative: true, BareValue: BareValuePositiveInteger, BareValueSuffix: "px", LiteralReadings: map[string]Reading{"auto": Reading{Order: []int{306}, Count: 1}}},
	"w":                   {Reading: Reading{Order: []int{45}, Count: 1}, ThemeKeys: []string{"--width", "--spacing", "--container"}, SupportsFractions: true, BareValue: BareValueSpacingMultiplier, RefusesEmptyValue: true},
}

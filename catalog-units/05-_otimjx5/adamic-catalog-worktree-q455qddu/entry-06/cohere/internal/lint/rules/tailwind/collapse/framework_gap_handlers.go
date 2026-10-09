// The ported `handle` bodies of the framework roots whose reading partitions on the value.
//
// The third and last slice of the emitter port. frameworkhandlers.go covers the 33 roots that emit
// one declaration and frameworkmultihandlers.go the 152 that emit a fixed list; both answer a root
// with a single declaration list, because for those roots there is only one.
//
// These 57 are the roots where that is false. `border-[3px]` is a width and `border-red-500` is a
// colour, `text-[3px]` is a font size and `text-red-500` is a colour, and the two emit different
// properties from the same registration. So an emitter here is a function of the value rather than a
// list stored under a root, which is the shape upstream's own handle bodies have: every one of these
// opens with a `switch` on an inferred type, or a fallthrough between a colour lookup and a width
// lookup, and returns a different array per arm.
//
// Same contract as the other two files otherwise. Each body was read at the v4.3.3 tag, the `decl`
// shape is ported and the value expression is not, and the reading is `PropertySort` over what an
// emitter produced rather than a number stored beside it.
//
// # What a branch needs, and why it is not `ResolvedUtilityValue`
//
// The other two slices take a `ResolvedUtilityValue`, because a root with one declaration list does
// not read it. These do, and measured against the descriptor rows they read four things, not one:
//
//   - whether the class carried a value at all. `border` is a width and `border-red-500` is a
//     colour, and upstream's first line is `if (!candidate.value)`.
//   - whether that value was arbitrary or named. `scale-[2]` emits one declaration and `scale-150`
//     emits four, and no inferred type separates them: upstream's arbitrary arm returns early.
//   - the inferred data type, against the root's own ordered type list. This is the branch the
//     milestone named, and it is the one `InferDataType` already answers.
//   - whether a modifier was written. `text-[3px]` emits `font-size` alone and `text-[3px]/50`
//     emits `font-size` and `line-height`; `shadow-md` writes `--tw-shadow-alpha` with an absent
//     value and `shadow-md/50` writes it with a real one, and `PropertySort` skips an absent value
//     and counts a present one. That is the whole of the `#2` against `#3` difference on the
//     shadow family's own axis, and it is a fact about the candidate rather than about the value.
//
// `ResolvedUtilityValue` carries the first three between `Value`, `DataType` and `IsStaticValue`,
// and carries nothing at all about the fourth: resolution consumes the modifier and does not report
// it. So the emitter takes a `UtilityBranch` instead, which names all four, and
// `UtilityBranchFor` builds one from the pieces a caller already has. Passing the resolved value
// alone would make the shadow family unportable rather than merely awkward, and would do it
// silently, since every arm would still return a well-formed list.
//
// # Where the descriptor rows say something upstream's source does not
//
// A descriptor row is keyed on a theme namespace as well as on a type, and the two are not the same
// question. `border-red-500` resolves through `--color` and `border-4` through `--border-width`, and
// neither is an inferred type: `red-500` infers as nothing and `4` infers as an integer on a root
// whose type list does not branch on integers.
//
// Upstream that distinction is a fallthrough rather than a switch. The colour lookup runs first and
// returns if it hit, the width lookup runs second. So a named value is a colour when the theme says
// so and a width otherwise, and the emitter takes that as an input (`ResolvedAsColor`) rather than
// re-deriving it, because deriving it would mean carrying every repository's `--color` keys into a
// file that describes Tailwind. That is the same seam descriptor_base_table.go draws, for the same
// reason.
package tailwind

// UtilityBranch is what one of these handle bodies reads before it picks an arm.
//
// Four fields rather than a resolved value, for the reason in the package comment: the shadow family
// branches on the modifier, which resolution consumes and does not report.
//
// The zero value is a class written with no value at all, which is a real state here and the one
// several of these roots answer differently from every other: `border` is a width, `border-red-500`
// is a colour.
type UtilityBranch struct {
	// HasValue is false for a class written as the bare root, such as `border` or `ring`.
	HasValue bool
	// IsArbitrary is true for a bracketed value, such as `border-[3px]`.
	//
	// Separate from DataType because it decides an arm on its own for three roots. `scale-[2]`
	// returns from upstream's arbitrary arm with one declaration while `scale-150` falls through to
	// four, and both infer the same type.
	IsArbitrary bool
	// DataType is the value's inferred type, or the empty string when it infers as nothing.
	//
	// Callers get this from InferDataType against the root's own ordered type list, or from the
	// author's own annotation on a value written as `bg-[color:var(--x)]`.
	DataType DataType
	// ResolvedAsColor is whether a named value resolved through a colour namespace.
	//
	// An input rather than something derived here. Upstream this is `resolveThemeColor` returning a
	// value, which consults the repository's own `--color` keys; deriving it in this file would put
	// one repository's tokens in a table describing Tailwind.
	ResolvedAsColor bool
	// HasModifier is whether the class carried a `/...` modifier.
	//
	// The shadow family writes an alpha declaration whose value is absent without one and present
	// with one, and PropertySort skips the first and counts the second. `text` uses it differently
	// again: a modifier turns a font size into a font size plus a line height.
	HasModifier bool
	// ResolvedNamespace is the theme namespace a named value resolved through, when one that is not
	// a colour decides an arm.
	//
	// `ResolvedAsColor` covers the colour namespaces, which is the split most of these roots make.
	// One root splits on a namespace that is not a colour: `font` consults `--font` before
	// `--font-weight`, so `font-mono` emits `font-family` and `font-medium` emits `font-weight`,
	// and both are named values that infer as nothing. That is the branch `ClassDeclaredProperties`
	// recorded as three per-class overrides.
	//
	// An input for the same reason `ResolvedAsColor` is: which keys are in `--font` is this
	// repository's `@theme`, not Tailwind's.
	ResolvedNamespace string
	// Value is the resolved value the class carried, when a caller needs the declarations to differ
	// between two values of one root.
	//
	// Empty for every caller that reads property names, which is all of them but one: the branch
	// decides which arm runs and the arm decides the properties, and neither consults this. It
	// exists for `ComposesFor`, which emits a root twice and asks whether the declarations an author
	// can see came out the same, and that question is meaningless if both emissions carry one
	// sentinel.
	//
	// A root's own `--tw-*` variables take this. The shared property its family layers through does
	// not: `box-shadow` carries a constant naming every shadow variable whatever the value was, and
	// that constant is what makes two shadows stack rather than overwrite. Emitting the value there
	// instead is exactly the flattening that made composition unanswerable from these bodies.
	Value string
}

// valueOr returns the branch's value, falling back to the sentinel every property-name caller wants.
//
// One place rather than a check at each `Declaration` call, so a new arm gets the behaviour by
// construction instead of by remembering.
func (branch UtilityBranch) valueOr() string {
	if branch.Value != "" {
		return branch.Value
	}
	return gapEmitterValue
}

// UtilityBranchFor builds a branch from a parsed candidate and the root's own type list.
//
// The colour question is passed in rather than asked here, for the reason on the field. Everything
// else is read off the candidate the same way upstream's handle body reads it: the presence of a
// value, its kind, and the type inferred against the root's list in the root's own order, which
// InferDataType already implements and which the descriptor rows already carry per root.
func UtilityBranchFor(candidate *ParsedCandidate, typeList []DataType, resolvedAsColor bool, resolvedNamespace string) UtilityBranch {
	branch := UtilityBranch{ResolvedAsColor: resolvedAsColor, ResolvedNamespace: resolvedNamespace}
	if candidate == nil {
		return branch
	}
	branch.HasModifier = candidate.Modifier != nil
	if candidate.Value == nil {
		return branch
	}
	branch.HasValue = true
	if candidate.Value.Kind == ParsedValueKindArbitrary {
		branch.IsArbitrary = true
		// An author's own annotation wins over inference, matching Lookup: the engine trusts
		// `bg-[color:var(--x)]` rather than inferring from a value it has already been told about.
		if candidate.Value.DataType != "" {
			branch.DataType = DataType(candidate.Value.DataType)
			return branch
		}
	}
	branch.DataType = InferDataType(candidate.Value.Value, typeList)
	return branch
}

// isOneOf reports whether the branch's type is any of the given ones.
//
// Upstream these arms are `switch` cases that fall through to one body, as in `case 'length': case
// 'line-width':`. Written as a predicate rather than a Go switch so an arm reads as the set of types
// it accepts, which is how the source spells it.
func (branch UtilityBranch) isOneOf(types ...DataType) bool {
	for _, dataType := range types {
		if branch.DataType == dataType {
			return true
		}
	}
	return false
}

// GapEmitter is one of these roots' handle bodies.
//
// Returns nil where upstream returns without producing declarations, which is a class that compiles
// to nothing rather than one that compiles to an empty rule. Every consumer here distinguishes the
// two: a class with no rules has no reading at all.
type GapEmitter func(branch UtilityBranch) []*Node

// declarations is the common tail of these bodies: a fixed property list, all carrying the value.
//
// The value is a sentinel for the same reason it is in the other two slices. `PropertySort` reads a
// property name, counts declarations, and reads a value only for `--tw-sort`, which none of these
// emit.
func declarations(properties ...string) []*Node {
	nodes := make([]*Node, 0, len(properties))
	for _, property := range properties {
		nodes = append(nodes, Declaration(property, gapEmitterValue))
	}
	return nodes
}

// gapEmitterValue is what every ported declaration here carries.
//
// Named rather than spelled at 200 call sites so that the one place a value is meaningful, the
// absent-valued alpha declaration below, is visibly a different thing rather than a different
// string.
const gapEmitterValue = "zzsentinel"

// absentValued is a declaration upstream emits with an undefined value.
//
// `decl('--tw-shadow-alpha', alpha)` where `alpha` is undefined is not a declaration that generates
// `--tw-shadow-alpha:;`. Upstream's own `decl` stores the undefined, and `PropertySort` skips a
// declaration whose value is absent while counting one whose value is empty, with a comment saying
// `--tw-foo:;` is valid CSS.
//
// That is the whole of the shadow family's `#2` against `#3`: without a modifier the alpha is
// undefined and does not count, with one it is a percentage and does. Emitting an empty string here
// instead would count it either way and quietly move four roots' unmodified readings.
func absentValued(property string) *Node {
	return &Node{Kind: KindDeclaration, Property: property}
}

// The border family. Eleven roots registered through one `borderSideUtility`, differing only in the
// three property names each passes it.
//
// The branch is upstream's, not a restatement of it: a bare `border` is a width, an arbitrary value
// is a width when it infers as `line-width` or `length` and a colour otherwise, and a named value is
// a colour when the theme says so and a width otherwise. The width arm emits the style declaration
// and the width declaration; the colour arm emits one.
func borderSideEmitter(styleProperty, widthProperty, colorProperty string) GapEmitter {
	return func(branch UtilityBranch) []*Node {
		if !branch.HasValue {
			return declarations(styleProperty, widthProperty)
		}
		if branch.IsArbitrary {
			if branch.isOneOf(DataTypeLineWidth, DataTypeLength) {
				return declarations(styleProperty, widthProperty)
			}
			return declarations(colorProperty)
		}
		if branch.ResolvedAsColor {
			return declarations(colorProperty)
		}
		return declarations(styleProperty, widthProperty)
	}
}

// The mask edge stops. Twelve roots registered through one `maskEdgeUtility`, differing only in
// which of the four edges they write.
//
// Each emits three common declarations and then, per active edge, the edge's own gradient variable
// and one stop variable that is a colour on the colour arm and a position on the position arm. The
// edge loop is upstream's fixed `['top', 'right', 'bottom', 'left']` rather than the order the
// caller names them, which is what puts `--tw-mask-right` before `--tw-mask-left` on `mask-x-*`.
//
// The order does not reach a reading, since PropertySort sorts positions, and it is kept anyway so
// this can be checked against the source line by line.
func maskEdgeEmitter(stop string, edges ...string) GapEmitter {
	return func(branch UtilityBranch) []*Node {
		suffix := "position"
		if branch.maskStopIsColor() {
			suffix = "color"
		}
		properties := []string{"mask-image", "mask-composite", "--tw-mask-linear"}
		for _, edge := range []string{"top", "right", "bottom", "left"} {
			if !containsString(edges, edge) {
				continue
			}
			properties = append(properties,
				"--tw-mask-"+edge,
				"--tw-mask-"+edge+"-"+stop+"-"+suffix,
			)
		}
		return declarations(properties...)
	}
}

// The mask linear, radial and conic stops. Six roots, each emitting four common declarations and one
// stop declaration that is a colour or a position.
func maskGradientEmitter(gradient, stop string) GapEmitter {
	return func(branch UtilityBranch) []*Node {
		suffix := "position"
		if branch.maskStopIsColor() {
			suffix = "color"
		}
		return declarations(
			"mask-image",
			"mask-composite",
			"--tw-mask-"+gradient+"-stops",
			"--tw-mask-"+gradient,
			"--tw-mask-"+gradient+"-"+stop+"-"+suffix,
		)
	}
}

// maskStopIsColor is the shared dispatcher in `maskStopUtility`, which all 18 stop roots run.
//
// An arbitrary value is a colour when it infers as one, against the list `['length', 'percentage',
// 'color']`; a named value is a colour when the theme's colour namespaces answered it, and a
// position otherwise. Both halves are read off the branch rather than recomputed, for the reason in
// the package comment.
func (branch UtilityBranch) maskStopIsColor() bool {
	if branch.IsArbitrary {
		return branch.DataType == DataTypeColor
	}
	return branch.ResolvedAsColor
}

// containsString is a membership test over a small fixed slice.
//
// Written out rather than reaching for `slices.Contains` to keep this file's imports empty, which is
// what the other two slices do as well.
func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

// The shadow family's alpha-carrying bodies.
//
// `shadow`, `inset-shadow`, `drop-shadow` and `text-shadow` each open by reading the modifier into an
// `alpha` that stays undefined without one, then emit `decl('<prefix>-alpha', alpha)` on the size
// arm. That declaration counts only when a modifier was written, which is the whole of the `#2`
// against `#3` split on their measured rows.
//
// The colour arm does not emit the alpha declaration at all, and reads the same modified or not.
func shadowFamilyEmitter(alphaProperty, colorProperty string, sizeProperties ...string) GapEmitter {
	return func(branch UtilityBranch) []*Node {
		if branch.HasValue && branch.shadowIsColor() {
			return declarations(colorProperty)
		}
		nodes := []*Node{absentValued(alphaProperty)}
		if branch.HasModifier {
			nodes[0] = Declaration(alphaProperty, gapEmitterValue)
		}
		return append(nodes, declarations(sizeProperties...)...)
	}
}

// shadowIsColor is the shadow family's own type question.
//
// The arbitrary arm infers against the single-element list `['color']`, so anything that is not a
// colour falls to the size arm. A named value is a colour when the theme's colour namespaces
// answered it.
func (branch UtilityBranch) shadowIsColor() bool {
	if branch.IsArbitrary {
		return branch.DataType == DataTypeColor
	}
	return branch.ResolvedAsColor
}

// gapEmitters is the ported body of each root, keyed the way the descriptor rows are.
//
// Grouped by family rather than sorted, because these bodies are shared: eleven border roots run one
// function and twelve mask edge roots run another, and the grouping is what makes that visible.
var gapEmitters = map[string]GapEmitter{
	// The eleven `borderSideUtility` registrations, in the source's own order.
	"border":    borderSideEmitter("border-style", "border-width", "border-color"),
	"border-x":  borderSideEmitter("border-inline-style", "border-inline-width", "border-inline-color"),
	"border-y":  borderSideEmitter("border-block-style", "border-block-width", "border-block-color"),
	"border-s":  borderSideEmitter("border-inline-start-style", "border-inline-start-width", "border-inline-start-color"),
	"border-e":  borderSideEmitter("border-inline-end-style", "border-inline-end-width", "border-inline-end-color"),
	"border-bs": borderSideEmitter("border-block-start-style", "border-block-start-width", "border-block-start-color"),
	"border-be": borderSideEmitter("border-block-end-style", "border-block-end-width", "border-block-end-color"),
	"border-t":  borderSideEmitter("border-top-style", "border-top-width", "border-top-color"),
	"border-r":  borderSideEmitter("border-right-style", "border-right-width", "border-right-color"),
	"border-b":  borderSideEmitter("border-bottom-style", "border-bottom-width", "border-bottom-color"),
	"border-l":  borderSideEmitter("border-left-style", "border-left-width", "border-left-color"),

	// The twelve `maskEdgeUtility` registrations, with the edge sets upstream passes them.
	"mask-x-from": maskEdgeEmitter("from", "right", "left"),
	"mask-x-to":   maskEdgeEmitter("to", "right", "left"),
	"mask-y-from": maskEdgeEmitter("from", "top", "bottom"),
	"mask-y-to":   maskEdgeEmitter("to", "top", "bottom"),
	"mask-t-from": maskEdgeEmitter("from", "top"),
	"mask-t-to":   maskEdgeEmitter("to", "top"),
	"mask-r-from": maskEdgeEmitter("from", "right"),
	"mask-r-to":   maskEdgeEmitter("to", "right"),
	"mask-b-from": maskEdgeEmitter("from", "bottom"),
	"mask-b-to":   maskEdgeEmitter("to", "bottom"),
	"mask-l-from": maskEdgeEmitter("from", "left"),
	"mask-l-to":   maskEdgeEmitter("to", "left"),

	// The six linear, radial and conic stops.
	"mask-linear-from": maskGradientEmitter("linear", "from"),
	"mask-linear-to":   maskGradientEmitter("linear", "to"),
	"mask-radial-from": maskGradientEmitter("radial", "from"),
	"mask-radial-to":   maskGradientEmitter("radial", "to"),
	"mask-conic-from":  maskGradientEmitter("conic", "from"),
	"mask-conic-to":    maskGradientEmitter("conic", "to"),

	// The shadow family. The size arm's property lists are upstream's, including the `box-shadow`
	// three of them end with and the two `--tw-drop-shadow*` declarations drop-shadow emits.
	"shadow":       shadowFamilyEmitter("--tw-shadow-alpha", "--tw-shadow-color", "--tw-shadow", "box-shadow"),
	"inset-shadow": shadowFamilyEmitter("--tw-inset-shadow-alpha", "--tw-inset-shadow-color", "--tw-inset-shadow", "box-shadow"),
	"text-shadow":  shadowFamilyEmitter("--tw-text-shadow-alpha", "--tw-text-shadow-color", "text-shadow"),

	// `drop-shadow` is the shadow family's odd member and does not share their emitter. Its colour
	// arm emits two declarations rather than one, because it writes `--tw-drop-shadow` as well as
	// the colour, and its size arm emits the size variable, the shared variable and `filter`.
	"drop-shadow": func(branch UtilityBranch) []*Node {
		if branch.HasValue && branch.shadowIsColor() {
			return declarations("--tw-drop-shadow-color", "--tw-drop-shadow")
		}
		nodes := []*Node{absentValued("--tw-drop-shadow-alpha")}
		if branch.HasModifier {
			nodes[0] = Declaration("--tw-drop-shadow-alpha", gapEmitterValue)
		}
		return append(nodes, declarations("--tw-drop-shadow-size", "--tw-drop-shadow", "filter")...)
	},

	// `bg` is a five-way switch on the inferred type, and the only root here with that many arms.
	// Every arm emits one declaration; which one is the whole content of the branch.
	"bg": func(branch UtilityBranch) []*Node {
		if !branch.HasValue {
			return nil
		}
		if branch.IsArbitrary {
			switch {
			case branch.isOneOf(DataTypePercentage, DataTypePosition):
				return declarations("background-position")
			case branch.isOneOf(DataTypeBackgroundSize, DataTypeLength):
				return declarations("background-size")
			case branch.isOneOf(DataTypeImage, DataTypeURL):
				return declarations("background-image")
			default:
				return declarations("background-color")
			}
		}
		if branch.ResolvedAsColor {
			return declarations("background-color")
		}
		return declarations("background-image")
	},

	// `mask` takes arbitrary values only. Upstream returns before the switch for a named one, which
	// is why its descriptor row carries no namespace bucket at all.
	"mask": func(branch UtilityBranch) []*Node {
		if !branch.HasValue || !branch.IsArbitrary {
			return nil
		}
		switch {
		case branch.isOneOf(DataTypePercentage, DataTypePosition):
			return declarations("mask-position")
		case branch.isOneOf(DataTypeBackgroundSize, DataTypeLength):
			return declarations("mask-size")
		default:
			return declarations("mask-image")
		}
	},

	// `text` is the root whose branch needs the modifier. A size with a modifier emits the line
	// height as well, which is the whole of its Absent-against-Alpha split.
	"text": func(branch UtilityBranch) []*Node {
		if !branch.HasValue {
			return nil
		}
		isSize := branch.isOneOf(DataTypeLength, DataTypePercentage, DataTypeAbsoluteSize, DataTypeRelativeSize)
		if branch.IsArbitrary {
			if !isSize {
				return declarations("color")
			}
			if branch.HasModifier {
				return declarations("font-size", "line-height")
			}
			return declarations("font-size")
		}
		if branch.ResolvedAsColor {
			return declarations("color")
		}
		// A `--text` key resolves with its own line height alongside, which is why the named size
		// path reads two where the arbitrary one reads one.
		return declarations("font-size", "line-height")
	},

	// `font` splits three ways rather than two: a family name emits `font-family` alone, a `--font`
	// key emits it with the two settings declarations, and everything else is a weight.
	"font": func(branch UtilityBranch) []*Node {
		if !branch.HasValue || branch.HasModifier {
			return nil
		}
		if branch.IsArbitrary {
			if branch.isOneOf(DataTypeGenericName, DataTypeFamilyName) {
				return declarations("font-family")
			}
			return declarations("--tw-font-weight", "font-weight")
		}
		// `--font` is consulted before `--font-weight`, and its arm resolves the family together
		// with two settings values from the same theme entry. So `font-mono` declares three
		// properties and `font-medium` declares one, from one root and two named values that both
		// infer as nothing.
		if branch.ResolvedNamespace == "--font" {
			// The two settings declarations carry `options['--font-feature-settings']`, which is
			// undefined unless the theme entry defined it alongside the family. Upstream emits them
			// regardless and an undefined value makes the declaration absent, which PropertySort
			// skips and a conflict comparison must not see. Measured on ahra: neither `--font-mono`
			// nor `--font-sans` carries either option, so both declarations are absent there.
			//
			// Emitted rather than elided so the shape matches the source, with the presence of the
			// value carrying the difference the way it does for the shadow family's alpha.
			return []*Node{
				Declaration("font-family", gapEmitterValue),
				absentValued("font-feature-settings"),
				absentValued("font-variation-settings"),
			}
		}
		return declarations("--tw-font-weight", "font-weight")
	},

	// `decoration` puts the thickness lookup before the colour one for a named value, which is the
	// reverse of the border family's order and the reason it cannot share their emitter.
	"decoration": func(branch UtilityBranch) []*Node {
		if !branch.HasValue {
			return nil
		}
		if branch.IsArbitrary {
			if branch.isOneOf(DataTypeLength, DataTypePercentage) {
				return declarations("text-decoration-thickness")
			}
			return declarations("text-decoration-color")
		}
		if branch.ResolvedAsColor {
			return declarations("text-decoration-color")
		}
		return declarations("text-decoration-thickness")
	},

	// `outline` emits two on its width arm, the style and the width, and one on its colour arm.
	"outline": func(branch UtilityBranch) []*Node {
		if !branch.HasValue {
			return declarations("outline-style", "outline-width")
		}
		if branch.IsArbitrary {
			if branch.isOneOf(DataTypeLength, DataTypeNumber, DataTypePercentage) {
				return declarations("outline-style", "outline-width")
			}
			return declarations("outline-color")
		}
		if branch.ResolvedAsColor {
			return declarations("outline-color")
		}
		return declarations("outline-style", "outline-width")
	},

	// `stroke` splits width against colour like the border family, with its own type set and a
	// single declaration on each arm.
	"stroke": func(branch UtilityBranch) []*Node {
		if !branch.HasValue {
			return nil
		}
		if branch.IsArbitrary {
			if branch.isOneOf(DataTypeNumber, DataTypeLength, DataTypePercentage) {
				return declarations("stroke-width")
			}
			return declarations("stroke")
		}
		if branch.ResolvedAsColor {
			return declarations("stroke")
		}
		return declarations("stroke-width")
	},

	// `fill` has one arm. It is here rather than in the single-declaration slice because it is a
	// colour utility whose descriptor row sits with these, not because it branches.
	"fill": func(branch UtilityBranch) []*Node {
		if !branch.HasValue {
			return nil
		}
		return declarations("fill")
	},

	// `ring` and `inset-ring` emit two on the width arm and one on the colour arm, and the width
	// arm's two are the shared `box-shadow` plus the root's own shadow variable.
	"ring":       ringEmitter("--tw-ring-shadow", "--tw-ring-color"),
	"inset-ring": ringEmitter("--tw-inset-ring-shadow", "--tw-inset-ring-color"),

	// `scale` is the root where arbitrary against named decides the arm on its own. An arbitrary
	// value returns from the arbitrary arm with one declaration; a named one falls through to the
	// three axis variables and the shorthand.
	"scale":  scaleEmitter,
	"-scale": scaleEmitter,

	// `rotate` emits one declaration on every arm.
	"rotate":  rotateEmitter,
	"-rotate": rotateEmitter,

	// The two gradient roots. Both emit the same pair whatever the value, including bare.
	"bg-conic":  gradientPositionEmitter,
	"-bg-conic": gradientPositionEmitter,
	"bg-radial": func(branch UtilityBranch) []*Node {
		// `bg-radial` takes no named value: upstream falls off the end of the handle for one, which
		// is why its descriptor row carries no `@none` bucket where `bg-conic`'s does.
		if branch.HasValue && !branch.IsArbitrary {
			return nil
		}
		return declarations("--tw-gradient-position", "background-image")
	},

	// `filter`, `transform` and `backdrop-filter` take a bare or arbitrary value and emit a fixed
	// list. `backdrop-filter` emits the `-webkit-` prefixed declaration as well, which
	// `PropertyOrder` does not know, so it counts and contributes no position: two declarations at
	// one position, which is what its row says.
	"filter": func(branch UtilityBranch) []*Node {
		return declarations("filter")
	},
	"transform": func(branch UtilityBranch) []*Node {
		if branch.HasValue && !branch.IsArbitrary {
			return nil
		}
		return declarations("transform")
	},
	"backdrop-filter": func(branch UtilityBranch) []*Node {
		return declarations("-webkit-backdrop-filter", "backdrop-filter")
	},

	// `duration` emits its own variable and the real property. `--tw-duration` is outside
	// PropertyOrder, so the two share one position.
	"duration": func(branch UtilityBranch) []*Node {
		if !branch.HasValue || branch.HasModifier {
			return nil
		}
		return declarations("--tw-duration", "transition-duration")
	},

	// `flex`, `block` and `inline` are three roots where a static utility and a functional one share
	// a name. The functional `flex` emits `flex`; `block` and `inline` have no functional
	// registration at all, so a value form emits nothing and only their static form reads. See
	// TestBlockAndInlineHaveNoFunctionalRegistration.
	"flex": func(branch UtilityBranch) []*Node {
		if !branch.HasValue {
			return nil
		}
		return declarations("flex")
	},
	"block":  func(branch UtilityBranch) []*Node { return nil },
	"inline": func(branch UtilityBranch) []*Node { return nil },
}

// ringEmitter is the shared body of `ring` and `inset-ring`.
func ringEmitter(shadowProperty, colorProperty string) GapEmitter {
	return func(branch UtilityBranch) []*Node {
		if !branch.HasValue {
			return declarations(shadowProperty, "box-shadow")
		}
		if branch.IsArbitrary {
			if branch.DataType == DataTypeLength {
				return declarations(shadowProperty, "box-shadow")
			}
			return declarations(colorProperty)
		}
		if branch.ResolvedAsColor {
			return declarations(colorProperty)
		}
		return declarations(shadowProperty, "box-shadow")
	}
}

// scaleEmitter is the shared body of `scale` and `-scale`.
//
// The two registrations differ only in the negation they apply to the value, which is not a property
// name and so is not ported, matching how the other two slices treat a negative root.
func scaleEmitter(branch UtilityBranch) []*Node {
	if !branch.HasValue || branch.HasModifier {
		return nil
	}
	if branch.IsArbitrary {
		return declarations("scale")
	}
	return declarations("--tw-scale-x", "--tw-scale-y", "--tw-scale-z", "scale")
}

// rotateEmitter is the shared body of `rotate` and `-rotate`.
func rotateEmitter(branch UtilityBranch) []*Node {
	if !branch.HasValue || branch.HasModifier {
		return nil
	}
	return declarations("rotate")
}

// gradientPositionEmitter is the shared body of `bg-conic` and `-bg-conic`.
func gradientPositionEmitter(branch UtilityBranch) []*Node {
	return declarations("--tw-gradient-position", "background-image")
}

// EmitGapRoot returns the declarations one of these roots emits for a branch.
//
// Returns nil for a root with no ported emitter, matching the other two slices: a fabricated
// declaration would let the acceptance test compare a table against itself.
func EmitGapRoot(root string, branch UtilityBranch) []*Node {
	emitter, ported := gapEmitters[root]
	if !ported {
		return nil
	}
	nodes := emitter(branch)
	if branch.Value == "" {
		return nodes
	}
	return withResolvedValue(root, nodes, branch.Value)
}

// withResolvedValue restates a gap root's declarations for one concrete value.
//
// The 58 `declarations` call sites in this file all carry the sentinel, because every consumer but
// one reads property names. Rewriting here rather than threading a value through all 58 keeps one
// seam: an arm added later gets this by construction, and the arms stay readable against upstream's
// source, which is the property this file was built for.
//
// Two rules, and the split is the whole of composition. A `--tw-*` custom property is the root's own
// contribution and takes the value, so `shadow-sm` and `shadow-lg` write different `--tw-shadow`.
// A shared property its family layers through takes that family's constant, which names every
// sibling variable and mentions no value, so both write an identical `box-shadow` and stack rather
// than overwrite. A property in neither set is an ordinary declaration and takes the value, which is
// what makes `border-2` and `border-4` collide.
//
// The declarations are copied rather than mutated. The emitters build fresh nodes per call today,
// and a caller that reads property names must not observe a value another caller asked for.
func withResolvedValue(root string, nodes []*Node, value string) []*Node {
	rewritten := make([]*Node, 0, len(nodes))
	for _, node := range nodes {
		if node.Kind != KindDeclaration || !node.ValuePresent {
			rewritten = append(rewritten, node)
			continue
		}
		replacement := *node
		if constant, shared := composedAggregateValues[composedAggregate{root: root, property: node.Property}]; shared {
			replacement.Value = constant
		} else {
			replacement.Value = value
		}
		rewritten = append(rewritten, &replacement)
	}
	return rewritten
}

// composedAggregateValues is the constant each layering family writes to its shared property.
//
// Read out of the shipped Tailwind 4.3.3 source rather than reconstructed, the same way the multi
// declaration slice's constants were. Every one names its family's variables and none mentions a
// value, which is the mechanism: two utilities of one root write different variables and the same
// shared declaration, so the browser sees both.
//
// `mask-composite` is here for the same reason with a different shape: it is a bare keyword rather
// than a var chain, and it is constant across every mask root, so it composes for the same reason.
//
// # Keyed on the root as well as the property, because a property is not constant everywhere
//
// A first version keyed on the property alone and was wrong on two roots, both caught by the
// non-composing gap controls rather than by reading. `bg` emits `background-image` carrying a
// gradient name for a named non-colour value, so `bg-linear-to-r` and `bg-conic` genuinely differ
// there, while `bg-conic` emits the same property carrying a constant. `filter` emits `filter`
// carrying the chain only when the class is the bare `filter`; with a value it carries the value.
//
// So the constant belongs to a root's use of a property rather than to the property. A pair absent
// from this map takes the value, which is the correct default: a root that writes its value into
// the shared property overwrites rather than layers, which is what `bg-red-500` against
// `bg-blue-500` does.
var composedAggregateValues = map[composedAggregate]string{
	{root: "shadow", property: "box-shadow"}:               cssBoxShadowValue,
	{root: "inset-shadow", property: "box-shadow"}:         cssBoxShadowValue,
	{root: "ring", property: "box-shadow"}:                 cssBoxShadowValue,
	{root: "inset-ring", property: "box-shadow"}:           cssBoxShadowValue,
	{root: "ring-offset", property: "box-shadow"}:          cssBoxShadowValue,
	{root: "drop-shadow", property: "filter"}:              cssFilterValue,
	{root: "scale", property: "scale"}:                     cssScaleZValue,
	{root: "bg-conic", property: "background-image"}:       cssConicGradientValue,
	{root: "bg-linear", property: "background-image"}:      cssLinearGradientValue,
	{root: "bg-radial", property: "background-image"}:      cssRadialGradientValue,
	{root: "mask-x-from", property: "mask-image"}:          cssMaskImageValue,
	{root: "mask-x-to", property: "mask-image"}:            cssMaskImageValue,
	{root: "mask-y-from", property: "mask-image"}:          cssMaskImageValue,
	{root: "mask-y-to", property: "mask-image"}:            cssMaskImageValue,
	{root: "mask-t-from", property: "mask-image"}:          cssMaskImageValue,
	{root: "mask-t-to", property: "mask-image"}:            cssMaskImageValue,
	{root: "mask-r-from", property: "mask-image"}:          cssMaskImageValue,
	{root: "mask-r-to", property: "mask-image"}:            cssMaskImageValue,
	{root: "mask-b-from", property: "mask-image"}:          cssMaskImageValue,
	{root: "mask-b-to", property: "mask-image"}:            cssMaskImageValue,
	{root: "mask-l-from", property: "mask-image"}:          cssMaskImageValue,
	{root: "mask-l-to", property: "mask-image"}:            cssMaskImageValue,
	{root: "mask-linear-from", property: "mask-image"}:     cssMaskImageValue,
	{root: "mask-linear-to", property: "mask-image"}:       cssMaskImageValue,
	{root: "mask-radial-from", property: "mask-image"}:     cssMaskImageValue,
	{root: "mask-radial-to", property: "mask-image"}:       cssMaskImageValue,
	{root: "mask-conic-from", property: "mask-image"}:      cssMaskImageValue,
	{root: "mask-conic-to", property: "mask-image"}:        cssMaskImageValue,
	{root: "mask-x-from", property: "mask-composite"}:      cssMaskCompositeValue,
	{root: "mask-x-to", property: "mask-composite"}:        cssMaskCompositeValue,
	{root: "mask-y-from", property: "mask-composite"}:      cssMaskCompositeValue,
	{root: "mask-y-to", property: "mask-composite"}:        cssMaskCompositeValue,
	{root: "mask-t-from", property: "mask-composite"}:      cssMaskCompositeValue,
	{root: "mask-t-to", property: "mask-composite"}:        cssMaskCompositeValue,
	{root: "mask-r-from", property: "mask-composite"}:      cssMaskCompositeValue,
	{root: "mask-r-to", property: "mask-composite"}:        cssMaskCompositeValue,
	{root: "mask-b-from", property: "mask-composite"}:      cssMaskCompositeValue,
	{root: "mask-b-to", property: "mask-composite"}:        cssMaskCompositeValue,
	{root: "mask-l-from", property: "mask-composite"}:      cssMaskCompositeValue,
	{root: "mask-l-to", property: "mask-composite"}:        cssMaskCompositeValue,
	{root: "mask-linear-from", property: "mask-composite"}: cssMaskCompositeValue,
	{root: "mask-linear-to", property: "mask-composite"}:   cssMaskCompositeValue,
	{root: "mask-radial-from", property: "mask-composite"}: cssMaskCompositeValue,
	{root: "mask-radial-to", property: "mask-composite"}:   cssMaskCompositeValue,
	{root: "mask-conic-from", property: "mask-composite"}:  cssMaskCompositeValue,
	{root: "mask-conic-to", property: "mask-composite"}:    cssMaskCompositeValue,
}

// composedAggregate is one root's use of one shared property.
type composedAggregate struct {
	root     string
	property string
}

// The shared constants these gap roots write, read from the shipped 4.3.3 source.
const (
	cssBoxShadowValue      = "var(--tw-inset-shadow), var(--tw-inset-ring-shadow), var(--tw-ring-offset-shadow), var(--tw-ring-shadow), var(--tw-shadow)"
	cssConicGradientValue  = "conic-gradient(var(--tw-gradient-stops))"
	cssLinearGradientValue = "linear-gradient(var(--tw-gradient-stops))"
	cssRadialGradientValue = "radial-gradient(var(--tw-gradient-stops))"
	cssMaskCompositeValue  = "intersect"
)

package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// RestrictTemplateExpressionsOptions configures which non-string types may be interpolated.
//
// Every boolean is a POINTER so that absent stays distinguishable from an explicit false. Five of
// the six default to TRUE, so collapsing that distinction would turn an unconfigured rule into a
// far stricter one than upstream's default, which is the exact inversion the brief warns about for
// a default-true option.
type RestrictTemplateExpressionsOptions struct {
	// Allow are specifiers for types that may be interpolated regardless of the flags below.
	//
	// Upstream's default is not empty: `Error`, `URL` and `URLSearchParams` from the standard
	// library. A decoder that left this nil would report every `${error}` in the tree.
	Allow []type_checking.TypeOrValueSpecifier

	// AllowInline is the bare-string form of an allow entry, matched against the type's own name.
	AllowInline []string

	AllowAny     *bool
	AllowArray   *bool
	AllowBoolean *bool
	AllowNever   *bool
	AllowNullish *bool
	AllowNumber  *bool
	AllowRegExp  *bool
}

// defaultAllowSpecifiers is upstream's `allow` default, verbatim.
//
// One specifier naming three standard-library types. `from: 'lib'` means the declaration has to come
// from a default library file rather than from anything in the project that happens to be called
// `Error`.
func defaultAllowSpecifiers() []type_checking.TypeOrValueSpecifier {
	return []type_checking.TypeOrValueSpecifier{{
		From: type_checking.TypeOrValueSpecifierFromLib,
		Name: []string{"Error", "URL", "URLSearchParams"},
	}}
}

// restrictTemplateExpressionsEnabled reads a default-true flag.
func restrictTemplateExpressionsEnabled(flag *bool, whenAbsent bool) bool {
	if flag == nil {
		return whenAbsent
	}
	return *flag
}

// restrictTemplateExpressionsRawOptions is the wire shape.
//
// The booleans bind straight through as pointers. `allow` does not: it is one heterogeneous array
// whose entries are a bare string or a specifier object, and it leaves as two typed fields.
type restrictTemplateExpressionsRawOptions struct {
	Allow        []onlyThrowErrorRawSpecifier `json:"allow"`
	AllowAny     *bool                        `json:"allowAny"`
	AllowArray   *bool                        `json:"allowArray"`
	AllowBoolean *bool                        `json:"allowBoolean"`
	AllowNever   *bool                        `json:"allowNever"`
	AllowNullish *bool                        `json:"allowNullish"`
	AllowNumber  *bool                        `json:"allowNumber"`
	AllowRegExp  *bool                        `json:"allowRegExp"`
}

// DecodeRestrictTemplateExpressionsOptions maps upstream's JSON onto the struct the rule reads.
//
// Hand-written rather than `rule.DecodeOptionsInto` for two reasons, and the second is the one that
// would silently break the rule. `allow` is heterogeneous and leaves as two fields, reusing the
// specifier decoder `only-throw-error` already ships for the identical wire shape. And the `allow`
// DEFAULT is not the zero value: absent `allow` means upstream's three standard-library types
// rather than an empty list, so a decoder returning nil there would report every interpolated
// `Error` in the tree while every fixture built from an explicit struct still passed.
func DecodeRestrictTemplateExpressionsOptions(raw []byte) (any, error) {
	options := RestrictTemplateExpressionsOptions{Allow: defaultAllowSpecifiers()}
	if len(raw) == 0 {
		return options, nil
	}

	decoded, err := rule.DecodeOptionsInto[restrictTemplateExpressionsRawOptions]()(raw)
	if err != nil {
		return options, err
	}
	wire, _ := decoded.(restrictTemplateExpressionsRawOptions)

	options.AllowAny = wire.AllowAny
	options.AllowArray = wire.AllowArray
	options.AllowBoolean = wire.AllowBoolean
	options.AllowNever = wire.AllowNever
	options.AllowNullish = wire.AllowNullish
	options.AllowNumber = wire.AllowNumber
	options.AllowRegExp = wire.AllowRegExp

	// An explicitly configured `allow` REPLACES the default rather than adding to it, matching
	// upstream, where `defaultOptions` is merged key by key and a supplied array wins outright.
	if wire.Allow != nil {
		options.Allow = nil
		options.AllowInline = nil
		for _, entry := range wire.Allow {
			if entry.inline != "" {
				options.AllowInline = append(options.AllowInline, entry.inline)
				continue
			}
			specifier := type_checking.TypeOrValueSpecifier{
				Name: entry.Name, Path: entry.Path, Package: entry.Package,
			}
			switch entry.From {
			case "file":
				specifier.From = type_checking.TypeOrValueSpecifierFromFile
			case "lib":
				specifier.From = type_checking.TypeOrValueSpecifierFromLib
			case "package":
				specifier.From = type_checking.TypeOrValueSpecifierFromPackage
			default:
				continue
			}
			options.Allow = append(options.Allow, specifier)
		}
	}

	return options, nil
}

// RestrictTemplateExpressions flags a value interpolated into a template literal whose type is not
// a string and is not one of the kinds the configuration allows.
//
//	valid:   `${'a'}`
//	valid:   `${1}`                      allowNumber defaults on
//	valid:   `${new Error('x')}`         the default allow list names Error
//	invalid: `${{}}`
//	invalid: `${x}`                      where x is unknown
//
// Interpolating a non-string calls its `toString`, and for most values that produces something
// nobody wants in a message: a plain object becomes `[object Object]`, an array of objects becomes a
// row of them separated by commas. The rule exists so those reach a reviewer rather than a log line,
// and the option surface is wide because which of them are actually fine differs by project.
//
// # Seven allowances, five of them on by default, and the defaults are the whole option surface
//
// `allowAny`, `allowBoolean`, `allowNullish`, `allowNumber` and `allowRegExp` default to TRUE;
// `allowArray` and `allowNever` default to false. So the rule as shipped reports objects, unknown,
// arrays, never, and anything else without a sensible string form, and lets the rest through. The
// options are pointers all the way through the decoder for exactly this reason: reading an absent
// key as false would turn an unconfigured rule into upstream's `strict` preset.
//
// `allow` is the other half and its default is not empty either. Upstream ships
// `[{ name: ['Error', 'URL', 'URLSearchParams'], from: 'lib' }]`, so an interpolated `Error` is fine
// out of the box. A decoder defaulting it to nil passes every fixture that supplies options and
// reports every `${error}` on a real tree.
//
// # The recursion, and why the two composite arms differ
//
// A union is allowed only if EVERY constituent is; an intersection if ANY constituent is. That
// asymmetry is correct rather than a typo: a union value could be any of its members at run time, so
// all of them have to be printable, while an intersection value is all of its members at once, so
// one printable facet is enough.
//
// The array tester recurses too, through the number index type, which is what makes `string[]`
// allowed under `allowArray` while `object[]` is not.
//
// # A tagged template is not this rule's business
//
// A tagged template hands the raw parts to a function that can do anything with them, so there is no
// `toString` to be wrong about. Upstream excludes it by testing the parent, and so does this.
//
// # Cost
//
// The anchor is a template literal, and the body exits immediately on one with no interpolations,
// which is the common case for a template used only for its multi-line-ness.
var RestrictTemplateExpressions = rule.Rule{
	Name: "@typescript-eslint/restrict-template-expressions",

	// Every judgment is about the TYPE of an interpolated value.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	// Compiler options and the default library, through type_checking's builtin and specifier helpers.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDefaultLibrary,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, isSettings := rule.OptionsAs[RestrictTemplateExpressionsOptions](options)
		if !isSettings {
			// A rule configured as a bare severity reaches here with nil options, and the zero
			// value of this struct is upstream's STRICT preset rather than its default. Routing
			// through the decoder keeps `"error"` meaning what upstream means by it.
			decoded, _ := DecodeRestrictTemplateExpressionsOptions(nil)
			settings, _ = decoded.(RestrictTemplateExpressionsOptions)
		}

		allowAny := restrictTemplateExpressionsEnabled(settings.AllowAny, true)
		allowArray := restrictTemplateExpressionsEnabled(settings.AllowArray, false)
		allowBoolean := restrictTemplateExpressionsEnabled(settings.AllowBoolean, true)
		allowNever := restrictTemplateExpressionsEnabled(settings.AllowNever, false)
		allowNullish := restrictTemplateExpressionsEnabled(settings.AllowNullish, true)
		allowNumber := restrictTemplateExpressionsEnabled(settings.AllowNumber, true)
		allowRegExp := restrictTemplateExpressionsEnabled(settings.AllowRegExp, true)

		var isAllowedType func(candidate *checker.Type) bool
		isAllowedType = func(candidate *checker.Type) bool {
			if candidate == nil {
				return true
			}

			// A union is allowed only if every constituent is, because the value could be any of
			// them at run time.
			if candidate.IsUnion() {
				for _, part := range candidate.AsUnionType().Types() {
					if !isAllowedType(part) {
						return false
					}
				}
				return true
			}

			// An intersection is allowed if any constituent is, because the value is all of them at
			// once and one printable facet is enough.
			if candidate.IsIntersection() {
				for _, part := range candidate.AsIntersectionType().Types() {
					if isAllowedType(part) {
						return true
					}
				}
				return false
			}

			if type_checking.IsTypeFlagSet(candidate, checker.TypeFlagsStringLike) {
				return true
			}

			// The configured allowlist, checked against the type and its base types so a subclass
			// of an allowed type is allowed too. That is upstream's `matchesTypeOrBaseType`.
			if matchesAllowedTypeOrBaseType(ctx, candidate, settings,
				map[*checker.Type]struct{}{}) {
				return true
			}

			if allowAny && type_checking.IsTypeAnyType(candidate) {
				return true
			}
			if allowBoolean && type_checking.IsTypeFlagSet(candidate, checker.TypeFlagsBooleanLike) {
				return true
			}
			if allowNullish && type_checking.IsTypeFlagSet(candidate,
				checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
				return true
			}
			if allowNumber && type_checking.IsTypeFlagSet(candidate,
				checker.TypeFlagsNumberLike|checker.TypeFlagsBigIntLike) {
				return true
			}
			if allowRegExp && type_checking.GetTypeName(ctx.TypeChecker, candidate) == "RegExp" {
				return true
			}
			if allowNever && type_checking.IsTypeFlagSet(candidate, checker.TypeFlagsNever) {
				return true
			}

			// The array tester recurses through the number index type, which is what separates
			// `string[]` from `object[]` under the same flag.
			if allowArray && (checker.Checker_isArrayType(ctx.TypeChecker, candidate) ||
				checker.IsTupleType(candidate)) {
				return isAllowedType(type_checking.GetNumberIndexType(ctx.TypeChecker, candidate))
			}

			return false
		}

		return rule.Listeners{
			ast.KindTemplateExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}

				// A tagged template hands the raw parts to a function, so there is no `toString`
				// for this rule to have an opinion about.
				if node.Parent != nil && node.Parent.Kind == ast.KindTaggedTemplateExpression {
					return
				}

				templateExpression := node.AsTemplateExpression()
				if templateExpression.TemplateSpans == nil {
					return
				}

				for _, span := range templateExpression.TemplateSpans.Nodes {
					expression := span.AsTemplateSpan().Expression
					if expression == nil {
						continue
					}

					expressionType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker,
						expression)
					if expressionType == nil || isAllowedType(expressionType) {
						continue
					}

					ctx.ReportNode(expression, buildInvalidTypeMessage(
						ctx.TypeChecker.TypeToString(expressionType)))
				}
			},
		}
	},
}

// matchesAllowedTypeOrBaseType answers upstream's `matchesTypeOrBaseType` over the allowlist.
//
// The type itself first, then its base types RECURSIVELY, so a class extending `Error` is allowed by
// the default specifier naming `Error` and so is a class extending that class. The recursion is not
// optional and a single level is not enough: upstream's corpus pins the transitive case in both the
// class and the interface direction, with `class DerivedTwice extends Derived extends Base` and an
// interface reached through two `extends` hops, and a one-level walk reports both. This port shipped
// that defect until those two cases ran.
//
// `seen` reproduces upstream's cycle guard rather than adding one. It is load-bearing on interfaces,
// where a declaration-merged hierarchy can reach itself.
func matchesAllowedTypeOrBaseType(ctx rule.Context, candidate *checker.Type,
	settings RestrictTemplateExpressionsOptions, seen map[*checker.Type]struct{}) bool {
	if candidate == nil {
		return false
	}
	if _, visited := seen[candidate]; visited {
		return false
	}
	seen[candidate] = struct{}{}

	if type_checking.TypeMatchesSomeSpecifier(candidate, settings.Allow, settings.AllowInline,
		ctx.Program) {
		return true
	}

	for _, baseType := range baseTypesOf(ctx, candidate) {
		if matchesAllowedTypeOrBaseType(ctx, baseType, settings, seen) {
			return true
		}
	}
	return false
}

// baseTypesOf answers upstream's `getBaseTypesForType`.
//
// Two hops that are easy to skip and both matter. A generic type arrives as a type REFERENCE whose
// base types live on its target rather than on the reference, and `getBaseTypes` answers only for a
// class or interface, so anything else returns nothing rather than being asked.
func baseTypesOf(ctx rule.Context, candidate *checker.Type) []*checker.Type {
	if !type_checking.IsObjectType(candidate) {
		return nil
	}

	interfaceTarget := candidate
	if checker.IsNonDeferredTypeReference(candidate) {
		if target := candidate.Target(); target != nil {
			interfaceTarget = target
		}
	}

	if checker.Type_objectFlags(interfaceTarget)&
		(checker.ObjectFlagsInterface|checker.ObjectFlagsClass) == 0 {
		return nil
	}

	return checker.Checker_getBaseTypes(ctx.TypeChecker, interfaceTarget)
}

func buildInvalidTypeMessage(typeName string) rule.Message {
	return rule.Message{
		Id:          "invalidType",
		Description: "Invalid type \"" + typeName + "\" of template literal expression.",
	}
}

package structure

import (
	"path"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// messageRequireMatchingFileName names the component, the file, and the rename.
//
// The second option is offered only when the file's own name could be a component's, because
// advising `function financeFloorView` to satisfy `financeFloorView.tsx` would trade this finding
// for a component nobody can render as JSX.
// consistencyRequireMatchingFileNameText is the rule's message, whose wording lives in
// `policy/messages/consistency-require-matching-file-name.json`, with the second rename as the
// `alternative` phrase.
var (
	consistencyRequireMatchingFileNameText            = policy.MessageOf("structure/consistency-require-matching-file-name", "requireMatchingFileName")
	consistencyRequireMatchingFileNameRenameComponent = consistencyRequireMatchingFileNameText.Option("alternative", "renameComponent")
	consistencyRequireMatchingFileNameNoAlternative   = consistencyRequireMatchingFileNameText.Option("alternative", "none")
)

func messageRequireMatchingFileName(componentName string, fileName string, fileBaseName string) rule.Message {
	values := map[string]string{
		"componentName":     componentName,
		"fileName":          fileName,
		"componentFileName": componentName + path.Ext(fileName),
	}
	alternative := consistencyRequireMatchingFileNameNoAlternative
	if react.IsLikelyComponentName(fileBaseName) && isIdentifierText(fileBaseName) {
		values["fileBaseName"] = fileBaseName
		alternative = consistencyRequireMatchingFileNameRenameComponent
	}
	return rule.Message{
		Id:          consistencyRequireMatchingFileNameText.Id,
		Description: consistencyRequireMatchingFileNameText.Render(values, alternative),
	}
}

// nextJsConventionFileBaseNames are the base names Next.js reads from the filesystem by name.
//
// This is the framework's grammar, not an allowance. Next resolves a route segment by finding a file
// literally called `page` or `layout`, so the file's name is fixed before anyone writes the component
// in it, and there is no rename that would satisfy this rule without breaking the route. The
// component name in these files is free to carry what the path cannot, and it does: measured on the
// ahra tree on 2026-10-01, 77 of 80 `page.tsx` files default-export a `<Route>PageRoute` and every
// layout a `<Segment>Layout`. That is a naming convention of its own, and enforcing it would be a
// separate rule rather than a reading of this one.
//
// The list is Next 16.3.1's own, read from the installed build rather than from memory:
// `FILE_TYPES` and `HTTP_ACCESS_FALLBACKS` in `next/dist/build/webpack/loaders/next-app-loader`
// (layout, template, error, loading, global-error, global-not-found, not-found, forbidden,
// unauthorized), plus page, route and default, plus `mdx-components`, which Next resolves at the
// project root. The metadata image conventions are below, because they take a numeric suffix.
//
// It is wider than `FileContext.IsSpecialNextJsFile`, which names six of these plus route. Widening
// that one would move every rule reading it, `react-component-require-named-export` among them, and
// whether `template.tsx` should be exempt from those is their question rather than this rule's.
//
// The base name alone decides, as it does in `FileContextFor`, so a `page.tsx` outside an app
// directory is exempt too. Measured, there is none in the ahra tree, and such a file's name is
// lowercase, which no component name can match anyway.
var nextJsConventionFileBaseNames = map[string]bool{
	"page":             true,
	"layout":           true,
	"template":         true,
	"loading":          true,
	"error":            true,
	"global-error":     true,
	"not-found":        true,
	"global-not-found": true,
	"forbidden":        true,
	"unauthorized":     true,
	"default":          true,
	"route":            true,
	"mdx-components":   true,

	// The metadata routes that are plain functions rather than images. They are `.ts` in practice
	// and listed for the `.tsx` spelling Next also accepts.
	"sitemap":  true,
	"robots":   true,
	"manifest": true,
}

// nextJsMetadataImageBaseNames are the generated-image conventions, which Next also matches with a
// single trailing digit: `icon1.tsx` and `opengraph-image2.tsx` are routes. The pattern is
// `<name>\d?` in `next/dist/lib/metadata/is-metadata-route.js`.
var nextJsMetadataImageBaseNames = []string{"icon", "apple-icon", "opengraph-image", "twitter-image"}

// companionFileSuffixes mark a file that is about a component rather than the component's home.
//
// `Button.test.tsx` and `Button.stories.tsx` exist because `Button.tsx` does, and the components
// declared inside them are fixtures and harnesses for that one. Requiring a harness to be named
// `Button.test` is impossible, and requiring it to be named `Button` would claim the subject's name
// for its scaffolding. Measured: the ahra tree holds no file with any of these suffixes today, so
// this is a decision for the day one appears rather than an exemption anything is relying on.
//
// Only these four. Any other dotted name, `Button.client.tsx` say, is compared whole and reports,
// because an exact match is the rule and a role suffix this list does not name is a name that does
// not match.
var companionFileSuffixes = []string{".test", ".spec", ".stories", ".story"}

// ConsistencyRequireMatchingFileName flags a file whose one component is named differently.
//
//	valid:   FinanceFloorViewRow.tsx    export function FinanceFloorViewRow() { ... }
//	valid:   page.tsx                   export default function OsPageRoute() { ... }
//	valid:   Table.tsx                  two components             react/no-multi-comp's finding
//	invalid: FinanceFloorView.tsx       export function FinanceFloorViewRow() { ... }
//	invalid: financeFloorView.tsx       export function FinanceFloorView() { ... }
//
// One component per file is `react/no-multi-comp`'s rule, and it is strict. This is the other half
// of the same promise: once a file holds one component, its name is the component's, so the tree can
// be navigated by name in both directions forever. Knowing `FinanceFloorViewRow` means knowing the
// file, and seeing the file means knowing the component.
//
// # What counts, and which way the detector errs
//
// Top-level declarations only, because the file's component is the one the file declares, not a
// component nested inside it. Each of these is one component, named by its binding:
//
//	a capitalized function declaration
//	a capitalized const holding an arrow or function expression
//	a capitalized const holding memo(...) or forwardRef(...), bare or React.-qualified
//	a capitalized const holding Object.assign(Root, ...), when Root is one of the above
//	a class extending Component or PureComponent, declared or assigned to a const
//	an anonymous default export of a function, which has no name
//
// A function or function literal is a component if either predicate in `react_detection.go` says so:
// `IsLikelyReactComponent` (a `properties` or `props` parameter, or a top-level JSX return) or
// `HasJsxOrReactHookCalls` (JSX or a hook call anywhere in a bounded walk). The house rules use one or
// the other, and this one takes both, because of which direction a miss costs here. This rule
// reports when exactly one component is found, so a detector that misses one of two components
// reports the other under a false premise: the file is not named for it because it is not the file's
// component. A detector that over-counts only makes a file read as "several", which is silent. So
// the detector is built to over-count, and classes are counted although no house rule counts them,
// for the same reason: an error boundary class beside a function fallback is two components, and
// counting only the function would report the fallback against the boundary's file.
//
// # A wrapper absorbs the function it wraps
//
// `function MenuItemComponent() {}` with `export const MenuItem = React.memo(MenuItemComponent)` is
// one component, `MenuItem`, implemented by the function it names. Five files in the structure
// library are written this way. Counting both would make every one of them silent as "several", so
// a wrapper's first argument, when it is an identifier naming a function this file declares, is
// dropped from the count and the wrapper stands for it. Type assertions around the call are looked
// through, since `React.memo(Inner, compare) as typeof Inner` is how three of those five keep a
// generic component's type parameter.
//
// The structure library's compound components take the same treatment:
// `export const Calendar = Object.assign(CalendarRoot, { Header: CalendarHeader })` is the component
// `Calendar`, implemented by `CalendarRoot`. Unlike memo, `Object.assign` builds plenty of things
// that are not components, so it counts only when its first argument names a component this file
// declares. Both shapes were found by the first run on the ahra tree, where each reported an
// implementation function against the file named for its wrapper.
//
// # When the rule stays silent, and whose finding that is instead
//
//	zero components               nothing to name the file for
//	two or more components        react/no-multi-comp reports every one after the first, and this
//	                              rule waits until the split makes the question answerable
//	an anonymous component        react/display-name reports it; once named, this rule compares
//	a Next.js convention file     the framework owns the name; see nextJsConventionFileBaseNames
//	a test or stories file        see companionFileSuffixes
//	a .ts or .js file             a component file is a .tsx or .jsx file
//
// `index.tsx` is not exempt. The house bans barrel files, so an `index.tsx` that declares a component
// is a component living under a name that is not its own, and it reports like any other. A barrel
// that only re-exports declares no component and is silent here; whether it should exist is not this
// rule's question.
//
// # No fix
//
// The repair is renaming a file, which changes every import of it across the tree. That is a rename
// with its own tooling and its own review, not a lint edit applied unattended to one file, and an
// edit to one file cannot express it anyway.
var ConsistencyRequireMatchingFileName = rule.Rule{
	Name:       "structure/consistency-require-matching-file-name",
	NoListener: rule.NoListenerAnswersInRun,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		// No backslash fold before the base name, unlike FileContextFor. Every file name reaching a
		// rule has already been through `tspath.NormalizePath`, in the program and in the test
		// harness alike, so a fold here could never change anything. Measured rather than assumed:
		// a probe rule run on `C:\repository\source\Button.tsx` read `C:/repository/source/Button.tsx`,
		// and a mutation deleting the fold survived a fixture written to catch it.
		fileName := path.Base(ctx.SourceFile.FileName().AsString())
		fileBaseName, isComponentFile := componentFileBaseName(fileName)
		if !isComponentFile || isNextJsConventionFile(fileBaseName) || isCompanionFile(fileBaseName) {
			return nil
		}

		// The decision is about the file rather than any node, so it is made once here over the
		// top-level statements, and no listener is registered.
		components := componentsNamingTheFile(ctx.SourceFile.AsNode())
		if len(components) != 1 {
			return nil
		}
		component := components[0]
		if component.nameNode == nil {
			return nil
		}
		if component.nameNode.Text() == fileBaseName {
			return nil
		}
		ctx.ReportNode(component.nameNode,
			messageRequireMatchingFileName(component.nameNode.Text(), fileName, fileBaseName))
		return nil
	},
}

// componentFileBaseName strips the React extension, and reports whether there was one.
func componentFileBaseName(fileName string) (string, bool) {
	for _, extension := range []string{".tsx", ".jsx"} {
		if strings.HasSuffix(fileName, extension) {
			return strings.TrimSuffix(fileName, extension), true
		}
	}
	return "", false
}

// isNextJsConventionFile reports a base name Next.js reads by name.
func isNextJsConventionFile(fileBaseName string) bool {
	if nextJsConventionFileBaseNames[fileBaseName] {
		return true
	}
	for _, imageBaseName := range nextJsMetadataImageBaseNames {
		if fileBaseName == imageBaseName {
			return true
		}
		// One digit and only one, matching Next's `\d?`. `icon12.tsx` is not a route.
		if len(fileBaseName) == len(imageBaseName)+1 && strings.HasPrefix(fileBaseName, imageBaseName) {
			last := fileBaseName[len(fileBaseName)-1]
			if last >= '0' && last <= '9' {
				return true
			}
		}
	}
	return false
}

// isCompanionFile reports a test or stories file.
func isCompanionFile(fileBaseName string) bool {
	for _, suffix := range companionFileSuffixes {
		if strings.HasSuffix(fileBaseName, suffix) {
			return true
		}
	}
	return false
}

// isIdentifierText reports whether a file's base name could be written as a component's name.
//
// ASCII letters, digits, `_` and `$`, not starting with a digit. A base name with a dash or a dot
// cannot be a binding, so offering it as the component's new name would offer something that does
// not parse.
func isIdentifierText(text string) bool {
	if text == "" {
		return false
	}
	for index, character := range text {
		switch {
		case character >= 'a' && character <= 'z', character >= 'A' && character <= 'Z',
			character == '_', character == '$':
		case character >= '0' && character <= '9' && index > 0:
		default:
			return false
		}
	}
	return true
}

// fileComponent is one component a file declares at the top level. A nil nameNode is an anonymous
// default export, which counts toward "several" and is never compared.
type fileComponent struct {
	nameNode *ast.Node
}

// compoundComponent is a capitalized const built by `Object.assign(Root, { ... })`, waiting to learn
// whether Root is one of this file's components.
type compoundComponent struct {
	nameNode *ast.Node
	wrapped  string
}

// isObjectAssignCall reports `Object.assign(...)`, the structure library's compound-component idiom.
func isObjectAssignCall(call *ast.Node) bool {
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := callee.AsPropertyAccessExpression()
	receiver := ast.SkipParentheses(access.Expression)
	name := access.Name()
	return receiver != nil && receiver.Kind == ast.KindIdentifier && receiver.Text() == "Object" &&
		name != nil && name.Kind == ast.KindIdentifier && name.Text() == "assign"
}

// componentsNamingTheFile collects the file's top-level components in source order, with wrapped
// implementations absorbed and repeated names counted once.
func componentsNamingTheFile(sourceFile *ast.Node) []fileComponent {
	var components []fileComponent
	var compounds []compoundComponent
	seen := map[string]bool{}
	absorbed := map[string]bool{}

	add := func(nameNode *ast.Node) {
		if nameNode == nil {
			components = append(components, fileComponent{})
			return
		}
		// Overload signatures each declare the name again, and `function Thing(properties: A);`
		// reads as a component by its parameter, so a name is one component however many times it
		// is declared.
		if seen[nameNode.Text()] {
			return
		}
		seen[nameNode.Text()] = true
		components = append(components, fileComponent{nameNode: nameNode})
	}

	sourceFile.ForEachChild(func(statement *ast.Node) bool {
		switch statement.Kind {
		case ast.KindFunctionDeclaration:
			name := statement.AsFunctionDeclaration().Name()
			if name == nil {
				// `export default function () {}`. A nameless declaration is only legal as a
				// default export, so no export check is needed to know what this is.
				if isFunctionComponent(statement) {
					add(nil)
				}
				return false
			}
			if react.IsLikelyComponentName(name.Text()) && isFunctionComponent(statement) {
				add(name)
			}

		case ast.KindClassDeclaration:
			name := statement.AsClassDeclaration().Name()
			if react.IsEs6ComponentClass(statement) {
				// `export default class extends Component {}` is the class form of the anonymous
				// default, and is counted without a name for the same reason.
				if name == nil {
					add(nil)
				} else {
					add(name)
				}
			}

		case ast.KindVariableStatement:
			declarationList := statement.AsVariableStatement().DeclarationList
			if declarationList == nil {
				return false
			}
			for _, declarationNode := range declarationList.AsVariableDeclarationList().Declarations.Nodes {
				declaration := declarationNode.AsVariableDeclaration()
				name := declaration.Name()
				if name == nil || name.Kind != ast.KindIdentifier || !react.IsLikelyComponentName(name.Text()) {
					continue
				}
				// The nil check precedes the skip, which dereferences its argument, and `let Thing;`
				// is ordinary code.
				if declaration.Initializer == nil {
					continue
				}
				// Assertions as well as parentheses. `React.memo(Inner, compare) as typeof Inner`
				// is how a memoized generic component keeps its type parameter, and three files in
				// the structure library are written that way. Skipping only parentheses left the
				// wrapper unseen and reported the implementation against the wrapper's file, which
				// is how the first run on the ahra tree found it.
				initializer := ast.SkipOuterExpressions(declaration.Initializer, ast.OEKParentheses|ast.OEKAssertions)
				if initializer == nil {
					continue
				}
				switch initializer.Kind {
				case ast.KindArrowFunction, ast.KindFunctionExpression:
					if isFunctionComponent(initializer) {
						add(name)
					}
				case ast.KindClassExpression:
					if react.IsEs6ComponentClass(initializer) {
						add(name)
					}
				case ast.KindCallExpression:
					if isAnonymousWrapperCall(initializer) {
						add(name)
						if wrapped := wrappedIdentifier(initializer); wrapped != "" {
							absorbed[wrapped] = true
						}
					} else if isObjectAssignCall(initializer) {
						if wrapped := wrappedIdentifier(initializer); wrapped != "" {
							compounds = append(compounds, compoundComponent{nameNode: name, wrapped: wrapped})
						}
					}
				}
			}

		case ast.KindExportAssignment:
			// `export default () => <div />` and `export default function () {}` as an expression.
			// A named function expression carries a name, and that name is the component's.
			expression := statement.AsExportAssignment().Expression
			if expression == nil {
				return false
			}
			expression = ast.SkipParentheses(expression)
			if expression == nil {
				return false
			}
			switch expression.Kind {
			case ast.KindArrowFunction:
				if isFunctionComponent(expression) {
					add(nil)
				}
			case ast.KindFunctionExpression:
				if isFunctionComponent(expression) {
					add(expression.AsFunctionExpression().Name())
				}
			}
		}
		return false
	})

	// Resolved after the walk, because the compound is declared below the root it attaches to and
	// whether the root is a component is only known once the walk has passed it.
	for _, compound := range compounds {
		if seen[compound.wrapped] {
			add(compound.nameNode)
			absorbed[compound.wrapped] = true
		}
	}

	if len(absorbed) == 0 {
		return components
	}
	kept := components[:0]
	for _, component := range components {
		if component.nameNode != nil && absorbed[component.nameNode.Text()] {
			continue
		}
		kept = append(kept, component)
	}
	return kept
}

// isFunctionComponent is either of react_detection.go's predicates. See the rule's doc comment for
// why this rule takes the union rather than one of them.
func isFunctionComponent(functionLike *ast.Node) bool {
	return IsLikelyReactComponent(functionLike) || HasJsxOrReactHookCalls(functionLike)
}

// wrappedIdentifier is the name a memo or forwardRef call wraps, when its first argument is a bare
// identifier, and empty otherwise. An inline function argument is part of the wrapper and was never
// a top-level component to absorb.
func wrappedIdentifier(call *ast.Node) string {
	arguments := call.AsCallExpression().Arguments
	if arguments == nil || len(arguments.Nodes) == 0 || arguments.Nodes[0] == nil {
		return ""
	}
	first := ast.SkipParentheses(arguments.Nodes[0])
	if first == nil || first.Kind != ast.KindIdentifier {
		return ""
	}
	return first.Text()
}

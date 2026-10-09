package nexus

import (
	"regexp"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/ecmascript/nextjs"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// ConsistencyRequireConstantCasingOptions names exported constants a framework or library reads
// verbatim, so the spelling is a contract rather than a choice.
//
// Next.js reads "runtime" and "dynamic" off a route module; graphql-codegen requires a plugin module
// to export "plugin". The rule cannot know which of these apply, so the consumer names them.
type ConsistencyRequireConstantCasingOptions struct {
	FrameworkConstantNames []string `json:"frameworkConstantNames"`
}

// The rule's messages, whose wording lives in `policy/messages/consistency-require-constant-casing.json`.
// The reasoning behind each message is the part that ports. A rule that says only what is wrong gets
// disabled the first time it is inconvenient; one that says why gets fixed.
var (
	constantCasingRequirePascalCaseExportedText = policy.MessageOf("nexus/consistency-require-constant-casing", "requirePascalCaseExported")
	constantCasingDropUnimportedExportText      = policy.MessageOf("nexus/consistency-require-constant-casing", "dropUnimportedExport")
	constantCasingRequireCamelCaseLocalText     = policy.MessageOf("nexus/consistency-require-constant-casing", "requireCamelCaseLocal")
	constantCasingRequireCamelCaseFunctionText  = policy.MessageOf("nexus/consistency-require-constant-casing", "requireCamelCaseFunction")
	constantCasingRequireCamelCaseInstanceText  = policy.MessageOf("nexus/consistency-require-constant-casing", "requireCamelCaseInstance")
)

func messageRequirePascalCaseExported(name string, suggestion string) rule.Message {
	return rule.Message{
		Id:          constantCasingRequirePascalCaseExportedText.Id,
		Description: constantCasingRequirePascalCaseExportedText.Render(map[string]string{"name": name, "suggestion": suggestion}),
	}
}

// messageDropUnimportedExport is the exported-camelCase finding when nothing in the program imports
// the name, so the export itself is the mistake rather than the casing.
//
// Telling the author to PascalCase a name nobody imports makes the export look intended and keeps a
// public surface nobody uses. MediumConversationSource.ts exported two camelCase constants that only
// its own file read; the honest repair is to drop the `export`, after which camelCase is already
// right. The last sentence is there because the program is not the world: a library file can be
// imported from another repository this run cannot see.
func messageDropUnimportedExport(name string, suggestion string) rule.Message {
	return rule.Message{
		Id:          constantCasingDropUnimportedExportText.Id,
		Description: constantCasingDropUnimportedExportText.Render(map[string]string{"name": name, "suggestion": suggestion}),
	}
}

func messageRequireCamelCaseLocal(name string, suggestion string) rule.Message {
	return rule.Message{
		Id:          constantCasingRequireCamelCaseLocalText.Id,
		Description: constantCasingRequireCamelCaseLocalText.Render(map[string]string{"name": name, "suggestion": suggestion}),
	}
}

func messageRequireCamelCaseFunction(name string, suggestion string) rule.Message {
	return rule.Message{
		Id:          constantCasingRequireCamelCaseFunctionText.Id,
		Description: constantCasingRequireCamelCaseFunctionText.Render(map[string]string{"name": name, "suggestion": suggestion}),
	}
}

func messageRequireCamelCaseInstance(name string, suggestion string) rule.Message {
	return rule.Message{
		Id:          constantCasingRequireCamelCaseInstanceText.Id,
		Description: constantCasingRequireCamelCaseInstanceText.Render(map[string]string{"name": name, "suggestion": suggestion}),
	}
}

// ConsistencyRequireConstantCasing maps a constant's casing to what it is.
//
//	valid:   export const OrderColumns = [...]
//	valid:   const orderColumns = [...]
//	valid:   export const networkService = new NetworkService()
//	valid:   export const formatNumber = (value: number) => String(value)
//	invalid: export const numberAbsentPlaceholder = '-'
//	invalid: const OrderColumns = [...]
//	invalid: export const FormatNumber = (value: number) => String(value)
//
// Three clauses. An exported constant holding a value is PascalCase; a file-local one is camelCase;
// and a constant holding a class instance or a function is camelCase whatever its reach.
//
// The third clause is where casing stops meaning reach and starts meaning kind. PascalCase is the
// older, louder signal for "this is a type", so a PascalCase name on an instance reads as its class
// and quietly lies: IncantationsStore looks like the class, while the class is actually
// AhraOsIncantationsStore. An instance is a thing that acts, and incantationsStore.findById() says
// so at the call site. Built-in containers are the exception, because they hold rather than act: a
// new Map() is a data structure with an access syntax, so an exported one is exported data.
//
// SCREAMING_SNAKE_CASE is deliberately not flagged here. It is neither casing, so it would fire, but
// consistency-no-screaming-snake-case already owns that shape with a better message and with the
// exemptions it needs. Two rules, no overlap, each sharp at its own failure site.
//
// A constant that is not exported is renamed by the fix, at its declaration and every reference the
// checker resolves to it, as one edit; `fileLocalRenameFix` says when it declines and why. An
// exported one is reported without a fix, because its references cross files and a rename there is
// a change to every importer.
//
// An exported camelCase constant that nothing in the program imports is told to drop the export
// instead of being told to PascalCase, since then the export is the mistake. "Nothing imports it"
// is read from every module specifier in the program and refuses on any doubt; see
// `constant_casing_importers.go`.
var ConsistencyRequireConstantCasing = rule.Rule{
	Name: "nexus/consistency-require-constant-casing",
	// For the rename fix, which resolves references through the checker, and only at a finding.
	NeedsTypeChecker: true,
	// Every other file's imports, for the drop-the-export advice: resolving another file's imports is
	// a read of other files, which the view enforces, and a resolution, which the source guard sees.
	// The run's directory, which the fingerprint names files relative to.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsModuleResolution | rule.ReadsOtherFiles,
	// What a file's verdict reads beyond the file and its types is its own entry in the importer index
	// and whether it is a root, so the index and the roots are the fingerprint (#f96cnry). See
	// constantCasingFingerprint.
	ProgramFingerprint: constantCasingFingerprint,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		frameworkConstantNames := map[string]bool{}
		if settings, hasSettings := rule.OptionsAs[ConsistencyRequireConstantCasingOptions](options); hasSettings {
			for _, name := range settings.FrameworkConstantNames {
				frameworkConstantNames[name] = true
			}
		}

		return rule.Listeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				declaration := node.AsVariableDeclaration()
				if declaration == nil {
					return
				}

				// Destructuring binds several names at once and takes most of them from the shape of
				// the value rather than from a choice the author made, so only a plain identifier is
				// this rule's business.
				name := declaration.Name()
				if name == nil || name.Kind != ast.KindIdentifier {
					return
				}

				// let and var are rebindable, which is the premise the whole convention rests on.
				// Only const carries the immutability that frees its casing to mean scope.
				declarationList := node.Parent
				if declarationList == nil || declarationList.Flags&ast.NodeFlagsConst == 0 {
					return
				}
				statement := declarationList.Parent

				// A "declare const X" names something that already exists somewhere this file does
				// not control: a runtime global, a script host's API, a value injected by a bundler.
				// The declaration is a description rather than a naming decision, so the upstream
				// spelling is the contract. "declare const DocumentApp" describes Google Apps
				// Script's own global, and lowercasing it would make the script throw at runtime
				// while the type-checker stayed happy, because the declaration is all TypeScript
				// ever sees.
				if hasDeclareModifier(statement) {
					return
				}

				declaredName := name.Text()

				// Owned by consistency-no-screaming-snake-case, which knows about process.env and
				// carries an allowlist for constants mirroring an external system's grammar.
				if isScreamingSnakeCase(declaredName) {
					return
				}

				// The trailing underscore is escaping a keyword, so recasing it produces a name the
				// parser rejects. "do_" is a Durable Object handle, and "do" is a loop.
				if escapesAReservedWord(declaredName) {
					return
				}

				initializer := declaration.Initializer

				// A React context is read in element position as X.Provider.
				if isReactContextBinding(initializer) {
					return
				}
				// A component wrapped in memo or forwardRef; JSX needs its capital.
				if isReactComponentFactory(initializer) {
					return
				}
				// A component that says so in its annotation; JSX needs its capital just as much.
				if isReactComponentAnnotation(declaration) {
					return
				}
				// Mirrors an upstream constructor's own spelling.
				if isPascalCaseReBinding(initializer) {
					return
				}
				// consistency-require-type-suffix requires these to end in "Kind".
				if isCasingConstEnumShape(initializer) {
					return
				}

				usage := usageIndexFor(ctx)

				// Exported means a name crosses files, by the modifier or by a local export clause
				// naming this binding. `export default X` does not count: the importer picks its own
				// name, so the local spelling reaches no one.
				exportedNames := exportedNamesOf(ctx, statement, name)
				exported := len(exportedNames) > 0

				// Next.js reads this export off the route file by name, so its spelling and its
				// `export` are both the framework's. Every clause below would otherwise advise a
				// change Next silently ignores: PascalCase for `revalidate`, camelCase for an arrow
				// `POST`, or dropping the export from a `metadata` nothing imports. Checked before
				// the clauses rather than beside the consumer's list, because that list is optional
				// and this floor is not.
				if exported && nextjs.IsRouteContractExport(imports.NormalizedFileName(ctx.SourceFile), declaredName) {
					return
				}

				// A function is named for what it does, not for how far it reaches, so like an
				// instance it takes camelCase whatever its reach. Five shapes reach this: an inline
				// function, an alias for one, a name the file itself calls, a factory call whose
				// name says a function comes back, and a call to a local factory whose return type
				// says so. The last three exist because a function exported for other modules to
				// call has no local evidence at all, since every call site is elsewhere.
				//
				// Requiring the casing rather than merely permitting it is what makes this useful.
				// An exemption only means "do not flag", which leaves a PascalCase function sitting
				// wrong and unreported: LocaleMiddleware was exactly that, a function whose consumer
				// calls localeMiddleware(request) in another repository.
				if isFunctionValued(initializer) ||
					usage.isFunctionValuedIdentifier(initializer) ||
					usage.isCalledLikeAFunction(declaredName) ||
					isFactoryProducedFunction(initializer) ||
					usage.isLocallyDeclaredFunctionReturn(initializer) {
					if isCamelCase(declaredName) {
						return
					}
					// A component is a function too, and JSX needs its capital.
					if usage.isUsedAsJsxElementName(declaredName) {
						return
					}
					if shadowsCamelCaseSource(declaration) {
						return
					}
					reportCamelCase(ctx, exported, name, messageRequireCamelCaseFunction(
						declaredName, constantNameToCamelCase(declaredName),
					))
					return
				}

				// A camelCase const whose type is a function is exempt. The type needs no evidence in
				// this file, which is why it reaches a client made by an imported factory (#s3jynnx),
				// and it only permits camelCase, never requires it: it cannot see how the value is used
				// elsewhere, and Base's Verify* decorators and GraphQlQuery are callables from imported
				// factories that stay PascalCase because every use is `@VerifyEmail`. Requiring it
				// flagged 36 valid constants there. See holdsAFunctionType.
				if isCamelCase(declaredName) && holdsAFunctionType(ctx, name) {
					return
				}

				// An import wearing a const. The namespace rule owns this spelling.
				if isDynamicImportBinding(initializer) {
					return
				}

				// An instance takes camelCase whatever its reach, so it is judged here rather than
				// falling through to the exported/local split below. This is the clause where kind
				// outranks scope: PascalCase would read as the class rather than the object.
				if isClassInstance(ctx, initializer, usage) || hasClassTypeAnnotation(declaration) {
					if isCamelCase(declaredName) {
						return
					}
					reportCamelCase(ctx, exported, name, messageRequireCamelCaseInstance(
						declaredName, constantNameToCamelCase(declaredName),
					))
					return
				}

				if exported {
					if isPascalCase(declaredName) || frameworkConstantNames[declaredName] {
						return
					}
					if isCamelCase(declaredName) &&
						importerIndexFor(ctx.Program).provablyUnimported(ctx.SourceFile, exportedNames...) {
						ctx.ReportNode(name, messageDropUnimportedExport(
							declaredName, constantNameToPascalCase(declaredName),
						))
						return
					}
					ctx.ReportNode(name, messageRequirePascalCaseExported(
						declaredName, constantNameToPascalCase(declaredName),
					))
					return
				}

				if isCamelCase(declaredName) {
					return
				}
				// JSX owns the capital letter: <Icon /> is a component, <icon /> is an HTML tag.
				if usage.isUsedAsJsxElementName(declaredName) {
					return
				}
				// The capital is what keeps the declaration from referencing itself.
				if shadowsCamelCaseSource(declaration) {
					return
				}
				reportCamelCase(ctx, exported, name, messageRequireCamelCaseLocal(
					declaredName, constantNameToCamelCase(declaredName),
				))
			},
		}
	},
}

// reportCamelCase reports a constant that should be camelCase, with the rename fix when nothing
// outside the file can name it.
//
// The suggestion inside the message is the rename the fix performs, so the two cannot disagree:
// both come from `constantNameToCamelCase`.
func reportCamelCase(ctx rule.Context, exported bool, name *ast.Node, message rule.Message) {
	if !exported {
		if fix, isSafe := fileLocalRenameFix(ctx, name, constantNameToCamelCase(name.Text())); isSafe {
			ctx.ReportNodeWithFixes(name, message, fix)
			return
		}
	}
	ctx.ReportNode(name, message)
}

// hasDeclareModifier reports whether a statement carries the declare keyword.
func hasDeclareModifier(statement *ast.Node) bool {
	if statement == nil {
		return false
	}
	modifiers := statement.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == ast.KindDeclareKeyword {
			return true
		}
	}
	return false
}

// factoryNamePattern matches a callee whose name claims a function comes back.
//
// Creator and Factory only. The create and build prefixes were too eager: buildRegistry() returns a
// class instance and createLocaleMiddleware() returns a function, and the name alone cannot tell
// those apart. What can tell them apart is the declared return type, which the usage index reads
// when the factory lives in the same file.
var factoryNamePattern = regexp.MustCompile(`(Creator|Factory)$`)

// isFactoryProducedFunction reports whether the initializer calls a factory whose name says it
// returns a function.
//
// An "export const createEsLintRule = EsLintUtilities.RuleCreator(...)" holds a function, but
// nothing local proves it: the calls all live in other modules, so reading usage in this file finds
// none. Without type information the only remaining evidence is the factory's own name.
//
// This is a naming heuristic rather than a fact, which is why it is narrow. A factory that returns a
// value rather than a function is still judged on its casing, and the author can say so with a
// disable comment.
func isFactoryProducedFunction(initializer *ast.Node) bool {
	expression := unwrapAssertions(initializer)
	if expression == nil || expression.Kind != ast.KindCallExpression {
		return false
	}
	name := calleeName(expression.AsCallExpression().Expression)
	return name != "" && factoryNamePattern.MatchString(name)
}

// isClassInstance reports whether the constant holds an instance of a class written here.
//
// This is the "kind beats reach" case: the value is a thing that acts, so it takes camelCase no
// matter how far it travels. Three shapes reach it:
//
//   - the direct new NetworkService()
//   - the resolver resolveAgentManagerSingleton() whose declared return type is the class, because a
//     singleton routed through a function to survive module re-evaluation is still a singleton
//   - the cached singleton globalForThing.__thing ?? new Thing(), which is the same singleton
//     wearing the guard that keeps it alive across a hot reload
//
// The third shape was missing once, and its absence was a real defect rather than an omission of
// taste: tasksDatabase, artDatabase and discordClient are all class instances the rule intends to
// exempt, and all three were being told to become PascalCase. That is the rule contradicting its own
// third clause, and renaming them would have made the names lie about what they hold.
// A cast is deliberately not unwrapped here, which is the one place this rule reads an initializer
// raw. Measured on the ahra tree, unwrapping produced 95 findings in two generated GraphQL files,
// every one of the form:
//
//	export const AccountDocument = new TypedDocumentString(...) as unknown as TypedDocumentString<R, V>
//
// TypedDocumentString is a real class, so peeling the cast finds a real "new" and the instance
// clause fires correctly on its own terms. The original does not fire, because it reads node.init
// and a cast matches none of its branches, so these fall through and pass as PascalCase exports.
//
// The original is right and the peeled reading is wrong, for a reason the clause states itself: an
// instance is exempted from PascalCase because it is a thing that acts, and the call site reads
// accountDocument.method(). A value cast to something else is not being handed to the reader as the
// class it was constructed from; it is a document handed to a client. The cast is the author saying
// so, and this rule has no type information to overrule it with.
//
// Parity with the gate cohere replaces is the acceptance criterion, and 95 findings on a clean
// baseline is the shape a false-positive family takes.
func isClassInstance(ctx rule.Context, initializer *ast.Node, usage *fileUsageIndex) bool {
	expression := initializer
	if expression == nil {
		return false
	}

	// The cached-singleton form: globalForThing.__thing ?? new Thing().
	//
	// Either side may be the construction, since the guard can be written in either order, so both
	// are checked. Only ?? and || qualify: those choose between two values that should be the same
	// kind, whereas && yields the right operand only when the left is truthy and is not a singleton
	// guard.
	if expression.Kind == ast.KindBinaryExpression {
		binary := expression.AsBinaryExpression()
		if binary.OperatorToken == nil {
			return false
		}
		if binary.OperatorToken.Kind != ast.KindQuestionQuestionToken &&
			binary.OperatorToken.Kind != ast.KindBarBarToken {
			return false
		}
		return isClassInstance(ctx, binary.Left, usage) || isClassInstance(ctx, binary.Right, usage)
	}

	// The direct form: new NetworkService().
	if expression.Kind == ast.KindNewExpression {
		callee := unwrapAssertions(expression.AsNewExpression().Expression)
		if callee == nil || callee.Kind != ast.KindIdentifier {
			return false
		}
		return !dataStructureConstructors[callee.Text()] && !constructsAContainer(ctx, callee)
	}

	// The resolver form, read from the local factory's declared return type.
	return usage.isResolverReturningAClass(expression)
}

// shapeTypeSuffixes are the suffixes our conventions give to shapes rather than to classes.
var shapeTypeSuffixes = []string{"Type", "Interface", "Properties", "Options", "Kind", "Record"}

// hasClassTypeAnnotation reports whether the binding has a declared type that names a class.
//
// An "export const discordClient: Client = ..." and an "export const tasksDatabase:
// TasksSqliteDatabase = ..." both say what they hold right on the declaration, and the annotation is
// more reliable than anything the initializer can show, since the initializer may be a global
// lookup, a cache read, or a factory imported from another module.
func hasClassTypeAnnotation(declaration *ast.VariableDeclaration) bool {
	if declaration == nil || declaration.Type == nil || declaration.Type.Kind != ast.KindTypeReference {
		return false
	}
	reference := declaration.Type.AsTypeReferenceNode()
	if reference.TypeName == nil || reference.TypeName.Kind != ast.KindIdentifier {
		return false
	}

	// The annotation only proves a class when it names one bare. A generic container like
	// Record<K, V> is PascalCase too, and reading that as a class turned every exported lookup table
	// into a false report: 129 of them, all correct as PascalCase data.
	//
	// So require a bare type reference with no type arguments, which is what an instance annotation
	// looks like: ": Client", ": TasksSqliteDatabase".
	if reference.TypeArguments != nil && len(reference.TypeArguments.Nodes) > 0 {
		return false
	}

	// The annotation alone still cannot tell a class from a type alias for a plain object. A
	// "const CollapsibleTransition: Transition = { type: 'spring' }" names a Framer Motion shape,
	// not an instance. An object or array literal on the right is data whatever its annotation says,
	// so the initializer gets the final word.
	if initializer := unwrapAssertions(declaration.Initializer); initializer != nil {
		if initializer.Kind == ast.KindObjectLiteralExpression ||
			initializer.Kind == ast.KindArrayLiteralExpression {
			return false
		}
	}

	typeName := reference.TypeName.Text()
	if !startsUppercase(typeName) || dataStructureConstructors[typeName] {
		return false
	}
	for _, suffix := range shapeTypeSuffixes {
		if strings.HasSuffix(typeName, suffix) {
			return false
		}
	}
	return true
}

// shadowsCamelCaseSource reports a PascalCase local that reads a camelCase binding of the same name
// in its own initializer.
//
// A "const ProcessingIcon = processingIcon ?? BrokenCircleIcon" is the standard way a component
// property becomes a renderable local. The two names differ only in the first letter, and the
// capital is load bearing: rename it to camelCase and the declaration references itself.
func shadowsCamelCaseSource(declaration *ast.VariableDeclaration) bool {
	name := declaration.Name()
	if name == nil || name.Kind != ast.KindIdentifier || declaration.Initializer == nil {
		return false
	}

	camelCaseName := lowerFirstCharacter(name.Text())
	if camelCaseName == name.Text() {
		return false
	}

	found := false
	var walk func(*ast.Node)
	walk = func(current *ast.Node) {
		if current == nil || found {
			return
		}
		if current.Kind == ast.KindIdentifier && current.Text() == camelCaseName {
			found = true
			return
		}
		current.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return found
		})
	}
	walk(declaration.Initializer)
	return found
}

// holdsAFunctionType reports whether the constant's type says it holds a function: every part of it,
// once null and undefined are set aside, can be called and cannot be constructed.
//
// An "export const connectedGraphQlClient = createGraphQlClient({...})" holds a function, and nothing
// in its file proves it: the factory is imported and every call is in another module. This rule used
// to guess from the factory instead, by its name (a callee ending in Creator or Factory) or by its
// declared return type when it lived in the same file, and both guesses missed this line. The
// constant's own type is the fact both of them were approximating, so it replaces them: it answers
// for an imported factory, a local one, an annotation, and a cast alike.
//
// What the type must show, and why each part is there:
//
//   - A call signature. That is what "a function" means to the checker, and it is the only thing a
//     caller of `connectedGraphQlClient(query)` relies on.
//   - Properties do not disqualify it. createGraphQlClient's result is a callable that also carries
//     a `requestCapturingSetCookie` method, and it is still called by name 77 times; a function with
//     a property hung on it is a function. An object with methods and no call signature is not, and
//     stays a constant.
//   - No construct signature. A value that can be built with `new` is a constructor, which keeps the
//     capital every reader expects at `new Decimal(...)`, even when it can also be called. A class
//     has only a construct signature, so it never reaches this clause.
//   - Every part of a union. `string | (() => string)` is sometimes a string, so nothing about it is
//     reliably a function.
//   - No call that returns a React element. A component is a function too, and JSX needs its
//     capital, so it is not this clause's business; see `mayReturnAnElement`. The first version of
//     this check had no such part and told ten Structure components to become camelCase: seven
//     compound components built as `Object.assign(DialogRoot, { Trigger, Header })` and three
//     `React.memo(Row, compare)`.
//
// `any` and `unknown` have no signatures, so they prove nothing here and the syntactic and usage
// evidence beside this answers for them, as it did before the checker was asked.
//
// Without a checker (a syntax-only run) this answers false, and that evidence is all there is.
func holdsAFunctionType(ctx rule.Context, name *ast.Node) bool {
	if ctx.TypeChecker == nil {
		return false
	}
	valueType := ctx.TypeChecker.GetTypeAtLocation(name)
	if valueType == nil {
		return false
	}
	for part := range type_checking.UnionTypePartsSeq(ctx.TypeChecker.GetNonNullableType(valueType)) {
		callSignatures := ctx.TypeChecker.GetSignaturesOfType(part, checker.SignatureKindCall)
		if len(callSignatures) == 0 {
			return false
		}
		if len(ctx.TypeChecker.GetSignaturesOfType(part, checker.SignatureKindConstruct)) > 0 {
			return false
		}
		for _, signature := range callSignatures {
			if mayReturnAnElement(ctx.TypeChecker, checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)) {
				return false
			}
		}
	}
	return true
}

// elementMemberNames are the members of a React element: what `ReactElement` declares and what
// `JSX.Element` inherits.
var elementMemberNames = []string{"type", "props", "key"}

// mayReturnAnElement reports whether a call's return type is, or may be, a React element.
//
// Read from the members rather than from the name `ReactElement`, so it holds for `JSX.Element`,
// which extends it, for the `ReactNode` union a `FunctionComponent` or a `memo` result declares,
// and for any JSX library typed the same way, without asking where a declaration lives. It only
// ever withholds the function clause, so reading it broadly can cost a report and never adds one:
// the constant then falls through to be judged as data, which is what it was before the checker
// was asked.
//
// A return the checker cannot see, `any` or `unknown`, may be an element. That is what a
// component's return becomes when the JSX types do not resolve, and counting it as a plain function
// would tell every `Object.assign(DialogRoot, {...})` in such a file to become camelCase.
func mayReturnAnElement(typeChecker *checker.Checker, returnType *checker.Type) bool {
	if returnType == nil {
		return true
	}
	if type_checking.IsTypeFlagSet(returnType, checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
		return true
	}
	for part := range type_checking.UnionTypePartsSeq(returnType) {
		if !type_checking.IsObjectType(part) && !type_checking.IsIntersectionType(part) {
			continue
		}
		hasEveryMember := true
		for _, memberName := range elementMemberNames {
			if checker.Checker_getPropertyOfType(typeChecker, part, memberName) == nil {
				hasEveryMember = false
				break
			}
		}
		if hasEveryMember {
			return true
		}
	}
	return false
}

// constructsAContainer reports whether a `new X()` callee's class extends a built-in container,
// directly or through classes of its own: `class TypedDocumentString extends String` holds one
// operation's text, so it is data by the same reason Map is (#w26f1b0).
//
// The class is found through the checker, so an imported one counts, and without a checker through
// the class this file declares under that name. A class that cannot be found is not exempt, which is a
// missed exemption and never an invented finding.
func constructsAContainer(ctx rule.Context, callee *ast.Node) bool {
	class := constantCasingClassOf(ctx, callee)
	for depth := 0; class != nil && depth < 16; depth++ {
		base := ast.GetClassExtendsHeritageElement(class)
		if base == nil {
			return false
		}
		baseName := unwrapAssertions(base.AsExpressionWithTypeArguments().Expression)
		if baseName == nil || baseName.Kind != ast.KindIdentifier {
			return false
		}
		if dataStructureConstructors[baseName.Text()] {
			return true
		}
		class = constantCasingClassOf(ctx, baseName)
	}
	return false
}

// constantCasingClassOf returns the class declaration an identifier names, or nil.
func constantCasingClassOf(ctx rule.Context, name *ast.Node) *ast.Node {
	if ctx.TypeChecker != nil {
		symbol := ctx.TypeChecker.GetSymbolAtLocation(name)
		if symbol != nil {
			symbol = checker.SkipAlias(symbol, ctx.TypeChecker)
			for _, declaration := range symbol.Declarations {
				if declaration.Kind == ast.KindClassDeclaration || declaration.Kind == ast.KindClassExpression {
					return declaration
				}
			}
		}
		return nil
	}
	var found *ast.Node
	ctx.SourceFile.AsNode().ForEachChild(func(statement *ast.Node) bool {
		if statement.Kind == ast.KindClassDeclaration && statement.Name() != nil && statement.Name().Text() == name.Text() {
			found = statement
			return true
		}
		return false
	})
	return found
}

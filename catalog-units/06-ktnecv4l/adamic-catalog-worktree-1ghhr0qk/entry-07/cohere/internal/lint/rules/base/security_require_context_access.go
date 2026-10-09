package base

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/decorators"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// securityRequireContextAccessText is the rule's message, whose wording lives in
// `policy/messages/security-require-context-access.json`.
var securityRequireContextAccessText = policy.MessageOf("base/security-require-context-access", "missingProtector")

// SecurityRequireContextAccessRequirement pairs a protected context key with the decorators that unlock it.
type SecurityRequireContextAccessRequirement struct {
	// ContextKey is the name of the injected key that needs protecting, as it is EXPORTED rather
	// than as it is spelled at the injection site. An aliased import is resolved back to this name.
	ContextKey string `json:"contextKey"`

	// RequiresAny is the set of decorators any one of which establishes the access. Empty means the
	// requirement can never be satisfied, which the decoder rejects.
	RequiresAny []string `json:"requiresAny"`
}

// SecurityRequireContextAccessOptions configures the rule.
//
// The shape is the source's own. The judgment is generic -- `@InjectRequestContext` and
// `@GraphQlFieldResolver` are base concepts every consumer has -- while WHICH keys are protected and
// WHICH decorators unlock them is project configuration, so the rule ships knowing neither.
type SecurityRequireContextAccessOptions struct {
	// Requirements is the list. Empty means the rule has nothing to enforce and registers nothing,
	// which is the source's `if(requirementMap.size === 0) return {}`.
	Requirements []SecurityRequireContextAccessRequirement `json:"requirements"`
}

// DecodeContextRequiresAccessOptions reads this rule's configuration from the config layer.
//
// Hand-rolled to enforce the two constraints the source states in its JSON schema, which we have no
// schema layer for. `contextKey` and `requiresAny` are both `required`, and `requiresAny` carries
// `minItems: 1` -- an empty list would be a requirement no decorator can ever satisfy, which reports
// on every injection of that key with a message naming nothing to add.
func DecodeContextRequiresAccessOptions(raw []byte) (any, error) {
	var options SecurityRequireContextAccessOptions
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, err
	}
	for index, requirement := range options.Requirements {
		if requirement.ContextKey == "" {
			return options, fmt.Errorf(
				"base/security-require-context-access: requirement %d names no contextKey", index)
		}
		if len(requirement.RequiresAny) == 0 {
			return options, fmt.Errorf(
				"base/security-require-context-access: requirement %d for %q lists no decorator in "+
					"requiresAny, so nothing could ever satisfy it", index, requirement.ContextKey)
		}
	}
	return options, nil
}

// securityRequireContextAccessDownstreamResolvers are the decorators that exempt a method entirely.
//
// A field resolver runs only after something produced the parent object, and that something already
// established access. Checking again at the field would report every resolver in the codebase.
var securityRequireContextAccessDownstreamResolvers = map[string]struct{}{
	"GraphQlFieldResolver": {},
}

// SecurityRequireContextAccess reports a method that injects a protected request-context key without
// carrying a decorator that establishes access to it.
//
//	valid:   @RequireSessionAccess() m(@InjectRequestContext(AccountRequestContextKey) a: string) {}
//	valid:   m(@InjectRequestContext(AccountRequestContextKey) a?: string) {}
//	valid:   @GraphQlFieldResolver() m(@InjectRequestContext(AccountRequestContextKey) a: string) {}
//	invalid: m(@InjectRequestContext(AccountRequestContextKey) a: string) {}
//
// # What it protects against
//
// `@InjectRequestContext(K)` hands a method a value out of the request context. For an unprotected
// key that value arrives whether or not anyone verified the request, so a method acting on it is
// making an authorization decision on unverified input. The access decorators are what perform that
// verification, and this rule is the check that the two travel together.
//
// The rule is generic and its configuration is not: the key names and the decorators that unlock
// them belong to the project, so the rule enforces nothing until somebody supplies them.
//
// # Ported from api-phi-health rather than from an upstream linter
//
// `libraries/base/code-quality/lint/rules/SecurityRequireContextAccessRule.ts`. There is no upstream corpus
// to import, so every fixture here was written against a measured verdict: the source rule was
// loaded into the ESLint Linter with the live wiring from `BaseLintConfiguration.ts` and driven over
// forty-six inputs, and each fixture records what it answered. Three of those inputs turned out to
// be parse errors rather than cases, and two answered the opposite of what the source reads like.
//
// # Where our AST differs from the one the source was written against
//
//	a constructor          ESTree calls it a `MethodDefinition`, so one visitor covers it. Our
//	                       parser gives it `KindConstructor`, so it needs its own listener -- and a
//	                       constructor is exactly where a parameter property is injected, which is
//	                       the shape this rule most needs to see.
//	decorators             live in the MODIFIER list rather than in a `.decorators` field, which is
//	                       what `decorators.Of` exists to hide.
//	a parameter property   is an ordinary parameter carrying `private`/`readonly` modifiers rather
//	                       than a `TSParameterProperty` wrapper, so no unwrapping step is needed.
var SecurityRequireContextAccess = rule.Rule{
	Name: "base/security-require-context-access",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, configured := rule.OptionsAs[SecurityRequireContextAccessOptions](options)
		if !configured || len(settings.Requirements) == 0 {
			// Nothing configured means nothing to enforce, and the source returns an empty visitor
			// object rather than a no-op listener. Reproduced so the rule costs nothing at all in a
			// project that has not wired it.
			return rule.Listeners{}
		}

		required := make(map[string][]string, len(settings.Requirements))
		for _, requirement := range settings.Requirements {
			required[requirement.ContextKey] = requirement.RequiresAny
		}

		// Import aliases are resolved per file, and the source relies on import declarations being
		// visited before the methods that use them. Our walk is pre-order over the whole file from
		// one listener, so the aliases are gathered in a first pass instead of relying on order.
		return rule.Listeners{
			ast.KindSourceFile: func(file *ast.Node) {
				aliases := securityRequireContextAccessImportAliases(file)
				securityRequireContextAccessWalk(ctx, file, required, aliases)
			},
		}
	},
}

// securityRequireContextAccessWalk visits every method and constructor in the file.
//
// A single walk from the source file rather than two listeners, because the alias map has to be
// complete before any method is judged and there is no exit hook to order that with.
func securityRequireContextAccessWalk(
	ctx rule.Context,
	node *ast.Node,
	required map[string][]string,
	aliases map[string]string,
) {
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindMethodDeclaration, ast.KindConstructor,
		ast.KindGetAccessor, ast.KindSetAccessor:
		// The accessors are included because the source's `MethodDefinition` visitor covers them in
		// ESTree, where a getter and a setter are method definitions with a `kind`. Measured: a
		// setter with a decorated parameter reports.
		checkContextRequiresAccess(ctx, node, required, aliases)
	}
	node.ForEachChild(func(child *ast.Node) bool {
		securityRequireContextAccessWalk(ctx, child, required, aliases)
		return false
	})
}

// checkContextRequiresAccess judges one method or constructor.
func checkContextRequiresAccess(
	ctx rule.Context,
	node *ast.Node,
	required map[string][]string,
	aliases map[string]string,
) {
	// A downstream field resolver is exempt before anything else is asked, matching the source.
	if decorators.HasDecoratorInSet(node, securityRequireContextAccessDownstreamResolvers) {
		return
	}

	// The keys this method injects that are configured as protected, in the order the parameters
	// appear, with duplicates collapsed. A method injecting one key twice reports once.
	var injected []string
	seen := map[string]struct{}{}
	for _, parameter := range node.Parameters() {
		for _, decorator := range decorators.Of(parameter) {
			surfaceName := securityRequireContextAccessInjectedKey(decorator)
			if surfaceName == "" {
				continue
			}
			key := surfaceName
			if exported, aliased := aliases[surfaceName]; aliased {
				key = exported
			}
			if _, protected := required[key]; !protected {
				continue
			}
			// A parameter that admits a missing value is "soft use": the author has committed to
			// handling the absent case inline, so the access requirement is skipped.
			if securityRequireContextAccessParameterIsSoftUse(parameter) {
				continue
			}
			if _, already := seen[key]; already {
				continue
			}
			seen[key] = struct{}{}
			injected = append(injected, key)
		}
	}
	if len(injected) == 0 {
		return
	}

	present := securityRequireContextAccessDecoratorNames(node)

	for _, key := range injected {
		requiresAny := required[key]
		satisfied := false
		for _, decoratorName := range requiresAny {
			if _, found := present[decoratorName]; found {
				satisfied = true
				break
			}
		}
		if satisfied {
			continue
		}

		wanted := make([]string, 0, len(requiresAny))
		for _, decoratorName := range requiresAny {
			wanted = append(wanted, "`@"+decoratorName+"()`")
		}
		ctx.ReportNode(securityRequireContextAccessReportTarget(node), rule.Message{
			Id: securityRequireContextAccessText.Id,
			Description: securityRequireContextAccessText.Render(map[string]string{
				"contextKey": key,
				"protectors": strings.Join(wanted, " or "),
			}),
		})
	}
}

// securityRequireContextAccessReportTarget returns the node a finding points at.
//
// The source reports on `node.key`, the method's NAME, rather than on the whole method. That keeps
// the caret on the thing a reader searches for and off the decorators, and it matters here because
// a method with a long parameter list would otherwise underline the whole signature.
//
// A constructor has no name node, so the finding falls back to the constructor itself. That is a
// shape the source never meets, because ESTree gives a constructor a `key` naming it.
func securityRequireContextAccessReportTarget(node *ast.Node) *ast.Node {
	if name := node.Name(); name != nil {
		return name
	}
	return node
}

// securityRequireContextAccessImportAliases maps a local name back to the name it was exported under.
//
// `import { AccountRequestContextKey as AK }` then `@InjectRequestContext(AK)` has to be matched
// against the configured key rather than silently bypassing it.
//
// The map is keyed by LOCAL name, which has a consequence worth stating because it looks like a
// bug and is the source's behaviour: `import { Something as AccountRequestContextKey }` maps the
// configured name AWAY to `Something`, so injecting it is CLEAN. Measured against the source rule.
// That is right rather than unfortunate -- the local name is a different symbol that happens to
// share a spelling, and protecting it would be protecting the wrong thing.
func securityRequireContextAccessImportAliases(file *ast.Node) map[string]string {
	aliases := map[string]string{}
	file.ForEachChild(func(statement *ast.Node) bool {
		if statement.Kind != ast.KindImportDeclaration {
			return false
		}
		// `imports.BindingsOf` decides which of the three shapes one import statement carries, so
		// the named list arrives without this rule reaching past it into the clause. Only the named
		// bindings can be aliased in the way this rule cares about: a default or namespace import
		// binds one local name with no exported name to resolve it back to.
		for _, element := range imports.BindingsOf(statement).Named {
			specifier := element.AsImportSpecifier()
			// `PropertyName` is set only when the import is aliased, and it holds the name the
			// symbol was exported under. Without an alias the local name IS the exported name and
			// no entry is needed.
			if specifier.PropertyName == nil {
				continue
			}
			local := specifier.Name()
			if local == nil || specifier.PropertyName.Kind != ast.KindIdentifier {
				continue
			}
			aliases[local.Text()] = specifier.PropertyName.Text()
		}
		return false
	})
	return aliases
}

// securityRequireContextAccessInjectedKey reads the key name out of an `@InjectRequestContext(X)`
// decorator.
//
// Two argument shapes resolve and everything else answers nothing. A bare identifier answers itself.
// A non-computed member access answers its PROPERTY, so `@InjectRequestContext(Keys.AccountKey)`
// matches a requirement configured for `AccountKey` -- without that arm a namespaced key silently
// bypasses the requirement, which the source's own comment records as a defect it fixed.
//
// A computed access answers nothing, because the property it names is not knowable before it runs.
func securityRequireContextAccessInjectedKey(decorator *ast.Node) string {
	if decorator == nil || decorator.Kind != ast.KindDecorator {
		return ""
	}
	expression := decorator.AsDecorator().Expression
	if expression == nil || expression.Kind != ast.KindCallExpression {
		return ""
	}
	call := expression.AsCallExpression()
	callee := call.Expression
	if callee == nil || !ast.IsIdentifier(callee) || callee.Text() != "InjectRequestContext" {
		return ""
	}
	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return ""
	}
	argument := call.Arguments.Nodes[0]
	if argument == nil {
		return ""
	}
	if ast.IsIdentifier(argument) {
		return argument.Text()
	}
	if argument.Kind == ast.KindPropertyAccessExpression {
		name := argument.AsPropertyAccessExpression().Name()
		if name != nil && ast.IsIdentifier(name) {
			return name.Text()
		}
	}
	return ""
}

// securityRequireContextAccessDecoratorNames collects the decorator names on a member and on its class.
//
// An access decorator may sit on either, so both are gathered into one set.
//
// # Why this is not `decorators.CallName`
//
// The shelf helper accepts a bare identifier CALL only, which is exactly right for the
// `GraphQlFieldResolver` exemption above and wrong here. The source uses a different helper for this
// question, one that also accepts a decorator written WITHOUT parentheses: measured against the
// source rule, `@RequireSessionAccess` with no call protects just as `@RequireSessionAccess()` does.
//
// Both helpers agree on the other direction: a qualified `@Access.RequireSessionAccess()` does NOT
// protect, because the configured name is a bare identifier and a namespaced one is a different
// symbol that happens to end in the same word. Also measured.
func securityRequireContextAccessDecoratorNames(node *ast.Node) map[string]struct{} {
	names := map[string]struct{}{}
	securityRequireContextAccessCollectNames(node, names)
	if class := securityRequireContextAccessEnclosingClass(node); class != nil {
		securityRequireContextAccessCollectNames(class, names)
	}
	return names
}

// securityRequireContextAccessCollectNames adds a node's decorator names to a set.
func securityRequireContextAccessCollectNames(node *ast.Node, into map[string]struct{}) {
	for _, decorator := range decorators.Of(node) {
		if name := securityRequireContextAccessDecoratorName(decorator); name != "" {
			into[name] = struct{}{}
		}
	}
}

// securityRequireContextAccessDecoratorName reads a decorator's name, call or not.
//
// The source's `getDecoratorName` with its default options: a bare identifier call answers its
// callee, a bare identifier with no call answers itself, and a qualified name answers nothing
// because `resolveMemberExpression` defaults to false.
func securityRequireContextAccessDecoratorName(decorator *ast.Node) string {
	if decorator == nil || decorator.Kind != ast.KindDecorator {
		return ""
	}
	expression := decorator.AsDecorator().Expression
	if expression == nil {
		return ""
	}
	if expression.Kind == ast.KindCallExpression {
		callee := expression.AsCallExpression().Expression
		if callee != nil && ast.IsIdentifier(callee) {
			return callee.Text()
		}
		return ""
	}
	if ast.IsIdentifier(expression) {
		return expression.Text()
	}
	return ""
}

// securityRequireContextAccessEnclosingClass walks up to the class a member belongs to.
//
// The source reaches it as `methodNode.parent.parent`, which is the class BODY then the class.
// typescript-go hangs members directly off the class, so this is one step -- written as a walk
// rather than a single parent read so it stays correct if the parser ever inserts a body node.
func securityRequireContextAccessEnclosingClass(node *ast.Node) *ast.Node {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression:
			return parent
		case ast.KindSourceFile:
			return nil
		}
	}
	return nil
}

// securityRequireContextAccessParameterIsSoftUse answers whether a parameter admits a missing value.
//
// Two ways: the parameter is optional (`?:`), or its annotation is or unions in a type that can be
// absent. Either way the author has committed to handling the missing case, so the access
// requirement is skipped.
//
// A parameter with NO annotation at all is not soft use, which is worth stating because it reads the
// other way round: an unannotated parameter is implicitly `any`, and `any` IS soft use when written
// out. The source tests the annotation and answers false when there is none, so the implicit case
// reports. Measured, and reproduced rather than tidied.
func securityRequireContextAccessParameterIsSoftUse(parameter *ast.Node) bool {
	declaration := parameter.AsParameterDeclaration()
	if declaration.QuestionToken != nil {
		return true
	}
	return securityRequireContextAccessTypeAdmitsMissing(declaration.Type)
}

// securityRequireContextAccessTypeAdmitsMissing answers whether a type annotation can be absent.
//
// Syntactic rather than type-driven, matching the source, which reads AST type nodes rather than
// asking the checker. So an alias for `string | undefined` does NOT count, because the annotation
// this sees is an identifier.
func securityRequireContextAccessTypeAdmitsMissing(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindUndefinedKeyword, ast.KindNullKeyword, ast.KindAnyKeyword,
		ast.KindUnknownKeyword, ast.KindVoidKeyword:
		return true
	case ast.KindLiteralType:
		// `null` parses as a literal type wrapping the null keyword rather than as a bare keyword.
		return securityRequireContextAccessTypeAdmitsMissing(node.AsLiteralTypeNode().Literal)
	case ast.KindParenthesizedType:
		return securityRequireContextAccessTypeAdmitsMissing(node.AsParenthesizedTypeNode().Type)
	case ast.KindUnionType:
		for _, member := range node.AsUnionTypeNode().Types.Nodes {
			if securityRequireContextAccessTypeAdmitsMissing(member) {
				return true
			}
		}
	}
	return false
}

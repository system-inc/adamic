package nexus

import (
	"strings"
	"unicode"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/module"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// ConsistencyNoScreamingSnakeCaseOptions names constants that may keep the shouting form because
// they mirror an external system's grammar. Exact match, never a prefix.
type ConsistencyNoScreamingSnakeCaseOptions struct {
	Allow []string `json:"allow"`
}

// The rule's messages, whose wording lives in `policy/messages/consistency-no-screaming-snake-case.json`.
// The reasoning is identical for both, so it is written once there as the `reasoning` phrase. A rule
// that says only what is wrong gets disabled the first time it is inconvenient.
var (
	consistencyNoScreamingSnakeCaseExportedText = policy.MessageOf("nexus/consistency-no-screaming-snake-case", "noScreamingSnakeCaseExported")
	consistencyNoScreamingSnakeCaseLocalText    = policy.MessageOf("nexus/consistency-no-screaming-snake-case", "noScreamingSnakeCaseLocal")
)

func messageScreamingSnakeCaseExported(name string, suggestion string) rule.Message {
	return rule.Message{
		Id:          consistencyNoScreamingSnakeCaseExportedText.Id,
		Description: consistencyNoScreamingSnakeCaseExportedText.Render(map[string]string{"name": name, "suggestion": suggestion}),
	}
}

func messageScreamingSnakeCaseLocal(name string, suggestion string) rule.Message {
	return rule.Message{
		Id:          consistencyNoScreamingSnakeCaseLocalText.Id,
		Description: consistencyNoScreamingSnakeCaseLocalText.Render(map[string]string{"name": name, "suggestion": suggestion}),
	}
}

// ConsistencyNoScreamingSnakeCase bans SCREAMING_SNAKE_CASE constant names in our own code.
//
//	valid:   const orderColumns = [...]
//	valid:   export const OrderColumns = [...]
//	valid:   const DATABASE_URL = process.env.DATABASE_URL
//	invalid: const MAX_RETRY_COUNT = 3
//	invalid: export const HTTP_TIMEOUT = 5000
//
// The exemptions are the whole design. SCREAMING_SNAKE is right exactly when the value mirrors an
// external grammar, because then the casing stops being decorative and starts doing real work. Two
// ways to say so: read it from `process.env` directly, where the value is the environment variable
// by definition, or name it in the allow option.
//
// No fix. The suggested name is offered in the message rather than applied, because the right name
// depends on whether the value is exported and may collide with a binding already in scope, and
// because MAX_RETRY_COUNT to maximumRetryCount is a rename the casing conversion cannot do.
var ConsistencyNoScreamingSnakeCase = rule.Rule{
	Name: "nexus/consistency-no-screaming-snake-case",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		allowed := map[string]bool{}
		if settings, hasSettings := rule.OptionsAs[ConsistencyNoScreamingSnakeCaseOptions](options); hasSettings {
			for _, name := range settings.Allow {
				allowed[name] = true
			}
		}

		return rule.Listeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				declaration := node.AsVariableDeclaration()
				if declaration == nil {
					return
				}
				name := declaration.Name()
				if name == nil || name.Kind != ast.KindIdentifier {
					return
				}
				if !isScreamingSnakeCase(name.Text()) || allowed[name.Text()] {
					return
				}
				if isProcessEnvironmentAccess(declaration.Initializer) {
					return
				}

				// Only const declarations. A shouting `let` is exotic enough to ignore, and the
				// immutability premise the reasoning rests on does not apply to it.
				declarationList := node.Parent
				if declarationList == nil || declarationList.Flags&ast.NodeFlagsConst == 0 {
					return
				}

				if module.IsExported(declarationList.Parent) {
					ctx.ReportNode(name, messageScreamingSnakeCaseExported(name.Text(), toPascalCase(name.Text())))
					return
				}
				ctx.ReportNode(name, messageScreamingSnakeCaseLocal(name.Text(), toCamelCase(name.Text())))
			},
		}
	},
}

// isScreamingSnakeCase reports the shouting shape: uppercase letters, digits, and underscores, with
// at least one underscore.
//
// The underscore is required on purpose. A single all-caps word like RED is a different naming
// choice, and flagging it here would put this rule in an argument it was not written to have.
func isScreamingSnakeCase(name string) bool {
	if !strings.Contains(name, "_") {
		return false
	}
	for index, character := range name {
		isUppercase := character >= 'A' && character <= 'Z'
		isDigit := character >= '0' && character <= '9'

		if index == 0 && !isUppercase {
			return false
		}
		if !isUppercase && !isDigit && character != '_' {
			return false
		}
	}
	// A trailing underscore means the name is not the shape either, matching the original's anchor.
	return !strings.HasSuffix(name, "_")
}

// isProcessEnvironmentAccess reports whether an initializer reads from process.env.
//
// One level of fallback is unwrapped, so `process.env.FOO ?? 'default'` still counts. That pattern
// is how an environment read is almost always written, and a rule that missed it would flag exactly
// the names it means to exempt.
func isProcessEnvironmentAccess(initializer *ast.Node) bool {
	if initializer == nil {
		return false
	}

	current := initializer
	if current.Kind == ast.KindBinaryExpression {
		binary := current.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil {
			operator := binary.OperatorToken.Kind
			if operator == ast.KindQuestionQuestionToken || operator == ast.KindBarBarToken {
				current = binary.Left
			}
		}
	}

	if current == nil || current.Kind != ast.KindPropertyAccessExpression {
		return false
	}

	// Walk leftward looking for an `env` property on a chain rooted at `process`. Both halves
	// matter: a chain rooted at something else is unrelated, and `process.argv` is not an
	// environment read even though it is rooted correctly.
	rootIsProcess := false
	sawEnvironment := false
	for cursor := current; cursor != nil; {
		if cursor.Kind == ast.KindPropertyAccessExpression {
			access := cursor.AsPropertyAccessExpression()
			if access == nil {
				break
			}
			if propertyName := access.Name(); propertyName != nil && propertyName.Text() == "env" {
				sawEnvironment = true
			}
			cursor = access.Expression
			continue
		}
		if cursor.Kind == ast.KindIdentifier && cursor.Text() == "process" {
			rootIsProcess = true
		}
		break
	}
	return rootIsProcess && sawEnvironment
}

// screamingSnakeParts splits a shouting name into its non-empty words.
func screamingSnakeParts(name string) []string {
	parts := []string{}
	for _, part := range strings.Split(name, "_") {
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

// titleCase lowercases a word and raises its first letter.
func titleCase(part string) string {
	runes := []rune(strings.ToLower(part))
	if len(runes) == 0 {
		return ""
	}
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// toCamelCase converts a shouting name to camelCase. This handles the casing only, never the
// abbreviations, so MAX_RETRY_COUNT becomes maxRetryCount rather than maximumRetryCount.
func toCamelCase(name string) string {
	parts := screamingSnakeParts(name)
	if len(parts) == 0 {
		return name
	}

	var builder strings.Builder
	builder.WriteString(strings.ToLower(parts[0]))
	for _, part := range parts[1:] {
		builder.WriteString(titleCase(part))
	}
	return builder.String()
}

// toPascalCase converts a shouting name to PascalCase.
func toPascalCase(name string) string {
	var builder strings.Builder
	for _, part := range screamingSnakeParts(name) {
		builder.WriteString(titleCase(part))
	}
	return builder.String()
}

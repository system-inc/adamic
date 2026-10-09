package structure

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// The rule's messages, one handle per id, whose wording lives in
// `policy/messages/next-require-api-parameter-name.json`.
var (
	nextRequireApiParameterNameUseParamsNotParametersText             = policy.MessageOf("structure/next-require-api-parameter-name", "useParamsNotParameters")
	nextRequireApiParameterNameUseSearchParamsNotSearchParametersText = policy.MessageOf("structure/next-require-api-parameter-name", "useSearchParamsNotSearchParameters")
)

// messageUseParamsNotParameters is the finding, rendered when it is reported so the text comes from the current catalog.
func messageUseParamsNotParameters() rule.Message {
	return rule.Message{
		Id:          nextRequireApiParameterNameUseParamsNotParametersText.Id,
		Description: nextRequireApiParameterNameUseParamsNotParametersText.Render(nil),
	}
}

// messageUseSearchParamsNotSearchParameters is the finding, rendered when it is reported so the text comes from the current catalog.
func messageUseSearchParamsNotSearchParameters() rule.Message {
	return rule.Message{
		Id:          nextRequireApiParameterNameUseSearchParamsNotSearchParametersText.Id,
		Description: nextRequireApiParameterNameUseSearchParamsNotSearchParametersText.Render(nil),
	}
}

// nextApiFunctionNames are the framework-called functions whose argument shape Next dictates.
//
// A closed set rather than a prefix test on `generate`. A page's own `generateReport` takes
// whatever its author decides, and holding it to Next's contract would rename a field the
// framework never reads.
var nextApiFunctionNames = map[string]bool{
	"generateMetadata":     true,
	"generateStaticParams": true,
	"generateViewport":     true,
}

// NextRequireApiParameterName flags a Next API function whose argument fields use the house
// spelling rather than the framework's.
//
//	valid:   export async function generateMetadata({ params }: { params: Promise<Parameters> })
//	invalid: export async function generateMetadata({ parameters }: { parameters: Promise<...> })
//	invalid: export async function generateMetadata({ searchParameters }: { searchParameters: ... })
//
// This rule exists because the house convention and the framework disagree, and the framework has
// to win. Everywhere else in this codebase `parameters` is the correct spelling and abbreviations
// are the defect; here Next reads the field by name, so the abbreviation is the contract and the
// full word silently produces undefined.
//
// Only the type annotation's members are judged, not the destructuring pattern. The binding name is
// the author's to choose and Next never sees it; the field name in the object Next constructs is
// what has to match. Renaming a binding would be a different rule.
//
// Fixable, since the repair is exactly one identifier and there is only one right answer.
var NextRequireApiParameterName = rule.Rule{
	Name: "structure/next-require-api-parameter-name",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		fileContext := FileContextFor(ctx.SourceFile.FileName().AsString())
		if !fileContext.IsPageFile && !fileContext.IsLayoutFile {
			// Only the files Next calls these functions in. A helper module exporting a function
			// named generateMetadata is not one the framework will ever call.
			return nil
		}

		return rule.Listeners{
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				declaration := node.AsFunctionDeclaration()

				name := declaration.Name()
				if name == nil || !nextApiFunctionNames[name.Text()] {
					return
				}

				parameters := parameterNodes(declaration.Parameters)
				if len(parameters) == 0 {
					return
				}

				// The first argument only. Next passes one object and these functions take no
				// others, so a second parameter is somebody's own and not the framework's.
				typeNode := parameterTypeNode(parameters[0])
				if typeNode == nil || typeNode.Kind != ast.KindTypeLiteral {
					// A named type reference is not judged. The mistake would live in that type's
					// own declaration, which is somewhere else and may be shared, so reporting here
					// would point at the wrong file.
					return
				}

				for _, member := range typeNode.AsTypeLiteralNode().Members.Nodes {
					if member.Kind != ast.KindPropertySignature {
						continue
					}
					key := member.AsPropertySignatureDeclaration().Name()
					if key == nil || key.Kind != ast.KindIdentifier {
						continue
					}

					switch key.Text() {
					case "parameters":
						ctx.ReportNodeWithFixes(key, messageUseParamsNotParameters(),
							ctx.ReplaceNode(key, "params"))
					case "searchParameters":
						ctx.ReportNodeWithFixes(key, messageUseSearchParamsNotSearchParameters(),
							ctx.ReplaceNode(key, "searchParams"))
					}
				}
			},
		}
	},
}

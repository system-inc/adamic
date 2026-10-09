package structure

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
	"github.com/system-inc/cohere/internal/lint/ecmascript/module"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// messagePageRequireDefaultExportText is the rule's `pageRequireDefaultExport` message, whose
// wording lives in `policy/messages/next-require-page-default-export.json`.
var messagePageRequireDefaultExportText = policy.MessageOf("structure/next-require-page-default-export", "pageRequireDefaultExport")

// messagePageRequireDefaultExport is the finding, rendered when it is reported so the text comes
// from the current catalog.
func messagePageRequireDefaultExport() rule.Message {
	return rule.Message{Id: messagePageRequireDefaultExportText.Id, Description: messagePageRequireDefaultExportText.Render(nil)}
}

// messagePageDefaultExportInlineText is the rule's `pageDefaultExportInline` message, whose wording
// lives in `policy/messages/next-require-page-default-export.json`.
var messagePageDefaultExportInlineText = policy.MessageOf("structure/next-require-page-default-export", "pageDefaultExportInline")

// messagePageDefaultExportInline is the finding, rendered when it is reported so the text comes
// from the current catalog.
func messagePageDefaultExportInline() rule.Message {
	return rule.Message{Id: messagePageDefaultExportInlineText.Id, Description: messagePageDefaultExportInlineText.Render(nil)}
}

// messagePageDefaultExportNameSuffixText is the rule's `pageDefaultExportNameSuffix` message, whose
// wording lives in `policy/messages/next-require-page-default-export.json`.
var messagePageDefaultExportNameSuffixText = policy.MessageOf("structure/next-require-page-default-export", "pageDefaultExportNameSuffix")

// messagePageDefaultExportNameSuffix is the finding, rendered when it is reported so the text comes
// from the current catalog.
func messagePageDefaultExportNameSuffix() rule.Message {
	return rule.Message{Id: messagePageDefaultExportNameSuffixText.Id, Description: messagePageDefaultExportNameSuffixText.Render(nil)}
}

// NextRequirePageDefaultExport flags a Next.js page whose default export is missing or misshapen.
//
//	valid:   export default function AccountSettingsPageRoute() { ... }
//	invalid: (a page.tsx with no default export at all)
//	invalid: function AccountSettingsPageRoute() { ... } export default AccountSettingsPageRoute
//	invalid: export default function AccountSettings() { ... }
//
// Three message ids for three repairs: adding an export, moving it onto the declaration, and
// renaming the function.
//
// Only page.tsx and page.jsx. The convention is about the file Next resolves a route from, and a
// layout or an error boundary has its own contract that this one would misdescribe.
//
// The rule applies to any page file, including one with a default export that is not a function at
// all, which is why the missing-export check runs at the file rather than per export.
var NextRequirePageDefaultExport = rule.Rule{
	Name:       "structure/next-require-page-default-export",
	NoListener: rule.NoListenerAnswersInRun,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		fileContext := FileContextFor(ctx.SourceFile.FileName().AsString())
		if !fileContext.IsReactFile || !fileContext.IsPageFile {
			return nil
		}

		sourceFile := ctx.SourceFile.AsNode()
		hasDefaultExport := false

		// Imports bind at module scope wherever they are written, so they are gathered before any
		// export is judged: an export of an imported binding asks nothing of this file.
		imported := importedLocalNames(sourceFile)

		// The whole file is walked once here rather than through a listener, because the
		// missing-export finding is a claim about the file and needs every statement seen before it
		// can be made. A listener would have to report at the last node and guess which that is.
		sourceFile.ForEachChild(func(statement *ast.Node) bool {
			defaultExport, ok := pageDefaultExportOf(statement, imported)
			if !ok {
				return false
			}
			hasDefaultExport = true

			if defaultExport.separate {
				ctx.ReportNode(statement, messagePageDefaultExportInline())
			}
			if defaultExport.name != "" && !strings.HasSuffix(defaultExport.name, "PageRoute") {
				ctx.ReportNode(statement, messagePageDefaultExportNameSuffix())
			}
			return false
		})

		if !hasDefaultExport {
			ctx.ReportNode(sourceFile, messagePageRequireDefaultExport())
		}

		return nil
	},
}

// pageDefaultExport is what one default export asks of the file that writes it.
type pageDefaultExport struct {
	// name is the function name this file chose for the route, or empty when the file chose none:
	// an anonymous declaration, an expression, or a binding declared in another module.
	name string

	// separate is whether the export is a statement apart from a declaration in this file, which is
	// the one case where moving `export default` onto the declaration is an edit this file can make.
	separate bool
}

// pageDefaultExportOf reads a statement's default export, if it is one.
//
// Every form Next resolves a route from counts, because the missing-export finding is a claim that
// the route does not exist and each of these makes it exist:
//
//	export default function NamePageRoute() {}        on the declaration
//	export default NamePageRoute                       a separate statement
//	export { NamePageRoute as default }                the same, spelled as a specifier
//	export { NamePage as default } from './NamePage'   a re-export
//	export { default } from './other/page'             another page's default, re-exported
//
// The re-exports were missed, and www-phi-health has 78 page files that shim a Structure page that
// way, every one reported as having no default export. A type-only specifier carries no value, so
// Next finds no page in it and neither does this.
//
// Only a binding declared in this file is held to the inline and suffix conventions. A re-export
// or an imported binding has its declaration in another module, so there is no declaration here to
// carry the export and no name here to choose; reporting either would demand an edit this file
// cannot make. An expression such as `memo(NamePageRoute)` has neither a declaration nor a name.
// ESLint's original agrees on the expression and the re-exports; on an imported binding it reports
// both, which is the false positive this declines.
func pageDefaultExportOf(statement *ast.Node, imported map[string]bool) (pageDefaultExport, bool) {
	switch statement.Kind {
	case ast.KindFunctionDeclaration:
		declaration := statement.AsFunctionDeclaration()
		if !module.HasDefaultModifier(declaration.Modifiers()) {
			return pageDefaultExport{}, false
		}
		if declared := declaration.Name(); declared != nil {
			return pageDefaultExport{name: declared.Text()}, true
		}
		return pageDefaultExport{}, true

	case ast.KindClassDeclaration:
		declaration := statement.AsClassDeclaration()
		if !module.HasDefaultModifier(declaration.Modifiers()) {
			return pageDefaultExport{}, false
		}
		if declared := declaration.Name(); declared != nil {
			return pageDefaultExport{name: declared.Text()}, true
		}
		return pageDefaultExport{}, true

	case ast.KindExportAssignment:
		assignment := statement.AsExportAssignment()
		if assignment.IsExportEquals {
			// `export = x` is the CommonJS form and is not a default export.
			return pageDefaultExport{}, false
		}
		expression := ast.SkipParentheses(assignment.Expression)
		if expression == nil || expression.Kind != ast.KindIdentifier {
			return pageDefaultExport{}, true
		}
		return localDefaultExport(expression.Text(), imported), true

	case ast.KindExportDeclaration:
		declaration := statement.AsExportDeclaration()
		if declaration.IsTypeOnly || declaration.ExportClause == nil || declaration.ExportClause.Kind != ast.KindNamedExports {
			return pageDefaultExport{}, false
		}
		elements := declaration.ExportClause.AsNamedExports().Elements
		if elements == nil {
			return pageDefaultExport{}, false
		}
		for _, specifier := range elements.Nodes {
			if specifier.AsExportSpecifier().IsTypeOnly {
				continue
			}
			exported := specifier.Name()
			if exported == nil || exported.Text() != "default" {
				continue
			}
			if declaration.ModuleSpecifier != nil {
				return pageDefaultExport{}, true
			}
			local := exported
			if propertyName := specifier.PropertyName(); propertyName != nil {
				local = propertyName
			}
			return localDefaultExport(local.Text(), imported), true
		}
	}

	return pageDefaultExport{}, false
}

// localDefaultExport is a default export naming a binding: held to the conventions when this file
// declared it, and to nothing when it was imported.
func localDefaultExport(name string, imported map[string]bool) pageDefaultExport {
	if imported[name] {
		return pageDefaultExport{}
	}
	return pageDefaultExport{name: name, separate: true}
}

// importedLocalNames is every local name the file's import declarations bind.
func importedLocalNames(sourceFile *ast.Node) map[string]bool {
	names := map[string]bool{}
	sourceFile.ForEachChild(func(statement *ast.Node) bool {
		bindings := imports.BindingsOf(statement)
		if bindings.Default != nil {
			names[bindings.Default.Text()] = true
		}
		if bindings.Namespace != nil {
			if name := bindings.Namespace.Name(); name != nil {
				names[name.Text()] = true
			}
		}
		for _, specifier := range bindings.Named {
			if name := specifier.Name(); name != nil {
				names[name.Text()] = true
			}
		}
		return false
	})
	return names
}

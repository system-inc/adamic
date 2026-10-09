package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageVarsOnTop = rule.Message{
	Id: "top",
	Description: "All 'var' declarations must be at the top of the function scope. " +
		"A `var` is hoisted to the top of its function whatever line it is written on, so a " +
		"declaration buried in a loop or a branch reads as scoped to that block and is not. " +
		"Writing them together at the top makes the text say what the engine already does.",
}

// VarsOnTop requires every `var` declaration to sit at the top of its containing scope.
//
//	valid:   function foo() { var first; var second = 1; first = second; }
//	valid:   'use strict'; var x; f();
//	valid:   import React from 'react'; var y;
//	valid:   class C { static { var x; foo(); } }
//	invalid: function foo() { var first; first = 1; var second = 1; }
//	invalid: function foo() { for (var i = 0; i < 10; i++) { } }
//	invalid: class C { static { foo(); var x; } }
//
// # What "on top" means
//
// The declaration has to be a statement of a container this rule accepts, and every statement
// before it in that container has to be a `var` declaration too, once a leading run of directives
// and imports is skipped. So a second `var` after a first is still on top, and the first statement
// that is not a `var` closes the run for everything after it.
//
// Three containers are accepted and nothing else: the program, a function body block, and a class
// static block. A `var` inside an `if`, a loop, a `switch`, or a `try` is reported wherever it sits,
// which is most of upstream's invalid cases, because those blocks do not scope it and the text
// implies they do.
//
// # `export var` needs no special handling here, and upstream needs a lot
//
// Upstream's parser wraps an exported declaration in an ExportNamedDeclaration node, so the `var`
// is not itself a member of the statement list. That costs it two pieces of machinery: it reports
// the PARENT so the finding does not point inside the export, and it counts an export-wrapped
// declaration as a declaration when scanning the run, without which `var x; export var y; var z;`
// would report on `z`.
//
// Our parser makes `export` a MODIFIER on the variable statement itself, so the statement in the
// list is the declaration, its parent is the source file, and both pieces are unnecessary. Probed
// before this was written rather than after: the three exported cases in upstream's corpus each
// produce one VariableStatement with one modifier and a source-file parent.
//
// This is fidelity to the decision rather than to the mechanism. The three cases behave identically
// and the reasoning is here so the absence does not read as an oversight.
//
// # Two things the directive skip gets right that are easy to get wrong
//
// The skip is a PREFIX scan rather than a filter. `'use strict'; var x; 'directive'; var y;` reports
// on `y`, because the run of directives ended at the first `var` and the later string is an
// ordinary expression statement sitting between two declarations. A filter that removed every
// directive-looking statement anywhere would call it clean.
//
// The skip does not happen at all inside a class static block, which upstream states outright and
// its corpus pins: `class C { static { 'use strict'; var x; } }` REPORTS. A static block has no
// directive prologue and no import, so a string there is just an expression. This is the one place
// the two container kinds behave differently and it is easy to unify them by accident.
var VarsOnTop = rule.Rule{
	Name: "vars-on-top",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Anchored on the declaration LIST rather than on the variable statement, because a `var`
		// in a loop header has no statement wrapping it.
		//
		// Upstream can anchor on its VariableDeclaration because its parser uses one node for both
		// positions. Ours splits them: `var x;` is a VariableStatement holding a list, while
		// `for (var i = 0; ...)` puts the list straight into the loop with no statement at all.
		// Three of upstream's invalid cases are loop headers and a listener on the statement kind
		// alone goes silent on every one of them, which is how this was found.
		return rule.Listeners{
			ast.KindVariableDeclarationList: func(declarationList *ast.Node) {
				// Only `var`. Everything else is already scoped to the block it is written in, so
				// the complaint does not apply.
				if declarationList.Flags&(ast.NodeFlagsLet|ast.NodeFlagsConst|ast.NodeFlagsUsing|
					ast.NodeFlagsAwaitUsing) != 0 {
					return
				}

				// The node that sits in a statement list, and the node the finding points at. For
				// an ordinary declaration that is the statement wrapping the list; in a loop header
				// there is none, and the list itself is reported where it stands.
				subject := declarationList
				if declarationList.Parent != nil &&
					declarationList.Parent.Kind == ast.KindVariableStatement {
					subject = declarationList.Parent
				}
				container := subject.Parent
				if container == nil {
					return
				}

				switch container.Kind {
				case ast.KindSourceFile:
					if !varsOnTopIsOnTop(subject, container.AsSourceFile().Statements, false) {
						ctx.ReportNode(subject, messageVarsOnTop)
					}
				case ast.KindBlock:
					grandparent := container.Parent
					if grandparent != nil && ast.IsFunctionLike(grandparent) &&
						varsOnTopIsOnTop(subject, container.AsBlock().Statements, false) {
						return
					}
					if grandparent != nil && grandparent.Kind == ast.KindClassStaticBlockDeclaration &&
						varsOnTopIsOnTop(subject, container.AsBlock().Statements, true) {
						return
					}
					ctx.ReportNode(subject, messageVarsOnTop)
				default:
					ctx.ReportNode(subject, messageVarsOnTop)
				}
			},
		}
	},
}

// varsOnTopIsOnTop answers whether `subject` is reached without passing anything but directives,
// imports and other variable declarations.
//
// `insideStaticBlock` turns off the directive and import skip, which upstream does explicitly
// because a static block has neither.
func varsOnTopIsOnTop(subject *ast.Node, statements *ast.NodeList, insideStaticBlock bool) bool {
	if statements == nil {
		return false
	}
	index := 0
	if !insideStaticBlock {
		for index < len(statements.Nodes) {
			statement := statements.Nodes[index]
			if !varsOnTopLooksLikeDirective(statement) && !varsOnTopLooksLikeImport(statement) {
				break
			}
			index++
		}
	}
	for ; index < len(statements.Nodes); index++ {
		statement := statements.Nodes[index]
		if !varsOnTopIsVariableDeclaration(statement) {
			return false
		}
		if statement == subject {
			return true
		}
	}
	return false
}

// varsOnTopLooksLikeDirective is upstream's structural test rather than a semantic one: any
// expression statement whose expression is a string literal, not only `'use strict'`.
//
// The corpus pins the looseness. `'use strict'; 'directive'; var x; var y; f();` is clean, so a
// second unrelated string still counts as part of the prologue.
//
// A TEMPLATE literal deliberately does not count, and this is the one place a plausible improvement
// would have been a real divergence. Upstream tests `expression.type === "Literal"`, and a template
// is a TemplateLiteral in that grammar rather than a Literal, so it falls out. Our parser gives a
// backtick string its own kind too, and a first draft accepted it on the reasoning that a template
// with no substitutions is a string. Measured against the installed build at 10.8.1 with controls
// either side:
//
//	"use strict"; var x; f();   clean
//	`use strict`; var x; f();   REPORTS
//	`hello`; var x;             REPORTS
//
// A template is never a directive to any engine either, so upstream is right and the improvement
// was not one. Found by a surviving mutant, because no upstream case writes a backtick here.
func varsOnTopLooksLikeDirective(node *ast.Node) bool {
	if node.Kind != ast.KindExpressionStatement {
		return false
	}
	expression := node.AsExpressionStatement().Expression
	return expression != nil && expression.Kind == ast.KindStringLiteral
}

// varsOnTopLooksLikeImport answers whether a statement is an import.
//
// Upstream also lists the three specifier kinds, which cannot appear as statements in any parser and
// are unreachable there too. Only the declaration is named here, and the omission is deliberate
// rather than a gap: a specifier is a child of the declaration and never a member of a statement
// list, so a case for it cannot be written.
//
// An `import x = require('y')` is the TypeScript spelling and our parser produces it where upstream
// has no node at all. It is accepted, because it is an import by every reading of what the skip is
// for, and the divergence is stated rather than silent.
func varsOnTopLooksLikeImport(node *ast.Node) bool {
	return node.Kind == ast.KindImportDeclaration || node.Kind == ast.KindImportEqualsDeclaration
}

// varsOnTopIsVariableDeclaration counts a variable statement.
//
// Upstream's second arm, for a declaration wrapped in an export, has no counterpart here because
// our parser does not wrap. See the rule doc comment.
func varsOnTopIsVariableDeclaration(node *ast.Node) bool {
	return node.Kind == ast.KindVariableStatement
}

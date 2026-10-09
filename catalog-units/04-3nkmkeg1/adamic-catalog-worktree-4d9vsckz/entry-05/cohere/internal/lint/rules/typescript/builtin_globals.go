package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// The globals an ESLint run under typescript-eslint's parser holds before any rule runs, which is what
// upstream's `builtinGlobals` options compare against. One home for the two rules that carry the
// option, no-redeclare and no-shadow, so the set they mean cannot drift apart.
//
// The helpers take the Program rather than a rule.Context, so each rule's call site names ctx.Program and
// the guard counts the read against it: a rule reading this set declares rule.ReadsDefaultLibrary, since
// the lib half asks whether a global's declarations live in a default library file.

// eslintLatestGlobals is ESLint 10.8.1's own ECMAScript globals at `ecmaVersion: "latest"` (es2026 in
// conf/globals.js), which its Linter declares read-only in every global scope before any rule runs
// (lib/languages/js/source-code/source-code.js:955 at 10.8.1, `getGlobalsForEcmaVersion`), on top of
// whatever the parser's scope manager put there. They are why `var NaN = 1;` and
// `function toString() {}` report in a script although the TypeScript lib gives neither a type.
var eslintLatestGlobals = map[string]bool{
	"AggregateError": true, "Array": true, "ArrayBuffer": true, "AsyncDisposableStack": true,
	"Atomics": true, "BigInt": true, "BigInt64Array": true, "BigUint64Array": true, "Boolean": true,
	"DataView": true, "Date": true, "DisposableStack": true, "Error": true, "EvalError": true,
	"FinalizationRegistry": true, "Float16Array": true, "Float32Array": true, "Float64Array": true,
	"Function": true, "Infinity": true, "Int16Array": true, "Int32Array": true, "Int8Array": true,
	"Intl": true, "Iterator": true, "JSON": true, "Map": true, "Math": true, "NaN": true,
	"Number": true, "Object": true, "Promise": true, "Proxy": true, "RangeError": true,
	"ReferenceError": true, "Reflect": true, "RegExp": true, "Set": true, "SharedArrayBuffer": true,
	"String": true, "SuppressedError": true, "Symbol": true, "SyntaxError": true, "Temporal": true,
	"TypeError": true, "URIError": true, "Uint16Array": true, "Uint32Array": true, "Uint8Array": true,
	"Uint8ClampedArray": true, "WeakMap": true, "WeakRef": true, "WeakSet": true, "constructor": true,
	"decodeURI": true, "decodeURIComponent": true, "encodeURI": true, "encodeURIComponent": true,
	"escape": true, "eval": true, "globalThis": true, "hasOwnProperty": true, "isFinite": true,
	"isNaN": true, "isPrototypeOf": true, "parseFloat": true, "parseInt": true,
	"propertyIsEnumerable": true, "toLocaleString": true, "toString": true, "undefined": true,
	"unescape": true, "valueOf": true,
}

// isBuiltinGlobal reports whether an ESLint run under typescript-eslint's parser holds name as a
// read-only global: ESLint's own ECMAScript globals, or a type the program's lib declares.
func isBuiltinGlobal(typeChecker *checker.Checker, program rule.Program, name string) bool {
	return eslintLatestGlobals[name] || isLibraryTypeGlobal(typeChecker, program, name)
}

// builtinGlobalIsValue reports whether upstream counts a builtin global as a value, which no-shadow's
// function-type-parameter exemption asks of what it shadows.
//
// A lib global is the scope manager's ImplicitLibVariable, a value only when the lib declares it as
// both (`Object`, `Promise`), so `Record` and `Partial` are types alone. A name ESLint alone declares
// (`NaN`, `parseInt`) is a plain scope variable with no such flag, and upstream reads it as a value.
// Where both declare a name, the variable is the lib's, so the lib's answer wins. Measured against the
// installed rule on the edge rows of no-shadow's replay.
func builtinGlobalIsValue(typeChecker *checker.Checker, program rule.Program, name string) bool {
	if !isLibraryTypeGlobal(typeChecker, program, name) {
		return true
	}
	symbol := typeChecker.GetGlobalSymbol(name, ast.SymbolFlagsAll, nil)
	for _, declaration := range symbol.Declarations {
		switch declaration.Kind {
		case ast.KindVariableDeclaration, ast.KindFunctionDeclaration, ast.KindClassDeclaration,
			ast.KindEnumDeclaration:
		default:
			continue
		}
		if file := ast.GetSourceFileOfNode(declaration); file != nil && program.IsSourceFileDefaultLibrary(file.PathKey()) {
			return true
		}
	}
	return false
}

// isLibraryTypeGlobal reports whether the program's standard library declares name as a type, which
// is what typescript-eslint's scope manager seeds the global scope with.
//
// Read through the global symbol rather than a list of the lib files, which would be a read of other
// files that no findings key covers. The checker merges lib files into the globals first, and when a
// script's declaration conflicts with a lib one, `mergeSymbol` reports it and keeps the lib's symbol,
// so a lib declaration is on the global whether the user's merged into it or not.
//
// The kinds are the ones the scope manager's lib generator keeps, the type variables: a
// `declare var NaN` or a `declare function parseInt` alone is a value with no type and is not a
// builtin to upstream, while `interface Object` beside `declare var Object` is.
func isLibraryTypeGlobal(typeChecker *checker.Checker, program rule.Program, name string) bool {
	symbol := typeChecker.GetGlobalSymbol(name, ast.SymbolFlagsAll, nil)
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		switch declaration.Kind {
		case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindClassDeclaration,
			ast.KindEnumDeclaration, ast.KindModuleDeclaration:
		default:
			continue
		}
		if file := ast.GetSourceFileOfNode(declaration); file != nil && program.IsSourceFileDefaultLibrary(file.PathKey()) {
			return true
		}
	}
	return false
}

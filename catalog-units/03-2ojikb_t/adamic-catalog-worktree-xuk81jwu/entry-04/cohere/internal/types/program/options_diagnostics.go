package program

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Options diagnostics are TypeScript's findings about the compiler options themselves rather than about
// any source file: an option TypeScript 7 removed, two options that contradict each other, a path
// mapping with two asterisks. They are reported, and they fail the run, but they do not stop lint
// (@system_cohere, 2026-10-04, from the quiet hundred's #wvgxtey).
//
// The types phase stops lint because findings against wrong semantics are noise. A removed option does
// not make the semantics wrong by itself, and where it does, the source fails to type-check and that
// diagnostic stops lint on its own. excalidraw sets `baseUrl`, which is TS5102 under TypeScript 7, and
// its first run showed no lint at all over that one line of tsconfig.
//
// Each code is listed with where upstream raises it, so the list can be checked against upstream when
// the pin moves. A code is an options diagnostic only when it also sits in no file or in a file outside
// the program, the tsconfig: TS6053 "File not found" is also what a `/// <reference path>` in source
// raises, and that one is about source.

// optionsDiagnosticCodes are the codes the program raises from the compiler options and project
// references (compiler.Program.verifyCompilerOptions and verifyProjectReferences), with each one's
// message. Messages upstream only chains under one of these (TS5068, TS5106, TS5111) are not codes a
// diagnostic is raised with, so they are not listed.
var optionsDiagnosticCodes = map[int32]string{
	5009:  "Cannot find the common subdirectory path for the input files",
	5011:  "The common source directory is not the inferred rootDir: rootDir must be set explicitly",
	5051:  "An option can only be used when inlineSourceMap or sourceMap is provided",
	5052:  "An option cannot be specified without specifying another",
	5053:  "An option cannot be specified with another",
	5055:  "Cannot write a file because it would overwrite an input file",
	5056:  "Cannot write a file because multiple input files would write it",
	5059:  "Invalid value for reactNamespace",
	5061:  "A paths pattern can have at most one asterisk",
	5062:  "A paths substitution can have at most one asterisk",
	5063:  "A paths pattern's substitutions should be an array",
	5066:  "A paths pattern's substitutions should not be an empty array",
	5067:  "Invalid value for jsxFactory",
	5069:  "An option cannot be specified without specifying one of two others",
	5074:  "incremental is only valid with a known configuration file or tsBuildInfoFile",
	5089:  "An option cannot be specified with this jsx setting",
	5090:  "Non-relative paths are not allowed in paths without a leading ./",
	5091:  "preserveConstEnums cannot be disabled under isolatedModules or verbatimModuleSyntax",
	5095:  "An option can only be used when module is preserve, commonjs or es2015 or later",
	5096:  "allowImportingTsExtensions needs noEmit, emitDeclarationOnly or rewriteRelativeImportExtensions",
	5098:  "An option can only be used when moduleResolution is node16, nodenext or bundler",
	5102:  "An option has been removed (TypeScript 7: baseUrl, outFile, downlevelIteration)",
	5108:  "An option's value has been removed (TypeScript 7: target ES5, module AMD, System or UMD, and others)",
	5109:  "moduleResolution must match module",
	5110:  "module must match moduleResolution",
	6053:  "A referenced project's tsconfig was not found",
	6304:  "Composite projects may not disable declaration emit",
	6306:  "A referenced project must set composite",
	6307:  "A file is not listed within a composite project's file list",
	6310:  "A referenced project may not disable emit",
	6377:  "Cannot write a file because it would overwrite a referenced project's tsbuildinfo",
	6379:  "Composite projects may not disable incremental compilation",
	18035: "Invalid value for jsxFragmentFactory",
}

// configParsingOptionCodes are the codes reading a tsconfig raises about one option's name or value
// (tsoptions, while it converts compilerOptions). Each is an options diagnostic, and the graph still
// builds over it, the option ignored or left at its default, as tsc does (@system_cohere, 2026-10-04).
//
// The reason these matter is TS5023: TypeScript 7 does not know the options TypeScript 5.5 removed
// (importsNotUsedAsValues, suppressImplicitAnyIndexErrors, keyofStringsOnly and the rest), so many
// TypeScript 5 tsconfigs fail to read before they could reach TS5102. Every other diagnostic from
// reading a tsconfig still refuses the build: a JSON syntax error or an unreadable extends means the
// options cohere would check against are not the ones written.
var configParsingOptionCodes = map[int32]string{
	5023: "Unknown compiler option",
	5024: "A compiler option requires a value of another type",
	5025: "Unknown compiler option, with a suggestion",
	6046: "An option's value is not one of the values it takes",
}

// IsOptionsDiagnostic reports whether a diagnostic is about the compiler options rather than about
// source: a code from either list, located in no file or in one outside this program.
func (g *Graph) IsOptionsDiagnostic(diagnostic *ast.Diagnostic) bool {
	_, raisedByProgram := optionsDiagnosticCodes[diagnostic.Code()]
	_, raisedByConfig := configParsingOptionCodes[diagnostic.Code()]
	if !raisedByProgram && !raisedByConfig {
		return false
	}
	sourceFile := diagnostic.File()
	return sourceFile == nil || g.Program.GetSourceFile(sourceFile.FileName()) != sourceFile
}

// onlyOptionValueDiagnostics reports whether every diagnostic from reading a tsconfig is about one
// option's name or value, so the graph can still be built. See configParsingOptionCodes.
func onlyOptionValueDiagnostics(diagnostics []*ast.Diagnostic) bool {
	for _, diagnostic := range diagnostics {
		if _, isOptionValue := configParsingOptionCodes[diagnostic.Code()]; !isOptionValue {
			return false
		}
	}
	return true
}

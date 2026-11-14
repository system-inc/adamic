package lower

import (
	"strings"
	"testing"
)

// A probe, never for landing (#5d9y3rs): what checkRecordGuardedReads makes of tsc's
// isTypeOnlyImportDeclaration shape, with each piece taken away in turn.
const recordGuardTypes = "enum SyntaxKind { Identifier, ImportClause, NamedImports, ImportSpecifier, NamespaceImport, TypeKeyword, DeferKeyword }\ninterface Node {\n\treadonly kind: SyntaxKind;\n\treadonly parent: Node;\n}\ninterface ImportClause extends Node {\n\treadonly kind: SyntaxKind.ImportClause;\n\treadonly phaseModifier: SyntaxKind.TypeKeyword | SyntaxKind.DeferKeyword | undefined;\n}\ninterface NamedImports extends Node {\n\treadonly kind: SyntaxKind.NamedImports;\n\treadonly parent: ImportClause;\n}\ninterface ImportSpecifier extends Node {\n\treadonly kind: SyntaxKind.ImportSpecifier;\n\treadonly parent: NamedImports;\n\treadonly isTypeOnly: boolean;\n}\n"

const recordGuardRecord = "const table: { [key: string]: number } = {};\nfunction put(key: string): void {\n\ttable[key] = 1;\n}\nput(String(1));\n"

func recordGuardFunction(clauseCase bool) string {
	source := "function isTypeOnly(node: Node): boolean {\n\tswitch (node.kind) {\n\t\tcase SyntaxKind.ImportSpecifier:\n\t\t\treturn (node as ImportSpecifier).isTypeOnly || (node as ImportSpecifier).parent.parent.phaseModifier === SyntaxKind.TypeKeyword;\n"
	if clauseCase {
		source += "\t\tcase SyntaxKind.ImportClause:\n\t\t\treturn (node as ImportClause).phaseModifier === SyntaxKind.TypeKeyword;\n"
	}
	return source + "\t}\n\treturn false;\n}\n"
}

// The root's parent is itself, set in its constructor, so no node needs a cast to be made.
const recordGuardMain = "class Root {\n\treadonly kind: SyntaxKind = SyntaxKind.Identifier;\n\treadonly parent: Node;\n\tconstructor() {\n\t\tthis.parent = this;\n\t}\n}\nconst clause: ImportClause = { kind: SyntaxKind.ImportClause, parent: new Root(), phaseModifier: SyntaxKind.TypeKeyword };\nconst named: NamedImports = { kind: SyntaxKind.NamedImports, parent: clause };\nconst specifier: ImportSpecifier = { kind: SyntaxKind.ImportSpecifier, parent: named, isTypeOnly: false };\nconsole.log(String(isTypeOnly(specifier)));\nconsole.log(String(isTypeOnly(clause)));\n"

// The hazard: a record passed as a weak type, downcast to ImportClause by its kind, and read plainly.
const recordGuardHazard = "interface Weak {\n\treadonly kind?: SyntaxKind;\n}\nfunction plain(c: ImportClause): string {\n\treturn String(c.phaseModifier);\n}\nfunction phase(w: Weak): string {\n\tif (w.kind === SyntaxKind.ImportClause) {\n\t\treturn plain(w as ImportClause);\n\t}\n\treturn 'no';\n}\nconsole.log(phase(table));\n"

func TestZZRecordGuardProbe(t *testing.T) {
	for _, probe := range []struct{ name, source string }{
		{"tsc shape", recordGuardTypes + recordGuardRecord + recordGuardFunction(true) + recordGuardMain},
		{"no ImportClause cast", recordGuardTypes + recordGuardRecord + recordGuardFunction(false) + recordGuardMain},
		{"no record", recordGuardTypes + recordGuardFunction(true) + recordGuardMain},
		{"hazard alone", recordGuardTypes + recordGuardRecord + recordGuardHazard},
		{"hazard beside tsc shape", recordGuardTypes + recordGuardRecord + recordGuardFunction(true) + recordGuardMain + recordGuardHazard},
	} {
		_, err := lowerSourceAs(t, probe.source, ".ts")
		verdict := "lowers"
		if err != nil {
			verdict = strings.ReplaceAll(err.Error(), "\n", " / ")
		}
		t.Logf("PROBE %s: %s", probe.name, verdict)
	}
}

package lower

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// arrayPredicateNodeBase is tsc's Node, NodeArray and VisitResult in small, with a cloneNode-like cast written as an
// intersection over a type parameter constrained to Node, and tsc's isArray (core.ts:1772). Before rulings.md:260's
// amendment (Nov 27, @system_adamic on #67fpbxv), Cloned counted as stored as an array whatever T's constraint, so
// Node's kind, flags and parent were names an array could have, every name Node requires was one, and Node was out of
// Array.isArray's domain. With it, Cloned counts only while Node may hold an array, and kind keeps it from holding one.
const arrayPredicateNodeBase = `interface TextRange { pos: number; end: number }
interface Node extends TextRange { readonly kind: number; readonly flags: number; readonly parent: Node | undefined }
interface SourceFile extends Node { readonly fileName: string }
interface NodeArray<T extends Node> extends ReadonlyArray<T>, TextRange { readonly hasTrailingComma?: boolean }
type VisitResult<T extends Node | undefined> = T | readonly Node[];
type Cloned<T extends Node> = T & SourceFile;
function isArray(value: unknown): value is readonly unknown[] { return Array.isArray(value); }
`

// arrayPredicateNodeWrapper and arrayPredicateNodeDirect narrow VisitResult<Node | undefined> by an array test, as
// visitNode does at visitorPublic.ts:169, through tsc's isArray and through Array.isArray itself.
const (
	arrayPredicateNodeWrapper = "function single(visited: VisitResult<Node | undefined>): Node | undefined {\n    return isArray(visited) ? visited[0] : visited;\n}\n"
	arrayPredicateNodeDirect  = "function single(visited: VisitResult<Node | undefined>): Node | undefined {\n    return Array.isArray(visited) ? visited[0] : visited;\n}\n"
)

// arrayPredicateNodeUse hands single a Node and an array of one, so a brand that called the Node an array would print
// false on the first line, and one that called the array no array would print false on the second.
const arrayPredicateNodeUse = `const leaf: Node = { kind: 1, flags: 0, parent: undefined, pos: 0, end: 1 };
console.log(String(single(leaf) === leaf));
console.log(String(single([leaf]) === leaf));
`

// The two stops Node's arm meets outside Array.isArray's domain: the wrapper's narrowed read (locals.go's checked
// narrowed-union read) and the direct call's argument (arrayIsArray).
const (
	arrayPredicateNarrowedWords = "a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables"
	arrayPredicateCallWords     = "Array.isArray on a tuple or an erased object/any/unknown view"
)

// arrayPredicateConstraintRows are TestArrayPredicateConstraintAdmitsNode's rows, each with the stop its program met
// before the amendment, by position and words, which the mutant restoring the old clause must bring back
// (TestArrayStoredConstraintMutantIsCaught).
var arrayPredicateConstraintRows = []struct{ name, source, where, what string }{
	{"wrapper", arrayPredicateNodeBase + arrayPredicateNodeWrapper + arrayPredicateNodeUse, ":9:31", arrayPredicateNarrowedWords},
	{"direct", arrayPredicateNodeBase + arrayPredicateNodeDirect + arrayPredicateNodeUse, ":9:26", arrayPredicateCallWords},
	{"two-parameters", arrayPredicateNodeBase + "type Paired<T extends Node, U> = T & U & SourceFile;\n" + arrayPredicateNodeWrapper + arrayPredicateNodeUse, ":10:31", arrayPredicateNarrowedWords},
}

// Node enters Array.isArray's domain once Cloned is settled by T's constraint (rulings.md:260's amendment, Nov 27,
// @system_adamic on #67fpbxv): Node requires kind, and no type stored as an array has it, so Cloned, `T & SourceFile`
// with T extends Node, holds no array and adds no names. VisitResult<Node | undefined> narrowed by tsc's isArray and by
// Array.isArray lowers and agrees with Node. two-parameters holds the rule for an intersection holding two type
// parameters: U has no constraint and may hold an array, but T's can't, and a value of Paired is a value of T, so
// Paired counts only when every parameter's constraint may hold one, and here it doesn't.
// Mutant: the old clause restored, every intersection holding a type parameter counted as stored as an array
// (TestArrayStoredConstraintMutantIsCaught). Every row goes red at its old stop.
func TestArrayPredicateConstraintAdmitsNode(t *testing.T) {
	t.Parallel()
	for _, row := range arrayPredicateConstraintRows {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, row.source, ".ts")
		})
	}
}

// Where a type parameter's constraint may hold an array, its intersection still counts, so its names keep Node out of
// the domain and each program keeps its stop, by its position and words:
//   - arraynode-wrapper and arraynode-direct, Language's pin on #67fpbxv: `interface ArrayNode extends Array<number>,
//     Node {}` is stored as an array and declares kind, so Node, Cloned's constraint, may hold an array; Cloned counts
//     again, the least fixpoint reaching it, and flags and parent are array names with kind;
//   - unconstrained, any, unknown and object: a Loose alias whose parameter's constraint is missing, any, unknown or
//     object, each of which may hold an array, so Loose counts and its kind, flags and parent with it.
//
// Node prints true twice for each, as it does for the rows that lower.
func TestArrayPredicateConstraintMayHoldStops(t *testing.T) {
	t.Parallel()
	const arrayNode = "interface ArrayNode extends Array<number>, Node {}\n"
	for _, row := range []struct{ name, source, where, what string }{
		{"arraynode-wrapper", arrayPredicateNodeBase + arrayNode + arrayPredicateNodeWrapper + arrayPredicateNodeUse, ":10:31", arrayPredicateNarrowedWords},
		{"arraynode-direct", arrayPredicateNodeBase + arrayNode + arrayPredicateNodeDirect + arrayPredicateNodeUse, ":10:26", arrayPredicateCallWords},
		{"unconstrained", arrayPredicateNodeBase + "type Loose<T> = T & SourceFile;\n" + arrayPredicateNodeWrapper + arrayPredicateNodeUse, ":10:31", arrayPredicateNarrowedWords},
		{"any", arrayPredicateNodeBase + "type Loose<T extends any> = T & SourceFile;\n" + arrayPredicateNodeWrapper + arrayPredicateNodeUse, ":10:31", arrayPredicateNarrowedWords},
		{"unknown", arrayPredicateNodeBase + "type Loose<T extends unknown> = T & SourceFile;\n" + arrayPredicateNodeWrapper + arrayPredicateNodeUse, ":10:31", arrayPredicateNarrowedWords},
		{"object", arrayPredicateNodeBase + "type Loose<T extends object> = T & SourceFile;\n" + arrayPredicateNodeWrapper + arrayPredicateNodeUse, ":10:31", arrayPredicateNarrowedWords},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			path := writeAgreementSource(t, "main.ts", row.source)
			if node := runAgreementNode(t, path); string(node.stdout) != "true\ntrue\n" || node.code != 0 {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q; want true twice", node.code, node.stdout, node.stderr)
			}
			program, err := load.Load([]string{path})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			_, err = Lower(context.Background(), program)
			var notYet *NotYet
			if !errors.As(err, &notYet) || notYet.Where != path+row.where || notYet.What != row.what {
				t.Fatalf("want NotYet at %s, %q, got %v", row.where, row.what, err)
			}
		})
	}
}

// tscNodeNames are the seven names tsc's Node and JSDoc require: kind, flags, modifierFlagsCache, parent and
// transformFlags, and TextRange's pos and end.
var tscNodeNames = []string{"end", "flags", "kind", "modifierFlagsCache", "parent", "pos", "transformFlags"}

// The library axiom under rulings.md:260's amendment (Nov 27, @system_adamic on #67fpbxv): the library's types stored
// as an array, Array and ReadonlyArray and every interface or class the library or the prelude declares that extends
// one, found by arrayStoredPropertyNames' own library pass (libraryExtensions), are RegExpExecArray, RegExpMatchArray,
// RegExpIndicesArray and TemplateStringsArray, and none of them has any of Node's seven names. If one did, Node would
// be a constraint that may hold an array, and tsc's cloneNode casts would keep Node out of the domain again, so a
// library bump that adds such a type, or such a name, fails here naming it, and the ruling's row is read again.
func TestArrayStoredLibraryTypesDeclareNoNodeName(t *testing.T) {
	t.Parallel()
	program, err := load.Load([]string{writeAgreementSource(t, "main.ts", "console.log('library');\n")})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	l := &lowering{program: program, checker: checked}
	stored := map[string]*checker.Type{
		"Array":         predicateReadonlyArray(checked, checked.GetUnknownType(), false),
		"ReadonlyArray": predicateReadonlyArray(checked, checked.GetUnknownType(), true),
	}
	extending := []string{}
	for _, extension := range l.libraryExtensions() {
		symbol := checked.GetSymbolOfDeclaration(extension)
		if symbol == nil {
			continue
		}
		declared := checked.GetDeclaredTypeOfSymbol(symbol)
		if !l.arrayStored(declared, map[*checker.Type]bool{}) {
			continue
		}
		if _, met := stored[symbol.Name]; !met {
			extending = append(extending, symbol.Name)
		}
		stored[symbol.Name] = declared
	}
	slices.Sort(extending)
	if want := []string{"RegExpExecArray", "RegExpIndicesArray", "RegExpMatchArray", "TemplateStringsArray"}; !slices.Equal(extending, want) {
		t.Errorf("the library's array-extending types are %v, want %v: read rulings.md:260's library axiom again and update both", extending, want)
	}
	for name, proven := range stored {
		for _, property := range checked.GetPropertiesOfType(proven) {
			if slices.Contains(tscNodeNames, property.Name) {
				t.Errorf("%s, stored as an array, has %s, one of Node's names (%s), so Node may hold an array", name, property.Name, strings.Join(tscNodeNames, ", "))
			}
		}
	}
}

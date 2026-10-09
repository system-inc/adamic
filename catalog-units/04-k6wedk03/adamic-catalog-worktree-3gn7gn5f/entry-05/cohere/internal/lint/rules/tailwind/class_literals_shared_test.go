package tailwind

import (
	"reflect"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// Thirteen rules read class surfaces through ClassLiteralSurfaces.ReaderFor, and the saving is only worth
// having if sharing never changes a reading. These pin the three things that could: rules with the
// same settings must share one reader per file, rules with different settings must never share one,
// and a remembered reading must be the reading a fresh reader would make.

// sharedReaderSource puts a class string on every surface and in every value position the reader
// enters, plus a `tw` attribute that only a reader configured for it should see.
const sharedReaderSource = "const className = (open ? 'flex flex' : 'hidden') as const;\n" +
	"const merged = cn('p-2', open && 'gap-2', ['m-1', size ?? 'm-2'], `px-${size} py-1`);\n" +
	"const element = <div tw=\"text-left text-left\" className={`flex ${open ? ' rotate-90 ' : ''}`} />;\n" +
	"const other = <span className=\"items-center\" title=\"not classes\" />;\n" +
	"const notClasses = 'flex flex';\n"

func TestClassLiteralReaderIsSharedPerSettingsPerFile(t *testing.T) {
	t.Parallel()
	twSettings := DefaultClassLiteralSettings()
	twSettings.AttributePatterns = []string{"tw"}

	cache := rule.NewFileCache()
	first := ClassLiteralSurfacesFor(DefaultClassLiteralSettings()).ReaderFor(cache)
	second := DefaultClassLiteralSurfaces().ReaderFor(cache)
	if first != second {
		t.Fatal("two rules with the same settings got two readers for one file, so each reads every node again")
	}
	if fills := cache.Fills()["tailwind.classValues:"+DefaultClassLiteralSettings().key()]; fills != 1 {
		t.Fatalf("the shared reader was built %d times for one file, want 1", fills)
	}
	if ClassLiteralSurfacesFor(twSettings).ReaderFor(cache) == first {
		t.Fatal("rules reading different attributes shared a reader, so one reads the other's surfaces")
	}

	// Names that run together the same way are still different settings.
	joined := ClassLiteralSettings{AttributePatterns: []string{"ab"}}
	split := ClassLiteralSettings{AttributePatterns: []string{"a", "b"}}
	moved := ClassLiteralSettings{AttributePatterns: []string{"a"}, CalleeNamePatterns: []string{"b"}}
	if joined.key() == split.key() || split.key() == moved.key() || joined.key() == moved.key() {
		t.Fatalf("distinct settings share a key: %q %q %q", joined.key(), split.key(), moved.key())
	}

	// A new file gets its own memo, and still the patterns compiled once for the run.
	nextFile := DefaultClassLiteralSurfaces().ReaderFor(rule.NewFileCache())
	if nextFile == first {
		t.Fatal("two files shared a memo, so a node pointer from one could answer for the other")
	}
	if nextFile.variables == nil || nextFile.variables != first.variables {
		t.Fatal("the variable patterns were compiled again for a second file")
	}
}

func TestSharedReadingAgreesWithAFreshOne(t *testing.T) {
	t.Parallel()
	fileName := tspath.NormalizePath("/Component.tsx")
	sourceFile := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: tspath.RootedFilePath(fileName), PathKey: tspath.PathKey(fileName),
	}, sharedReaderSource, core.ScriptKindTSX)
	if sourceFile == nil {
		t.Fatal("could not parse the fixture")
	}

	twSettings := DefaultClassLiteralSettings()
	twSettings.AttributePatterns = []string{"tw"}

	// Both readers share one file's cache, as two rules configured differently would.
	cache := rule.NewFileCache()
	for _, settings := range []ClassLiteralSettings{DefaultClassLiteralSettings(), twSettings} {
		bound := ClassLiteralSurfacesFor(settings).ReaderFor(cache)
		fresh := NewClassLiteralReader(settings)

		read := 0
		var walk func(node *ast.Node)
		walk = func(node *ast.Node) {
			// Twice through the bound reader, so the second answer comes from the memo.
			for range 2 {
				shared := bound.classValuesIn(node)
				if expected := fresh.classValuesIn(node); !reflect.DeepEqual(shared, expected) {
					t.Fatalf("settings %q: a shared reading of %s differs from a fresh one:\n  shared: %+v\n  fresh:  %+v",
						settings.AttributePatterns, node.Kind, shared, expected)
				}
			}
			read += len(fresh.classValuesIn(node).literals)
			node.ForEachChild(func(child *ast.Node) bool {
				walk(child)
				return false
			})
		}
		walk(sourceFile.AsNode())

		// Proof the comparison saw class strings and the memo was used, not a vacuous agreement.
		if read == 0 || len(bound.values) == 0 {
			t.Fatalf("settings %q: read %d literals and remembered %d nodes; the fixture reached nothing",
				settings.AttributePatterns, read, len(bound.values))
		}
	}
}

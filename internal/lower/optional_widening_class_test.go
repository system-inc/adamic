package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Mutant run and restored: the plain !isClassInstance(source) exemption lets
// the original .a and .ts programs lower again. The expression, mixin, module,
// grandchild and unresolved-base pins also lose this rule's refusal.
func TestOptionalWideningClassRefused(t *testing.T) {
	t.Parallel()
	base := "class Base { x = 1; }\n"
	site := "const wider: { x: number; y?: number } = view;\nconsole.log(typeof wider.y);\n"
	fixed := "class Base { x = 1; y = 'w'; }\nclass Derived extends Base {}\nconst view: Base = new Derived();\nconst wider: { x: number; y: string } = view;\nconsole.log(typeof wider.y);\n"
	for _, row := range []struct{ name, source, fixed, where string }{
		{"original", base + "class Derived extends Base { y = 'wrong'; }\nconst view: Base = new Derived();\n" + site, fixed, "4:42"},
		{"expression", base + "const D = class extends Base { y = 'w' };\nconst view: Base = new D();\n" + site, fixed, "4:42"},
		{"mixin", base + "function m(B: typeof Base) { return class extends B { y = 'w' } }\nconst D = m(Base);\nconst view: Base = new D();\n" + site, fixed, "5:42"},
		{"opaque constructor", "class Base { x: number; constructor(x: number) { this.x = x } }\nfunction m(B: new () => { x: number }) { return class extends B { z = 2 } }\nconst view: Base = new Base(1);\n" + site, fixed, "4:42"},
		{"grandchild", base + "class Child extends Base {}\nclass Grandchild extends Child { y = 'w' }\nconst view: Base = new Grandchild();\n" + site, "class Base { x = 1; y = 'w'; }\nclass Child extends Base {}\nclass Grandchild extends Child {}\nconst view: Base = new Grandchild();\nconst wider: { x: number; y: string } = view;\nconsole.log(typeof wider.y);\n", "5:42"},
		// Here no subclass declares y. Only the unresolved constructor makes its
		// possible subclass set open. The fixed program closes it explicitly.
		{"unresolved base", base + "function m(B: typeof Base) { return class extends B { z = 2 } }\nconst view: Base = new Base();\n" + site, base + "class D extends Base { z = 2 }\nconst view: Base = new D();\n" + site, "4:42"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(row.name+extension, func(t *testing.T) {
				t.Parallel()
				want := "main" + extension + ":" + row.where + ": Adamic 0.1 refuses optional property y in { x: number; y?: number; } absent from structural source Base, which can hide fields; declare y on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
				_, err := lowerSourceAs(t, row.source, extension)
				var refused *Refused
				if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), want) {
					t.Errorf("want %s, got %v", want, err)
				}
				// Main does not lower class expressions. These fixed copies use the
				// equivalent named declaration as well as declaring y on Base.
				lowersAndAgreesWithNodeAs(t, row.fixed, extension)
			})
		}
	}
}

func TestOptionalWideningClassOtherModule(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			classes := filepath.Join(directory, "classes"+extension)
			main := filepath.Join(directory, "main"+extension)
			source := "import { Base, Derived } from './classes" + extension + "';\nconst view: Base = new Derived();\nconst wider: { x: number; y?: number } = view;\nconsole.log(typeof wider.y);\n"
			if err := os.WriteFile(classes, []byte("export class Base { x = 1 }\nexport class Derived extends Base { y = 'w' }\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(main, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{main})
			if err != nil {
				t.Fatal(err)
			}
			_, err = Lower(context.Background(), program)
			want := "main" + extension + ":3:42: Adamic 0.1 refuses optional property y in { x: number; y?: number; } absent from structural source Base, which can hide fields; declare y on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
			var refused *Refused
			if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), want) {
				t.Errorf("want %s, got %v", want, err)
			}
			if err := os.WriteFile(classes, []byte("export class Base { x = 1; y = 'w' }\nexport class Derived extends Base {}\n"), 0600); err != nil {
				t.Fatal(err)
			}
			source = strings.Replace(source, "y?: number", "y: string", 1)
			if err := os.WriteFile(main, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			agreeEntry(t, main, nil, 0, nil, nil)
		})
	}
}

// Error's lib declaration is an interface plus a constructor, so it already
// reaches the structural refusal on main. Retain that measured stop.
func TestOptionalWideningClassLibraryError(t *testing.T) {
	t.Parallel()
	source := "const error = new Error('e');\nconst view: { message: string; x?: number } = error;\nconsole.log(String(view.x));\n"
	_, err := lowerSource(t, source)
	var refused *Refused
	want := "main.a:2:47: Adamic 0.1 refuses optional property x in { message: string; x?: number; } absent from structural source Error, which can hide fields; declare x on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
	if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), want) {
		t.Fatalf("want %s, got %v", want, err)
	}
	lowersAndAgreesWithNode(t, "const error = new Error('e');\nconst view: { message: string; x?: number } = { message: error.message };\nconsole.log(String(view.x));\n")
}

func TestOptionalWideningClassClosedControls(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"class Base { x = 1 }\nconst view: Base = new Base();\nconst wider: { x: number; y?: number } = view;\nconsole.log(typeof wider.y);\n",
		"class Base { x = 1 }\nclass D extends Base { z = 'w' }\nconst view: Base = new D();\nconst wider: { x: number; y?: number } = view;\nconsole.log(typeof wider.y);\n",
	} {
		lowersAndAgreesWithNode(t, source)
	}
}

// tsc exposes Base as the mixin's base constraint. That does not identify B's
// runtime constructor: other compatible classes may be passed to m.
func TestOptionalWideningClassMixinConstraint(t *testing.T) {
	t.Parallel()
	path := writeAgreementSource(t, "main.a", "class Base { x = 1 }\nfunction m(B: typeof Base) { return class extends B { z = 2 } }\nconsole.log('constraint');\n")
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	checked, release := program.Checker(context.Background(), program.Files()[0])
	defer release()
	l := lowering{program: program, checker: checked, result: &ir.Program{}}
	var instance *checker.Type
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindClassExpression {
			signatures := checked.GetSignaturesOfType(checked.GetTypeAtLocation(node), checker.SignatureKindConstruct)
			if len(signatures) == 1 {
				instance = checked.GetReturnTypeOfSignature(signatures[0])
			}
		}
		return node.ForEachChild(visit)
	}
	program.Files()[0].AsNode().ForEachChild(visit)
	if instance == nil {
		t.Fatal("missing mixin class")
	}
	if instance.Target() != nil {
		instance = instance.Target()
	}
	bases := checked.GetBaseTypes(instance)
	if len(bases) != 1 || checked.TypeToString(bases[0]) != "Base" {
		t.Fatalf("want the Base constraint, got %v", bases)
	}
	l.noteOptionalClasses()
	if !l.optionalClasses.unknown[bases[0].Symbol()] {
		t.Fatal("constraint was treated as the runtime constructor")
	}
}

// Stats is an actual class merged with an interface in Node's declarations.
// Its implementation and subclass set are outside the program's class walk.
func TestOptionalWideningClassLibraryStats(t *testing.T) {
	t.Parallel()
	source := "import { statSync } from 'node:fs';\nconst stats = statSync('.');\nconst view: { size: number; x?: number } = stats;\nconsole.log(typeof view.x);\n"
	_, err := lowerSource(t, source)
	var refused *Refused
	want := "main.a:3:44: Adamic 0.1 refuses optional property x in { size: number; x?: number; } absent from structural source Stats, which can hide fields; declare x on the source type, or build a fresh object with known fields (adamic/no-optional-widening)"
	if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), want) {
		t.Fatalf("want %s, got %v", want, err)
	}
	lowersAndAgreesWithNode(t, "import { statSync } from 'node:fs';\nconst stats = statSync('.');\nconst view: { size: number; x?: number } = { size: stats.size };\nconsole.log(typeof view.x);\n")
}

// The class optional-widening exemption relies on the nominal-class refusal:
// a structural value must stop before it can be held in a class slot.
func TestOptionalWideningClassRequiresNominalAncestry(t *testing.T) {
	t.Parallel()
	source := "class Base { x = 1; }\nconst o = { x: 1, y: 'w' };\nconst view: Base = o;\nconst wider: { x: number; y?: number } = view;\nconsole.log(typeof wider.y);\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, source, extension)
			var refused *Refused
			want := "main" + extension + ":3:20: Adamic 0.1 refuses a value without nominal ancestry seen as Base; construct that class or a subclass; use an interface for structural values (adamic/nominal-class)"
			if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), want) {
				t.Fatalf("want %s, got %v", want, err)
			}
		})
	}
}

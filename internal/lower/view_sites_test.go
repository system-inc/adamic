package lower

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// viewSitesCasts makes three casts into checked views, Other's first, in .ts, which keeps a cast's first tag checked
// where .a refuses a downcast whose checks no proof erases. either.b, read at the top level after them, has the
// union One | Two for its receiver, and nothing lowers it before One's view: Other's walk judges it unchecked, One's
// viewSchema then adds One:b, and One's walk must ask it again and register the union's key for b. A function body
// holding the read wouldn't do: bodies are lowered before the top level's casts, and the read itself registers it.
const viewSitesCasts = `interface Base {readonly kind:'one'|'two'|'other'}
interface One extends Base {readonly kind:'one';readonly b:number}
interface Two extends Base {readonly kind:'two';readonly b:number}
interface Other extends Base {readonly kind:'other';readonly c:number}
const other={kind:'other' as const,c:3};
const otherHeld:Base=other;
console.log(String((otherHeld as Other).c));
const one={kind:'one' as const,b:5};
const oneHeld:Base=one;
const viewed:One=oneHeld as One;
const two={kind:'two' as const,b:7};
const twoHeld:Base=two;
const pick:boolean=String(one.b)==="5";
const either:One|Two=pick?viewed:twoHeld as Two;
console.log(String(either.b));
`

// viewSitesTwoWrites writes a checked object field twice, on lines 5 and 6, so its view has two
// accesses to refuse.
const viewSitesTwoWrites = `interface Base {readonly kind:'box'|'other'}
interface Box extends Base {readonly kind:'box';child:{ready:boolean}}
const raw={kind:'box' as const,child:{ready:true,other:42}};
const held:Base=raw;const view=held as Box;
view.child={ready:true};
view.child={ready:false};
console.log(String(raw.child.other));
`

// A program's site index is built once however many casts it makes into checked views, which is the
// work #q1hyhzc took off: view walked every node of every module at each one. Mutant: the index built
// again at every view, and this counts three.
//
// Not parallel: viewSiteIndexBuilds counts every build in the package while it's set, and no parallel
// test runs beside a sequential one.
func TestViewSiteIndexIsBuiltOncePerLowering(t *testing.T) {
	var builds atomic.Int64
	viewSiteIndexBuilds.Store(&builds)
	defer viewSiteIndexBuilds.Store(nil)
	keyed, err := lowerKeyed(t, writeAgreementSource(t, "main.ts", viewSitesCasts))
	if err != nil {
		t.Fatal(err)
	}
	if len(keyed.lowering.viewSources) < 3 {
		t.Fatalf("the fixture made %d views, not its three casts", len(keyed.lowering.viewSources))
	}
	if built := builds.Load(); built != 1 {
		t.Fatalf("%d views built the site index %d times", len(keyed.lowering.viewSources), built)
	}
}

// An access found unchecked is asked again once a key of its name is added: either.b, unchecked at Other's
// view, is checked through its union once One's view registers One:b, and One's walk registers the union's
// own key for b, as the whole walk does at that call. Mutant: the dirty-name recheck skipped, so only a name never
// judged is asked, x.b stays unchecked, and this and TestViewWalkIndexedAgreesWithWhole go red.
//
// Not parallel: it lowers with viewWalkCompare on, so One's walk is held to the whole walk's at that call. The key
// would be registered by the time lowering ends either way, since either.b's own lowering registers it, so only the
// call-by-call comparison can tell the recheck was made.
func TestViewSiteAskedAgainWhenItsNameGainsAKey(t *testing.T) {
	viewWalkCompare.Store(true)
	defer viewWalkCompare.Store(false)
	var keyed keyedProgram
	var err error
	func() {
		defer func() {
			if failure := recover(); failure != nil {
				t.Fatalf("%v", failure)
			}
		}()
		keyed, err = lowerKeyed(t, writeAgreementSource(t, "main.ts", viewSitesCasts))
	}()
	if err != nil {
		t.Fatal(err)
	}
	// The premise: Other's view comes before One's, so x.b was judged before One:b was a key.
	order := map[string]int{}
	for index, source := range keyed.lowering.viewSources {
		if _, found := order[keyed.lowering.checker.TypeToString(source.target)]; !found {
			order[keyed.lowering.checker.TypeToString(source.target)] = index
		}
	}
	otherAt, otherFound := order["Other"]
	oneAt, oneFound := order["One"]
	if !otherFound || !oneFound || otherAt > oneAt {
		t.Fatalf("the fixture's views aren't Other's before One's: %v", order)
	}
	var receiver *checker.Type
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == "b" && ast.IsIdentifier(node.AsPropertyAccessExpression().Expression) && node.AsPropertyAccessExpression().Expression.Text() == "either" {
			receiver = keyed.lowering.checker.GetTypeAtLocation(node.AsPropertyAccessExpression().Expression)
		}
		node.ForEachChild(visit)
		return false
	}
	keyed.entry.AsNode().ForEachChild(visit)
	if receiver == nil || receiver.Flags()&checker.TypeFlagsUnion == 0 {
		t.Fatal("the fixture has no either.b through a union receiver")
	}
	if !keyed.lowering.result.CheckedFields[ir.CheckedFieldKey(int(receiver.Id()), "b")] {
		t.Fatalf("either.b through %s isn't checked: One's view registered One:b after Other's walk judged x.b, and nothing asked it again", keyed.lowering.checker.TypeToString(receiver))
	}
}

// Of two accesses a view refuses, the error names the first in the modules' preorder, as the whole walk
// did: the write on line 5, not the one on line 6. Mutant: the indexed walk going on past a refusal and
// keeping the last.
func TestViewRefusesItsFirstSiteInPreorder(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, viewSitesTwoWrites)
	var notYet *NotYet
	if !errors.As(err, &notYet) || notYet.What != "writing a checked object field without its source-slot type certificate" {
		t.Fatalf("got %v, want the checked object field write refused", err)
	}
	if !strings.Contains(notYet.Where, "main.a:5:") {
		t.Fatalf("refused at %s, not the first write's line 5", notYet.Where)
	}
}

// With viewWalkCompare on, every view under no type mapper runs the whole walk and the indexed walk from
// one state and panics where their refusal, getters, CheckedFields or PlainReceivers differ. This lowers
// every .a and .ts program of internal/oracle's corpus, the one internal/split's analysis is exercised
// over, the widest set of views the repository holds, and the pins' own programs, so a mutant the
// corpus never meets still meets them.
//
// Not parallel: viewWalkCompare changes every view in the package while it's set, and no parallel test
// runs beside a sequential one. Its subtests are parallel with one another, and the switch goes off only
// once they're done.
func TestViewWalkIndexedAgreesWithWhole(t *testing.T) {
	viewWalkCompare.Store(true)
	t.Cleanup(func() { viewWalkCompare.Store(false) })
	paths := []string{writeAgreementSource(t, "casts.ts", viewSitesCasts), writeAgreementSource(t, "writes.a", viewSitesTwoWrites)}
	// Reads: internal/oracle/testdata/*
	err := filepath.WalkDir("../oracle/testdata", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// A .d.ts is no program, only declarations another file names.
		if !entry.IsDir() && (filepath.Ext(path) == ".a" || filepath.Ext(path) == ".ts" && !strings.HasSuffix(path, ".d.ts")) {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	var viewed atomic.Int64
	t.Run("programs", func(t *testing.T) {
		for _, path := range paths {
			t.Run(strings.TrimPrefix(filepath.ToSlash(path), "../oracle/testdata/"), func(t *testing.T) {
				t.Parallel()
				loaded, err := load.Load([]string{path})
				if err != nil {
					return // A program that doesn't load makes no view.
				}
				defer func() {
					if failure := recover(); failure != nil {
						t.Fatalf("%s: %v", path, failure)
					}
				}()
				if program, err := Lower(context.Background(), loaded); err == nil && len(program.ViewOrigins) != 0 {
					viewed.Add(1)
				}
			})
		}
	})
	if viewed.Load() == 0 {
		t.Fatal("no program of the corpus lowered with a view, so nothing was compared")
	}
	t.Logf("%d of %d programs lowered with views, each walk compared", viewed.Load(), len(paths))
}
